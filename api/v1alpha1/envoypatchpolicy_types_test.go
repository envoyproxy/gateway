package v1alpha1

import (
	"encoding/json"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	"strings"
	"testing"
)

const jsonPatch = `
{
   "metadata":{
      "name":"authn_policy1.envoygateway.tetrate.io",
      "creationTimestamp":null
   },
   "spec":{
      "type":"JSONPatch",
      "jsonPatches":[
         {
            "type":"type.googleapis.com/envoy.config.listener.v3.Listener",
            "name":"default_gw1_listener1",
            "operation":{
               "op":"add",
               "path":"/filterChains/0/filters/0/typedConfig/httpFilters/0",
               "value":{
                  "name":"envoy.filters.http.oauth2",
                  "typed_config":{
                     "@type":"type.googleapis.com/envoy.extensions.filters.http.oauth2.v3.OAuth2Config",
                     "auth_scopes":[
                        "openid"
                     ],
                     "auth_type":"BASIC_AUTH",
                     "authorization_endpoint":"https://accounts.google.com/o/oauth2/v2/auth",
                     "credentials":{
                        "client_id":"250344188863-lddmgbasbdom9qpt1mr0rkln62281s6d.apps.googleusercontent.com",
                        "hmac_secret":{
                           "name":"authn_policy1_hmac",
                           "sds_config":{
                              "api_config_source":{
                                 "api_type":"GRPC",
                                 "grpc_services":[
                                    {
                                       "envoy_grpc":{
                                          "cluster_name":"teg_sds_cluster"
                                       }
                                    }
                                 ],
                                 "transport_api_version":"V3"
                              }
                           }
                        },
                        "token_secret":{
                           "name":"default/client-secret",
                           "sds_config":{
                              "api_config_source":{
                                 "api_type":"GRPC",
                                 "grpc_services":[
                                    {
                                       "envoy_grpc":{
                                          "cluster_name":"teg_sds_cluster"
                                       }
                                    }
                                 ],
                                 "transport_api_version":"V3"
                              }
                           }
                        }
                     },
                     "forward_bearer_token":true,
                     "redirect_path_matcher":{
                        "path":{
                           "exact":"/oauth2/callback"
                        }
                     },
                     "redirect_uri":"https://foo.example.com/oauth2/callback",
                     "signout_path":{
                        "path":{
                           "exact":"/signout"
                        }
                     },
                     "token_endpoint":{
                        "cluster":"accounts.google.com_authn_policy1_authn_server",
                        "timeout":"10s",
                        "uri":"https://oauth2.googleapis.com/token"
                     }
                  }
               }
            }
         }
      ],
      "targetRef":{
         "group":"gateway.networking.k8s.io",
         "kind":"Gateway",
         "name":"gw1"
      },
      "priority":0
   },
   "status":{
      
   }
}
`

func TestJSONPatchOperation_MarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		patch   EnvoyPatchPolicy
		wantErr bool
		want    []byte
	}{
		{
			name: "yaml patch",
			patch: EnvoyPatchPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name: "authn_policy1.envoygateway.tetrate.io",
				},
				Spec: EnvoyPatchPolicySpec{
					Type: JSONPatchEnvoyPatchType,
					TargetRef: gwapiv1a2.PolicyTargetReference{
						Group: "gateway.networking.k8s.io",
						Kind:  "Gateway",
						Name:  "gw1",
					},
					JSONPatches: []*EnvoyJSONPatchConfig{
						{
							Type: ListenerEnvoyResourceType,
							Name: "default_gw1_listener1",
							Operation: JSONPatchOperation{
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
							},
						},
					},
				},
			},
			wantErr: false,
			want:    []byte(removeSpacesAndNewLines(jsonPatch)),
		},
		{
			name: "json patch",
			patch: EnvoyPatchPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name: "authn_policy1.envoygateway.tetrate.io",
				},
				Spec: EnvoyPatchPolicySpec{
					Type: JSONPatchEnvoyPatchType,
					TargetRef: gwapiv1a2.PolicyTargetReference{
						Group: "gateway.networking.k8s.io",
						Kind:  "Gateway",
						Name:  "gw1",
					},
					JSONPatches: []*EnvoyJSONPatchConfig{
						{
							Type: ListenerEnvoyResourceType,
							Name: "default_gw1_listener1",
							Operation: JSONPatchOperation{
								Op:   "add",
								Path: "/filterChains/0/filters/0/typedConfig/httpFilters/0",
								Value: `{
  "name": "envoy.filters.http.oauth2",
  "typed_config": {
    "@type": "type.googleapis.com/envoy.extensions.filters.http.oauth2.v3.OAuth2Config",
    "auth_scopes": [
      "openid"
    ],
    "auth_type": "BASIC_AUTH",
    "authorization_endpoint": "https://accounts.google.com/o/oauth2/v2/auth",
    "credentials": {
      "client_id": "250344188863-lddmgbasbdom9qpt1mr0rkln62281s6d.apps.googleusercontent.com",
      "hmac_secret": {
        "name": "authn_policy1_hmac",
        "sds_config": {
          "api_config_source": {
            "api_type": "GRPC",
            "grpc_services": [
              {
                "envoy_grpc": {
                  "cluster_name": "teg_sds_cluster"
                }
              }
            ],
            "transport_api_version": "V3"
          }
        }
      },
      "token_secret": {
        "name": "default/client-secret",
        "sds_config": {
          "api_config_source": {
            "api_type": "GRPC",
            "grpc_services": [
              {
                "envoy_grpc": {
                  "cluster_name": "teg_sds_cluster"
                }
              }
            ],
            "transport_api_version": "V3"
          }
        }
      }
    },
    "forward_bearer_token": true,
    "redirect_path_matcher": {
      "path": {
        "exact": "/oauth2/callback"
      }
    },
    "redirect_uri": "https://foo.example.com/oauth2/callback",
    "signout_path": {
      "path": {
        "exact": "/signout"
      }
    },
    "token_endpoint": {
      "cluster": "accounts.google.com_authn_policy1_authn_server",
      "timeout": "10s",
      "uri": "https://oauth2.googleapis.com/token"
    }
  }
}
`,
							},
						},
					},
				},
			},
			wantErr: false,
			want:    []byte(removeSpacesAndNewLines(jsonPatch)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.patch)
			if (err != nil) != tt.wantErr {
				t.Errorf("MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MarshalJSON() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func removeSpacesAndNewLines(old string) string {
	return strings.ReplaceAll(strings.ReplaceAll(old, "\n", ""), " ", "")
}
