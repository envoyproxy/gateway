// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"errors"
	"sort"
	"strings"

	accesslog "github.com/envoyproxy/go-control-plane/envoy/config/accesslog/v3"
	cfgcore "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	fileaccesslog "github.com/envoyproxy/go-control-plane/envoy/extensions/access_loggers/file/v3"
	grpcaccesslog "github.com/envoyproxy/go-control-plane/envoy/extensions/access_loggers/grpc/v3"
	otelaccesslog "github.com/envoyproxy/go-control-plane/envoy/extensions/access_loggers/open_telemetry/v3"
	celformatter "github.com/envoyproxy/go-control-plane/envoy/extensions/formatter/cel/v3"
	metadataformatter "github.com/envoyproxy/go-control-plane/envoy/extensions/formatter/metadata/v3"
	reqwithoutqueryformatter "github.com/envoyproxy/go-control-plane/envoy/extensions/formatter/req_without_query/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	otlpcommonv1 "go.opentelemetry.io/proto/otlp/common/v1"
	"golang.org/x/exp/maps"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/structpb"
	"k8s.io/utils/ptr"

	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/utils/protocov"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

const (
	// EnvoyTextLogFormat is the default log format for Envoy.
	// See https://www.envoyproxy.io/docs/envoy/latest/configuration/observability/access_log/usage#default-format-string
	EnvoyTextLogFormat = "{\"start_time\":\"%START_TIME%\",\"method\":\"%REQ(:METHOD)%\"," +
		"\"x-envoy-origin-path\":\"%REQ(X-ENVOY-ORIGINAL-PATH?:PATH)%\",\"protocol\":\"%PROTOCOL%\"," +
		"\"response_code\":\"%RESPONSE_CODE%\",\"response_flags\":\"%RESPONSE_FLAGS%\"," +
		"\"response_code_details\":\"%RESPONSE_CODE_DETAILS%\"," +
		"\"connection_termination_details\":\"%CONNECTION_TERMINATION_DETAILS%\"," +
		"\"upstream_transport_failure_reason\":\"%UPSTREAM_TRANSPORT_FAILURE_REASON%\"," +
		"\"bytes_received\":\"%BYTES_RECEIVED%\",\"bytes_sent\":\"%BYTES_SENT%\"," +
		"\"duration\":\"%DURATION%\",\"x-envoy-upstream-service-time\":\"%RESP(X-ENVOY-UPSTREAM-SERVICE-TIME)%\"," +
		"\"x-forwarded-for\":\"%REQ(X-FORWARDED-FOR)%\",\"user-agent\":\"%REQ(USER-AGENT)%\"," +
		"\"x-request-id\":\"%REQ(X-REQUEST-ID)%\",\":authority\":\"%REQ(:AUTHORITY)%\"," +
		"\"upstream_host\":\"%UPSTREAM_HOST%\",\"upstream_cluster\":\"%UPSTREAM_CLUSTER%\"," +
		"\"upstream_local_address\":\"%UPSTREAM_LOCAL_ADDRESS%\"," +
		"\"downstream_local_address\":\"%DOWNSTREAM_LOCAL_ADDRESS%\"," +
		"\"downstream_remote_address\":\"%DOWNSTREAM_REMOTE_ADDRESS%\"," +
		"\"requested_server_name\":\"%REQUESTED_SERVER_NAME%\",\"route_name\":\"%ROUTE_NAME%\"}\n"

	otelLogName   = "otel_envoy_accesslog"
	otelAccessLog = "envoy.access_loggers.open_telemetry"

	reqWithoutQueryCommandOperator = "%REQ_WITHOUT_QUERY"
	metadataCommandOperator        = "%METADATA"
	celCommandOperator             = "%CEL"
)

// for the case when a route does not exist to upstream, hcm logs will not be present
var listenerAccessLogFilter = &accesslog.AccessLogFilter{
	FilterSpecifier: &accesslog.AccessLogFilter_ResponseFlagFilter{
		ResponseFlagFilter: &accesslog.ResponseFlagFilter{Flags: []string{"NR"}},
	},
}

var (
	// reqWithoutQueryFormatter configures additional formatters needed for some of the format strings like "REQ_WITHOUT_QUERY"
	reqWithoutQueryFormatter = &cfgcore.TypedExtensionConfig{
		Name:        "envoy.formatter.req_without_query",
		TypedConfig: protocov.ToAny(&reqwithoutqueryformatter.ReqWithoutQuery{}),
	}

	// metadataFormatter configures additional formatters needed for some of the format strings like "METADATA"
	// for more information, see https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/formatter/metadata/v3/metadata.proto
	metadataFormatter = &cfgcore.TypedExtensionConfig{
		Name:        "envoy.formatter.metadata",
		TypedConfig: protocov.ToAny(&metadataformatter.Metadata{}),
	}

	// celFormatter configures additional formatters needed for some of the format strings like "CEL"
	// for more information, see https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/formatter/cel/v3/cel.proto
	celFormatter = &cfgcore.TypedExtensionConfig{
		Name:        "envoy.formatter.cel",
		TypedConfig: protocov.ToAny(&celformatter.Cel{}),
	}
)

