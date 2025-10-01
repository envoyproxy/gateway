// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package runner

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/docker/docker/pkg/fileutils"
	"github.com/telepresenceio/watchable"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/crypto"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	extension "github.com/envoyproxy/gateway/internal/extension/types"
	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/message"
	"github.com/envoyproxy/gateway/internal/utils"
	"github.com/envoyproxy/gateway/internal/wasm"
)

const (
	// Default certificates path for envoy-gateway with Kubernetes provider.
	serveTLSCertFilepath = "/certs/tls.crt"
	serveTLSKeyFilepath  = "/certs/tls.key"
	serveTLSCaFilepath   = "/certs/ca.crt"

	// TODO: Make these path configurable.
	// Default certificates path for envoy-gateway with Host infrastructure provider.
	localTLSCertFilepath = "/tmp/envoy-gateway/certs/envoy-gateway/tls.crt"
	localTLSKeyFilepath  = "/tmp/envoy-gateway/certs/envoy-gateway/tls.key"
	localTLSCaFilepath   = "/tmp/envoy-gateway/certs/envoy-gateway/ca.crt"

	hmacSecretName = "envoy-oidc-hmac" // nolint: gosec
	hmacSecretKey  = "hmac-secret"
	hmacSecretPath = "/tmp/envoy-gateway/certs/envoy-oidc-hmac/hmac-secret" // nolint: gosec
)

type Config struct {
	config.Server
	ProviderResources *message.ProviderResources
	XdsIR             *message.XdsIR
	InfraIR           *message.InfraIR
	ExtensionManager  extension.Manager
}

type Runner struct {
	Config
	wasmCache wasm.Cache
}

func New(cfg *Config) *Runner {
	return &Runner{
		Config: *cfg,
	}
}

// Close implements Runner interface.
func (r *Runner) Close() error { return nil }

// Name implements Runner interface.
func (r *Runner) Name() string {
	return string(egv1a1.LogComponentGatewayAPIRunner)
}

// Start starts the gateway-api translator runner
func (r *Runner) Start(ctx context.Context) (err error) {
	r.Logger = r.Logger.WithName(r.Name()).WithValues("runner", r.Name())

	go r.startWasmCache(ctx)
	// Do not call .Subscribe() inside Goroutine since it is supposed to be called from the same
	// Goroutine where Close() is called.
	c := r.ProviderResources.GatewayAPIResources.Subscribe(ctx)
	go r.subscribeAndTranslate(c)
	r.Logger.Info("started")
	return
}

func (r *Runner) startWasmCache(ctx context.Context) {
	// Start the wasm cache server
	// EG reuse the OIDC HMAC secret as a hash salt to generate an unguessable
	// downloading path for the Wasm module.
	tlsConfig, salt, err := r.loadTLSConfig(ctx)
	if err != nil {
		r.Logger.Error(err, "failed to start wasm cache")
		return
	}
	cacheOption := wasm.CacheOptions{}
	if r.EnvoyGateway.Provider.Type == egv1a1.ProviderTypeKubernetes {
		cacheOption.CacheDir = "/var/lib/eg/wasm"
	} else {
		h, _ := os.UserHomeDir() // Assume we always get the home directory.
		cacheOption.CacheDir = path.Join(h, ".eg", "wasm")
	}
	// Create the file directory if it does not exist.
	if err = fileutils.CreateIfNotExists(cacheOption.CacheDir, true); err != nil {
		r.Logger.Error(err, "Failed to create Wasm cache directory")
		return
	}
	r.wasmCache = wasm.NewHTTPServerWithFileCache(
		// HTTP server options
		wasm.SeverOptions{
			Salt:      salt,
			TLSConfig: tlsConfig,
		},
		cacheOption, r.ControllerNamespace, r.Logger)
	r.wasmCache.Start(ctx)
}

