---
title: "Use egctl"
---

`egctl` is a command line tool to provide additional functionality for Envoy Gateway users.

## Installing egctl

This guide shows how to install the egctl CLI. egctl can be installed either from source, or from pre-built binary releases.

### From The Envoy Gateway Project

The Envoy Gateway project provides two ways to fetch and install egctl. These are the official methods to get egctl releases. Installation through those methods can be found below the official methods.

### From the Binary Releases

Every [release](https://github.com/envoyproxy/gateway/releases) of egctl provides binary releases for a variety of OSes. These binary versions can be manually downloaded and installed.

1. Download your [desired version](https://github.com/envoyproxy/gateway/releases)
2. Unpack it (tar -zxvf egctl_v0.5.0_linux_amd64.tar.gz)
3. Find the egctl binary in the unpacked directory, and move it to its desired destination (mv bin/linux/amd64/egctl /usr/local/bin/egctl)

From there, you should be able to run: `egctl help`.

### From Script

`egctl` now has an installer script that will automatically grab the v0.5.0 release version of egctl and install it locally.

You can fetch that script, and then execute it locally. It's well documented so that you can read through it and understand what it is doing before you run it.

```shell
curl -fsSL -o get-egctl.sh https://gateway.envoyproxy.io/get-egctl.sh

chmod +x get-egctl.sh

# get help info of the 
bash get-egctl.sh --help

# install the v0.5.0 development version of egctl
bash VERSION=v0.5.0 get-egctl.sh
```

Yes, you can just use the below command if you want to live on the edge.

```shell
curl https://gateway.envoyproxy.io/get-egctl.sh | VERSION=v0.5.0 bash 
```

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
      - connectTimeout: 10s
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
      connectTimeout: 10s
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

You can also use the `--type`/`-t` flag to retrieve specific output types. In the below example, we will translate the Kubernetes resources (including the Gateway API resources) into xDS `route` resources.

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

### Add Missing Resources

You can pass the `--add-missing-resources` flag to use dummy non Gateway API resources instead of specifying them explicitly.

For example, this will provide the similar result as the above:

```shell
cat <<EOF | egctl x translate --add-missing-resources --from gateway-api --to gateway-api -t route -f -
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

You can see the output contains a [EnvoyProxy](https://gateway.envoyproxy.io/v0.5.0/api/config_types.html#envoyproxy) resource that
can be used as a starting point to modify the xDS bootstrap resource for the managed Envoy Proxy fleet.

```yaml
envoyProxy:
  metadata:
    creationTimestamp: null
    name: default-envoy-proxy
    namespace: envoy-gateway-system
  spec:
    bootstrap: |
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
      static_resources:
        clusters:
        - connect_timeout: 10s
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
              ads: {}
              resource_api_version: V3
            name: runtime-0
    logging: {}
  status: {}
gatewayClass:
  metadata:
    creationTimestamp: null
    name: eg
    namespace: envoy-gateway-system
  spec:
    controllerName: gateway.envoyproxy.io/gatewayclass-controller
    parametersRef:
      group: config.gateway.envoyproxy.io
      kind: EnvoyProxy
      name: default-envoy-proxy
      namespace: envoy-gateway-system
  status:
    conditions:
    - lastTransitionTime: "2023-04-19T20:30:46Z"
      message: Valid GatewayClass
      reason: Accepted
      status: "True"
      type: Accepted
gateways:
- metadata:
    creationTimestamp: null
    name: eg
    namespace: default
  spec:
    gatewayClassName: eg
    listeners:
    - name: http
      port: 80
      protocol: HTTP
  status:
    listeners:
    - attachedRoutes: 1
      conditions:
      - lastTransitionTime: "2023-04-19T20:30:46Z"
        message: Sending translated listener configuration to the data plane
        reason: Programmed
        status: "True"
        type: Programmed
      - lastTransitionTime: "2023-04-19T20:30:46Z"
        message: Listener has been successfully translated
        reason: Accepted
        status: "True"
        type: Accepted
      name: http
      supportedKinds:
      - group: gateway.networking.k8s.io
        kind: HTTPRoute
      - group: gateway.networking.k8s.io
        kind: GRPCRoute
httpRoutes:
- metadata:
    creationTimestamp: null
    name: backend
    namespace: default
  spec:
    hostnames:
    - www.example.com
    parentRefs:
    - name: eg
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
  status:
    parents:
    - conditions:
      - lastTransitionTime: "2023-04-19T20:30:46Z"
        message: Route is accepted
        reason: Accepted
        status: "True"
        type: Accepted
      - lastTransitionTime: "2023-04-19T20:30:46Z"
        message: Resolved all the Object references for the Route
        reason: ResolvedRefs
        status: "True"
        type: ResolvedRefs
      controllerName: gateway.envoyproxy.io/gatewayclass-controller
      parentRef:
        name: eg
```

Sometimes you might find that egctl doesn't provide an expected result. For example, the following example provides an empty route resource:

```shell
cat <<EOF | egctl x translate --from gateway-api --type route --to xds -f -
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
spec:
  gatewayClassName: eg
  listeners:
    - name: tls
      protocol: TLS
      port: 8443
---
apiVersion: gateway.networking.k8s.io/v1beta1
kind: HTTPRoute
metadata:
  name: backend
spec:
  parentRefs:
    - name: eg
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
xds:
  envoy-gateway-system-eg:
    '@type': type.googleapis.com/envoy.admin.v3.RoutesConfigDump
```

### Validating Gateway API Configuration

You can add an additional target `gateway-api` to show the processed Gateway API resources. For example, translating the above resources with the new argument shows that the HTTPRoute is rejected because there is no ready listener for it:

```shell
cat <<EOF | egctl x translate --from gateway-api --type route --to gateway-api,xds -f -
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
spec:
  gatewayClassName: eg
  listeners:
    - name: tls
      protocol: TLS
      port: 8443
---
apiVersion: gateway.networking.k8s.io/v1beta1
kind: HTTPRoute
metadata:
  name: backend
spec:
  parentRefs:
    - name: eg
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
gatewayClass:
  metadata:
    creationTimestamp: null
    name: eg
    namespace: envoy-gateway-system
  spec:
    controllerName: gateway.envoyproxy.io/gatewayclass-controller
  status:
    conditions:
    - lastTransitionTime: "2023-04-19T20:54:52Z"
      message: Valid GatewayClass
      reason: Accepted
      status: "True"
      type: Accepted
gateways:
- metadata:
    creationTimestamp: null
    name: eg
    namespace: envoy-gateway-system
  spec:
    gatewayClassName: eg
    listeners:
    - name: tls
      port: 8443
      protocol: TLS
  status:
    listeners:
    - attachedRoutes: 0
      conditions:
      - lastTransitionTime: "2023-04-19T20:54:52Z"
        message: Listener must have TLS set when protocol is TLS.
        reason: Invalid
        status: "False"
        type: Programmed
      name: tls
      supportedKinds:
      - group: gateway.networking.k8s.io
        kind: TLSRoute
httpRoutes:
- metadata:
    creationTimestamp: null
    name: backend
    namespace: envoy-gateway-system
  spec:
    parentRefs:
    - name: eg
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
  status:
    parents:
    - conditions:
      - lastTransitionTime: "2023-04-19T20:54:52Z"
        message: There are no ready listeners for this parent ref
        reason: NoReadyListeners
        status: "False"
        type: Accepted
      - lastTransitionTime: "2023-04-19T20:54:52Z"
        message: Service envoy-gateway-system/backend not found
        reason: BackendNotFound
        status: "False"
        type: ResolvedRefs
      controllerName: gateway.envoyproxy.io/gatewayclass-controller
      parentRef:
        name: eg
xds:
  envoy-gateway-system-eg:
    '@type': type.googleapis.com/envoy.admin.v3.RoutesConfigDump
```
