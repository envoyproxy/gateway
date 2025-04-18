// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package envoy_test

import (
	"context"
	"encoding/json"
	"testing"

	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/envoyproxy/gateway/internal/cmd/envoy"
	"github.com/envoyproxy/gateway/internal/gatewayapi"
)

func TestPodBindingMutator_Handle_UpdateAddsTopologyLabels(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	decoder := admission.NewDecoder(scheme)

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "bar",
			Labels: map[string]string{
				"app.kubernetes.io/component":      "proxy",
				gatewayapi.OwningGatewayClassLabel: "eg",
			},
		},
		Spec: corev1.PodSpec{},
	}
	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:            "node-A",
			Labels:          map[string]string{corev1.LabelTopologyZone: "zone1"},
			ResourceVersion: "",
		},
	}
	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithRuntimeObjects(node, pod).
		Build()

	mutator := &envoy.ProxyTopologyMutator{
		Client:  fakeClient,
		Decoder: decoder,
	}

	bindObj := &corev1.Binding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pod.Name,
			Namespace: pod.Namespace,
		},
		Target: corev1.ObjectReference{Name: node.Name},
	}

	oldBytes, err := json.Marshal(bindObj)
	if err != nil {
		t.Fatalf("marshal oldPod: %v", err)
	}

	req := admission.Request{
		AdmissionRequest: admissionv1.AdmissionRequest{
			UID:         types.UID("1234"),
			Kind:        metav1.GroupVersionKind{Group: "", Version: "v1", Kind: "Pod"},
			Resource:    metav1.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"},
			SubResource: "binding",
			Name:        bindObj.Name,
			Namespace:   bindObj.Namespace,
			Operation:   admissionv1.Update,
			Object:      runtime.RawExtension{Raw: oldBytes},
		},
	}

	resp := mutator.Handle(context.Background(), req)
	if !resp.Allowed {
		t.Fatalf("expected Allowed response, got: %v", resp.Result)
	}

	updatedPod := &corev1.Pod{}
	if err = fakeClient.Get(context.Background(), types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}, updatedPod); err != nil {
		t.Fatalf("get pod: %v", err)
	}

	if zone, ok := updatedPod.Labels[corev1.LabelTopologyZone]; !ok || zone != "zone1" {
		t.Fatalf("expected zone1, got: %v", zone)
	}
}
