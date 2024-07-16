---
title: "Extension APIs"
---

## Packages
- [gateway.envoyproxy.io/v1alpha1](#gatewayenvoyproxyiov1alpha1)


## gateway.envoyproxy.io/v1alpha1

Package v1alpha1 contains API schema definitions for the gateway.envoyproxy.io API group.


### Resource Types
- [AuthenticationFilter](#authenticationfilter)
- [RateLimitFilter](#ratelimitfilter)



## AuthenticationFilter







| Field | Description |
| --- | --- |
| `apiVersion` _string_ | `gateway.envoyproxy.io/v1alpha1`
| `kind` _string_ | `AuthenticationFilter`
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |
| `spec` _[AuthenticationFilterSpec](#authenticationfilterspec)_ | Spec defines the desired state of the AuthenticationFilter type. |


## AuthenticationFilterSpec



AuthenticationFilterSpec defines the desired state of the AuthenticationFilter type.

_Appears in:_
- [AuthenticationFilter](#authenticationfilter)

| Field | Description |
| --- | --- |
| `type` _[AuthenticationFilterType](#authenticationfiltertype)_ | Type defines the type of authentication provider to use. Supported provider types are "JWT". |
| `jwtProviders` _[JwtAuthenticationFilterProvider](#jwtauthenticationfilterprovider) array_ | JWT defines the JSON Web Token (JWT) authentication provider type. When multiple jwtProviders are specified, the JWT is considered valid if any of the providers successfully validate the JWT. For additional details, see https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/jwt_authn_filter.html. |


## AuthenticationFilterType

_Underlying type:_ `string`

AuthenticationFilterType is a type of authentication provider.

_Appears in:_
- [AuthenticationFilterSpec](#authenticationfilterspec)



## GlobalRateLimit



GlobalRateLimit defines global rate limit configuration.

_Appears in:_
- [RateLimitFilterSpec](#ratelimitfilterspec)

| Field | Description |
| --- | --- |
| `rules` _[RateLimitRule](#ratelimitrule) array_ | Rules are a list of RateLimit selectors and limits. Each rule and its associated limit is applied in a mutually exclusive way i.e. if multiple rules get selected, each of their associated limits get applied, so a single traffic request might increase the rate limit counters for multiple rules if selected. |


## HeaderMatch



HeaderMatch defines the match attributes within the HTTP Headers of the request.

_Appears in:_
- [RateLimitSelectCondition](#ratelimitselectcondition)

| Field | Description |
| --- | --- |
| `type` _[HeaderMatchType](#headermatchtype)_ | Type specifies how to match against the value of the header. |
| `name` _string_ | Name of the HTTP header. |
| `value` _string_ | Value within the HTTP header. Due to the case-insensitivity of header names, "foo" and "Foo" are considered equivalent. Do not set this field when Type="Distinct", implying matching on any/all unique values within the header. |


## HeaderMatchType

_Underlying type:_ `string`

HeaderMatchType specifies the semantics of how HTTP header values should be compared. Valid HeaderMatchType values are "Exact", "RegularExpression", and "Distinct".

_Appears in:_
- [HeaderMatch](#headermatch)



## JwtAuthenticationFilterProvider



JwtAuthenticationFilterProvider defines the JSON Web Token (JWT) authentication provider type and how JWTs should be verified:

_Appears in:_
- [AuthenticationFilterSpec](#authenticationfilterspec)

| Field | Description |
| --- | --- |
| `name` _string_ | Name defines a unique name for the JWT provider. A name can have a variety of forms, including RFC1123 subdomains, RFC 1123 labels, or RFC 1035 labels. |
| `issuer` _string_ | Issuer is the principal that issued the JWT and takes the form of a URL or email address. For additional details, see https://tools.ietf.org/html/rfc7519#section-4.1.1 for URL format and https://rfc-editor.org/rfc/rfc5322.html for email format. If not provided, the JWT issuer is not checked. |
| `audiences` _string array_ | Audiences is a list of JWT audiences allowed access. For additional details, see https://tools.ietf.org/html/rfc7519#section-4.1.3. If not provided, JWT audiences are not checked. |
| `remoteJWKS` _[RemoteJWKS](#remotejwks)_ | RemoteJWKS defines how to fetch and cache JSON Web Key Sets (JWKS) from a remote HTTP/HTTPS endpoint. |


## RateLimitFilter



RateLimitFilter allows the user to limit the number of incoming requests to a predefined value based on attributes within the traffic flow.



| Field | Description |
| --- | --- |
| `apiVersion` _string_ | `gateway.envoyproxy.io/v1alpha1`
| `kind` _string_ | `RateLimitFilter`
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |
| `spec` _[RateLimitFilterSpec](#ratelimitfilterspec)_ | Spec defines the desired state of RateLimitFilter. |


## RateLimitFilterSpec



RateLimitFilterSpec defines the desired state of RateLimitFilter.

_Appears in:_
- [RateLimitFilter](#ratelimitfilter)

| Field | Description |
| --- | --- |
| `type` _[RateLimitType](#ratelimittype)_ | Type decides the scope for the RateLimits. Valid RateLimitType values are "Global". |
| `global` _[GlobalRateLimit](#globalratelimit)_ | Global defines global rate limit configuration. |


## RateLimitRule



RateLimitRule defines the semantics for matching attributes from the incoming requests, and setting limits for them.

_Appears in:_
- [GlobalRateLimit](#globalratelimit)

| Field | Description |
| --- | --- |
| `clientSelectors` _[RateLimitSelectCondition](#ratelimitselectcondition) array_ | ClientSelectors holds the list of select conditions to select specific clients using attributes from the traffic flow. All individual select conditions must hold True for this rule and its limit to be applied. If this field is empty, it is equivalent to True, and the limit is applied. |
| `limit` _[RateLimitValue](#ratelimitvalue)_ | Limit holds the rate limit values. This limit is applied for traffic flows when the selectors compute to True, causing the request to be counted towards the limit. The limit is enforced and the request is ratelimited, i.e. a response with 429 HTTP status code is sent back to the client when the selected requests have reached the limit. |


## RateLimitSelectCondition



RateLimitSelectCondition specifies the attributes within the traffic flow that can be used to select a subset of clients to be ratelimited. All the individual conditions must hold True for the overall condition to hold True.

_Appears in:_
- [RateLimitRule](#ratelimitrule)

| Field | Description |
| --- | --- |
| `headers` _[HeaderMatch](#headermatch) array_ | Headers is a list of request headers to match. Multiple header values are ANDed together, meaning, a request MUST match all the specified headers. |


## RateLimitType

_Underlying type:_ `string`

RateLimitType specifies the types of RateLimiting.

_Appears in:_
- [RateLimitFilterSpec](#ratelimitfilterspec)



## RateLimitUnit

_Underlying type:_ `string`

RateLimitUnit specifies the intervals for setting rate limits. Valid RateLimitUnit values are "Second", "Minute", "Hour", and "Day".

_Appears in:_
- [RateLimitValue](#ratelimitvalue)



## RateLimitValue



RateLimitValue defines the limits for rate limiting.

_Appears in:_
- [RateLimitRule](#ratelimitrule)

| Field | Description |
| --- | --- |
| `requests` _integer_ |  |
| `unit` _[RateLimitUnit](#ratelimitunit)_ |  |


## RemoteJWKS



RemoteJWKS defines how to fetch and cache JSON Web Key Sets (JWKS) from a remote HTTP/HTTPS endpoint.

_Appears in:_
- [JwtAuthenticationFilterProvider](#jwtauthenticationfilterprovider)

| Field | Description |
| --- | --- |
| `uri` _string_ | URI is the HTTPS URI to fetch the JWKS. Envoy's system trust bundle is used to validate the server certificate. |


