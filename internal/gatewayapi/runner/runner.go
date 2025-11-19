// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package runner

import (
	"context"
	"crypto/tls"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sync"

	"github.com/docker/docker/pkg/fileutils"
	"github.com/telepresenceio/watchable"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
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
	"github.com/envoyproxy/gateway/internal/infrastructure/host"
	"github.com/envoyproxy/gateway/internal/message"
	"github.com/envoyproxy/gateway/internal/utils"
	"github.com/envoyproxy/gateway/internal/wasm"
)

const (
	// Default certificates path for envoy-gateway with Kubernetes provider.
	serveTLSCertFilepath = "/certs/tls.crt"
	serveTLSKeyFilepath  = "/certs/tls.key"
	serveTLSCaFilepath   = "/certs/ca.crt"

	hmacSecretName = "envoy-oidc-hmac" // nolint: gosec
	hmacSecretKey  = "hmac-secret"
)

var tracer = otel.Tracer("envoy-gateway/gateway-api")

type Config struct {
	config.Server
	ProviderResources *message.ProviderResources
	RunnerErrors      *message.RunnerErrors
	XdsIR             *message.XdsIR
	InfraIR           *message.InfraIR
	ExtensionManager  extension.Manager
}

type Runner struct {
	Config
	wasmCache wasm.Cache

	// Key tracking for mark and sweep - avoids expensive LoadAll operations
	keyCache *KeyCache

	// Goroutine synchronization
	done sync.WaitGroup
}

func New(cfg *Config) *Runner {
	return &Runner{
		Config:   *cfg,
		keyCache: newKeyCache(),
	}
}

// Close implements Runner interface.
func (r *Runner) Close() error {
	r.done.Wait()
	return nil
}

// Name implements Runner interface.
func (r *Runner) Name() string {
	return string(egv1a1.LogComponentGatewayAPIRunner)
}

