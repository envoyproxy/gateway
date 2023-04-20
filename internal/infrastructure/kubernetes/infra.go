// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"

	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/envoyproxy/gateway/api/config/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
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
	// Namespace is the Namespace used for managed infra.
	Namespace string

	// EnvoyGateway is the configuration used to startup Envoy Gateway.
	EnvoyGateway *v1alpha1.EnvoyGateway

	// Client wrap k8s client.
	Client *InfraClient
}

// NewInfra returns a new Infra.
func NewInfra(cli client.Client, cfg *config.Server) *Infra {
	return &Infra{
		Namespace:    cfg.Namespace,
		EnvoyGateway: cfg.EnvoyGateway,
		Client:       New(cli),
	}
}

// createOrUpdate creates a ServiceAccount/ConfigMap/Deployment/Service in the kube api server based on the
// provided ResourceRender, if it doesn't exist and updates it if it does.
func (i *Infra) createOrUpdate(ctx context.Context, r ResourceRender) error {
	if err := i.createOrUpdateServiceAccount(ctx, r); err != nil {
		return errors.Wrapf(err, "failed to create or update serviceaccount %s/%s", i.Namespace, r.Name())
	}

	if err := i.createOrUpdateConfigMap(ctx, r); err != nil {
		return errors.Wrapf(err, "failed to create or update configmap %s/%s", i.Namespace, r.Name())
	}

	if err := i.createOrUpdateDeployment(ctx, r); err != nil {
		return errors.Wrapf(err, "failed to create or update deployment %s/%s", i.Namespace, r.Name())
	}

	if err := i.createOrUpdateService(ctx, r); err != nil {
		return errors.Wrapf(err, "failed to create or update service %s/%s", i.Namespace, r.Name())
	}

	return nil
}

// delete deletes the ServiceAccount/ConfigMap/Deployment/Service in the kube api server, if it exists.
func (i *Infra) delete(ctx context.Context, r ResourceRender) error {
	if err := i.deleteServiceAccount(ctx, r); err != nil {
		return errors.Wrapf(err, "failed to delete serviceaccount %s/%s", i.Namespace, r.Name())
	}
	if err := i.deleteConfigMap(ctx, r); err != nil {
		return errors.Wrapf(err, "failed to delete configmap %s/%s", i.Namespace, r.Name())
	}
	if err := i.deleteDeployment(ctx, r); err != nil {
		return errors.Wrapf(err, "failed to delete deployment %s/%s", i.Namespace, r.Name())
	}
	if err := i.deleteService(ctx, r); err != nil {
		return errors.Wrapf(err, "failed to delete service %s/%s", i.Namespace, r.Name())
	}

	return nil
}
