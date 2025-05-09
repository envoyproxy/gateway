// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"
	"fmt"

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
)

var _ ResourceRender = &proxy.ResourceRender{}

var _ ResourceRender = &ratelimit.ResourceRender{}

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

	// DNSDomain is the dns domain used by k8s services. Defaults to "cluster.local".
	DNSDomain string

	// EnvoyGateway is the configuration used to startup Envoy Gateway.
	EnvoyGateway *egv1a1.EnvoyGateway

	// Client wrap k8s client.
	Client *InfraClient

	logger logging.Logger
}

// NewInfra returns a new Infra.
func NewInfra(cli client.Client, cfg *config.Server) *Infra {
	return &Infra{
		// Always set infra namespace to cfg.ControllerNamespace,
		// Otherwise RateLimit resource provider will failed to create/delete.
		ControllerNamespace: cfg.ControllerNamespace,
		DNSDomain:           cfg.DNSDomain,
		EnvoyGateway:        cfg.EnvoyGateway,
		Client:              New(cli),
		logger:              cfg.Logger.WithName(string(egv1a1.LogComponentInfrastructureRunner)),
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
