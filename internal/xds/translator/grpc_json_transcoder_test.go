// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"encoding/base64"
	"testing"

	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	grpcjsontranscoder "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/grpc_json_transcoder/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	"github.com/stretchr/testify/require"

	"github.com/envoyproxy/gateway/internal/ir"
)

// grpcEchoDescriptor is the FileDescriptorSet for the conformance grpcecho service.
const grpcEchoDescriptor = "CoAMCg5ncnBjZWNoby5wcm90bxIrZ2F0ZXdheV9hcGlfY29uZm9ybWFuY2UuZWNob19iYXNpYy5ncnBjZWNobxocZ29vZ2xlL2FwaS9hbm5vdGF0aW9ucy5wcm90byIwCgZIZWFkZXISEAoDa2V5GAEgASgJUgNrZXkSFAoFdmFsdWUYAiABKAlSBXZhbHVlInYKB0NvbnRleHQSHAoJbmFtZXNwYWNlGAEgASgJUgluYW1lc3BhY2USGAoHaW5ncmVzcxgCIAEoCVIHaW5ncmVzcxIhCgxzZXJ2aWNlX25hbWUYAyABKAlSC3NlcnZpY2VOYW1lEhAKA3BvZBgEIAEoCVIDcG9kIssBCg1UTFNBc3NlcnRpb25zEhgKB3ZlcnNpb24YASABKAlSB3ZlcnNpb24SLwoTbmVnb3RpYXRlZF9wcm90b2NvbBgCIAEoCVISbmVnb3RpYXRlZFByb3RvY29sEh8KC3NlcnZlcl9uYW1lGAMgASgJUgpzZXJ2ZXJOYW1lEiEKDGNpcGhlcl9zdWl0ZRgEIAEoCVILY2lwaGVyU3VpdGUSKwoRcGVlcl9jZXJ0aWZpY2F0ZXMYBSADKAlSEHBlZXJDZXJ0aWZpY2F0ZXMi4gIKCkFzc2VydGlvbnMSNAoWZnVsbHlfcXVhbGlmaWVkX21ldGhvZBgBIAEoCVIUZnVsbHlRdWFsaWZpZWRNZXRob2QSTQoHaGVhZGVycxgCIAMoCzIzLmdhdGV3YXlfYXBpX2NvbmZvcm1hbmNlLmVjaG9fYmFzaWMuZ3JwY2VjaG8uSGVhZGVyUgdoZWFkZXJzEhwKCWF1dGhvcml0eRgDIAEoCVIJYXV0aG9yaXR5Ek4KB2NvbnRleHQYBCABKAsyNC5nYXRld2F5X2FwaV9jb25mb3JtYW5jZS5lY2hvX2Jhc2ljLmdycGNlY2hvLkNvbnRleHRSB2NvbnRleHQSYQoOdGxzX2Fzc2VydGlvbnMYBSABKAsyOi5nYXRld2F5X2FwaV9jb25mb3JtYW5jZS5lY2hvX2Jhc2ljLmdycGNlY2hvLlRMU0Fzc2VydGlvbnNSDXRsc0Fzc2VydGlvbnMiDQoLRWNob1JlcXVlc3QiuwEKDEVjaG9SZXNwb25zZRJXCgphc3NlcnRpb25zGAEgASgLMjcuZ2F0ZXdheV9hcGlfY29uZm9ybWFuY2UuZWNob19iYXNpYy5ncnBjZWNoby5Bc3NlcnRpb25zUgphc3NlcnRpb25zElIKB3JlcXVlc3QYAiABKAsyOC5nYXRld2F5X2FwaV9jb25mb3JtYW5jZS5lY2hvX2Jhc2ljLmdycGNlY2hvLkVjaG9SZXF1ZXN0UgdyZXF1ZXN0Mq8DCghHcnBjRWNobxJ9CgRFY2hvEjguZ2F0ZXdheV9hcGlfY29uZm9ybWFuY2UuZWNob19iYXNpYy5ncnBjZWNoby5FY2hvUmVxdWVzdBo5LmdhdGV3YXlfYXBpX2NvbmZvcm1hbmNlLmVjaG9fYmFzaWMuZ3JwY2VjaG8uRWNob1Jlc3BvbnNlIgASngEKB0VjaG9Ud28SOC5nYXRld2F5X2FwaV9jb25mb3JtYW5jZS5lY2hvX2Jhc2ljLmdycGNlY2hvLkVjaG9SZXF1ZXN0GjkuZ2F0ZXdheV9hcGlfY29uZm9ybWFuY2UuZWNob19iYXNpYy5ncnBjZWNoby5FY2hvUmVzcG9uc2UiHoLT5JMCGBIWL3YxL2dycGMtZWNoby9lY2hvLXR3bxKCAQoJRWNob1RocmVlEjguZ2F0ZXdheV9hcGlfY29uZm9ybWFuY2UuZWNob19iYXNpYy5ncnBjZWNoby5FY2hvUmVxdWVzdBo5LmdhdGV3YXlfYXBpX2NvbmZvcm1hbmNlLmVjaG9fYmFzaWMuZ3JwY2VjaG8uRWNob1Jlc3BvbnNlIgBCP1o9c2lncy5rOHMuaW8vZ2F0ZXdheS1hcGkvY29uZm9ybWFuY2UvZWNoby1iYXNpYy9ncnBjZWNob3NlcnZlcmIGcHJvdG8z"

