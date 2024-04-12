---
title: "Use egctl"
---

`egctl` is a command line tool to provide additional functionality for Envoy Gateway users.



## egctl experimental translate

This subcommand allows users to translate from an input configuration type to an output configuration type.

The `translate` subcommand can translate Kubernetes resources to:
* Gateway API resources
  This is useful in order to see how validation would occur if these resources were applied to Kubernetes. 
  
  Use the `--to gateway-api` parameter to translate to Gateway API resources.
  
* Envoy Gateway intermediate representation (IR)
  This represents Envoy Gateway's translation of the Gateway API resources.
  
  Use the `--to ir` parameter to translate to Envoy Gateway intermediate representation.
  
* Envoy Proxy xDS
  This is the xDS configuration provided to Envoy Proxy.
  
  Use the `--to xds` parameter to translate to Envoy Proxy xDS.
  

In the below example, we will translate the Kubernetes resources (including the Gateway API resources) into xDS
resources.

```shell
cat <<EOF | egctl x translate --from gateway-api --to xds -f -
apiVersion: gateway.networking.k8s.io/v1
kind: GatewayClass
metadata:
  name: eg
spec:
  controllerName: gateway.envoyproxy.io/gatewayclass-controller
---
apiVersion: gateway.networking.k8s.io/v1
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
apiVersion: gateway.networking.k8s.io/v1
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
        - www.example.com
        name: default-eg-http-www.example.com
        routes:
        - match:
            prefix: /
          route:
            cluster: default-backend-rule-0-match-0-www.example.com
resourceType: all
```  

You can also use the `--type`/`-t` flag to retrieve specific output types. In the below example, we will translate the Kubernetes resources (including the Gateway API resources) into xDS `route` resources.

```shell
cat <<EOF | egctl x translate --from gateway-api --to xds -t route -f -
apiVersion: gateway.networking.k8s.io/v1
kind: GatewayClass
metadata:
  name: eg
spec:
  controllerName: gateway.envoyproxy.io/gatewayclass-controller
---
apiVersion: gateway.networking.k8s.io/v1
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
apiVersion: gateway.networking.k8s.io/v1
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
      - www.example.com
      name: default-eg-http
      routes:
      - match:
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
apiVersion: gateway.networking.k8s.io/v1
kind: GatewayClass
metadata:
  name: eg
spec:
  controllerName: gateway.envoyproxy.io/gatewayclass-controller
---
apiVersion: gateway.networking.k8s.io/v1
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
apiVersion: gateway.networking.k8s.io/v1
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

You can see the output contains a [EnvoyProxy][] resource that
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
      group: gateway.envoyproxy.io
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
apiVersion: gateway.networking.k8s.io/v1
kind: GatewayClass
metadata:
  name: eg
spec:
  controllerName: gateway.envoyproxy.io/gatewayclass-controller
---
apiVersion: gateway.networking.k8s.io/v1
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
apiVersion: gateway.networking.k8s.io/v1
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
apiVersion: gateway.networking.k8s.io/v1
kind: GatewayClass
metadata:
  name: eg
spec:
  controllerName: gateway.envoyproxy.io/gatewayclass-controller
---
apiVersion: gateway.networking.k8s.io/v1
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
apiVersion: gateway.networking.k8s.io/v1
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

## egctl experimental status

This subcommand allows users to show the summary of the status of specific or all resource types, in order to quickly find
out the status of any resources.

By default, `egctl x status` display all the conditions for one resource type. You can either add `--quiet` to only
display the latest condition, or add `--verbose` to display more details about current status.

{{% alert title="Note" color="primary" %}}

Currently, this subcommand only supports: `GatewayClass`, `Gateway`, `HTTPRoute`, `GRPCRoute`,
`TLSRoute`, `TCPRoute`, `UDPRoute`, `BackendTLSPolicy`, 
`BackendTrafficPolicy`, `ClientTrafficPolicy`, `EnvoyPatchPolicy`, `SecurityPolicy` resource types and `all`.

{{% /alert %}}

Some examples of this command after installing [Multi-tenancy][] example manifest:

- Show the summary of GatewayClass.

```console
~ egctl x status gatewayclass

