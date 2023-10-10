---
title: "Extension APIs"
---

## Packages
- [gateway.envoyproxy.io/v1alpha1](#gatewayenvoyproxyiov1alpha1)


## gateway.envoyproxy.io/v1alpha1

Package v1alpha1 contains API schema definitions for the gateway.envoyproxy.io API group.


### Resource Types
- [AuthenticationFilter](#authenticationfilter)
- [EnvoyPatchPolicy](#envoypatchpolicy)
- [EnvoyPatchPolicyList](#envoypatchpolicylist)
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



## ClaimToHeader



ClaimToHeader defines a configuration to convert JWT claims into HTTP headers

_Appears in:_
- [JwtAuthenticationFilterProvider](#jwtauthenticationfilterprovider)

| Field | Description |
| --- | --- |
| `header` _string_ | Header defines the name of the HTTP request header that the JWT Claim will be saved into. |
| `claim` _string_ | Claim is the JWT Claim that should be saved into the header : it can be a nested claim of type (eg. "claim.nested.key", "sub"). The nested claim name must use dot "." to separate the JSON name path. |


## EnvoyJSONPatchConfig



EnvoyJSONPatchConfig defines the configuration for patching a Envoy xDS Resource using JSONPatch semantic

_Appears in:_
- [EnvoyPatchPolicySpec](#envoypatchpolicyspec)

| Field | Description |
| --- | --- |
| `type` _[EnvoyResourceType](#envoyresourcetype)_ | Type is the typed URL of the Envoy xDS Resource |
| `name` _string_ | Name is the name of the resource |
| `operation` _[JSONPatchOperation](#jsonpatchoperation)_ | Patch defines the JSON Patch Operation |


## EnvoyPatchPolicy



EnvoyPatchPolicy allows the user to modify the generated Envoy xDS resources by Envoy Gateway using this patch API

_Appears in:_
- [EnvoyPatchPolicyList](#envoypatchpolicylist)

| Field | Description |
| --- | --- |
| `apiVersion` _string_ | `gateway.envoyproxy.io/v1alpha1`
| `kind` _string_ | `EnvoyPatchPolicy`
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |
| `spec` _[EnvoyPatchPolicySpec](#envoypatchpolicyspec)_ | Spec defines the desired state of EnvoyPatchPolicy. |


## EnvoyPatchPolicyList



EnvoyPatchPolicyList contains a list of EnvoyPatchPolicy resources.



| Field | Description |
| --- | --- |
| `apiVersion` _string_ | `gateway.envoyproxy.io/v1alpha1`
| `kind` _string_ | `EnvoyPatchPolicyList`
| `metadata` _[ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#listmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |
| `items` _[EnvoyPatchPolicy](#envoypatchpolicy) array_ |  |


## EnvoyPatchPolicySpec



EnvoyPatchPolicySpec defines the desired state of EnvoyPatchPolicy.

_Appears in:_
- [EnvoyPatchPolicy](#envoypatchpolicy)

| Field | Description |
| --- | --- |
| `type` _[EnvoyPatchType](#envoypatchtype)_ | Type decides the type of patch. Valid EnvoyPatchType values are "JSONPatch". |
| `jsonPatches` _[EnvoyJSONPatchConfig](#envoyjsonpatchconfig) array_ | JSONPatch defines the JSONPatch configuration. |
| `targetRef` _[PolicyTargetReference](#policytargetreference)_ | TargetRef is the name of the Gateway API resource this policy is being attached to. Currently only attaching to Gateway is supported This Policy and the TargetRef MUST be in the same namespace for this Policy to have effect and be applied to the Gateway TargetRef |
| `priority` _integer_ | Priority of the EnvoyPatchPolicy. If multiple EnvoyPatchPolicies are applied to the same TargetRef, they will be applied in the ascending order of the priority i.e. int32.min has the highest priority and int32.max has the lowest priority. Defaults to 0. |




## EnvoyPatchType

_Underlying type:_ `string`

EnvoyPatchType specifies the types of Envoy patching mechanisms.

_Appears in:_
- [EnvoyPatchPolicySpec](#envoypatchpolicyspec)



## EnvoyResourceType

_Underlying type:_ `string`

EnvoyResourceType specifies the type URL of the Envoy resource.

_Appears in:_
- [EnvoyJSONPatchConfig](#envoyjsonpatchconfig)



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



## JSONPatchOperation



JSONPatchOperation defines the JSON Patch Operation as defined in https://datatracker.ietf.org/doc/html/rfc6902

_Appears in:_
- [EnvoyJSONPatchConfig](#envoyjsonpatchconfig)

| Field | Description |
| --- | --- |
| `op` _[JSONPatchOperationType](#jsonpatchoperationtype)_ | Op is the type of operation to perform |
| `path` _string_ | Path is the location of the target document/field where the operation will be performed Refer to https://datatracker.ietf.org/doc/html/rfc6901 for more details. |
| `value` _[JSON](#json)_ | Value is the new value of the path location. |


## JSONPatchOperationType

_Underlying type:_ `string`

JSONPatchOperationType specifies the JSON Patch operations that can be performed.

_Appears in:_
- [JSONPatchOperation](#jsonpatchoperation)



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
| `claimToHeaders` _[ClaimToHeader](#claimtoheader) array_ | ClaimToHeaders is a list of JWT claims that must be extracted into HTTP request headers For examples, following config: The claim must be of type; string, int, double, bool. Array type claims are not supported |


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
| `sourceIP` _string_ | Deprecated: Use SourceCIDR instead. |
| `sourceCIDR` _[SourceMatch](#sourcematch)_ | SourceCIDR is the client IP Address range to match on. |


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


## SourceMatch





_Appears in:_
- [RateLimitSelectCondition](#ratelimitselectcondition)

| Field | Description |
| --- | --- |
| `type` _[SourceMatchType](#sourcematchtype)_ |  |
| `value` _string_ | Value is the IP CIDR that represents the range of Source IP Addresses of the client. These could also be the intermediate addresses through which the request has flown through and is part of the  `X-Forwarded-For` header. For example, `192.168.0.1/32`, `192.168.0.0/24`, `001:db8::/64`. |


## SourceMatchType

_Underlying type:_ `string`



_Appears in:_
- [SourceMatch](#sourcematch)



