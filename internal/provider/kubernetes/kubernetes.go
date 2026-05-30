// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"
	"fmt"
	"net/http"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	policyv1 "k8s.io/api/policy/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/config"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	gwapiv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway"
	ec "github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/infrastructure/kubernetes/proxy"
	"github.com/envoyproxy/gateway/internal/message"
)

// Provider is the scaffolding for the Kubernetes provider. It sets up dependencies
// and defines the topology of the provider and its managed components, wiring
// them together.
type Provider struct {
	client        client.Client
	manager       manager.Manager
	providerReady chan struct{}
}

const (
	QPS   = 50
	BURST = 100
)

// Exposed to allow disabling health probe listener in tests.
var (
	healthProbeBindAddress = ":8081"

	// webhookTLSCert is the filename within webhookTLSCertDir containing
	// the webhook server TLS certificate.
	webhookTLSCert = "tls.crt"
	// webhookTLSKey is the filename within webhookTLSCertDir containing
	// the webhook server TLS private key.
	webhookTLSKey = "tls.key"
	// webhookTLSCertDir is the directory container the webhook server
	// TLS certificate files.
	webhookTLSCertDir = "/certs"
	// webhookTLSPort is the port for the webhook server to listen on.
	webhookTLSPort = 9443
)

// cacheReadyCheck returns a healthz.Checker that verifies the manager's cache has synced.
// This ensures the control plane has populated its cache with all resources from the API server
// before reporting ready. This prevents serving inconsistent xDS configuration to Envoy proxies
// when running multiple control plane replicas during periods of resource churn.
func cacheReadyCheck(mgr manager.Manager) healthz.Checker {
	return func(req *http.Request) error {
		// Use a short timeout to avoid blocking the health check indefinitely.
		// The readiness probe will retry periodically until the cache syncs.
		ctx, cancel := context.WithTimeout(req.Context(), 1*time.Second)
		defer cancel()

		// WaitForCacheSync returns true if the cache has synced, false if the context is cancelled.
		if !mgr.GetCache().WaitForCacheSync(ctx) {
			return fmt.Errorf("cache not synced yet")
		}

		return nil
	}
}

func New(ctx context.Context, restCfg *rest.Config, svrCfg *ec.Server,
	resources *message.ProviderResources, errNotifier message.RunnerErrorNotifier,
) (*Provider, error) {
	return newProvider(ctx, restCfg, svrCfg, nil, resources, errNotifier)
}

