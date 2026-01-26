---
title: HTTP Header and Method Based Authentication
description:
  Configure request authentication using HTTP headers and HTTP methods with
  SecurityPolicy.
---

## Overview

Envoy Gateway allows request authentication using HTTP headers and HTTP methods through `SecurityPolicy`.

This enables restricting access to routes based on specific header values, allowed HTTP methods, or a combination of both.

---

## Header-Based Authentication

Header-based authentication allows matching incoming requests based on the presence and value of specific HTTP headers.

This is commonly used for simple mechanisms such as API key validation using custom headers.

### Example
```yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: SecurityPolicy
metadata:
  name: header-auth
spec:
  targetRefs:
    - kind: HTTPRoute
      name: example-route
  authentication:
    rules:
      - headerMatches:
        - name: x-user
          exact: example-user
```

In this example, requests are allowed only if the request headers match the configured header match conditions.

---

## Method-Based Authentication

Method-based authentication restricts access based on the HTTP method of incoming requests.

This can be used to allow or deny specific operations such as `GET`, `POST`, or `DELETE`.

### Example
```yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: SecurityPolicy
metadata:
  name: method-auth
spec:
  targetRefs:
    - kind: HTTPRoute
      name: example-route
  authentication:
    rules:
      - methods:
          - GET
          - POST
```

In this configuration, only `GET` and `POST` requests are permitted. Any other HTTP methods (such as `PUT` or `DELETE`) will be blocked by the policy.

---

## Combined Header and Method Authentication

Header-based and method-based authentication can be combined within a single authentication rule for more granular control.

### Example
```yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: SecurityPolicy
metadata:
  name: combined-auth
spec:
  targetRefs:
    - kind: HTTPRoute
      name: example-route
authentication:
  rules:
    - headerMatches:
        - name: x-user
          exact: example-user
      methods:
        - GET
```

In this scenario, a request is only authorized if it matches the configured header conditions and uses the allowed HTTP method.

---

## Behavior Notes

- **Logical AND**: Authentication conditions within a rule use logical AND semantics. Requests must satisfy all configured header and method requirements.
- **Rule Evaluation**: Rules are evaluated in the order they are defined in the list.
- **Enforcement**: If a request does not meet the specified criteria, Envoy Gateway will reject the request before it reaches the backend service.