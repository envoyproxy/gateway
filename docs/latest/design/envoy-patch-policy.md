# EnvoyPatchPolicy

## Overview

This design introduces the `EnvoyPatchPolicy` API allowing users to modify the generated Envoy xDS Configuration
that Envoy Gateway generates before sending it to Envoy Proxy.

Envoy Gateway allows users to configure networking and security intent using the
upstream [Gateway API][] as well as implementation specific [Extension APIs][] defined in this project
to provide a more batteries included experience for application developers.
* These APIs are an abstacted version of the underlying Envoy xDS API to provide a a better user experience for the application developer, exposing and setting only a subset of the fields for a specific feature, sometimes in a opinionated way (e.g [RateLimit][])
* These APIs do not expose all the features capabilities that Envoy has either because these features are desired but the API
is not defined yet or the project cannot support such an extensive list of features.
To alleviate this problem, and provide an interim solution for a small section of advanced users who are well versed in
Envoy xDS API and its capabilities, this API is being introduced.

## Goals

## Non Goals

## Example

## State of the World

## Design Decisions
* This API will only support a single `targetRef` and can bind to only a `Gateway` resource. This simplifies reasoning of how
patches will work.
* This API will always be an experimental API and cannot be graduated into a stable API because Envoy Gateway cannot garuntee
  * that the naming scheme for the generated resources names wont change across releases
  * that the underlying Envoy Proxy API wont change across releases

## Alternatives

* Users can customize the Envoy [Bootstrap configuration using EnvoyProxy API][] and provide static xDS configuration.
* Users can extend funtionality by [Extending the Control Plane][] and adding gRPC hooks to modify the generated xDS configuration.



[PolicyAttachment]: https://gateway-api.sigs.k8s.io/references/policy-attachment/
[Gateway API]: https://gateway-api.sigs.k8s.io/
[Extension APIs]: https://gateway.envoyproxy.io/latest/api/extension_types.html
[RateLimit]: https://gateway.envoyproxy.io/latest/user/rate-limit.html
[Extending the Control Plane]: https://gateway.envoyproxy.io/latest/design/extending-envoy-gateway.html
[Bootstrap configuration using EnvoyProxy API]: https://gateway.envoyproxy.io/latest/user/customize-envoyproxy.html#customize-envoyproxy-bootstrap-config
