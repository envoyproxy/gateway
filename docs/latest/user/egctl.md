# egctl

`egctl` is a command line tool to provide additional functionality for Envoy Gateway users.

## egctl experimental translate

This subcommand allows users to translate from an input configuration type to an output configuration type.

In the below example, we will translate the Kubernetes resources (including the Gateway API resources) into xDS
resources.

```shell
cat <<EOF | egctl x translate --from gateway-api --to xds -f -
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

```yaml
configKey: default-eg
configs:
- '@type': type.googleapis.com/envoy.admin.v3.BootstrapConfigDump
  bootstrap:
    admin:
      accessLog:
      - name: envoy.access_loggers.file
        typedConfig:
          '@type': type.googleapis.com/envoy.extensions.access_loggers.file.v3.FileAccessLog
          path: /dev/null
      address:
        socketAddress:
          address: 127.0.0.1
          portValue: 19000
    dynamicResources:
      cdsConfig:
        apiConfigSource:
          apiType: DELTA_GRPC
          grpcServices:
          - envoyGrpc:
              clusterName: xds_cluster
          setNodeOnFirstMessageOnly: true
          transportApiVersion: V3
        resourceApiVersion: V3
      ldsConfig:
        apiConfigSource:
          apiType: DELTA_GRPC
          grpcServices:
          - envoyGrpc:
              clusterName: xds_cluster
          setNodeOnFirstMessageOnly: true
          transportApiVersion: V3
        resourceApiVersion: V3
    layeredRuntime:
      layers:
      - name: runtime-0
        rtdsLayer:
          name: runtime-0
          rtdsConfig:
            apiConfigSource:
              apiType: DELTA_GRPC
              grpcServices:
              - envoyGrpc:
                  clusterName: xds_cluster
              transportApiVersion: V3
            resourceApiVersion: V3
    staticResources:
      clusters:
      - connectTimeout: 1s
        loadAssignment:
          clusterName: xds_cluster
          endpoints:
          - lbEndpoints:
            - endpoint:
                address:
                  socketAddress:
                    address: envoy-gateway
                    portValue: 18000
        name: xds_cluster
        transportSocket:
          name: envoy.transport_sockets.tls
          typedConfig:
            '@type': type.googleapis.com/envoy.extensions.transport_sockets.tls.v3.UpstreamTlsContext
            commonTlsContext:
              tlsCertificateSdsSecretConfigs:
              - name: xds_certificate
                sdsConfig:
                  pathConfigSource:
                    path: /sds/xds-certificate.json
                  resourceApiVersion: V3
              tlsParams:
                tlsMaximumProtocolVersion: TLSv1_3
              validationContextSdsSecretConfig:
                name: xds_trusted_ca
                sdsConfig:
                  pathConfigSource:
                    path: /sds/xds-trusted-ca.json
                  resourceApiVersion: V3
        type: STRICT_DNS
        typedExtensionProtocolOptions:
          envoy.extensions.upstreams.http.v3.HttpProtocolOptions:
            '@type': type.googleapis.com/envoy.extensions.upstreams.http.v3.HttpProtocolOptions
            explicitHttpConfig:
              http2ProtocolOptions: {}
- '@type': type.googleapis.com/envoy.admin.v3.ClustersConfigDump
  dynamicActiveClusters:
  - cluster:
      '@type': type.googleapis.com/envoy.config.cluster.v3.Cluster
      commonLbConfig:
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
- '@type': type.googleapis.com/envoy.admin.v3.ListenersConfigDump
  dynamicListeners:
  - activeState:
      listener:
        '@type': type.googleapis.com/envoy.config.listener.v3.Listener
        accessLog:
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
        name: default-eg-http
- '@type': type.googleapis.com/envoy.admin.v3.RoutesConfigDump
  dynamicRouteConfigs:
  - routeConfig:
      '@type': type.googleapis.com/envoy.config.route.v3.RouteConfiguration
      name: default-eg-http
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
resourceType: all
```  

You can also use the `--type`/`-t` flag to retrieve specific output types. In the below example, we will translate the 
Kubernetes resources (including the Gateway API resources) into xDS `route` resources.

```shell
cat <<EOF | egctl x translate --from gateway-api --to xds -t route -f -
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

```yaml
'@type': type.googleapis.com/envoy.admin.v3.RoutesConfigDump
configKey: default-eg
dynamicRouteConfigs:
- routeConfig:
    '@type': type.googleapis.com/envoy.config.route.v3.RouteConfiguration
    name: default-eg-http
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
resourceType: route
```