// Start starts the gateway-api translator runner
func (r *Runner) Start(ctx context.Context) error {
	r.Logger = r.Logger.WithName(r.Name()).WithValues("runner", r.Name())
	r.done.Go(func() {
		r.startWasmCache(ctx)
	})
	// Do not call .Subscribe() inside Goroutine since it is supposed to be called from the same
	// Goroutine where Close() is called.
	c := r.ProviderResources.GatewayAPIResources.Subscribe(ctx)
	r.done.Go(func() {
		r.subscribeAndTranslate(c)
	})
	r.Logger.Info("started")
	return nil
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

func (r *Runner) subscribeAndTranslate(sub <-chan watchable.Snapshot[string, *resource.ControllerResourcesContext]) {
	message.HandleSubscription(
		r.Logger,
		message.Metadata{Runner: r.Name(), Message: message.ProviderResourcesMessageName}, sub,
		func(update message.Update[string, *resource.ControllerResourcesContext], errChan chan error) {
			parentCtx := context.Background()
			if update.Value != nil && update.Value.Context != nil {
				parentCtx = update.Value.Context
			}

			traceCtx, span := tracer.Start(parentCtx, "GatewayApiRunner.subscribeAndTranslate")
			defer span.End()
			traceLogger := r.Logger.WithTrace(traceCtx)
			traceLogger.Info("received an update", "key", update.Key)

			valWrapper := update.Value
			// There is only 1 key which is the controller name
			// so when a delete is triggered, delete all keys
			if update.Delete || valWrapper == nil || valWrapper.Resources == nil {
				r.deleteAllKeys()
				return
			}

			val := valWrapper.Resources

			// Add span attributes for observability
			span.SetAttributes(
				attribute.String("controller.key", update.Key),
				attribute.Bool("update.delete", update.Delete),
			)
			if val != nil {
				span.SetAttributes(attribute.Int("resources.count", len(*val)))
			}

			// Initialize keysToDelete with tracked keys (mark and sweep approach)
			keysToDelete := r.keyCache.copy()

			// Aggregate metric counters for batch publishing
			var infraIRCount, xdsIRCount, gatewayStatusCount, xListenerSetStatusCount, httpRouteStatusCount, grpcRouteStatusCount int
			var tlsRouteStatusCount, tcpRouteStatusCount, udpRouteStatusCount int
			var backendTLSPolicyStatusCount, clientTrafficPolicyStatusCount, backendTrafficPolicyStatusCount int
			var securityPolicyStatusCount, envoyExtensionPolicyStatusCount, backendStatusCount, extensionServerPolicyStatusCount int

			// `aggregatedStatuses` aggregates status result of resources from all
			// parents/ancestors, and then stores the status once for every resource.
			aggregatedStatuses := struct {
				HTTPRoutes              map[types.NamespacedName]*gwapiv1.RouteStatus
				GRPCRoutes              map[types.NamespacedName]*gwapiv1.RouteStatus
				TLSRoutes               map[types.NamespacedName]*gwapiv1a2.RouteStatus
				TCPRoutes               map[types.NamespacedName]*gwapiv1a2.RouteStatus
				UDPRoutes               map[types.NamespacedName]*gwapiv1a2.RouteStatus
				BackendTLSPolicies      map[types.NamespacedName]*gwapiv1.PolicyStatus
				ClientTrafficPolicies   map[types.NamespacedName]*gwapiv1.PolicyStatus
				BackendTrafficPolicies  map[types.NamespacedName]*gwapiv1.PolicyStatus
				SecurityPolicies        map[types.NamespacedName]*gwapiv1.PolicyStatus
				EnvoyExtensionPolicies  map[types.NamespacedName]*gwapiv1.PolicyStatus
				ExtensionServerPolicies map[message.NamespacedNameAndGVK]*gwapiv1.PolicyStatus
			}{
				HTTPRoutes:              make(map[types.NamespacedName]*gwapiv1.RouteStatus),
				GRPCRoutes:              make(map[types.NamespacedName]*gwapiv1.RouteStatus),
				TLSRoutes:               make(map[types.NamespacedName]*gwapiv1a2.RouteStatus),
				TCPRoutes:               make(map[types.NamespacedName]*gwapiv1a2.RouteStatus),
				UDPRoutes:               make(map[types.NamespacedName]*gwapiv1a2.RouteStatus),
				BackendTLSPolicies:      make(map[types.NamespacedName]*gwapiv1.PolicyStatus),
				ClientTrafficPolicies:   make(map[types.NamespacedName]*gwapiv1.PolicyStatus),
				BackendTrafficPolicies:  make(map[types.NamespacedName]*gwapiv1.PolicyStatus),
				SecurityPolicies:        make(map[types.NamespacedName]*gwapiv1.PolicyStatus),
				EnvoyExtensionPolicies:  make(map[types.NamespacedName]*gwapiv1.PolicyStatus),
				ExtensionServerPolicies: make(map[message.NamespacedNameAndGVK]*gwapiv1.PolicyStatus),
			}

			mergeRouteStatus := func(status, other *gwapiv1.RouteStatus) *gwapiv1.RouteStatus {
				if other != nil {
					if status != nil {
						status.Parents = append(status.Parents, other.Parents...)
					} else {
						return other
					}
				}
				return status
			}
			mergePolicyStatus := func(status, other *gwapiv1.PolicyStatus) *gwapiv1.PolicyStatus {
				if other != nil {
					if status != nil {
						status.Ancestors = append(status.Ancestors, other.Ancestors...)
					} else {
						return other
					}
				}
				return status
			}

			span.AddEvent("translate", trace.WithAttributes(attribute.Int("resources.count", len(*val))))
			for _, resources := range *val {
				// Translate and publish IRs.
				t := &gatewayapi.Translator{
					GatewayControllerName:           r.EnvoyGateway.Gateway.ControllerName,
					GatewayClassName:                gwapiv1.ObjectName(resources.GatewayClass.Name),
					GlobalRateLimitEnabled:          r.EnvoyGateway.RateLimit != nil,
					EnvoyPatchPolicyEnabled:         r.EnvoyGateway.ExtensionAPIs != nil && r.EnvoyGateway.ExtensionAPIs.EnableEnvoyPatchPolicy,
					BackendEnabled:                  r.EnvoyGateway.ExtensionAPIs != nil && r.EnvoyGateway.ExtensionAPIs.EnableBackend,
					ControllerNamespace:             r.ControllerNamespace,
					GatewayNamespaceMode:            r.EnvoyGateway.GatewayNamespaceMode(),
					MergeGateways:                   gatewayapi.IsMergeGatewaysEnabled(resources),
					WasmCache:                       r.wasmCache,
					RunningOnHost:                   r.EnvoyGateway.Provider != nil && r.EnvoyGateway.Provider.IsRunningOnHost(),
					Logger:                          traceLogger,
					LuaEnvoyExtensionPolicyDisabled: r.EnvoyGateway.ExtensionAPIs != nil && r.EnvoyGateway.ExtensionAPIs.DisableLua,
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
					traceLogger.Info("extension resources", "GVKs count", len(extGKs))
				}
				// Translate to IR
				_, translateToIRSpan := tracer.Start(traceCtx, "GatewayApiRunner.ResoureTranslationCycle.TranslateToIR")
				result, err := t.Translate(resources)
				translateToIRSpan.End()
				if err != nil {
					// Currently all errors that Translate returns should just be logged
					traceLogger.Error(err, "errors detected during translation", "gateway-class", resources.GatewayClass.Name)
					// Notify the main control loop about translation errors. This may be a critical error in standalone mode, so
					// notify the control loop in case this needs to be handled.
					r.RunnerErrors.Store(r.Name(), message.NewWatchableError(err))
				}

				// Publish the IRs.
				// Also validate the ir before sending it.
				for key, val := range result.InfraIR {
					logV := traceLogger.V(1).WithValues(string(message.InfraIRMessageName), key)
					if logV.Enabled() {
						logV.Info(val.JSONString())
					}
					if err := val.Validate(); err != nil {
						traceLogger.Error(err, "unable to validate infra ir, skipped sending it")
						errChan <- err
					} else {
						r.InfraIR.Store(key, val)
						infraIRCount++
						// Track IR key for mark and sweep
						r.keyCache.IR[key] = true
						delete(keysToDelete.IR, key)
					}
				}

				for key, val := range result.XdsIR {
					logV := traceLogger.V(1).WithValues(string(message.XDSIRMessageName), key)
					if logV.Enabled() {
						logV.Info(val.JSONString())
					}
					if err := val.Validate(); err != nil {
						traceLogger.Error(err, "unable to validate xds ir, skipped sending it")
						errChan <- err
					} else {
						m := message.XdsIRWithContext{
							XdsIR:   val,
							Context: traceCtx,
						}
						r.XdsIR.Store(key, &m)
						xdsIRCount++
					}
				}

				// Update Status
				_, statusUpdateSpan := tracer.Start(traceCtx, "GatewayApiRunner.ResoureTranslationCycle.UpdateStatus")
				if result.GatewayClass != nil {
					key := utils.NamespacedName(result.GatewayClass)
					r.ProviderResources.GatewayClassStatuses.Store(key, &result.GatewayClass.Status)
				}

				// 1. Resources which can only belong to 1 GatewayClass (at most) get their statuses stored right away.
				for _, gateway := range result.Gateways {
					key := utils.NamespacedName(gateway)
					r.ProviderResources.GatewayStatuses.Store(key, &gateway.Status)
					gatewayStatusCount++
					delete(keysToDelete.GatewayStatus, key)
					r.keyCache.GatewayStatus[key] = true
				}
				for _, backend := range result.Backends {
					key := utils.NamespacedName(backend)
					if len(backend.Status.Conditions) > 0 {
						r.ProviderResources.BackendStatuses.Store(key, &backend.Status)
						backendStatusCount++
					}
					delete(keysToDelete.BackendStatus, key)
					r.keyCache.BackendStatus[key] = true
				}
				for _, xListenerSet := range result.XListenerSets {
					key := utils.NamespacedName(xListenerSet)
					r.ProviderResources.XListenerSetStatuses.Store(key, &xListenerSet.Status)
					xListenerSetStatusCount++
					delete(keysToDelete.XListenerSetStatus, key)
					r.keyCache.XListenerSetStatus[key] = true
				}
				// 2. Resources which can belong to multiple GatewayClasses get their
				//    status aggregated, then stored once after iterating over all GatewayClasses.
				for _, httpRoute := range result.HTTPRoutes {
					if len(httpRoute.Status.Parents) != 0 {
						key := utils.NamespacedName(httpRoute)
						aggregatedStatuses.HTTPRoutes[key] = mergeRouteStatus(aggregatedStatuses.HTTPRoutes[key], &httpRoute.Status.RouteStatus)
					}
				}
				for _, grpcRoute := range result.GRPCRoutes {
					if len(grpcRoute.Status.Parents) != 0 {
						key := utils.NamespacedName(grpcRoute)
						aggregatedStatuses.GRPCRoutes[key] = mergeRouteStatus(aggregatedStatuses.GRPCRoutes[key], &grpcRoute.Status.RouteStatus)
					}
				}
				for _, tlsRoute := range result.TLSRoutes {
					if len(tlsRoute.Status.Parents) != 0 {
						key := utils.NamespacedName(tlsRoute)
						aggregatedStatuses.TLSRoutes[key] = mergeRouteStatus(aggregatedStatuses.TLSRoutes[key], &tlsRoute.Status.RouteStatus)
					}
				}
				for _, tcpRoute := range result.TCPRoutes {
					if len(tcpRoute.Status.Parents) != 0 {
						key := utils.NamespacedName(tcpRoute)
						aggregatedStatuses.TCPRoutes[key] = mergeRouteStatus(aggregatedStatuses.TCPRoutes[key], &tcpRoute.Status.RouteStatus)
					}
				}
				for _, udpRoute := range result.UDPRoutes {
					if len(udpRoute.Status.Parents) != 0 {
						key := utils.NamespacedName(udpRoute)
						aggregatedStatuses.UDPRoutes[key] = mergeRouteStatus(aggregatedStatuses.UDPRoutes[key], &udpRoute.Status.RouteStatus)
					}
				}
				for _, backendTLSPolicy := range result.BackendTLSPolicies {
					if len(backendTLSPolicy.Status.Ancestors) != 0 {
						key := utils.NamespacedName(backendTLSPolicy)
						aggregatedStatuses.BackendTLSPolicies[key] = mergePolicyStatus(aggregatedStatuses.BackendTLSPolicies[key], &backendTLSPolicy.Status)
					}
				}
				for _, clientTrafficPolicy := range result.ClientTrafficPolicies {
					if len(clientTrafficPolicy.Status.Ancestors) != 0 {
						key := utils.NamespacedName(clientTrafficPolicy)
						aggregatedStatuses.ClientTrafficPolicies[key] = mergePolicyStatus(aggregatedStatuses.ClientTrafficPolicies[key], &clientTrafficPolicy.Status)
					}
				}
				for _, backendTrafficPolicy := range result.BackendTrafficPolicies {
					if len(backendTrafficPolicy.Status.Ancestors) != 0 {
						key := utils.NamespacedName(backendTrafficPolicy)
						aggregatedStatuses.BackendTrafficPolicies[key] = mergePolicyStatus(aggregatedStatuses.BackendTrafficPolicies[key], &backendTrafficPolicy.Status)
					}
				}
				for _, securityPolicy := range result.SecurityPolicies {
					if len(securityPolicy.Status.Ancestors) != 0 {
						key := utils.NamespacedName(securityPolicy)
						aggregatedStatuses.SecurityPolicies[key] = mergePolicyStatus(aggregatedStatuses.SecurityPolicies[key], &securityPolicy.Status)
					}
				}
				for _, envoyExtensionPolicy := range result.EnvoyExtensionPolicies {
					if len(envoyExtensionPolicy.Status.Ancestors) != 0 {
						key := utils.NamespacedName(envoyExtensionPolicy)
						aggregatedStatuses.EnvoyExtensionPolicies[key] = mergePolicyStatus(aggregatedStatuses.EnvoyExtensionPolicies[key], &envoyExtensionPolicy.Status)
					}
				}
				for _, extServerPolicy := range result.ExtensionServerPolicies {
					status := gatewayapi.ExtServerPolicyStatusAsPolicyStatus(&extServerPolicy)
					if len(status.Ancestors) != 0 {
						key := message.NamespacedNameAndGVK{
							NamespacedName:   utils.NamespacedName(&extServerPolicy),
							GroupVersionKind: extServerPolicy.GroupVersionKind(),
						}
						aggregatedStatuses.ExtensionServerPolicies[key] = mergePolicyStatus(aggregatedStatuses.ExtensionServerPolicies[key], &status)
					}
				}
				statusUpdateSpan.End()
			}

			// Store the stauses of all objects atomically with the aggregated status.
			for key, status := range aggregatedStatuses.HTTPRoutes {
				s := gwapiv1.HTTPRouteStatus{RouteStatus: *status}
				r.ProviderResources.HTTPRouteStatuses.Store(key, &s)
				httpRouteStatusCount++
				delete(keysToDelete.HTTPRouteStatus, key)
				r.keyCache.HTTPRouteStatus[key] = true
			}
			for key, status := range aggregatedStatuses.GRPCRoutes {
				s := gwapiv1.GRPCRouteStatus{RouteStatus: *status}
				r.ProviderResources.GRPCRouteStatuses.Store(key, &s)
				grpcRouteStatusCount++
				delete(keysToDelete.GRPCRouteStatus, key)
				r.keyCache.GRPCRouteStatus[key] = true
			}
			for key, status := range aggregatedStatuses.TLSRoutes {
				s := gwapiv1a2.TLSRouteStatus{RouteStatus: *status}
				r.ProviderResources.TLSRouteStatuses.Store(key, &s)
				tlsRouteStatusCount++
				delete(keysToDelete.TLSRouteStatus, key)
				r.keyCache.TLSRouteStatus[key] = true
			}
			for key, status := range aggregatedStatuses.TCPRoutes {
				s := gwapiv1a2.TCPRouteStatus{RouteStatus: *status}
				r.ProviderResources.TCPRouteStatuses.Store(key, &s)
				tcpRouteStatusCount++
				delete(keysToDelete.TCPRouteStatus, key)
				r.keyCache.TCPRouteStatus[key] = true
			}
			for key, status := range aggregatedStatuses.UDPRoutes {
				s := gwapiv1a2.UDPRouteStatus{RouteStatus: *status}
				r.ProviderResources.UDPRouteStatuses.Store(key, &s)
				udpRouteStatusCount++
				delete(keysToDelete.UDPRouteStatus, key)
				r.keyCache.UDPRouteStatus[key] = true
			}
			for key, status := range aggregatedStatuses.BackendTLSPolicies {
				r.ProviderResources.BackendTLSPolicyStatuses.Store(key, status)
				backendTLSPolicyStatusCount++
				delete(keysToDelete.BackendTLSPolicyStatus, key)
				r.keyCache.BackendTLSPolicyStatus[key] = true
			}
			for key, status := range aggregatedStatuses.ClientTrafficPolicies {
				r.ProviderResources.ClientTrafficPolicyStatuses.Store(key, status)
				clientTrafficPolicyStatusCount++
				delete(keysToDelete.ClientTrafficPolicyStatus, key)
				r.keyCache.ClientTrafficPolicyStatus[key] = true
			}
			for key, status := range aggregatedStatuses.BackendTrafficPolicies {
				r.ProviderResources.BackendTrafficPolicyStatuses.Store(key, status)
				backendTrafficPolicyStatusCount++
				delete(keysToDelete.BackendTrafficPolicyStatus, key)
				r.keyCache.BackendTrafficPolicyStatus[key] = true
			}
			for key, status := range aggregatedStatuses.SecurityPolicies {
				r.ProviderResources.SecurityPolicyStatuses.Store(key, status)
				securityPolicyStatusCount++
				delete(keysToDelete.SecurityPolicyStatus, key)
				r.keyCache.SecurityPolicyStatus[key] = true
			}
			for key, status := range aggregatedStatuses.EnvoyExtensionPolicies {
				r.ProviderResources.EnvoyExtensionPolicyStatuses.Store(key, status)
				envoyExtensionPolicyStatusCount++
				delete(keysToDelete.EnvoyExtensionPolicyStatus, key)
				r.keyCache.EnvoyExtensionPolicyStatus[key] = true
			}
			for key, status := range aggregatedStatuses.ExtensionServerPolicies {
				r.ProviderResources.ExtensionPolicyStatuses.Store(key, status)
				extensionServerPolicyStatusCount++
				delete(keysToDelete.ExtensionServerPolicyStatus, key)
				r.keyCache.ExtensionServerPolicyStatus[key] = true
			}
			// Publish aggregated metrics
			message.PublishMetric(message.Metadata{Runner: r.Name(), Message: message.InfraIRMessageName}, infraIRCount)
			message.PublishMetric(message.Metadata{Runner: r.Name(), Message: message.XDSIRMessageName}, xdsIRCount)
			message.PublishMetric(message.Metadata{Runner: r.Name(), Message: message.GatewayClassStatusMessageName}, 1)
			message.PublishMetric(message.Metadata{Runner: r.Name(), Message: message.GatewayStatusMessageName}, gatewayStatusCount)
			message.PublishMetric(message.Metadata{Runner: r.Name(), Message: message.XListenerSetStatusMessageName}, xListenerSetStatusCount)
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

			// Delete keys using mark and sweep
			r.deleteKeys(keysToDelete)
		},
	)
	r.Logger.Info("shutting down")
}

func (r *Runner) loadTLSConfig(ctx context.Context) (*tls.Config, []byte, error) {
	switch {
	case r.EnvoyGateway.Provider.IsRunningOnKubernetes():
		salt, err := hmac(ctx, r.ControllerNamespace)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get hmac secret: %w", err)
		}

		tlsConfig, err := crypto.LoadTLSConfig(serveTLSCertFilepath, serveTLSKeyFilepath, serveTLSCaFilepath)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create tls config: %w", err)
		}
		return tlsConfig, salt, nil

	case r.EnvoyGateway.Provider.IsRunningOnHost():
		// Get config
		var hostCfg *egv1a1.EnvoyGatewayHostInfrastructureProvider
		if p := r.EnvoyGateway.Provider; p != nil && p.Custom != nil &&
			p.Custom.Infrastructure != nil && p.Custom.Infrastructure.Host != nil {
			hostCfg = p.Custom.Infrastructure.Host
		}

		paths, err := host.GetPaths(hostCfg)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to determine paths: %w", err)
		}

		// Read HMAC secret
		hmacPath := filepath.Join(paths.CertDir("envoy-oidc-hmac"), "hmac-secret")
		salt, err := os.ReadFile(hmacPath)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get hmac secret: %w", err)
		}

		certDir := paths.CertDir("envoy-gateway")
		certPath := filepath.Join(certDir, "tls.crt")
		keyPath := filepath.Join(certDir, "tls.key")
		caPath := filepath.Join(certDir, "ca.crt")

		tlsConfig, err := crypto.LoadTLSConfig(certPath, keyPath, caPath)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create tls config: %w", err)
		}
		return tlsConfig, salt, nil

	default:
		return nil, nil, fmt.Errorf("no valid tls certificates")
	}
}

