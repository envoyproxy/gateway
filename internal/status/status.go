// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

// This file contains code derived from Contour,
// https://github.com/projectcontour/contour
// from the source file
// https://github.com/projectcontour/contour/blob/main/internal/k8s/status.go
// and is provided here subject to the following:
// Copyright Project Contour Authors
// SPDX-License-Identifier: Apache-2.0

package status

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

// Update contains an all the information needed to update an object's status.
// Send down a channel to the goroutine that actually writes the changes back.
type Update struct {
	NamespacedName types.NamespacedName
	Resource       client.Object
	Mutator        Mutator
}

// Mutator is an interface to hold mutator functions for status updates.
type Mutator interface {
	Mutate(obj client.Object) client.Object
}

// MutatorFunc is a function adaptor for Mutators.
type MutatorFunc func(client.Object) client.Object

// Mutate adapts the MutatorFunc to fit through the Mutator interface.
func (m MutatorFunc) Mutate(old client.Object) client.Object {
	if m == nil {
		return nil
	}

	return m(old)
}

// UpdateHandler holds the details required to actually write an Update back to the referenced object.
type UpdateHandler struct {
	log           logr.Logger
	client        client.Client
	sendUpdates   chan struct{}
	updateChannel chan Update
}

func NewUpdateHandler(log logr.Logger, client client.Client) *UpdateHandler {
	return &UpdateHandler{
		log:           log,
		client:        client,
		sendUpdates:   make(chan struct{}),
		updateChannel: make(chan Update, 100),
	}
}

func (u *UpdateHandler) apply(update Update) {
	var statusUpdateErr error
	obj := update.Resource
	objKind := kindOf(obj)

	startTime := time.Now()

	statusUpdateTotal.With(kindLabel.Value(objKind)).Increment()

	defer func() {
		updateDuration := time.Since(startTime)
		if statusUpdateErr != nil {
			statusUpdateDurationSeconds.With(kindLabel.Value(objKind)).Record(updateDuration.Seconds())
			statusUpdateFailed.With(kindLabel.Value(objKind)).Increment()
		} else {
			statusUpdateDurationSeconds.With(kindLabel.Value(objKind)).Record(updateDuration.Seconds())
			statusUpdateSuccess.With(kindLabel.Value(objKind)).Increment()
		}
	}()

	if statusUpdateErr = retry.OnError(retry.DefaultBackoff, func(err error) bool {
		if kerrors.IsConflict(err) {
			statusUpdateConflict.With(kindLabel.Value(objKind)).Increment()
			return true
		}
		return false
	}, func() error {
		// Get the resource.
		if err := u.client.Get(context.Background(), update.NamespacedName, obj); err != nil {
			if kerrors.IsNotFound(err) {
				return nil
			}
			return err
		}

		newObj := update.Mutator.Mutate(obj)

		if isStatusEqual(obj, newObj) {
			u.log.WithName(update.NamespacedName.Name).
				WithName(update.NamespacedName.Namespace).
				Info("status unchanged, bypassing update")
			statusUpdateNoop.With(kindLabel.Value(objKind)).Increment()
			return nil
		}

		newObj.SetUID(obj.GetUID())

		return u.client.Status().Update(context.Background(), newObj)
	}); statusUpdateErr != nil {
		u.log.Error(statusUpdateErr, "unable to update status", "name", update.NamespacedName.Name,
			"namespace", update.NamespacedName.Namespace)
	}
}

func (u *UpdateHandler) NeedLeaderElection() bool {
	return true
}

// Start runs the goroutine to perform status writes.
func (u *UpdateHandler) Start(ctx context.Context) error {
	u.log.Info("started status update handler")
	defer u.log.Info("stopped status update handler")

	// Enable Updaters to start sending updates to this handler.
	close(u.sendUpdates)

	for {
		select {
		case <-ctx.Done():
			return nil
		case update := <-u.updateChannel:
			u.log.Info("received a status update", "namespace", update.NamespacedName.Namespace,
				"name", update.NamespacedName.Name)

			u.apply(update)
		}
	}
}

// Writer retrieves the interface that should be used to write to the UpdateHandler.
func (u *UpdateHandler) Writer() Updater {
	return &UpdateWriter{
		enabled:       u.sendUpdates,
		updateChannel: u.updateChannel,
	}
}

// Updater describes an interface to send status updates somewhere.
type Updater interface {
	Send(u Update)
}

