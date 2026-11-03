// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build celvalidation

package celvalidation

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/util/validation/field"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
)

// TestRateLimitServiceSettings checks that EnvoyGateway.RateLimit.ServiceSettings, which is
// validated in code since EnvoyGateway isn't admitted as a CRD, accepts and rejects the same
// ClusterSettings as the API server does. It reuses the BackendTrafficPolicy cases, keeping
// only their ClusterSettings and admitting them as a BackendTrafficPolicy with a valid target.
func TestRateLimitServiceSettings(t *testing.T) {
	ctx := context.Background()

	for _, tc := range backendTrafficPolicyCases() {
		if tc.mutate == nil {
			continue
		}
		in := baseBTP.DeepCopy()
		tc.mutate(in)
		cs := in.Spec.ClusterSettings
		if reflect.DeepEqual(cs, egv1a1.ClusterSettings{}) {
			continue
		}

		t.Run(tc.desc, func(t *testing.T) {
			btp := baseBTP.DeepCopy()
			btp.Name = fmt.Sprintf("btp-%v", time.Now().UnixNano())
			btp.Spec.TargetRefs = []gwapiv1.LocalPolicyTargetReferenceWithSectionName{{
				LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
					Group: "gateway.networking.k8s.io",
					Kind:  "Gateway",
					Name:  "eg",
				},
			}}
			btp.Spec.ClusterSettings = *cs.DeepCopy()
			apiServerErr := c.Create(ctx, btp)

			inCodeErr := config.ValidateClusterSettings(field.NewPath("spec"), &cs)

			if (apiServerErr == nil) != (inCodeErr == nil) {
				t.Fatalf("API server and in-code validation disagree:\nAPI server: %v\nin-code:    %v", apiServerErr, inCodeErr)
			}
		})
	}
}
