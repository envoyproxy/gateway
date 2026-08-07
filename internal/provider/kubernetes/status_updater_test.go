// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"
	"sync/atomic"
	"testing"

	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/envoyproxy/gateway/internal/envoygateway"
)

func TestUpdateHandlerApplyReadsStatusFromStatusReader(t *testing.T) {
	key := types.NamespacedName{Namespace: "default", Name: "gateway"}
	staleCachedGateway := gatewayWithProgrammedStatus(key, metav1.ConditionTrue)
	liveGateway := gatewayWithProgrammedStatus(key, metav1.ConditionFalse)

	writerClient := fakeclient.NewClientBuilder().
		WithScheme(envoygateway.GetScheme()).
		WithStatusSubresource(&gwapiv1.Gateway{}).
		WithObjects(staleCachedGateway).
		Build()
	liveReader := fakeclient.NewClientBuilder().
		WithScheme(envoygateway.GetScheme()).
		WithStatusSubresource(&gwapiv1.Gateway{}).
		WithObjects(liveGateway).
		Build()

	statusWriter := &recordingStatusWriter{StatusWriter: writerClient.Status()}
	handler := NewUpdateHandler(logr.Discard(), &recordingStatusClient{
		Client:       writerClient,
		statusWriter: statusWriter,
	}, liveReader)

	handler.apply(Update{
		NamespacedName: key,
		Resource:       new(gwapiv1.Gateway),
		Mutator: MutatorFunc(func(obj client.Object) client.Object {
			gateway := obj.(*gwapiv1.Gateway).DeepCopy()
			gateway.Status.Conditions = []metav1.Condition{
				programmedCondition(metav1.ConditionTrue),
			}
			return gateway
		}),
	})

	if got := atomic.LoadInt32(&statusWriter.updates); got != 1 {
		t.Fatalf("expected status update to be written after comparing against live status, got %d writes", got)
	}
}

func TestUpdateHandlerApplySkipsUnchangedStatus(t *testing.T) {
	key := types.NamespacedName{Namespace: "default", Name: "gateway"}
	gateway := gatewayWithProgrammedStatus(key, metav1.ConditionTrue)

	writerClient := fakeclient.NewClientBuilder().
		WithScheme(envoygateway.GetScheme()).
		WithStatusSubresource(&gwapiv1.Gateway{}).
		WithObjects(gateway).
		Build()

	statusWriter := &recordingStatusWriter{StatusWriter: writerClient.Status()}
	handler := NewUpdateHandler(logr.Discard(), &recordingStatusClient{
		Client:       writerClient,
		statusWriter: statusWriter,
	}, nil)

	handler.apply(Update{
		NamespacedName: key,
		Resource:       new(gwapiv1.Gateway),
		Mutator: MutatorFunc(func(obj client.Object) client.Object {
			return obj.(*gwapiv1.Gateway).DeepCopy()
		}),
	})

	if got := atomic.LoadInt32(&statusWriter.updates); got != 0 {
		t.Fatalf("expected unchanged status to bypass status update, got %d writes", got)
	}
}

func TestUpdateHandlerApplyIgnoresNotFoundFromStatusReader(t *testing.T) {
	key := types.NamespacedName{Namespace: "default", Name: "gateway"}

	writerClient := fakeclient.NewClientBuilder().
		WithScheme(envoygateway.GetScheme()).
		WithStatusSubresource(&gwapiv1.Gateway{}).
		Build()
	liveReader := fakeclient.NewClientBuilder().
		WithScheme(envoygateway.GetScheme()).
		WithStatusSubresource(&gwapiv1.Gateway{}).
		Build()

	statusWriter := &recordingStatusWriter{StatusWriter: writerClient.Status()}
	handler := NewUpdateHandler(logr.Discard(), &recordingStatusClient{
		Client:       writerClient,
		statusWriter: statusWriter,
	}, liveReader)

	mutatorCalled := false
	handler.apply(Update{
		NamespacedName: key,
		Resource:       new(gwapiv1.Gateway),
		Mutator: MutatorFunc(func(obj client.Object) client.Object {
			mutatorCalled = true
			return obj
		}),
	})

	if mutatorCalled {
		t.Fatal("expected missing live object to bypass mutation")
	}
	if got := atomic.LoadInt32(&statusWriter.updates); got != 0 {
		t.Fatalf("expected missing live object to bypass status update, got %d writes", got)
	}
}

type recordingStatusClient struct {
	client.Client
	statusWriter client.StatusWriter
}

func (c *recordingStatusClient) Status() client.StatusWriter {
	return c.statusWriter
}

type recordingStatusWriter struct {
	client.StatusWriter
	updates int32
}

func (w *recordingStatusWriter) Update(ctx context.Context, obj client.Object, opts ...client.SubResourceUpdateOption) error {
	atomic.AddInt32(&w.updates, 1)
	return w.StatusWriter.Update(ctx, obj, opts...)
}

func gatewayWithProgrammedStatus(key types.NamespacedName, status metav1.ConditionStatus) *gwapiv1.Gateway {
	return &gwapiv1.Gateway{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: key.Namespace,
			Name:      key.Name,
		},
		Status: gwapiv1.GatewayStatus{
			Conditions: []metav1.Condition{
				programmedCondition(status),
			},
		},
	}
}

func programmedCondition(status metav1.ConditionStatus) metav1.Condition {
	reason := string(gwapiv1.GatewayReasonProgrammed)
	message := "Address assigned to the Gateway, 1/1 envoy replicas available"
	if status == metav1.ConditionFalse {
		reason = string(gwapiv1.GatewayReasonNoResources)
		message = "Envoy replicas unavailable"
	}
	return metav1.Condition{
		Type:               string(gwapiv1.GatewayConditionProgrammed),
		Status:             status,
		Reason:             reason,
		Message:            message,
		ObservedGeneration: 1,
	}
}
