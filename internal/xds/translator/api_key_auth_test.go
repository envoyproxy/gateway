// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"testing"

	apikeyauthv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/api_key_auth/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	"github.com/stretchr/testify/require"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
)

// TestPatchHCMWithAPIKeyAuth asserts that the listener level filter carries no configuration
// of its own, so that a credential change on one route stays a route level update instead of
// rewriting the listener.
func TestPatchHCMWithAPIKeyAuth(t *testing.T) {
	listener := &ir.HTTPListener{
		Name: "listener-1",
		Routes: []*ir.HTTPRoute{
			{
				Name: "route-1",
				Security: &ir.SecurityFeatures{
					APIKeyAuth: &ir.APIKeyAuth{
						Credentials:           []ir.APIKeyCredential{{Client: []byte("client-1"), Key: []byte("key1")}},
						ExtractFrom:           []*ir.ExtractFrom{{Headers: []string{"X-API-KEY"}}},
						ForwardClientIDHeader: new("X-API-KEY-CLIENT-ID"),
						Sanitize:              new(true),
					},
				},
			},
			{
				Name: "route-2",
				Security: &ir.SecurityFeatures{
					APIKeyAuth: &ir.APIKeyAuth{
						Credentials: []ir.APIKeyCredential{{Client: []byte("client-2"), Key: []byte("key2")}},
						ExtractFrom: []*ir.ExtractFrom{{Params: []string{"api-key"}}},
					},
				},
			},
		},
	}

	mgr := &hcmv3.HttpConnectionManager{}
	require.NoError(t, (&apiKeyAuth{}).patchHCM(mgr, listener))

	require.Len(t, mgr.HttpFilters, 1)
	filter := mgr.HttpFilters[0]
	require.Equal(t, egv1a1.EnvoyFilterAPIKeyAuth.String(), filter.Name)
	require.True(t, filter.Disabled)

	cfg := &apikeyauthv3.ApiKeyAuth{}
	require.NoError(t, filter.GetTypedConfig().UnmarshalTo(cfg))
	require.Empty(t, cfg.Credentials, "route credentials must not be published in the listener")
	require.Empty(t, cfg.KeySources, "key sources are per route and must not be inherited")
	require.Nil(t, cfg.Forwarding, "forwarding is per route and must not be published in the listener")
}

func TestPatchHCMWithAPIKeyAuthNotNeeded(t *testing.T) {
	mgr := &hcmv3.HttpConnectionManager{}
	listener := &ir.HTTPListener{
		Name:   "listener-1",
		Routes: []*ir.HTTPRoute{{Name: "route-1"}},
	}

	require.NoError(t, (&apiKeyAuth{}).patchHCM(mgr, listener))
	require.Empty(t, mgr.HttpFilters, "no route uses api key auth")
}

func TestPatchHCMWithAPIKeyAuthAlreadyPresent(t *testing.T) {
	mgr := &hcmv3.HttpConnectionManager{
		HttpFilters: []*hcmv3.HttpFilter{{Name: egv1a1.EnvoyFilterAPIKeyAuth.String()}},
	}
	listener := &ir.HTTPListener{
		Name: "listener-1",
		Routes: []*ir.HTTPRoute{
			{
				Name: "route-1",
				Security: &ir.SecurityFeatures{
					APIKeyAuth: &ir.APIKeyAuth{
						Credentials: []ir.APIKeyCredential{{Client: []byte("client-1"), Key: []byte("key1")}},
						ExtractFrom: []*ir.ExtractFrom{{Headers: []string{"X-API-KEY"}}},
					},
				},
			},
		},
	}

	require.NoError(t, (&apiKeyAuth{}).patchHCM(mgr, listener))
	require.Len(t, mgr.HttpFilters, 1, "the filter must not be appended twice")
}