NAME           TYPE       STATUS    REASON
eg-marketing   Accepted   True      Accepted
eg-product     Accepted   True      Accepted
```

- Show the summary of all resource types under all namespaces, the resource type with empty status will be ignored.

```console
~ egctl x status all -A

NAME                        TYPE       STATUS    REASON
gatewayclass/eg-marketing   Accepted   True      Accepted
gatewayclass/eg-product     Accepted   True      Accepted

NAMESPACE   NAME         TYPE         STATUS    REASON
marketing   gateway/eg   Programmed   True      Programmed
                         Accepted     True      Accepted
product     gateway/eg   Programmed   True      Programmed
                         Accepted     True      Accepted

NAMESPACE   NAME                TYPE           STATUS    REASON
marketing   httproute/backend   ResolvedRefs   True      ResolvedRefs
                                Accepted       True      Accepted
product     httproute/backend   ResolvedRefs   True      ResolvedRefs
                                Accepted       True      Accepted
```

- Show the summary of all the Gateways with details under all namespaces.

```console
~ egctl x status gateway --verbose --all-namespaces

NAMESPACE   NAME      TYPE         STATUS    REASON       MESSAGE                                                                    OBSERVED GENERATION   LAST TRANSITION TIME
marketing   eg        Programmed   True      Programmed   Address assigned to the Gateway, 1/1 envoy Deployment replicas available   1                     2024-02-02 18:17:14 +0800 CST
                      Accepted     True      Accepted     The Gateway has been scheduled by Envoy Gateway                            1                     2024-02-01 17:50:39 +0800 CST
product     eg        Programmed   True      Programmed   Address assigned to the Gateway, 1/1 envoy Deployment replicas available   1                     2024-02-02 18:17:14 +0800 CST
                      Accepted     True      Accepted     The Gateway has been scheduled by Envoy Gateway                            1                     2024-02-01 17:52:42 +0800 CST
```

- Show the summary of latest Gateways condition under `product` namespace.

```console
~ egctl x status gateway --quiet -n product

NAME      TYPE         STATUS    REASON
eg        Programmed   True      Programmed
```

- Show the summary of latest HTTPRoutes condition under all namespaces.

```console
~ egctl x status httproute --quiet --all-namespaces

NAMESPACE   NAME      TYPE           STATUS    REASON
marketing   backend   ResolvedRefs   True      ResolvedRefs
product     backend   ResolvedRefs   True      ResolvedRefs
```


[Multi-tenancy]: ../deployment-mode#multi-tenancy
[EnvoyProxy]: ../../../api/extension_types#envoyproxy


## egctl experimental dashboard

This subcommand streamlines the process for users to access the Envoy admin dashboard. By executing the following command:

```bash
egctl x dashboard envoy-proxy -n envoy-gateway-system envoy-engw-eg-a9c23fbb-558f94486c-82wh4
```

You will see the following output:

```bash
egctl x dashboard envoy-proxy -n envoy-gateway-system envoy-engw-eg-a9c23fbb-558f94486c-82wh4
http://localhost:19000
```

the Envoy admin dashboard will automatically open in your default web browser. This eliminates the need to manually locate and expose the admin port.


## egctl experimental install

This subcommand can be used to install envoy-gateway.

```bash
egctl x install
```

By default, this will install both the envoy-gateway workload resource and the required gateway-api and envoy-gatewayCRDs.

We can specify to install only workload resources via `--skip-crds`

```bash
egctl x install --skip-crds
```

We can specify to install only CRDs resources via `--only-crds`

```bash
egctl x install --only-crds
```

We can specify `--release-name` and `--namespace` to install envoy-gateway in different places to support multi-tenant mode.
> Note: If CRDs are already installed, then we need to specify `--skip-crds` to avoid repeated installation of CRDs resources.

```bash
egctl x install --release-name shop-backend --namespace shop
```


## egctl experimental uninstall

This subcommand can be used to uninstall envoy-gateway.

```bash
egctl x uninstall
```

By default, this will only uninstall the envoy-gateway workload resource, if we want to also uninstall CRDs, we need to specify `--with-crds`

```bash
egctl x uninstall --with-crds
```