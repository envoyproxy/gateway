admin:
  access_log:
  - name: envoy.access_loggers.file
    typed_config:
      "@type": type.googleapis.com/envoy.extensions.access_loggers.file.v3.FileAccessLog
      path: {{ .AdminServer.AccessLogPath }}
  address:
    socket_address:
      address: {{ .AdminServer.Address }}
      port_value: {{ .AdminServer.Port }}
cluster_manager:
  local_cluster_name: $(ENVOY_PROXY_INFRA_NAME)
node:
  locality:
    zone: $(ENVOY_SERVICE_ZONE)
{{- if .StatsMatcher  }}
stats_config:
  stats_matcher:
    inclusion_list:
      patterns:
      {{- range $_, $item := .StatsMatcher.Exacts }}
      - exact: {{$item}}
      {{- end}}
      {{- range $_, $item := .StatsMatcher.Prefixes }}
      - prefix: {{$item}}
      {{- end}}
      {{- range $_, $item := .StatsMatcher.Suffixes }}
      - suffix: {{$item}}
      {{- end}}
      {{- range $_, $item := .StatsMatcher.RegularExpressions }}
      - safe_regex:
          google_re2: {}
          regex: {{js $item}}
      {{- end}}
{{- end }}
layered_runtime:
  layers:
  - name: global_config
    static_layer:
      envoy.restart_features.use_eds_cache_for_ads: true
      re2.max_program_size.error_level: 4294967295
      re2.max_program_size.warn_level: 1000
