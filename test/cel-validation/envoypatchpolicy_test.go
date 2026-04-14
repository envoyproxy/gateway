// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build celvalidation

package celvalidation

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

func TestEnvoyPatchPolicyTarget(t *testing.T) {
	ctx := context.Background()
	baseEPP := egv1a1.EnvoyPatchPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "epp",
			Namespace: metav1.NamespaceDefault,
		},
		Spec: egv1a1.EnvoyPatchPolicySpec{
			Type: egv1a1.JSONPatchEnvoyPatchType,
			JSONPatches: []egv1a1.EnvoyJSONPatchConfig{
				{
					Type: egv1a1.ListenerEnvoyResourceType,
					Name: "test-listener",
					Operation: egv1a1.JSONPatchOperation{
						Op:    "add",
						Path:  new("/foo"),
						Value: &apiextensionsv1.JSON{Raw: []byte(`"bar"`)},
					},
				},
			},
		},
	}

	cases := []struct {
		desc         string
		mutate       func(epp *egv1a1.EnvoyPatchPolicy)
		mutateStatus func(epp *egv1a1.EnvoyPatchPolicy)
		wantErrors   []string
	}{
		{
			desc: "valid gateway targetRef",
			mutate: func(epp *egv1a1.EnvoyPatchPolicy) {
				epp.Spec.TargetRef = &gwapiv1.LocalPolicyTargetReference{
					Group: gwapiv1.Group("gateway.networking.k8s.io"),
					Kind:  gwapiv1.Kind("Gateway"),
					Name:  gwapiv1.ObjectName("eg"),
				}
			},
			wantErrors: []string{},
		},
		{
			desc: "valid gateway targetRefs",
			mutate: func(epp *egv1a1.EnvoyPatchPolicy) {
				epp.Spec.TargetRefs = []gwapiv1.LocalPolicyTargetReference{
					{
						Group: gwapiv1.Group("gateway.networking.k8s.io"),
						Kind:  gwapiv1.Kind("Gateway"),
						Name:  gwapiv1.ObjectName("eg"),
					},
				}
			},
			wantErrors: []string{},
		},
		{
			desc: "both targetRef and targetRefs",
			mutate: func(epp *egv1a1.EnvoyPatchPolicy) {
				epp.Spec.TargetRef = &gwapiv1.LocalPolicyTargetReference{
					Group: gwapiv1.Group("gateway.networking.k8s.io"),
					Kind:  gwapiv1.Kind("Gateway"),
					Name:  gwapiv1.ObjectName("eg"),
				}
				epp.Spec.TargetRefs = []gwapiv1.LocalPolicyTargetReference{
					{
						Group: gwapiv1.Group("gateway.networking.k8s.io"),
						Kind:  gwapiv1.Kind("Gateway"),
						Name:  gwapiv1.ObjectName("eg"),
					},
				}
			},
			wantErrors: []string{
				"spec: Invalid value:",
				"exactly one of targetRef or targetRefs must be set",
			},
		},
		{
			desc: "no targetRef or targetRefs",
			mutate: func(_ *egv1a1.EnvoyPatchPolicy) {
				// Don't set either field
			},
			wantErrors: []string{
				"spec: Invalid value:",
				"exactly one of targetRef or targetRefs must be set",
			},
		},
		{
			desc: "targetRef set but targetRefs empty array",
			mutate: func(epp *egv1a1.EnvoyPatchPolicy) {
				epp.Spec.TargetRef = &gwapiv1.LocalPolicyTargetReference{
					Group: gwapiv1.Group("gateway.networking.k8s.io"),
					Kind:  gwapiv1.Kind("Gateway"),
					Name:  gwapiv1.ObjectName("eg"),
				}
				epp.Spec.TargetRefs = []gwapiv1.LocalPolicyTargetReference{}
			},
			wantErrors: []string{},
		},
		{
			desc: "targetRefs with multiple gateways",
			mutate: func(epp *egv1a1.EnvoyPatchPolicy) {
				epp.Spec.TargetRefs = []gwapiv1.LocalPolicyTargetReference{
					{
						Group: gwapiv1.Group("gateway.networking.k8s.io"),
						Kind:  gwapiv1.Kind("Gateway"),
						Name:  gwapiv1.ObjectName("eg1"),
					},
					{
						Group: gwapiv1.Group("gateway.networking.k8s.io"),
						Kind:  gwapiv1.Kind("Gateway"),
						Name:  gwapiv1.ObjectName("eg2"),
					},
				}
			},
			wantErrors: []string{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			epp := baseEPP.DeepCopy()
			epp.Name = fmt.Sprintf("epp-%v", time.Now().UnixNano())

			if tc.mutate != nil {
				tc.mutate(epp)
			}
			err := c.Create(ctx, epp)

			if tc.mutateStatus != nil {
				tc.mutateStatus(epp)
				err = c.Status().Update(ctx, epp)
			}

			if (len(tc.wantErrors) != 0) != (err != nil) {
				t.Fatalf("Unexpected response while creating EnvoyPatchPolicy; got err=\n%v\n;want error=%v", err, tc.wantErrors)
			}

			var missingErrorStrings []string
			for _, wantError := range tc.wantErrors {
				if !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(wantError)) {
					missingErrorStrings = append(missingErrorStrings, wantError)
				}
			}

			if len(missingErrorStrings) != 0 {
				t.Errorf("Unexpected response while creating EnvoyPatchPolicy; got err=\n%v\n;missing strings within error=%q", err, missingErrorStrings)
			}
		})
	}
}