func buildXdsAccessLog(al *ir.AccessLog, forListener bool) []*accesslog.AccessLog {
	if al == nil {
		return nil
	}

	totalLen := len(al.Text) + len(al.JSON) + len(al.OpenTelemetry)
	accessLogs := make([]*accesslog.AccessLog, 0, totalLen)
	// handle text file access logs
	for _, text := range al.Text {
		filelog := &fileaccesslog.FileAccessLog{
			Path: text.Path,
		}
		format := EnvoyTextLogFormat
		if text.Format != nil {
			format = *text.Format
		}

		filelog.AccessLogFormat = &fileaccesslog.FileAccessLog_LogFormat{
			LogFormat: &cfgcore.SubstitutionFormatString{
				Format: &cfgcore.SubstitutionFormatString_TextFormatSource{
					TextFormatSource: &cfgcore.DataSource{
						Specifier: &cfgcore.DataSource_InlineString{
							InlineString: format,
						},
					},
				},
			},
		}

		formatters := accessLogTextFormatters(format)
		if len(formatters) != 0 {
			filelog.GetLogFormat().Formatters = formatters
		}

		// TODO: find a better way to handle this
		accesslogAny, _ := anypb.New(filelog)
		accessLogs = append(accessLogs, &accesslog.AccessLog{
			Name: wellknown.FileAccessLog,
			ConfigType: &accesslog.AccessLog_TypedConfig{
				TypedConfig: accesslogAny,
			},
		})
	}
	// handle json file access logs
	for _, json := range al.JSON {
		jsonFormat := &structpb.Struct{
			Fields: make(map[string]*structpb.Value, len(json.JSON)),
		}

		// sort keys to ensure consistent ordering
		keys := maps.Keys(json.JSON)
		sort.Strings(keys)

		for _, key := range keys {
			jsonFormat.Fields[key] = &structpb.Value{
				Kind: &structpb.Value_StringValue{
					StringValue: json.JSON[key],
				},
			}
		}

		filelog := &fileaccesslog.FileAccessLog{
			Path: json.Path,
			AccessLogFormat: &fileaccesslog.FileAccessLog_LogFormat{
				LogFormat: &cfgcore.SubstitutionFormatString{
					Format: &cfgcore.SubstitutionFormatString_JsonFormat{
						JsonFormat: jsonFormat,
					},
				},
			},
		}

		formatters := accessLogJSONFormatters(json.JSON)
		if len(formatters) != 0 {
			filelog.GetLogFormat().Formatters = formatters
		}

		accesslogAny, _ := anypb.New(filelog)
		accessLogs = append(accessLogs, &accesslog.AccessLog{
			Name: wellknown.FileAccessLog,
			ConfigType: &accesslog.AccessLog_TypedConfig{
				TypedConfig: accesslogAny,
			},
		})
	}
	// handle open telemetry access logs
	for _, otel := range al.OpenTelemetry {
		al := &otelaccesslog.OpenTelemetryAccessLogConfig{
			CommonConfig: &grpcaccesslog.CommonGrpcAccessLogConfig{
				LogName: otelLogName,
				GrpcService: &cfgcore.GrpcService{
					TargetSpecifier: &cfgcore.GrpcService_EnvoyGrpc_{
						EnvoyGrpc: &cfgcore.GrpcService_EnvoyGrpc{
							ClusterName: buildClusterName("accesslog", otel.Host, otel.Port),
							Authority:   otel.Host,
						},
					},
				},
				TransportApiVersion: cfgcore.ApiVersion_V3,
			},
			ResourceAttributes: convertToKeyValueList(otel.Resources, false),
		}

		format := EnvoyTextLogFormat
		if otel.Text != nil {
			format = *otel.Text
		}

		if format != "" {
			al.Body = &otlpcommonv1.AnyValue{
				Value: &otlpcommonv1.AnyValue_StringValue{
					StringValue: format,
				},
			}
		}

		al.Attributes = convertToKeyValueList(otel.Attributes, true)

		accesslogAny, _ := anypb.New(al)
		accessLogs = append(accessLogs, &accesslog.AccessLog{
			Name: otelAccessLog,
			ConfigType: &accesslog.AccessLog_TypedConfig{
				TypedConfig: accesslogAny,
			},
		})
	}

	// add filter for listener access logs
	if forListener {
		for _, al := range accessLogs {
			al.Filter = listenerAccessLogFilter
		}
	}

	return accessLogs
}

