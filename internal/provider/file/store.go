// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package file

import (
	"fmt"

	"github.com/fsnotify/fsnotify"
	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

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

// HandleEvent handles resources update according to the event type.
func (r *resourcesStore) HandleEvent(event fsnotify.Event, files, dirs []string) {
	switch event.Op {
	case fsnotify.Create, fsnotify.Write, fsnotify.Rename:
		// For these events, update only the file that triggered this event.
		if err := r.updateAndStore(event.Name, files, dirs); err != nil {
			r.logger.Error(err, "failed to update and store resources", "file", event.Name)
		}
	case fsnotify.Remove:
		// For remove event, just simply reload all resources.
		if err := r.reloadAndStore(files, dirs); err != nil {
			r.logger.Error(err, "failed to reload and store resources")
		}
	}
}

// TODO(sh2): For now, we assume that one file only contains one GatewayClass and all its other
// related resources, like Gateway, HTTPRoute, etc. If we managed to extend Resources structure,
// we also need to process all the resources and its relationship, like what is done in
// Kubernetes provider. However, this will cause us to maintain two places of the same logic
// in each provider. The ideal case is two different providers share the same resources process logic.
//
// - This issue is tracked by https://github.com/envoyproxy/gateway/issues/3213

// reloadAndStore reloads and stores all resources from given files and directories.
func (r *resourcesStore) reloadAndStore(files, dirs []string) error {
	resources, err := loadFromFilesAndDirs(files, dirs)
	if err != nil {
		return err
	}

	// Ignore every resources that has no GatewayClass.
	gwcResources := make(resource.ControllerResources, 0, 1)
	for _, res := range resources {
		if res.GatewayClass != nil {
			gwcResources = append(gwcResources, res)
		}
	}

	r.resources.GatewayAPIResources.Store(r.name, &gwcResources)
	r.logger.Info("store resources successfully", "len", len(gwcResources))

	return nil
}

// updateAndStore updates and stores resources load from the given target file.
// This method is optimized for the case that GatewayClass does not involve any changes.
func (r *resourcesStore) updateAndStore(target string, files, dirs []string) error {
	targetResource, err := loadFromFile(target)
	if err != nil {
		return err
	}

	allResourcesRef, ok := r.resources.GatewayAPIResources.Load(r.name)
	if !ok {
		// In case it happens, which very unlikely, reload all resources.
		return r.reloadAndStore(files, dirs)
	}

	allResources := *allResourcesRef
	allResourcesList := []*resource.Resources(allResources)

	diff, found := false, false
	targetKey := matchingKey(targetResource.GatewayClass)
	// If target resource is found in current resources, update it.
	for i, curr := range allResourcesList {
		if matchingKey(curr.GatewayClass) != targetKey {
			continue
		}

		found = true
		opts := []cmp.Option{
			cmpopts.IgnoreFields(resource.Resources{}, "serviceMap"),
		}
		if len(cmp.Diff(curr, targetResource, opts...)) > 0 {
			allResourcesList[i] = targetResource
			diff = true
		}
		break
	}

	// It only has two possibilities for the target which is not found in current resources:
	// 1) target is a new resource
	// 2) target is the one in current resources where the matching key of GatewayClass has changed
	// Hard to tell the difference between these two cases, so just reload.
	if !found {
		return r.reloadAndStore(files, dirs)
	}

	if diff {
		allResources = allResourcesList
		r.resources.GatewayAPIResources.Store(r.name, &allResources)
		r.logger.Info("update resources successfully", "key", targetKey)
	}

	return nil
}

func matchingKey(gc *gwapiv1.GatewayClass) string {
	return fmt.Sprintf("%s/%s", gc.Namespace, gc.Name)
}
