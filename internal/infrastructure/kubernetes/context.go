package kubernetes

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/ir"
)

const (
	KindServiceAccount Kind = "ServiceAccount"
)

type Kind string

// Context holds all the translated Infra IR resources and provides
// the scaffolding for the managing Kubernetes infrastructure.
type Context struct {
	Client    client.Client
	Log       logr.Logger
	Resources *Resources
}

// Resources are managed Kubernetes resources.
type Resources struct {
	ServiceAccount *corev1.ServiceAccount
}

// NewContext returns a new Context.
func NewContext(cli client.Client, cfg *config.Server) *Context {
	return &Context{
		Client: cli,
		Log:    cfg.Logger,
		Resources: &Resources{
			ServiceAccount: new(corev1.ServiceAccount),
		},
	}
}

// addResource adds the resource to the context resources, using kind to
// identify the object kind to add.
func (c *Context) addResource(kind Kind, obj client.Object) error {
	if c.Resources == nil {
		c.Resources = new(Resources)
	}

	switch kind {
	case KindServiceAccount:
		sa, ok := obj.(*corev1.ServiceAccount)
		if !ok {
			return fmt.Errorf("unexpected object kind %s", obj.GetObjectKind())
		}
		c.Resources.ServiceAccount = sa
	default:
		return fmt.Errorf("unexpected object kind %s", obj.GetObjectKind())
	}

	return nil
}

// CreateIfNeeded creates the managed kube infra if it doesn't exist.
func (c *Context) CreateIfNeeded(ctx context.Context, infra *ir.Infra) error {
	if infra == nil {
		return errors.New("infra ir is nil")
	}

	if infra.Proxy == nil {
		return errors.New("infra proxy ir is nil")
	}

	if c.Resources == nil {
		c.Resources = &Resources{
			ServiceAccount: new(corev1.ServiceAccount),
		}
	}

	if err := c.createServiceAccountIfNeeded(ctx, infra); err != nil {
		return err
	}

	return nil
}