func accessLogTextFormatters(text string) []*cfgcore.TypedExtensionConfig {
	formatters := make([]*cfgcore.TypedExtensionConfig, 0, 3)

	if strings.Contains(text, reqWithoutQueryCommandOperator) {
		formatters = append(formatters, reqWithoutQueryFormatter)
	}

	if strings.Contains(text, metadataCommandOperator) {
		formatters = append(formatters, metadataFormatter)
	}

	if strings.Contains(text, celCommandOperator) {
		formatters = append(formatters, celFormatter)
	}

	return formatters
}

func accessLogJSONFormatters(json map[string]string) []*cfgcore.TypedExtensionConfig {
	reqWithoutQuery, metadata, cel := false, false, false

	for _, value := range json {
		if reqWithoutQuery && metadata && cel {
			break
		}

		if strings.Contains(value, reqWithoutQueryCommandOperator) {
			reqWithoutQuery = true
		}

		if strings.Contains(value, metadataCommandOperator) {
			metadata = true
		}

		if strings.Contains(value, celCommandOperator) {
			cel = true
		}
	}

	formatters := make([]*cfgcore.TypedExtensionConfig, 0, 3)

	if reqWithoutQuery {
		formatters = append(formatters, reqWithoutQueryFormatter)
	}

	if metadata {
		formatters = append(formatters, metadataFormatter)
	}

	if cel {
		formatters = append(formatters, celFormatter)
	}

	return formatters
}

// read more here: https://opentelemetry.io/docs/specs/otel/resource/semantic_conventions/k8s/
const (
	k8sNamespaceNameKey = "k8s.namespace.name"
	k8sPodNameKey       = "k8s.pod.name"
)

func convertToKeyValueList(attributes map[string]string, additionalLabels bool) *otlpcommonv1.KeyValueList {
	maxLen := len(attributes)
	if additionalLabels {
		maxLen += 2
	}
	keyValueList := &otlpcommonv1.KeyValueList{
		Values: make([]*otlpcommonv1.KeyValue, 0, maxLen),
	}

	// always set the k8s namespace and pod name for better UX
	// EG cannot know the client namespace and pod name,
	// so we set these on attributes that read from the environment.
	if additionalLabels {
		// TODO: check the provider type and set the appropriate attributes
		keyValueList.Values = append(keyValueList.Values, &otlpcommonv1.KeyValue{
			Key:   k8sNamespaceNameKey,
			Value: &otlpcommonv1.AnyValue{Value: &otlpcommonv1.AnyValue_StringValue{StringValue: "%ENVIRONMENT(ENVOY_GATEWAY_NAMESPACE)%"}},
		})

		keyValueList.Values = append(keyValueList.Values, &otlpcommonv1.KeyValue{
			Key:   k8sPodNameKey,
			Value: &otlpcommonv1.AnyValue{Value: &otlpcommonv1.AnyValue_StringValue{StringValue: "%ENVIRONMENT(ENVOY_POD_NAME)%"}},
		})
	}

	if len(attributes) == 0 {
		return keyValueList
	}

	// sort keys to ensure consistent ordering
	keys := maps.Keys(attributes)
	sort.Strings(keys)

	for _, key := range keys {
		keyValueList.Values = append(keyValueList.Values, &otlpcommonv1.KeyValue{
			Key:   key,
			Value: &otlpcommonv1.AnyValue{Value: &otlpcommonv1.AnyValue_StringValue{StringValue: attributes[key]}},
		})
	}

	return keyValueList
}

func processClusterForAccessLog(tCtx *types.ResourceVersionTable, al *ir.AccessLog, metrics *ir.Metrics) error {
	if al == nil {
		return nil
	}

	for _, otel := range al.OpenTelemetry {
		clusterName := buildClusterName("accesslog", otel.Host, otel.Port)

		ds := &ir.DestinationSetting{
			Weight:    ptr.To[uint32](1),
			Protocol:  ir.GRPC,
			Endpoints: []*ir.DestinationEndpoint{ir.NewDestEndpoint(otel.Host, otel.Port)},
		}
		if err := addXdsCluster(tCtx, &xdsClusterArgs{
			name:         clusterName,
			settings:     []*ir.DestinationSetting{ds},
			tSocket:      nil,
			endpointType: EndpointTypeDNS,
			metrics:      metrics,
		}); err != nil && !errors.Is(err, ErrXdsClusterExists) {
			return err
		}

	}

	return nil
}
