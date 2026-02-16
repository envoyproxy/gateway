---
title: "Rate Limit Design"
---

## Overview

Rate limit is a feature that allows the user to limit the number of incoming requests
to a predefined value based on attributes within the traffic flow.

Here are some reasons why a user may want to implement Rate limits

* To prevent malicious activity such as DDoS attacks.
* To prevent applications and its resources (such as a database) from getting overloaded.
* To create API limits based on user entitlements.

## Scope Types

The rate limit type here describes the scope of rate limits.

* Global - In this case, the rate limit is common across all the instances of Envoy proxies
where its applied i.e. if the data plane has 2 replicas of Envoy running, and the rate limit is
10 requests/second, this limit is common and will be hit if 5 requests pass through the first replica
and 5 requests pass through the second replica within the same second.

* Local - In this case, the rate limits are specific to each instance/replica of Envoy running.
Note - This is not part of the initial design and will be added as a future enhancement. 

## Match Types 

### Rate limit a specific traffic flow 

* Here is an example of a ratelimit implemented by the application developer to limit a specific user
by matching on a custom `x-user-id` header with a value set to `one`

```yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: BackendTrafficPolicy
metadata:
  name: ratelimit-specific-user
spec:
  targetRef:
    group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: example
  rateLimit:
    type: Global
    global:
      rules:
      - clientSelectors:
        - headers:
          - name: x-user-id
            value: one
        limit:
          requests: 10
          unit: Hour
---
apiVersion: gateway.networking.k8s.io/v1
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
        kind: RateLimitFilter
        name: ratelimit-specific-user
    backendRefs:
    - name: backend
      port: 3000
```

### Rate limit all traffic flows

* Here is an example of a rate limit implemented by the application developer that limits the total requests made
to a specific route to safeguard health of internal application components. In this case, no specific `headers` match
is specified, and the rate limit is applied to all traffic flows accepted by this `HTTPRoute`.

```yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: BackendTrafficPolicy
metadata:
  name: ratelimit-all-requests
spec:
  targetRef:
    group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: example
  rateLimit:
    type: Global
    global:
      rules:
      - limit:
          requests: 1000
          unit: Second
---
apiVersion: gateway.networking.k8s.io/v1
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
        kind: RateLimitFilter
        name: ratelimit-all-requests
    backendRefs:
    - name: backend
      port: 3000
```

### Rate limit per distinct value

* Here is an example of a rate limit implemented by the application developer to limit any unique user
by matching on a custom `x-user-id` header. Here, user A (recognised from the traffic flow using the header
`x-user-id` and value `a`) will be rate limited at 10 requests/hour and so will user B 
(recognised from the traffic flow using the header `x-user-id` and value `b`).

```yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: BackendTrafficPolicy
metadata:
  name: ratelimit-per-user
spec:
  targetRef:
    group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: example
  rateLimit:
    type: Global
    global:
      rules:
      - clientSelectors:
        - headers:
          - type: Distinct
            name: x-user-id
        limit:
          requests: 10
          unit: Hour
---
apiVersion: gateway.networking.k8s.io/v1
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
        kind: RateLimitFilter
        name: ratelimit-per-user 
    backendRefs:
    - name: backend
      port: 3000
```

### Rate limit per source IP

* Here is an example of a rate limit implemented by the application developer that limits the total requests made
to a specific route by matching on source IP. In this case, requests from `x.x.x.x` will be rate limited at 10 requests/hour.

```yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: BackendTrafficPolicy
metadata:
  name: ratelimit-per-ip
spec:
  targetRef:
    group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: example
  rateLimit:
    type: Global
    global:
      rules:
      - clientSelectors:
        - sourceIP: x.x.x.x/32
        limit:
          requests: 10
          unit: Hour
---
apiVersion: gateway.networking.k8s.io/v1
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
        kind: RateLimitFilter
        name: ratelimit-per-user 
    backendRefs:
    - name: backend
      port: 3000
```

### Rate limit based on JWT claims

