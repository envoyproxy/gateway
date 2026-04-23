// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"fmt"
	"sort"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	tracecfg "github.com/envoyproxy/go-control-plane/envoy/config/trace/v3"
	hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	resourcedetectorsv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/tracers/opentelemetry/resource_detectors/v3"
	samplersv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/tracers/opentelemetry/samplers/v3"
	tracingtype "github.com/envoyproxy/go-control-plane/envoy/type/tracing/v3"
	xdstype "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"k8s.io/utils/ptr"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/utils/proto"
	"github.com/envoyproxy/gateway/internal/xds/types"
	"github.com/envoyproxy/gateway/internal/xds/utils/fractionalpercent"
)

const (
	envoyOpenTelemetry = "envoy.tracers.opentelemetry"
	envoyZipkin        = "envoy.traces.zipkin"
	envoyDatadog       = "envoy.tracers.datadog"
)

type typConfigGen func() (*anypb.Any, error)

func buildHCMTracing(tracing *ir.Tracing) (*hcm.HttpConnectionManager_Tracing, error) {
	if tracing == nil {
		return nil, nil
	}

	var providerName string
	var providerConfig typConfigGen

	switch tracing.Provider.Type {
	case egv1a1.TracingProviderTypeDatadog:
		providerName = envoyDatadog

		providerConfig = func() (*anypb.Any, error) {
			config := &tracecfg.DatadogConfig{
				ServiceName:      tracing.ServiceName,
				CollectorCluster: tracing.Destination.Name,
			}
			return proto.ToAnyWithValidation(config)
		}
	case egv1a1.TracingProviderTypeOpenTelemetry:
		providerName = envoyOpenTelemetry

		providerConfig = func() (*anypb.Any, error) {
			var otelSampler *egv1a1.OTelSampler
			if tracing.Provider.OpenTelemetry != nil {
				otelSampler = tracing.Provider.OpenTelemetry.Sampler
			}
			sampler, sErr := buildSampler(otelSampler)
			if sErr != nil {
				return nil, sErr
			}
			config := &tracecfg.OpenTelemetryConfig{
				GrpcService: &corev3.GrpcService{
					TargetSpecifier: &corev3.GrpcService_EnvoyGrpc_{
						EnvoyGrpc: &corev3.GrpcService_EnvoyGrpc{
							ClusterName: tracing.Destination.Name,
							Authority:   tracing.Authority,
						},
					},
					InitialMetadata: buildGrpcInitialMetadata(tracing.Headers),
				},
				ServiceName:       tracing.ServiceName,
				ResourceDetectors: buildResourceDetectors(tracing.ResourceAttributes),
				Sampler:           sampler,
			}

			return proto.ToAnyWithValidation(config)
		}
	case egv1a1.TracingProviderTypeZipkin:
		providerName = envoyZipkin

		providerConfig = func() (*anypb.Any, error) {
			config := &tracecfg.ZipkinConfig{
				CollectorCluster:         tracing.Destination.Name,
				CollectorEndpoint:        "/api/v2/spans",
				TraceId_128Bit:           ptr.Deref(tracing.Provider.Zipkin.Enable128BitTraceID, false),
				SharedSpanContext:        wrapperspb.Bool(!ptr.Deref(tracing.Provider.Zipkin.DisableSharedSpanContext, false)),
				CollectorEndpointVersion: tracecfg.ZipkinConfig_HTTP_JSON,
			}

			return proto.ToAnyWithValidation(config)
		}
	default:
		return nil, fmt.Errorf("unknown tracing provider type: %s", tracing.Provider.Type)
	}

	ocAny, err := providerConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal tracing configuration: %w", err)
	}

	tags, err := buildTracingTags(tracing.CustomTags, tracing.Tags)
	if err != nil {
		return nil, fmt.Errorf("failed to build tracing tags: %w", err)
	}

	op, upstreamOp := buildTracingOperation(tracing.SpanName)

	return &hcm.HttpConnectionManager_Tracing{
		ClientSampling: &xdstype.Percent{
			Value: 100.0,
		},
		OverallSampling: &xdstype.Percent{
			Value: 100.0,
		},
		RandomSampling: &xdstype.Percent{
			Value: randomSamplingValue(tracing),
		},
		Provider: &tracecfg.Tracing_Http{
			Name: providerName,
			ConfigType: &tracecfg.Tracing_Http_TypedConfig{
				TypedConfig: ocAny,
			},
		},
		CustomTags:        tags,
		SpawnUpstreamSpan: wrapperspb.Bool(true),
		Operation:         op,
		UpstreamOperation: upstreamOp,
	}, nil
}

