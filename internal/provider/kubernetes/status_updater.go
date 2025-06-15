// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"
	"sync"
	"time"

	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	gwapiv1a3 "sigs.k8s.io/gateway-api/apis/v1alpha3"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/metrics"
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
	updateChannel chan Update
	wg            *sync.WaitGroup
}

func NewUpdateHandler(log logr.Logger, client client.Client) *UpdateHandler {
	u := &UpdateHandler{
		log:           log,
		client:        client,
		updateChannel: make(chan Update, 1000),
		wg:            new(sync.WaitGroup),
	}

	u.wg.Add(1)

	return u
}

func (u *UpdateHandler) apply(update Update) {
	var (
		startTime = time.Now()
		obj       = update.Resource
		objKind   = KindOf(obj)
	)
	log := u.log.WithValues("name", update.NamespacedName.Name,
		"namespace", update.NamespacedName.Namespace, "kind", objKind)
	defer func() {
		updateDuration := time.Since(startTime)
		statusUpdateDurationSeconds.With(kindLabel.Value(objKind)).Record(updateDuration.Seconds())
	}()

	if err := retry.OnError(retry.DefaultBackoff, func(err error) bool {
		if kerrors.IsConflict(err) {
			statusUpdateTotal.WithFailure(metrics.ReasonConflict, kindLabel.Value(objKind)).Increment()
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
			log.Info("status unchanged, bypassing update")

			statusUpdateTotal.WithStatus(statusNoAction, kindLabel.Value(objKind)).Increment()
			return nil
		}

		newObj.SetUID(obj.GetUID())

		return u.client.Status().Update(context.Background(), newObj)
	}); err != nil {
		log.Error(err, "unable to update status")

		statusUpdateTotal.WithFailure(metrics.ReasonError, kindLabel.Value(objKind)).Increment()
	} else {
		statusUpdateTotal.WithSuccess(kindLabel.Value(objKind)).Increment()
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
	u.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return nil
		case update := <-u.updateChannel:
			u.log.Info("received a status update", "namespace", update.NamespacedName.Namespace,
				"name", update.NamespacedName.Name, "kind", KindOf(update.Resource))

			u.apply(update)
		}
	}
}

// Writer retrieves the interface that should be used to write to the UpdateHandler.
func (u *UpdateHandler) Writer() Updater {
	return &UpdateWriter{
		updateChannel: u.updateChannel,
		wg:            u.wg,
	}
}

// Updater describes an interface to send status updates somewhere.
type Updater interface {
	Send(u Update)
}

// UpdateWriter takes status updates and sends these to the UpdateHandler via a channel.
type UpdateWriter struct {
	updateChannel chan<- Update
	wg            *sync.WaitGroup
}

// Send sends the given Update off to the update channel for writing by the UpdateHandler.
func (u *UpdateWriter) Send(update Update) {
	// Wait until updater is ready
	u.wg.Wait()
	u.updateChannel <- update
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
//	BackendTrafficPolicy
//	SecurityPolicy
//	BackendTLSPolicy
//	EnvoyExtensionPolicy
//	Unstructured (for server extension policies)
func isStatusEqual(objA, objB interface{}) bool {
	opts := cmp.Options{
		cmpopts.IgnoreFields(metav1.Condition{}, "LastTransitionTime"),
		cmpopts.IgnoreMapEntries(func(k string, _ any) bool {
			return k == "lastTransitionTime"
		}),
	}
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
	case *gwapiv1.GRPCRoute:
		if b, ok := objB.(*gwapiv1.GRPCRoute); ok {
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
	case *gwapiv1a3.BackendTLSPolicy:
		if b, ok := objB.(*gwapiv1a3.BackendTLSPolicy); ok {
			if cmp.Equal(a.Status, b.Status, opts) {
				return true
			}
		}
	case *egv1a1.EnvoyExtensionPolicy:
		if b, ok := objB.(*egv1a1.EnvoyExtensionPolicy); ok {
			if cmp.Equal(a.Status, b.Status, opts) {
				return true
			}
		}
	case *unstructured.Unstructured:
		if b, ok := objB.(*unstructured.Unstructured); ok {
			if cmp.Equal(a.Object["status"], b.Object["status"], opts) {
				return true
			}
		}
	case *egv1a1.Backend:
		if b, ok := objB.(*egv1a1.Backend); ok {
			if cmp.Equal(a.Status, b.Status, opts) {
				return true
			}
		}
	}

	return false
}

// KindOf returns the known kind string for the given Kubernetes object,
// returns Unknown for the unsupported object.
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
//	BackendTrafficPolicy
//	SecurityPolicy
//	BackendTLSPolicy
//	EnvoyExtensionPolicy
//	Unstructured (for Extension Policies)
func KindOf(obj interface{}) string {
	var kind string
	switch o := obj.(type) {
	case *gwapiv1.GatewayClass:
		kind = resource.KindGatewayClass
	case *gwapiv1.Gateway:
		kind = resource.KindGateway
	case *gwapiv1.HTTPRoute:
		kind = resource.KindHTTPRoute
	case *gwapiv1a2.TLSRoute:
		kind = resource.KindTLSRoute
	case *gwapiv1a2.TCPRoute:
		kind = resource.KindTCPRoute
	case *gwapiv1a2.UDPRoute:
		kind = resource.KindUDPRoute
	case *gwapiv1.GRPCRoute:
		kind = resource.KindGRPCRoute
	case *egv1a1.EnvoyPatchPolicy:
		kind = resource.KindEnvoyPatchPolicy
	case *egv1a1.ClientTrafficPolicy:
		kind = resource.KindClientTrafficPolicy
	case *egv1a1.BackendTrafficPolicy:
		kind = resource.KindBackendTrafficPolicy
	case *egv1a1.SecurityPolicy:
		kind = resource.KindSecurityPolicy
	case *egv1a1.EnvoyExtensionPolicy:
		kind = resource.KindEnvoyExtensionPolicy
	case *gwapiv1a3.BackendTLSPolicy:
		kind = resource.KindBackendTLSPolicy
	case *unstructured.Unstructured:
		kind = o.GetKind()
	case *egv1a1.Backend:
		kind = resource.KindBackend
	default:
		kind = "Unknown"
	}

	return kind
}
