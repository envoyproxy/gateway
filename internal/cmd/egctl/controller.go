// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package egctl

import (
	"context"
	"os"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/message"
	"github.com/envoyproxy/gateway/internal/provider/file"
	"github.com/envoyproxy/gateway/internal/provider/kubernetes"
)

// NewSimpleController creates a simple controller that only has basic load
// and reconcile capabilities.
func NewSimpleController(ctx context.Context,
	pr *message.ProviderResources,
	namespace string,
) (loadAndReconcile func(*resource.LoadResources) error) {
	cfg, _ := config.New(os.Stdout)
	cfg.EnvoyGateway.Provider = &egv1a1.EnvoyGatewayProvider{Type: egv1a1.ProviderTypeCustom}
	cfg.ControllerNamespace = namespace

	reconciler, err := kubernetes.NewOfflineGatewayAPIController(ctx, cfg, nil, pr)
	if err != nil {
		panic(err)
	}
	store := file.NewResourcesStore(cfg.EnvoyGateway.Gateway.ControllerName, reconciler.Client, pr, cfg.Logger.Logger)

	return func(rs *resource.LoadResources) error {
		_, err := store.Store(ctx, rs, true)
		if err != nil {
			return err
		}
		return reconciler.Reconcile(ctx)
	}
}
