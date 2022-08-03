package kubernetes

import (
	"context"
	"errors"
	"fmt"
	"sync"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/envoyproxy/gateway/internal/ir"
)

// Infra holds all the translated Infra IR resources and provides
// the scaffolding for the managing Kubernetes infrastructure.
type Infra struct {
	mu        sync.Mutex
	Client    client.Client
	Resources *Resources
}

// Resources are managed Kubernetes resources.
type Resources struct {
	ServiceAccount *corev1.ServiceAccount
}

// NewInfra returns a new Infra.
func NewInfra(cli client.Client) *Infra {
	return &Infra{
		mu:     sync.Mutex{},
		Client: cli,
		Resources: &Resources{
			ServiceAccount: new(corev1.ServiceAccount),
		},
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
		i.Resources = &Resources{
			ServiceAccount: new(corev1.ServiceAccount),
		}
	}

	if err := i.createServiceAccountIfNeeded(ctx, infra); err != nil {
		return err
	}

	return nil
}