func mustDescriptor(t *testing.T) []byte {
	t.Helper()
	bin, err := base64.StdEncoding.DecodeString(grpcEchoDescriptor)
	require.NoError(t, err)
	return bin
}

func testTranscoder(t *testing.T) *ir.GRPCJSONTranscoder {
	return &ir.GRPCJSONTranscoder{
		Name:               "httproutefilter/default/transcode",
		ProtoDescriptorBin: mustDescriptor(t),
		Services:           []string{"gateway_api_conformance.echo_basic.grpcecho.GrpcEcho"},
	}
}

// Guards the bug that made Envoy reject the listener: an empty HCM-level filter config,
// when descriptor_set is a required oneof.
func TestPatchHCMGRPCJSONTranscoderSetsDescriptor(t *testing.T) {
	transcoder := testTranscoder(t)
	mgr := &hcmv3.HttpConnectionManager{}
	listener := &ir.HTTPListener{
		Routes: []*ir.HTTPRoute{{
			Name:               "route-1",
			GRPCJSONTranscoder: transcoder,
		}},
	}

	require.NoError(t, (&grpcJSONTranscoder{}).patchHCM(mgr, listener))
	require.Len(t, mgr.HttpFilters, 1)

	filter := mgr.HttpFilters[0]
	require.Equal(t, "envoy.filters.http.grpc_json_transcoder/httproutefilter/default/transcode", filter.Name)
	require.True(t, filter.Disabled, "filter must be disabled by default so untargeted routes are unaffected")

	got := &grpcjsontranscoder.GrpcJsonTranscoder{}
	require.NoError(t, filter.GetTypedConfig().UnmarshalTo(got))
	require.NoError(t, got.ValidateAll(), "listener-level config must pass Envoy proto validation")
	require.Equal(t, mustDescriptor(t), got.GetProtoDescriptorBin())
	require.Equal(t, transcoder.Services, got.GetServices())
}

func TestPatchHCMGRPCJSONTranscoderDedupes(t *testing.T) {
	transcoder := testTranscoder(t)
	mgr := &hcmv3.HttpConnectionManager{}
	listener := &ir.HTTPListener{
		Routes: []*ir.HTTPRoute{
			{Name: "route-1", GRPCJSONTranscoder: transcoder},
			{Name: "route-2", GRPCJSONTranscoder: transcoder},
			{Name: "route-3"},
		},
	}

	require.NoError(t, (&grpcJSONTranscoder{}).patchHCM(mgr, listener))
	require.Len(t, mgr.HttpFilters, 1)
}

func TestPatchHCMGRPCJSONTranscoderNoPolicy(t *testing.T) {
	mgr := &hcmv3.HttpConnectionManager{}
	listener := &ir.HTTPListener{Routes: []*ir.HTTPRoute{{Name: "route-1"}}}

	require.NoError(t, (&grpcJSONTranscoder{}).patchHCM(mgr, listener))
	require.Empty(t, mgr.HttpFilters)
}

func TestPatchRouteGRPCJSONTranscoder(t *testing.T) {
	transcoder := testTranscoder(t)
	route := &routev3.Route{}
	irRoute := &ir.HTTPRoute{Name: "route-1", GRPCJSONTranscoder: transcoder}

	require.NoError(t, (&grpcJSONTranscoder{}).patchRoute(route, irRoute, nil))

	cfg := route.GetTypedPerFilterConfig()["envoy.filters.http.grpc_json_transcoder/httproutefilter/default/transcode"]
	require.NotNil(t, cfg, "per-route config must use the HCM filter instance name")

	got := &routev3.FilterConfig{}
	require.NoError(t, cfg.UnmarshalTo(got))
	require.NotContains(t, string(cfg.GetValue()), string(mustDescriptor(t)),
		"the descriptor belongs on the HCM filter only, not duplicated per route")
}

func TestBuildGRPCJSONTranscoderConfigOptions(t *testing.T) {
	transcoder := testTranscoder(t)
	transcoder.PrintOptions = &ir.JSONPrintOptions{
		AddWhitespace:           new(true),
		PreserveProtoFieldNames: new(false),
	}
	transcoder.AutoMapping = new(true)

	cfg, err := buildGRPCJSONTranscoderConfig(transcoder)
	require.NoError(t, err)

	got := &grpcjsontranscoder.GrpcJsonTranscoder{}
	require.NoError(t, cfg.UnmarshalTo(got))
	require.NoError(t, got.ValidateAll())
	require.True(t, got.GetPrintOptions().GetAddWhitespace())
	require.False(t, got.GetPrintOptions().GetPreserveProtoFieldNames())
	require.False(t, got.GetPrintOptions().GetAlwaysPrintEnumsAsInts(), "unset pointer must mean false, not true")
	require.True(t, got.GetAutoMapping())
	require.False(t, got.GetConvertGrpcStatus())
}