func (r *Runner) subscribeAndTranslate(sub <-chan watchable.Snapshot[string, *resource.ControllerResources]) {
	message.HandleSubscription(message.Metadata{Runner: r.Name(), Message: message.ProviderResourcesMessageName}, sub,
		func(update message.Update[string, *resource.ControllerResources], errChan chan error) {
			r.Logger.Info("received an update")
			val := update.Value
			// There is only 1 key which is the controller name
			// so when a delete is triggered, delete all IR keys
			if update.Delete || val == nil {
				r.deleteAllIRKeys()
				r.deleteAllStatusKeys()
				return
			}

			// IR keys for watchable
			var curIRKeys, newIRKeys []string

			// Get current IR keys
			for key := range r.InfraIR.LoadAll() {
				curIRKeys = append(curIRKeys, key)
			}

			// Get all status keys from watchable and save them in this StatusesToDelete structure.
			// Iterating through the controller resources, any valid keys will be removed from statusesToDelete.
			// Remaining keys will be deleted from watchable before we exit this function.
			statusesToDelete := r.getAllStatuses()

			// Aggregate metric counters for batch publishing
			var infraIRCount, xdsIRCount, gatewayStatusCount, httpRouteStatusCount, grpcRouteStatusCount int
			var tlsRouteStatusCount, tcpRouteStatusCount, udpRouteStatusCount int
			var backendTLSPolicyStatusCount, clientTrafficPolicyStatusCount, backendTrafficPolicyStatusCount int
			var securityPolicyStatusCount, envoyExtensionPolicyStatusCount, backendStatusCount, extensionServerPolicyStatusCount int

			for _, resources := range *val {
				// Translate and publish IRs.
				t := &gatewayapi.Translator{
					GatewayControllerName:     r.EnvoyGateway.Gateway.ControllerName,
					GatewayClassName:          gwapiv1.ObjectName(resources.GatewayClass.Name),
					GlobalRateLimitEnabled:    r.EnvoyGateway.RateLimit != nil,
					EnvoyPatchPolicyEnabled:   r.EnvoyGateway.ExtensionAPIs != nil && r.EnvoyGateway.ExtensionAPIs.EnableEnvoyPatchPolicy,
					BackendEnabled:            r.EnvoyGateway.ExtensionAPIs != nil && r.EnvoyGateway.ExtensionAPIs.EnableBackend,
					ControllerNamespace:       r.ControllerNamespace,
					GatewayNamespaceMode:      r.EnvoyGateway.GatewayNamespaceMode(),
					MergeGateways:             gatewayapi.IsMergeGatewaysEnabled(resources),
					WasmCache:                 r.wasmCache,
					ListenerPortShiftDisabled: r.EnvoyGateway.Provider != nil && r.EnvoyGateway.Provider.IsRunningOnHost(),
				}

				// If an extension is loaded, pass its supported groups/kinds to the translator
				if r.EnvoyGateway.ExtensionManager != nil {
					var extGKs []schema.GroupKind
					for _, gvk := range r.EnvoyGateway.ExtensionManager.Resources {
						extGKs = append(extGKs, schema.GroupKind{Group: gvk.Group, Kind: gvk.Kind})
					}
					// Include backend resources in extension group kinds for custom backend support
					for _, gvk := range r.EnvoyGateway.ExtensionManager.BackendResources {
						extGKs = append(extGKs, schema.GroupKind{Group: gvk.Group, Kind: gvk.Kind})
					}
					t.ExtensionGroupKinds = extGKs
					r.Logger.Info("extension resources", "GVKs count", len(extGKs))
				}
				// Translate to IR
				result, err := t.Translate(resources)
				if err != nil {
					// In standalone mode, suppress known non-actionable TLS-secret-not-found errors
					if strings.Contains(err.Error(), "TLS secret") && strings.Contains(err.Error(), "not found") && r.EnvoyGateway.Provider.Type != egv1a1.ProviderTypeKubernetes {
						// noop
					} else {
						r.Logger.Error(err, "errors detected during translation", "gateway-class", resources.GatewayClass.Name)
					}
				}

				// Publish the IRs.
				// Also validate the ir before sending it.
				for key, val := range result.InfraIR {
					logger := r.Logger.V(1).WithValues(string(message.InfraIRMessageName), key)
					if logger.Enabled() {
						logger.Info(val.JSONString())
					}
					if err := val.Validate(); err != nil {
						r.Logger.Error(err, "unable to validate infra ir, skipped sending it")
						errChan <- err
					} else {
						r.InfraIR.Store(key, val)
						infraIRCount++
						newIRKeys = append(newIRKeys, key)
					}
				}

				for key, val := range result.XdsIR {
					logger := r.Logger.V(1).WithValues(string(message.XDSIRMessageName), key)
					if logger.Enabled() {
						logger.Info(val.JSONString())
					}
					if err := val.Validate(); err != nil {
						r.Logger.Error(err, "unable to validate xds ir, skipped sending it")
						errChan <- err
					} else {
						r.XdsIR.Store(key, val)
						xdsIRCount++
					}
				}

				// Update Status
				for _, gateway := range result.Gateways {
					key := utils.NamespacedName(gateway)
					r.ProviderResources.GatewayStatuses.Store(key, &gateway.Status)
					gatewayStatusCount++
					delete(statusesToDelete.GatewayStatusKeys, key)
				}
				for _, httpRoute := range result.HTTPRoutes {
					key := utils.NamespacedName(httpRoute)
					r.ProviderResources.HTTPRouteStatuses.Store(key, &httpRoute.Status)
					httpRouteStatusCount++
					delete(statusesToDelete.HTTPRouteStatusKeys, key)
				}
				for _, grpcRoute := range result.GRPCRoutes {
					key := utils.NamespacedName(grpcRoute)
					r.ProviderResources.GRPCRouteStatuses.Store(key, &grpcRoute.Status)
					grpcRouteStatusCount++
					delete(statusesToDelete.GRPCRouteStatusKeys, key)
				}
				for _, tlsRoute := range result.TLSRoutes {
					key := utils.NamespacedName(tlsRoute)
					r.ProviderResources.TLSRouteStatuses.Store(key, &tlsRoute.Status)
					tlsRouteStatusCount++
					delete(statusesToDelete.TLSRouteStatusKeys, key)
				}
				for _, tcpRoute := range result.TCPRoutes {
					key := utils.NamespacedName(tcpRoute)
					r.ProviderResources.TCPRouteStatuses.Store(key, &tcpRoute.Status)
					tcpRouteStatusCount++
					delete(statusesToDelete.TCPRouteStatusKeys, key)
				}
				for _, udpRoute := range result.UDPRoutes {
					key := utils.NamespacedName(udpRoute)
					r.ProviderResources.UDPRouteStatuses.Store(key, &udpRoute.Status)
					udpRouteStatusCount++
					delete(statusesToDelete.UDPRouteStatusKeys, key)
				}

				// Skip updating status for policies with empty status
				// They may have been skipped in this translation because
				// their target is not found (not relevant)

				for _, backendTLSPolicy := range result.BackendTLSPolicies {
					key := utils.NamespacedName(backendTLSPolicy)
					if len(backendTLSPolicy.Status.Ancestors) > 0 {
						r.ProviderResources.BackendTLSPolicyStatuses.Store(key, &backendTLSPolicy.Status)
						backendTLSPolicyStatusCount++
					}
					delete(statusesToDelete.BackendTLSPolicyStatusKeys, key)
				}

				for _, clientTrafficPolicy := range result.ClientTrafficPolicies {
					key := utils.NamespacedName(clientTrafficPolicy)
					if len(clientTrafficPolicy.Status.Ancestors) > 0 {
						r.ProviderResources.ClientTrafficPolicyStatuses.Store(key, &clientTrafficPolicy.Status)
						clientTrafficPolicyStatusCount++
					}
					delete(statusesToDelete.ClientTrafficPolicyStatusKeys, key)
				}
				for _, backendTrafficPolicy := range result.BackendTrafficPolicies {
					key := utils.NamespacedName(backendTrafficPolicy)
					if len(backendTrafficPolicy.Status.Ancestors) > 0 {
						r.ProviderResources.BackendTrafficPolicyStatuses.Store(key, &backendTrafficPolicy.Status)
						backendTrafficPolicyStatusCount++
					}
					delete(statusesToDelete.BackendTrafficPolicyStatusKeys, key)
				}
				for _, securityPolicy := range result.SecurityPolicies {
					key := utils.NamespacedName(securityPolicy)
					if len(securityPolicy.Status.Ancestors) > 0 {
						r.ProviderResources.SecurityPolicyStatuses.Store(key, &securityPolicy.Status)
						securityPolicyStatusCount++
					}
					delete(statusesToDelete.SecurityPolicyStatusKeys, key)
				}
				for _, envoyExtensionPolicy := range result.EnvoyExtensionPolicies {
					key := utils.NamespacedName(envoyExtensionPolicy)
					if len(envoyExtensionPolicy.Status.Ancestors) > 0 {
						r.ProviderResources.EnvoyExtensionPolicyStatuses.Store(key, &envoyExtensionPolicy.Status)
						envoyExtensionPolicyStatusCount++
					}
					delete(statusesToDelete.EnvoyExtensionPolicyStatusKeys, key)
				}
				for _, backend := range result.Backends {
					key := utils.NamespacedName(backend)
					if len(backend.Status.Conditions) > 0 {
						r.ProviderResources.BackendStatuses.Store(key, &backend.Status)
						backendStatusCount++
					}
					delete(statusesToDelete.BackendStatusKeys, key)
				}
				for _, extServerPolicy := range result.ExtensionServerPolicies {
					key := message.NamespacedNameAndGVK{
						NamespacedName:   utils.NamespacedName(&extServerPolicy),
						GroupVersionKind: extServerPolicy.GroupVersionKind(),
					}
					if statusObj, hasStatus := extServerPolicy.Object["status"]; hasStatus && statusObj != nil {
						if statusMap, ok := statusObj.(map[string]any); ok && len(statusMap) > 0 {
							policyStatus := unstructuredToPolicyStatus(statusMap)
							r.ProviderResources.ExtensionPolicyStatuses.Store(key, &policyStatus)
							extensionServerPolicyStatusCount++
						}
					}
					delete(statusesToDelete.ExtensionServerPolicyStatusKeys, key)
				}
			}

			// Publish aggregated metrics
			message.PublishMetric(message.Metadata{Runner: r.Name(), Message: message.InfraIRMessageName}, infraIRCount)
			message.PublishMetric(message.Metadata{Runner: r.Name(), Message: message.XDSIRMessageName}, xdsIRCount)
			message.PublishMetric(message.Metadata{Runner: r.Name(), Message: message.GatewayStatusMessageName}, gatewayStatusCount)
			message.PublishMetric(message.Metadata{Runner: r.Name(), Message: message.HTTPRouteStatusMessageName}, httpRouteStatusCount)
			message.PublishMetric(message.Metadata{Runner: r.Name(), Message: message.GRPCRouteStatusMessageName}, grpcRouteStatusCount)
			message.PublishMetric(message.Metadata{Runner: r.Name(), Message: message.TLSRouteStatusMessageName}, tlsRouteStatusCount)
			message.PublishMetric(message.Metadata{Runner: r.Name(), Message: message.TCPRouteStatusMessageName}, tcpRouteStatusCount)
			message.PublishMetric(message.Metadata{Runner: r.Name(), Message: message.UDPRouteStatusMessageName}, udpRouteStatusCount)
			message.PublishMetric(message.Metadata{Runner: r.Name(), Message: message.BackendTLSPolicyStatusMessageName}, backendTLSPolicyStatusCount)
			message.PublishMetric(message.Metadata{Runner: r.Name(), Message: message.ClientTrafficPolicyStatusMessageName}, clientTrafficPolicyStatusCount)
			message.PublishMetric(message.Metadata{Runner: r.Name(), Message: message.BackendTrafficPolicyStatusMessageName}, backendTrafficPolicyStatusCount)
			message.PublishMetric(message.Metadata{Runner: r.Name(), Message: message.SecurityPolicyStatusMessageName}, securityPolicyStatusCount)
			message.PublishMetric(message.Metadata{Runner: r.Name(), Message: message.EnvoyExtensionPolicyStatusMessageName}, envoyExtensionPolicyStatusCount)
			message.PublishMetric(message.Metadata{Runner: r.Name(), Message: message.BackendStatusMessageName}, backendStatusCount)
			message.PublishMetric(message.Metadata{Runner: r.Name(), Message: message.ExtensionServerPoliciesStatusMessageName}, extensionServerPolicyStatusCount)

			// Delete IR keys
			// There is a 1:1 mapping between infra and xds IR keys
			delKeys := getIRKeysToDelete(curIRKeys, newIRKeys)
			for _, key := range delKeys {
				r.InfraIR.Delete(key)
				r.XdsIR.Delete(key)
			}

			// Delete status keys
			r.deleteStatusKeys(statusesToDelete)
		},
	)
	r.Logger.Info("shutting down")
}