// newProvider creates a new Provider from the provided EnvoyGateway.
func newProvider(ctx context.Context, restCfg *rest.Config, svrCfg *ec.Server,
	metricsOpts *metricsserver.Options,
	resources *message.ProviderResources, _ message.RunnerErrorNotifier,
) (*Provider, error) {
	// TODO: Decide which mgr opts should be exposed through envoygateway.provider.kubernetes API.
	mgrOpts := manager.Options{
		Scheme:                  envoygateway.GetScheme(),
		Logger:                  svrCfg.Logger.Logger,
		HealthProbeBindAddress:  healthProbeBindAddress,
		LeaderElectionID:        "5b9825d2.gateway.envoyproxy.io",
		LeaderElectionNamespace: svrCfg.ControllerNamespace,
		Client: client.Options{
			Cache: &client.CacheOptions{
				Unstructured: true,
			},
		},
	}

	if metricsOpts != nil {
		mgrOpts.Metrics = *metricsOpts
	}

	log.SetLogger(mgrOpts.Logger)
	klog.SetLogger(mgrOpts.Logger)

	restCfg.QPS, restCfg.Burst = svrCfg.EnvoyGateway.Provider.Kubernetes.Client.RateLimit.GetQPSAndBurst()

	if !ptr.Deref(svrCfg.EnvoyGateway.Provider.Kubernetes.LeaderElection.Disable, false) {
		mgrOpts.LeaderElection = true
		if svrCfg.EnvoyGateway.Provider.Kubernetes.LeaderElection.LeaseDuration != nil {
			ld, err := time.ParseDuration(string(*svrCfg.EnvoyGateway.Provider.Kubernetes.LeaderElection.LeaseDuration))
			if err != nil {
				return nil, err
			}
			mgrOpts.LeaseDuration = new(ld)
		}

		if svrCfg.EnvoyGateway.Provider.Kubernetes.LeaderElection.RetryPeriod != nil {
			rp, err := time.ParseDuration(string(*svrCfg.EnvoyGateway.Provider.Kubernetes.LeaderElection.RetryPeriod))
			if err != nil {
				return nil, err
			}
			mgrOpts.RetryPeriod = new(rp)
		}

		if svrCfg.EnvoyGateway.Provider.Kubernetes.LeaderElection.RenewDeadline != nil {
			rd, err := time.ParseDuration(string(*svrCfg.EnvoyGateway.Provider.Kubernetes.LeaderElection.RenewDeadline))
			if err != nil {
				return nil, err
			}
			mgrOpts.RenewDeadline = new(rd)
		}
		mgrOpts.Controller = config.Controller{NeedLeaderElection: new(false)}
	}

	if svrCfg.EnvoyGateway.Provider.Kubernetes.CacheSyncPeriod != nil {
		csp, err := time.ParseDuration(string(*svrCfg.EnvoyGateway.Provider.Kubernetes.CacheSyncPeriod))
		if err != nil {
			return nil, err
		}
		mgrOpts.Cache.SyncPeriod = new(csp)
	}

	// Disable deepcopy for some read only resources to reduce CPU and memory usage.
	// These resources are not modified by the provider, so it is safe to skip deepcopy.
	// If any of these resources need to be modified in the future, deepcopy should be re-enabled for that resource.
	if mgrOpts.Cache.ByObject == nil {
		mgrOpts.Cache.ByObject = map[client.Object]cache.ByObject{
			&corev1.Secret{}: {
				UnsafeDisableDeepCopy: new(true),
			},
			&corev1.ConfigMap{}: {
				UnsafeDisableDeepCopy: new(true),
				Transform:             composeTransforms(cache.TransformStripManagedFields(), transformConfigMapData),
			},
			&corev1.Service{}: {
				UnsafeDisableDeepCopy: new(true),
			},
			&discoveryv1.EndpointSlice{}: {
				UnsafeDisableDeepCopy: new(true),
			},
			&corev1.Node{}: {
				UnsafeDisableDeepCopy: new(true),
			},
			&gwapiv1b1.ReferenceGrant{}: {
				UnsafeDisableDeepCopy: new(true),
			},
		}
	}

	// Limit the cache to only Envoy proxy Pods to reduce memory and sync churn.
	// ProxyTopologyInjector is the only component that interacts with Pods.
	mgrOpts.Cache.ByObject[&corev1.Pod{}] = cache.ByObject{
		UnsafeDisableDeepCopy: new(true),
		Label:                 labels.SelectorFromSet(proxy.EnvoyAppLabel()),
	}

	namesReq, err := labels.NewRequirement("app.kubernetes.io/name", selection.In,
		[]string{"envoy", "envoy-ratelimit"})
	if err != nil {
		panic(err)
	}
	managedSelector := labels.NewSelector().Add(*namesReq)

	// If GatewayNamespaceMode is enabled, we need to watch all namespaces for the envoy proxy infrastructure resources.
	// If not, we only watch the controller namespace to avoid unnecessary RBAC permissions.
	// A label selector is still applied in both cases to limit the cache to only the resources owned by EG to reduce memory and sync churn.
	if svrCfg.EnvoyGateway.GatewayNamespaceMode() {
		// Keep ServiceAccount/Deployment unfiltered because the Envoy Gateway controller service account and deployment
		// are needed to watch for changes, and EG controller's labels can be customized by users while installation
		// and may not be present in the cache.
		mgrOpts.Cache.ByObject[&corev1.ServiceAccount{}] = cache.ByObject{
			UnsafeDisableDeepCopy: new(true),
		}
		mgrOpts.Cache.ByObject[&appsv1.Deployment{}] = cache.ByObject{
			UnsafeDisableDeepCopy: new(true),
		}
		// Filtering these envoy proxy and ratelimit infra resources by labels to reduce the cache size and memory usage.
		mgrOpts.Cache.ByObject[&appsv1.DaemonSet{}] = cache.ByObject{
			UnsafeDisableDeepCopy: new(true),
			Label:                 managedSelector,
		}
		mgrOpts.Cache.ByObject[&autoscalingv2.HorizontalPodAutoscaler{}] = cache.ByObject{
			UnsafeDisableDeepCopy: new(true),
			Label:                 managedSelector,
		}
		mgrOpts.Cache.ByObject[&policyv1.PodDisruptionBudget{}] = cache.ByObject{
			UnsafeDisableDeepCopy: new(true),
			Label:                 managedSelector,
		}
	} else {
		// Keep ServiceAccount/Deployment unfiltered because the Envoy Gateway controller service account and deployment
		// are needed to watch for changes, and EG controller's labels can be customized by users while installation
		// and may not be present in the cache.
		mgrOpts.Cache.ByObject[&corev1.ServiceAccount{}] = cache.ByObject{
			UnsafeDisableDeepCopy: new(true),
			Namespaces: map[string]cache.Config{
				svrCfg.ControllerNamespace: {},
			},
		}
		mgrOpts.Cache.ByObject[&appsv1.Deployment{}] = cache.ByObject{
			UnsafeDisableDeepCopy: new(true),
			Namespaces: map[string]cache.Config{
				svrCfg.ControllerNamespace: {},
			},
		}
		// Filtering these envoy proxy and ratelimit infra resources by labels to reduce the cache size and memory usage.
		mgrOpts.Cache.ByObject[&appsv1.DaemonSet{}] = cache.ByObject{
			UnsafeDisableDeepCopy: new(true),
			Label:                 managedSelector,
			Namespaces: map[string]cache.Config{
				svrCfg.ControllerNamespace: {},
			},
		}
		mgrOpts.Cache.ByObject[&autoscalingv2.HorizontalPodAutoscaler{}] = cache.ByObject{
			UnsafeDisableDeepCopy: new(true),
			Label:                 managedSelector,
			Namespaces: map[string]cache.Config{
				svrCfg.ControllerNamespace: {},
			},
		}
		mgrOpts.Cache.ByObject[&policyv1.PodDisruptionBudget{}] = cache.ByObject{
			UnsafeDisableDeepCopy: new(true),
			Label:                 managedSelector,
			Namespaces: map[string]cache.Config{
				svrCfg.ControllerNamespace: {},
			},
		}
	}

	mgrOpts.Cache.DefaultTransform = cache.TransformStripManagedFields()

	if svrCfg.EnvoyGateway.NamespaceMode() {
		mgrOpts.Cache.DefaultNamespaces = make(map[string]cache.Config)
		for _, watchNS := range svrCfg.EnvoyGateway.Provider.Kubernetes.Watch.Namespaces {
			mgrOpts.Cache.DefaultNamespaces[watchNS] = cache.Config{}
		}
	}
	if svrCfg.EnvoyGateway.Provider.Kubernetes.TopologyInjector == nil || !ptr.Deref(svrCfg.EnvoyGateway.Provider.Kubernetes.TopologyInjector.Disable, false) {
		mgrOpts.WebhookServer = webhook.NewServer(webhook.Options{
			CertDir:  webhookTLSCertDir,
			CertName: webhookTLSCert,
			KeyName:  webhookTLSKey,
			Port:     webhookTLSPort,
		})
	}

	mgr, err := ctrl.NewManager(restCfg, mgrOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to create manager: %w", err)
	}

	if svrCfg.EnvoyGateway.Provider.Kubernetes.TopologyInjector == nil || !ptr.Deref(svrCfg.EnvoyGateway.Provider.Kubernetes.TopologyInjector.Disable, false) {
		mgr.GetWebhookServer().Register("/inject-pod-topology", &webhook.Admission{
			Handler: &ProxyTopologyInjector{
				Client:    mgr.GetClient(),
				APIReader: mgr.GetAPIReader(),
				Logger:    svrCfg.Logger.WithName("proxy-topology-injector"),
				Decoder:   admission.NewDecoder(mgr.GetScheme()),
			},
		})
	}
	updateHandler := NewUpdateHandler(mgr.GetLogger(), mgr.GetClient())
	if err := mgr.Add(updateHandler); err != nil {
		return nil, fmt.Errorf("failed to add status update handler %w", err)
	}

	// Create and register the controllers with the manager.
	if err := newGatewayAPIController(ctx, mgr, svrCfg, updateHandler.Writer(), resources); err != nil {
		return nil, fmt.Errorf("failed to create gatewayapi controller: %w", err)
	}

	// Add health check health probes.
	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		return nil, fmt.Errorf("unable to set up health check: %w", err)
	}

	// Add ready check to wait for a successful sync of the cache.
	if err := mgr.AddReadyzCheck("cache-sync", cacheReadyCheck(mgr)); err != nil {
		return nil, fmt.Errorf("unable to set up ready check: %w", err)
	}

	// Emit elected & continue with the tasks that require leadership.
	go func() {
		<-mgr.Elected()
		// Close the elected channel to signal that this EG instance has been elected as leader.
		// This allows the tasks that require leadership to proceed.
		close(svrCfg.Elected)
	}()

	return &Provider{
		manager:       mgr,
		client:        mgr.GetClient(),
		providerReady: svrCfg.ProviderReady,
	}, nil
}

func (p *Provider) Type() egv1a1.ProviderType {
	return egv1a1.ProviderTypeKubernetes
}

// GetClient returns the controller-runtime client created by the Kubernetes provider.
func (p *Provider) GetClient() client.Client {
	return p.client
}

// Start starts the Provider synchronously until a message is received from ctx.
func (p *Provider) Start(ctx context.Context) error {
	errChan := make(chan error)
	go func() {
		errChan <- p.manager.Start(ctx)
	}()
	go signalProviderReady(ctx, p.manager.GetCache().WaitForCacheSync, p.providerReady)

	// Wait for the manager to exit or an explicit stop.
	select {
	case <-ctx.Done():
		return nil
	case err := <-errChan:
		return err
	}
}

func signalProviderReady(
	ctx context.Context,
	waitForCacheSync func(context.Context) bool,
	providerReady chan struct{},
) {
	if !waitForCacheSync(ctx) {
		return
	}

	select {
	case <-providerReady:
		return
	default:
		close(providerReady)
	}
}
