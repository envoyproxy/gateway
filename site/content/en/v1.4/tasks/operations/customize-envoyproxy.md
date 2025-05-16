---
title: "Customize EnvoyProxy"
---

Envoy Gateway provides an [EnvoyProxy][] CRD that can be linked to the ParametersRef
in a Gateway and GatewayClass, allowing cluster admins to customize the managed EnvoyProxy Deployment and
Service. To learn more about GatewayClass and ParametersRef, please refer to [Gateway API documentation][].

## Prerequisites

{{< boilerplate prerequisites >}}

Before you start, you need to add `Infrastructure.ParametersRef` in Gateway, and refer to EnvoyProxy Config:
**Note**: `MergeGateways` cannot be set to `true` in your EnvoyProxy config if attaching to the Gateway.

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: eg
spec:
  gatewayClassName: eg
  infrastructure:
    parametersRef:
      group: gateway.envoyproxy.io
      kind: EnvoyProxy
      name: custom-proxy-config
  listeners:
    - name: http
      protocol: HTTP
      port: 80
EOF
```

{{% /tab %}}
{{% tab header="Apply from file" %}}
Save and apply the following resource to your cluster:

```yaml
---
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: eg
spec:
  gatewayClassName: eg
  infrastructure:
    parametersRef:
      group: gateway.envoyproxy.io
      kind: EnvoyProxy
      name: custom-proxy-config
  listeners:
    - name: http
      protocol: HTTP
      port: 80
```

{{% /tab %}}
{{< /tabpane >}}

You can also attach the EnvoyProxy resource to the GatewayClass using the `parametersRef` field.
This configuration is discouraged if you plan on creating multiple Gateways linking to the same
GatewayClass and would like different infrastructure configurations for each of them.

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.networking.k8s.io/v1
kind: GatewayClass
metadata:
  name: eg
spec:
  controllerName: gateway.envoyproxy.io/gatewayclass-controller
  parametersRef:
    group: gateway.envoyproxy.io
    kind: EnvoyProxy
    name: custom-proxy-config
    namespace: default
EOF
```

{{% /tab %}}
{{% tab header="Apply from file" %}}
Save and apply the following resource to your cluster:

```yaml
---
apiVersion: gateway.networking.k8s.io/v1
kind: GatewayClass
metadata:
  name: eg
spec:
  controllerName: gateway.envoyproxy.io/gatewayclass-controller
  parametersRef:
    group: gateway.envoyproxy.io
    kind: EnvoyProxy
    name: custom-proxy-config
    namespace: default
```

{{% /tab %}}
{{< /tabpane >}}

## Customize EnvoyProxy Deployment Replicas

You can customize the EnvoyProxy Deployment Replicas via EnvoyProxy Config like:

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: custom-proxy-config
  namespace: default
spec:
  provider:
    type: Kubernetes
    kubernetes:
      envoyDeployment:
        replicas: 2
EOF
```

{{% /tab %}}
{{% tab header="Apply from file" %}}
Save and apply the following resource to your cluster:

```yaml
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: custom-proxy-config
  namespace: default
spec:
  provider:
    type: Kubernetes
    kubernetes:
      envoyDeployment:
        replicas: 2
```

{{% /tab %}}
{{< /tabpane >}}

After you apply the config, you should see the replicas of envoyproxy changes to 2.
And also you can dynamically change the value.

``` shell
kubectl get deployment -l gateway.envoyproxy.io/owning-gateway-name=eg -n envoy-gateway-system
```

## Customize EnvoyProxy Image

You can customize the EnvoyProxy Image via EnvoyProxy Config like:

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: custom-proxy-config
  namespace: default
spec:
  provider:
    type: Kubernetes
    kubernetes:
      envoyDeployment:
        container:
          image: envoyproxy/envoy:v1.25-latest
EOF
```

{{% /tab %}}
{{% tab header="Apply from file" %}}
Save and apply the following resource to your cluster:

```yaml
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: custom-proxy-config
  namespace: default
spec:
  provider:
    type: Kubernetes
    kubernetes:
      envoyDeployment:
        container:
          image: envoyproxy/envoy:v1.25-latest
```

{{% /tab %}}
{{< /tabpane >}}