// deleteAllIRKeys deletes all XdsIR and InfraIR using tracked keys
func (r *Runner) deleteAllKeys() {
	// Delete IR keys
	for key := range r.keyCache.IR {
		r.InfraIR.Delete(key)
		r.XdsIR.Delete(key)
	}

	// Delete status keys
	for key := range r.keyCache.GatewayStatus {
		r.ProviderResources.GatewayStatuses.Delete(key)
	}
	for key := range r.keyCache.XListenerSetStatus {
		r.ProviderResources.XListenerSetStatuses.Delete(key)
	}
	for key := range r.keyCache.HTTPRouteStatus {
		r.ProviderResources.HTTPRouteStatuses.Delete(key)
	}
	for key := range r.keyCache.GRPCRouteStatus {
		r.ProviderResources.GRPCRouteStatuses.Delete(key)
	}
	for key := range r.keyCache.TLSRouteStatus {
		r.ProviderResources.TLSRouteStatuses.Delete(key)
	}
	for key := range r.keyCache.TCPRouteStatus {
		r.ProviderResources.TCPRouteStatuses.Delete(key)
	}
	for key := range r.keyCache.UDPRouteStatus {
		r.ProviderResources.UDPRouteStatuses.Delete(key)
	}
	for key := range r.keyCache.BackendTLSPolicyStatus {
		r.ProviderResources.BackendTLSPolicyStatuses.Delete(key)
	}
	for key := range r.keyCache.ClientTrafficPolicyStatus {
		r.ProviderResources.ClientTrafficPolicyStatuses.Delete(key)
	}
	for key := range r.keyCache.BackendTrafficPolicyStatus {
		r.ProviderResources.BackendTrafficPolicyStatuses.Delete(key)
	}
	for key := range r.keyCache.SecurityPolicyStatus {
		r.ProviderResources.SecurityPolicyStatuses.Delete(key)
	}
	for key := range r.keyCache.EnvoyExtensionPolicyStatus {
		r.ProviderResources.EnvoyExtensionPolicyStatuses.Delete(key)
	}
	for key := range r.keyCache.ExtensionServerPolicyStatus {
		r.ProviderResources.ExtensionPolicyStatuses.Delete(key)
	}
	for key := range r.keyCache.BackendStatus {
		r.ProviderResources.BackendStatuses.Delete(key)
	}

	// Clear all tracking
	r.keyCache = newKeyCache()
}

