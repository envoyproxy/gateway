# Envoy Gateway Authentication Policy API

## Overview

This authentication policy declares the authentication mechanisms, to be enforced on connection and request going though Envoy Gateway. This includes the credential (X.509, JWT, etc), parameters (cipher suites, key algorithms)
The policy is similar to [OpenAPI 3.1 security objects](https://github.com/OAI/OpenAPI-Specification/blob/main/versions/3.1.0.md#securitySchemeObject) without the API key part, and should be easily translatable from it with some additions.

## Authentication mechanisms
This policy should support the following authentication mechanisms:
- JWT Bearer Token
- mutualTLS (client certificate)
- OAuth2
- OIDC
- External authentication

In the first phase, Envoy Gateway will implement JWT Bearer Token authentication.

In general those policy translates into Envoy HTTP filters at HTTP connection manager level, and route specific settings will be applied for each route. These APIs are expressed in a Policy CRD and attached to Gateway API resources with [Policy Attachement](https://gateway-api.sigs.k8s.io/references/policy-attachment/).

## JWT Bearer Token

A JWT Bearer Token authentication policy will look like the following:

```
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: Authentication
metadata:
  name: productpage
spec:
  type: jwt
  jwt:
    iss: https://www.okta.com
    aud: bookinfo.com
    jwksUri: https://bookinfo.com/jwt/public-key/jwks.json
  targetRef:
    kind: HTTPRoute
    name: httpbin
```

JWT Bearer token will be translate to Envoy's JWT authentication filter. The JWKS URI need to be translated to a separate cluster for JWKS fetch and refresh.
