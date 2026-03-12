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

func TestEnvoyPatchPolicy(t *testing.T) {
	ctx := context.Background()
	baseEPP := egv1a1.EnvoyPatchPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "epp",
			Namespace: metav1.NamespaceDefault,
		},
		Spec: egv1a1.EnvoyPatchPolicySpec{},
	}

	listenerName := "listener-1"
	patchPath := "/name"

	cases := []struct {
		desc       string
		mutate     func(epp *egv1a1.EnvoyPatchPolicy)
		wantErrors []string
	}{
		{
			desc: "valid json patch with name only",
			mutate: func(epp *egv1a1.EnvoyPatchPolicy) {
				epp.Spec = egv1a1.EnvoyPatchPolicySpec{
					Type: egv1a1.JSONPatchEnvoyPatchType,
					JSONPatches: []egv1a1.EnvoyJSONPatchConfig{
						{
							Type: egv1a1.ListenerEnvoyResourceType,
							Name: &listenerName,
							Operation: egv1a1.JSONPatchOperation{
								Op:   egv1a1.JSONPatchOperationType("test"),
								Path: &patchPath,
								Value: &apiextensionsv1.JSON{
									Raw: []byte(`"listener-1"`),
								},
							},
						},
					},
					TargetRef: gwapiv1.LocalPolicyTargetReference{
						Group: gwapiv1.Group(gwapiv1.GroupName),
						Kind:  gwapiv1.Kind("Gateway"),
						Name:  gwapiv1.ObjectName("eg"),
					},
				}
			},
			wantErrors: []string{},
		},
		{
			desc: "valid json patch with nameSelector only",
			mutate: func(epp *egv1a1.EnvoyPatchPolicy) {
				epp.Spec = egv1a1.EnvoyPatchPolicySpec{
					Type: egv1a1.JSONPatchEnvoyPatchType,
					JSONPatches: []egv1a1.EnvoyJSONPatchConfig{
						{
							Type: egv1a1.ListenerEnvoyResourceType,
							NameSelector: &egv1a1.StringMatch{
								Value: listenerName,
							},
							Operation: egv1a1.JSONPatchOperation{
								Op:   egv1a1.JSONPatchOperationType("test"),
								Path: &patchPath,
								Value: &apiextensionsv1.JSON{
									Raw: []byte(`"listener-1"`),
								},
							},
						},
					},
					TargetRef: gwapiv1.LocalPolicyTargetReference{
						Group: gwapiv1.Group(gwapiv1.GroupName),
						Kind:  gwapiv1.Kind("Gateway"),
						Name:  gwapiv1.ObjectName("eg"),
					},
				}
			},
			wantErrors: []string{},
		},
		{
			desc: "invalid json patch with both name and nameSelector",
			mutate: func(epp *egv1a1.EnvoyPatchPolicy) {
				epp.Spec = egv1a1.EnvoyPatchPolicySpec{
					Type: egv1a1.JSONPatchEnvoyPatchType,
					JSONPatches: []egv1a1.EnvoyJSONPatchConfig{
						{
							Type: egv1a1.ListenerEnvoyResourceType,
							Name: &listenerName,
							NameSelector: &egv1a1.StringMatch{
								Value: listenerName,
							},
							Operation: egv1a1.JSONPatchOperation{
								Op:   egv1a1.JSONPatchOperationType("test"),
								Path: &patchPath,
								Value: &apiextensionsv1.JSON{
									Raw: []byte(`"listener-1"`),
								},
							},
						},
					},
					TargetRef: gwapiv1.LocalPolicyTargetReference{
						Group: gwapiv1.Group(gwapiv1.GroupName),
						Kind:  gwapiv1.Kind("Gateway"),
						Name:  gwapiv1.ObjectName("eg"),
					},
				}
			},
			wantErrors: []string{"only one of name and nameSelector can be specified"},
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