// randomSamplingValue returns the HCM RandomSampling percentage. When an OTel
// sampler is configured, this returns 100% so the sampler alone decides whether
// to record spans. Otherwise, HCM would drop requests before the sampler runs.
func randomSamplingValue(tracing *ir.Tracing) float64 {
	if tracing.Provider.OpenTelemetry != nil && tracing.Provider.OpenTelemetry.Sampler != nil {
		return 100.0
	}
	return tracing.SamplingRate
}

func buildTracingOperation(span *egv1a1.TracingSpanName) (string, string) {
	if span == nil {
		return "", ""
	}

	return span.Client, span.Server
}

func processClusterForTracing(tCtx *types.ResourceVersionTable, tracing *ir.Tracing, metrics *ir.Metrics) error {
	if tracing == nil {
		return nil
	}

	args := &xdsClusterArgs{
		name:         tracing.Destination.Name,
		settings:     tracing.Destination.Settings,
		tSocket:      nil,
		endpointType: buildEndpointType(tracing.Destination.Settings),
		metrics:      metrics,
		metadata:     tracing.Destination.Metadata,
	}

	applyTraffic(args, tracing.Traffic)

	return addXdsCluster(tCtx, args)
}

func buildTracingTags(customTags []ir.CustomTagMapEntry, tags []ir.MapEntry) ([]*tracingtype.CustomTag, error) {
	out := make(map[string]*tracingtype.CustomTag)

	for _, entry := range customTags {
		k, v := entry.Key, entry.Value
		switch v.Type {
		case egv1a1.CustomTagTypeLiteral:
			out[k] = &tracingtype.CustomTag{
				Tag: k,
				Type: &tracingtype.CustomTag_Literal_{
					Literal: &tracingtype.CustomTag_Literal{
						Value: v.Literal.Value,
					},
				},
			}
		case egv1a1.CustomTagTypeEnvironment:
			defaultVal := ""
			if v.Environment.DefaultValue != nil {
				defaultVal = *v.Environment.DefaultValue
			}

			out[k] = &tracingtype.CustomTag{
				Tag: k,
				Type: &tracingtype.CustomTag_Environment_{
					Environment: &tracingtype.CustomTag_Environment{
						Name:         v.Environment.Name,
						DefaultValue: defaultVal,
					},
				},
			}
		case egv1a1.CustomTagTypeRequestHeader:
			defaultVal := ""
			if v.RequestHeader.DefaultValue != nil {
				defaultVal = *v.RequestHeader.DefaultValue
			}

			out[k] = &tracingtype.CustomTag{
				Tag: k,
				Type: &tracingtype.CustomTag_RequestHeader{
					RequestHeader: &tracingtype.CustomTag_Header{
						Name:         v.RequestHeader.Name,
						DefaultValue: defaultVal,
					},
				},
			}
		default:
			return nil, fmt.Errorf("unknown custom tag type: %s", v.Type)
		}
	}

	// same key in tags will override tracingTags
	for _, entry := range tags {
		out[entry.Key] = &tracingtype.CustomTag{
			Tag: entry.Key,
			Type: &tracingtype.CustomTag_Value{
				Value: entry.Value,
			},
		}
	}

	result := make([]*tracingtype.CustomTag, 0, len(out))
	for _, v := range out {
		result = append(result, v)
	}

	// sort tags by tag name, make result consistent
	sort.Slice(result, func(i, j int) bool {
		return result[i].Tag < result[j].Tag
	})

	return result, nil
}