* Here is an example of rate limit implemented by the application developer that limits the total requests made
to a specific route by matching on the jwt claim. In this case, requests with jwt claim information of `{"name":"John Doe"}`
will be rate limited at 10 requests/hour.

```yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: SecurityPolicy
metadata:
  name: jwt-example
spec:
  targetRef:
    group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: example
  jwt:
    providers:
      - name: example
        remoteJWKS:
          uri: https://raw.githubusercontent.com/envoyproxy/gateway/main/examples/kubernetes/jwt/jwks.json
        claimToHeaders:
        - claim: name
          header: custom-request-header
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: BackendTrafficPolicy
metadata:
  name: ratelimit-specific-user
spec:
  targetRef:
    group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: example
  rateLimit:
    type: Global
    global:
      rules:
      - clientSelectors:
        - headers:
          - name: custom-request-header
            value: John Doe
        limit:
          requests: 10
          unit: Hour
---
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: example
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
        value: /foo
```


## Multiple RateLimitFilters, rules and clientSelectors
* Users can create multiple `RateLimitFilter`s and apply it to the same `HTTPRoute`. In such a case each
`RateLimitFilter` will be applied to the route and matched (and limited) in a mutually exclusive way, independent of each other.
* Rate limits are applied for each `RateLimitFilter` `rule` when ALL the conditions under `clientSelectors` hold true.

Here's an example highlighting this -

```yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: BackendTrafficPolicy
metadata:
  name: ratelimit-all-safeguard-app 
spec:
  targetRef:
    group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: example
  rateLimit:
    type: Global
    global:
      rules:
      - limit:
          requests: 100
          unit: Hour
---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: BackendTrafficPolicy
metadata:
  name: ratelimit-per-user
spec:
  targetRef:
    group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: example
  rateLimit:
    type: Global
    global:
      rules:
      - clientSelectors:
        - headers:
          - type: Distinct
            name: x-user-id
        limit:
          requests: 100
          unit: Hour
---
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
        kind: RateLimitFilter
        name: ratelimit-per-user
    - type: ExtensionRef
      extensionRef:
        group: gateway.envoyproxy.io
        kind: RateLimitFilter
        name: ratelimit-all-safeguard-app
    backendRefs:
    - name: backend
      port: 3000
```

* The user has created two `RateLimitFilter`s  and has attached it to a `HTTPRoute` - one(`ratelimit-all-safeguard-app`) to
ensure that the backend does not get overwhelmed with requests, any excess requests are rate limited irrespective of
the attributes within the traffic flow, and another(`ratelimit-per-user`) to rate limit each distinct user client
who can be differentiated using the `x-user-id` header, to ensure that each client does not make excessive requests to the backend.
* If user `baz` (identified with the header and value of `x-user-id: baz`) sends 90 requests within the first second, and
user `bar` sends 11 more requests during that same interval of 1 second, and user `bar` sends the 101th request within that second,
the rule defined in `ratelimit-all-safeguard-app` gets activated and Envoy Gateway will ratelimit the request sent by `bar` (and any other
request sent within that 1 second). After 1 second, the rate limit counter associated with the `ratelimit-all-safeguard-app` rule
is reset and again evaluated.
* If user `bar` also ends up sending 90 more requests within the hour, summing up `bar`'s total request count to 101, the rate limit rule
defined within `ratelimit-per-user` will get activated, and `bar`'s requests will be rate limited again until the hour interval ends.
* Within the same above hour, if `baz` sends 991 more requests, summing up `baz`'s total request count to 1001, the rate limit rule defined
within `ratelimit-per-user` will get activated for `baz`, and `baz`'s requests will also be rate limited until the hour interval ends. 

## Design Decisions

