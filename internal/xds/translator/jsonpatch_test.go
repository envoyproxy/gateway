// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"testing"
	"time"

	clusterv3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	resourcev3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/protobuf/types/known/durationpb"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

func TestProcessJSONPatchesSpan(t *testing.T) {
	// A zero-valued span is what tells an operator that EnvoyPatchPolicy is not the
	// suspect, so the phase is always recorded, even with nothing to patch.
	t.Run("span reports zero counts when there are no EnvoyPatchPolicies", func(t *testing.T) {
		sr := recordSpans(t)

		require.NoError(t, processJSONPatches(t.Context(), new(types.ResourceVersionTable), nil))

		spans := sr.Ended()
		require.Len(t, spans, 1)
		require.Equal(t, "XdsTranslator.processJSONPatches", spans[0].Name())
		require.ElementsMatch(t, []attribute.KeyValue{
			attribute.Int("envoy-patch-policies.count", 0),
			attribute.Int("json-patches.count", 0),
			attribute.Int("json-patches.applied", 0),
			attribute.Int("json-patches.resource-not-found", 0),
			attribute.Int("json-patches.failed", 0),
		}, spans[0].Attributes())
	})

	t.Run("span reports the outcome of every patch", func(t *testing.T) {
		sr := recordSpans(t)

		tCtx := new(types.ResourceVersionTable)
		require.NoError(t, tCtx.AddXdsResource(resourcev3.ClusterType, &clusterv3.Cluster{
			Name:                 "test-cluster",
			ClusterDiscoveryType: &clusterv3.Cluster_Type{Type: clusterv3.Cluster_STATIC},
			ConnectTimeout:       durationpb.New(10 * time.Second),
		}))

		policies := []*ir.EnvoyPatchPolicy{
			{
				EnvoyPatchPolicyStatus: ir.EnvoyPatchPolicyStatus{
					Name:      "policy",
					Namespace: "default",
					Status:    &gwapiv1.PolicyStatus{},
				},
				JSONPatches: []*ir.JSONPatchConfig{
					// Applied.
					{
						Type: resourcev3.ClusterType,
						Name: "test-cluster",
						Operation: ir.JSONPatchOperation{
							Op:    ir.JSONPatchOpReplace,
							Path:  new("/connect_timeout"),
							Value: &apiextensionsv1.JSON{Raw: []byte(`"30s"`)},
						},
					},
					// The targeted resource does not exist.
					{
						Type: resourcev3.ClusterType,
						Name: "missing-cluster",
						Operation: ir.JSONPatchOperation{
							Op:    ir.JSONPatchOpReplace,
							Path:  new("/connect_timeout"),
							Value: &apiextensionsv1.JSON{Raw: []byte(`"30s"`)},
						},
					},
					// Invalid: a replace operation requires a path or a jsonPath.
					{
						Type: resourcev3.ClusterType,
						Name: "test-cluster",
						Operation: ir.JSONPatchOperation{
							Op:    ir.JSONPatchOpReplace,
							Value: &apiextensionsv1.JSON{Raw: []byte(`"30s"`)},
						},
					},
				},
			},
		}

		// JSONPatch errors are user-facing errors, they are returned but don't fail the translation.
		require.Error(t, processJSONPatches(t.Context(), tCtx, policies))

		spans := sr.Ended()
		require.Len(t, spans, 1)
		require.Equal(t, "XdsTranslator.processJSONPatches", spans[0].Name())
		require.ElementsMatch(t, []attribute.KeyValue{
			attribute.Int("envoy-patch-policies.count", 1),
			attribute.Int("json-patches.count", 3),
			attribute.Int("json-patches.applied", 1),
			attribute.Int("json-patches.resource-not-found", 1),
			attribute.Int("json-patches.failed", 1),
		}, spans[0].Attributes())
	})
}
