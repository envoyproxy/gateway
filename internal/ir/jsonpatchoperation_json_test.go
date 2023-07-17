// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package ir

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestJSONPatchOperation_MarshalAndUnMarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   *JSONPatchOperation
		wantErr bool
		want    *JSONPatchOperation
	}{
		{
			name:    "json patch",
			input:   EnvoyPatchPolicyForTest(),
			want:    EnvoyPatchPolicyForTest(),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonPatch, err := json.Marshal(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Marshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			got := &JSONPatchOperation{}
			err = json.Unmarshal(jsonPatch, got)
			if (err != nil) != tt.wantErr {
				t.Errorf("Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got = %v, want %v", got, tt.want)
			}
		})
	}
}

func EnvoyPatchPolicyForTest() *JSONPatchOperation {
	return &JSONPatchOperation{
		Op:   "add",
		Path: "/filterChains/0/filters/0/typedConfig/httpFilters/0",
		Value: `name: envoy.filters.http.oauth2
typed_config:
  '@type': type.googleapis.com/envoy.extensions.filters.http.oauth2.v3.OAuth2Config
  auth_scopes:
  - openid
  auth_type: BASIC_AUTH
  authorization_endpoint: https://accounts.google.com/o/oauth2/v2/auth
  credentials:
    client_id: 250344188863-lddmgbasbdom9qpt1mr0rkln62281s6d.apps.googleusercontent.com
    hmac_secret:
      name: authn_policy1_hmac
      sds_config:
        api_config_source:
          api_type: GRPC
          grpc_services:
          - envoy_grpc:
              cluster_name: teg_sds_cluster
          transport_api_version: V3
    token_secret:
      name: default/client-secret
      sds_config:
        api_config_source:
          api_type: GRPC
          grpc_services:
          - envoy_grpc:
              cluster_name: teg_sds_cluster
          transport_api_version: V3
  forward_bearer_token: true
  redirect_path_matcher:
    path:
      exact: /oauth2/callback
  redirect_uri: https://foo.example.com/oauth2/callback
  signout_path:
    path:
      exact: /signout
  token_endpoint:
    cluster: accounts.google.com_authn_policy1_authn_server
    timeout: 10s
    uri: https://oauth2.googleapis.com/token
`,
	}
}