dynamic_resources:
  ads_config:
    api_type: DELTA_GRPC
    transport_api_version: V3
    grpc_services:
    - envoy_grpc:
        cluster_name: xds_cluster
    set_node_on_first_message_only: true
  lds_config:
    ads: {}
    resource_api_version: V3
  cds_config:
    ads: {}
    resource_api_version: V3
{{- if .OtelMetricSinks }}
stats_sinks:
{{- range $idx, $sink := .OtelMetricSinks }}
- name: "envoy.stat_sinks.open_telemetry"
  typed_config:
    "@type": type.googleapis.com/envoy.extensions.stat_sinks.open_telemetry.v3.SinkConfig
    grpc_service:
      envoy_grpc:
        cluster_name: otel_metric_sink_{{ $idx }}
{{- end }}
{{- end }}
static_resources:
  {{- if .EnablePrometheus }}
  listeners:
  - name: envoy-gateway-proxy-stats-{{ .StatsServer.Address }}-{{ .StatsServer.Port }}
    address:
      socket_address:
        address: '{{ .StatsServer.Address }}'
        port_value: {{ .StatsServer.Port }}
        protocol: TCP
        {{- if eq .IPFamily "DualStack" "IPv6" }}
        ipv4_compat: true
        {{- end }}
    bypass_overload_manager: true
    filter_chains:
    - filters:
      - name: envoy.filters.network.http_connection_manager
        typed_config:
          "@type": type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
          stat_prefix: eg-stats-http
          normalize_path: true
          route_config:
            name: local_route
            virtual_hosts:
            - name: prometheus_stats
              domains:
              - "*"
              {{- if .EnablePrometheusCompression }}
              typed_per_filter_config:
                envoy.filters.http.compression:
                  "@type": type.googleapis.com/envoy.extensions.filters.http.compressor.v3.CompressorPerRoute
                  disabled: true
              {{- end }}
              routes:
              - match:
                  path: /stats/prometheus
                  headers:
                  - name: ":method"
                    string_match:
                      exact: GET
                route:
                  cluster: prometheus_stats
                {{- if .EnablePrometheusCompression }}
                typed_per_filter_config:
                  envoy.filters.http.compression:
                    "@type": type.googleapis.com/envoy.extensions.filters.http.compressor.v3.CompressorPerRoute
                    overrides:
                      response_direction_config:
                {{- end }}
          http_filters:
          {{- if .EnablePrometheusCompression }}
          - name: envoy.filters.http.compressor
            typed_config:
              "@type": type.googleapis.com/envoy.extensions.filters.http.compressor.v3.Compressor
              {{- if eq .PrometheusCompressionLibrary "Gzip"}}
              compressor_library:
                name: text_optimized
                typed_config:
                  "@type": type.googleapis.com/envoy.extensions.compression.gzip.compressor.v3.Gzip
              {{- end }}
              {{- if eq .PrometheusCompressionLibrary "Brotli"}}
              compressor_library:
                name: text_optimized
                typed_config:
                  "@type": type.googleapis.com/envoy.extensions.compression.brotli.compressor.v3.Brotli
              {{- end }}
          {{- end }}
          - name: envoy.filters.http.router
            typed_config:
              "@type": type.googleapis.com/envoy.extensions.filters.http.router.v3.Router
  {{- end }}
  clusters:
  {{- if .EnablePrometheus }}
  - name: prometheus_stats
    connect_timeout: 0.250s
    type: STATIC
    lb_policy: ROUND_ROBIN
    load_assignment:
      cluster_name: prometheus_stats
      endpoints:
      - lb_endpoints:
        - endpoint:
            address:
              socket_address:
                address: {{ .AdminServer.Address }}
                port_value: {{ .AdminServer.Port }}
  {{- end }}
  {{- range $idx, $sink := .OtelMetricSinks }}
  - name: otel_metric_sink_{{ $idx }}
    connect_timeout: 0.250s
    type: STRICT_DNS
    typed_extension_protocol_options:
      envoy.extensions.upstreams.http.v3.HttpProtocolOptions:
        "@type": "type.googleapis.com/envoy.extensions.upstreams.http.v3.HttpProtocolOptions"
        explicit_http_config:
          http2_protocol_options: {}
    lb_policy: ROUND_ROBIN
    load_assignment:
      cluster_name: otel_metric_sink_{{ $idx }}
      endpoints:
      - lb_endpoints:
        - endpoint:
            address:
              socket_address:
                address: {{ $sink.Address }}
                port_value: {{ $sink.Port }}
  {{- end }}
  - connect_timeout: 10s
    eds_cluster_config:
      eds_config:
        ads: {}
        resource_api_version: 'V3'
      service_name: $(ENVOY_PROXY_INFRA_NAME)
    load_balancing_policy:
      policies:
        - typed_extension_config:
            name: 'envoy.load_balancing_policies.least_request'
            typed_config:
              '@type': 'type.googleapis.com/envoy.extensions.load_balancing_policies.least_request.v3.LeastRequest'
              locality_lb_config:
                locality_weighted_lb_config: {}
    name: $(ENVOY_PROXY_INFRA_NAME)
    type: EDS
  - connect_timeout: 10s
    load_assignment:
      cluster_name: xds_cluster
      endpoints:
      - load_balancing_weight: 1
        lb_endpoints:
        - load_balancing_weight: 1
          endpoint:
            address:
              socket_address:
                address: {{ .XdsServer.Address }}
                port_value: {{ .XdsServer.Port }}
    typed_extension_protocol_options:
      envoy.extensions.upstreams.http.v3.HttpProtocolOptions:
        "@type": "type.googleapis.com/envoy.extensions.upstreams.http.v3.HttpProtocolOptions"
        explicit_http_config:
          http2_protocol_options:
            connection_keepalive:
              interval: 30s
              timeout: 5s
  {{- if .GatewayNamespaceMode }}
        http_filters:
        - name: envoy.filters.http.credential_injector
          typed_config:
            "@type": type.googleapis.com/envoy.extensions.filters.http.credential_injector.v3.CredentialInjector
            credential:
              name: envoy.http.injected_credentials.generic
              typed_config:
                "@type": type.googleapis.com/envoy.extensions.http.injected_credentials.generic.v3.Generic
                credential:
                  name: jwt-sa-bearer
            overwrite: true
        - name: envoy.extensions.filters.http.upstream_codec.v3.UpstreamCodec
          typed_config:
            "@type": type.googleapis.com/envoy.extensions.filters.http.upstream_codec.v3.UpstreamCodec
  {{- end }}
    name: xds_cluster
    type: STRICT_DNS
    transport_socket:
      name: envoy.transport_sockets.tls
      typed_config:
        "@type": type.googleapis.com/envoy.extensions.transport_sockets.tls.v3.UpstreamTlsContext
        common_tls_context:
          tls_params:
            tls_maximum_protocol_version: TLSv1_3
{{- if not .GatewayNamespaceMode }}
          tls_certificate_sds_secret_configs:
          - name: xds_certificate
            sds_config:
              path_config_source:
                path: {{ .SdsCertificatePath }}
              resource_api_version: V3
{{- end }}
          validation_context_sds_secret_config:
            name: xds_trusted_ca
            sds_config:
              path_config_source:
                path: {{ .SdsTrustedCAPath }}
              resource_api_version: V3
  {{- if .GatewayNamespaceMode }}
  secrets:
  - name: jwt-sa-bearer
    generic_secret:
      secret:
        filename: "/var/run/secrets/token/sa-token"
  {{- end }}
overload_manager:
  refresh_interval: 0.25s
  resource_monitors:
  - name: "envoy.resource_monitors.global_downstream_max_connections"
    typed_config:
      "@type": type.googleapis.com/envoy.extensions.resource_monitors.downstream_connections.v3.DownstreamConnectionsConfig
      max_active_downstream_connections: 50000
  {{- with .OverloadManager.MaxHeapSizeBytes }}
  - name: "envoy.resource_monitors.fixed_heap"
    typed_config:
      "@type": type.googleapis.com/envoy.extensions.resource_monitors.fixed_heap.v3.FixedHeapConfig
      max_heap_size_bytes: {{ . }}
  actions:
  - name: "envoy.overload_actions.shrink_heap"
    triggers:
    - name: "envoy.resource_monitors.fixed_heap"
      threshold:
        value: 0.95
  - name: "envoy.overload_actions.stop_accepting_requests"
    triggers:
    - name: "envoy.resource_monitors.fixed_heap"
      threshold:
        value: 0.98
  {{- end }}
