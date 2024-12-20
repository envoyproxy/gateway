// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package file

import (
	"github.com/go-logr/logr"

	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/message"
)

type resourcesStore struct {
	name      string
	resources *message.ProviderResources

	logger logr.Logger
}

func newResourcesStore(name string, resources *message.ProviderResources, logger logr.Logger) *resourcesStore {
	return &resourcesStore{
		name:      name,
		resources: resources,
		logger:    logger,
	}
}

// HandleEvent simply removes all the resources and triggers a resources reload from files
// and directories despite of the event type.
// TODO: Enhance this method by respecting the event type, and add support for multiple GatewayClass.
func (r *resourcesStore) HandleEvent(files, dirs []string) {
	r.logger.Info("reload all resources")

	r.resources.GatewayAPIResources.Delete(r.name)
	if err := r.LoadAndStore(files, dirs); err != nil {
		r.logger.Error(err, "failed to load and store resources")
	}
}

// LoadAndStore loads and stores all resources from files and directories.
func (r *resourcesStore) LoadAndStore(files, dirs []string) error {
	resources, err := loadFromFilesAndDirs(files, dirs)
	if err != nil {
		return err
	}

	// TODO(sh2): For now, we assume that one file only contains one GatewayClass and all its other
	// related resources, like Gateway, HTTPRoute, etc. If we managed to extend Resources structure,
	// we also need to process all the resources and its relationship, like what is done in
	// Kubernetes provider. However, this will cause us to maintain two places of the same logic
	// in each provider. The ideal case is two different providers share the same resources process logic.
	//
	// - This issue is tracked by https://github.com/envoyproxy/gateway/issues/3213

	// We cannot make sure by the time the Write event was triggered, whether the GatewayClass exist,
	// so here we just simply Store the first gatewayapi.Resources that has GatewayClass.
	gwcResources := make(resource.ControllerResources, 0, 1)
	for _, res := range resources {
		if res.GatewayClass != nil {
			gwcResources = append(gwcResources, res)
		}
	}
	if len(gwcResources) == 0 {
		return nil
	}

	r.resources.GatewayAPIResources.Store(r.name, &gwcResources)
	r.logger.Info("loaded and stored resources successfully")

	return nil
}
