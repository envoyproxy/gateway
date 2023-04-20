// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/envoyproxy/gateway/api/config/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/infrastructure/kubernetes/applier"
	"github.com/envoyproxy/gateway/internal/infrastructure/kubernetes/proxy"
	"github.com/envoyproxy/gateway/internal/infrastructure/kubernetes/ratelimit"
)

var (
	_ ResourceRender = &proxy.ResourceRender{}
	_ ResourceRender = &ratelimit.ResourceRender{}
)

// ResourceRender renders Kubernetes infrastructure resources
// based on Infra IR resources.
type ResourceRender interface {
	Name() string
	ServiceAccount() (*corev1.ServiceAccount, error)
	Service() (*corev1.Service, error)
	ConfigMap() (*corev1.ConfigMap, error)
	Deployment() (*appsv1.Deployment, error)
}

// Infra manages the creation and deletion of Kubernetes infrastructure
// based on Infra IR resources.
type Infra struct {
	Client client.Client

	// Namespace is the Namespace used for managed infra.
	Namespace string

	// EnvoyGateway is the configuration used to startup Envoy Gateway.
	EnvoyGateway *v1alpha1.EnvoyGateway
	applier      *applier.Instance
}

// NewInfra returns a new Infra.
func NewInfra(cli client.Client, cfg *config.Server) *Infra {
	return &Infra{
		Client:       cli,
		Namespace:    cfg.Namespace,
		EnvoyGateway: cfg.EnvoyGateway,
		applier:      applier.New(cli),
	}
}

func (i *Infra) createOrUpdate(ctx context.Context, r ResourceRender) error {
	if err := i.createOrUpdateServiceAccount(ctx, r); err != nil {
		return err
	}

	if err := i.createOrUpdateConfigMap(ctx, r); err != nil {
		return err
	}

	if err := i.createOrUpdateDeployment(ctx, r); err != nil {
		return err
	}

	if err := i.createOrUpdateService(ctx, r); err != nil {
		return err
	}

	return nil
}

// createOrUpdateServiceAccount creates a ServiceAccount in the kube api server based on the
// provided ResourceRender, if it doesn't exist and updates it if it does.
func (i *Infra) createOrUpdateServiceAccount(ctx context.Context, r ResourceRender) error {
	sa, err := r.ServiceAccount()
	if err != nil {
		return err
	}
	return i.applier.CreateOrUpdateServiceAccount(ctx, sa)
}

// createOrUpdateConfigMap creates a ConfigMap in the Kube api server based on the provided
// ResourceRender, if it doesn't exist and updates it if it does.
func (i *Infra) createOrUpdateConfigMap(ctx context.Context, r ResourceRender) error {
	cm, err := r.ConfigMap()
	if err != nil {
		return err
	}

	return i.applier.CreateOrUpdateConfigMap(ctx, cm)
}

// createOrUpdateDeployment creates a Deployment in the kube api server based on the provided
// ResourceRender, if it doesn't exist and updates it if it does.
func (i *Infra) createOrUpdateDeployment(ctx context.Context, r ResourceRender) error {
	deployment, err := r.Deployment()
	if err != nil {
		return err
	}
	return i.applier.CreateOrUpdateDeployment(ctx, deployment)
}

// createOrUpdateRateLimitService creates a Service in the kube api server based on the provided ResourceRender,
// if it doesn't exist or updates it if it does.
func (i *Infra) createOrUpdateService(ctx context.Context, r ResourceRender) error {
	svc, err := r.Service()
	if err != nil {
		return err
	}

	return i.applier.CreateOrUpdateService(ctx, svc)
}

func (i *Infra) delete(ctx context.Context, r ResourceRender) error {
	if err := i.deleteServiceAccount(ctx, r); err != nil {
		return err
	}

	if err := i.deleteConfigMap(ctx, r); err != nil {
		return err
	}

	if err := i.deleteDeployment(ctx, r); err != nil {
		return err
	}

	if err := i.deleteService(ctx, r); err != nil {
		return err
	}

	return nil
}

// deleteServiceAccount deletes the ServiceAccount in the kube api server,
// if it exists.
func (i *Infra) deleteServiceAccount(ctx context.Context, r ResourceRender) error {
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: i.Namespace,
			Name:      r.Name(),
		},
	}

	return i.applier.DeleteServiceAccount(ctx, sa)
}

// deleteDeployment deletes the Envoy Deployment in the kube api server, if it exists.
func (i *Infra) deleteDeployment(ctx context.Context, r ResourceRender) error {
	deploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: i.Namespace,
			Name:      r.Name(),
		},
	}

	return i.applier.DeleteDeployment(ctx, deploy)
}

// deleteConfigMap deletes the ConfigMap in the kube api server, if it exists.
func (i *Infra) deleteConfigMap(ctx context.Context, r ResourceRender) error {
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: i.Namespace,
			Name:      r.Name(),
		},
	}

	return i.applier.DeleteConfigMap(ctx, cm)
}

// deleteService deletes the Service in the kube api server, if it exists.
func (i *Infra) deleteService(ctx context.Context, r ResourceRender) error {
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: i.Namespace,
			Name:      r.Name(),
		},
	}

	return i.applier.DeleteService(ctx, svc)
}