After applying the config, you can get the deployment image, and see it has changed.

## Customize EnvoyProxy Pod Annotations

You can customize the EnvoyProxy Pod Annotations via EnvoyProxy Config like:

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: custom-proxy-config
  namespace: default
spec:
  provider:
    type: Kubernetes
    kubernetes:
      envoyDeployment:
        pod:
          annotations:
            custom1: deploy-annotation1
            custom2: deploy-annotation2
EOF
```

{{% /tab %}}
{{% tab header="Apply from file" %}}
Save and apply the following resource to your cluster:

```yaml
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: custom-proxy-config
  namespace: default
spec:
  provider:
    type: Kubernetes
    kubernetes:
      envoyDeployment:
        pod:
          annotations:
            custom1: deploy-annotation1
            custom2: deploy-annotation2
```

{{% /tab %}}
{{< /tabpane >}}

After applying the config, you can get the envoyproxy pods, and see new annotations has been added.

## Customize EnvoyProxy Deployment Resources

You can customize the EnvoyProxy Deployment Resources via EnvoyProxy Config like:

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: custom-proxy-config
  namespace: default
spec:
  provider:
    type: Kubernetes
    kubernetes:
      envoyDeployment:
        container:
          resources:
            requests:
              cpu: 150m
              memory: 640Mi
            limits:
              cpu: 500m
              memory: 1Gi
EOF
```

{{% /tab %}}
{{% tab header="Apply from file" %}}
Save and apply the following resource to your cluster:

```yaml
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: custom-proxy-config
  namespace: default
spec:
  provider:
    type: Kubernetes
    kubernetes:
      envoyDeployment:
        container:
          resources:
            requests:
              cpu: 150m
              memory: 640Mi
            limits:
              cpu: 500m
              memory: 1Gi
```

{{% /tab %}}
{{< /tabpane >}}

## Customize EnvoyProxy Deployment Env

You can customize the EnvoyProxy Deployment Env via EnvoyProxy Config like:

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: custom-proxy-config
  namespace: default
spec:
  provider:
    type: Kubernetes
    kubernetes:
      envoyDeployment:
        container:
          env:
          - name: env_a
            value: env_a_value
          - name: env_b
            value: env_b_value
EOF
```

{{% /tab %}}
{{% tab header="Apply from file" %}}
Save and apply the following resource to your cluster:

```yaml
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: custom-proxy-config
  namespace: default
spec:
  provider:
    type: Kubernetes
    kubernetes:
      envoyDeployment:
        container:
          env:
          - name: env_a
            value: env_a_value
          - name: env_b
            value: env_b_value
```

{{% /tab %}}
{{< /tabpane >}}

> Envoy Gateway has provided two initial `env` `ENVOY_GATEWAY_NAMESPACE` and `ENVOY_POD_NAME` for envoyproxy container.

After applying the config, you can get the envoyproxy deployment, and see resources has been changed.

## Customize EnvoyProxy Deployment Volumes or VolumeMounts

You can customize the EnvoyProxy Deployment Volumes or VolumeMounts via EnvoyProxy Config like:

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: custom-proxy-config
  namespace: default
spec:
  provider:
    type: Kubernetes
    kubernetes:
      envoyDeployment:
        container:
          volumeMounts:
          - mountPath: /certs
            name: certs
            readOnly: true
        pod:
          volumes:
          - name: certs
            secret:
              secretName: envoy-cert
EOF
```

{{% /tab %}}
{{% tab header="Apply from file" %}}
Save and apply the following resource to your cluster:

```yaml
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: custom-proxy-config
  namespace: default
spec:
  provider:
    type: Kubernetes
    kubernetes:
      envoyDeployment:
        container:
          volumeMounts:
          - mountPath: /certs
            name: certs
            readOnly: true
        pod:
          volumes:
          - name: certs
            secret:
              secretName: envoy-cert
```

{{% /tab %}}
{{< /tabpane >}}

After applying the config, you can get the envoyproxy deployment, and see resources has been changed.

## Customize EnvoyProxy Service Annotations

You can customize the EnvoyProxy Service Annotations via EnvoyProxy Config like:

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: custom-proxy-config
  namespace: default