func (r *Runner) loadTLSConfig(ctx context.Context) (tlsConfig *tls.Config, salt []byte, err error) {
	switch {
	case r.EnvoyGateway.Provider.IsRunningOnKubernetes():
		salt, err = hmac(ctx, r.ControllerNamespace)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get hmac secret: %w", err)
		}

		tlsConfig, err = crypto.LoadTLSConfig(serveTLSCertFilepath, serveTLSKeyFilepath, serveTLSCaFilepath)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create tls config: %w", err)
		}

	case r.EnvoyGateway.Provider.IsRunningOnHost():
		salt, err = os.ReadFile(hmacSecretPath)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get hmac secret: %w", err)
		}

		tlsConfig, err = crypto.LoadTLSConfig(localTLSCertFilepath, localTLSKeyFilepath, localTLSCaFilepath)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create tls config: %w", err)
		}

	default:
		return nil, nil, fmt.Errorf("no valid tls certificates")
	}
	return
}

func unstructuredToPolicyStatus(policyStatus map[string]any) gwapiv1a2.PolicyStatus {
	var ret gwapiv1a2.PolicyStatus
	// No need to check the json marshal/unmarshal error, the policyStatus was
	// created via a typed object so the marshalling/unmarshalling will always
	// work
	d, _ := json.Marshal(policyStatus)
	_ = json.Unmarshal(d, &ret)
	return ret
}

