package kubernetes

import (
	"context"
	"errors"
	"fmt"
	"sync"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/utils/env"
)

// Infra holds all the translated Infra IR resources and provides
// the scaffolding for the managing Kubernetes infrastructure.
type Infra struct {
	mu     sync.Mutex
	Client client.Client
	// Namespace is the Namespace used for managed infra.
	Namespace string
	Resources *Resources
}

// Resources are managed Kubernetes resources.
type Resources struct {
	ServiceAccount *corev1.ServiceAccount
	Deployment     *appsv1.Deployment
	Service        *corev1.Service
}

// NewInfra returns a new Infra.
func NewInfra(cli client.Client) *Infra {
	infra := &Infra{
		mu:        sync.Mutex{},
		Client:    cli,
		Resources: newResources(),
	}

	// Set the namespace used for the managed infra.
	infra.Namespace = env.Lookup("ENVOY_GATEWAY_NAMESPACE", config.EnvoyGatewayNamespace)

	return infra
}

// newResources returns a new Resources.
func newResources() *Resources {
	return &Resources{
		ServiceAccount: new(corev1.ServiceAccount),
		Deployment:     new(appsv1.Deployment),
		Service:        new(corev1.Service),
	}
}

// updateResource updates the obj to the infra resources, using the object type
// to identify the object kind to add.
func (i *Infra) updateResource(obj client.Object) error {
	i.mu.Lock()
	defer i.mu.Unlock()
	if i.Resources == nil {
		i.Resources = new(Resources)
	}

	switch o := obj.(type) {
	case *corev1.ServiceAccount:
		i.Resources.ServiceAccount = o
	case *appsv1.Deployment:
		i.Resources.Deployment = o
	case *corev1.Service:
		i.Resources.Service = o
	default:
		return fmt.Errorf("unexpected object kind %s", obj.GetObjectKind())
	}

	return nil
}

// CreateInfra creates the managed kube infra, if it doesn't exist.
func (i *Infra) CreateInfra(ctx context.Context, infra *ir.Infra) error {
	if infra == nil {
		return errors.New("infra ir is nil")
	}

	if infra.Proxy == nil {
		return errors.New("infra proxy ir is nil")
	}

	if i.Resources == nil {
		i.Resources = newResources()
	}

	if err := i.createOrUpdateServiceAccount(ctx, infra); err != nil {
		return err
	}

	if _, err := i.createOrUpdateConfigMap(ctx); err != nil {
		return err
	}

	if err := i.createOrUpdateDeployment(ctx, infra); err != nil {
		return err
	}

	if err := i.createOrUpdateService(ctx, infra); err != nil {
		return err
	}

	return nil
}

// DeleteInfra removes the managed kube infra, if it doesn't exist.
func (i *Infra) DeleteInfra(ctx context.Context, infra *ir.Infra) error {
	if infra == nil {
		return errors.New("infra ir is nil")
	}

	if err := i.deleteService(ctx, infra); err != nil {
		return err
	}

	if err := i.deleteDeployment(ctx, infra); err != nil {
		return err
	}

	if err := i.deleteConfigMap(ctx); err != nil {
		return err
	}

	if err := i.deleteServiceAccount(ctx, infra); err != nil {
		return err
	}

	return nil
}
