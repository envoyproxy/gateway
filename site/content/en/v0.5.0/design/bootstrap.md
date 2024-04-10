---
title: "Bootstrap Design"
---

## Overview

[Issue 31][] specifies the need for allowing advanced users to specify their custom
Envoy Bootstrap configuration rather than using the default Bootstrap configuration
defined in Envoy Gateway. This allows advanced users to extend Envoy Gateway and
support their custom use cases such setting up tracing and stats configuration
that is not supported by Envoy Gateway.

## Goals

* Define an API field to allow a user to specify a custom Bootstrap
* Provide tooling to allow the user to generate the default Bootstrap configuration
  as well as validate their custom Bootstrap.

## Non Goals

* Allow user to configure only a section of the Bootstrap

## API

Leverage the existing [EnvoyProxy][] resource which can be attached to the [GatewayClass][] using
the [parametersRef][] field, and define a `Bootstrap` field within the resource. If this field is set,
the value is used as the Bootstrap configuration for all managed Envoy Proxies created by Envoy Gateway.

```go
// EnvoyProxySpec defines the desired state of EnvoyProxy.
type EnvoyProxySpec struct {
    ......
	// Bootstrap defines the Envoy Bootstrap as a YAML string.
	// Visit https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/bootstrap/v3/bootstrap.proto#envoy-v3-api-msg-config-bootstrap-v3-bootstrap
	// to learn more about the syntax.
	// If set, this is the Bootstrap configuration used for the managed Envoy Proxy fleet instead of the default Bootstrap configuration
	// set by Envoy Gateway.
	// Some fields within the Bootstrap that are required to communicate with the xDS Server (Envoy Gateway) and receive xDS resources
	// from it are not configurable and will result in the `EnvoyProxy` resource being rejected.
	// Backward compatibility across minor versions is not guaranteed.
	// We strongly recommend using `egctl x translate` to generate a `EnvoyProxy` resource with the `Bootstrap` field set to the default
	// Bootstrap configuration used. You can edit this configuration, and rerun `egctl x translate` to ensure there are no validation errors.
	//
	// +optional
	Bootstrap *string `json:"bootstrap,omitempty"`
}
```

## Tooling

A CLI tool `egctl x translate` will be provided to the user to help generate a working Bootstrap configuration.
Here is an example where a user inputs a `GatewayClass` and the CLI generates the `EnvoyProxy` resource with the `Bootstrap` field populated.

```
cat <<EOF | egctl x translate --from gateway-api --to gateway-api -f -
apiVersion: gateway.networking.k8s.io/v1beta1
kind: GatewayClass
metadata:
  name: eg
spec:
  controllerName: gateway.envoyproxy.io/gatewayclass-controller
---

EOF
```

```
apiVersion: gateway.networking.k8s.io/v1beta1
kind: GatewayClass
metadata:
  name: eg
spec:
  controllerName: gateway.envoyproxy.io/gatewayclass-controller
  parametersRef:
    group: config.gateway.envoyproxy.io/v1alpha1
    kind: EnvoyProxy
    name: with-bootstrap-config
---
apiVersion: config.gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: with-bootstrap-config
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

```

The user can now modify the output, for their use case. Lets say for this example, the user wants to change the admin server port
from `19000` to `18000`, they can do so by editing the previous output and running `egctl x translate` again to see if there any validation
errors. Validation errors should be surfaced in the Status subresource. The internal validator will ensure that the Bootstrap string can be
unmarshalled into the Bootstrap object as well as ensure the user can override certain fields within the Bootstrap configuration such as the
`address` and tls context within the `xds_cluster` which are essential for xDS communication between Envoy Gateway and Envoy Proxy.

```
cat <<EOF | egctl x translate --from gateway-api --to gateway-api -f -
apiVersion: gateway.networking.k8s.io/v1beta1
kind: GatewayClass
metadata:
  name: eg
spec:
  controllerName: gateway.envoyproxy.io/gatewayclass-controller
  parametersRef:
    group: config.gateway.envoyproxy.io/v1alpha1
    kind: EnvoyProxy
    name: with-bootstrap-config
---
apiVersion: config.gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: with-bootstrap-config
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
          port_value: 18000
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

EOF
```

```
apiVersion: gateway.networking.k8s.io/v1beta1
kind: GatewayClass
metadata:
  name: eg
spec:
  controllerName: gateway.envoyproxy.io/gatewayclass-controller
  parametersRef:
    group: config.gateway.envoyproxy.io/v1alpha1
    kind: EnvoyProxy
    name: with-bootstrap-config
---
apiVersion: config.gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: with-bootstrap-config
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
          port_value: 18000
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

```

[Issue 31]: https://github.com/envoyproxy/gateway/issues/31
[EnvoyProxy]: https://gateway.envoyproxy.io/latest/api/config_types.html#envoyproxy
[GatewayClass]: https://gateway-api.sigs.k8s.io/references/spec/#gateway.networking.k8s.io/v1beta1.GatewayClass
[parametersRef]: https://gateway-api.sigs.k8s.io/references/spec/#gateway.networking.k8s.io/v1beta1.ParametersReference 
