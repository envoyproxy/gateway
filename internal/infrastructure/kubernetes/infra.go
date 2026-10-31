// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/infrastructure/kubernetes/proxy"
	"github.com/envoyproxy/gateway/internal/infrastructure/kubernetes/ratelimit"
	"github.com/envoyproxy/gateway/internal/logging"
	"github.com/envoyproxy/gateway/internal/message"
)

var (
	_ ResourceRender = &proxy.ResourceRender{}
	_ ResourceRender = &ratelimit.ResourceRender{}
)

// ResourceRender renders Kubernetes infrastructure resources
// based on Infra IR resources.
type ResourceRender interface {
	Name() string
	Namespace() string
	LabelSelector() labels.Selector
	ServiceAccount() (*corev1.ServiceAccount, error)
	Service() (*corev1.Service, error)
	ConfigMap(cert string) (*corev1.ConfigMap, error)
	Deployment() (*appsv1.Deployment, error)
	DaemonSet() (*appsv1.DaemonSet, error)
	HorizontalPodAutoscaler() (*autoscalingv2.HorizontalPodAutoscaler, error)
	PodDisruptionBudget() (*policyv1.PodDisruptionBudget, error)
}

// Infra manages the creation and deletion of Kubernetes infrastructure
// based on Infra IR resources.
type Infra struct {
	// ControllerNamespace is the namespace where Envoy Gateway is deployed.
	ControllerNamespace string

	// ControllerName is the well-known name of the Envoy Gateway controller's own resources.
	ControllerName string

	// ControllerFullName is the Helm release fullname used for certgen and other release-scoped resources.
	ControllerFullName string

	// ControllerServiceAccountName is the service account name of the Envoy Gateway controller pod.
	ControllerServiceAccountName string

	// DNSDomain is the dns domain used by k8s services. Defaults to "cluster.local".
	DNSDomain string

	// EnvoyGateway is the configuration used to startup Envoy Gateway.
	EnvoyGateway *egv1a1.EnvoyGateway

	// Client wrap k8s client.
	Client *InfraClient

	logger logging.Logger

	// errors is the notifier used to send async errors to the main control loop.
	errors message.RunnerErrorNotifier

	// kubeServerVersion discovers the API server git version. The result is cached
	// in kubeVersion so proxy renders do not call discovery again.
	kubeServerVersion *config.KubernetesServerVersionHolder
	kubeVersionOnce   sync.Once
	kubeVersion       string
}

// kubernetesVersionDiscoveryTimeout bounds a single API server version lookup.
// A timeout, error, or empty result selects the legacy shutdown-manager layout
// and is cached so reconciliation is not blocked again.
var kubernetesVersionDiscoveryTimeout = 10 * time.Second

// NewInfra returns a new Infra.
func NewInfra(cli client.Client, cfg *config.Server, errors message.RunnerErrorNotifier) *Infra {
	return &Infra{
		// Always set infra namespace to cfg.ControllerNamespace,
		// Otherwise RateLimit resource provider will failed to create/delete.
		ControllerNamespace:          cfg.ControllerNamespace,
		ControllerName:               cfg.ControllerName,
		ControllerFullName:           cfg.ControllerFullName,
		ControllerServiceAccountName: cfg.ControllerServiceAccountName,
		DNSDomain:                    cfg.DNSDomain,
		EnvoyGateway:                 cfg.EnvoyGateway,
		Client:                       New(cli),
		logger:                       cfg.Logger.WithName(string(egv1a1.LogComponentInfrastructureRunner)),
		errors:                       errors,
		kubeServerVersion:            cfg.KubernetesServerVersion,
	}
}

// ResolveKubernetesVersion discovers the API server version at most once.
// Failures are cached as an empty version so callers keep the legacy layout.
func (i *Infra) ResolveKubernetesVersion(ctx context.Context) {
	i.kubeVersionOnce.Do(func() {
		i.kubeVersion = i.discoverKubernetesVersion(ctx)
	})
}

