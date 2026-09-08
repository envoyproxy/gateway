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
	"time"

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

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/crypto"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	extension "github.com/envoyproxy/gateway/internal/extension/types"
	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/gatewayapi/status"
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

// uid is the identity of the object the status was computed from; see storeStatus.
type aggregatedPolicyStatus struct {
	status     *gwapiv1.PolicyStatus
	generation int64
	uid        types.UID
}

type aggregatedRouteStatus struct {
	status     *gwapiv1.RouteStatus
	generation int64
	uid        types.UID
}

type aggregatedEnvoyProxyStatus struct {
	status *egv1a1.EnvoyProxyStatus
	uid    types.UID
}

func mergeAggregatedRouteStatus(aggregated aggregatedRouteStatus, incoming *gwapiv1.RouteStatus, generation int64, uid types.UID) aggregatedRouteStatus {
	// The same object is merged once per GatewayClass, so its identity is invariant
	// across merges and is recorded before any early return.
	aggregated.uid = uid
	if incoming == nil {
		return aggregated
	}

	if aggregated.status == nil {
		aggregated.status = incoming
		aggregated.generation = generation
		return aggregated
	}

	aggregated.status.Parents = append(aggregated.status.Parents, incoming.Parents...)
	if generation > aggregated.generation {
		aggregated.generation = generation
	}

	return aggregated
}

func mergePolicyStatus(aggregated aggregatedPolicyStatus, incoming *gwapiv1.PolicyStatus, generation int64, uid types.UID) aggregatedPolicyStatus {
	aggregated.uid = uid
	if incoming == nil {
		return aggregated
	}

	if aggregated.status == nil {
		aggregated.status = incoming
		aggregated.generation = generation
		return aggregated
	}

	// Prevent self-merge when aggregated and incoming reference the same object
	if aggregated.status != incoming {
		aggregated.status.Ancestors = append(aggregated.status.Ancestors, incoming.Ancestors...)
	}
	if generation > aggregated.generation {
		aggregated.generation = generation
	}

	return aggregated
}