type KeyCache struct {
	// IR keys
	IR map[string]bool

	// Status keys
	GatewayStatus          map[types.NamespacedName]bool
	XListenerSetStatus     map[types.NamespacedName]bool
	HTTPRouteStatus        map[types.NamespacedName]bool
	GRPCRouteStatus        map[types.NamespacedName]bool
	TLSRouteStatus         map[types.NamespacedName]bool
	TCPRouteStatus         map[types.NamespacedName]bool
	UDPRouteStatus         map[types.NamespacedName]bool
	BackendTLSPolicyStatus map[types.NamespacedName]bool

	ClientTrafficPolicyStatus   map[types.NamespacedName]bool
	BackendTrafficPolicyStatus  map[types.NamespacedName]bool
	SecurityPolicyStatus        map[types.NamespacedName]bool
	EnvoyExtensionPolicyStatus  map[types.NamespacedName]bool
	ExtensionServerPolicyStatus map[message.NamespacedNameAndGVK]bool

	BackendStatus map[types.NamespacedName]bool
}

// copy creates a deep copy of the KeyCache for mark-and-sweep deletion
func (kc *KeyCache) copy() *KeyCache {
	copied := newKeyCache()

	// Copy IR keys
	for key := range kc.IR {
		copied.IR[key] = true
	}

	// Copy status keys
	for key := range kc.GatewayStatus {
		copied.GatewayStatus[key] = true
	}
	for key := range kc.XListenerSetStatus {
		copied.XListenerSetStatus[key] = true
	}
	for key := range kc.HTTPRouteStatus {
		copied.HTTPRouteStatus[key] = true
	}
	for key := range kc.GRPCRouteStatus {
		copied.GRPCRouteStatus[key] = true
	}
	for key := range kc.TLSRouteStatus {
		copied.TLSRouteStatus[key] = true
	}
	for key := range kc.TCPRouteStatus {
		copied.TCPRouteStatus[key] = true
	}
	for key := range kc.UDPRouteStatus {
		copied.UDPRouteStatus[key] = true
	}
	for key := range kc.BackendTLSPolicyStatus {
		copied.BackendTLSPolicyStatus[key] = true
	}
	for key := range kc.ClientTrafficPolicyStatus {
		copied.ClientTrafficPolicyStatus[key] = true
	}
	for key := range kc.BackendTrafficPolicyStatus {
		copied.BackendTrafficPolicyStatus[key] = true
	}
	for key := range kc.SecurityPolicyStatus {
		copied.SecurityPolicyStatus[key] = true
	}
	for key := range kc.EnvoyExtensionPolicyStatus {
		copied.EnvoyExtensionPolicyStatus[key] = true
	}
	for key := range kc.ExtensionServerPolicyStatus {
		copied.ExtensionServerPolicyStatus[key] = true
	}
	for key := range kc.BackendStatus {
		copied.BackendStatus[key] = true
	}

	return copied
}

