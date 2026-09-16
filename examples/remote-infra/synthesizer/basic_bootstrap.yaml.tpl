cluster_manager:
  local_cluster_name: {{ .ServiceClusterName }}
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
    initial_fetch_timeout: 0s
    resource_api_version: V3
  cds_config:
    ads: {}
    initial_fetch_timeout: 0s
    resource_api_version: V3
static_resources:
  clusters:
    - connect_timeout: 10s
      eds_cluster_config:
        eds_config:
          ads: {}
          resource_api_version: 'V3'
        service_name: {{ .ServiceClusterName }}
      load_balancing_policy:
        policies:
          - typed_extension_config:
              name: 'envoy.load_balancing_policies.least_request'
              typed_config:
                '@type': 'type.googleapis.com/envoy.extensions.load_balancing_policies.least_request.v3.LeastRequest'
                locality_lb_config:
                  zone_aware_lb_config:
                    min_cluster_size: '1'
      name: {{ .ServiceClusterName }}
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
                      address: envoy-gateway.envoy-gateway-system.svc.cluster.local.
                      port_value: 18000
      typed_extension_protocol_options:
        envoy.extensions.upstreams.http.v3.HttpProtocolOptions:
          "@type": "type.googleapis.com/envoy.extensions.upstreams.http.v3.HttpProtocolOptions"
          explicit_http_config:
            http2_protocol_options:
              connection_keepalive:
                interval: 30s
                timeout: 5s
      name: xds_cluster
      type: STRICT_DNS
      transport_socket:
        name: envoy.transport_sockets.tls
        typed_config:
          "@type": type.googleapis.com/envoy.extensions.transport_sockets.tls.v3.UpstreamTlsContext
          common_tls_context:
            tls_params:
              tls_maximum_protocol_version: TLSv1_3
            tls_certificates:
              - certificate_chain:
                  filename: /certs/tls.crt
                private_key:
                  filename: /certs/tls.key
            validation_context:
              trusted_ca:
                filename: /certs/ca.crt
              match_typed_subject_alt_names:
                - san_type: DNS
                  matcher:
                    exact: envoy-gateway