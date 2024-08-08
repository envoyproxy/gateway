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
	cel "github.com/envoyproxy/go-control-plane/envoy/extensions/access_loggers/filters/cel/v3"
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

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
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

	tcpGRPCAccessLog = "envoy.access_loggers.tcp_grpc"
	celFilter        = "envoy.access_loggers.extension_filters.cel"
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
			Filter: buildAccessLogFilter(text.CELMatches, forListener),
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
			Filter: buildAccessLogFilter(json.CELMatches, forListener),
		})
	}
	// handle ALS access logs
	for _, als := range al.ALS {
		cc := &grpcaccesslog.CommonGrpcAccessLogConfig{
			LogName: als.LogName,
			GrpcService: &cfgcore.GrpcService{
				TargetSpecifier: &cfgcore.GrpcService_EnvoyGrpc_{
					EnvoyGrpc: &cfgcore.GrpcService_EnvoyGrpc{
						ClusterName: als.Destination.Name,
					},
				},
			},
			TransportApiVersion: cfgcore.ApiVersion_V3,
		}

		switch als.Type {
		case egv1a1.ALSEnvoyProxyAccessLogTypeHTTP:
			alCfg := &grpcaccesslog.HttpGrpcAccessLogConfig{
				CommonConfig: cc,
			}

			if als.HTTP != nil {
				alCfg.AdditionalRequestHeadersToLog = als.HTTP.RequestHeaders
				alCfg.AdditionalResponseHeadersToLog = als.HTTP.ResponseHeaders
				alCfg.AdditionalResponseTrailersToLog = als.HTTP.ResponseTrailers
			}

			accesslogAny, _ := anypb.New(alCfg)
			accessLogs = append(accessLogs, &accesslog.AccessLog{
				Name: wellknown.HTTPGRPCAccessLog,
				ConfigType: &accesslog.AccessLog_TypedConfig{
					TypedConfig: accesslogAny,
				},
				Filter: buildAccessLogFilter(als.CELMatches, forListener),
			})
		case egv1a1.ALSEnvoyProxyAccessLogTypeTCP:
			alCfg := &grpcaccesslog.TcpGrpcAccessLogConfig{
				CommonConfig: cc,
			}

			accesslogAny, _ := anypb.New(alCfg)
			accessLogs = append(accessLogs, &accesslog.AccessLog{
				Name: tcpGRPCAccessLog,
				ConfigType: &accesslog.AccessLog_TypedConfig{
					TypedConfig: accesslogAny,
				},
				Filter: buildAccessLogFilter(als.CELMatches, forListener),
			})
		}
	}
	// handle open telemetry access logs
	for _, otel := range al.OpenTelemetry {
		al := &otelaccesslog.OpenTelemetryAccessLogConfig{
			CommonConfig: &grpcaccesslog.CommonGrpcAccessLogConfig{
				LogName: otelLogName,
				GrpcService: &cfgcore.GrpcService{
					TargetSpecifier: &cfgcore.GrpcService_EnvoyGrpc_{
						EnvoyGrpc: &cfgcore.GrpcService_EnvoyGrpc{
							ClusterName: otel.Destination.Name,
							Authority:   otel.Authority,
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

		formatters := accessLogOpenTelemetryFormatters(format, otel.Attributes)
		if len(formatters) != 0 {
			al.Formatters = formatters
		}

		accesslogAny, _ := anypb.New(al)
		accessLogs = append(accessLogs, &accesslog.AccessLog{
			Name: otelAccessLog,
			ConfigType: &accesslog.AccessLog_TypedConfig{
				TypedConfig: accesslogAny,
			},
			Filter: buildAccessLogFilter(otel.CELMatches, forListener),
		})
	}

	return accessLogs
}

func celAccessLogFilter(expr string) *accesslog.AccessLogFilter {
	fl := &cel.ExpressionFilter{
		Expression: expr,
	}

	return &accesslog.AccessLogFilter{
		FilterSpecifier: &accesslog.AccessLogFilter_ExtensionFilter{
			ExtensionFilter: &accesslog.ExtensionFilter{
				Name:       celFilter,
				ConfigType: &accesslog.ExtensionFilter_TypedConfig{TypedConfig: protocov.ToAny(fl)},
			},
		},
	}
}

func buildAccessLogFilter(exprs []string, forListener bool) *accesslog.AccessLogFilter {
	// add filter for access logs
	var filters []*accesslog.AccessLogFilter
	for _, expr := range exprs {
		filters = append(filters, celAccessLogFilter(expr))
	}
	if forListener {
		filters = append(filters, listenerAccessLogFilter)
	}

	if len(filters) == 0 {
		return nil
	}

	if len(filters) == 1 {
		return filters[0]
	}

	return &accesslog.AccessLogFilter{
		FilterSpecifier: &accesslog.AccessLogFilter_AndFilter{
			AndFilter: &accesslog.AndFilter{
				Filters: filters,
			},
		},
	}
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

func accessLogOpenTelemetryFormatters(body string, attributes map[string]string) []*cfgcore.TypedExtensionConfig {
	reqWithoutQuery, metadata, cel := false, false, false

	if strings.Contains(body, reqWithoutQueryCommandOperator) {
		reqWithoutQuery = true
	}

	if strings.Contains(body, metadataCommandOperator) {
		metadata = true
	}

	if strings.Contains(body, celCommandOperator) {
		cel = true
	}

	for _, value := range attributes {
		if reqWithoutQuery && metadata && cel {
			break
		}

		if !reqWithoutQuery && strings.Contains(value, reqWithoutQueryCommandOperator) {
			reqWithoutQuery = true
		}

		if !metadata && strings.Contains(value, metadataCommandOperator) {
			metadata = true
		}

		if !cel && strings.Contains(value, celCommandOperator) {
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
	// add clusters for ALS access logs
	for _, als := range al.ALS {
		traffic := als.Traffic
		// Make sure that there are safe defaults for the traffic
		if traffic == nil {
			traffic = &ir.TrafficFeatures{}
		}
		if err := addXdsCluster(tCtx, &xdsClusterArgs{
			name:              als.Destination.Name,
			settings:          als.Destination.Settings,
			tSocket:           nil,
			endpointType:      EndpointTypeStatic,
			loadBalancer:      traffic.LoadBalancer,
			proxyProtocol:     traffic.ProxyProtocol,
			circuitBreaker:    traffic.CircuitBreaker,
			healthCheck:       traffic.HealthCheck,
			timeout:           traffic.Timeout,
			tcpkeepalive:      traffic.TCPKeepalive,
			backendConnection: traffic.BackendConnection,
			dns:               traffic.DNS,
			http2Settings:     traffic.HTTP2,
		}); err != nil && !errors.Is(err, ErrXdsClusterExists) {
			return err
		}
	}

	// add clusters for Open Telemetry access logs
	for _, otel := range al.OpenTelemetry {
		traffic := otel.Traffic
		// Make sure that there are safe defaults for the traffic
		if traffic == nil {
			traffic = &ir.TrafficFeatures{}
		}

		if err := addXdsCluster(tCtx, &xdsClusterArgs{
			name:              otel.Destination.Name,
			settings:          otel.Destination.Settings,
			tSocket:           nil,
			endpointType:      EndpointTypeDNS,
			metrics:           metrics,
			loadBalancer:      traffic.LoadBalancer,
			proxyProtocol:     traffic.ProxyProtocol,
			circuitBreaker:    traffic.CircuitBreaker,
			healthCheck:       traffic.HealthCheck,
			timeout:           traffic.Timeout,
			tcpkeepalive:      traffic.TCPKeepalive,
			backendConnection: traffic.BackendConnection,
			dns:               traffic.DNS,
			http2Settings:     traffic.HTTP2,
		}); err != nil && !errors.Is(err, ErrXdsClusterExists) {
			return err
		}
	}

	return nil
}
