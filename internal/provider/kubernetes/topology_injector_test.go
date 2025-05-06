// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"gomodules.xyz/jsonpatch/v2"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/envoyproxy/gateway/internal/gatewayapi"
)

func TestProxyTopologyInjector_Handle(t *testing.T) {
	defaultPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "bar",
			Labels: map[string]string{
				"app.kubernetes.io/component":      "proxy",
				gatewayapi.OwningGatewayClassLabel: "eg",
				"app.kubernetes.io/managed-by":     "envoy-gateway",
				"app.kubernetes.io/name":           "envoy",
			},
		},
		Spec: corev1.PodSpec{},
	}
	defaultNode := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "node-A",
			Labels: map[string]string{corev1.LabelTopologyZone: "zone1"},
		},
	}

	cases := []struct {
		caseName          string
		obj               client.Object
		node              *corev1.Node
		pod               *corev1.Pod
		expectedPatchResp []jsonpatch.JsonPatchOperation
	}{
		{
			caseName: "valid binding",
			obj: &corev1.Binding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      defaultPod.Name,
					Namespace: defaultPod.Namespace,
				},
				Target: corev1.ObjectReference{Name: defaultNode.Name},
			},
			node: defaultNode,
			pod:  defaultPod,
			expectedPatchResp: []jsonpatch.JsonPatchOperation{{
				Operation: "add",
				Path:      "/metadata/annotations",
				Value: map[string]interface{}{
					"topology.kubernetes.io/zone": "zone1",
				},
			}},
		},
		{
			caseName: "empty target",
			obj: &corev1.Binding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      defaultPod.Name,
					Namespace: defaultPod.Namespace,
				},
			},
			node:              defaultNode,
			pod:               defaultPod,
			expectedPatchResp: nil,
		},
		{
			caseName: "skip binding - no label",
			obj: &corev1.Binding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "baz",
					Namespace: "bar",
				},
			},
			node:              defaultNode,
			pod:               &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Namespace: "bar", Name: "baz"}},
			expectedPatchResp: nil,
		},
		{
			caseName: "no matching pod",
			obj: &corev1.Binding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "baz",
					Namespace: "bar",
				},
			},
			node:              defaultNode,
			pod:               defaultPod,
			expectedPatchResp: nil,
		},
	}
	for _, tc := range cases {
		t.Run(tc.caseName, func(t *testing.T) {
			scheme := runtime.NewScheme()
			if err := corev1.AddToScheme(scheme); err != nil {
				t.Fatal(err)
			}
			decoder := admission.NewDecoder(scheme)

			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				WithRuntimeObjects(tc.node, tc.pod).
				Build()

			mutator := &ProxyTopologyInjector{
				Client:  fakeClient,
				Decoder: decoder,
			}

			objBytes, err := json.Marshal(tc.obj)
			require.NoError(t, err)

			req := admission.Request{
				AdmissionRequest: admissionv1.AdmissionRequest{
					UID:       types.UID("1234"),
					Name:      tc.obj.GetName(),
					Namespace: tc.obj.GetNamespace(),
					Operation: admissionv1.Update,
					Object:    runtime.RawExtension{Raw: objBytes},
				},
			}

			resp := mutator.Handle(context.Background(), req)
			require.True(t, resp.Allowed)

			require.Equal(t, tc.expectedPatchResp, resp.Patches)
		})
	}
}