func mergeEnvoyProxyStatus(aggregated, incoming *egv1a1.EnvoyProxyStatus) *egv1a1.EnvoyProxyStatus {
	if incoming == nil {
		return aggregated
	}

	if aggregated == nil {
		return incoming
	}

	// Prevent self-merge when aggregated and incoming reference the same object
	if aggregated != incoming {
		aggregated.Ancestors = append(aggregated.Ancestors, incoming.Ancestors...)
	}
	return aggregated
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
	if r.EnvoyGateway.Provider.IsRunningOnKubernetes() {
		cacheOption.CacheDir = "/var/lib/eg/wasm"
	} else {
		h, _ := os.UserHomeDir() // Assume we always get the home directory.
		cacheOption.CacheDir = path.Join(h, ".eg", "wasm")
	}
	// Create the file directory if it does not exist.
	if err = os.MkdirAll(cacheOption.CacheDir, 0o755); err != nil {
		r.Logger.Error(err, "Failed to create Wasm cache directory")
		return
	}
	r.wasmCache = wasm.NewHTTPServerWithFileCache(
		// HTTP server options
		wasm.ServerOptions{
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
			message.PublishRunnerEventMetric(r.Name(), update.Delete)

			parentCtx := update.Value.ParentContext(context.Background())
			var startOpts []trace.SpanStartOption
			if !update.Delete && !update.Initial {
				parentCtx, startOpts = message.RecordQueueWait(parentCtx, tracer, r.Name(), update.Value.StoredAtTime())
			}

			traceCtx, span := tracer.Start(parentCtx, "GatewayApiRunner.subscribeAndTranslate", startOpts...)
			defer span.End()
			traceLogger := r.Logger.WithTrace(traceCtx)
			traceLogger.Info("received an update", "key", update.Key)

			valWrapper := update.Value
			// There is only 1 key which is the controller name
			// so when a delete is triggered, delete all keys
			if update.Delete || valWrapper == nil || valWrapper.Resources == nil {
				// Clear EnvoyProxy statuses before deleting to remove stale ancestor conditions
				for key := range r.keyCache.EnvoyProxyStatus {
					emptyStatus := &egv1a1.EnvoyProxyStatus{
						Ancestors: []egv1a1.EnvoyProxyAncestorStatus{},
					}
					r.ProviderResources.EnvoyProxyStatuses.Store(key, emptyStatus)
				}
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
			var infraIRCount, xdsIRCount, gatewayStatusCount, listenerSetStatusCount, httpRouteStatusCount, grpcRouteStatusCount int
			var tlsRouteStatusCount, tcpRouteStatusCount, udpRouteStatusCount int
			var backendTLSPolicyStatusCount, clientTrafficPolicyStatusCount, backendTrafficPolicyStatusCount int
			var securityPolicyStatusCount, envoyExtensionPolicyStatusCount, backendStatusCount, extensionServerPolicyStatusCount int
			var envoyproxyStatusCount int

			// `aggregatedStatuses` aggregates status result of resources from all
			// parents/ancestors, and then stores the status once for every resource.
			aggregatedStatuses := struct {
				HTTPRoutes              map[types.NamespacedName]aggregatedRouteStatus
				GRPCRoutes              map[types.NamespacedName]aggregatedRouteStatus
				TLSRoutes               map[types.NamespacedName]aggregatedRouteStatus
				TCPRoutes               map[types.NamespacedName]aggregatedRouteStatus
				UDPRoutes               map[types.NamespacedName]aggregatedRouteStatus
				BackendTLSPolicies      map[types.NamespacedName]aggregatedPolicyStatus
				ClientTrafficPolicies   map[types.NamespacedName]aggregatedPolicyStatus
				BackendTrafficPolicies  map[types.NamespacedName]aggregatedPolicyStatus
				SecurityPolicies        map[types.NamespacedName]aggregatedPolicyStatus
				EnvoyExtensionPolicies  map[types.NamespacedName]aggregatedPolicyStatus
				ExtensionServerPolicies map[message.NamespacedNameAndGVK]aggregatedPolicyStatus
				EnvoyProxies            map[types.NamespacedName]aggregatedEnvoyProxyStatus
			}{
				HTTPRoutes:              make(map[types.NamespacedName]aggregatedRouteStatus),
				GRPCRoutes:              make(map[types.NamespacedName]aggregatedRouteStatus),
				TLSRoutes:               make(map[types.NamespacedName]aggregatedRouteStatus),
				TCPRoutes:               make(map[types.NamespacedName]aggregatedRouteStatus),
				UDPRoutes:               make(map[types.NamespacedName]aggregatedRouteStatus),
				BackendTLSPolicies:      make(map[types.NamespacedName]aggregatedPolicyStatus),
				ClientTrafficPolicies:   make(map[types.NamespacedName]aggregatedPolicyStatus),
				BackendTrafficPolicies:  make(map[types.NamespacedName]aggregatedPolicyStatus),
				SecurityPolicies:        make(map[types.NamespacedName]aggregatedPolicyStatus),
				EnvoyExtensionPolicies:  make(map[types.NamespacedName]aggregatedPolicyStatus),
				ExtensionServerPolicies: make(map[message.NamespacedNameAndGVK]aggregatedPolicyStatus),
				EnvoyProxies:            make(map[types.NamespacedName]aggregatedEnvoyProxyStatus),
			}

			span.AddEvent("translate", trace.WithAttributes(attribute.Int("resources.count", len(*val))))

			rtcTraceCtx, rtcSpan := tracer.Start(traceCtx, "GatewayApiRunner.ResoureTranslationCycle")
			defer rtcSpan.End()
			translateGatewayClass := func(resources *resource.Resources) {
				// The GatewayClass name is deliberately kept out of the span name and passed as
				// an attribute instead: span names are what most trace backends group/aggregate
				// by (latency percentiles, error rates, etc.), and this loop runs once per
				// GatewayClass in the cluster. Baking the name into the span name would fragment
				// those aggregates into one bucket per class - unbounded and growing over the
				// cluster's lifetime - for no benefit, since the attribute already makes the span
				// filterable/searchable by GatewayClass in any trace UI.
				translateGCCtx, translateGCSpan := tracer.Start(rtcTraceCtx, "GatewayApiRunner.ResoureTranslationCycle.TranslateGatewayClass",
					trace.WithAttributes(attribute.String("gatewayclass.name", resources.GatewayClass.Name)),
				)
				defer translateGCSpan.End()
				// Translate and publish IRs.
				t := &gatewayapi.Translator{
					GatewayControllerName:           r.EnvoyGateway.Gateway.ControllerName,
					GatewayClassName:                gwapiv1.ObjectName(resources.GatewayClass.Name),
					GlobalRateLimitEnabled:          r.EnvoyGateway.RateLimit != nil,
					EnvoyPatchPolicyEnabled:         r.EnvoyGateway.ExtensionAPIs != nil && r.EnvoyGateway.ExtensionAPIs.EnableEnvoyPatchPolicy,
					BackendEnabled:                  r.EnvoyGateway.ExtensionAPIs != nil && r.EnvoyGateway.ExtensionAPIs.EnableBackend,
					SDSSecretRefEnabled:             r.EnvoyGateway.ExtensionAPIs != nil && r.EnvoyGateway.ExtensionAPIs.EnableSDSSecretRef,
					ControllerNamespace:             r.ControllerNamespace,
					DNSDomain:                       r.DNSDomain,
					GatewayNamespaceMode:            r.EnvoyGateway.GatewayNamespaceMode(),
					MergeGateways:                   gatewayapi.IsMergeGatewaysEnabled(resources),
					MergeBackends:                   gatewayapi.ResolveMergeBackendsConfig(resources),
					PerResourceSystemCASecret:       r.EnvoyGateway.RuntimeFlags.IsEnabled(egv1a1.PerResourceSystemCASecret),
					WasmCache:                       r.wasmCache,
					RunningOnHost:                   r.EnvoyGateway.Provider != nil && r.EnvoyGateway.Provider.IsRunningOnHost(),
					InfraRemotelyManaged:            r.EnvoyGateway.Provider != nil && r.EnvoyGateway.Provider.IsInfraManagedRemotely(),
					Logger:                          traceLogger,
					LuaEnvoyExtensionPolicyDisabled: r.EnvoyGateway.ExtensionAPIs.LuaDisabled(),
				}

				// If extensions are loaded, pass their supported groups/kinds to the translator
				if extensions := r.EnvoyGateway.GetExtensionManagers(); len(extensions) > 0 {
					var extGKs []schema.GroupKind
					for _, em := range extensions {
						for _, gvk := range em.Resources {
							extGKs = append(extGKs, schema.GroupKind{Group: gvk.Group, Kind: gvk.Kind})
						}
						// Include backend resources in extension group kinds for custom backend support
						for _, gvk := range em.BackendResources {
							extGKs = append(extGKs, schema.GroupKind{Group: gvk.Group, Kind: gvk.Kind})
						}
					}
					t.ExtensionGroupKinds = extGKs
					traceLogger.Info("extension resources", "GVKs count", len(extGKs))
				}
				// Translate to IR.
				// The span is ended by a deferred call inside the closure: the translator
				// panics on some inputs and HandleSubscription recovers from it, and an
				// unended span is never exported, so a plain End() here would drop the
				// stage span and the input sizes recorded on it. The closure keeps that
				// defer scoped to this iteration instead of the whole update.
				result, err := func() (*gatewayapi.TranslateResult, error) {
					translateToIRCtx, translateToIRSpan := tracer.Start(translateGCCtx, "GatewayApiRunner.ResoureTranslationCycle.TranslateToIR")
					defer translateToIRSpan.End()
					return t.Translate(translateToIRCtx, resources)
				}()
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
						r.InfraIR.Store(key, &message.InfraIRWithContext{
							Infra:    val,
							Context:  translateGCCtx,
							StoredAt: time.Now(),
						})
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
							XdsIR:    val,
							Context:  translateGCCtx,
							StoredAt: time.Now(),
						}
						r.XdsIR.Store(key, &m)
						xdsIRCount++
					}
				}

				// Update Status
				_, statusUpdateSpan := tracer.Start(translateGCCtx, "GatewayApiRunner.ResoureTranslationCycle.UpdateStatus")
				defer statusUpdateSpan.End()
				if result.GatewayClass != nil {
					key := utils.NamespacedName(result.GatewayClass)
					r.ProviderResources.GatewayClassStatuses.Store(key, &result.GatewayClass.Status)
				}

				// Resources which can only belong to 1 GatewayClass (at most) get their statuses stored right away.
				for _, gateway := range result.Gateways {
					key := utils.NamespacedName(gateway)
					storeStatus(&r.ProviderResources.GatewayStatuses, r.keyCache.GatewayStatus, key, gateway.UID, &gateway.Status)
					gatewayStatusCount++
					delete(keysToDelete.GatewayStatus, key)
				}
				for _, listenerSet := range result.ListenerSets {
					key := utils.NamespacedName(listenerSet)
					storeStatus(&r.ProviderResources.ListenerSetStatuses, r.keyCache.ListenerSetStatus, key, listenerSet.UID, &listenerSet.Status)
					listenerSetStatusCount++
					delete(keysToDelete.ListenerSetStatus, key)
				}

				// Backend statuses have no parents, so they are not aggregated.
				for _, backend := range result.Backends {
					key := utils.NamespacedName(backend)
					if len(backend.Status.Conditions) > 0 {
						storeStatus(&r.ProviderResources.BackendStatuses, r.keyCache.BackendStatus, key, backend.UID, &backend.Status)
						backendStatusCount++
					}
					delete(keysToDelete.BackendStatus, key)
				}

				// Resources which can belong to multiple GatewayClasses get their statuses aggregated,
				// then stored once after iterating over all GatewayClasses.
				for _, httpRoute := range result.HTTPRoutes {
					if len(httpRoute.Status.Parents) != 0 {
						key := utils.NamespacedName(httpRoute)
						aggregatedStatuses.HTTPRoutes[key] = mergeAggregatedRouteStatus(aggregatedStatuses.HTTPRoutes[key], &httpRoute.Status.RouteStatus, httpRoute.Generation, httpRoute.UID)
					}
				}
				for _, grpcRoute := range result.GRPCRoutes {
					if len(grpcRoute.Status.Parents) != 0 {
						key := utils.NamespacedName(grpcRoute)
						aggregatedStatuses.GRPCRoutes[key] = mergeAggregatedRouteStatus(aggregatedStatuses.GRPCRoutes[key], &grpcRoute.Status.RouteStatus, grpcRoute.Generation, grpcRoute.UID)
					}
				}
				for _, tlsRoute := range result.TLSRoutes {
					if len(tlsRoute.Status.Parents) != 0 {
						key := utils.NamespacedName(tlsRoute)
						aggregatedStatuses.TLSRoutes[key] = mergeAggregatedRouteStatus(aggregatedStatuses.TLSRoutes[key], &tlsRoute.Status.RouteStatus, tlsRoute.Generation, tlsRoute.UID)
					}
				}
				for _, tcpRoute := range result.TCPRoutes {
					if len(tcpRoute.Status.Parents) != 0 {
						key := utils.NamespacedName(tcpRoute)
						aggregatedStatuses.TCPRoutes[key] = mergeAggregatedRouteStatus(aggregatedStatuses.TCPRoutes[key], &tcpRoute.Status.RouteStatus, tcpRoute.Generation, tcpRoute.UID)
					}
				}
				for _, udpRoute := range result.UDPRoutes {
					if len(udpRoute.Status.Parents) != 0 {
						key := utils.NamespacedName(udpRoute)
						aggregatedStatuses.UDPRoutes[key] = mergeAggregatedRouteStatus(aggregatedStatuses.UDPRoutes[key], &udpRoute.Status.RouteStatus, udpRoute.Generation, udpRoute.UID)
					}
				}
				for _, backendTLSPolicy := range result.BackendTLSPolicies {
					if len(backendTLSPolicy.Status.Ancestors) != 0 {
						key := utils.NamespacedName(backendTLSPolicy)
						aggregatedStatuses.BackendTLSPolicies[key] = mergePolicyStatus(aggregatedStatuses.BackendTLSPolicies[key], &backendTLSPolicy.Status, backendTLSPolicy.Generation, backendTLSPolicy.UID)
					}
				}
				for _, clientTrafficPolicy := range result.ClientTrafficPolicies {
					if len(clientTrafficPolicy.Status.Ancestors) != 0 {
						key := utils.NamespacedName(clientTrafficPolicy)
						aggregatedStatuses.ClientTrafficPolicies[key] = mergePolicyStatus(aggregatedStatuses.ClientTrafficPolicies[key], &clientTrafficPolicy.Status, clientTrafficPolicy.Generation, clientTrafficPolicy.UID)
					}
				}
				for _, backendTrafficPolicy := range result.BackendTrafficPolicies {
					if len(backendTrafficPolicy.Status.Ancestors) != 0 {
						key := utils.NamespacedName(backendTrafficPolicy)
						aggregatedStatuses.BackendTrafficPolicies[key] = mergePolicyStatus(aggregatedStatuses.BackendTrafficPolicies[key], &backendTrafficPolicy.Status, backendTrafficPolicy.Generation, backendTrafficPolicy.UID)
					}
				}
				for _, securityPolicy := range result.SecurityPolicies {
					if len(securityPolicy.Status.Ancestors) != 0 {
						key := utils.NamespacedName(securityPolicy)
						aggregatedStatuses.SecurityPolicies[key] = mergePolicyStatus(aggregatedStatuses.SecurityPolicies[key], &securityPolicy.Status, securityPolicy.Generation, securityPolicy.UID)
					}
				}
				for _, envoyExtensionPolicy := range result.EnvoyExtensionPolicies {
					if len(envoyExtensionPolicy.Status.Ancestors) != 0 {
						key := utils.NamespacedName(envoyExtensionPolicy)
						aggregatedStatuses.EnvoyExtensionPolicies[key] = mergePolicyStatus(aggregatedStatuses.EnvoyExtensionPolicies[key], &envoyExtensionPolicy.Status, envoyExtensionPolicy.Generation, envoyExtensionPolicy.UID)
					}
				}
				for _, extServerPolicy := range result.ExtensionServerPolicies {
					policyStatus := gatewayapi.ExtServerPolicyStatusAsPolicyStatus(&extServerPolicy)
					if len(policyStatus.Ancestors) != 0 {
						key := message.NamespacedNameAndGVK{
							NamespacedName:   utils.NamespacedName(&extServerPolicy),
							GroupVersionKind: extServerPolicy.GroupVersionKind(),
						}
						aggregatedStatuses.ExtensionServerPolicies[key] = mergePolicyStatus(aggregatedStatuses.ExtensionServerPolicies[key], &policyStatus, extServerPolicy.GetGeneration(), extServerPolicy.GetUID())
					}
				}
				// EnvoyProxy status
				for _, ep := range result.EnvoyProxiesForGateways {
					key := utils.NamespacedName(ep)
					r.Logger.Info("update envoyproxy status", "key", key)
					aggregatedStatuses.EnvoyProxies[key] = aggregatedEnvoyProxyStatus{
						status: mergeEnvoyProxyStatus(aggregatedStatuses.EnvoyProxies[key].status, &ep.Status),
						uid:    ep.UID,
					}
				}
				if ep := result.EnvoyProxyForGatewayClass; ep != nil {
					key := utils.NamespacedName(ep)
					r.Logger.Info("update envoyproxy status", "key", key)
					aggregatedStatuses.EnvoyProxies[key] = aggregatedEnvoyProxyStatus{
						status: mergeEnvoyProxyStatus(aggregatedStatuses.EnvoyProxies[key].status, &ep.Status),
						uid:    ep.UID,
					}
				}
			}
			for _, resources := range *val {
				translateGatewayClass(resources)
			}

			// Store the stauses of all objects atomically with the aggregated status.
			for key, entry := range aggregatedStatuses.HTTPRoutes {
				status.TruncateRouteParents(entry.status, entry.generation)
				s := gwapiv1.HTTPRouteStatus{RouteStatus: *entry.status}
				storeStatus(&r.ProviderResources.HTTPRouteStatuses, r.keyCache.HTTPRouteStatus, key, entry.uid, &s)
				httpRouteStatusCount++
				delete(keysToDelete.HTTPRouteStatus, key)
			}
			for key, entry := range aggregatedStatuses.GRPCRoutes {
				status.TruncateRouteParents(entry.status, entry.generation)
				s := gwapiv1.GRPCRouteStatus{RouteStatus: *entry.status}
				storeStatus(&r.ProviderResources.GRPCRouteStatuses, r.keyCache.GRPCRouteStatus, key, entry.uid, &s)
				grpcRouteStatusCount++
				delete(keysToDelete.GRPCRouteStatus, key)
			}
			for key, entry := range aggregatedStatuses.TLSRoutes {
				status.TruncateRouteParents(entry.status, entry.generation)
				s := gwapiv1.TLSRouteStatus{RouteStatus: *entry.status}
				storeStatus(&r.ProviderResources.TLSRouteStatuses, r.keyCache.TLSRouteStatus, key, entry.uid, &s)
				tlsRouteStatusCount++
				delete(keysToDelete.TLSRouteStatus, key)
			}
			for key, entry := range aggregatedStatuses.TCPRoutes {
				status.TruncateRouteParents(entry.status, entry.generation)
				s := gwapiv1.TCPRouteStatus{RouteStatus: *entry.status}
				storeStatus(&r.ProviderResources.TCPRouteStatuses, r.keyCache.TCPRouteStatus, key, entry.uid, &s)
				tcpRouteStatusCount++
				delete(keysToDelete.TCPRouteStatus, key)
			}
			for key, entry := range aggregatedStatuses.UDPRoutes {
				status.TruncateRouteParents(entry.status, entry.generation)
				s := gwapiv1.UDPRouteStatus{RouteStatus: *entry.status}
				storeStatus(&r.ProviderResources.UDPRouteStatuses, r.keyCache.UDPRouteStatus, key, entry.uid, &s)
				udpRouteStatusCount++
				delete(keysToDelete.UDPRouteStatus, key)
			}
			for key, entry := range aggregatedStatuses.BackendTLSPolicies {
				status.TruncatePolicyAncestors(entry.status, r.EnvoyGateway.Gateway.ControllerName, entry.generation)
				storeStatus(&r.ProviderResources.BackendTLSPolicyStatuses, r.keyCache.BackendTLSPolicyStatus, key, entry.uid, entry.status)
				backendTLSPolicyStatusCount++
				delete(keysToDelete.BackendTLSPolicyStatus, key)
			}
			for key, entry := range aggregatedStatuses.ClientTrafficPolicies {
				status.TruncatePolicyAncestors(entry.status, r.EnvoyGateway.Gateway.ControllerName, entry.generation)
				storeStatus(&r.ProviderResources.ClientTrafficPolicyStatuses, r.keyCache.ClientTrafficPolicyStatus, key, entry.uid, entry.status)
				clientTrafficPolicyStatusCount++
				delete(keysToDelete.ClientTrafficPolicyStatus, key)
			}
			for key, entry := range aggregatedStatuses.BackendTrafficPolicies {
				status.TruncatePolicyAncestors(entry.status, r.EnvoyGateway.Gateway.ControllerName, entry.generation)
				storeStatus(&r.ProviderResources.BackendTrafficPolicyStatuses, r.keyCache.BackendTrafficPolicyStatus, key, entry.uid, entry.status)
				backendTrafficPolicyStatusCount++
				delete(keysToDelete.BackendTrafficPolicyStatus, key)
			}
			for key, entry := range aggregatedStatuses.SecurityPolicies {
				status.TruncatePolicyAncestors(entry.status, r.EnvoyGateway.Gateway.ControllerName, entry.generation)
				storeStatus(&r.ProviderResources.SecurityPolicyStatuses, r.keyCache.SecurityPolicyStatus, key, entry.uid, entry.status)
				securityPolicyStatusCount++
				delete(keysToDelete.SecurityPolicyStatus, key)
			}
			for key, entry := range aggregatedStatuses.EnvoyExtensionPolicies {
				status.TruncatePolicyAncestors(entry.status, r.EnvoyGateway.Gateway.ControllerName, entry.generation)
				storeStatus(&r.ProviderResources.EnvoyExtensionPolicyStatuses, r.keyCache.EnvoyExtensionPolicyStatus, key, entry.uid, entry.status)
				envoyExtensionPolicyStatusCount++
				delete(keysToDelete.EnvoyExtensionPolicyStatus, key)
			}
			for key, entry := range aggregatedStatuses.ExtensionServerPolicies {
				status.TruncatePolicyAncestors(entry.status, r.EnvoyGateway.Gateway.ControllerName, entry.generation)
				storeStatus(&r.ProviderResources.ExtensionPolicyStatuses, r.keyCache.ExtensionServerPolicyStatus, key, entry.uid, entry.status)
				extensionServerPolicyStatusCount++
				delete(keysToDelete.ExtensionServerPolicyStatus, key)
			}
			for key, entry := range aggregatedStatuses.EnvoyProxies {
				s := egv1a1.EnvoyProxyStatus{
					Ancestors: entry.status.Ancestors,
				}
				storeStatus(&r.ProviderResources.EnvoyProxyStatuses, r.keyCache.EnvoyProxyStatus, key, entry.uid, &s)
				envoyproxyStatusCount++
				delete(keysToDelete.EnvoyProxyStatus, key)
			}
			// Publish aggregated metrics
			message.PublishMetric(message.Metadata{Runner: r.Name(), Message: message.InfraIRMessageName}, infraIRCount)
			message.PublishMetric(message.Metadata{Runner: r.Name(), Message: message.XDSIRMessageName}, xdsIRCount)
			message.PublishMetric(message.Metadata{Runner: r.Name(), Message: message.GatewayClassStatusMessageName}, 1)
			message.PublishMetric(message.Metadata{Runner: r.Name(), Message: message.GatewayStatusMessageName}, gatewayStatusCount)
			message.PublishMetric(message.Metadata{Runner: r.Name(), Message: message.ListenerSetStatusMessageName}, listenerSetStatusCount)
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
			message.PublishMetric(message.Metadata{Runner: r.Name(), Message: message.EnvoyProxyStatusMessageName}, envoyproxyStatusCount)

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
	for key := range r.keyCache.ListenerSetStatus {
		r.ProviderResources.ListenerSetStatuses.Delete(key)
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
	for key := range r.keyCache.EnvoyProxyStatus {
		r.ProviderResources.EnvoyProxyStatuses.Delete(key)
	}

	// Clear all tracking
	r.keyCache = newKeyCache()
}

type KeyCache struct {
	// IR keys
	IR map[string]bool

	// Status keys, mapped to the UID of the object the published status was computed
	// from. The UID lets storeStatus tell "nothing changed" apart from "the object
	// behind this name was replaced"; see its doc comment.
	GatewayStatus          map[types.NamespacedName]types.UID
	ListenerSetStatus      map[types.NamespacedName]types.UID
	HTTPRouteStatus        map[types.NamespacedName]types.UID
	GRPCRouteStatus        map[types.NamespacedName]types.UID
	TLSRouteStatus         map[types.NamespacedName]types.UID
	TCPRouteStatus         map[types.NamespacedName]types.UID
	UDPRouteStatus         map[types.NamespacedName]types.UID
	BackendTLSPolicyStatus map[types.NamespacedName]types.UID

	ClientTrafficPolicyStatus   map[types.NamespacedName]types.UID
	BackendTrafficPolicyStatus  map[types.NamespacedName]types.UID
	SecurityPolicyStatus        map[types.NamespacedName]types.UID
	EnvoyExtensionPolicyStatus  map[types.NamespacedName]types.UID
	ExtensionServerPolicyStatus map[message.NamespacedNameAndGVK]types.UID
	EnvoyProxyStatus            map[types.NamespacedName]types.UID

	BackendStatus map[types.NamespacedName]types.UID
}

// copy creates a deep copy of the KeyCache for mark-and-sweep deletion.
// Only the keys are carried over: the copy is the sweep set, and deleteKeys reads
// keys alone.
func (kc *KeyCache) copy() *KeyCache {
	copied := newKeyCache()

	// Copy IR keys
	for key := range kc.IR {
		copied.IR[key] = true
	}

	// Copy status keys
	for key := range kc.GatewayStatus {
		copied.GatewayStatus[key] = ""
	}
	for key := range kc.ListenerSetStatus {
		copied.ListenerSetStatus[key] = ""
	}
	for key := range kc.HTTPRouteStatus {
		copied.HTTPRouteStatus[key] = ""
	}
	for key := range kc.GRPCRouteStatus {
		copied.GRPCRouteStatus[key] = ""
	}
	for key := range kc.TLSRouteStatus {
		copied.TLSRouteStatus[key] = ""
	}
	for key := range kc.TCPRouteStatus {
		copied.TCPRouteStatus[key] = ""
	}
	for key := range kc.UDPRouteStatus {
		copied.UDPRouteStatus[key] = ""
	}
	for key := range kc.BackendTLSPolicyStatus {
		copied.BackendTLSPolicyStatus[key] = ""
	}
	for key := range kc.ClientTrafficPolicyStatus {
		copied.ClientTrafficPolicyStatus[key] = ""
	}
	for key := range kc.BackendTrafficPolicyStatus {
		copied.BackendTrafficPolicyStatus[key] = ""
	}
	for key := range kc.SecurityPolicyStatus {
		copied.SecurityPolicyStatus[key] = ""
	}
	for key := range kc.EnvoyExtensionPolicyStatus {
		copied.EnvoyExtensionPolicyStatus[key] = ""
	}
	for key := range kc.ExtensionServerPolicyStatus {
		copied.ExtensionServerPolicyStatus[key] = ""
	}
	for key := range kc.EnvoyProxyStatus {
		copied.EnvoyProxyStatus[key] = ""
	}
	for key := range kc.BackendStatus {
		copied.BackendStatus[key] = ""
	}

	return copied
}

// storeStatus publishes status under key and remembers which object it came from.
//
// The translated status is a pure function of spec and generation, so an object that is
// deleted and recreated under the same name produces an identical value, which the
// watchable subscriber drops as unchanged. Deleting the key first lets the store reach
// subscribers, so the recreated object does not keep the default status from its CRD.
// See issue #9536.
//
// A key absent from seen was never published, not replaced. An empty uid, which the
// file provider produces, matches itself, so replacement is not detected there.
func storeStatus[K comparable, V any](
	m *watchable.Map[K, V],
	seen map[K]types.UID,
	key K,
	uid types.UID,
	status V,
) {
	if prev, ok := seen[key]; ok && prev != uid {
		m.Delete(key)
	}
	m.Store(key, status)
	seen[key] = uid
}

func newKeyCache() *KeyCache {
	return &KeyCache{
		IR:                          make(map[string]bool),
		GatewayStatus:               make(map[types.NamespacedName]types.UID),
		ListenerSetStatus:           make(map[types.NamespacedName]types.UID),
		HTTPRouteStatus:             make(map[types.NamespacedName]types.UID),
		GRPCRouteStatus:             make(map[types.NamespacedName]types.UID),
		TLSRouteStatus:              make(map[types.NamespacedName]types.UID),
		TCPRouteStatus:              make(map[types.NamespacedName]types.UID),
		UDPRouteStatus:              make(map[types.NamespacedName]types.UID),
		BackendTLSPolicyStatus:      make(map[types.NamespacedName]types.UID),
		ClientTrafficPolicyStatus:   make(map[types.NamespacedName]types.UID),
		BackendTrafficPolicyStatus:  make(map[types.NamespacedName]types.UID),
		SecurityPolicyStatus:        make(map[types.NamespacedName]types.UID),
		EnvoyExtensionPolicyStatus:  make(map[types.NamespacedName]types.UID),
		ExtensionServerPolicyStatus: make(map[message.NamespacedNameAndGVK]types.UID),
		EnvoyProxyStatus:            make(map[types.NamespacedName]types.UID),
		BackendStatus:               make(map[types.NamespacedName]types.UID),
	}
}

// populateKeyCache initializes the keyCache with existing keys from watchable stores
// This is needed for restart scenarios where stores may already contain data
//
// Status keys are seeded with an empty UID, meaning "identity unknown": the store holds
// a status but this process has not seen the object it came from, so storeStatus treats
// the next publish as a plain store rather than as a replacement.
func (r *Runner) populateKeyCache() {
	// Populate IR keys
	for key := range r.InfraIR.LoadAll() {
		r.keyCache.IR[key] = true
	}

	// Populate status keys
	for key := range r.ProviderResources.GatewayStatuses.LoadAll() {
		r.keyCache.GatewayStatus[key] = ""
	}
	for key := range r.ProviderResources.ListenerSetStatuses.LoadAll() {
		r.keyCache.ListenerSetStatus[key] = ""
	}
	for key := range r.ProviderResources.HTTPRouteStatuses.LoadAll() {
		r.keyCache.HTTPRouteStatus[key] = ""
	}
	for key := range r.ProviderResources.GRPCRouteStatuses.LoadAll() {
		r.keyCache.GRPCRouteStatus[key] = ""
	}
	for key := range r.ProviderResources.TLSRouteStatuses.LoadAll() {
		r.keyCache.TLSRouteStatus[key] = ""
	}
	for key := range r.ProviderResources.TCPRouteStatuses.LoadAll() {
		r.keyCache.TCPRouteStatus[key] = ""
	}
	for key := range r.ProviderResources.UDPRouteStatuses.LoadAll() {
		r.keyCache.UDPRouteStatus[key] = ""
	}
	for key := range r.ProviderResources.BackendTLSPolicyStatuses.LoadAll() {
		r.keyCache.BackendTLSPolicyStatus[key] = ""
	}
	for key := range r.ProviderResources.ClientTrafficPolicyStatuses.LoadAll() {
		r.keyCache.ClientTrafficPolicyStatus[key] = ""
	}
	for key := range r.ProviderResources.BackendTrafficPolicyStatuses.LoadAll() {
		r.keyCache.BackendTrafficPolicyStatus[key] = ""
	}
	for key := range r.ProviderResources.SecurityPolicyStatuses.LoadAll() {
		r.keyCache.SecurityPolicyStatus[key] = ""
	}
	for key := range r.ProviderResources.EnvoyExtensionPolicyStatuses.LoadAll() {
		r.keyCache.EnvoyExtensionPolicyStatus[key] = ""
	}
	for key := range r.ProviderResources.ExtensionPolicyStatuses.LoadAll() {
		r.keyCache.ExtensionServerPolicyStatus[key] = ""
	}
	for key := range r.ProviderResources.EnvoyProxyStatuses.LoadAll() {
		r.keyCache.EnvoyProxyStatus[key] = ""
	}
	for key := range r.ProviderResources.BackendStatuses.LoadAll() {
		r.keyCache.BackendStatus[key] = ""
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
	for key := range kc.ListenerSetStatus {
		r.ProviderResources.ListenerSetStatuses.Delete(key)
		delete(r.keyCache.ListenerSetStatus, key)
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
	for key := range kc.EnvoyProxyStatus {
		r.ProviderResources.EnvoyProxyStatuses.Delete(key)
		delete(r.keyCache.EnvoyProxyStatus, key)
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