func newKeyCache() *KeyCache {
	return &KeyCache{
		IR:                          make(map[string]bool),
		GatewayStatus:               make(map[types.NamespacedName]bool),
		XListenerSetStatus:          make(map[types.NamespacedName]bool),
		HTTPRouteStatus:             make(map[types.NamespacedName]bool),
		GRPCRouteStatus:             make(map[types.NamespacedName]bool),
		TLSRouteStatus:              make(map[types.NamespacedName]bool),
		TCPRouteStatus:              make(map[types.NamespacedName]bool),
		UDPRouteStatus:              make(map[types.NamespacedName]bool),
		BackendTLSPolicyStatus:      make(map[types.NamespacedName]bool),
		ClientTrafficPolicyStatus:   make(map[types.NamespacedName]bool),
		BackendTrafficPolicyStatus:  make(map[types.NamespacedName]bool),
		SecurityPolicyStatus:        make(map[types.NamespacedName]bool),
		EnvoyExtensionPolicyStatus:  make(map[types.NamespacedName]bool),
		ExtensionServerPolicyStatus: make(map[message.NamespacedNameAndGVK]bool),
		BackendStatus:               make(map[types.NamespacedName]bool),
	}
}

// populateKeyCache initializes the keyCache with existing keys from watchable stores
// This is needed for restart scenarios where stores may already contain data
func (r *Runner) populateKeyCache() {
	// Populate IR keys
	for key := range r.InfraIR.LoadAll() {
		r.keyCache.IR[key] = true
	}

	// Populate status keys
	for key := range r.ProviderResources.GatewayStatuses.LoadAll() {
		r.keyCache.GatewayStatus[key] = true
	}
	for key := range r.ProviderResources.XListenerSetStatuses.LoadAll() {
		r.keyCache.XListenerSetStatus[key] = true
	}
	for key := range r.ProviderResources.HTTPRouteStatuses.LoadAll() {
		r.keyCache.HTTPRouteStatus[key] = true
	}
	for key := range r.ProviderResources.GRPCRouteStatuses.LoadAll() {
		r.keyCache.GRPCRouteStatus[key] = true
	}
	for key := range r.ProviderResources.TLSRouteStatuses.LoadAll() {
		r.keyCache.TLSRouteStatus[key] = true
	}
	for key := range r.ProviderResources.TCPRouteStatuses.LoadAll() {
		r.keyCache.TCPRouteStatus[key] = true
	}
	for key := range r.ProviderResources.UDPRouteStatuses.LoadAll() {
		r.keyCache.UDPRouteStatus[key] = true
	}
	for key := range r.ProviderResources.BackendTLSPolicyStatuses.LoadAll() {
		r.keyCache.BackendTLSPolicyStatus[key] = true
	}
	for key := range r.ProviderResources.ClientTrafficPolicyStatuses.LoadAll() {
		r.keyCache.ClientTrafficPolicyStatus[key] = true
	}
	for key := range r.ProviderResources.BackendTrafficPolicyStatuses.LoadAll() {
		r.keyCache.BackendTrafficPolicyStatus[key] = true
	}
	for key := range r.ProviderResources.SecurityPolicyStatuses.LoadAll() {
		r.keyCache.SecurityPolicyStatus[key] = true
	}
	for key := range r.ProviderResources.EnvoyExtensionPolicyStatuses.LoadAll() {
		r.keyCache.EnvoyExtensionPolicyStatus[key] = true
	}
	for key := range r.ProviderResources.ExtensionPolicyStatuses.LoadAll() {
		r.keyCache.ExtensionServerPolicyStatus[key] = true
	}
	for key := range r.ProviderResources.BackendStatuses.LoadAll() {
		r.keyCache.BackendStatus[key] = true
	}
}