// UpdateWriter takes status updates and sends these to the UpdateHandler via a channel.
type UpdateWriter struct {
	enabled       <-chan struct{}
	updateChannel chan<- Update
}

// Send sends the given Update off to the update channel for writing by the UpdateHandler.
func (u *UpdateWriter) Send(update Update) {
	// Non-blocking receive to see if we should pass along update.
	select {
	case <-u.enabled:
		u.updateChannel <- update
	default:
	}
}

// isStatusEqual checks if two objects have equivalent status.
//
// Supported objects:
//
//	GatewayClasses
//	Gateway
//	HTTPRoute
//	TLSRoute
//	TCPRoute
//	UDPRoute
//	GRPCRoute
//	EnvoyPatchPolicy
//	ClientTrafficPolicy
//	SecurityPolicy
func isStatusEqual(objA, objB interface{}) bool {
	opts := cmpopts.IgnoreFields(metav1.Condition{}, "LastTransitionTime")
	switch a := objA.(type) {
	case *gwapiv1.GatewayClass:
		if b, ok := objB.(*gwapiv1.GatewayClass); ok {
			if cmp.Equal(a.Status, b.Status, opts) {
				return true
			}
		}
	case *gwapiv1.Gateway:
		if b, ok := objB.(*gwapiv1.Gateway); ok {
			if cmp.Equal(a.Status, b.Status, opts) {
				return true
			}
		}
	case *gwapiv1.HTTPRoute:
		if b, ok := objB.(*gwapiv1.HTTPRoute); ok {
			if cmp.Equal(a.Status, b.Status, opts) {
				return true
			}
		}
	case *gwapiv1a2.TLSRoute:
		if b, ok := objB.(*gwapiv1a2.TLSRoute); ok {
			if cmp.Equal(a.Status, b.Status, opts) {
				return true
			}
		}
	case *gwapiv1a2.TCPRoute:
		if b, ok := objB.(*gwapiv1a2.TCPRoute); ok {
			if cmp.Equal(a.Status, b.Status, opts) {
				return true
			}
		}
	case *gwapiv1a2.UDPRoute:
		if b, ok := objB.(*gwapiv1a2.UDPRoute); ok {
			if cmp.Equal(a.Status, b.Status, opts) {
				return true
			}
		}
	case *gwapiv1a2.GRPCRoute:
		if b, ok := objB.(*gwapiv1a2.GRPCRoute); ok {
			if cmp.Equal(a.Status, b.Status, opts) {
				return true
			}
		}
	case *egv1a1.EnvoyPatchPolicy:
		if b, ok := objB.(*egv1a1.EnvoyPatchPolicy); ok {
			if cmp.Equal(a.Status, b.Status, opts) {
				return true
			}
		}
	case *egv1a1.ClientTrafficPolicy:
		if b, ok := objB.(*egv1a1.ClientTrafficPolicy); ok {
			if cmp.Equal(a.Status, b.Status, opts) {
				return true
			}
		}
	case *egv1a1.BackendTrafficPolicy:
		if b, ok := objB.(*egv1a1.BackendTrafficPolicy); ok {
			if cmp.Equal(a.Status, b.Status, opts) {
				return true
			}
		}
	case *egv1a1.SecurityPolicy:
		if b, ok := objB.(*egv1a1.SecurityPolicy); ok {
			if cmp.Equal(a.Status, b.Status, opts) {
				return true
			}
		}
	}

	return false
}

// kindOf returns the kind string for the given Kubernetes object.
func kindOf(obj interface{}) string {
	switch obj.(type) {
	case *gwapiv1.GatewayClass:
		return "GatewayClass"
	case *gwapiv1.Gateway:
		return "Gateway"
	case *gwapiv1.HTTPRoute:
		return "HTTPRoute"
	case *gwapiv1a2.TLSRoute:
		return "TLSRoute"
	case *gwapiv1a2.TCPRoute:
		return "TCPRoute"
	case *gwapiv1a2.UDPRoute:
		return "UDPRoute"
	case *gwapiv1a2.GRPCRoute:
		return "GRPCRoute"
	case *egv1a1.EnvoyPatchPolicy:
		return "EnvoyPatchPolicy"
	case *egv1a1.ClientTrafficPolicy:
		return "ClientTrafficPolicy"
	case *egv1a1.BackendTrafficPolicy:
		return "BackendTrafficPolicy"
	case *egv1a1.SecurityPolicy:
		return "SecurityPolicy"
	}

	return ""
}