func (i *Infra) discoverKubernetesVersion(ctx context.Context) string {
	if i.kubeServerVersion == nil {
		return ""
	}
	discover := i.kubeServerVersion.Get()
	if discover == nil {
		return ""
	}

	ctx, cancel := context.WithTimeout(ctx, kubernetesVersionDiscoveryTimeout)
	defer cancel()
	version, err := callKubernetesVersionDiscover(ctx, discover)
	if err != nil || strings.TrimSpace(version) == "" {
		if err != nil {
			i.logger.Error(err, "failed to discover Kubernetes server version; using legacy shutdown-manager layout")
		} else {
			i.logger.Info("Kubernetes server version unavailable; using legacy shutdown-manager layout")
		}
		return ""
	}
	i.logger.Info("discovered Kubernetes server version", "version", version)
	return strings.TrimSpace(version)
}

// callKubernetesVersionDiscover runs discover until it returns or ctx is done, so a
// stuck client cannot block proxy reconciliation indefinitely.
func callKubernetesVersionDiscover(ctx context.Context, discover func(context.Context) (string, error)) (string, error) {
	type result struct {
		version string
		err     error
	}
	ch := make(chan result, 1)
	go func() {
		version, err := discover(ctx)
		ch <- result{version: version, err: err}
	}()
	select {
	case res := <-ch:
		return res.version, res.err
	case <-ctx.Done():
		select {
		case res := <-ch:
			return res.version, res.err
		default:
			return "", ctx.Err()
		}
	}
}

// Close implements Manager interface.
func (i *Infra) Close() error { return nil }

// createOrUpdate creates a ServiceAccount/ConfigMap/Deployment/Service in the kube api server based on the
// provided ResourceRender, if it doesn't exist and updates it if it does.
func (i *Infra) createOrUpdate(ctx context.Context, r ResourceRender) error {
	if err := i.createOrUpdateServiceAccount(ctx, r); err != nil {
		return fmt.Errorf("failed to create or update serviceaccount %s/%s: %w", r.Namespace(), r.Name(), err)
	}

	if err := i.createOrUpdateConfigMap(ctx, r); err != nil {
		return fmt.Errorf("failed to create or update configmap %s/%s: %w", r.Namespace(), r.Name(), err)
	}

	if err := i.createOrUpdateDeployment(ctx, r); err != nil {
		return fmt.Errorf("failed to create or update deployment %s/%s: %w", r.Namespace(), r.Name(), err)
	}

	if err := i.createOrUpdateDaemonSet(ctx, r); err != nil {
		return fmt.Errorf("failed to create or update daemonset %s/%s: %w", r.Namespace(), r.Name(), err)
	}

	if err := i.createOrUpdateService(ctx, r); err != nil {
		return fmt.Errorf("failed to create or update service %s/%s: %w", r.Namespace(), r.Name(), err)
	}

	if err := i.createOrUpdateHPA(ctx, r); err != nil {
		return fmt.Errorf("failed to create or update hpa %s/%s: %w", r.Namespace(), r.Name(), err)
	}

	if err := i.createOrUpdatePodDisruptionBudget(ctx, r); err != nil {
		return fmt.Errorf("failed to create or update pdb %s/%s: %w", r.Namespace(), r.Name(), err)
	}

	return nil
}

// delete deletes the ServiceAccount/ConfigMap/Deployment/Service in the kube api server, if it exists.
func (i *Infra) delete(ctx context.Context, r ResourceRender) error {
	if err := i.deleteServiceAccount(ctx, r); err != nil {
		return fmt.Errorf("failed to delete serviceaccount %s/%s: %w", r.Namespace(), r.Name(), err)
	}

	if err := i.deleteConfigMap(ctx, r); err != nil {
		return fmt.Errorf("failed to delete configmap %s/%s: %w", r.Namespace(), r.Name(), err)
	}

	if err := i.deleteDeployment(ctx, r); err != nil {
		return fmt.Errorf("failed to delete deployment %s/%s: %w", r.Namespace(), r.Name(), err)
	}

	if err := i.deleteDaemonSet(ctx, r); err != nil {
		return fmt.Errorf("failed to delete daemonset %s/%s: %w", r.Namespace(), r.Name(), err)
	}

	if err := i.deleteService(ctx, r); err != nil {
		return fmt.Errorf("failed to delete service %s/%s: %w", r.Namespace(), r.Name(), err)
	}

	if err := i.deleteHPA(ctx, r); err != nil {
		return fmt.Errorf("failed to delete hpa %s/%s: %w", r.Namespace(), r.Name(), err)
	}

	if err := i.deletePDB(ctx, r); err != nil {
		return fmt.Errorf("failed to delete pdb %s/%s: %w", r.Namespace(), r.Name(), err)
	}

	return nil
}