// deleteAllIRKeys deletes all XdsIR and InfraIR
func (r *Runner) deleteAllIRKeys() {
	for key := range r.InfraIR.LoadAll() {
		r.InfraIR.Delete(key)
		r.XdsIR.Delete(key)
	}
}

type StatusesToDelete struct {
	GatewayStatusKeys          map[types.NamespacedName]bool
	HTTPRouteStatusKeys        map[types.NamespacedName]bool
	GRPCRouteStatusKeys        map[types.NamespacedName]bool
	TLSRouteStatusKeys         map[types.NamespacedName]bool
	TCPRouteStatusKeys         map[types.NamespacedName]bool
	UDPRouteStatusKeys         map[types.NamespacedName]bool
	BackendTLSPolicyStatusKeys map[types.NamespacedName]bool

	ClientTrafficPolicyStatusKeys   map[types.NamespacedName]bool
	BackendTrafficPolicyStatusKeys  map[types.NamespacedName]bool
	SecurityPolicyStatusKeys        map[types.NamespacedName]bool
	EnvoyExtensionPolicyStatusKeys  map[types.NamespacedName]bool
	ExtensionServerPolicyStatusKeys map[message.NamespacedNameAndGVK]bool

	BackendStatusKeys map[types.NamespacedName]bool
}