* The initial design uses an Extension filter to apply the Rate Limit functionality on a specific [HTTPRoute][].
This was preferred over the [PolicyAttachment][] extension mechanism, because it is unclear whether Rate Limit
will be required to be enforced or overridden by the platform administrator or not.
* The RateLimitFilter can only be applied as a filter to a [HTTPRouteRule][], applying it across all backends within a [HTTPRoute][]
and cannot be applied a filter within a [HTTPBackendRef][] for a specific backend.
* The [HTTPRoute][] API has a [matches][] field within each [rule][] to select a specific traffic flow to be routed to
the destination backend. The RateLimitFilter API that can be attached to an HTTPRoute via an [extensionRef][] filter,
also has a `clientSelectors` field within each `rule` to select attributes within the traffic flow to rate limit specific clients.
The two levels of selectors/matches allow for flexibility and aim to hold match information specific to its use, allowing the author/owner
of each configuration to be different. It also allows the `clientSelectors` field within the RateLimitFilter to be enhanced with other matchable
attribute such as [IP subnet][] in the future that are not relevant in the [HTTPRoute][] API.

## Implementation Details

### Global Rate limiting

* [Global rate limiting][] in Envoy Proxy can be achieved using the following -
  * [Actions][] can be configured per [xDS Route][].
  * If the match criteria defined within these actions is met for a specific HTTP Request, a set of key value pairs called [descriptors][]
  defined within the above actions is sent to a remote [rate limit service][], whose configuration (such as the URL for the rate limit service) is defined 
  using a [rate limit filter][].
  * Based on information received by the rate limit service and its programmed configuration, a decision is computed, whether to rate limit
  the HTTP Request or not, and is sent back to Envoy, which enforces this decision on the data plane.
* Envoy Gateway will leverage this Envoy Proxy feature by -
  * Translating the user facing RateLimitFilter API into Rate limit [Actions][] as well as Rate limit service configuration to implement
  the desired API intent.
  * Envoy Gateway will use the existing [reference implementation][] of the rate limit service.
    * The Infrastructure administrator will need to enable the rate limit service using new settings that will be defined in the [EnvoyGateway][] config API.
  * The xDS IR will be enhanced to hold the user facing rate limit intent.
  * The xDS Translator will be enhanced to translate the rate limit field within the xDS IR into Rate limit [Actions][] as well as instantiate the [rate limit filter][].
  * A new runner called `rate-limit` will be added that subscribes to the xDS IR messages and translates it into a new Rate Limit Infra IR which contains
  the [rate limit service configuration][] as well as other information needed to deploy the rate limit service.
  * The infrastructure service will be enhanced to subscribe to the Rate Limit Infra IR and deploy a provider specific rate limit service runnable entity.
  * A Status field within the RateLimitFilter API will be added to reflect whether the specific configuration was programmed correctly in these multiple locations or not.

[PolicyAttachment]: https://gateway-api.sigs.k8s.io/references/policy-attachment/
[HTTPRoute]: https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1.HTTPRoute
[HTTPRouteRule]: https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1.HTTPRouteRule
[HTTPBackendRef]: https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1.HTTPBackendRef
[matches]: https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1.HTTPRouteMatch
[rule]: https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1.HTTPRouteMatch
[extensionRef]: https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1.HTTPRouteFilterType
[IP subnet]: https://en.wikipedia.org/wiki/Subnetwork
[Actions]: https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/route/v3/route_components.proto#envoy-v3-api-msg-config-route-v3-ratelimit-action
[descriptors]: https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/rate_limit_filter.html?highlight=descriptor#example-1
[Global rate limiting]: https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/other_features/global_rate_limiting
[xDS Route]: https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/route/v3/route_components.proto#envoy-v3-api-msg-config-route-v3-routeaction
[rate limit filter]: https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/filters/http/ratelimit/v3/rate_limit.proto#envoy-v3-api-msg-extensions-filters-http-ratelimit-v3-ratelimit 
[rate limit service]: https://www.envoyproxy.io/docs/envoy/latest/configuration/other_features/rate_limit#config-rate-limit-service
[reference implementation]: https://github.com/envoyproxy/ratelimit
[EnvoyGateway]: https://github.com/envoyproxy/gateway/blob/main/api/v1alpha1/envoygateway_types.go
[rate limit service configuration]: https://github.com/envoyproxy/ratelimit#configuration
