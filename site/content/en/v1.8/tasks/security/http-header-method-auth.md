---
title: HTTP Header and Method Based Authorization
description:
  Configure request authorization using HTTP headers and HTTP methods with
  SecurityPolicy.
---

## Overview

Envoy Gateway allows controlling access to requests based on HTTP headers and HTTP methods using `SecurityPolicy` **authorization rules**.

This enables restricting access to routes based on specific header values, allowed HTTP methods, or a combination of both.

> **Note:** Header and method based access control is implemented using `SecurityPolicy` authorization rules, not request authentication.

---

## Header-Based Authorization

Header-based authorization allows controlling access based on values present in HTTP request headers.

This can be used to allow requests only from specific users or identities represented via request headers.

### Example
```yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: SecurityPolicy
metadata:
  name: header-auth
spec:
  targetRef:
    group: gateway.networking.k8s.io
    kind: Gateway
    name: gateway-1
  authorization:
    defaultAction: Deny
    rules:
      - name: allow-specific-users
        action: Allow
        principal:
          headers:
            - name: x-user-id
              values:
                - example-user
```

In this example, requests are allowed only if the `x-user-id` request header matches one of the configured allowed values.

---

## Method-Based Authorization

Method-based authorization restricts access based on the HTTP method of incoming requests.

This can be used to allow or deny specific operations such as `GET`, `POST`, or `DELETE`.

### Example
```yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: SecurityPolicy
metadata:
  name: method-auth
spec:
  targetRef:
    group: gateway.networking.k8s.io
    kind: Gateway
    name: gateway-1
  authorization:
    defaultAction: Deny
    rules:
      - name: allow-read-methods
        action: Allow
        operation:
          methods:
            - GET
            - POST
```

In this configuration, only `GET` and `POST` requests are permitted. Any other HTTP methods (such as `PUT` or `DELETE`) will be denied by default.

---

## Combined Header and Method Authorization

Header-based and method-based authorization can be combined within a single authorization rule for more granular access control.

### Example
```yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: SecurityPolicy
metadata:
  name: combined-auth
spec:
  targetRef:
    group: gateway.networking.k8s.io
    kind: Gateway
    name: gateway-1
  authorization:
    defaultAction: Deny
    rules:
      - name: allow-user-get
        action: Allow
        operation:
          methods:
            - GET
        principal:
          headers:
            - name: x-user-id
              values:
                - example-user
```

In this scenario, a request is authorized only if it uses an allowed HTTP method and matches the configured header-based principal conditions.

---

## Behavior Notes

- **Authorization semantics:** Rules define authorization behavior, not request authentication.
- **Logical AND:** Conditions within a rule use logical AND semantics. Requests must satisfy all configured header and method requirements.
- **Rule Evaluation:** Rules are evaluated in the order they are defined.
- **Default Action:** Requests that do not match any rule are handled according to the configured `defaultAction`.