func (r *Runner) getAllStatuses() *StatusesToDelete {
	// Maps storing status keys to be deleted
	ds := &StatusesToDelete{
		GatewayStatusKeys:   make(map[types.NamespacedName]bool),
		HTTPRouteStatusKeys: make(map[types.NamespacedName]bool),
		GRPCRouteStatusKeys: make(map[types.NamespacedName]bool),
		TLSRouteStatusKeys:  make(map[types.NamespacedName]bool),
		TCPRouteStatusKeys:  make(map[types.NamespacedName]bool),
		UDPRouteStatusKeys:  make(map[types.NamespacedName]bool),

		ClientTrafficPolicyStatusKeys:   make(map[types.NamespacedName]bool),
		BackendTrafficPolicyStatusKeys:  make(map[types.NamespacedName]bool),
		SecurityPolicyStatusKeys:        make(map[types.NamespacedName]bool),
		BackendTLSPolicyStatusKeys:      make(map[types.NamespacedName]bool),
		EnvoyExtensionPolicyStatusKeys:  make(map[types.NamespacedName]bool),
		ExtensionServerPolicyStatusKeys: make(map[message.NamespacedNameAndGVK]bool),

		BackendStatusKeys: make(map[types.NamespacedName]bool),
	}

	// Get current status keys
	for key := range r.ProviderResources.GatewayStatuses.LoadAll() {
		ds.GatewayStatusKeys[key] = true
	}
	for key := range r.ProviderResources.HTTPRouteStatuses.LoadAll() {
		ds.HTTPRouteStatusKeys[key] = true
	}
	for key := range r.ProviderResources.GRPCRouteStatuses.LoadAll() {
		ds.GRPCRouteStatusKeys[key] = true
	}
	for key := range r.ProviderResources.TLSRouteStatuses.LoadAll() {
		ds.TLSRouteStatusKeys[key] = true
	}
	for key := range r.ProviderResources.TCPRouteStatuses.LoadAll() {
		ds.TCPRouteStatusKeys[key] = true
	}
	for key := range r.ProviderResources.UDPRouteStatuses.LoadAll() {
		ds.UDPRouteStatusKeys[key] = true
	}
	for key := range r.ProviderResources.BackendTLSPolicyStatuses.LoadAll() {
		ds.BackendTLSPolicyStatusKeys[key] = true
	}

	for key := range r.ProviderResources.ClientTrafficPolicyStatuses.LoadAll() {
		ds.ClientTrafficPolicyStatusKeys[key] = true
	}
	for key := range r.ProviderResources.BackendTrafficPolicyStatuses.LoadAll() {
		ds.BackendTrafficPolicyStatusKeys[key] = true
	}
	for key := range r.ProviderResources.SecurityPolicyStatuses.LoadAll() {
		ds.SecurityPolicyStatusKeys[key] = true
	}
	for key := range r.ProviderResources.EnvoyExtensionPolicyStatuses.LoadAll() {
		ds.EnvoyExtensionPolicyStatusKeys[key] = true
	}
	for key := range r.ProviderResources.BackendStatuses.LoadAll() {
		ds.BackendStatusKeys[key] = true
	}
	return ds
}

