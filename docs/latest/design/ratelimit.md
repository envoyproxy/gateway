# Rate Limit Design

## Overview

Rate limit is a feature that allows the user to limit the number of incoming requests
to a predefined value based on attributes within the traffic flow.

Here are some reasons why a user may want to implements Rate limits

* To prevent malicious activity such as DDoS attacks.
* To prevent applications and its resources (such as a database) from getting overloaded.
* To create API limits based on user entitlements.

## Scope Types

The rate limit type here describe the scope of rate limits

* Global - In this case, the rate limit is common across all the instances of Envoy proxies
where its applied i.e. if the data plane has 2 replicas of Envoy running, and the rate limit is
10 requests/second, this limit is common and will be hit if 5 requests pass through the first replica
and 5 requests pass through the second replica within the same second.

* Local - In this case, the rate limits are specific to each instance/replica of Envoy running.
Note - This is not part of the initial design and will be added as a future enhancement. 

## Match Types 

### Rate limit a specifc traffic flow 

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

## Design Decisions

* The initial design uses an Extension filter to apply the Rate Limit functionality on a specific HTTPRoute.
This was preferred over the PolicyAttachment Extension mechanism, because it is unclear whether Rate Limit
will be required to be enforced or overridden by the platform administrator or not.
* The Rate limits are applied across all backends within a HTTPRoute, and are not applied per backend.
* The HTTPRoute API has a `matches` field within each `rule` to select a specific traffic flow to be routed to
the destination backend. The RateLimitFilter API that can be attached to an HTTPRoute via an `extensionRef` filter,
also has a `matches` field within each `rule` to select attributes within the traffic flow to rate limit specific clients.
The two levels of `matches` allow for flexibility and aim to hold match information specific to its use, allowing the author/owner
of each configuration to be different.