// buildResourceDetectors creates resource detectors for OpenTelemetry tracing
// using the StaticConfigResourceDetector extension with the given attributes.
func buildResourceDetectors(resources []ir.MapEntry) []*corev3.TypedExtensionConfig {
	if len(resources) == 0 {
		return nil
	}
	staticConfig := &resourcedetectorsv3.StaticConfigResourceDetectorConfig{
		Attributes: ir.SliceToMap(resources),
	}
	any, err := proto.ToAnyWithValidation(staticConfig)
	if err != nil {
		return nil
	}
	return []*corev3.TypedExtensionConfig{
		{
			Name:        "envoy.tracers.opentelemetry.resource_detectors.static_config",
			TypedConfig: any,
		},
	}
}

// buildSampler creates a sampler TypedExtensionConfig for the OpenTelemetry tracer.
func buildSampler(sampler *egv1a1.OTelSampler) (*corev3.TypedExtensionConfig, error) {
	if sampler == nil {
		return nil, nil
	}
	zero := &gwapiv1.Fraction{Numerator: 0}
	switch sampler.Type {
	case egv1a1.OTelSamplerTypeAlwaysOn:
		return buildAlwaysOnSampler()
	case egv1a1.OTelSamplerTypeAlwaysOff:
		return buildTraceIDRatioSampler(zero)
	case egv1a1.OTelSamplerTypeTraceIDRatio:
		return buildTraceIDRatioSampler(sampler.SamplingPercentage)
	case egv1a1.OTelSamplerTypeParentBasedAlwaysOn:
		wrapped, err := buildAlwaysOnSampler()
		if err != nil {
			return nil, err
		}
		return buildParentBasedSampler(wrapped)
	case egv1a1.OTelSamplerTypeParentBasedAlwaysOff:
		wrapped, err := buildTraceIDRatioSampler(zero)
		if err != nil {
			return nil, err
		}
		return buildParentBasedSampler(wrapped)
	case egv1a1.OTelSamplerTypeParentBasedTraceIDRatio:
		wrapped, err := buildTraceIDRatioSampler(sampler.SamplingPercentage)
		if err != nil {
			return nil, err
		}
		return buildParentBasedSampler(wrapped)
	default:
		return nil, fmt.Errorf("unknown sampler type: %s", sampler.Type)
	}
}

func buildAlwaysOnSampler() (*corev3.TypedExtensionConfig, error) {
	cfg, err := proto.ToAnyWithValidation(&samplersv3.AlwaysOnSamplerConfig{})
	if err != nil {
		return nil, err
	}
	return &corev3.TypedExtensionConfig{
		Name:        "envoy.tracers.opentelemetry.samplers.always_on",
		TypedConfig: cfg,
	}, nil
}

func buildTraceIDRatioSampler(fraction *gwapiv1.Fraction) (*corev3.TypedExtensionConfig, error) {
	var fp *xdstype.FractionalPercent
	if fraction != nil {
		fp = fractionalpercent.FromFraction(fraction)
	} else {
		// Default to 100% sampling.
		fp = &xdstype.FractionalPercent{
			Numerator:   100,
			Denominator: xdstype.FractionalPercent_HUNDRED,
		}
	}
	cfg, err := proto.ToAnyWithValidation(&samplersv3.TraceIdRatioBasedSamplerConfig{
		SamplingPercentage: fp,
	})
	if err != nil {
		return nil, err
	}
	return &corev3.TypedExtensionConfig{
		Name:        "envoy.tracers.opentelemetry.samplers.trace_id_ratio_based",
		TypedConfig: cfg,
	}, nil
}

func buildParentBasedSampler(wrapped *corev3.TypedExtensionConfig) (*corev3.TypedExtensionConfig, error) {
	cfg, err := proto.ToAnyWithValidation(&samplersv3.ParentBasedSamplerConfig{
		WrappedSampler: wrapped,
	})
	if err != nil {
		return nil, err
	}
	return &corev3.TypedExtensionConfig{
		Name:        "envoy.tracers.opentelemetry.samplers.parent_based",
		TypedConfig: cfg,
	}, nil
}