spec:
  provider:
    type: Kubernetes
    kubernetes:
      envoyService:
        annotations:
          custom1: svc-annotation1
          custom2: svc-annotation2
EOF
```

{{% /tab %}}
{{% tab header="Apply from file" %}}
Save and apply the following resource to your cluster:

```yaml
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: custom-proxy-config
  namespace: default
spec:
  provider:
    type: Kubernetes
    kubernetes:
      envoyService:
        annotations:
          custom1: svc-annotation1
          custom2: svc-annotation2
```

{{% /tab %}}
{{< /tabpane >}}

After applying the config, you can get the envoyproxy service, and see annotations has been added.

## Customize EnvoyProxy Bootstrap Config

You can customize the EnvoyProxy bootstrap config via EnvoyProxy Config.
There are three ways to customize it:

* Replace: the whole bootstrap config will be replaced by the config you provided.
* Merge: the config you provided will be merged into the default bootstrap config.
* JSONPatch: the list of JSON Patches you provided will be applied to the bootstrap config. JSON Patch is a standard format specified in [RFC 6902](https://datatracker.ietf.org/doc/html/rfc6902/).

{{< tabpane text=true >}}
{{% tab header="Replace: apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: custom-proxy-config
  namespace: default 
spec:
  bootstrap:
    type: Replace
    value: |
      admin:
        access_log:
        - name: envoy.access_loggers.file
          typed_config:
            "@type": type.googleapis.com/envoy.extensions.access_loggers.file.v3.FileAccessLog
            path: /dev/null
        address:
          socket_address:
            address: 127.0.0.1
            port_value: 20000
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
EOF
```

{{% /tab %}}
{{% tab header="Replace: apply from file" %}}
Save and apply the following resource to your cluster:

```yaml
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: custom-proxy-config
  namespace: default
spec:
  bootstrap:
    type: Replace
    value: |
      admin:
        access_log:
        - name: envoy.access_loggers.file
          typed_config:
            "@type": type.googleapis.com/envoy.extensions.access_loggers.file.v3.FileAccessLog
            path: /dev/null
        address:
          socket_address:
            address: 127.0.0.1
            port_value: 20000
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
```

{{% /tab %}}
{{% tab header="Merge: apply from stdin" %}}
Save and apply the following resource to your cluster:

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: custom-proxy-config
  namespace: default
spec:
  bootstrap:
    type: Merge
    value: |
      layered_runtime:
        layers:
        - name: "static-runtime"
          static_layer:
            re2.max_program_size.error_level: 1000
EOF
```

{{% /tab %}}
{{% tab header="Merge: apply from file" %}}
Save and apply the following resource to your cluster:

```yaml
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: custom-proxy-config
  namespace: default
spec:
  bootstrap:
    type: Merge
    value: |
      layered_runtime:
        layers:
        - name: "static-runtime"
          static_layer:
            re2.max_program_size.error_level: 1000
```

{{% /tab %}}
{{% tab header="JSONPatch: apply from stdin" %}}
Save and apply the following resource to your cluster:

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: custom-proxy-config
  namespace: default 
spec:
  bootstrap:
    type: JSONPatch
    jsonPatches:
    - {"op": "add", "path": "/static_resources/clusters/0/dns_lookup_family", "value": "V4_ONLY"}
    - {"op": "replace", "path": "/admin/address/socket_address/port_value", "value": 9901}
EOF
```

{{% /tab %}}
{{% tab header="JSONPatch: apply from file" %}}
Save and apply the following resource to your cluster:

```yaml
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: custom-proxy-config
  namespace: default
spec:
  bootstrap:
    type: JSONPatch
    jsonPatches:
      - {"op": "add", "path": "/static_resources/clusters/0/dns_lookup_family", "value": "V4_ONLY"}
      - {"op": "replace", "path": "/admin/address/socket_address/port_value", "value": 9901}
```

{{% /tab %}}
{{< /tabpane >}}

You can use [egctl x translate][]
to get the default xDS Bootstrap configuration used by Envoy Gateway.

After applying the config, the bootstrap config will be overridden by the new config you provided.
Any errors in the configuration will be surfaced as status within the `GatewayClass` resource.
You can also validate this configuration using [egctl x translate][].

## Customize EnvoyProxy Horizontal Pod Autoscaler

