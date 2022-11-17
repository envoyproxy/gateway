# Rate Limiting Design

## Overview

Rate limiting is a feature that allows the user to limit the number of incoming requests
to a predefined value based on attributes within the traffic flow.

Here are some reasons why a user may want to implements Rate limits

* To prevent malicious activity such as DDoS attacks.
* To prevent applications and its resources (such as a database) from getting overloaded.
* To create API limits based on user entitlements.

## API

* Here is an example of a rate limit implemented by the application developer that limits the total requests made
to a specific route to safeguard health of internal application components.

```yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: RateLimit
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
            kind: RateLimit
            name: ratelimit-all-requests
      backendRefs:
        - name: backend
          port: 3000
```

* Here is an example of a ratelimit implemented by the application developer to limit a specific user
by matching on a custom `x-user-id` header with a value set to `one`

```yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: RateLimit
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
            kind: RateLimit
            name: ratelimit-specific-user
      backendRefs:
        - name: backend
          port: 3000
```

* Here is an example of a rate limit implemented by the application developer to limit any unique user
by matching on a custom `x-user-id` header. Here, user A (recognised from the traffic flow using the header
`x-user-id` and value `a`) will be rate limited at 10 requests/hour and so will user B 
(recognised from the traffic flow using the header `x-user-id` and value `b`).

```yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: RateLimit
metadata:
  name: ratelimit-per-user
spec:
  type: Global
  rules:
  - matches:
    - Type: Any
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
            kind: RateLimit
            name: ratelimit-per-user 
      backendRefs:
        - name: backend
          port: 3000
```

* The initial design uses an Extension filter to apply the Rate Limiting functionality on a specific HTTPRoute.
This was preferred over the PolicyAttachment Extension mechanism, because it is unclear whether Rate Limiting
will be required to be enforced or overridden by the platform administrator or not.
