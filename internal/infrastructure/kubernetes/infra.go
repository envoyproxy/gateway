package kubernetes

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/envoyproxy/gateway/internal/ir"
)

const (
	// envoyGatewayNamespace is the namespace where envoy-gateway is running.
	envoyGatewayNamespace = "envoy-gateway-system"
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
	ns, found := os.LookupEnv("ENVOY_GATEWAY_NAMESPACE")

	if found {
		infra.Namespace = ns
	} else {
		infra.Namespace = envoyGatewayNamespace
	}

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

// addResource adds obj to the infra resources, using the object type
// to identify the object kind to add.
func (i *Infra) addResource(obj client.Object) error {
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

// CreateInfra creates the managed kube infra if it doesn't exist.
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

	if err := i.createServiceAccountIfNeeded(ctx, infra); err != nil {
		return err
	}

	if err := i.createDeploymentIfNeeded(ctx, infra); err != nil {
		return err
	}

	if err := i.createServiceIfNeeded(ctx, infra); err != nil {
		return err
	}

	return nil
}
