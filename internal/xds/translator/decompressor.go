// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"errors"
	"fmt"
	"strings"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	brotlidecompressorv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/compression/brotli/decompressor/v3"
	gzipdecompressorv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/compression/gzip/decompressor/v3"
	zstddecompressorv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/compression/zstd/decompressor/v3"
	decompressorv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/decompressor/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	protobuf "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/utils/proto"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

func init() {
	registerHTTPFilter(&decompressor{})
}

type decompressor struct{}

var _ httpFilter = &decompressor{}

// patchHCM builds and appends the decompressor Filter to the HTTP Connection Manager
// if applicable, and it does not already exist.
func (*decompressor) patchHCM(mgr *hcmv3.HttpConnectionManager, irListener *ir.HTTPListener) error {
	if mgr == nil {
		return errors.New("hcm is nil")
	}
	if irListener == nil {
		return errors.New("ir listener is nil")
	}

	// Return early if no decompression is configured
	if len(irListener.Decompression) == 0 {
		return nil
	}

	var (
		filter *hcmv3.HttpFilter
		err    error
	)

	// Process each decompression type
	for _, irDecomp := range irListener.Decompression {
		filterName := decompressorFilterName(irDecomp.Type)

		// Skip if filter already exists
		if hcmContainsFilter(mgr, filterName) {
			continue
		}

		// Extract window bits for gzip if configured
		var windowBits *uint32
		if irDecomp.Type == egv1a1.GzipDecompressorType && irDecomp.Gzip != nil {
			windowBits = irDecomp.Gzip.WindowBits
		}

		if filter, err = buildDecompressorFilter(irDecomp.Type, windowBits); err != nil {
			return err
		}

		mgr.HttpFilters = append(mgr.HttpFilters, filter)
	}

	return nil
}

func (*decompressor) patchResources(*types.ResourceVersionTable, []*ir.HTTPRoute) error {
	return nil
}

// patchRoute enables the decompressor filters for all routes when decompression is configured globally.
func (*decompressor) patchRoute(route *routev3.Route, irRoute *ir.HTTPRoute, _ *ir.HTTPListener) error {
	// Decompressor is a listener-level filter, no per-route configuration needed
	return nil
}

func decompressorFilterName(decompressorType egv1a1.DecompressorType) string {
	return fmt.Sprintf("%s.%s", egv1a1.EnvoyFilterDecompressor.String(), strings.ToLower(string(decompressorType)))
}

// buildDecompressorFilter builds a decompressor filter with the provided decompressionType.
func buildDecompressorFilter(decompressionType egv1a1.DecompressorType, windowBits *uint32) (*hcmv3.HttpFilter, error) {
	var (
		decompressorProto *decompressorv3.Decompressor
		extensionName     string
		extensionMsg      protobuf.Message
		extensionAny      *anypb.Any
		decompressorAny   *anypb.Any
		err               error
	)

	switch decompressionType {
	case egv1a1.BrotliDecompressorType:
		extensionName = "envoy.compression.brotli.decompressor"
		extensionMsg = &brotlidecompressorv3.Brotli{}
	case egv1a1.GzipDecompressorType:
		extensionName = "envoy.compression.gzip.decompressor"
		gzipMsg := &gzipdecompressorv3.Gzip{}
		if windowBits != nil {
			gzipMsg.WindowBits = wrapperspb.UInt32(*windowBits)
		}
		extensionMsg = gzipMsg
	case egv1a1.ZstdDecompressorType:
		extensionName = "envoy.compression.zstd.decompressor"
		extensionMsg = &zstddecompressorv3.Zstd{}
	default:
		return nil, fmt.Errorf("unsupported decompressor type: %s", decompressionType)
	}

	if extensionAny, err = proto.ToAnyWithValidation(extensionMsg); err != nil {
		return nil, err
	}

	decompressorProto = &decompressorv3.Decompressor{
		DecompressorLibrary: &corev3.TypedExtensionConfig{
			Name:        extensionName,
			TypedConfig: extensionAny,
		},
		// When this filter is added (only when explicitly configured in ClientTrafficPolicy),
		// enable decompression for both request and response directions
		RequestDirectionConfig: &decompressorv3.Decompressor_RequestDirectionConfig{
			CommonConfig: &decompressorv3.Decompressor_CommonDirectionConfig{
				Enabled: &corev3.RuntimeFeatureFlag{
					DefaultValue: wrapperspb.Bool(true),
				},
			},
		},
		ResponseDirectionConfig: &decompressorv3.Decompressor_ResponseDirectionConfig{
			CommonConfig: &decompressorv3.Decompressor_CommonDirectionConfig{
				Enabled: &corev3.RuntimeFeatureFlag{
					DefaultValue: wrapperspb.Bool(true),
				},
			},
		},
	}

	if decompressorAny, err = proto.ToAnyWithValidation(decompressorProto); err != nil {
		return nil, err
	}

	return &hcmv3.HttpFilter{
		Name: decompressorFilterName(decompressionType),
		ConfigType: &hcmv3.HttpFilter_TypedConfig{
			TypedConfig: decompressorAny,
		},
	}, nil
}
