// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"
	"errors"

	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// watchAndReconcileSource is a concrete implementation of the Source interface.
type watchAndReconcileSource struct {
	condition <-chan struct{}
	object    client.Object
}

func NewWatchAndReconcileSource(cond <-chan struct{}, obj client.Object) source.Source {
	return &watchAndReconcileSource{condition: cond, object: obj}
}

// Start implements the Source interface. It registers the EventHandler with the Informer.
func (s *watchAndReconcileSource) Start(ctx context.Context, eh handler.EventHandler, queue workqueue.RateLimitingInterface, _ ...predicate.Predicate) error {
	if s.object == nil {
		return errors.New("object to queue is required")
	}
	// do not block controller startup
	go func() {
		select {
		case <-ctx.Done():
			return
		case <-s.condition:
			// Triggers a reconcile
			eh.Generic(ctx, event.GenericEvent{Object: s.object}, queue)
		}
	}()
	return nil
}
