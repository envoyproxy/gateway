// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"math"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/ir"
)

func TestInt64ToUint32(t *testing.T) {
	type testCase struct {
		Name    string
		In      int64
		Out     uint32
		Success bool
	}

	testCases := []testCase{
		{
			Name:    "valid",
			In:      1024,
			Out:     1024,
			Success: true,
		},
		{
			Name:    "invalid-underflow",
			In:      -1,
			Out:     0,
			Success: false,
		},
		{
			Name:    "invalid-overflow",
			In:      math.MaxUint32 + 1,
			Out:     0,
			Success: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			out, success := int64ToUint32(tc.In)
			require.Equal(t, tc.Out, out)
			require.Equal(t, tc.Success, success)
		})
	}
}

func TestMakeIrStatusSet(t *testing.T) {
	tests := []struct {
		name string
		in   []egv1a1.HTTPStatus
		want []ir.HTTPStatus
	}{
		{
			name: "no duplicates",
			in:   []egv1a1.HTTPStatus{200, 404},
			want: []ir.HTTPStatus{200, 404},
		},
		{
			name: "with duplicates",
			in:   []egv1a1.HTTPStatus{200, 404, 200},
			want: []ir.HTTPStatus{200, 404},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := makeIrStatusSet(tt.in); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("makeIrStatusSet() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMakeIrTriggerSet(t *testing.T) {
	tests := []struct {
		name string
		in   []egv1a1.TriggerEnum
		want []ir.TriggerEnum
	}{
		{
			name: "no duplicates",
			in:   []egv1a1.TriggerEnum{"5xx", "reset"},
			want: []ir.TriggerEnum{"5xx", "reset"},
		},
		{
			name: "with duplicates",
			in:   []egv1a1.TriggerEnum{"5xx", "reset", "5xx"},
			want: []ir.TriggerEnum{"5xx", "reset"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := makeIrTriggerSet(tt.in); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("makeIrTriggerSet() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_translateRateLimitCost(t *testing.T) {
	for _, tc := range []struct {
		name string
		cost *egv1a1.RateLimitCostSpecifier
		exp  *ir.RateLimitCost
	}{
		{
			name: "number",
			cost: &egv1a1.RateLimitCostSpecifier{Number: ptr.To[uint64](1)},
			exp:  &ir.RateLimitCost{Number: ptr.To[uint64](1)},
		},
		{
			name: "metadata",
			cost: &egv1a1.RateLimitCostSpecifier{Metadata: &egv1a1.RateLimitCostMetadata{Namespace: "something.com", Key: "name"}},
			exp:  &ir.RateLimitCost{Format: ptr.To(`%DYNAMIC_METADATA(something.com:name)%`)},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			act := translateRateLimitCost(tc.cost)
			require.Equal(t, tc.exp, act)
		})
	}
}

func TestBuildHTTPProtocolUpgradeConfig(t *testing.T) {
	cases := []struct {
		name     string
		cfgs     []*egv1a1.ProtocolUpgradeConfig
		expected []string
	}{
		{
			name:     "empty",
			cfgs:     nil,
			expected: nil,
		},
		{
			name: "spdy",
			cfgs: []*egv1a1.ProtocolUpgradeConfig{
				{
					Type: "spdy/3.1",
				},
			},
			expected: []string{"spdy/3.1"},
		},
		{
			name: "websockets-spdy",
			cfgs: []*egv1a1.ProtocolUpgradeConfig{
				{
					Type: "websockets",
				},
				{
					Type: "spdy/3.1",
				},
			},
			expected: []string{"websockets", "spdy/3.1"},
		},
		{
			name: "spdy-websockets",
			cfgs: []*egv1a1.ProtocolUpgradeConfig{
				{
					Type: "spdy/3.1",
				},
				{
					Type: "websockets",
				},
			},
			expected: []string{"spdy/3.1", "websockets"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := buildHTTPProtocolUpgradeConfig(tc.cfgs)
			require.Equal(t, tc.expected, got)
		})
	}
}

func TestBuildResponseOverride(t *testing.T) {
	// Helper function to create a basic policy with TypeMeta
	createPolicy := func(name string, overrides []*egv1a1.ResponseOverride) *egv1a1.BackendTrafficPolicy {
		return &egv1a1.BackendTrafficPolicy{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "gateway.envoyproxy.io/v1alpha1",
				Kind:       "BackendTrafficPolicy",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: "default",
			},
			Spec: egv1a1.BackendTrafficPolicySpec{
				ResponseOverride: overrides,
			},
		}
	}

	tests := []struct {
		name     string
		policy   *egv1a1.BackendTrafficPolicy
		expected *ir.ResponseOverride
		wantErr  bool
	}{
		{
			name: "WithResponseHeadersToAddAndJsonFormat",
			policy: createPolicy("test-policy", []*egv1a1.ResponseOverride{
				{
					Match: egv1a1.CustomResponseMatch{
						StatusCodes: []egv1a1.StatusCodeMatch{{Type: ptr.To(egv1a1.StatusCodeValueTypeValue), Value: ptr.To(429)}},
					},
					Response: &egv1a1.CustomResponse{
						StatusCode:  ptr.To(429),
						ContentType: ptr.To("application/json"),
						ResponseHeadersToAdd: []egv1a1.ResponseHeaderToAdd{
							{Name: "X-RateLimit-Limit", Value: "100", Append: ptr.To(false)},
							{Name: "Retry-After", Value: "60", Append: ptr.To(false)},
						},
						BodyFormat: &egv1a1.ResponseBodyFormat{
							JSONFormat:  map[string]string{"error": "rate_limit_exceeded", "status_code": "%RESPONSE_CODE%"},
							ContentType: ptr.To("application/json; charset=utf-8"),
						},
					},
				},
			}),
			expected: &ir.ResponseOverride{
				Name: "backendtrafficpolicy/default/test-policy",
				Rules: []ir.ResponseOverrideRule{
					{
						Name: "backendtrafficpolicy/default/test-policy/responseoverride/rule/0",
						Match: ir.CustomResponseMatch{
							StatusCodes: []ir.StatusCodeMatch{{Value: ptr.To(429)}},
						},
						Response: ir.CustomResponse{
							StatusCode:  ptr.To(uint32(429)),
							ContentType: ptr.To("application/json"),
							ResponseHeadersToAdd: []ir.AddHeader{
								{Name: "X-RateLimit-Limit", Value: []string{"100"}, Append: false},
								{Name: "Retry-After", Value: []string{"60"}, Append: false},
							},
							BodyFormat: &ir.ResponseBodyFormat{
								JSONFormat:  map[string]string{"error": "rate_limit_exceeded", "status_code": "%RESPONSE_CODE%"},
								ContentType: ptr.To("application/json; charset=utf-8"),
							},
						},
					},
				},
			},
		},
		{
			name: "WithBodyFormatTextFormat",
			policy: createPolicy("test-policy", []*egv1a1.ResponseOverride{
				{
					Match: egv1a1.CustomResponseMatch{
						StatusCodes: []egv1a1.StatusCodeMatch{{Type: ptr.To(egv1a1.StatusCodeValueTypeRange), Range: &egv1a1.StatusCodeRange{Start: 400, End: 499}}},
					},
					Response: &egv1a1.CustomResponse{
						StatusCode: ptr.To(400),
						ResponseHeadersToAdd: []egv1a1.ResponseHeaderToAdd{
							{Name: "X-Error-Type", Value: "client-error", Append: ptr.To(false)},
						},
						BodyFormat: &egv1a1.ResponseBodyFormat{
							TextFormat:  ptr.To("Error %RESPONSE_CODE%: %LOCAL_REPLY_BODY%"),
							ContentType: ptr.To("text/plain"),
						},
					},
				},
			}),
			expected: &ir.ResponseOverride{
				Name: "backendtrafficpolicy/default/test-policy",
				Rules: []ir.ResponseOverrideRule{
					{
						Name: "backendtrafficpolicy/default/test-policy/responseoverride/rule/0",
						Match: ir.CustomResponseMatch{
							StatusCodes: []ir.StatusCodeMatch{{Range: &ir.StatusCodeRange{Start: 400, End: 499}}},
						},
						Response: ir.CustomResponse{
							StatusCode: ptr.To(uint32(400)),
							ResponseHeadersToAdd: []ir.AddHeader{
								{Name: "X-Error-Type", Value: []string{"client-error"}, Append: false},
							},
							BodyFormat: &ir.ResponseBodyFormat{
								TextFormat:  ptr.To("Error %RESPONSE_CODE%: %LOCAL_REPLY_BODY%"),
								ContentType: ptr.To("text/plain"),
							},
						},
					},
				},
			},
		},
		{
			name: "WithHeaderAppendBehavior",
			policy: createPolicy("test-policy", []*egv1a1.ResponseOverride{
				{
					Match: egv1a1.CustomResponseMatch{
						StatusCodes: []egv1a1.StatusCodeMatch{{Type: ptr.To(egv1a1.StatusCodeValueTypeValue), Value: ptr.To(500)}},
					},
					Response: &egv1a1.CustomResponse{
						StatusCode: ptr.To(503),
						ResponseHeadersToAdd: []egv1a1.ResponseHeaderToAdd{
							{Name: "Cache-Control", Value: "no-cache", Append: ptr.To(true)},
							{Name: "X-Error-Source", Value: "backend-service", Append: ptr.To(false)},
						},
						Body: &egv1a1.CustomResponseBody{Type: ptr.To(egv1a1.ResponseValueTypeInline), Inline: ptr.To("Service unavailable")},
					},
				},
			}),
			expected: &ir.ResponseOverride{
				Name: "backendtrafficpolicy/default/test-policy",
				Rules: []ir.ResponseOverrideRule{
					{
						Name: "backendtrafficpolicy/default/test-policy/responseoverride/rule/0",
						Match: ir.CustomResponseMatch{
							StatusCodes: []ir.StatusCodeMatch{{Value: ptr.To(500)}},
						},
						Response: ir.CustomResponse{
							StatusCode: ptr.To(uint32(503)),
							Body:       ptr.To("Service unavailable"),
							ResponseHeadersToAdd: []ir.AddHeader{
								{Name: "Cache-Control", Value: []string{"no-cache"}, Append: true},
								{Name: "X-Error-Source", Value: []string{"backend-service"}, Append: false},
							},
						},
					},
				},
			},
		},
		{
			name: "WithConfigMapBodyAndBodyFormatOverride",
			policy: createPolicy("test-policy", []*egv1a1.ResponseOverride{
				{
					Match: egv1a1.CustomResponseMatch{
						StatusCodes: []egv1a1.StatusCodeMatch{{Type: ptr.To(egv1a1.StatusCodeValueTypeValue), Value: ptr.To(503)}},
					},
					Response: &egv1a1.CustomResponse{
						Body: &egv1a1.CustomResponseBody{
							Type:     ptr.To(egv1a1.ResponseValueTypeValueRef),
							ValueRef: &gwapiv1a2.LocalObjectReference{Kind: "ConfigMap", Name: "custom-error-responses"},
						},
						ResponseHeadersToAdd: []egv1a1.ResponseHeaderToAdd{
							{Name: "X-Custom-Response", Value: "true", Append: ptr.To(false)},
						},
						BodyFormat: &egv1a1.ResponseBodyFormat{
							JSONFormat:  map[string]string{"original_response": "%LOCAL_REPLY_BODY%", "enhanced_status": "%RESPONSE_CODE%"},
							ContentType: ptr.To("application/json"),
						},
					},
				},
			}),
			expected: &ir.ResponseOverride{
				Name: "backendtrafficpolicy/default/test-policy",
				Rules: []ir.ResponseOverrideRule{
					{
						Name: "backendtrafficpolicy/default/test-policy/responseoverride/rule/0",
						Match: ir.CustomResponseMatch{
							StatusCodes: []ir.StatusCodeMatch{{Value: ptr.To(503)}},
						},
						Response: ir.CustomResponse{
							Body: ptr.To(`{"error": "Service temporarily unavailable", "support_contact": "support@example.com"}`),
							ResponseHeadersToAdd: []ir.AddHeader{
								{Name: "X-Custom-Response", Value: []string{"true"}, Append: false},
							},
							BodyFormat: &ir.ResponseBodyFormat{
								JSONFormat:  map[string]string{"original_response": "%LOCAL_REPLY_BODY%", "enhanced_status": "%RESPONSE_CODE%"},
								ContentType: ptr.To("application/json"),
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resources := resource.NewResources()
			if tt.name == "WithConfigMapBodyAndBodyFormatOverride" {
				resources.ConfigMaps = append(resources.ConfigMaps, &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{Name: "custom-error-responses", Namespace: "default"},
					Data:       map[string]string{"response.body": `{"error": "Service temporarily unavailable", "support_contact": "support@example.com"}`},
				})
			}

			result, err := buildResponseOverride(tt.policy, resources)

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tt.expected.Name, result.Name)
				require.Len(t, result.Rules, len(tt.expected.Rules))

				for i, expectedRule := range tt.expected.Rules {
					actualRule := result.Rules[i]
					assert.Equal(t, expectedRule.Name, actualRule.Name)
					assert.Equal(t, expectedRule.Match, actualRule.Match)
					assert.Equal(t, expectedRule.Response.StatusCode, actualRule.Response.StatusCode)
					assert.Equal(t, expectedRule.Response.ContentType, actualRule.Response.ContentType)
					assert.Equal(t, expectedRule.Response.Body, actualRule.Response.Body)
					assert.Equal(t, expectedRule.Response.ResponseHeadersToAdd, actualRule.Response.ResponseHeadersToAdd)

					if expectedRule.Response.BodyFormat != nil {
						require.NotNil(t, actualRule.Response.BodyFormat)
						assert.Equal(t, expectedRule.Response.BodyFormat.JSONFormat, actualRule.Response.BodyFormat.JSONFormat)
						assert.Equal(t, expectedRule.Response.BodyFormat.TextFormat, actualRule.Response.BodyFormat.TextFormat)
						assert.Equal(t, expectedRule.Response.BodyFormat.ContentType, actualRule.Response.BodyFormat.ContentType)
					} else {
						assert.Nil(t, actualRule.Response.BodyFormat)
					}
				}
			}
		})
	}
}
