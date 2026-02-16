---
title: "Request Authentication Design"
---

## Overview

[Issue 336][] specifies the need for exposing a user-facing API to configure request authentication. Request
authentication is defined as an authentication mechanism to be enforced by Envoy on a per-request basis. A connection
will be rejected if it contains invalid authentication information, based on the `AuthenticationFilter` API type
proposed in this design document.

Envoy Gateway leverages [Gateway API][] for configuring managed Envoy proxies. Gateway API defines core, extended, and
implementation-specific API [support levels][] for implementers such as Envoy Gateway to expose features. Since
implementing request authentication is not covered by `Core` or `Extended` APIs, an `Implementation-specific` API will
be created for this purpose.

## Goals

* Define an API for configuring request authentication.
* Implement [JWT] as the first supported authentication type.
* Allow users that manage routes, e.g. [HTTPRoute][], to authenticate matching requests before forwarding to a backend
  service.
* Support HTTPRoutes as an Authentication API referent. HTTPRoute provides multiple [extension points][]. The
  [HTTPRouteFilter][] is the extension point supported by the Authentication API.

## Non-Goals

* Allow infrastructure administrators to override or establish default authentication policies.
* Support referents other than HTTPRoute.
* Support Gateway API extension points other than HTTPRouteFilter.

## Use-Cases

These use-cases are presented as an aid for how users may attempt to utilize the outputs of the design. They are not an
exhaustive list of features for authentication support in Envoy Gateway.

As a Service Producer, I need the ability to:
* Authenticate a request before forwarding it to a backend service.
* Have different authentication mechanisms per route rule.
* Choose from different authentication mechanisms supported by Envoy, e.g. OIDC.

### Authentication API Type

The Authentication API type defines authentication configuration for authenticating requests through managed Envoy
proxies.

```go
package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

)

type AuthenticationFilter struct {
	metav1.TypeMeta
	metav1.ObjectMeta

	// Spec defines the desired state of the Authentication type.
	Spec AuthenticationFilterSpec

	// Note: The status sub-resource has been excluded but may be added in the future.
}

// AuthenticationFilterSpec defines the desired state of the AuthenticationFilter type.
// +union
type AuthenticationFilterSpec struct {
	// Type defines the type of authentication provider to use. Supported provider types are:
	//
	// * JWT: A provider that uses JSON Web Token (JWT) for authenticating requests.
	//
	// +unionDiscriminator
	Type AuthenticationFilterType

	// JWT defines the JSON Web Token (JWT) authentication provider type. When multiple
	// jwtProviders are specified, the JWT is considered valid if any of the providers
	// successfully validate the JWT.
	JwtProviders []JwtAuthenticationFilterProvider
}

...
```

Refer to [PR 773][] for the detailed AuthenticationFilter API spec.

The status subresource is not included in the AuthenticationFilter API. Status will be surfaced by an HTTPRoute that
references an AuthenticationFilter. For example, an HTTPRoute will surface the `ResolvedRefs=False` status condition if it
references an AuthenticationFilter that does not exist. It may be beneficial to add AuthenticationFilter status fields in the future
based on defined use-cases. For example, a remote [JWKS][] can be validated based on the specified URI and have an
appropriate status condition surfaced.

#### AuthenticationFilter Example

The following is an AuthenticationFilter example with one JWT authentication provider:

```yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: AuthenticationFilter
metadata:
  name: example
spec:
  type: JWT
  jwtProviders:
  - name: example
    issuer: https://www.example.com
    audiences:
    - foo.com
    remoteJwks:
      uri: https://foo.com/jwt/public-key/jwks.json
      <TBD>
```

__Note:__ `type` is a union type, allowing only one of any supported provider type such as `jwtProviders` to be
specified.

The following is an example HTTPRoute configured to use the above JWT authentication provider:

```yaml
apiVersion: gateway.networking.k8s.io/v1beta1
kind: HTTPRoute
metadata:
  name: example
spec:
  parentRefs:
    - name: eg
  hostnames:
    - www.example.com
  rules:
    - matches:
        - path:
            type: PathPrefix
            value: /foo
      filters:
        - type: ExtensionRef
          extensionRef:
            group: gateway.envoyproxy.io
            kind: AuthenticationFilter
            name: example
      backendRefs:
        - name: backend
          port: 3000
```

Requests for `www.example.com/foo` will be authenticated using the referenced JWT provider before being forwarded to the
backend service named "backend".

## Implementation Details

The JWT authentication type is translated to an Envoy [JWT authentication filter][] and a cluster is created for each
remote [JWKS][]. The following examples provide additional details on how Gateway API and AuthenticationFilter resources are
translated into Envoy configuration.

### Example 1: One Route with One JWT Provider

The following cluster is created from the above HTTPRoute and AuthenticationFilter:

```yaml
dynamic_clusters:
  - name: foo.com|443
    load_assignment:
      cluster_name: foo.com|443
      endpoints:
        - lb_endpoints:
            - endpoint:
                address:
                  socket_address:
                    address: foo.com
                    port_value: 443
    transport_socket:
      name: envoy.transport_sockets.tls
      typed_config:
        "@type": type.googleapis.com/envoy.extensions.transport_sockets.tls.v3.UpstreamTlsContext
        sni: foo.com
        common_tls_context:
          validation_context:
            match_subject_alt_names:
              - exact: "*.foo.com"
            trusted_ca:
              filename: /etc/ssl/certs/ca-certificates.crt
```

A JWT authentication HTTP filter is added to the HTTP Connection Manager. For example:

```yaml
dynamic_resources:
  dynamic_listeners:
    - name: example_listener
      address:
        socket_address:
          address: 1.2.3.4
          port_value: 80
      filter_chains:
        - filters:
            - name: envoy.http_connection_manager
          http_filters:
          - name: envoy.filters.http.jwt_authn
            typed_config:
              "@type": type.googleapis.com/envoy.config.filter.http.jwt_authn.v2alpha.JwtAuthentication
```

This JWT authentication HTTP filter contains two fields:
* The `providers` field specifies how a JWT should be verified, such as where to extract the token, where to fetch the
  public key ([JWKS][]) and where to output its payload. This field is built from the source resource `namespace-name`, and
  the JWT provider name of an AuthenticationFilter.
* The `rules` field specifies matching rules and their requirements. If a request matches a rule, its requirement
  applies. The requirement specifies which JWT providers should be used. This field is built from a HTTPRoute
  `matches` rule that references the AuthenticationFilter. When a referenced Authentication specifies multiple
  `jwtProviders`, the JWT is considered valid if __any__ of the providers successfully validate the JWT.

The following JWT authentication HTTP filter `providers` configuration is created from the above AuthenticationFilter.

```yaml
providers:
   example:
     issuer: https://www.example.com
     audiences:
     - foo.com
     remote_jwks:
       http_uri:
         uri: https://foo.com/jwt/public-key/jwks.json
         cluster: example_jwks_cluster
         timeout: 1s
```

The following JWT authentication HTTP filter `rules` configuration is created from the above HTTPRoute.

```yaml
rules:
  - match:
      prefix: /foo
    requires:
      provider_name: default-example-example
```

### Example 2: Two HTTPRoutes with Different AuthenticationFilters

The following example contains:
* Two HTTPRoutes with different hostnames.
* Each HTTPRoute references a different AuthenticationFilter.
* Each AuthenticationFilter contains a different JWT provider.

```yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: AuthenticationFilter
metadata:
  name: example1
spec:
  type: JWT
  jwtProviders:
  - name: example1
    issuer: https://www.example1.com
    audiences:
    - foo.com
    remoteJwks:
      uri: https://foo.com/jwt/public-key/jwks.json
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: AuthenticationFilter
metadata:
  name: example2
spec:
  type: JWT
  jwtProviders:
    - name: example2
      issuer: https://www.example2.com
      audiences:
        - bar.com
      remoteJwks:
        uri: https://bar.com/jwt/public-key/jwks.json
---
apiVersion: gateway.networking.k8s.io/v1beta1
kind: HTTPRoute
metadata:
  name: example1
spec:
  hostnames:
    - www.example1.com
  parentRefs:
    - group: gateway.networking.k8s.io
      kind: Gateway
      name: eg
  rules:
    - matches:
        - path:
            type: PathPrefix
            value: /foo
      filters:
        - type: ExtensionRef
          extensionRef:
            group: gateway.envoyproxy.io
            kind: AuthenticationFilter
            name: example1
      backendRefs:
        - name: backend
          port: 3000
---
apiVersion: gateway.networking.k8s.io/v1beta1
kind: HTTPRoute
metadata:
  name: example2
spec:
  hostnames:
    - www.example2.com
  parentRefs:
    - group: gateway.networking.k8s.io
      kind: Gateway
      name: eg
  rules:
    - matches:
        - path:
            type: PathPrefix
            value: /bar
      filters:
        - type: ExtensionRef
          extensionRef:
            group: gateway.envoyproxy.io
            kind: AuthenticationFilter
            name: example2
      backendRefs:
        - name: backend2
          port: 3000
```

The following xDS configuration is created from the above example resources:

```yaml
configs:
...
dynamic_listeners:
  - name: default-eg-http
    ...
        default_filter_chain:
          filters:
          - name: envoy.filters.network.http_connection_manager
            typed_config:
              '@type': >-
                type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
              stat_prefix: http
              rds:
                config_source:
                  ...
                route_config_name: default-eg-http
              http_filters:
                - name: envoy.filters.http.jwt_authn
                  typed_config:
                    '@type': >-
                      type.googleapis.com/envoy.config.filter.http.jwt_authn.v2alpha.JwtAuthentication
                    providers:
                      default-example1-example1:
                        issuer: https://www.example1.com
                        audiences:
                          - foo.com
                        remote_jwks:
                          http_uri:
                            uri: https://foo.com/jwt/public-key/jwks.json
                            cluster: default-example1-example1-jwt
                      default-example2-example2:
                        issuer: https://www.example2.com
                        audiences:
                          - bar.com
                        remote_jwks:
                          http_uri:
                            uri: https://bar.com/jwt/public-key/jwks.json
                            cluster: default-example2-example2-jwt
                    rules:
                      - match:
                          exact: /foo
                        requires:
                          provider_name: default-example1-example1
                      - match:
                          exact: /bar
                        requires:
                          provider_name: default-example2-example2
                - name: envoy.filters.http.router
                  typed_config:
                    '@type': >-
                      type.googleapis.com/envoy.extensions.filters.http.router.v3.Router
dynamic_route_configs:
  - route_config:
      '@type': type.googleapis.com/envoy.config.route.v3.RouteConfiguration
      name: default-eg-http
      virtual_hosts:
        - name: default-eg-http
          domains:
            - '*'
          routes:
            - match:
                prefix: /foo
                headers:
                  - name: ':authority'
                    string_match:
                      exact: www.example1.com
              route:
                cluster: default-backend-rule-0-match-0-www.example1.com
            - match:
                prefix: /bar
                headers:
                  - name: ':authority'
                    string_match:
                      exact: www.example2.com
              route:
                cluster: default-backend2-rule-0-match-0-www.example2.com
dynamic_active_clusters:
  - cluster:
      name: default-backend-rule-0-match-0-www.example.com
      ...
        endpoints:
        - locality: {}
          lb_endpoints:
            - endpoint:
                address:
                  socket_address:
                    address: $BACKEND_SERVICE1_IP
                    port_value: 3000
  - cluster:
      '@type': type.googleapis.com/envoy.config.cluster.v3.Cluster
      name: default-backend-rule-1-match-0-www.example.com
      ...
        endpoints:
        - locality: {}
          lb_endpoints:
            - endpoint:
                address:
                  socket_address:
                    address: $BACKEND_SERVICE2_IP
                    port_value: 3000
...
```

__Note:__ The JWT provider cluster and route is omitted from the above example for brevity.

### Implementation Outline

* Update the Kubernetes provider to get/watch AuthenticationFilter resources that are referenced by managed HTTPRoutes.
  Add the referenced AuthenticationFilter object to the resource map and publish it.
* Update the resource translator to include the AuthenticationFilter API in HTTPRoute processing.
* Update the xDS translator to translate an AuthenticationFilter into xDS resources. The translator should perform the
  following:
  * Convert a list of JWT rules from the xds IR into an Envoy JWT filter config.
  * Create a JWT authentication HTTP filter.
  * Build the HTTP Connection Manager (HCM) HTTP filters.
  * Build the HCM.
  * When building the Listener, create an HCM for each filter-chain.

## Adding Authentication Types

Additional authentication types can be added in the future through the `AuthenticationFilterType` API. For
example, to add the `Foo` authentication type:

Define the `Foo` authentication provider:

```go
package v1alpha1

// FooAuthenticationFilterProvider defines the "Foo" authentication filter provider type.
type FooAuthenticationFilterProvider struct {
	// TODO: Define fields of the Foo authentication filter provider type.
}
```

Add the `FooAuthenticationFilterProvider` type to `AuthenticationFilterSpec`:

```go
package v1alpha1

type AuthenticationFilterSpec struct {
	...
	
	// Foo defines the Foo authentication type. For additional
	// details, see:
	//
	//   <INSERT_LINK>
	//
	// +optional
	Foo *FooAuthenticationFilterProvider
}
```

Lastly, add the type to the `AuthenticationType` enum:

```go
// AuthenticationType is a type of authentication provider.
// +kubebuilder:validation:Enum=JWT,FOO
type AuthenticationFilterType string

const (
	// JwtAuthenticationProviderType is the JWT authentication provider type.
	FooAuthenticationFilterProviderType AuthenticationFilterType = "FOO"
)
```

The AuthenticationFilter API should support additional authentication types in the future, for example:
- OAuth2
- OIDC

## Outstanding Questions

- If Envoy Gateway owns the AuthenticationFilter API, is an xDS IR equivalent needed?
- Should local [JWKS][] be implemented before remote [JWKS][]?
- How should Envoy obtain the trusted CA for a remote [JWKS][]?
- Should HTTPS be the only supported scheme for remote [JWKS][]?
- Should OR'ing JWT providers be supported?
- Should Authentication provide status?
- Are the API field validation rules acceptable?

[Issue 336]: https://github.com/envoyproxy/gateway/issues/336
[Gateway API]: https://gateway-api.sigs.k8s.io/
[support levels]: https://gateway-api.sigs.k8s.io/concepts/conformance/?h=extended#2-support-levels
[JWT]: https://jwt.io/
[HTTPRoute]: https://gateway-api.sigs.k8s.io/api-types/httproute/
[extension points]: https://gateway-api.sigs.k8s.io/concepts/api-overview/?h=extension#extension-points
[HTTPRouteFilter]: https://gateway-api.sigs.k8s.io/references/spec/#gateway.networking.k8s.io/v1beta1.HTTPRouteFilter
[JWKS]: https://www.rfc-editor.org/rfc/rfc7517
[JWT authentication filter]: https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/jwt_authn_filter#config-http-filters-jwt-authn
[PR 773]: https://github.com/envoyproxy/gateway/pull/733