func (r *Runner) deleteStatusKeys(ds *StatusesToDelete) {
	for key := range ds.GatewayStatusKeys {
		r.ProviderResources.GatewayStatuses.Delete(key)
		delete(ds.GatewayStatusKeys, key)
	}
	for key := range ds.HTTPRouteStatusKeys {
		r.ProviderResources.HTTPRouteStatuses.Delete(key)
		delete(ds.HTTPRouteStatusKeys, key)
	}
	for key := range ds.GRPCRouteStatusKeys {
		r.ProviderResources.GRPCRouteStatuses.Delete(key)
		delete(ds.GRPCRouteStatusKeys, key)
	}
	for key := range ds.TLSRouteStatusKeys {
		r.ProviderResources.TLSRouteStatuses.Delete(key)
		delete(ds.TLSRouteStatusKeys, key)
	}
	for key := range ds.TCPRouteStatusKeys {
		r.ProviderResources.TCPRouteStatuses.Delete(key)
		delete(ds.TCPRouteStatusKeys, key)
	}
	for key := range ds.UDPRouteStatusKeys {
		r.ProviderResources.UDPRouteStatuses.Delete(key)
		delete(ds.UDPRouteStatusKeys, key)
	}

	for key := range ds.ClientTrafficPolicyStatusKeys {
		r.ProviderResources.ClientTrafficPolicyStatuses.Delete(key)
		delete(ds.ClientTrafficPolicyStatusKeys, key)
	}
	for key := range ds.BackendTrafficPolicyStatusKeys {
		r.ProviderResources.BackendTrafficPolicyStatuses.Delete(key)
		delete(ds.BackendTrafficPolicyStatusKeys, key)
	}
	for key := range ds.SecurityPolicyStatusKeys {
		r.ProviderResources.SecurityPolicyStatuses.Delete(key)
		delete(ds.SecurityPolicyStatusKeys, key)
	}
	for key := range ds.BackendTLSPolicyStatusKeys {
		r.ProviderResources.BackendTLSPolicyStatuses.Delete(key)
		delete(ds.BackendTLSPolicyStatusKeys, key)
	}
	for key := range ds.EnvoyExtensionPolicyStatusKeys {
		r.ProviderResources.EnvoyExtensionPolicyStatuses.Delete(key)
		delete(ds.EnvoyExtensionPolicyStatusKeys, key)
	}
	for key := range ds.ExtensionServerPolicyStatusKeys {
		r.ProviderResources.ExtensionPolicyStatuses.Delete(key)
		delete(ds.ExtensionServerPolicyStatusKeys, key)
	}
	for key := range ds.BackendStatusKeys {
		r.ProviderResources.BackendStatuses.Delete(key)
		delete(ds.BackendStatusKeys, key)
	}
}