func (r *Runner) deleteKeys(kc *KeyCache) {
	// Delete IR keys
	for key := range kc.IR {
		r.InfraIR.Delete(key)
		r.XdsIR.Delete(key)
		delete(r.keyCache.IR, key)
	}

	// Delete status keys
	for key := range kc.GatewayStatus {
		r.ProviderResources.GatewayStatuses.Delete(key)
		delete(r.keyCache.GatewayStatus, key)
	}
	for key := range kc.XListenerSetStatus {
		r.ProviderResources.XListenerSetStatuses.Delete(key)
		delete(r.keyCache.XListenerSetStatus, key)
	}
	for key := range kc.HTTPRouteStatus {
		r.ProviderResources.HTTPRouteStatuses.Delete(key)
		delete(r.keyCache.HTTPRouteStatus, key)
	}
	for key := range kc.GRPCRouteStatus {
		r.ProviderResources.GRPCRouteStatuses.Delete(key)
		delete(r.keyCache.GRPCRouteStatus, key)
	}
	for key := range kc.TLSRouteStatus {
		r.ProviderResources.TLSRouteStatuses.Delete(key)
		delete(r.keyCache.TLSRouteStatus, key)
	}
	for key := range kc.TCPRouteStatus {
		r.ProviderResources.TCPRouteStatuses.Delete(key)
		delete(r.keyCache.TCPRouteStatus, key)
	}
	for key := range kc.UDPRouteStatus {
		r.ProviderResources.UDPRouteStatuses.Delete(key)
		delete(r.keyCache.UDPRouteStatus, key)
	}

	for key := range kc.ClientTrafficPolicyStatus {
		r.ProviderResources.ClientTrafficPolicyStatuses.Delete(key)
		delete(r.keyCache.ClientTrafficPolicyStatus, key)
	}
	for key := range kc.BackendTrafficPolicyStatus {
		r.ProviderResources.BackendTrafficPolicyStatuses.Delete(key)
		delete(r.keyCache.BackendTrafficPolicyStatus, key)
	}
	for key := range kc.SecurityPolicyStatus {
		r.ProviderResources.SecurityPolicyStatuses.Delete(key)
		delete(r.keyCache.SecurityPolicyStatus, key)
	}
	for key := range kc.BackendTLSPolicyStatus {
		r.ProviderResources.BackendTLSPolicyStatuses.Delete(key)
		delete(r.keyCache.BackendTLSPolicyStatus, key)
	}
	for key := range kc.EnvoyExtensionPolicyStatus {
		r.ProviderResources.EnvoyExtensionPolicyStatuses.Delete(key)
		delete(r.keyCache.EnvoyExtensionPolicyStatus, key)
	}
	for key := range kc.ExtensionServerPolicyStatus {
		r.ProviderResources.ExtensionPolicyStatuses.Delete(key)
		delete(r.keyCache.ExtensionServerPolicyStatus, key)
	}
	for key := range kc.BackendStatus {
		r.ProviderResources.BackendStatuses.Delete(key)
		delete(r.keyCache.BackendStatus, key)
	}
}

// hmac returns the HMAC secret generated by the CertGen job.
// hmac will be used as a hash salt to generate unguessable downloading paths for Wasm modules.
func hmac(ctx context.Context, namespace string) ([]byte, error) {
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
	return hmac, err
}
