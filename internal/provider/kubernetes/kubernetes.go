// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"
	"fmt"
	"time"

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
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway"
	ec "github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/message"
)

// Provider is the scaffolding for the Kubernetes provider. It sets up dependencies
// and defines the topology of the provider and its managed components, wiring
// them together.
type Provider struct {
	client  client.Client
	manager manager.Manager
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

// New creates a new Provider from the provided EnvoyGateway.
func New(ctx context.Context, restCfg *rest.Config, svrCfg *ec.Server, resources *message.ProviderResources) (*Provider, error) {
	// TODO: Decide which mgr opts should be exposed through envoygateway.provider.kubernetes API.

	mgrOpts := manager.Options{
		Scheme:                  envoygateway.GetScheme(),
		Logger:                  svrCfg.Logger.Logger,
		HealthProbeBindAddress:  healthProbeBindAddress,
		LeaderElectionID:        "5b9825d2.gateway.envoyproxy.io",
		LeaderElectionNamespace: svrCfg.ControllerNamespace,
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
			mgrOpts.LeaseDuration = ptr.To(ld)
		}

		if svrCfg.EnvoyGateway.Provider.Kubernetes.LeaderElection.RetryPeriod != nil {
			rp, err := time.ParseDuration(string(*svrCfg.EnvoyGateway.Provider.Kubernetes.LeaderElection.RetryPeriod))
			if err != nil {
				return nil, err
			}
			mgrOpts.RetryPeriod = ptr.To(rp)
		}

		if svrCfg.EnvoyGateway.Provider.Kubernetes.LeaderElection.RenewDeadline != nil {
			rd, err := time.ParseDuration(string(*svrCfg.EnvoyGateway.Provider.Kubernetes.LeaderElection.RenewDeadline))
			if err != nil {
				return nil, err
			}
			mgrOpts.RenewDeadline = ptr.To(rd)
		}
		mgrOpts.Controller = config.Controller{NeedLeaderElection: ptr.To(false)}
	}

	if svrCfg.EnvoyGateway.Provider.Kubernetes.CacheSyncPeriod != nil {
		csp, err := time.ParseDuration(string(*svrCfg.EnvoyGateway.Provider.Kubernetes.CacheSyncPeriod))
		if err != nil {
			return nil, err
		}
		mgrOpts.Cache.SyncPeriod = ptr.To(csp)
	}

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
				Client:  mgr.GetClient(),
				Logger:  svrCfg.Logger.WithName("proxy-topology-injector"),
				Decoder: admission.NewDecoder(mgr.GetScheme()),
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

	// Add ready check health probes.
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
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
		manager: mgr,
		client:  mgr.GetClient(),
	}, nil
}

func (p *Provider) Type() egv1a1.ProviderType {
	return egv1a1.ProviderTypeKubernetes
}

// Start starts the Provider synchronously until a message is received from ctx.
func (p *Provider) Start(ctx context.Context) error {
	errChan := make(chan error)
	go func() {
		errChan <- p.manager.Start(ctx)
	}()

	// Wait for the manager to exit or an explicit stop.
	select {
	case <-ctx.Done():
		return nil
	case err := <-errChan:
		return err
	}
}