// deleteAllStatusKeys deletes all status keys stored by the subscriber.
func (r *Runner) deleteAllStatusKeys() {
	// Fields of GatewayAPIStatuses
	for key := range r.ProviderResources.GatewayStatuses.LoadAll() {
		r.ProviderResources.GatewayStatuses.Delete(key)
	}
	for key := range r.ProviderResources.HTTPRouteStatuses.LoadAll() {
		r.ProviderResources.HTTPRouteStatuses.Delete(key)
	}
	for key := range r.ProviderResources.GRPCRouteStatuses.LoadAll() {
		r.ProviderResources.GRPCRouteStatuses.Delete(key)
	}
	for key := range r.ProviderResources.TLSRouteStatuses.LoadAll() {
		r.ProviderResources.TLSRouteStatuses.Delete(key)
	}
	for key := range r.ProviderResources.TCPRouteStatuses.LoadAll() {
		r.ProviderResources.TCPRouteStatuses.Delete(key)
	}
	for key := range r.ProviderResources.UDPRouteStatuses.LoadAll() {
		r.ProviderResources.UDPRouteStatuses.Delete(key)
	}
	for key := range r.ProviderResources.BackendTLSPolicyStatuses.LoadAll() {
		r.ProviderResources.BackendTLSPolicyStatuses.Delete(key)
	}

	// Fields of PolicyStatuses
	for key := range r.ProviderResources.ClientTrafficPolicyStatuses.LoadAll() {
		r.ProviderResources.ClientTrafficPolicyStatuses.Delete(key)
	}
	for key := range r.ProviderResources.BackendTrafficPolicyStatuses.LoadAll() {
		r.ProviderResources.BackendTrafficPolicyStatuses.Delete(key)
	}
	for key := range r.ProviderResources.SecurityPolicyStatuses.LoadAll() {
		r.ProviderResources.SecurityPolicyStatuses.Delete(key)
	}
	for key := range r.ProviderResources.EnvoyExtensionPolicyStatuses.LoadAll() {
		r.ProviderResources.EnvoyExtensionPolicyStatuses.Delete(key)
	}
	for key := range r.ProviderResources.ExtensionPolicyStatuses.LoadAll() {
		r.ProviderResources.ExtensionPolicyStatuses.Delete(key)
	}
	for key := range r.ProviderResources.BackendStatuses.LoadAll() {
		r.ProviderResources.BackendStatuses.Delete(key)
	}
}

// getIRKeysToDelete returns the list of IR keys to delete
// based on the difference between the current keys and the
// new keys parameters passed to the function.
func getIRKeysToDelete(curKeys, newKeys []string) []string {
	curSet := sets.NewString(curKeys...)
	newSet := sets.NewString(newKeys...)

	delSet := curSet.Difference(newSet)

	return delSet.List()
}

// hmac returns the HMAC secret generated by the CertGen job.
// hmac will be used as a hash salt to generate unguessable downloading paths for Wasm modules.
func hmac(ctx context.Context, namespace string) (hmac []byte, err error) {
	// Get the HMAC secret.
	// HMAC secret is generated by the CertGen job and stored in a secret
	cfg, err := ctrl.GetConfig()
	if err != nil {
		return nil, err
	}
	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	secret, err := client.CoreV1().Secrets(namespace).Get(ctx, hmacSecretName, metav1.GetOptions{})
	if err != nil {
		if kerrors.IsNotFound(err) {
			return nil, fmt.Errorf("HMAC secret %s/%s not found", namespace, hmacSecretName)
		}
		return nil, err
	}
	hmac, ok := secret.Data[hmacSecretKey]
	if !ok || len(hmac) == 0 {
		return nil, fmt.Errorf(
			"HMAC secret not found in secret %s/%s", namespace, hmacSecretName)
	}
	return
}