You can enable [Horizontal Pod Autoscaler](https://github.com/envoyproxy/gateway/issues/703) for EnvoyProxy Deployment. However, before enabling the HPA for EnvoyProxy, please ensure that the [metrics-server](https://github.com/kubernetes-sigs/metrics-server) component is installed in the cluster.

Once confirmed, you can apply it via EnvoyProxy Config as shown below:

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: custom-proxy-config
  namespace: default 
spec:
  provider:
    type: Kubernetes
    kubernetes:
      envoyHpa:
        minReplicas: 2
        maxReplicas: 10
        metrics:
          - resource:
              name: cpu
              target:
                averageUtilization: 60
                type: Utilization
            type: Resource
EOF
```

{{% /tab %}}
{{% tab header="Apply from file" %}}
Save and apply the following resource to your cluster:

```yaml
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: custom-proxy-config
  namespace: default 
spec:
  provider:
    type: Kubernetes
    kubernetes:
      envoyHpa:
        minReplicas: 2
        maxReplicas: 10
        metrics:
          - resource:
              name: cpu
              target:
                averageUtilization: 60
                type: Utilization
            type: Resource
```

{{% /tab %}}
{{< /tabpane >}}

After applying the config, the EnvoyProxy HPA (Horizontal Pod Autoscaler) is generated. However, upon activating the EnvoyProxy's HPA, the Envoy Gateway will no longer reference the `replicas` field specified in the `envoyDeployment`, as outlined [here](#customize-envoyproxy-deployment-replicas).

## Customize EnvoyProxy Command line options

You can customize the EnvoyProxy Command line options via `spec.extraArgs` in EnvoyProxy Config.
For example, the following configuration will add `--disable-extensions` arg in order to disable `envoy.access_loggers/envoy.access_loggers.wasm` extension:

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: custom-proxy-config
  namespace: default 
spec:
  extraArgs:
    - --disable-extensions envoy.access_loggers/envoy.access_loggers.wasm 
EOF
```

{{% /tab %}}
{{% tab header="Apply from file" %}}
Save and apply the following resource to your cluster:

```yaml
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: custom-proxy-config
  namespace: default 
spec:
  extraArgs:
    - --disable-extensions envoy.access_loggers/envoy.access_loggers.wasm 
```

{{% /tab %}}
{{< /tabpane >}}

## Customize EnvoyProxy with Patches

You can customize the EnvoyProxy using patches.

### Patching Deployment for EnvoyProxy

For example, the following configuration will add resource limits to the `envoy` and the `shutdown-manager` containers in the `envoyproxy` deployment:

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: eg
  namespace: default 
spec:
  provider:
    type: Kubernetes
    kubernetes:
      envoyDeployment:
        patch:
          type: StrategicMerge
          value:
            spec:
              template:
                spec:
                  containers:
                  - name: envoy
                    resources:
                      limits:
                        cpu: 500m
                        memory: 1024Mi
                  - name: shutdown-manager
                    resources:
                      limits:
                        cpu: 200m
                        memory: 1024Mi
EOF
```

{{% /tab %}}
{{% tab header="Apply from file" %}}
Save and apply the following resource to your cluster:

```yaml
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: eg
  namespace: default
spec:
  provider:
    type: Kubernetes
    kubernetes:
      envoyDeployment:
        patch:
          type: StrategicMerge
          value:
            spec:
              template:
                spec:
                  containers:
                  - name: envoy
                    resources:
                      limits:
                        cpu: 500m
                        memory: 1024Mi
                  - name: shutdown-manager
                    resources:
                      limits:
                        cpu: 200m
                        memory: 1024Mi
```

{{% /tab %}}
{{< /tabpane >}}

After applying the configuration, you will see the change in both containers in the `envoyproxy` deployment.

### Patching Service for EnvoyProxy

For example, the following configuration will add an annotation for the `envoyproxy` service:

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: eg
  namespace: default
spec:
  provider:
    type: Kubernetes
    kubernetes:
      envoyService:
        patch:
          type: StrategicMerge
          value:
            metadata:
              annotations:
                custom-annotation: foobar
EOF
```

{{% /tab %}}
{{% tab header="Apply from file" %}}
Save and apply the following resource to your cluster:

```yaml
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: eg
  namespace: default
spec:
  provider:
    type: Kubernetes
    kubernetes:
      envoyService:
        patch:
          type: StrategicMerge
          value:
            metadata:
              annotations:
                custom-annotation: foobar
```

{{% /tab %}}
{{< /tabpane >}}

After applying the configuration, you will see the `custom-annotation: foobar` has been added to the `envoyproxy` service.

## Customize Filter Order

Under the hood, Envoy Gateway uses a series of [Envoy HTTP filters](https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/http_filters)
to process HTTP requests and responses, and to apply various policies.

By default, Envoy Gateway applies the following filters in the order shown:
* envoy.filters.http.fault
* envoy.filters.http.cors
* envoy.filters.http.ext_authz
* envoy.filters.http.basic_authn
* envoy.filters.http.oauth2
* envoy.filters.http.jwt_authn
* envoy.filters.http.ext_proc
* envoy.filters.http.wasm
* envoy.filters.http.rbac
* envoy.filters.http.local_ratelimit
* envoy.filters.http.ratelimit
* envoy.filters.http.router

The default order in which these filters are applied is opinionated and may not suit all use cases. 
To address this, Envoy Gateway allows you to adjust the execution order of these filters with the `filterOrder` field in the [EnvoyProxy][] resource.

`filterOrder` is a list of customized filter order configurations. Each configuration can specify a filter
name and a filter to place it before or after. These configurations are applied in the order they are listed.
If a filter occurs in multiple configurations, the final order is the result of applying all these configurations in order.
To avoid conflicts, it is recommended to only specify one configuration per filter.

For example, the following configuration moves the `envoy.filters.http.wasm` filter before the `envoy.filters.http.jwt_authn`
filter and the `envoy.filters.http.cors` filter after the `envoy.filters.http.basic_authn` filter:

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: custom-proxy-config
  namespace: default
spec:
  filterOrder:
    - name: envoy.filters.http.wasm
      before: envoy.filters.http.jwt_authn
    - name: envoy.filters.http.cors
      after: envoy.filters.http.basic_authn
EOF
```

{{% /tab %}}
{{% tab header="Apply from file" %}}
Save and apply the following resource to your cluster:

```yaml
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: custom-proxy-config
  namespace: default
spec:
  filterOrder:
    - name: envoy.filters.http.wasm
      before: envoy.filters.http.jwt_authn
    - name: envoy.filters.http.cors
      after: envoy.filters.http.basic_authn
```

{{% /tab %}}
{{< /tabpane >}}

## Customize EnvoyProxy IP Family

You can customize the IP family configuration for EnvoyProxy via the EnvoyProxy Config.
This allows the Envoy Proxy fleet to serve external clients over IPv4 as well as IPv6.

The below configuration sets the `ipFamily` to `DualStack` to allow ingressing IPv4 as well as IPv6 traffic.

**Note**: Envoy Gateway relies on the [Service](https://kubernetes.io/docs/concepts/services-networking/dual-stack/#services) spec of the BackendRef resource (linked to xRoutes) to decide which type of IP addresses to use to route to them.

{{< tabpane text=true >}}
{{% tab header="Apply from stdin" %}}

```shell
cat <<EOF | kubectl apply -f -
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: custom-proxy-config
  namespace: default
spec:
  ipFamily: DualStack
EOF
```

{{% /tab %}}

{{% tab header="Apply from file" %}}
Save and apply the following resource to your cluster:

```yaml
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: custom-proxy-config
  namespace: default
spec:
  ipFamily: DualStack  # Supports: IPv4, IPv6, or DualStack
```

{{% /tab %}}
{{< /tabpane >}}

After applying the config, the EnvoyProxy deployment will be configured to use the specified IP family. When set to `DualStack`, both IPv4 and IPv6 networking will be enabled.

**Note**: Your cluster must support the selected IP family configuration. For DualStack support, ensure your Kubernetes cluster is properly configured for dual-stack networking.

[Gateway API documentation]: https://gateway-api.sigs.k8s.io/
[EnvoyProxy]: ../../../api/extension_types#envoyproxy
[egctl x translate]: ../operations/egctl#egctl-experimental-translate