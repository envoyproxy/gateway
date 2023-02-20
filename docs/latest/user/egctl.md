# egctl

`egctl` is a command line tool to provide additional functionality for Envoy Gateway users.

## egctl experimental translate

This subcommand allows users to translate from an input configuration type to an output configuration type.

In the below example, we will translate the Kubernetes resources (including the Gateway API resources) into xDS
resources.

```
cat <<EOF >> gateway-api-config.yaml
apiVersion: gateway.networking.k8s.io/v1beta1
kind: GatewayClass
metadata:
  name: eg
spec:
  controllerName: gateway.envoyproxy.io/gatewayclass-controller
---
apiVersion: gateway.networking.k8s.io/v1beta1
kind: Gateway
metadata:
  name: eg
  namespace: default
spec:
  gatewayClassName: eg
  listeners:
    - name: http
      protocol: HTTP
      port: 80
---
apiVersion: v1
kind: Namespace
metadata:
  name: default 
---
apiVersion: v1
kind: Service
metadata:
  name: backend
  namespace: default
  labels:
    app: backend
    service: backend
spec:
  clusterIP: "1.1.1.1"
  type: ClusterIP
  ports:
    - name: http
      port: 3000
      targetPort: 3000
      protocol: TCP
  selector:
    app: backend
---
apiVersion: gateway.networking.k8s.io/v1beta1
kind: HTTPRoute
metadata:
  name: backend
  namespace: default
spec:
  parentRefs:
    - name: eg
  hostnames:
    - "www.example.com"
  rules:
    - backendRefs:
        - group: ""
          kind: Service
          name: backend
          port: 3000
          weight: 1
      matches:
        - path:
            type: PathPrefix
            value: /
EOF
```

```
egctl x translate --from gateway-api --to xds -f gateway-api-config.yaml

xDS

Key: default-eg

Bootstrap:
admin:
  access_log:
  - name: envoy.access_loggers.file
    typed_config:
      "@type": type.googleapis.com/envoy.extensions.access_loggers.file.v3.FileAccessLog
      path: /dev/null
  address:
    socket_address:
      address: 127.0.0.1
      port_value: 19000
dynamic_resources:
  cds_config:
    resource_api_version: V3
    api_config_source:
      api_type: DELTA_GRPC
      transport_api_version: V3
      grpc_services:
      - envoy_grpc:
          cluster_name: xds_cluster
      set_node_on_first_message_only: true
  lds_config:
    resource_api_version: V3
    api_config_source:
      api_type: DELTA_GRPC
      transport_api_version: V3
      grpc_services:
      - envoy_grpc:
          cluster_name: xds_cluster
      set_node_on_first_message_only: true
static_resources:
  clusters:
  - connect_timeout: 1s
    load_assignment:
      cluster_name: xds_cluster
      endpoints:
      - lb_endpoints:
        - endpoint:
            address:
              socket_address:
                address: envoy-gateway
                port_value: 18000
    typed_extension_protocol_options:
      "envoy.extensions.upstreams.http.v3.HttpProtocolOptions":
         "@type": "type.googleapis.com/envoy.extensions.upstreams.http.v3.HttpProtocolOptions"
         "explicit_http_config":
           "http2_protocol_options": {}
    name: xds_cluster
    type: STRICT_DNS
    transport_socket:
      name: envoy.transport_sockets.tls
      typed_config:
        "@type": type.googleapis.com/envoy.extensions.transport_sockets.tls.v3.UpstreamTlsContext
        common_tls_context:
          tls_params:
            tls_maximum_protocol_version: TLSv1_3
          tls_certificate_sds_secret_configs:
          - name: xds_certificate
            sds_config:
              path_config_source:
                path: "/sds/xds-certificate.json"
              resource_api_version: V3
          validation_context_sds_secret_config:
            name: xds_trusted_ca
            sds_config:
              path_config_source:
                path: "/sds/xds-trusted-ca.json"
              resource_api_version: V3
layered_runtime:
  layers:
    - name: runtime-0
      rtds_layer:
        rtds_config:
          resource_api_version: V3
          api_config_source:
            transport_api_version: V3
            api_type: DELTA_GRPC
            grpc_services:
              envoy_grpc:
                cluster_name: xds_cluster
        name: runtime-0

Listeners:
- accessLog:
  - filter:
      responseFlagFilter:
        flags:
        - NR
    name: envoy.access_loggers.file
    typedConfig:
      '@type': type.googleapis.com/envoy.extensions.access_loggers.file.v3.FileAccessLog
      path: /dev/stdout
  address:
    socketAddress:
      address: 0.0.0.0
      portValue: 10080
  defaultFilterChain:
    filters:
    - name: envoy.filters.network.http_connection_manager
      typedConfig:
        '@type': type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
        accessLog:
        - name: envoy.access_loggers.file
          typedConfig:
            '@type': type.googleapis.com/envoy.extensions.access_loggers.file.v3.FileAccessLog
            path: /dev/stdout
        httpFilters:
        - name: envoy.filters.http.router
          typedConfig:
            '@type': type.googleapis.com/envoy.extensions.filters.http.router.v3.Router
        rds:
          configSource:
            apiConfigSource:
              apiType: DELTA_GRPC
              grpcServices:
              - envoyGrpc:
                  clusterName: xds_cluster
              setNodeOnFirstMessageOnly: true
              transportApiVersion: V3
            resourceApiVersion: V3
          routeConfigName: default-eg-http
        statPrefix: http
        upgradeConfigs:
        - upgradeType: websocket
        useRemoteAddress: true
        useRemoteAddress: true
  name: default-eg-http

Routes:
- name: default-eg-http
  virtualHosts:
  - domains:
    - '*'
    name: default-eg-http
    routes:
    - match:
        headers:
        - name: :authority
          stringMatch:
            exact: www.example.com
        prefix: /
      route:
        cluster: default-backend-rule-0-match-0-www.example.com

Clusters:
- commonLbConfig:
    localityWeightedLbConfig: {}
  connectTimeout: 5s
  dnsLookupFamily: V4_ONLY
  loadAssignment:
    clusterName: default-backend-rule-0-match-0-www.example.com
    endpoints:
    - lbEndpoints:
      - endpoint:
          address:
            socketAddress:
              address: 1.1.1.1
              portValue: 3000
        loadBalancingWeight: 1
      loadBalancingWeight: 1
      locality: {}
  name: default-backend-rule-0-match-0-www.example.com
  outlierDetection: {}
  type: STATIC
```  
