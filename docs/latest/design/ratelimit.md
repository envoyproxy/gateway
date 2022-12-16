# Rate Limit Design

## Overview

Rate limit is a feature that allows the user to limit the number of incoming requests
to a predefined value based on attributes within the traffic flow.

Here are some reasons why a user may want to implements Rate limits

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
kind: RateLimitFilter
metadata:
  name: ratelimit-specific-user
spec:
  type: Global
  rules:
  - matches:
    - header:
        name: x-user-id
        value: one
      limit:
        requests: 10
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
kind: RateLimitFilter
metadata:
  name: ratelimit-all-requests
spec:
  type: Global
  rules:
  - matches:
    - limit:
        requests: 1000
        unit: Second
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
kind: RateLimitFilter
metadata:
  name: ratelimit-per-user
spec:
  type: Global
  rules:
  - matches:
    - Type: Distinct
      header:
        name: x-user-id
      limit:
        requests: 10
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
      backendRefs:
        - name: backend
          port: 3000
```

## Multiple RateLimitFilters, rules and matches
* Users can create multiple `RateLimitFilter`s and apply it to the same `HTTPRoute`. In such a case each
`RateLimitFilter` will be applied to the route and matched (and limited) in a mutually exclusive way, independent of each other.
* Rate limits are applied for each `RateLimitFilter` `rule` when the conditions under `matches` hold true.
* A `match` holds true, when all conditions under the `match` hold true.

Here's an example highlighting this -

```
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: RateLimitFilter
metadata:
  name: ratelimit-all-safeguard-app 
spec:
  type: Global
  rules:
  - matches:
    - limit:
        requests: 100
        unit: Second
---

apiVersion: gateway.envoyproxy.io/v1alpha1
kind: RateLimitFilter
metadata:
  name: ratelimit-per-user
spec:
  type: Global
  rules:
  - matches:
    - Type: Distinct
      header:
        name: x-user-id
      limit:
        requests: 1000
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
who can be differentiated using the `x-user-id` header, to ensure that each client does not make exessive requests to the backend.
* If user `baz` (identified with the header and value of `x-user-id: baz`) sends 90 requests within the first second, and
user `bar` sends 11 more requests during that same interval of 1 second, and user `bar` sends the 101th request within that second,
the rule defined in `ratelimit-all-safeguard-app` gets activated and Envoy Gateway will ratelimit the request sent by `bar` (and any other
request sent within that 1 second). After 1 second, the rate limit counter associated with the `ratelimit-all-safeguard-app` rule
is reset and again evaluated.
* If user `bar` also ends up sending 90 more requests within the hour, summing up `bar`'s total request count to 101, the rate limit rule
defined within `ratelimit-per-user` will get activated, and `bar`'s requests will be rate limited again until the hour interval ends.
* Within the same above hour, if `baz` sends 11 more requests, summing up `baz`'s total request count to 101, the rate limit rule defined
within `ratelimit-per-user` will get activated for `baz`, and `baz`'s requests will also be rate limited until the hour interval ends. 

## Design Decisions

* The initial design uses an Extension filter to apply the Rate Limit functionality on a specific `HTTPRoute`.
This was preferred over the PolicyAttachment Extension mechanism, because it is unclear whether Rate Limit
will be required to be enforced or overridden by the platform administrator or not.
* The RateFilter can only be applied as a filter to a `HTTPRouteRule`, applying it across all backends within a `HTTPRoute`
and cannot be applied a filter within a `HTTPBackendRef` for a specific backend.
* The HTTPRoute API has a `matches` field within each `rule` to select a specific traffic flow to be routed to
the destination backend. The RateLimitFilter API that can be attached to an HTTPRoute via an `extensionRef` filter,
also has a `matches` field within each `rule` to select attributes within the traffic flow to rate limit specific clients.
The two levels of `matches` allow for flexibility and aim to hold match information specific to its use, allowing the author/owner
of each configuration to be different.
