// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package file

import (
	"github.com/fsnotify/fsnotify"
	"github.com/go-logr/logr"

	"github.com/envoyproxy/gateway/internal/gatewayapi"
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

func (r *resourcesStore) HandleEvent(event fsnotify.Event, files, dirs []string) {
	r.logger.Info("receive an event", "name", event.Name, "op", event.Op.String())

	// TODO(sh2): Find a better way to process the event. For now,
	// it only simply reload all the resources from files and
	// directories despite the event type.
	if err := r.LoadAndStore(files, dirs); err != nil {
		r.logger.Error(err, "error processing resources")
	}
}

// LoadAndStore loads and stores resources from files and directories.
func (r *resourcesStore) LoadAndStore(files, dirs []string) error {
	rs, err := loadFromFilesAndDirs(files, dirs)
	if err != nil {
		return err
	}

	// TODO(sh2): For now, we assume that one file only contains one GatewayClass and all its other
	// related resources, like Gateway, HTTPRoute, etc. If we manged to extend Resources structure,
	// we also need to process all the resources and its relationship, like what is done in
	// Kubernetes provider. However, this will cause us to maintain two places of the same logic
	// in each provider. The ideal case is two different providers share the same resources process logic.
	//
	// - This issue is tracked by https://github.com/envoyproxy/gateway/issues/3213
	//
	// So here we just simply Store each gatewayapi.Resources.
	var gwcResources gatewayapi.ControllerResources = rs
	r.resources.GatewayAPIResources.Store(r.name, &gwcResources)

	r.logger.Info("loaded and stored resources successfully")
	return nil
}
