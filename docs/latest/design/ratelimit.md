# Ratelimiting Design

## Overview

Ratelimiting is a feature that allows the user to limit the number of incoming requests
to a predefined value based on attributes within the traffic flow.

Here are some reasons why a user may want to implements Ratelimits

* To prevent malicious activity such as DDoS attacks.
* To prevent applications and its resources (such as a database) from getting overloaded.
* To create API limits based on user entitlements.

## API

* Here is an example of a ratelimit implemented by the platform engineer that limits requests made
by every unique client remote address to 100 requests per second, to help mitigate DDoS attacks.
```
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: RateLimiting 
metadata:
  name: ratelimit-per-client-ip
spec:
  type: Global
  rules:
  - matches:
    - remoteAddress: {}
    limit:
      requests: 100
      unit: Minute
```

* Here is an example of a ratelimit implemented by the application developer that limits total requests made
to a specific route to safeguard health of internal application components.
```
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: RateLimiting
metadata:
  name: ratelimit-all-requests
spec:
  type: Global
  rules:
  - matches:
    - limit:
        requests: 1000
	unit: Second
```

* Here is an example of a ratelimit implemented by the application developer to limit a specific set of users
by matching on a custom `x-user-tier` header with a value set to `one`
```
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: RateLimiting
metadata:
  name: ratelimit-specific-requests
spec:
  type: Global
  rules:
  - matches:
    - header:
        name: x-user-tier
	value: one
      limit:
        requests: 10
	unit: Hour
```
