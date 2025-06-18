+++
title = "Gateway API Extensions"
weight = 1
description = "Envoy Gateway provides these extensions to support additional features not available in the Gateway API today"
+++


## Packages
- [gateway.envoyproxy.io/v1alpha1](#gatewayenvoyproxyiov1alpha1)


## gateway.envoyproxy.io/v1alpha1

Package v1alpha1 contains API schema definitions for the gateway.envoyproxy.io
API group.


### Resource Types
- [Backend](#backend)
- [BackendTrafficPolicy](#backendtrafficpolicy)
- [ClientTrafficPolicy](#clienttrafficpolicy)
- [EnvoyExtensionPolicy](#envoyextensionpolicy)
- [EnvoyGateway](#envoygateway)
- [EnvoyPatchPolicy](#envoypatchpolicy)
- [EnvoyProxy](#envoyproxy)
- [HTTPRouteFilter](#httproutefilter)
- [SecurityPolicy](#securitypolicy)



#### ALPNProtocol

_Underlying type:_ _string_

ALPNProtocol specifies the protocol to be negotiated using ALPN

_Appears in:_
- [BackendTLSConfig](#backendtlsconfig)
- [ClientTLSSettings](#clienttlssettings)
- [TLSSettings](#tlssettings)

| Value | Description |
| ----- | ----------- |
| `http/1.0` | HTTPProtocolVersion1_0 specifies that HTTP/1.0 should be negotiable with ALPN<br /> | 
| `http/1.1` | HTTPProtocolVersion1_1 specifies that HTTP/1.1 should be negotiable with ALPN<br /> | 
| `h2` | HTTPProtocolVersion2 specifies that HTTP/2 should be negotiable with ALPN<br /> | 


#### ALSEnvoyProxyAccessLog



ALSEnvoyProxyAccessLog defines the gRPC Access Log Service (ALS) sink.
The service must implement the Envoy gRPC Access Log Service streaming API:
https://www.envoyproxy.io/docs/envoy/latest/api-v3/service/accesslog/v3/als.proto
Access log format information is passed in the form of gRPC metadata when the
stream is established.

_Appears in:_
- [ProxyAccessLogSink](#proxyaccesslogsink)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `backendRef` | _[BackendObjectReference](https://gateway-api.sigs.k8s.io/references/spec/#gateway.networking.k8s.io/v1.BackendObjectReference)_ |  false  |  | BackendRef references a Kubernetes object that represents the<br />backend server to which the authorization request will be sent.<br />Deprecated: Use BackendRefs instead. |
| `backendRefs` | _[BackendRef](#backendref) array_ |  false  |  | BackendRefs references a Kubernetes object that represents the<br />backend server to which the authorization request will be sent. |
| `backendSettings` | _[ClusterSettings](#clustersettings)_ |  false  |  | BackendSettings holds configuration for managing the connection<br />to the backend. |
| `logName` | _string_ |  false  |  | LogName defines the friendly name of the access log to be returned in<br />StreamAccessLogsMessage.Identifier. This allows the access log server<br />to differentiate between different access logs coming from the same Envoy. |
| `type` | _[ALSEnvoyProxyAccessLogType](#alsenvoyproxyaccesslogtype)_ |  true  |  | Type defines the type of accesslog. Supported types are "HTTP" and "TCP". |
| `http` | _[ALSEnvoyProxyHTTPAccessLogConfig](#alsenvoyproxyhttpaccesslogconfig)_ |  false  |  | HTTP defines additional configuration specific to HTTP access logs. |


#### ALSEnvoyProxyAccessLogType

_Underlying type:_ _string_



_Appears in:_
- [ALSEnvoyProxyAccessLog](#alsenvoyproxyaccesslog)

| Value | Description |
| ----- | ----------- |
| `HTTP` | ALSEnvoyProxyAccessLogTypeHTTP defines the HTTP access log type and will populate StreamAccessLogsMessage.http_logs.<br /> | 
| `TCP` | ALSEnvoyProxyAccessLogTypeTCP defines the TCP access log type and will populate StreamAccessLogsMessage.tcp_logs.<br /> | 


#### ALSEnvoyProxyHTTPAccessLogConfig





_Appears in:_
- [ALSEnvoyProxyAccessLog](#alsenvoyproxyaccesslog)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `requestHeaders` | _string array_ |  false  |  | RequestHeaders defines request headers to include in log entries sent to the access log service. |
| `responseHeaders` | _string array_ |  false  |  | ResponseHeaders defines response headers to include in log entries sent to the access log service. |
| `responseTrailers` | _string array_ |  false  |  | ResponseTrailers defines response trailers to include in log entries sent to the access log service. |


#### APIKeyAuth



APIKeyAuth defines the configuration for the API Key Authentication.

_Appears in:_
- [SecurityPolicySpec](#securitypolicyspec)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `credentialRefs` | _[SecretObjectReference](https://gateway-api.sigs.k8s.io/references/spec/#gateway.networking.k8s.io/v1.SecretObjectReference) array_ |  true  |  | CredentialRefs is the Kubernetes secret which contains the API keys.<br />This is an Opaque secret.<br />Each API key is stored in the key representing the client id.<br />If the secrets have a key for a duplicated client, the first one will be used. |
| `extractFrom` | _[ExtractFrom](#extractfrom) array_ |  true  |  | ExtractFrom is where to fetch the key from the coming request.<br />The value from the first source that has a key will be used. |


#### ActiveHealthCheck



ActiveHealthCheck defines the active health check configuration.
EG supports various types of active health checking including HTTP, TCP.

_Appears in:_
- [HealthCheck](#healthcheck)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `timeout` | _[Duration](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#duration-v1-meta)_ |  false  | 1s | Timeout defines the time to wait for a health check response. |
| `interval` | _[Duration](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#duration-v1-meta)_ |  false  | 3s | Interval defines the time between active health checks. |
| `unhealthyThreshold` | _integer_ |  false  | 3 | UnhealthyThreshold defines the number of unhealthy health checks required before a backend host is marked unhealthy. |
| `healthyThreshold` | _integer_ |  false  | 1 | HealthyThreshold defines the number of healthy health checks required before a backend host is marked healthy. |
| `type` | _[ActiveHealthCheckerType](#activehealthcheckertype)_ |  true  |  | Type defines the type of health checker. |
| `http` | _[HTTPActiveHealthChecker](#httpactivehealthchecker)_ |  false  |  | HTTP defines the configuration of http health checker.<br />It's required while the health checker type is HTTP. |
| `tcp` | _[TCPActiveHealthChecker](#tcpactivehealthchecker)_ |  false  |  | TCP defines the configuration of tcp health checker.<br />It's required while the health checker type is TCP. |
| `grpc` | _[GRPCActiveHealthChecker](#grpcactivehealthchecker)_ |  false  |  | GRPC defines the configuration of the GRPC health checker.<br />It's optional, and can only be used if the specified type is GRPC. |


#### ActiveHealthCheckPayload



ActiveHealthCheckPayload defines the encoding of the payload bytes in the payload.

_Appears in:_
- [HTTPActiveHealthChecker](#httpactivehealthchecker)
- [TCPActiveHealthChecker](#tcpactivehealthchecker)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `type` | _[ActiveHealthCheckPayloadType](#activehealthcheckpayloadtype)_ |  true  |  | Type defines the type of the payload. |
| `text` | _string_ |  false  |  | Text payload in plain text. |
| `binary` | _integer array_ |  false  |  | Binary payload base64 encoded. |


#### ActiveHealthCheckPayloadType

_Underlying type:_ _string_

ActiveHealthCheckPayloadType is the type of the payload.

_Appears in:_
- [ActiveHealthCheckPayload](#activehealthcheckpayload)

| Value | Description |
| ----- | ----------- |
| `Text` | ActiveHealthCheckPayloadTypeText defines the Text type payload.<br /> | 
| `Binary` | ActiveHealthCheckPayloadTypeBinary defines the Binary type payload.<br /> | 


#### ActiveHealthCheckerType

_Underlying type:_ _string_

ActiveHealthCheckerType is the type of health checker.

_Appears in:_
- [ActiveHealthCheck](#activehealthcheck)

| Value | Description |
| ----- | ----------- |
| `HTTP` | ActiveHealthCheckerTypeHTTP defines the HTTP type of health checking.<br /> | 
| `TCP` | ActiveHealthCheckerTypeTCP defines the TCP type of health checking.<br /> | 
| `GRPC` | ActiveHealthCheckerTypeGRPC defines the GRPC type of health checking.<br /> | 


#### AppProtocolType

_Underlying type:_ _string_

AppProtocolType defines various backend applications protocols supported by Envoy Gateway

_Appears in:_
- [BackendSpec](#backendspec)

| Value | Description |
| ----- | ----------- |
| `gateway.envoyproxy.io/h2c` | AppProtocolTypeH2C defines the HTTP/2 application protocol.<br /> | 
| `gateway.envoyproxy.io/ws` | AppProtocolTypeWS defines the WebSocket over HTTP protocol.<br /> | 
| `gateway.envoyproxy.io/wss` | AppProtocolTypeWSS defines the WebSocket over HTTPS protocol.<br /> | 


#### Authorization



Authorization defines the authorization configuration.

Note: if neither `Rules` nor `DefaultAction` is specified, the default action is to deny all requests.

_Appears in:_
- [SecurityPolicySpec](#securitypolicyspec)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `rules` | _[AuthorizationRule](#authorizationrule) array_ |  false  |  | Rules defines a list of authorization rules.<br />These rules are evaluated in order, the first matching rule will be applied,<br />and the rest will be skipped.<br />For example, if there are two rules: the first rule allows the request<br />and the second rule denies it, when a request matches both rules, it will be allowed. |
| `defaultAction` | _[AuthorizationAction](#authorizationaction)_ |  false  |  | DefaultAction defines the default action to be taken if no rules match.<br />If not specified, the default action is Deny. |


#### AuthorizationAction

_Underlying type:_ _string_

AuthorizationAction defines the action to be taken if a rule matches.

_Appears in:_
- [Authorization](#authorization)
- [AuthorizationRule](#authorizationrule)

| Value | Description |
| ----- | ----------- |
| `Allow` | AuthorizationActionAllow is the action to allow the request.<br /> | 
| `Deny` | AuthorizationActionDeny is the action to deny the request.<br /> | 


#### AuthorizationHeaderMatch



AuthorizationHeaderMatch specifies how to match against the value of an HTTP header within a authorization rule.

_Appears in:_
- [Principal](#principal)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `name` | _string_ |  true  |  | Name of the HTTP header.<br />The header name is case-insensitive unless PreserveHeaderCase is set to true.<br />For example, "Foo" and "foo" are considered the same header. |
| `values` | _string array_ |  true  |  | Values are the values that the header must match.<br />If multiple values are specified, the rule will match if any of the values match. |


#### AuthorizationRule



AuthorizationRule defines a single authorization rule.

_Appears in:_
- [Authorization](#authorization)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `name` | _string_ |  false  |  | Name is a user-friendly name for the rule.<br />If not specified, Envoy Gateway will generate a unique name for the rule. |
| `action` | _[AuthorizationAction](#authorizationaction)_ |  true  |  | Action defines the action to be taken if the rule matches. |
| `operation` | _[Operation](#operation)_ |  false  |  | Operation specifies the operation of a request, such as HTTP methods.<br />If not specified, all operations are matched on. |
| `principal` | _[Principal](#principal)_ |  true  |  | Principal specifies the client identity of a request.<br />If there are multiple principal types, all principals must match for the rule to match.<br />For example, if there are two principals: one for client IP and one for JWT claim,<br />the rule will match only if both the client IP and the JWT claim match. |


#### BackOffPolicy





_Appears in:_
- [PerRetryPolicy](#perretrypolicy)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `baseInterval` | _[Duration](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#duration-v1-meta)_ |  true  |  | BaseInterval is the base interval between retries. |
| `maxInterval` | _[Duration](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#duration-v1-meta)_ |  false  |  | MaxInterval is the maximum interval between retries. This parameter is optional, but must be greater than or equal to the base_interval if set.<br />The default is 10 times the base_interval |


#### Backend



Backend allows the user to configure the endpoints of a backend and
the behavior of the connection from Envoy Proxy to the backend.



| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `apiVersion` | _string_ | |`gateway.envoyproxy.io/v1alpha1`
| `kind` | _string_ | |`Backend`
| `metadata` | _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#objectmeta-v1-meta)_ |  true  |  | Refer to Kubernetes API documentation for fields of `metadata`. |
| `spec` | _[BackendSpec](#backendspec)_ |  true  |  | Spec defines the desired state of Backend. |
| `status` | _[BackendStatus](#backendstatus)_ |  true  |  | Status defines the current status of Backend. |


#### BackendCluster



BackendCluster contains all the configuration required for configuring access
to a backend. This can include multiple endpoints, and settings that apply for
managing the connection to all these endpoints.

_Appears in:_
- [ALSEnvoyProxyAccessLog](#alsenvoyproxyaccesslog)
- [ExtProc](#extproc)
- [GRPCExtAuthService](#grpcextauthservice)
- [HTTPExtAuthService](#httpextauthservice)
- [OIDCProvider](#oidcprovider)
- [OpenTelemetryEnvoyProxyAccessLog](#opentelemetryenvoyproxyaccesslog)
- [ProxyOpenTelemetrySink](#proxyopentelemetrysink)
- [RemoteJWKS](#remotejwks)
- [TracingProvider](#tracingprovider)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `backendRef` | _[BackendObjectReference](https://gateway-api.sigs.k8s.io/references/spec/#gateway.networking.k8s.io/v1.BackendObjectReference)_ |  false  |  | BackendRef references a Kubernetes object that represents the<br />backend server to which the authorization request will be sent.<br />Deprecated: Use BackendRefs instead. |
| `backendRefs` | _[BackendRef](#backendref) array_ |  false  |  | BackendRefs references a Kubernetes object that represents the<br />backend server to which the authorization request will be sent. |
| `backendSettings` | _[ClusterSettings](#clustersettings)_ |  false  |  | BackendSettings holds configuration for managing the connection<br />to the backend. |






#### BackendConnection



BackendConnection allows users to configure connection-level settings of backend

_Appears in:_
- [BackendTrafficPolicySpec](#backendtrafficpolicyspec)
- [ClusterSettings](#clustersettings)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `bufferLimit` | _[Quantity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#quantity-resource-api)_ |  false  |  | BufferLimit Soft limit on size of the cluster’s connections read and write buffers.<br />BufferLimit applies to connection streaming (maybe non-streaming) channel between processes, it's in user space.<br />If unspecified, an implementation defined default is applied (32768 bytes).<br />For example, 20Mi, 1Gi, 256Ki etc.<br />Note: that when the suffix is not provided, the value is interpreted as bytes. |


#### BackendEndpoint



BackendEndpoint describes a backend endpoint, which can be either a fully-qualified domain name, IP address or unix domain socket
corresponding to Envoy's Address: https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/core/v3/address.proto#config-core-v3-address

_Appears in:_
- [BackendSpec](#backendspec)
- [ExtensionService](#extensionservice)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `fqdn` | _[FQDNEndpoint](#fqdnendpoint)_ |  false  |  | FQDN defines a FQDN endpoint |
| `ip` | _[IPEndpoint](#ipendpoint)_ |  false  |  | IP defines an IP endpoint. Supports both IPv4 and IPv6 addresses. |
| `unix` | _[UnixSocket](#unixsocket)_ |  false  |  | Unix defines the unix domain socket endpoint |
| `zone` | _string_ |  false  |  | Zone defines the service zone of the backend endpoint. |


#### BackendRef



BackendRef defines how an ObjectReference that is specific to BackendRef.

_Appears in:_
- [ALSEnvoyProxyAccessLog](#alsenvoyproxyaccesslog)
- [BackendCluster](#backendcluster)
- [ExtProc](#extproc)
- [GRPCExtAuthService](#grpcextauthservice)
- [HTTPExtAuthService](#httpextauthservice)
- [OIDCProvider](#oidcprovider)
- [OpenTelemetryEnvoyProxyAccessLog](#opentelemetryenvoyproxyaccesslog)
- [ProxyOpenTelemetrySink](#proxyopentelemetrysink)
- [RemoteJWKS](#remotejwks)
- [TracingProvider](#tracingprovider)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `group` | _[Group](#group)_ |  false  |  | Group is the group of the referent. For example, "gateway.networking.k8s.io".<br />When unspecified or empty string, core API group is inferred. |
| `kind` | _[Kind](#kind)_ |  false  | Service | Kind is the Kubernetes resource kind of the referent. For example<br />"Service".<br />Defaults to "Service" when not specified.<br />ExternalName services can refer to CNAME DNS records that may live<br />outside of the cluster and as such are difficult to reason about in<br />terms of conformance. They also may not be safe to forward to (see<br />CVE-2021-25740 for more information). Implementations SHOULD NOT<br />support ExternalName Services.<br />Support: Core (Services with a type other than ExternalName)<br />Support: Implementation-specific (Services with type ExternalName) |
| `name` | _[ObjectName](#objectname)_ |  true  |  | Name is the name of the referent. |
| `namespace` | _[Namespace](#namespace)_ |  false  |  | Namespace is the namespace of the backend. When unspecified, the local<br />namespace is inferred.<br />Note that when a namespace different than the local namespace is specified,<br />a ReferenceGrant object is required in the referent namespace to allow that<br />namespace's owner to accept the reference. See the ReferenceGrant<br />documentation for details.<br />Support: Core |
| `port` | _[PortNumber](#portnumber)_ |  false  |  | Port specifies the destination port number to use for this resource.<br />Port is required when the referent is a Kubernetes Service. In this<br />case, the port number is the service port number, not the target port.<br />For other resources, destination port might be derived from the referent<br />resource or this field. |
| `fallback` | _boolean_ |  false  |  | Fallback indicates whether the backend is designated as a fallback.<br />Multiple fallback backends can be configured.<br />It is highly recommended to configure active or passive health checks to ensure that failover can be detected<br />when the active backends become unhealthy and to automatically readjust once the primary backends are healthy again.<br />The overprovisioning factor is set to 1.4, meaning the fallback backends will only start receiving traffic when<br />the health of the active backends falls below 72%. |


#### BackendSpec



BackendSpec describes the desired state of BackendSpec.

_Appears in:_
- [Backend](#backend)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `type` | _[BackendType](#backendtype)_ |  false  | Endpoints | Type defines the type of the backend. Defaults to "Endpoints" |
| `endpoints` | _[BackendEndpoint](#backendendpoint) array_ |  true  |  | Endpoints defines the endpoints to be used when connecting to the backend. |
| `appProtocols` | _[AppProtocolType](#appprotocoltype) array_ |  false  |  | AppProtocols defines the application protocols to be supported when connecting to the backend. |
| `fallback` | _boolean_ |  false  |  | Fallback indicates whether the backend is designated as a fallback.<br />It is highly recommended to configure active or passive health checks to ensure that failover can be detected<br />when the active backends become unhealthy and to automatically readjust once the primary backends are healthy again.<br />The overprovisioning factor is set to 1.4, meaning the fallback backends will only start receiving traffic when<br />the health of the active backends falls below 72%. |
| `tls` | _[BackendTLSSettings](#backendtlssettings)_ |  false  |  | TLS defines the TLS settings for the backend.<br />TLS.CACertificateRefs and TLS.WellKnownCACertificates can only be specified for DynamicResolver backends.<br />TLS.InsecureSkipVerify can be specified for any Backends |


#### BackendStatus



BackendStatus defines the state of Backend

_Appears in:_
- [Backend](#backend)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `conditions` | _[Condition](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#condition-v1-meta) array_ |  false  |  | Conditions describe the current conditions of the Backend. |


#### BackendTLSConfig



BackendTLSConfig describes the BackendTLS configuration for Envoy Proxy.

_Appears in:_
- [EnvoyProxySpec](#envoyproxyspec)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `clientCertificateRef` | _[SecretObjectReference](https://gateway-api.sigs.k8s.io/references/spec/#gateway.networking.k8s.io/v1.SecretObjectReference)_ |  false  |  | ClientCertificateRef defines the reference to a Kubernetes Secret that contains<br />the client certificate and private key for Envoy to use when connecting to<br />backend services and external services, such as ExtAuth, ALS, OpenTelemetry, etc.<br />This secret should be located within the same namespace as the Envoy proxy resource that references it. |
| `minVersion` | _[TLSVersion](#tlsversion)_ |  false  |  | Min specifies the minimal TLS protocol version to allow.<br />The default is TLS 1.2 if this is not specified. |
| `maxVersion` | _[TLSVersion](#tlsversion)_ |  false  |  | Max specifies the maximal TLS protocol version to allow<br />The default is TLS 1.3 if this is not specified. |
| `ciphers` | _string array_ |  false  |  | Ciphers specifies the set of cipher suites supported when<br />negotiating TLS 1.0 - 1.2. This setting has no effect for TLS 1.3.<br />In non-FIPS Envoy Proxy builds the default cipher list is:<br />- [ECDHE-ECDSA-AES128-GCM-SHA256\|ECDHE-ECDSA-CHACHA20-POLY1305]<br />- [ECDHE-RSA-AES128-GCM-SHA256\|ECDHE-RSA-CHACHA20-POLY1305]<br />- ECDHE-ECDSA-AES256-GCM-SHA384<br />- ECDHE-RSA-AES256-GCM-SHA384<br />In builds using BoringSSL FIPS the default cipher list is:<br />- ECDHE-ECDSA-AES128-GCM-SHA256<br />- ECDHE-RSA-AES128-GCM-SHA256<br />- ECDHE-ECDSA-AES256-GCM-SHA384<br />- ECDHE-RSA-AES256-GCM-SHA384 |
| `ecdhCurves` | _string array_ |  false  |  | ECDHCurves specifies the set of supported ECDH curves.<br />In non-FIPS Envoy Proxy builds the default curves are:<br />- X25519<br />- P-256<br />In builds using BoringSSL FIPS the default curve is:<br />- P-256 |
| `signatureAlgorithms` | _string array_ |  false  |  | SignatureAlgorithms specifies which signature algorithms the listener should<br />support. |
| `alpnProtocols` | _[ALPNProtocol](#alpnprotocol) array_ |  false  |  | ALPNProtocols supplies the list of ALPN protocols that should be<br />exposed by the listener or used by the proxy to connect to the backend.<br />Defaults:<br />1. HTTPS Routes: h2 and http/1.1 are enabled in listener context.<br />2. Other Routes: ALPN is disabled.<br />3. Backends: proxy uses the appropriate ALPN options for the backend protocol.<br />When an empty list is provided, the ALPN TLS extension is disabled.<br />Supported values are:<br />- http/1.0<br />- http/1.1<br />- h2 |


#### BackendTLSSettings



BackendTLSSettings holds the TLS settings for the backend.

_Appears in:_
- [BackendSpec](#backendspec)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `caCertificateRefs` | _LocalObjectReference array_ |  false  |  | CACertificateRefs contains one or more references to Kubernetes objects that<br />contain TLS certificates of the Certificate Authorities that can be used<br />as a trust anchor to validate the certificates presented by the backend.<br />A single reference to a Kubernetes ConfigMap or a Kubernetes Secret,<br />with the CA certificate in a key named `ca.crt` is currently supported.<br />If CACertificateRefs is empty or unspecified, then WellKnownCACertificates must be<br />specified. Only one of CACertificateRefs or WellKnownCACertificates may be specified,<br />not both.<br />Only used for DynamicResolver backends. |
| `wellKnownCACertificates` | _[WellKnownCACertificatesType](#wellknowncacertificatestype)_ |  false  |  | WellKnownCACertificates specifies whether system CA certificates may be used in<br />the TLS handshake between the gateway and backend pod.<br />If WellKnownCACertificates is unspecified or empty (""), then CACertificateRefs<br />must be specified with at least one entry for a valid configuration. Only one of<br />CACertificateRefs or WellKnownCACertificates may be specified, not both.<br />Only used for DynamicResolver backends. |
| `insecureSkipVerify` | _boolean_ |  false  | false | InsecureSkipVerify indicates whether the upstream's certificate verification<br />should be skipped. Defaults to "false". |


#### BackendTelemetry





_Appears in:_
- [BackendTrafficPolicySpec](#backendtrafficpolicyspec)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `tracing` | _[Tracing](#tracing)_ |  false  |  | Tracing configures the tracing settings for the backend or HTTPRoute. |


#### BackendTrafficPolicy



BackendTrafficPolicy allows the user to configure the behavior of the connection
between the Envoy Proxy listener and the backend service.



| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `apiVersion` | _string_ | |`gateway.envoyproxy.io/v1alpha1`
| `kind` | _string_ | |`BackendTrafficPolicy`
| `metadata` | _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#objectmeta-v1-meta)_ |  true  |  | Refer to Kubernetes API documentation for fields of `metadata`. |
| `spec` | _[BackendTrafficPolicySpec](#backendtrafficpolicyspec)_ |  true  |  | spec defines the desired state of BackendTrafficPolicy. |
| `status` | _[PolicyStatus](https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1alpha2.PolicyStatus)_ |  true  |  | status defines the current status of BackendTrafficPolicy. |


#### BackendTrafficPolicySpec



BackendTrafficPolicySpec defines the desired state of BackendTrafficPolicy.

_Appears in:_
- [BackendTrafficPolicy](#backendtrafficpolicy)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `targetRef` | _[LocalPolicyTargetReferenceWithSectionName](https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1alpha2.LocalPolicyTargetReferenceWithSectionName)_ |  true  |  | TargetRef is the name of the resource this policy is being attached to.<br />This policy and the TargetRef MUST be in the same namespace for this<br />Policy to have effect<br />Deprecated: use targetRefs/targetSelectors instead |
| `targetRefs` | _[LocalPolicyTargetReferenceWithSectionName](https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1alpha2.LocalPolicyTargetReferenceWithSectionName) array_ |  true  |  | TargetRefs are the names of the Gateway resources this policy<br />is being attached to. |
| `targetSelectors` | _[TargetSelector](#targetselector) array_ |  true  |  | TargetSelectors allow targeting resources for this policy based on labels |
| `loadBalancer` | _[LoadBalancer](#loadbalancer)_ |  false  |  | LoadBalancer policy to apply when routing traffic from the gateway to<br />the backend endpoints. Defaults to `LeastRequest`. |
| `retry` | _[Retry](#retry)_ |  false  |  | Retry provides more advanced usage, allowing users to customize the number of retries, retry fallback strategy, and retry triggering conditions.<br />If not set, retry will be disabled. |
| `proxyProtocol` | _[ProxyProtocol](#proxyprotocol)_ |  false  |  | ProxyProtocol enables the Proxy Protocol when communicating with the backend. |
| `tcpKeepalive` | _[TCPKeepalive](#tcpkeepalive)_ |  false  |  | TcpKeepalive settings associated with the upstream client connection.<br />Disabled by default. |
| `healthCheck` | _[HealthCheck](#healthcheck)_ |  false  |  | HealthCheck allows gateway to perform active health checking on backends. |
| `circuitBreaker` | _[CircuitBreaker](#circuitbreaker)_ |  false  |  | Circuit Breaker settings for the upstream connections and requests.<br />If not set, circuit breakers will be enabled with the default thresholds |
| `timeout` | _[Timeout](#timeout)_ |  false  |  | Timeout settings for the backend connections. |
| `connection` | _[BackendConnection](#backendconnection)_ |  false  |  | Connection includes backend connection settings. |
| `dns` | _[DNS](#dns)_ |  false  |  | DNS includes dns resolution settings. |
| `http2` | _[HTTP2Settings](#http2settings)_ |  false  |  | HTTP2 provides HTTP/2 configuration for backend connections. |
| `mergeType` | _[MergeType](#mergetype)_ |  false  |  | MergeType determines how this configuration is merged with existing BackendTrafficPolicy<br />configurations targeting a parent resource. When set, this configuration will be merged<br />into a parent BackendTrafficPolicy (i.e. the one targeting a Gateway or Listener).<br />This field cannot be set when targeting a parent resource (Gateway).<br />If unset, no merging occurs, and only the most specific configuration takes effect. |
| `rateLimit` | _[RateLimitSpec](#ratelimitspec)_ |  false  |  | RateLimit allows the user to limit the number of incoming requests<br />to a predefined value based on attributes within the traffic flow. |
| `faultInjection` | _[FaultInjection](#faultinjection)_ |  false  |  | FaultInjection defines the fault injection policy to be applied. This configuration can be used to<br />inject delays and abort requests to mimic failure scenarios such as service failures and overloads |
| `useClientProtocol` | _boolean_ |  false  |  | UseClientProtocol configures Envoy to prefer sending requests to backends using<br />the same HTTP protocol that the incoming request used. Defaults to false, which means<br />that Envoy will use the protocol indicated by the attached BackendRef. |
| `compression` | _[Compression](#compression) array_ |  false  |  | The compression config for the http streams. |
| `responseOverride` | _[ResponseOverride](#responseoverride) array_ |  false  |  | ResponseOverride defines the configuration to override specific responses with a custom one.<br />If multiple configurations are specified, the first one to match wins. |
| `httpUpgrade` | _[ProtocolUpgradeConfig](#protocolupgradeconfig) array_ |  false  |  | HTTPUpgrade defines the configuration for HTTP protocol upgrades.<br />If not specified, the default upgrade configuration(websocket) will be used. |
| `telemetry` | _[BackendTelemetry](#backendtelemetry)_ |  false  |  | Telemetry configures the telemetry settings for the policy target (Gateway or xRoute).<br />This will override the telemetry settings in the EnvoyProxy resource. |


#### BackendType

_Underlying type:_ _string_

BackendType defines the type of the Backend.

_Appears in:_
- [BackendSpec](#backendspec)

| Value | Description |
| ----- | ----------- |
| `Endpoints` | BackendTypeEndpoints defines the type of the backend as Endpoints.<br /> | 
| `DynamicResolver` | BackendTypeDynamicResolver defines the type of the backend as DynamicResolver.<br />When a backend is of type DynamicResolver, the Envoy will resolve the upstream<br />ip address and port from the host header of the incoming request. If the ip address<br />is directly set in the host header, the Envoy will use the ip address and port as the<br />upstream address. If the hostname is set in the host header, the Envoy will resolve the<br />ip address and port from the hostname using the DNS resolver.<br /> | 


#### BasicAuth



BasicAuth defines the configuration for 	the HTTP Basic Authentication.

_Appears in:_
- [SecurityPolicySpec](#securitypolicyspec)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `users` | _[SecretObjectReference](https://gateway-api.sigs.k8s.io/references/spec/#gateway.networking.k8s.io/v1.SecretObjectReference)_ |  true  |  | The Kubernetes secret which contains the username-password pairs in<br />htpasswd format, used to verify user credentials in the "Authorization"<br />header.<br />This is an Opaque secret. The username-password pairs should be stored in<br />the key ".htpasswd". As the key name indicates, the value needs to be the<br />htpasswd format, for example: "user1:\{SHA\}hashed_user1_password".<br />Right now, only SHA hash algorithm is supported.<br />Reference to https://httpd.apache.org/docs/2.4/programs/htpasswd.html<br />for more details.<br />Note: The secret must be in the same namespace as the SecurityPolicy. |
| `forwardUsernameHeader` | _string_ |  false  |  | This field specifies the header name to forward a successfully authenticated user to<br />the backend. The header will be added to the request with the username as the value.<br />If it is not specified, the username will not be forwarded. |


#### BodyToExtAuth



BodyToExtAuth defines the Body to Ext Auth configuration

_Appears in:_
- [ExtAuth](#extauth)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `maxRequestBytes` | _integer_ |  true  |  | MaxRequestBytes is the maximum size of a message body that the filter will hold in memory.<br />Envoy will return HTTP 413 and will not initiate the authorization process when buffer<br />reaches the number set in this field.<br />Note that this setting will have precedence over failOpen mode. |


#### BootstrapType

_Underlying type:_ _string_

BootstrapType defines the types of bootstrap supported by Envoy Gateway.

_Appears in:_
- [ProxyBootstrap](#proxybootstrap)

| Value | Description |
| ----- | ----------- |
| `Merge` | Merge merges the provided bootstrap with the default one. The provided bootstrap can add or override a value<br />within a map, or add a new value to a list.<br />Please note that the provided bootstrap can't override a value within a list.<br /> | 
| `Replace` | Replace replaces the default bootstrap with the provided one.<br /> | 
| `JSONPatch` | JSONPatch applies the provided JSONPatches to the default bootstrap.<br /> | 


#### BrotliCompressor



BrotliCompressor defines the config for the Brotli compressor.
The default values can be found here:
https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/compression/brotli/compressor/v3/brotli.proto#extension-envoy-compression-brotli-compressor

_Appears in:_
- [Compression](#compression)



#### CIDR

_Underlying type:_ _string_

CIDR defines a CIDR Address range.
A CIDR can be an IPv4 address range such as "192.168.1.0/24" or an IPv6 address range such as "2001:0db8:11a3:09d7::/64".

_Appears in:_
- [Principal](#principal)
- [XForwardedForSettings](#xforwardedforsettings)



#### CORS



CORS defines the configuration for Cross-Origin Resource Sharing (CORS).

_Appears in:_
- [SecurityPolicySpec](#securitypolicyspec)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `allowOrigins` | _[Origin](#origin) array_ |  false  |  | AllowOrigins defines the origins that are allowed to make requests.<br />It specifies the allowed origins in the Access-Control-Allow-Origin CORS response header.<br />The value "*" allows any origin to make requests. |
| `allowMethods` | _string array_ |  false  |  | AllowMethods defines the methods that are allowed to make requests.<br />It specifies the allowed methods in the Access-Control-Allow-Methods CORS response header..<br />The value "*" allows any method to be used. |
| `allowHeaders` | _string array_ |  false  |  | AllowHeaders defines the headers that are allowed to be sent with requests.<br />It specifies the allowed headers in the Access-Control-Allow-Headers CORS response header..<br />The value "*" allows any header to be sent. |
| `exposeHeaders` | _string array_ |  false  |  | ExposeHeaders defines which response headers should be made accessible to<br />scripts running in the browser.<br />It specifies the headers in the Access-Control-Expose-Headers CORS response header..<br />The value "*" allows any header to be exposed. |
| `maxAge` | _[Duration](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#duration-v1-meta)_ |  false  |  | MaxAge defines how long the results of a preflight request can be cached.<br />It specifies the value in the Access-Control-Max-Age CORS response header.. |
| `allowCredentials` | _boolean_ |  false  |  | AllowCredentials indicates whether a request can include user credentials<br />like cookies, authentication headers, or TLS client certificates.<br />It specifies the value in the Access-Control-Allow-Credentials CORS response header. |


#### CircuitBreaker



CircuitBreaker defines the Circuit Breaker configuration.

_Appears in:_
- [BackendTrafficPolicySpec](#backendtrafficpolicyspec)
- [ClusterSettings](#clustersettings)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `maxConnections` | _integer_ |  false  | 1024 | The maximum number of connections that Envoy will establish to the referenced backend defined within a xRoute rule. |
| `maxPendingRequests` | _integer_ |  false  | 1024 | The maximum number of pending requests that Envoy will queue to the referenced backend defined within a xRoute rule. |
| `maxParallelRequests` | _integer_ |  false  | 1024 | The maximum number of parallel requests that Envoy will make to the referenced backend defined within a xRoute rule. |
| `maxParallelRetries` | _integer_ |  false  | 1024 | The maximum number of parallel retries that Envoy will make to the referenced backend defined within a xRoute rule. |
| `maxRequestsPerConnection` | _integer_ |  false  |  | The maximum number of requests that Envoy will make over a single connection to the referenced backend defined within a xRoute rule.<br />Default: unlimited. |
| `perEndpoint` | _[PerEndpointCircuitBreakers](#perendpointcircuitbreakers)_ |  false  |  | PerEndpoint defines Circuit Breakers that will apply per-endpoint for an upstream cluster |


#### ClaimToHeader



ClaimToHeader defines a configuration to convert JWT claims into HTTP headers

_Appears in:_
- [JWTProvider](#jwtprovider)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `header` | _string_ |  true  |  | Header defines the name of the HTTP request header that the JWT Claim will be saved into. |
| `claim` | _string_ |  true  |  | Claim is the JWT Claim that should be saved into the header : it can be a nested claim of type<br />(eg. "claim.nested.key", "sub"). The nested claim name must use dot "."<br />to separate the JSON name path. |


#### ClientConnection



ClientConnection allows users to configure connection-level settings of client

_Appears in:_
- [ClientTrafficPolicySpec](#clienttrafficpolicyspec)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `connectionLimit` | _[ConnectionLimit](#connectionlimit)_ |  false  |  | ConnectionLimit defines limits related to connections |
| `bufferLimit` | _[Quantity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#quantity-resource-api)_ |  false  |  | BufferLimit provides configuration for the maximum buffer size in bytes for each incoming connection.<br />BufferLimit applies to connection streaming (maybe non-streaming) channel between processes, it's in user space.<br />For example, 20Mi, 1Gi, 256Ki etc.<br />Note that when the suffix is not provided, the value is interpreted as bytes.<br />Default: 32768 bytes. |


#### ClientIPDetectionSettings



ClientIPDetectionSettings provides configuration for determining the original client IP address for requests.

_Appears in:_
- [ClientTrafficPolicySpec](#clienttrafficpolicyspec)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `xForwardedFor` | _[XForwardedForSettings](#xforwardedforsettings)_ |  false  |  | XForwardedForSettings provides configuration for using X-Forwarded-For headers for determining the client IP address. |
| `customHeader` | _[CustomHeaderExtensionSettings](#customheaderextensionsettings)_ |  false  |  | CustomHeader provides configuration for determining the client IP address for a request based on<br />a trusted custom HTTP header. This uses the custom_header original IP detection extension.<br />Refer to https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/http/original_ip_detection/custom_header/v3/custom_header.proto<br />for more details. |


#### ClientTLSSettings





_Appears in:_
- [ClientTrafficPolicySpec](#clienttrafficpolicyspec)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `clientValidation` | _[ClientValidationContext](#clientvalidationcontext)_ |  false  |  | ClientValidation specifies the configuration to validate the client<br />initiating the TLS connection to the Gateway listener. |
| `minVersion` | _[TLSVersion](#tlsversion)_ |  false  |  | Min specifies the minimal TLS protocol version to allow.<br />The default is TLS 1.2 if this is not specified. |
| `maxVersion` | _[TLSVersion](#tlsversion)_ |  false  |  | Max specifies the maximal TLS protocol version to allow<br />The default is TLS 1.3 if this is not specified. |
| `ciphers` | _string array_ |  false  |  | Ciphers specifies the set of cipher suites supported when<br />negotiating TLS 1.0 - 1.2. This setting has no effect for TLS 1.3.<br />In non-FIPS Envoy Proxy builds the default cipher list is:<br />- [ECDHE-ECDSA-AES128-GCM-SHA256\|ECDHE-ECDSA-CHACHA20-POLY1305]<br />- [ECDHE-RSA-AES128-GCM-SHA256\|ECDHE-RSA-CHACHA20-POLY1305]<br />- ECDHE-ECDSA-AES256-GCM-SHA384<br />- ECDHE-RSA-AES256-GCM-SHA384<br />In builds using BoringSSL FIPS the default cipher list is:<br />- ECDHE-ECDSA-AES128-GCM-SHA256<br />- ECDHE-RSA-AES128-GCM-SHA256<br />- ECDHE-ECDSA-AES256-GCM-SHA384<br />- ECDHE-RSA-AES256-GCM-SHA384 |
| `ecdhCurves` | _string array_ |  false  |  | ECDHCurves specifies the set of supported ECDH curves.<br />In non-FIPS Envoy Proxy builds the default curves are:<br />- X25519<br />- P-256<br />In builds using BoringSSL FIPS the default curve is:<br />- P-256 |
| `signatureAlgorithms` | _string array_ |  false  |  | SignatureAlgorithms specifies which signature algorithms the listener should<br />support. |
| `alpnProtocols` | _[ALPNProtocol](#alpnprotocol) array_ |  false  |  | ALPNProtocols supplies the list of ALPN protocols that should be<br />exposed by the listener or used by the proxy to connect to the backend.<br />Defaults:<br />1. HTTPS Routes: h2 and http/1.1 are enabled in listener context.<br />2. Other Routes: ALPN is disabled.<br />3. Backends: proxy uses the appropriate ALPN options for the backend protocol.<br />When an empty list is provided, the ALPN TLS extension is disabled.<br />Supported values are:<br />- http/1.0<br />- http/1.1<br />- h2 |
| `session` | _[Session](#session)_ |  false  |  | Session defines settings related to TLS session management. |


#### ClientTimeout





_Appears in:_
- [ClientTrafficPolicySpec](#clienttrafficpolicyspec)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `tcp` | _[TCPClientTimeout](#tcpclienttimeout)_ |  false  |  | Timeout settings for TCP. |
| `http` | _[HTTPClientTimeout](#httpclienttimeout)_ |  false  |  | Timeout settings for HTTP. |


#### ClientTrafficPolicy



ClientTrafficPolicy allows the user to configure the behavior of the connection
between the downstream client and Envoy Proxy listener.



| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `apiVersion` | _string_ | |`gateway.envoyproxy.io/v1alpha1`
| `kind` | _string_ | |`ClientTrafficPolicy`
| `metadata` | _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#objectmeta-v1-meta)_ |  true  |  | Refer to Kubernetes API documentation for fields of `metadata`. |
| `spec` | _[ClientTrafficPolicySpec](#clienttrafficpolicyspec)_ |  true  |  | Spec defines the desired state of ClientTrafficPolicy. |
| `status` | _[PolicyStatus](https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1alpha2.PolicyStatus)_ |  true  |  | Status defines the current status of ClientTrafficPolicy. |


#### ClientTrafficPolicySpec



ClientTrafficPolicySpec defines the desired state of ClientTrafficPolicy.

_Appears in:_
- [ClientTrafficPolicy](#clienttrafficpolicy)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `targetRef` | _[LocalPolicyTargetReferenceWithSectionName](https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1alpha2.LocalPolicyTargetReferenceWithSectionName)_ |  true  |  | TargetRef is the name of the resource this policy is being attached to.<br />This policy and the TargetRef MUST be in the same namespace for this<br />Policy to have effect<br />Deprecated: use targetRefs/targetSelectors instead |
| `targetRefs` | _[LocalPolicyTargetReferenceWithSectionName](https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1alpha2.LocalPolicyTargetReferenceWithSectionName) array_ |  true  |  | TargetRefs are the names of the Gateway resources this policy<br />is being attached to. |
| `targetSelectors` | _[TargetSelector](#targetselector) array_ |  true  |  | TargetSelectors allow targeting resources for this policy based on labels |
| `tcpKeepalive` | _[TCPKeepalive](#tcpkeepalive)_ |  false  |  | TcpKeepalive settings associated with the downstream client connection.<br />If defined, sets SO_KEEPALIVE on the listener socket to enable TCP Keepalives.<br />Disabled by default. |
| `enableProxyProtocol` | _boolean_ |  false  |  | EnableProxyProtocol interprets the ProxyProtocol header and adds the<br />Client Address into the X-Forwarded-For header.<br />Note Proxy Protocol must be present when this field is set, else the connection<br />is closed. |
| `clientIPDetection` | _[ClientIPDetectionSettings](#clientipdetectionsettings)_ |  false  |  | ClientIPDetectionSettings provides configuration for determining the original client IP address for requests. |
| `tls` | _[ClientTLSSettings](#clienttlssettings)_ |  false  |  | TLS settings configure TLS termination settings with the downstream client. |
| `path` | _[PathSettings](#pathsettings)_ |  false  |  | Path enables managing how the incoming path set by clients can be normalized. |
| `headers` | _[HeaderSettings](#headersettings)_ |  false  |  | HeaderSettings provides configuration for header management. |
| `timeout` | _[ClientTimeout](#clienttimeout)_ |  false  |  | Timeout settings for the client connections. |
| `connection` | _[ClientConnection](#clientconnection)_ |  false  |  | Connection includes client connection settings. |
| `http1` | _[HTTP1Settings](#http1settings)_ |  false  |  | HTTP1 provides HTTP/1 configuration on the listener. |
| `http2` | _[HTTP2Settings](#http2settings)_ |  false  |  | HTTP2 provides HTTP/2 configuration on the listener. |
| `http3` | _[HTTP3Settings](#http3settings)_ |  false  |  | HTTP3 provides HTTP/3 configuration on the listener. |
| `healthCheck` | _[HealthCheckSettings](#healthchecksettings)_ |  false  |  | HealthCheck provides configuration for determining whether the HTTP/HTTPS listener is healthy. |


#### ClientValidationContext



ClientValidationContext holds configuration that can be used to validate the client initiating the TLS connection
to the Gateway.
By default, no client specific configuration is validated.

_Appears in:_
- [ClientTLSSettings](#clienttlssettings)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `optional` | _boolean_ |  false  |  | Optional set to true accepts connections even when a client doesn't present a certificate.<br />Defaults to false, which rejects connections without a valid client certificate. |
| `caCertificateRefs` | _[SecretObjectReference](https://gateway-api.sigs.k8s.io/references/spec/#gateway.networking.k8s.io/v1.SecretObjectReference) array_ |  false  |  | CACertificateRefs contains one or more references to<br />Kubernetes objects that contain TLS certificates of<br />the Certificate Authorities that can be used<br />as a trust anchor to validate the certificates presented by the client.<br />A single reference to a Kubernetes ConfigMap or a Kubernetes Secret,<br />with the CA certificate in a key named `ca.crt` is currently supported.<br />References to a resource in different namespace are invalid UNLESS there<br />is a ReferenceGrant in the target namespace that allows the certificate<br />to be attached. |


#### ClusterSettings



ClusterSettings provides the various knobs that can be set to control how traffic to a given
backend will be configured.

_Appears in:_
- [ALSEnvoyProxyAccessLog](#alsenvoyproxyaccesslog)
- [BackendCluster](#backendcluster)
- [BackendTrafficPolicySpec](#backendtrafficpolicyspec)
- [ExtProc](#extproc)
- [GRPCExtAuthService](#grpcextauthservice)
- [HTTPExtAuthService](#httpextauthservice)
- [OIDCProvider](#oidcprovider)
- [OpenTelemetryEnvoyProxyAccessLog](#opentelemetryenvoyproxyaccesslog)
- [ProxyOpenTelemetrySink](#proxyopentelemetrysink)
- [RemoteJWKS](#remotejwks)
- [TracingProvider](#tracingprovider)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `loadBalancer` | _[LoadBalancer](#loadbalancer)_ |  false  |  | LoadBalancer policy to apply when routing traffic from the gateway to<br />the backend endpoints. Defaults to `LeastRequest`. |
| `retry` | _[Retry](#retry)_ |  false  |  | Retry provides more advanced usage, allowing users to customize the number of retries, retry fallback strategy, and retry triggering conditions.<br />If not set, retry will be disabled. |
| `proxyProtocol` | _[ProxyProtocol](#proxyprotocol)_ |  false  |  | ProxyProtocol enables the Proxy Protocol when communicating with the backend. |
| `tcpKeepalive` | _[TCPKeepalive](#tcpkeepalive)_ |  false  |  | TcpKeepalive settings associated with the upstream client connection.<br />Disabled by default. |
| `healthCheck` | _[HealthCheck](#healthcheck)_ |  false  |  | HealthCheck allows gateway to perform active health checking on backends. |
| `circuitBreaker` | _[CircuitBreaker](#circuitbreaker)_ |  false  |  | Circuit Breaker settings for the upstream connections and requests.<br />If not set, circuit breakers will be enabled with the default thresholds |
| `timeout` | _[Timeout](#timeout)_ |  false  |  | Timeout settings for the backend connections. |
| `connection` | _[BackendConnection](#backendconnection)_ |  false  |  | Connection includes backend connection settings. |
| `dns` | _[DNS](#dns)_ |  false  |  | DNS includes dns resolution settings. |
| `http2` | _[HTTP2Settings](#http2settings)_ |  false  |  | HTTP2 provides HTTP/2 configuration for backend connections. |


#### Compression



Compression defines the config of enabling compression.
This can help reduce the bandwidth at the expense of higher CPU.

_Appears in:_
- [BackendTrafficPolicySpec](#backendtrafficpolicyspec)
- [ProxyPrometheusProvider](#proxyprometheusprovider)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `type` | _[CompressorType](#compressortype)_ |  true  |  | CompressorType defines the compressor type to use for compression. |
| `brotli` | _[BrotliCompressor](#brotlicompressor)_ |  false  |  | The configuration for Brotli compressor. |
| `gzip` | _[GzipCompressor](#gzipcompressor)_ |  false  |  | The configuration for GZIP compressor. |


#### CompressorType

_Underlying type:_ _string_

CompressorType defines the types of compressor library supported by Envoy Gateway.

_Appears in:_
- [Compression](#compression)

| Value | Description |
| ----- | ----------- |
| `Gzip` |  | 
| `Brotli` |  | 


#### ConnectConfig





_Appears in:_
- [ProtocolUpgradeConfig](#protocolupgradeconfig)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `terminate` | _boolean_ |  false  |  | Terminate the CONNECT request, and forwards the payload as raw TCP data. |


#### ConnectionLimit





_Appears in:_
- [ClientConnection](#clientconnection)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `value` | _integer_ |  true  |  | Value of the maximum concurrent connections limit.<br />When the limit is reached, incoming connections will be closed after the CloseDelay duration. |
| `closeDelay` | _[Duration](https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1.Duration)_ |  false  |  | CloseDelay defines the delay to use before closing connections that are rejected<br />once the limit value is reached.<br />Default: none. |


#### ConsistentHash



ConsistentHash defines the configuration related to the consistent hash
load balancer policy.

_Appears in:_
- [LoadBalancer](#loadbalancer)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `type` | _[ConsistentHashType](#consistenthashtype)_ |  true  |  | ConsistentHashType defines the type of input to hash on. Valid Type values are<br />"SourceIP",<br />"Header",<br />"Cookie". |
| `header` | _[Header](#header)_ |  false  |  | Header configures the header hash policy when the consistent hash type is set to Header. |
| `cookie` | _[Cookie](#cookie)_ |  false  |  | Cookie configures the cookie hash policy when the consistent hash type is set to Cookie. |
| `tableSize` | _integer_ |  false  | 65537 | The table size for consistent hashing, must be prime number limited to 5000011. |


#### ConsistentHashType

_Underlying type:_ _string_

ConsistentHashType defines the type of input to hash on.

_Appears in:_
- [ConsistentHash](#consistenthash)

| Value | Description |
| ----- | ----------- |
| `SourceIP` | SourceIPConsistentHashType hashes based on the source IP address.<br /> | 
| `Header` | HeaderConsistentHashType hashes based on a request header.<br /> | 
| `Cookie` | CookieConsistentHashType hashes based on a cookie.<br /> | 


#### Cookie



Cookie defines the cookie hashing configuration for consistent hash based
load balancing.

_Appears in:_
- [ConsistentHash](#consistenthash)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `name` | _string_ |  true  |  | Name of the cookie to hash.<br />If this cookie does not exist in the request, Envoy will generate a cookie and set<br />the TTL on the response back to the client based on Layer 4<br />attributes of the backend endpoint, to ensure that these future requests<br />go to the same backend endpoint. Make sure to set the TTL field for this case. |
| `ttl` | _[Duration](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#duration-v1-meta)_ |  false  |  | TTL of the generated cookie if the cookie is not present. This value sets the<br />Max-Age attribute value. |
| `attributes` | _object (keys:string, values:string)_ |  false  |  | Additional Attributes to set for the generated cookie. |


#### CustomHeaderExtensionSettings



CustomHeaderExtensionSettings provides configuration for determining the client IP address for a request based on
a trusted custom HTTP header. This uses the the custom_header original IP detection extension.
Refer to https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/http/original_ip_detection/custom_header/v3/custom_header.proto
for more details.

_Appears in:_
- [ClientIPDetectionSettings](#clientipdetectionsettings)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `name` | _string_ |  true  |  | Name of the header containing the original downstream remote address, if present. |
| `failClosed` | _boolean_ |  false  |  | FailClosed is a switch used to control the flow of traffic when client IP detection<br />fails. If set to true, the listener will respond with 403 Forbidden when the client<br />IP address cannot be determined. |


#### CustomRedirect



CustomRedirect contains configuration for returning a custom redirect.

_Appears in:_
- [ResponseOverride](#responseoverride)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `scheme` | _string_ |  false  |  | Scheme is the scheme to be used in the value of the `Location` header in<br />the response. When empty, the scheme of the request is used. |
| `hostname` | _[PreciseHostname](#precisehostname)_ |  false  |  | Hostname is the hostname to be used in the value of the `Location`<br />header in the response.<br />When empty, the hostname in the `Host` header of the request is used. |
| `path` | _[HTTPPathModifier](#httppathmodifier)_ |  false  |  | Path defines parameters used to modify the path of the incoming request.<br />The modified path is then used to construct the `Location` header. When<br />empty, the request path is used as-is.<br />Only ReplaceFullPath path modifier is supported currently. |
| `port` | _[PortNumber](#portnumber)_ |  false  |  | Port is the port to be used in the value of the `Location`<br />header in the response.<br />If redirect scheme is not-empty, the well-known port associated with the redirect scheme will be used.<br />Specifically "http" to port 80 and "https" to port 443. If the redirect scheme does not have a<br />well-known port or redirect scheme is empty, the listener port of the Gateway will be used.<br />Port will not be added in the 'Location' header if scheme is HTTP and port is 80<br />or scheme is HTTPS and port is 443. |
| `statusCode` | _integer_ |  false  | 302 | StatusCode is the HTTP status code to be used in response. |


#### CustomResponse



CustomResponse defines the configuration for returning a custom response.

_Appears in:_
- [ResponseOverride](#responseoverride)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `contentType` | _string_ |  false  |  | Content Type of the response. This will be set in the Content-Type header. |
| `body` | _[CustomResponseBody](#customresponsebody)_ |  false  |  | Body of the Custom Response<br />Supports Envoy command operators for dynamic content (see https://www.envoyproxy.io/docs/envoy/latest/configuration/observability/access_log/usage#command-operators). |
| `statusCode` | _integer_ |  false  |  | Status Code of the Custom Response<br />If unset, does not override the status of response. |


#### CustomResponseBody



CustomResponseBody

_Appears in:_
- [CustomResponse](#customresponse)
- [HTTPDirectResponseFilter](#httpdirectresponsefilter)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `type` | _[ResponseValueType](#responsevaluetype)_ |  true  | Inline | Type is the type of method to use to read the body value.<br />Valid values are Inline and ValueRef, default is Inline. |
| `inline` | _string_ |  false  |  | Inline contains the value as an inline string. |
| `valueRef` | _[LocalObjectReference](#localobjectreference)_ |  false  |  | ValueRef contains the contents of the body<br />specified as a local object reference.<br />Only a reference to ConfigMap is supported.<br />The value of key `response.body` in the ConfigMap will be used as the response body.<br />If the key is not found, the first value in the ConfigMap will be used. |


#### CustomResponseMatch



CustomResponseMatch defines the configuration for matching a user response to return a custom one.

_Appears in:_
- [ResponseOverride](#responseoverride)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `statusCodes` | _[StatusCodeMatch](#statuscodematch) array_ |  true  |  | Status code to match on. The match evaluates to true if any of the matches are successful. |


#### CustomTag





_Appears in:_
- [ProxyTracing](#proxytracing)
- [Tracing](#tracing)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `type` | _[CustomTagType](#customtagtype)_ |  true  | Literal | Type defines the type of custom tag. |
| `literal` | _[LiteralCustomTag](#literalcustomtag)_ |  true  |  | Literal adds hard-coded value to each span.<br />It's required when the type is "Literal". |
| `environment` | _[EnvironmentCustomTag](#environmentcustomtag)_ |  true  |  | Environment adds value from environment variable to each span.<br />It's required when the type is "Environment". |
| `requestHeader` | _[RequestHeaderCustomTag](#requestheadercustomtag)_ |  true  |  | RequestHeader adds value from request header to each span.<br />It's required when the type is "RequestHeader". |


#### CustomTagType

_Underlying type:_ _string_



_Appears in:_
- [CustomTag](#customtag)

| Value | Description |
| ----- | ----------- |
| `Literal` | CustomTagTypeLiteral adds hard-coded value to each span.<br /> | 
| `Environment` | CustomTagTypeEnvironment adds value from environment variable to each span.<br /> | 
| `RequestHeader` | CustomTagTypeRequestHeader adds value from request header to each span.<br /> | 


#### DNS





_Appears in:_
- [BackendTrafficPolicySpec](#backendtrafficpolicyspec)
- [ClusterSettings](#clustersettings)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `dnsRefreshRate` | _[Duration](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#duration-v1-meta)_ |  true  |  | DNSRefreshRate specifies the rate at which DNS records should be refreshed.<br />Defaults to 30 seconds. |
| `respectDnsTtl` | _boolean_ |  true  |  | RespectDNSTTL indicates whether the DNS Time-To-Live (TTL) should be respected.<br />If the value is set to true, the DNS refresh rate will be set to the resource record’s TTL.<br />Defaults to true. |
| `lookupFamily` | _[DNSLookupFamily](#dnslookupfamily)_ |  false  |  | LookupFamily determines how Envoy would resolve DNS for Routes where the backend is specified as a fully qualified domain name (FQDN).<br />If set, this configuration overrides other defaults. |


#### DNSLookupFamily

_Underlying type:_ _string_

DNSLookupFamily defines the behavior of Envoy when resolving DNS for hostnames

_Appears in:_
- [DNS](#dns)

| Value | Description |
| ----- | ----------- |
| `IPv4` | IPv4DNSLookupFamily means the DNS resolver will first perform a lookup for addresses in the IPv4 family.<br /> | 
| `IPv6` | IPv6DNSLookupFamily means the DNS resolver will first perform a lookup for addresses in the IPv6 family.<br /> | 
| `IPv4Preferred` | IPv4PreferredDNSLookupFamily means the DNS resolver will first perform a lookup for addresses in the IPv4 family and fallback<br />to a lookup for addresses in the IPv6 family.<br /> | 
| `IPv6Preferred` | IPv6PreferredDNSLookupFamily means the DNS resolver will first perform a lookup for addresses in the IPv6 family and fallback<br />to a lookup for addresses in the IPv4 family.<br /> | 
| `IPv4AndIPv6` | IPv4AndIPv6DNSLookupFamily mean the DNS resolver will perform a lookup for both IPv4 and IPv6 families, and return all resolved<br />addresses. When this is used, Happy Eyeballs will be enabled for upstream connections.<br /> | 


#### EnvironmentCustomTag



EnvironmentCustomTag adds value from environment variable to each span.

_Appears in:_
- [CustomTag](#customtag)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `name` | _string_ |  true  |  | Name defines the name of the environment variable which to extract the value from. |
| `defaultValue` | _string_ |  false  |  | DefaultValue defines the default value to use if the environment variable is not set. |


#### EnvoyExtensionPolicy



EnvoyExtensionPolicy allows the user to configure various envoy extensibility options for the Gateway.



| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `apiVersion` | _string_ | |`gateway.envoyproxy.io/v1alpha1`
| `kind` | _string_ | |`EnvoyExtensionPolicy`
| `metadata` | _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#objectmeta-v1-meta)_ |  true  |  | Refer to Kubernetes API documentation for fields of `metadata`. |
| `spec` | _[EnvoyExtensionPolicySpec](#envoyextensionpolicyspec)_ |  true  |  | Spec defines the desired state of EnvoyExtensionPolicy. |
| `status` | _[PolicyStatus](https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1alpha2.PolicyStatus)_ |  true  |  | Status defines the current status of EnvoyExtensionPolicy. |


#### EnvoyExtensionPolicySpec



EnvoyExtensionPolicySpec defines the desired state of EnvoyExtensionPolicy.

_Appears in:_
- [EnvoyExtensionPolicy](#envoyextensionpolicy)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `targetRef` | _[LocalPolicyTargetReferenceWithSectionName](https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1alpha2.LocalPolicyTargetReferenceWithSectionName)_ |  true  |  | TargetRef is the name of the resource this policy is being attached to.<br />This policy and the TargetRef MUST be in the same namespace for this<br />Policy to have effect<br />Deprecated: use targetRefs/targetSelectors instead |
| `targetRefs` | _[LocalPolicyTargetReferenceWithSectionName](https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1alpha2.LocalPolicyTargetReferenceWithSectionName) array_ |  true  |  | TargetRefs are the names of the Gateway resources this policy<br />is being attached to. |
| `targetSelectors` | _[TargetSelector](#targetselector) array_ |  true  |  | TargetSelectors allow targeting resources for this policy based on labels |
| `wasm` | _[Wasm](#wasm) array_ |  false  |  | Wasm is a list of Wasm extensions to be loaded by the Gateway.<br />Order matters, as the extensions will be loaded in the order they are<br />defined in this list. |
| `extProc` | _[ExtProc](#extproc) array_ |  false  |  | ExtProc is an ordered list of external processing filters<br />that should be added to the envoy filter chain |
| `lua` | _[Lua](#lua) array_ |  false  |  | Lua is an ordered list of Lua filters<br />that should be added to the envoy filter chain |


#### EnvoyFilter

_Underlying type:_ _string_

EnvoyFilter defines the type of Envoy HTTP filter.

_Appears in:_
- [FilterPosition](#filterposition)

| Value | Description |
| ----- | ----------- |
| `envoy.filters.http.health_check` | EnvoyFilterHealthCheck defines the Envoy HTTP health check filter.<br /> | 
| `envoy.filters.http.fault` | EnvoyFilterFault defines the Envoy HTTP fault filter.<br /> | 
| `envoy.filters.http.cors` | EnvoyFilterCORS defines the Envoy HTTP CORS filter.<br /> | 
| `envoy.filters.http.ext_authz` | EnvoyFilterExtAuthz defines the Envoy HTTP external authorization filter.<br /> | 
| `envoy.filters.http.api_key_auth` | EnvoyFilterAPIKeyAuth defines the Envoy HTTP api key authentication filter.<br /> | 
| `envoy.filters.http.basic_auth` | EnvoyFilterBasicAuth defines the Envoy HTTP basic authentication filter.<br /> | 
| `envoy.filters.http.oauth2` | EnvoyFilterOAuth2 defines the Envoy HTTP OAuth2 filter.<br /> | 
| `envoy.filters.http.jwt_authn` | EnvoyFilterJWTAuthn defines the Envoy HTTP JWT authentication filter.<br /> | 
| `envoy.filters.http.stateful_session` | EnvoyFilterSessionPersistence defines the Envoy HTTP session persistence filter.<br /> | 
| `envoy.filters.http.ext_proc` | EnvoyFilterExtProc defines the Envoy HTTP external process filter.<br /> | 
| `envoy.filters.http.wasm` | EnvoyFilterWasm defines the Envoy HTTP WebAssembly filter.<br /> | 
| `envoy.filters.http.lua` | EnvoyFilterLua defines the Envoy HTTP Lua filter.<br /> | 
| `envoy.filters.http.rbac` | EnvoyFilterRBAC defines the Envoy RBAC filter.<br /> | 
| `envoy.filters.http.local_ratelimit` | EnvoyFilterLocalRateLimit defines the Envoy HTTP local rate limit filter.<br /> | 
| `envoy.filters.http.ratelimit` | EnvoyFilterRateLimit defines the Envoy HTTP rate limit filter.<br /> | 
| `envoy.filters.http.custom_response` | EnvoyFilterCustomResponse defines the Envoy HTTP custom response filter.<br /> | 
| `envoy.filters.http.credential_injector` | EnvoyFilterCredentialInjector defines the Envoy HTTP credential injector filter.<br /> | 
| `envoy.filters.http.compressor` | EnvoyFilterCompressor defines the Envoy HTTP compressor filter.<br /> | 
| `envoy.filters.http.router` | EnvoyFilterRouter defines the Envoy HTTP router filter.<br /> | 
| `envoy.filters.http.buffer` | EnvoyFilterBuffer defines the Envoy HTTP buffer filter<br /> | 


#### EnvoyGateway



EnvoyGateway is the schema for the envoygateways API.



| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `apiVersion` | _string_ | |`gateway.envoyproxy.io/v1alpha1`
| `kind` | _string_ | |`EnvoyGateway`
| `gateway` | _[Gateway](#gateway)_ |  false  |  | Gateway defines desired Gateway API specific configuration. If unset,<br />default configuration parameters will apply. |
| `provider` | _[EnvoyGatewayProvider](#envoygatewayprovider)_ |  false  |  | Provider defines the desired provider and provider-specific configuration.<br />If unspecified, the Kubernetes provider is used with default configuration<br />parameters. |
| `logging` | _[EnvoyGatewayLogging](#envoygatewaylogging)_ |  false  | \{ default:info \} | Logging defines logging parameters for Envoy Gateway. |
| `admin` | _[EnvoyGatewayAdmin](#envoygatewayadmin)_ |  false  |  | Admin defines the desired admin related abilities.<br />If unspecified, the Admin is used with default configuration<br />parameters. |
| `telemetry` | _[EnvoyGatewayTelemetry](#envoygatewaytelemetry)_ |  false  |  | Telemetry defines the desired control plane telemetry related abilities.<br />If unspecified, the telemetry is used with default configuration. |
| `rateLimit` | _[RateLimit](#ratelimit)_ |  false  |  | RateLimit defines the configuration associated with the Rate Limit service<br />deployed by Envoy Gateway required to implement the Global Rate limiting<br />functionality. The specific rate limit service used here is the reference<br />implementation in Envoy. For more details visit https://github.com/envoyproxy/ratelimit.<br />This configuration is unneeded for "Local" rate limiting. |
| `extensionManager` | _[ExtensionManager](#extensionmanager)_ |  false  |  | ExtensionManager defines an extension manager to register for the Envoy Gateway Control Plane. |
| `extensionApis` | _[ExtensionAPISettings](#extensionapisettings)_ |  false  |  | ExtensionAPIs defines the settings related to specific Gateway API Extensions<br />implemented by Envoy Gateway |


#### EnvoyGatewayAdmin



EnvoyGatewayAdmin defines the Envoy Gateway Admin configuration.

_Appears in:_
- [EnvoyGateway](#envoygateway)
- [EnvoyGatewaySpec](#envoygatewayspec)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `address` | _[EnvoyGatewayAdminAddress](#envoygatewayadminaddress)_ |  false  |  | Address defines the address of Envoy Gateway Admin Server. |
| `enableDumpConfig` | _boolean_ |  false  |  | EnableDumpConfig defines if enable dump config in Envoy Gateway logs. |
| `enablePprof` | _boolean_ |  false  |  | EnablePprof defines if enable pprof in Envoy Gateway Admin Server. |


#### EnvoyGatewayAdminAddress



EnvoyGatewayAdminAddress defines the Envoy Gateway Admin Address configuration.

_Appears in:_
- [EnvoyGatewayAdmin](#envoygatewayadmin)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `port` | _integer_ |  false  | 19000 | Port defines the port the admin server is exposed on. |
| `host` | _string_ |  false  | 127.0.0.1 | Host defines the admin server hostname. |


#### EnvoyGatewayCustomProvider



EnvoyGatewayCustomProvider defines configuration for the Custom provider.

_Appears in:_
- [EnvoyGatewayProvider](#envoygatewayprovider)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `resource` | _[EnvoyGatewayResourceProvider](#envoygatewayresourceprovider)_ |  true  |  | Resource defines the desired resource provider.<br />This provider is used to specify the provider to be used<br />to retrieve the resource configurations such as Gateway API<br />resources |
| `infrastructure` | _[EnvoyGatewayInfrastructureProvider](#envoygatewayinfrastructureprovider)_ |  false  |  | Infrastructure defines the desired infrastructure provider.<br />This provider is used to specify the provider to be used<br />to provide an environment to deploy the out resources like<br />the Envoy Proxy data plane.<br />Infrastructure is optional, if provider is not specified,<br />No infrastructure provider is available. |


#### EnvoyGatewayFileResourceProvider



EnvoyGatewayFileResourceProvider defines configuration for the File Resource provider.

_Appears in:_
- [EnvoyGatewayResourceProvider](#envoygatewayresourceprovider)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `paths` | _string array_ |  true  |  | Paths are the paths to a directory or file containing the resource configuration.<br />Recursive subdirectories are not currently supported. |


#### EnvoyGatewayHostInfrastructureProvider



EnvoyGatewayHostInfrastructureProvider defines configuration for the Host Infrastructure provider.

_Appears in:_
- [EnvoyGatewayInfrastructureProvider](#envoygatewayinfrastructureprovider)



#### EnvoyGatewayInfrastructureProvider



EnvoyGatewayInfrastructureProvider defines configuration for the Custom Infrastructure provider.

_Appears in:_
- [EnvoyGatewayCustomProvider](#envoygatewaycustomprovider)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `type` | _[InfrastructureProviderType](#infrastructureprovidertype)_ |  true  |  | Type is the type of infrastructure providers to use. Supported types are "Host". |
| `host` | _[EnvoyGatewayHostInfrastructureProvider](#envoygatewayhostinfrastructureprovider)_ |  false  |  | Host defines the configuration of the Host provider. Host provides runtime<br />deployment of the data plane as a child process on the host environment. |


#### EnvoyGatewayKubernetesProvider



EnvoyGatewayKubernetesProvider defines configuration for the Kubernetes provider.

_Appears in:_
- [EnvoyGatewayProvider](#envoygatewayprovider)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `rateLimitDeployment` | _[KubernetesDeploymentSpec](#kubernetesdeploymentspec)_ |  false  |  | RateLimitDeployment defines the desired state of the Envoy ratelimit deployment resource.<br />If unspecified, default settings for the managed Envoy ratelimit deployment resource<br />are applied. |
| `rateLimitHpa` | _[KubernetesHorizontalPodAutoscalerSpec](#kuberneteshorizontalpodautoscalerspec)_ |  false  |  | RateLimitHpa defines the Horizontal Pod Autoscaler settings for Envoy ratelimit Deployment.<br />If the HPA is set, Replicas field from RateLimitDeployment will be ignored. |
| `watch` | _[KubernetesWatchMode](#kuberneteswatchmode)_ |  false  |  | Watch holds configuration of which input resources should be watched and reconciled. |
| `leaderElection` | _[LeaderElection](#leaderelection)_ |  false  |  | LeaderElection specifies the configuration for leader election.<br />If it's not set up, leader election will be active by default, using Kubernetes' standard settings. |
| `shutdownManager` | _[ShutdownManager](#shutdownmanager)_ |  false  |  | ShutdownManager defines the configuration for the shutdown manager. |
| `client` | _[KubernetesClient](#kubernetesclient)_ |  true  |  | Client holds the configuration for the Kubernetes client. |
| `proxyTopologyInjector` | _[EnvoyGatewayTopologyInjector](#envoygatewaytopologyinjector)_ |  false  |  | TopologyInjector defines the configuration for topology injector MutatatingWebhookConfiguration |
| `cacheSyncPeriod` | _[Duration](https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1.Duration)_ |  false  |  | CacheSyncPeriod determines the minimum frequency at which watched resources are synced.<br />Note that a sync in the provider layer will not lead to a full reconciliation (including translation),<br />unless there are actual changes in the provider resources.<br />This option can be used to protect against missed events or issues in Envoy Gateway where resources<br />are not requeued when they should be, at the cost of increased resource consumption.<br />Learn more about the implications of this option: https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/cache#Options<br />Default: 10 hours |


#### EnvoyGatewayLogComponent

_Underlying type:_ _string_

EnvoyGatewayLogComponent defines a component that supports a configured logging level.

_Appears in:_
- [EnvoyGatewayLogging](#envoygatewaylogging)

| Value | Description |
| ----- | ----------- |
| `default` | LogComponentGatewayDefault defines the "default"-wide logging component. When specified,<br />all other logging components are ignored.<br /> | 
| `provider` | LogComponentProviderRunner defines the "provider" runner component.<br /> | 
| `gateway-api` | LogComponentGatewayAPIRunner defines the "gateway-api" runner component.<br /> | 
| `xds-translator` | LogComponentXdsTranslatorRunner defines the "xds-translator" runner component.<br /> | 
| `xds-server` | LogComponentXdsServerRunner defines the "xds-server" runner component.<br /> | 
| `infrastructure` | LogComponentInfrastructureRunner defines the "infrastructure" runner component.<br /> | 
| `global-ratelimit` | LogComponentGlobalRateLimitRunner defines the "global-ratelimit" runner component.<br /> | 


#### EnvoyGatewayLogging



EnvoyGatewayLogging defines logging for Envoy Gateway.

_Appears in:_
- [EnvoyGateway](#envoygateway)
- [EnvoyGatewaySpec](#envoygatewayspec)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `level` | _object (keys:[EnvoyGatewayLogComponent](#envoygatewaylogcomponent), values:[LogLevel](#loglevel))_ |  true  | \{ default:info \} | Level is the logging level. If unspecified, defaults to "info".<br />EnvoyGatewayLogComponent options: default/provider/gateway-api/xds-translator/xds-server/infrastructure/global-ratelimit.<br />LogLevel options: debug/info/error/warn. |


#### EnvoyGatewayMetricSink



EnvoyGatewayMetricSink defines control plane
metric sinks where metrics are sent to.

_Appears in:_
- [EnvoyGatewayMetrics](#envoygatewaymetrics)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `type` | _[MetricSinkType](#metricsinktype)_ |  true  | OpenTelemetry | Type defines the metric sink type.<br />EG control plane currently supports OpenTelemetry. |
| `openTelemetry` | _[EnvoyGatewayOpenTelemetrySink](#envoygatewayopentelemetrysink)_ |  true  |  | OpenTelemetry defines the configuration for OpenTelemetry sink.<br />It's required if the sink type is OpenTelemetry. |


#### EnvoyGatewayMetrics



EnvoyGatewayMetrics defines control plane push/pull metrics configurations.

_Appears in:_
- [EnvoyGatewayTelemetry](#envoygatewaytelemetry)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `sinks` | _[EnvoyGatewayMetricSink](#envoygatewaymetricsink) array_ |  true  |  | Sinks defines the metric sinks where metrics are sent to. |
| `prometheus` | _[EnvoyGatewayPrometheusProvider](#envoygatewayprometheusprovider)_ |  true  |  | Prometheus defines the configuration for prometheus endpoint. |


#### EnvoyGatewayOpenTelemetrySink





_Appears in:_
- [EnvoyGatewayMetricSink](#envoygatewaymetricsink)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `host` | _string_ |  true  |  | Host define the sink service hostname. |
| `protocol` | _string_ |  true  |  | Protocol define the sink service protocol. |
| `port` | _integer_ |  false  | 4317 | Port defines the port the sink service is exposed on. |
| `exportInterval` | _[Duration](https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1.Duration)_ |  true  |  | ExportInterval configures the intervening time between exports for a<br />Sink. This option overrides any value set for the<br />OTEL_METRIC_EXPORT_INTERVAL environment variable.<br />If ExportInterval is less than or equal to zero, 60 seconds<br />is used as the default. |
| `exportTimeout` | _[Duration](https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1.Duration)_ |  true  |  | ExportTimeout configures the time a Sink waits for an export to<br />complete before canceling it. This option overrides any value set for the<br />OTEL_METRIC_EXPORT_TIMEOUT environment variable.<br />If ExportTimeout is less than or equal to zero, 30 seconds<br />is used as the default. |


#### EnvoyGatewayPrometheusProvider



EnvoyGatewayPrometheusProvider will expose prometheus endpoint in pull mode.

_Appears in:_
- [EnvoyGatewayMetrics](#envoygatewaymetrics)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `disable` | _boolean_ |  true  |  | Disable defines if disables the prometheus metrics in pull mode. |


#### EnvoyGatewayProvider



EnvoyGatewayProvider defines the desired configuration of a provider.

_Appears in:_
- [EnvoyGateway](#envoygateway)
- [EnvoyGatewaySpec](#envoygatewayspec)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `type` | _[ProviderType](#providertype)_ |  true  |  | Type is the type of provider to use. Supported types are "Kubernetes", "Custom". |
| `kubernetes` | _[EnvoyGatewayKubernetesProvider](#envoygatewaykubernetesprovider)_ |  false  |  | Kubernetes defines the configuration of the Kubernetes provider. Kubernetes<br />provides runtime configuration via the Kubernetes API. |
| `custom` | _[EnvoyGatewayCustomProvider](#envoygatewaycustomprovider)_ |  false  |  | Custom defines the configuration for the Custom provider. This provider<br />allows you to define a specific resource provider and an infrastructure<br />provider. |


#### EnvoyGatewayResourceProvider



EnvoyGatewayResourceProvider defines configuration for the Custom Resource provider.

_Appears in:_
- [EnvoyGatewayCustomProvider](#envoygatewaycustomprovider)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `type` | _[ResourceProviderType](#resourceprovidertype)_ |  true  |  | Type is the type of resource provider to use. Supported types are "File". |
| `file` | _[EnvoyGatewayFileResourceProvider](#envoygatewayfileresourceprovider)_ |  false  |  | File defines the configuration of the File provider. File provides runtime<br />configuration defined by one or more files. |


#### EnvoyGatewaySpec



EnvoyGatewaySpec defines the desired state of Envoy Gateway.

_Appears in:_
- [EnvoyGateway](#envoygateway)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `gateway` | _[Gateway](#gateway)_ |  false  |  | Gateway defines desired Gateway API specific configuration. If unset,<br />default configuration parameters will apply. |
| `provider` | _[EnvoyGatewayProvider](#envoygatewayprovider)_ |  false  |  | Provider defines the desired provider and provider-specific configuration.<br />If unspecified, the Kubernetes provider is used with default configuration<br />parameters. |
| `logging` | _[EnvoyGatewayLogging](#envoygatewaylogging)_ |  false  | \{ default:info \} | Logging defines logging parameters for Envoy Gateway. |
| `admin` | _[EnvoyGatewayAdmin](#envoygatewayadmin)_ |  false  |  | Admin defines the desired admin related abilities.<br />If unspecified, the Admin is used with default configuration<br />parameters. |
| `telemetry` | _[EnvoyGatewayTelemetry](#envoygatewaytelemetry)_ |  false  |  | Telemetry defines the desired control plane telemetry related abilities.<br />If unspecified, the telemetry is used with default configuration. |
| `rateLimit` | _[RateLimit](#ratelimit)_ |  false  |  | RateLimit defines the configuration associated with the Rate Limit service<br />deployed by Envoy Gateway required to implement the Global Rate limiting<br />functionality. The specific rate limit service used here is the reference<br />implementation in Envoy. For more details visit https://github.com/envoyproxy/ratelimit.<br />This configuration is unneeded for "Local" rate limiting. |
| `extensionManager` | _[ExtensionManager](#extensionmanager)_ |  false  |  | ExtensionManager defines an extension manager to register for the Envoy Gateway Control Plane. |
| `extensionApis` | _[ExtensionAPISettings](#extensionapisettings)_ |  false  |  | ExtensionAPIs defines the settings related to specific Gateway API Extensions<br />implemented by Envoy Gateway |


#### EnvoyGatewayTelemetry



EnvoyGatewayTelemetry defines telemetry configurations for envoy gateway control plane.
Control plane will focus on metrics observability telemetry and tracing telemetry later.

_Appears in:_
- [EnvoyGateway](#envoygateway)
- [EnvoyGatewaySpec](#envoygatewayspec)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `metrics` | _[EnvoyGatewayMetrics](#envoygatewaymetrics)_ |  true  |  | Metrics defines metrics configuration for envoy gateway. |


#### EnvoyGatewayTopologyInjector



EnvoyGatewayTopologyInjector defines the configuration for topology injector MutatatingWebhookConfiguration

_Appears in:_
- [EnvoyGatewayKubernetesProvider](#envoygatewaykubernetesprovider)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `disabled` | _boolean_ |  false  |  |  |


#### EnvoyJSONPatchConfig



EnvoyJSONPatchConfig defines the configuration for patching a Envoy xDS Resource
using JSONPatch semantic

_Appears in:_
- [EnvoyPatchPolicySpec](#envoypatchpolicyspec)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `type` | _[EnvoyResourceType](#envoyresourcetype)_ |  true  |  | Type is the typed URL of the Envoy xDS Resource |
| `name` | _string_ |  true  |  | Name is the name of the resource |
| `operation` | _[JSONPatchOperation](#jsonpatchoperation)_ |  true  |  | Patch defines the JSON Patch Operation |


#### EnvoyPatchPolicy



EnvoyPatchPolicy allows the user to modify the generated Envoy xDS
resources by Envoy Gateway using this patch API



| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `apiVersion` | _string_ | |`gateway.envoyproxy.io/v1alpha1`
| `kind` | _string_ | |`EnvoyPatchPolicy`
| `metadata` | _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#objectmeta-v1-meta)_ |  true  |  | Refer to Kubernetes API documentation for fields of `metadata`. |
| `spec` | _[EnvoyPatchPolicySpec](#envoypatchpolicyspec)_ |  true  |  | Spec defines the desired state of EnvoyPatchPolicy. |
| `status` | _[PolicyStatus](https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1alpha2.PolicyStatus)_ |  true  |  | Status defines the current status of EnvoyPatchPolicy. |


#### EnvoyPatchPolicySpec



EnvoyPatchPolicySpec defines the desired state of EnvoyPatchPolicy.

_Appears in:_
- [EnvoyPatchPolicy](#envoypatchpolicy)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `type` | _[EnvoyPatchType](#envoypatchtype)_ |  true  |  | Type decides the type of patch.<br />Valid EnvoyPatchType values are "JSONPatch". |
| `jsonPatches` | _[EnvoyJSONPatchConfig](#envoyjsonpatchconfig) array_ |  false  |  | JSONPatch defines the JSONPatch configuration. |
| `targetRef` | _[LocalPolicyTargetReference](https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1alpha2.LocalPolicyTargetReference)_ |  true  |  | TargetRef is the name of the Gateway API resource this policy<br />is being attached to.<br />By default, attaching to Gateway is supported and<br />when mergeGateways is enabled it should attach to GatewayClass.<br />This Policy and the TargetRef MUST be in the same namespace<br />for this Policy to have effect and be applied to the Gateway<br />TargetRef |
| `priority` | _integer_ |  true  |  | Priority of the EnvoyPatchPolicy.<br />If multiple EnvoyPatchPolicies are applied to the same<br />TargetRef, they will be applied in the ascending order of<br />the priority i.e. int32.min has the highest priority and<br />int32.max has the lowest priority.<br />Defaults to 0. |


#### EnvoyPatchType

_Underlying type:_ _string_

EnvoyPatchType specifies the types of Envoy patching mechanisms.

_Appears in:_
- [EnvoyPatchPolicySpec](#envoypatchpolicyspec)

| Value | Description |
| ----- | ----------- |
| `JSONPatch` | JSONPatchEnvoyPatchType allows the user to patch the generated xDS resources using JSONPatch semantics.<br />For more details on the semantics, please refer to https://datatracker.ietf.org/doc/html/rfc6902<br /> | 


#### EnvoyProxy



EnvoyProxy is the schema for the envoyproxies API.



| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `apiVersion` | _string_ | |`gateway.envoyproxy.io/v1alpha1`
| `kind` | _string_ | |`EnvoyProxy`
| `metadata` | _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#objectmeta-v1-meta)_ |  true  |  | Refer to Kubernetes API documentation for fields of `metadata`. |
| `spec` | _[EnvoyProxySpec](#envoyproxyspec)_ |  true  |  | EnvoyProxySpec defines the desired state of EnvoyProxy. |
| `status` | _[EnvoyProxyStatus](#envoyproxystatus)_ |  true  |  | EnvoyProxyStatus defines the actual state of EnvoyProxy. |


#### EnvoyProxyKubernetesProvider



EnvoyProxyKubernetesProvider defines configuration for the Kubernetes resource
provider.

_Appears in:_
- [EnvoyProxyProvider](#envoyproxyprovider)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `envoyDeployment` | _[KubernetesDeploymentSpec](#kubernetesdeploymentspec)_ |  false  |  | EnvoyDeployment defines the desired state of the Envoy deployment resource.<br />If unspecified, default settings for the managed Envoy deployment resource<br />are applied. |
| `envoyDaemonSet` | _[KubernetesDaemonSetSpec](#kubernetesdaemonsetspec)_ |  false  |  | EnvoyDaemonSet defines the desired state of the Envoy daemonset resource.<br />Disabled by default, a deployment resource is used instead to provision the Envoy Proxy fleet |
| `envoyService` | _[KubernetesServiceSpec](#kubernetesservicespec)_ |  false  |  | EnvoyService defines the desired state of the Envoy service resource.<br />If unspecified, default settings for the managed Envoy service resource<br />are applied. |
| `envoyHpa` | _[KubernetesHorizontalPodAutoscalerSpec](#kuberneteshorizontalpodautoscalerspec)_ |  false  |  | EnvoyHpa defines the Horizontal Pod Autoscaler settings for Envoy Proxy Deployment. |
| `useListenerPortAsContainerPort` | _boolean_ |  false  |  | UseListenerPortAsContainerPort disables the port shifting feature in the Envoy Proxy.<br />When set to false (default value), if the service port is a privileged port (1-1023), add a constant to the value converting it into an ephemeral port.<br />This allows the container to bind to the port without needing a CAP_NET_BIND_SERVICE capability. |
| `envoyPDB` | _[KubernetesPodDisruptionBudgetSpec](#kubernetespoddisruptionbudgetspec)_ |  false  |  | EnvoyPDB allows to control the pod disruption budget of an Envoy Proxy. |


#### EnvoyProxyProvider



EnvoyProxyProvider defines the desired state of a resource provider.

_Appears in:_
- [EnvoyProxySpec](#envoyproxyspec)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `type` | _[ProviderType](#providertype)_ |  true  |  | Type is the type of resource provider to use. A resource provider provides<br />infrastructure resources for running the data plane, e.g. Envoy proxy, and<br />optional auxiliary control planes. Supported types are "Kubernetes". |
| `kubernetes` | _[EnvoyProxyKubernetesProvider](#envoyproxykubernetesprovider)_ |  false  |  | Kubernetes defines the desired state of the Kubernetes resource provider.<br />Kubernetes provides infrastructure resources for running the data plane,<br />e.g. Envoy proxy. If unspecified and type is "Kubernetes", default settings<br />for managed Kubernetes resources are applied. |


#### EnvoyProxySpec



EnvoyProxySpec defines the desired state of EnvoyProxy.

_Appears in:_
- [EnvoyProxy](#envoyproxy)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `provider` | _[EnvoyProxyProvider](#envoyproxyprovider)_ |  false  |  | Provider defines the desired resource provider and provider-specific configuration.<br />If unspecified, the "Kubernetes" resource provider is used with default configuration<br />parameters. |
| `logging` | _[ProxyLogging](#proxylogging)_ |  true  | \{ level:map[default:warn] \} | Logging defines logging parameters for managed proxies. |
| `telemetry` | _[ProxyTelemetry](#proxytelemetry)_ |  false  |  | Telemetry defines telemetry parameters for managed proxies. |
| `bootstrap` | _[ProxyBootstrap](#proxybootstrap)_ |  false  |  | Bootstrap defines the Envoy Bootstrap as a YAML string.<br />Visit https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/bootstrap/v3/bootstrap.proto#envoy-v3-api-msg-config-bootstrap-v3-bootstrap<br />to learn more about the syntax.<br />If set, this is the Bootstrap configuration used for the managed Envoy Proxy fleet instead of the default Bootstrap configuration<br />set by Envoy Gateway.<br />Some fields within the Bootstrap that are required to communicate with the xDS Server (Envoy Gateway) and receive xDS resources<br />from it are not configurable and will result in the `EnvoyProxy` resource being rejected.<br />Backward compatibility across minor versions is not guaranteed.<br />We strongly recommend using `egctl x translate` to generate a `EnvoyProxy` resource with the `Bootstrap` field set to the default<br />Bootstrap configuration used. You can edit this configuration, and rerun `egctl x translate` to ensure there are no validation errors. |
| `concurrency` | _integer_ |  false  |  | Concurrency defines the number of worker threads to run. If unset, it defaults to<br />the number of cpuset threads on the platform. |
| `routingType` | _[RoutingType](#routingtype)_ |  false  |  | RoutingType can be set to "Service" to use the Service Cluster IP for routing to the backend,<br />or it can be set to "Endpoint" to use Endpoint routing. The default is "Endpoint". |
| `extraArgs` | _string array_ |  false  |  | ExtraArgs defines additional command line options that are provided to Envoy.<br />More info: https://www.envoyproxy.io/docs/envoy/latest/operations/cli#command-line-options<br />Note: some command line options are used internally(e.g. --log-level) so they cannot be provided here. |
| `mergeGateways` | _boolean_ |  false  |  | MergeGateways defines if Gateway resources should be merged onto the same Envoy Proxy Infrastructure.<br />Setting this field to true would merge all Gateway Listeners under the parent Gateway Class.<br />This means that the port, protocol and hostname tuple must be unique for every listener.<br />If a duplicate listener is detected, the newer listener (based on timestamp) will be rejected and its status will be updated with a "Accepted=False" condition. |
| `shutdown` | _[ShutdownConfig](#shutdownconfig)_ |  false  |  | Shutdown defines configuration for graceful envoy shutdown process. |
| `filterOrder` | _[FilterPosition](#filterposition) array_ |  false  |  | FilterOrder defines the order of filters in the Envoy proxy's HTTP filter chain.<br />The FilterPosition in the list will be applied in the order they are defined.<br />If unspecified, the default filter order is applied.<br />Default filter order is:<br />- envoy.filters.http.health_check<br />- envoy.filters.http.fault<br />- envoy.filters.http.cors<br />- envoy.filters.http.ext_authz<br />- envoy.filters.http.basic_auth<br />- envoy.filters.http.oauth2<br />- envoy.filters.http.jwt_authn<br />- envoy.filters.http.stateful_session<br />- envoy.filters.http.lua<br />- envoy.filters.http.ext_proc<br />- envoy.filters.http.wasm<br />- envoy.filters.http.rbac<br />- envoy.filters.http.local_ratelimit<br />- envoy.filters.http.ratelimit<br />- envoy.filters.http.custom_response<br />- envoy.filters.http.router<br />Note: "envoy.filters.http.router" cannot be reordered, it's always the last filter in the chain. |
| `backendTLS` | _[BackendTLSConfig](#backendtlsconfig)_ |  false  |  | BackendTLS is the TLS configuration for the Envoy proxy to use when connecting to backends.<br />These settings are applied on backends for which TLS policies are specified. |
| `ipFamily` | _[IPFamily](#ipfamily)_ |  false  |  | IPFamily specifies the IP family for the EnvoyProxy fleet.<br />This setting only affects the Gateway listener port and does not impact<br />other aspects of the Envoy proxy configuration.<br />If not specified, the system will operate as follows:<br />- It defaults to IPv4 only.<br />- IPv6 and dual-stack environments are not supported in this default configuration.<br />Note: To enable IPv6 or dual-stack functionality, explicit configuration is required. |
| `preserveRouteOrder` | _boolean_ |  false  |  | PreserveRouteOrder determines if the order of matching for HTTPRoutes is determined by Gateway-API<br />specification (https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1.HTTPRouteRule)<br />or preserves the order defined by users in the HTTPRoute's HTTPRouteRule list.<br />Default: False |
| `disableLuaValidation` | _boolean_ |  false  | false | DisableLuaValidation disables the Lua script validation for Lua EnvoyExtensionPolicies |


#### EnvoyProxyStatus



EnvoyProxyStatus defines the observed state of EnvoyProxy. This type is not implemented
until https://github.com/envoyproxy/gateway/issues/1007 is fixed.

_Appears in:_
- [EnvoyProxy](#envoyproxy)



#### EnvoyResourceType

_Underlying type:_ _string_

EnvoyResourceType specifies the type URL of the Envoy resource.

_Appears in:_
- [EnvoyJSONPatchConfig](#envoyjsonpatchconfig)

| Value | Description |
| ----- | ----------- |
| `type.googleapis.com/envoy.config.listener.v3.Listener` | ListenerEnvoyResourceType defines the Type URL of the Listener resource<br /> | 
| `type.googleapis.com/envoy.config.route.v3.RouteConfiguration` | RouteConfigurationEnvoyResourceType defines the Type URL of the RouteConfiguration resource<br /> | 
| `type.googleapis.com/envoy.config.cluster.v3.Cluster` | ClusterEnvoyResourceType defines the Type URL of the Cluster resource<br /> | 
| `type.googleapis.com/envoy.config.endpoint.v3.ClusterLoadAssignment` | ClusterLoadAssignmentEnvoyResourceType defines the Type URL of the ClusterLoadAssignment resource<br /> | 
| `type.googleapis.com/envoy.extensions.transport_sockets.tls.v3.Secret` | SecretEnvoyResourceType defines the Type URL of the Secret resource<br /> | 


#### ExtAuth



ExtAuth defines the configuration for External Authorization.

_Appears in:_
- [SecurityPolicySpec](#securitypolicyspec)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `grpc` | _[GRPCExtAuthService](#grpcextauthservice)_ |  true  |  | GRPC defines the gRPC External Authorization service.<br />Either GRPCService or HTTPService must be specified,<br />and only one of them can be provided. |
| `http` | _[HTTPExtAuthService](#httpextauthservice)_ |  true  |  | HTTP defines the HTTP External Authorization service.<br />Either GRPCService or HTTPService must be specified,<br />and only one of them can be provided. |
| `headersToExtAuth` | _string array_ |  false  |  | HeadersToExtAuth defines the client request headers that will be included<br />in the request to the external authorization service.<br />Note: If not specified, the default behavior for gRPC and HTTP external<br />authorization services is different due to backward compatibility reasons.<br />All headers will be included in the check request to a gRPC authorization server.<br />Only the following headers will be included in the check request to an HTTP<br />authorization server: Host, Method, Path, Content-Length, and Authorization.<br />And these headers will always be included to the check request to an HTTP<br />authorization server by default, no matter whether they are specified<br />in HeadersToExtAuth or not. |
| `bodyToExtAuth` | _[BodyToExtAuth](#bodytoextauth)_ |  false  |  | BodyToExtAuth defines the Body to Ext Auth configuration. |
| `failOpen` | _boolean_ |  false  | false | FailOpen is a switch used to control the behavior when a response from the External Authorization service cannot be obtained.<br />If FailOpen is set to true, the system allows the traffic to pass through.<br />Otherwise, if it is set to false or not set (defaulting to false),<br />the system blocks the traffic and returns a HTTP 5xx error, reflecting a fail-closed approach.<br />This setting determines whether to prioritize accessibility over strict security in case of authorization service failure. |
| `recomputeRoute` | _boolean_ |  false  |  | RecomputeRoute clears the route cache and recalculates the routing decision.<br />This field must be enabled if the headers added or modified by the ExtAuth are used for<br />route matching decisions. If the recomputation selects a new route, features targeting<br />the new matched route will be applied. |


#### ExtProc



ExtProc defines the configuration for External Processing filter.

_Appears in:_
- [EnvoyExtensionPolicySpec](#envoyextensionpolicyspec)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `backendRef` | _[BackendObjectReference](https://gateway-api.sigs.k8s.io/references/spec/#gateway.networking.k8s.io/v1.BackendObjectReference)_ |  false  |  | BackendRef references a Kubernetes object that represents the<br />backend server to which the authorization request will be sent.<br />Deprecated: Use BackendRefs instead. |
| `backendRefs` | _[BackendRef](#backendref) array_ |  false  |  | BackendRefs references a Kubernetes object that represents the<br />backend server to which the authorization request will be sent. |
| `backendSettings` | _[ClusterSettings](#clustersettings)_ |  false  |  | BackendSettings holds configuration for managing the connection<br />to the backend. |
| `messageTimeout` | _[Duration](https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1.Duration)_ |  false  |  | MessageTimeout is the timeout for a response to be returned from the external processor<br />Default: 200ms |
| `failOpen` | _boolean_ |  false  |  | FailOpen defines if requests or responses that cannot be processed due to connectivity to the<br />external processor are terminated or passed-through.<br />Default: false |
| `processingMode` | _[ExtProcProcessingMode](#extprocprocessingmode)_ |  false  |  | ProcessingMode defines how request and response body is processed<br />Default: header and body are not sent to the external processor |
| `metadata` | _[ExtProcMetadata](#extprocmetadata)_ |  false  |  | Refer to Kubernetes API documentation for fields of `metadata`. |


#### ExtProcBodyProcessingMode

_Underlying type:_ _string_



_Appears in:_
- [ProcessingModeOptions](#processingmodeoptions)

| Value | Description |
| ----- | ----------- |
| `Streamed` | StreamedExtProcBodyProcessingMode will stream the body to the server in pieces as they arrive at the proxy.<br /> | 
| `Buffered` | BufferedExtProcBodyProcessingMode will buffer the message body in memory and send the entire body at once. If the body exceeds the configured buffer limit, then the downstream system will receive an error.<br /> | 
| `FullDuplexStreamed` | FullDuplexStreamedExtBodyProcessingMode will send the body in pieces, to be read in a stream. Full details here: https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/filters/http/ext_proc/v3/processing_mode.proto.html#enum-extensions-filters-http-ext-proc-v3-processingmode-bodysendmode<br /> | 
| `BufferedPartial` | BufferedPartialExtBodyHeaderProcessingMode will buffer the message body in memory and send the entire body in one chunk. If the body exceeds the configured buffer limit, then the body contents up to the buffer limit will be sent.<br /> | 


#### ExtProcMetadata



ExtProcMetadata defines options related to the sending and receiving of dynamic metadata to and from the
external processor service

_Appears in:_
- [ExtProc](#extproc)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `accessibleNamespaces` | _string array_ |  false  |  | AccessibleNamespaces are metadata namespaces that are sent to the external processor as context |
| `writableNamespaces` | _string array_ |  false  |  | WritableNamespaces are metadata namespaces that the external processor can write to |


#### ExtProcProcessingMode



ExtProcProcessingMode defines if and how headers and bodies are sent to the service.
https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/filters/http/ext_proc/v3/processing_mode.proto#envoy-v3-api-msg-extensions-filters-http-ext-proc-v3-processingmode

_Appears in:_
- [ExtProc](#extproc)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `request` | _[ProcessingModeOptions](#processingmodeoptions)_ |  false  |  | Defines processing mode for requests. If present, request headers are sent. Request body is processed according<br />to the specified mode. |
| `response` | _[ProcessingModeOptions](#processingmodeoptions)_ |  false  |  | Defines processing mode for responses. If present, response headers are sent. Response body is processed according<br />to the specified mode. |
| `allowModeOverride` | _boolean_ |  false  |  | AllowModeOverride allows the external processor to override the processing mode set via the<br />`mode_override` field in the gRPC response message. This defaults to false. |


#### ExtensionAPISettings



ExtensionAPISettings defines the settings specific to Gateway API Extensions.

_Appears in:_
- [EnvoyGateway](#envoygateway)
- [EnvoyGatewaySpec](#envoygatewayspec)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `enableEnvoyPatchPolicy` | _boolean_ |  true  |  | EnableEnvoyPatchPolicy enables Envoy Gateway to<br />reconcile and implement the EnvoyPatchPolicy resources. |
| `enableBackend` | _boolean_ |  true  |  | EnableBackend enables Envoy Gateway to<br />reconcile and implement the Backend resources. |


#### ExtensionHooks



ExtensionHooks defines extension hooks across all supported runners

_Appears in:_
- [ExtensionManager](#extensionmanager)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `xdsTranslator` | _[XDSTranslatorHooks](#xdstranslatorhooks)_ |  true  |  | XDSTranslator defines all the supported extension hooks for the xds-translator runner |


#### ExtensionManager



ExtensionManager defines the configuration for registering an extension manager to
the Envoy Gateway control plane.

_Appears in:_
- [EnvoyGateway](#envoygateway)
- [EnvoyGatewaySpec](#envoygatewayspec)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `resources` | _[GroupVersionKind](#groupversionkind) array_ |  false  |  | Resources defines the set of K8s resources the extension will handle as route<br />filter resources |
| `policyResources` | _[GroupVersionKind](#groupversionkind) array_ |  false  |  | PolicyResources defines the set of K8S resources the extension server will handle<br />as directly attached GatewayAPI policies |
| `hooks` | _[ExtensionHooks](#extensionhooks)_ |  true  |  | Hooks defines the set of hooks the extension supports |
| `service` | _[ExtensionService](#extensionservice)_ |  true  |  | Service defines the configuration of the extension service that the Envoy<br />Gateway Control Plane will call through extension hooks. |
| `failOpen` | _boolean_ |  false  |  | FailOpen defines if Envoy Gateway should ignore errors returned from the Extension Service hooks.<br />When set to false, Envoy Gateway does not ignore extension Service hook errors. As a result,<br />xDS updates are skipped for the relevant envoy proxy fleet and the previous state is preserved.<br />When set to true, if the Extension Service hooks return an error, no changes will be applied to the<br />source of the configuration which was sent to the extension server. The errors are ignored and the resulting<br />xDS configuration is updated in the xDS snapshot.<br />Default: false |
| `maxMessageSize` | _[Quantity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#quantity-resource-api)_ |  false  |  | MaxMessageSize defines the maximum message size in bytes that can be<br />sent to or received from the Extension Service.<br />Default: 4M |


#### ExtensionService



ExtensionService defines the configuration for connecting to a registered extension service.

_Appears in:_
- [ExtensionManager](#extensionmanager)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `fqdn` | _[FQDNEndpoint](#fqdnendpoint)_ |  false  |  | FQDN defines a FQDN endpoint |
| `ip` | _[IPEndpoint](#ipendpoint)_ |  false  |  | IP defines an IP endpoint. Supports both IPv4 and IPv6 addresses. |
| `unix` | _[UnixSocket](#unixsocket)_ |  false  |  | Unix defines the unix domain socket endpoint |
| `zone` | _string_ |  false  |  | Zone defines the service zone of the backend endpoint. |
| `host` | _string_ |  false  |  | Host define the extension service hostname.<br />Deprecated: use the appropriate transport attribute instead (FQDN,IP,Unix) |
| `port` | _integer_ |  false  | 80 | Port defines the port the extension service is exposed on.<br />Deprecated: use the appropriate transport attribute instead (FQDN,IP,Unix) |
| `tls` | _[ExtensionTLS](#extensiontls)_ |  false  |  | TLS defines TLS configuration for communication between Envoy Gateway and<br />the extension service. |
| `retry` | _[ExtensionServiceRetry](#extensionserviceretry)_ |  false  |  | Retry defines the retry policy for to use when errors are encountered in communication with<br />the extension service. |


#### ExtensionServiceRetry



ExtensionServiceRetry defines the retry policy for to use when errors are encountered in communication with the extension service.

_Appears in:_
- [ExtensionService](#extensionservice)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `maxAttempts` | _integer_ |  false  |  | MaxAttempts defines the maximum number of retry attempts.<br />Default: 4 |
| `initialBackoff` | _[Duration](https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1.Duration)_ |  false  |  | InitialBackoff defines the initial backoff in seconds for retries, details: https://github.com/grpc/proposal/blob/master/A6-client-retries.md#integration-with-service-config.<br />Default: 0.1s |
| `maxBackoff` | _[Duration](https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1.Duration)_ |  false  |  | MaxBackoff defines the maximum backoff in seconds for retries.<br />Default: 1s |
| `backoffMultiplier` | _[Fraction](https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1.Fraction)_ |  false  |  | BackoffMultiplier defines the multiplier to use for exponential backoff for retries.<br />Default: 2.0 |
| `RetryableStatusCodes` | _[RetryableGRPCStatusCode](#retryablegrpcstatuscode) array_ |  false  |  | RetryableStatusCodes defines the grpc status code for which retries will be attempted.<br />Default: [ "UNAVAILABLE" ] |


#### ExtensionTLS



ExtensionTLS defines the TLS configuration when connecting to an extension service.

_Appears in:_
- [ExtensionService](#extensionservice)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `certificateRef` | _[SecretObjectReference](https://gateway-api.sigs.k8s.io/references/spec/#gateway.networking.k8s.io/v1.SecretObjectReference)_ |  true  |  | CertificateRef is a reference to a Kubernetes Secret with a CA certificate in a key named "tls.crt".<br />The CA certificate is used by Envoy Gateway the verify the server certificate presented by the extension server.<br />At this time, Envoy Gateway does not support Client Certificate authentication of Envoy Gateway towards the extension server (mTLS). |


#### ExtractFrom



ExtractFrom is where to fetch the key from the coming request.
Only one of header, param or cookie is supposed to be specified.

_Appears in:_
- [APIKeyAuth](#apikeyauth)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `headers` | _string array_ |  false  |  | Headers is the names of the header to fetch the key from.<br />If multiple headers are specified, envoy will look for the api key in the order of the list.<br />This field is optional, but only one of headers, params or cookies is supposed to be specified. |
| `params` | _string array_ |  false  |  | Params is the names of the query parameter to fetch the key from.<br />If multiple params are specified, envoy will look for the api key in the order of the list.<br />This field is optional, but only one of headers, params or cookies is supposed to be specified. |
| `cookies` | _string array_ |  false  |  | Cookies is the names of the cookie to fetch the key from.<br />If multiple cookies are specified, envoy will look for the api key in the order of the list.<br />This field is optional, but only one of headers, params or cookies is supposed to be specified. |


#### FQDNEndpoint



FQDNEndpoint describes TCP/UDP socket address, corresponding to Envoy's Socket Address
https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/core/v3/address.proto#config-core-v3-socketaddress

_Appears in:_
- [BackendEndpoint](#backendendpoint)
- [ExtensionService](#extensionservice)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `hostname` | _string_ |  true  |  | Hostname defines the FQDN hostname of the backend endpoint. |
| `port` | _integer_ |  true  |  | Port defines the port of the backend endpoint. |


#### FaultInjection



FaultInjection defines the fault injection policy to be applied. This configuration can be used to
inject delays and abort requests to mimic failure scenarios such as service failures and overloads

_Appears in:_
- [BackendTrafficPolicySpec](#backendtrafficpolicyspec)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `delay` | _[FaultInjectionDelay](#faultinjectiondelay)_ |  false  |  | If specified, a delay will be injected into the request. |
| `abort` | _[FaultInjectionAbort](#faultinjectionabort)_ |  false  |  | If specified, the request will be aborted if it meets the configuration criteria. |


#### FaultInjectionAbort



FaultInjectionAbort defines the abort fault injection configuration

_Appears in:_
- [FaultInjection](#faultinjection)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `httpStatus` | _integer_ |  false  |  | StatusCode specifies the HTTP status code to be returned |
| `grpcStatus` | _integer_ |  false  |  | GrpcStatus specifies the GRPC status code to be returned |
| `percentage` | _float_ |  false  | 100 | Percentage specifies the percentage of requests to be aborted. Default 100%, if set 0, no requests will be aborted. Accuracy to 0.0001%. |


#### FaultInjectionDelay



FaultInjectionDelay defines the delay fault injection configuration

_Appears in:_
- [FaultInjection](#faultinjection)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `fixedDelay` | _[Duration](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#duration-v1-meta)_ |  true  |  | FixedDelay specifies the fixed delay duration |
| `percentage` | _float_ |  false  | 100 | Percentage specifies the percentage of requests to be delayed. Default 100%, if set 0, no requests will be delayed. Accuracy to 0.0001%. |


#### FileEnvoyProxyAccessLog





_Appears in:_
- [ProxyAccessLogSink](#proxyaccesslogsink)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `path` | _string_ |  true  |  | Path defines the file path used to expose envoy access log(e.g. /dev/stdout). |


#### FilterPosition



FilterPosition defines the position of an Envoy HTTP filter in the filter chain.

_Appears in:_
- [EnvoyProxySpec](#envoyproxyspec)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `name` | _[EnvoyFilter](#envoyfilter)_ |  true  |  | Name of the filter. |
| `before` | _[EnvoyFilter](#envoyfilter)_ |  true  |  | Before defines the filter that should come before the filter.<br />Only one of Before or After must be set. |
| `after` | _[EnvoyFilter](#envoyfilter)_ |  true  |  | After defines the filter that should come after the filter.<br />Only one of Before or After must be set. |


#### GRPCActiveHealthChecker



GRPCActiveHealthChecker defines the settings of the GRPC health check.

_Appears in:_
- [ActiveHealthCheck](#activehealthcheck)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `service` | _string_ |  false  |  | Service to send in the health check request.<br />If this is not specified, then the health check request applies to the entire<br />server and not to a specific service. |


#### GRPCExtAuthService



GRPCExtAuthService defines the gRPC External Authorization service
The authorization request message is defined in
https://www.envoyproxy.io/docs/envoy/latest/api-v3/service/auth/v3/external_auth.proto

_Appears in:_
- [ExtAuth](#extauth)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `backendRef` | _[BackendObjectReference](https://gateway-api.sigs.k8s.io/references/spec/#gateway.networking.k8s.io/v1.BackendObjectReference)_ |  false  |  | BackendRef references a Kubernetes object that represents the<br />backend server to which the authorization request will be sent.<br />Deprecated: Use BackendRefs instead. |
| `backendRefs` | _[BackendRef](#backendref) array_ |  false  |  | BackendRefs references a Kubernetes object that represents the<br />backend server to which the authorization request will be sent. |
| `backendSettings` | _[ClusterSettings](#clustersettings)_ |  false  |  | BackendSettings holds configuration for managing the connection<br />to the backend. |


#### Gateway



Gateway defines the desired Gateway API configuration of Envoy Gateway.

_Appears in:_
- [EnvoyGateway](#envoygateway)
- [EnvoyGatewaySpec](#envoygatewayspec)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `controllerName` | _string_ |  false  |  | ControllerName defines the name of the Gateway API controller. If unspecified,<br />defaults to "gateway.envoyproxy.io/gatewayclass-controller". See the following<br />for additional details:<br />  https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1.GatewayClass |


#### GlobalRateLimit



GlobalRateLimit defines global rate limit configuration.

_Appears in:_
- [RateLimitSpec](#ratelimitspec)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `rules` | _[RateLimitRule](#ratelimitrule) array_ |  true  |  | Rules are a list of RateLimit selectors and limits. Each rule and its<br />associated limit is applied in a mutually exclusive way. If a request<br />matches multiple rules, each of their associated limits get applied, so a<br />single request might increase the rate limit counters for multiple rules<br />if selected. The rate limit service will return a logical OR of the individual<br />rate limit decisions of all matching rules. For example, if a request<br />matches two rules, one rate limited and one not, the final decision will be<br />to rate limit the request. |


#### GroupVersionKind



GroupVersionKind unambiguously identifies a Kind.
It can be converted to k8s.io/apimachinery/pkg/runtime/schema.GroupVersionKind

_Appears in:_
- [ExtensionManager](#extensionmanager)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `group` | _string_ |  true  |  |  |
| `version` | _string_ |  true  |  |  |
| `kind` | _string_ |  true  |  |  |


#### GzipCompressor



GzipCompressor defines the config for the Gzip compressor.
The default values can be found here:
https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/compression/gzip/compressor/v3/gzip.proto#extension-envoy-compression-gzip-compressor

_Appears in:_
- [Compression](#compression)



#### HTTP10Settings



HTTP10Settings provides HTTP/1.0 configuration on the listener.

_Appears in:_
- [HTTP1Settings](#http1settings)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `useDefaultHost` | _boolean_ |  false  |  | UseDefaultHost defines if the HTTP/1.0 request is missing the Host header,<br />then the hostname associated with the listener should be injected into the<br />request.<br />If this is not set and an HTTP/1.0 request arrives without a host, then<br />it will be rejected. |


#### HTTP1Settings



HTTP1Settings provides HTTP/1 configuration on the listener.

_Appears in:_
- [ClientTrafficPolicySpec](#clienttrafficpolicyspec)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `enableTrailers` | _boolean_ |  false  |  | EnableTrailers defines if HTTP/1 trailers should be proxied by Envoy. |
| `preserveHeaderCase` | _boolean_ |  false  |  | PreserveHeaderCase defines if Envoy should preserve the letter case of headers.<br />By default, Envoy will lowercase all the headers. |
| `http10` | _[HTTP10Settings](#http10settings)_ |  false  |  | HTTP10 turns on support for HTTP/1.0 and HTTP/0.9 requests. |


#### HTTP2Settings



HTTP2Settings provides HTTP/2 configuration for listeners and backends.

_Appears in:_
- [BackendTrafficPolicySpec](#backendtrafficpolicyspec)
- [ClientTrafficPolicySpec](#clienttrafficpolicyspec)
- [ClusterSettings](#clustersettings)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `initialStreamWindowSize` | _[Quantity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#quantity-resource-api)_ |  false  |  | InitialStreamWindowSize sets the initial window size for HTTP/2 streams.<br />If not set, the default value is 64 KiB(64*1024). |
| `initialConnectionWindowSize` | _[Quantity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#quantity-resource-api)_ |  false  |  | InitialConnectionWindowSize sets the initial window size for HTTP/2 connections.<br />If not set, the default value is 1 MiB. |
| `maxConcurrentStreams` | _integer_ |  false  |  | MaxConcurrentStreams sets the maximum number of concurrent streams allowed per connection.<br />If not set, the default value is 100. |
| `onInvalidMessage` | _[InvalidMessageAction](#invalidmessageaction)_ |  false  |  | OnInvalidMessage determines if Envoy will terminate the connection or just the offending stream in the event of HTTP messaging error<br />It's recommended for L2 Envoy deployments to set this value to TerminateStream.<br />https://www.envoyproxy.io/docs/envoy/latest/configuration/best_practices/level_two<br />Default: TerminateConnection |


#### HTTP3Settings



HTTP3Settings provides HTTP/3 configuration on the listener.

_Appears in:_
- [ClientTrafficPolicySpec](#clienttrafficpolicyspec)



#### HTTPActiveHealthChecker



HTTPActiveHealthChecker defines the settings of http health check.

_Appears in:_
- [ActiveHealthCheck](#activehealthcheck)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `hostname` | _string_ |  false  |  | Hostname defines the HTTP host that will be requested during health checking.<br />Default: HTTPRoute or GRPCRoute hostname. |
| `path` | _string_ |  true  |  | Path defines the HTTP path that will be requested during health checking. |
| `method` | _string_ |  false  |  | Method defines the HTTP method used for health checking.<br />Defaults to GET |
| `expectedStatuses` | _[HTTPStatus](#httpstatus) array_ |  false  |  | ExpectedStatuses defines a list of HTTP response statuses considered healthy.<br />Defaults to 200 only |
| `expectedResponse` | _[ActiveHealthCheckPayload](#activehealthcheckpayload)_ |  false  |  | ExpectedResponse defines a list of HTTP expected responses to match. |


#### HTTPClientTimeout





_Appears in:_
- [ClientTimeout](#clienttimeout)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `requestReceivedTimeout` | _[Duration](https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1.Duration)_ |  false  |  | RequestReceivedTimeout is the duration envoy waits for the complete request reception. This timer starts upon request<br />initiation and stops when either the last byte of the request is sent upstream or when the response begins. |
| `idleTimeout` | _[Duration](https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1.Duration)_ |  false  |  | IdleTimeout for an HTTP connection. Idle time is defined as a period in which there are no active requests in the connection.<br />Default: 1 hour. |


#### HTTPCredentialInjectionFilter



HTTPCredentialInjectionFilter defines the configuration to inject credentials into the request.
This is useful when the backend service requires credentials in the request, and the original
request does not contain them. The filter can inject credentials into the request before forwarding
it to the backend service.

_Appears in:_
- [HTTPRouteFilterSpec](#httproutefilterspec)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `header` | _string_ |  false  |  | Header is the name of the header where the credentials are injected.<br />If not specified, the credentials are injected into the Authorization header. |
| `overwrite` | _boolean_ |  false  |  | Whether to overwrite the value or not if the injected headers already exist.<br />If not specified, the default value is false. |
| `credential` | _[InjectedCredential](#injectedcredential)_ |  true  |  | Credential is the credential to be injected. |


#### HTTPDirectResponseFilter



HTTPDirectResponseFilter defines the configuration to return a fixed response.

_Appears in:_
- [HTTPRouteFilterSpec](#httproutefilterspec)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `contentType` | _string_ |  false  |  | Content Type of the response. This will be set in the Content-Type header. |
| `body` | _[CustomResponseBody](#customresponsebody)_ |  false  |  | Body of the Response |
| `statusCode` | _integer_ |  false  |  | Status Code of the HTTP response<br />If unset, defaults to 200. |


#### HTTPExtAuthService



HTTPExtAuthService defines the HTTP External Authorization service

_Appears in:_
- [ExtAuth](#extauth)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `backendRef` | _[BackendObjectReference](https://gateway-api.sigs.k8s.io/references/spec/#gateway.networking.k8s.io/v1.BackendObjectReference)_ |  false  |  | BackendRef references a Kubernetes object that represents the<br />backend server to which the authorization request will be sent.<br />Deprecated: Use BackendRefs instead. |
| `backendRefs` | _[BackendRef](#backendref) array_ |  false  |  | BackendRefs references a Kubernetes object that represents the<br />backend server to which the authorization request will be sent. |
| `backendSettings` | _[ClusterSettings](#clustersettings)_ |  false  |  | BackendSettings holds configuration for managing the connection<br />to the backend. |
| `path` | _string_ |  false  |  | Path is the path of the HTTP External Authorization service.<br />If path is specified, the authorization request will be sent to that path,<br />or else the authorization request will use the path of the original request.<br />Please note that the original request path will be appended to the path specified here.<br />For example, if the original request path is "/hello", and the path specified here is "/auth",<br />then the path of the authorization request will be "/auth/hello". If the path is not specified,<br />the path of the authorization request will be "/hello". |
| `headersToBackend` | _string array_ |  false  |  | HeadersToBackend are the authorization response headers that will be added<br />to the original client request before sending it to the backend server.<br />Note that coexisting headers will be overridden.<br />If not specified, no authorization response headers will be added to the<br />original client request. |


#### HTTPHostnameModifier





_Appears in:_
- [HTTPURLRewriteFilter](#httpurlrewritefilter)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `type` | _[HTTPHostnameModifierType](#httphostnamemodifiertype)_ |  true  |  |  |
| `header` | _string_ |  false  |  | Header is the name of the header whose value would be used to rewrite the Host header |


#### HTTPHostnameModifierType

_Underlying type:_ _string_

HTTPPathModifierType defines the type of Hostname rewrite.

_Appears in:_
- [HTTPHostnameModifier](#httphostnamemodifier)

| Value | Description |
| ----- | ----------- |
| `Header` | HeaderHTTPHostnameModifier indicates that the Host header value would be replaced with the value of the header specified in header.<br />https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/route/v3/route_components.proto#envoy-v3-api-field-config-route-v3-routeaction-host-rewrite-header<br /> | 
| `Backend` | BackendHTTPHostnameModifier indicates that the Host header value would be replaced by the DNS name of the backend if it exists.<br />https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/route/v3/route_components.proto#envoy-v3-api-field-config-route-v3-routeaction-auto-host-rewrite<br /> | 


#### HTTPPathModifier





_Appears in:_
- [HTTPURLRewriteFilter](#httpurlrewritefilter)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `type` | _[HTTPPathModifierType](#httppathmodifiertype)_ |  true  |  |  |
| `replaceRegexMatch` | _[ReplaceRegexMatch](#replaceregexmatch)_ |  false  |  | ReplaceRegexMatch defines a path regex rewrite. The path portions matched by the regex pattern are replaced by the defined substitution.<br />https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/route/v3/route_components.proto#envoy-v3-api-field-config-route-v3-routeaction-regex-rewrite<br />Some examples:<br />(1) replaceRegexMatch:<br />      pattern: ^/service/([^/]+)(/.*)$<br />      substitution: \2/instance/\1<br />    Would transform /service/foo/v1/api into /v1/api/instance/foo.<br />(2) replaceRegexMatch:<br />      pattern: one<br />      substitution: two<br />    Would transform /xxx/one/yyy/one/zzz into /xxx/two/yyy/two/zzz.<br />(3) replaceRegexMatch:<br />      pattern: ^(.*?)one(.*)$<br />      substitution: \1two\2<br />    Would transform /xxx/one/yyy/one/zzz into /xxx/two/yyy/one/zzz.<br />(3) replaceRegexMatch:<br />      pattern: (?i)/xxx/<br />      substitution: /yyy/<br />    Would transform path /aaa/XxX/bbb into /aaa/yyy/bbb (case-insensitive). |


#### HTTPPathModifierType

_Underlying type:_ _string_

HTTPPathModifierType defines the type of path redirect or rewrite.

_Appears in:_
- [HTTPPathModifier](#httppathmodifier)

| Value | Description |
| ----- | ----------- |
| `ReplaceRegexMatch` | RegexHTTPPathModifier This type of modifier indicates that the portions of the path that match the specified<br /> regex would be substituted with the specified substitution value<br />https://www.envoyproxy.io/docs/envoy/latest/api-v3/type/matcher/v3/regex.proto#type-matcher-v3-regexmatchandsubstitute<br /> | 


#### HTTPRouteFilter



HTTPRouteFilter is a custom Envoy Gateway HTTPRouteFilter which provides extended
traffic processing options such as path regex rewrite, direct response and more.



| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `apiVersion` | _string_ | |`gateway.envoyproxy.io/v1alpha1`
| `kind` | _string_ | |`HTTPRouteFilter`
| `metadata` | _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#objectmeta-v1-meta)_ |  true  |  | Refer to Kubernetes API documentation for fields of `metadata`. |
| `spec` | _[HTTPRouteFilterSpec](#httproutefilterspec)_ |  true  |  | Spec defines the desired state of HTTPRouteFilter. |


#### HTTPRouteFilterSpec



HTTPRouteFilterSpec defines the desired state of HTTPRouteFilter.

_Appears in:_
- [HTTPRouteFilter](#httproutefilter)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `urlRewrite` | _[HTTPURLRewriteFilter](#httpurlrewritefilter)_ |  false  |  |  |
| `directResponse` | _[HTTPDirectResponseFilter](#httpdirectresponsefilter)_ |  false  |  |  |
| `credentialInjection` | _[HTTPCredentialInjectionFilter](#httpcredentialinjectionfilter)_ |  false  |  |  |


#### HTTPStatus

_Underlying type:_ _integer_

HTTPStatus defines the http status code.

_Appears in:_
- [HTTPActiveHealthChecker](#httpactivehealthchecker)
- [RetryOn](#retryon)



#### HTTPTimeout





_Appears in:_
- [Timeout](#timeout)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `connectionIdleTimeout` | _[Duration](https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1.Duration)_ |  false  |  | The idle timeout for an HTTP connection. Idle time is defined as a period in which there are no active requests in the connection.<br />Default: 1 hour. |
| `maxConnectionDuration` | _[Duration](https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1.Duration)_ |  false  |  | The maximum duration of an HTTP connection.<br />Default: unlimited. |
| `requestTimeout` | _[Duration](https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1.Duration)_ |  false  |  | RequestTimeout is the time until which entire response is received from the upstream. |


#### HTTPURLRewriteFilter



HTTPURLRewriteFilter define rewrites of HTTP URL components such as path and host

_Appears in:_
- [HTTPRouteFilterSpec](#httproutefilterspec)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `hostname` | _[HTTPHostnameModifier](#httphostnamemodifier)_ |  false  |  | Hostname is the value to be used to replace the Host header value during<br />forwarding. |
| `path` | _[HTTPPathModifier](#httppathmodifier)_ |  false  |  | Path defines a path rewrite. |


#### HTTPWasmCodeSource



HTTPWasmCodeSource defines the HTTP URL containing the Wasm code.

_Appears in:_
- [WasmCodeSource](#wasmcodesource)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `url` | _string_ |  true  |  | URL is the URL containing the Wasm code. |
| `sha256` | _string_ |  false  |  | SHA256 checksum that will be used to verify the Wasm code.<br />If not specified, Envoy Gateway will not verify the downloaded Wasm code.<br />kubebuilder:validation:Pattern=`^[a-f0-9]\{64\}$` |


#### Header



Header defines the header hashing configuration for consistent hash based
load balancing.

_Appears in:_
- [ConsistentHash](#consistenthash)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `name` | _string_ |  true  |  | Name of the header to hash. |


#### HeaderMatch



HeaderMatch defines the match attributes within the HTTP Headers of the request.

_Appears in:_
- [RateLimitSelectCondition](#ratelimitselectcondition)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `type` | _[HeaderMatchType](#headermatchtype)_ |  false  | Exact | Type specifies how to match against the value of the header. |
| `name` | _string_ |  true  |  | Name of the HTTP header.<br />The header name is case-insensitive unless PreserveHeaderCase is set to true.<br />For example, "Foo" and "foo" are considered the same header. |
| `value` | _string_ |  false  |  | Value within the HTTP header.<br />Do not set this field when Type="Distinct", implying matching on any/all unique<br />values within the header. |
| `invert` | _boolean_ |  false  | false | Invert specifies whether the value match result will be inverted.<br />Do not set this field when Type="Distinct", implying matching on any/all unique<br />values within the header. |


#### HeaderMatchType

_Underlying type:_ _string_

HeaderMatchType specifies the semantics of how HTTP header values should be compared.
Valid HeaderMatchType values are "Exact", "RegularExpression", and "Distinct".

_Appears in:_
- [HeaderMatch](#headermatch)

| Value | Description |
| ----- | ----------- |
| `Exact` | HeaderMatchExact matches the exact value of the Value field against the value of<br />the specified HTTP Header.<br /> | 
| `RegularExpression` | HeaderMatchRegularExpression matches a regular expression against the value of the<br />specified HTTP Header. The regex string must adhere to the syntax documented in<br />https://github.com/google/re2/wiki/Syntax.<br /> | 
| `Distinct` | HeaderMatchDistinct matches any and all possible unique values encountered in the<br />specified HTTP Header. Note that each unique value will receive its own rate limit<br />bucket.<br /> | 


#### HeaderSettings



HeaderSettings provides configuration options for headers on the listener.

_Appears in:_
- [ClientTrafficPolicySpec](#clienttrafficpolicyspec)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `enableEnvoyHeaders` | _boolean_ |  false  |  | EnableEnvoyHeaders configures Envoy Proxy to add the "X-Envoy-" headers to requests<br />and responses. |
| `disableRateLimitHeaders` | _boolean_ |  false  |  | DisableRateLimitHeaders configures Envoy Proxy to omit the "X-RateLimit-" response headers<br />when rate limiting is enabled. |
| `xForwardedClientCert` | _[XForwardedClientCert](#xforwardedclientcert)_ |  false  |  | XForwardedClientCert configures how Envoy Proxy handle the x-forwarded-client-cert (XFCC) HTTP header.<br />x-forwarded-client-cert (XFCC) is an HTTP header used to forward the certificate<br />information of part or all of the clients or proxies that a request has flowed through,<br />on its way from the client to the server.<br />Envoy proxy may choose to sanitize/append/forward the XFCC header before proxying the request.<br />If not set, the default behavior is sanitizing the XFCC header. |
| `withUnderscoresAction` | _[WithUnderscoresAction](#withunderscoresaction)_ |  false  |  | WithUnderscoresAction configures the action to take when an HTTP header with underscores<br />is encountered. The default action is to reject the request. |
| `preserveXRequestID` | _boolean_ |  false  |  | PreserveXRequestID configures Envoy to keep the X-Request-ID header if passed for a request that is edge<br />(Edge request is the request from external clients to front Envoy) and not reset it, which is the current Envoy behaviour.<br />Defaults to false and cannot be combined with RequestID.<br />Deprecated: use RequestID=Preserve instead |
| `requestID` | _[RequestIDAction](#requestidaction)_ |  false  |  | RequestID configures Envoy's behavior for handling the `X-Request-ID` header.<br />Defaults to `Generate` and builds the `X-Request-ID` for every request and ignores pre-existing values from the edge.<br />(An "edge request" refers to a request from an external client to the Envoy entrypoint.) |
| `earlyRequestHeaders` | _[HTTPHeaderFilter](#httpheaderfilter)_ |  false  |  | EarlyRequestHeaders defines settings for early request header modification, before envoy performs<br />routing, tracing and built-in header manipulation. |


#### HealthCheck



HealthCheck configuration to decide which endpoints
are healthy and can be used for routing.

_Appears in:_
- [BackendTrafficPolicySpec](#backendtrafficpolicyspec)
- [ClusterSettings](#clustersettings)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `active` | _[ActiveHealthCheck](#activehealthcheck)_ |  false  |  | Active health check configuration |
| `passive` | _[PassiveHealthCheck](#passivehealthcheck)_ |  false  |  | Passive passive check configuration |
| `panicThreshold` | _integer_ |  false  |  | When number of unhealthy endpoints for a backend reaches this threshold<br />Envoy will disregard health status and balance across all endpoints.<br />It's designed to prevent a situation in which host failures cascade throughout the cluster<br />as load increases. If not set, the default value is 50%. To disable panic mode, set value to `0`. |


#### HealthCheckSettings



HealthCheckSettings provides HealthCheck configuration on the HTTP/HTTPS listener.

_Appears in:_
- [ClientTrafficPolicySpec](#clienttrafficpolicyspec)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `path` | _string_ |  true  |  | Path specifies the HTTP path to match on for health check requests. |


#### IPEndpoint



IPEndpoint describes TCP/UDP socket address, corresponding to Envoy's Socket Address
https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/core/v3/address.proto#config-core-v3-socketaddress

_Appears in:_
- [BackendEndpoint](#backendendpoint)
- [ExtensionService](#extensionservice)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `address` | _string_ |  true  |  | Address defines the IP address of the backend endpoint.<br />Supports both IPv4 and IPv6 addresses. |
| `port` | _integer_ |  true  |  | Port defines the port of the backend endpoint. |


#### IPFamily

_Underlying type:_ _string_

IPFamily defines the IP family to use for the Envoy proxy.

_Appears in:_
- [EnvoyProxySpec](#envoyproxyspec)

| Value | Description |
| ----- | ----------- |
| `IPv4` | IPv4 defines the IPv4 family.<br /> | 
| `IPv6` | IPv6 defines the IPv6 family.<br /> | 
| `DualStack` | DualStack defines the dual-stack family.<br />When set to DualStack, Envoy proxy will listen on both IPv4 and IPv6 addresses<br />for incoming client traffic, enabling support for both IP protocol versions.<br /> | 


#### ImagePullPolicy

_Underlying type:_ _string_

ImagePullPolicy defines the policy to use when pulling an OIC image.

_Appears in:_
- [WasmCodeSource](#wasmcodesource)

| Value | Description |
| ----- | ----------- |
| `IfNotPresent` | ImagePullPolicyIfNotPresent will only pull the image if it does not already exist in the EG cache.<br /> | 
| `Always` | ImagePullPolicyAlways will pull the image when the EnvoyExtension resource version changes.<br />Note: EG does not update the Wasm module every time an Envoy proxy requests the Wasm module.<br /> | 


#### ImageWasmCodeSource



ImageWasmCodeSource defines the OCI image containing the Wasm code.

_Appears in:_
- [WasmCodeSource](#wasmcodesource)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `url` | _string_ |  true  |  | URL is the URL of the OCI image.<br />URL can be in the format of `registry/image:tag` or `registry/image@sha256:digest`. |
| `sha256` | _string_ |  false  |  | SHA256 checksum that will be used to verify the OCI image.<br />It must match the digest of the OCI image.<br />If not specified, Envoy Gateway will not verify the downloaded OCI image.<br />kubebuilder:validation:Pattern=`^[a-f0-9]\{64\}$` |
| `pullSecretRef` | _[SecretObjectReference](https://gateway-api.sigs.k8s.io/references/spec/#gateway.networking.k8s.io/v1.SecretObjectReference)_ |  false  |  | PullSecretRef is a reference to the secret containing the credentials to pull the image.<br />Only support Kubernetes Secret resource from the same namespace. |


#### InfrastructureProviderType

_Underlying type:_ _string_

InfrastructureProviderType defines the types of custom infrastructure providers supported by Envoy Gateway.

_Appears in:_
- [EnvoyGatewayInfrastructureProvider](#envoygatewayinfrastructureprovider)

| Value | Description |
| ----- | ----------- |
| `Host` | InfrastructureProviderTypeHost defines the "Host" provider.<br /> | 


#### InjectedCredential



InjectedCredential defines the credential to be injected.

_Appears in:_
- [HTTPCredentialInjectionFilter](#httpcredentialinjectionfilter)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `valueRef` | _[SecretObjectReference](https://gateway-api.sigs.k8s.io/references/spec/#gateway.networking.k8s.io/v1.SecretObjectReference)_ |  true  |  | ValueRef is a reference to the secret containing the credentials to be injected.<br />This is an Opaque secret. The credential should be stored in the key<br />"credential", and the value should be the credential to be injected.<br />For example, for basic authentication, the value should be "Basic <base64 encoded username:password>".<br />for bearer token, the value should be "Bearer <token>".<br />Note: The secret must be in the same namespace as the HTTPRouteFilter. |


#### InvalidMessageAction

_Underlying type:_ _string_



_Appears in:_
- [HTTP2Settings](#http2settings)

| Value | Description |
| ----- | ----------- |
| `TerminateConnection` |  | 
| `TerminateStream` |  | 


#### JSONPatchOperation



JSONPatchOperation defines the JSON Patch Operation as defined in
https://datatracker.ietf.org/doc/html/rfc6902

_Appears in:_
- [EnvoyJSONPatchConfig](#envoyjsonpatchconfig)
- [ProxyBootstrap](#proxybootstrap)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `op` | _[JSONPatchOperationType](#jsonpatchoperationtype)_ |  true  |  | Op is the type of operation to perform |
| `path` | _string_ |  false  |  | Path is a JSONPointer expression. Refer to https://datatracker.ietf.org/doc/html/rfc6901 for more details.<br />It specifies the location of the target document/field where the operation will be performed |
| `jsonPath` | _string_ |  false  |  | JSONPath is a JSONPath expression. Refer to https://datatracker.ietf.org/doc/rfc9535/ for more details.<br />It produces one or more JSONPointer expressions based on the given JSON document.<br />If no JSONPointer is found, it will result in an error.<br />If the 'Path' property is also set, it will be appended to the resulting JSONPointer expressions from the JSONPath evaluation.<br />This is useful when creating a property that does not yet exist in the JSON document.<br />The final JSONPointer expressions specifies the locations in the target document/field where the operation will be applied. |
| `from` | _string_ |  false  |  | From is the source location of the value to be copied or moved. Only valid<br />for move or copy operations<br />Refer to https://datatracker.ietf.org/doc/html/rfc6901 for more details. |
| `value` | _[JSON](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#json-v1-apiextensions-k8s-io)_ |  false  |  | Value is the new value of the path location. The value is only used by<br />the `add` and `replace` operations. |


#### JSONPatchOperationType

_Underlying type:_ _string_

JSONPatchOperationType specifies the JSON Patch operations that can be performed.

_Appears in:_
- [JSONPatchOperation](#jsonpatchoperation)



#### JWT



JWT defines the configuration for JSON Web Token (JWT) authentication.

_Appears in:_
- [SecurityPolicySpec](#securitypolicyspec)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `optional` | _boolean_ |  true  |  | Optional determines whether a missing JWT is acceptable, defaulting to false if not specified.<br />Note: Even if optional is set to true, JWT authentication will still fail if an invalid JWT is presented. |
| `providers` | _[JWTProvider](#jwtprovider) array_ |  true  |  | Providers defines the JSON Web Token (JWT) authentication provider type.<br />When multiple JWT providers are specified, the JWT is considered valid if<br />any of the providers successfully validate the JWT. For additional details,<br />see https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/jwt_authn_filter.html. |


#### JWTClaim



JWTClaim specifies a claim in a JWT token.

_Appears in:_
- [JWTPrincipal](#jwtprincipal)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `name` | _string_ |  true  |  | Name is the name of the claim.<br />If it is a nested claim, use a dot (.) separated string as the name to<br />represent the full path to the claim.<br />For example, if the claim is in the "department" field in the "organization" field,<br />the name should be "organization.department". |
| `valueType` | _[JWTClaimValueType](#jwtclaimvaluetype)_ |  false  | String | ValueType is the type of the claim value.<br />Only String and StringArray types are supported for now. |
| `values` | _string array_ |  true  |  | Values are the values that the claim must match.<br />If the claim is a string type, the specified value must match exactly.<br />If the claim is a string array type, the specified value must match one of the values in the array.<br />If multiple values are specified, one of the values must match for the rule to match. |


#### JWTClaimValueType

_Underlying type:_ _string_



_Appears in:_
- [JWTClaim](#jwtclaim)

| Value | Description |
| ----- | ----------- |
| `String` |  | 
| `StringArray` |  | 


#### JWTExtractor



JWTExtractor defines a custom JWT token extraction from HTTP request.
If specified, Envoy will extract the JWT token from the listed extractors (headers, cookies, or params) and validate each of them.
If any value extracted is found to be an invalid JWT, a 401 error will be returned.

_Appears in:_
- [JWTProvider](#jwtprovider)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `headers` | _[JWTHeaderExtractor](#jwtheaderextractor) array_ |  false  |  | Headers represents a list of HTTP request headers to extract the JWT token from. |
| `cookies` | _string array_ |  false  |  | Cookies represents a list of cookie names to extract the JWT token from. |
| `params` | _string array_ |  false  |  | Params represents a list of query parameters to extract the JWT token from. |


#### JWTHeaderExtractor



JWTHeaderExtractor defines an HTTP header location to extract JWT token

_Appears in:_
- [JWTExtractor](#jwtextractor)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `name` | _string_ |  true  |  | Name is the HTTP header name to retrieve the token |
| `valuePrefix` | _string_ |  false  |  | ValuePrefix is the prefix that should be stripped before extracting the token.<br />The format would be used by Envoy like "\{ValuePrefix\}<TOKEN>".<br />For example, "Authorization: Bearer <TOKEN>", then the ValuePrefix="Bearer " with a space at the end. |


#### JWTPrincipal



JWTPrincipal specifies the client identity of a request based on the JWT claims and scopes.
At least one of the claims or scopes must be specified.
Claims and scopes are And-ed together if both are specified.

_Appears in:_
- [Principal](#principal)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `provider` | _string_ |  true  |  | Provider is the name of the JWT provider that used to verify the JWT token.<br />In order to use JWT claims for authorization, you must configure the JWT<br />authentication with the same provider in the same `SecurityPolicy`. |
| `claims` | _[JWTClaim](#jwtclaim) array_ |  false  |  | Claims are the claims in a JWT token.<br />If multiple claims are specified, all claims must match for the rule to match.<br />For example, if there are two claims: one for the audience and one for the issuer,<br />the rule will match only if both the audience and the issuer match. |
| `scopes` | _[JWTScope](#jwtscope) array_ |  false  |  | Scopes are a special type of claim in a JWT token that represents the permissions of the client.<br />The value of the scopes field should be a space delimited string that is expected in the scope parameter,<br />as defined in RFC 6749: https://datatracker.ietf.org/doc/html/rfc6749#page-23.<br />If multiple scopes are specified, all scopes must match for the rule to match. |


#### JWTProvider



JWTProvider defines how a JSON Web Token (JWT) can be verified.

_Appears in:_
- [JWT](#jwt)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `name` | _string_ |  true  |  | Name defines a unique name for the JWT provider. A name can have a variety of forms,<br />including RFC1123 subdomains, RFC 1123 labels, or RFC 1035 labels. |
| `issuer` | _string_ |  false  |  | Issuer is the principal that issued the JWT and takes the form of a URL or email address.<br />For additional details, see https://tools.ietf.org/html/rfc7519#section-4.1.1 for<br />URL format and https://rfc-editor.org/rfc/rfc5322.html for email format. If not provided,<br />the JWT issuer is not checked. |
| `audiences` | _string array_ |  false  |  | Audiences is a list of JWT audiences allowed access. For additional details, see<br />https://tools.ietf.org/html/rfc7519#section-4.1.3. If not provided, JWT audiences<br />are not checked. |
| `remoteJWKS` | _[RemoteJWKS](#remotejwks)_ |  false  |  | RemoteJWKS defines how to fetch and cache JSON Web Key Sets (JWKS) from a remote<br />HTTP/HTTPS endpoint. |
| `localJWKS` | _[LocalJWKS](#localjwks)_ |  false  |  | LocalJWKS defines how to get the JSON Web Key Sets (JWKS) from a local source. |
| `claimToHeaders` | _[ClaimToHeader](#claimtoheader) array_ |  false  |  | ClaimToHeaders is a list of JWT claims that must be extracted into HTTP request headers<br />For examples, following config:<br />The claim must be of type; string, int, double, bool. Array type claims are not supported |
| `recomputeRoute` | _boolean_ |  false  |  | RecomputeRoute clears the route cache and recalculates the routing decision.<br />This field must be enabled if the headers generated from the claim are used for<br />route matching decisions. If the recomputation selects a new route, features targeting<br />the new matched route will be applied. |
| `extractFrom` | _[JWTExtractor](#jwtextractor)_ |  false  |  | ExtractFrom defines different ways to extract the JWT token from HTTP request.<br />If empty, it defaults to extract JWT token from the Authorization HTTP request header using Bearer schema<br />or access_token from query parameters. |


#### JWTScope

_Underlying type:_ _string_



_Appears in:_
- [JWTPrincipal](#jwtprincipal)



#### KubernetesClient





_Appears in:_
- [EnvoyGatewayKubernetesProvider](#envoygatewaykubernetesprovider)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `rateLimit` | _[KubernetesClientRateLimit](#kubernetesclientratelimit)_ |  true  |  | RateLimit defines the rate limit settings for the Kubernetes client. |


#### KubernetesClientRateLimit



KubernetesClientRateLimit defines the rate limit settings for the Kubernetes client.

_Appears in:_
- [KubernetesClient](#kubernetesclient)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `qps` | _integer_ |  false  | 50 | QPS defines the queries per second limit for the Kubernetes client. |
| `burst` | _integer_ |  false  | 100 | Burst defines the maximum burst of requests allowed when tokens have accumulated. |


#### KubernetesContainerSpec



KubernetesContainerSpec defines the desired state of the Kubernetes container resource.

_Appears in:_
- [KubernetesDaemonSetSpec](#kubernetesdaemonsetspec)
- [KubernetesDeploymentSpec](#kubernetesdeploymentspec)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `env` | _[EnvVar](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#envvar-v1-core) array_ |  false  |  | List of environment variables to set in the container. |
| `resources` | _[ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#resourcerequirements-v1-core)_ |  false  |  | Resources required by this container.<br />More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/ |
| `securityContext` | _[SecurityContext](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#securitycontext-v1-core)_ |  false  |  | SecurityContext defines the security options the container should be run with.<br />If set, the fields of SecurityContext override the equivalent fields of PodSecurityContext.<br />More info: https://kubernetes.io/docs/tasks/configure-pod-container/security-context/ |
| `image` | _string_ |  false  |  | Image specifies the EnvoyProxy container image to be used, instead of the default image. |
| `volumeMounts` | _[VolumeMount](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#volumemount-v1-core) array_ |  false  |  | VolumeMounts are volumes to mount into the container's filesystem.<br />Cannot be updated. |


#### KubernetesDaemonSetSpec



KubernetesDaemonSetSpec defines the desired state of the Kubernetes daemonset resource.

_Appears in:_
- [EnvoyProxyKubernetesProvider](#envoyproxykubernetesprovider)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `patch` | _[KubernetesPatchSpec](#kubernetespatchspec)_ |  false  |  | Patch defines how to perform the patch operation to daemonset |
| `strategy` | _[DaemonSetUpdateStrategy](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#daemonsetupdatestrategy-v1-apps)_ |  false  |  | The daemonset strategy to use to replace existing pods with new ones. |
| `pod` | _[KubernetesPodSpec](#kubernetespodspec)_ |  false  |  | Pod defines the desired specification of pod. |
| `container` | _[KubernetesContainerSpec](#kubernetescontainerspec)_ |  false  |  | Container defines the desired specification of main container. |
| `name` | _string_ |  false  |  | Name of the daemonSet.<br />When unset, this defaults to an autogenerated name. |


#### KubernetesDeployMode



KubernetesDeployMode holds configuration for how to deploy managed resources such as the Envoy Proxy
data plane fleet.

_Appears in:_
- [EnvoyGatewayKubernetesProvider](#envoygatewaykubernetesprovider)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `type` | _[KubernetesDeployModeType](#kubernetesdeploymodetype)_ |  false  | ControllerNamespace | Type indicates what deployment mode to use. "ControllerNamespace" and<br />"GatewayNamespace" are currently supported.<br />By default, when this field is unset or empty, Envoy Gateway will deploy Envoy Proxy fleet in the Controller namespace. |


#### KubernetesDeployModeType

_Underlying type:_ _string_

KubernetesDeployModeType defines the type of KubernetesDeployMode

_Appears in:_
- [KubernetesDeployMode](#kubernetesdeploymode)

| Value | Description |
| ----- | ----------- |
| `ControllerNamespace` | KubernetesDeployModeTypeControllerNamespace indicates that the controller namespace is used for the infra proxy deployments.<br /> | 
| `GatewayNamespace` | KubernetesDeployModeTypeGatewayNamespace indicates that the gateway namespace is used for the infra proxy deployments.<br /> | 


#### KubernetesDeploymentSpec



KubernetesDeploymentSpec defines the desired state of the Kubernetes deployment resource.

_Appears in:_
- [EnvoyGatewayKubernetesProvider](#envoygatewaykubernetesprovider)
- [EnvoyProxyKubernetesProvider](#envoyproxykubernetesprovider)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `patch` | _[KubernetesPatchSpec](#kubernetespatchspec)_ |  false  |  | Patch defines how to perform the patch operation to deployment |
| `replicas` | _integer_ |  false  |  | Replicas is the number of desired pods. Defaults to 1. |
| `strategy` | _[DeploymentStrategy](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#deploymentstrategy-v1-apps)_ |  false  |  | The deployment strategy to use to replace existing pods with new ones. |
| `pod` | _[KubernetesPodSpec](#kubernetespodspec)_ |  false  |  | Pod defines the desired specification of pod. |
| `container` | _[KubernetesContainerSpec](#kubernetescontainerspec)_ |  false  |  | Container defines the desired specification of main container. |
| `initContainers` | _[Container](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#container-v1-core) array_ |  false  |  | List of initialization containers belonging to the pod.<br />More info: https://kubernetes.io/docs/concepts/workloads/pods/init-containers/ |
| `name` | _string_ |  false  |  | Name of the deployment.<br />When unset, this defaults to an autogenerated name. |


#### KubernetesHorizontalPodAutoscalerSpec



KubernetesHorizontalPodAutoscalerSpec defines Kubernetes Horizontal Pod Autoscaler settings of Envoy Proxy Deployment.
When HPA is enabled, it is recommended that the value in `KubernetesDeploymentSpec.replicas` be removed, otherwise
Envoy Gateway will revert back to this value every time reconciliation occurs.
See k8s.io.autoscaling.v2.HorizontalPodAutoScalerSpec.

_Appears in:_
- [EnvoyGatewayKubernetesProvider](#envoygatewaykubernetesprovider)
- [EnvoyProxyKubernetesProvider](#envoyproxykubernetesprovider)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `minReplicas` | _integer_ |  false  |  | minReplicas is the lower limit for the number of replicas to which the autoscaler<br />can scale down. It defaults to 1 replica. |
| `maxReplicas` | _integer_ |  true  |  | maxReplicas is the upper limit for the number of replicas to which the autoscaler can scale up.<br />It cannot be less that minReplicas. |
| `metrics` | _[MetricSpec](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#metricspec-v2-autoscaling) array_ |  false  |  | metrics contains the specifications for which to use to calculate the<br />desired replica count (the maximum replica count across all metrics will<br />be used).<br />If left empty, it defaults to being based on CPU utilization with average on 80% usage. |
| `behavior` | _[HorizontalPodAutoscalerBehavior](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#horizontalpodautoscalerbehavior-v2-autoscaling)_ |  false  |  | behavior configures the scaling behavior of the target<br />in both Up and Down directions (scaleUp and scaleDown fields respectively).<br />If not set, the default HPAScalingRules for scale up and scale down are used.<br />See k8s.io.autoscaling.v2.HorizontalPodAutoScalerBehavior. |
| `patch` | _[KubernetesPatchSpec](#kubernetespatchspec)_ |  false  |  | Patch defines how to perform the patch operation to the HorizontalPodAutoscaler |
| `name` | _string_ |  false  |  | Name of the horizontalPodAutoScaler.<br />When unset, this defaults to an autogenerated name. |


#### KubernetesPatchSpec



KubernetesPatchSpec defines how to perform the patch operation.
Note that `value` can be an in-line YAML document, as can be seen in e.g. (the example of patching the Envoy proxy Deployment)[https://gateway.envoyproxy.io/docs/tasks/operations/customize-envoyproxy/#patching-deployment-for-envoyproxy].
Note also that, currently, strings containing literal JSON are _rejected_.

_Appears in:_
- [KubernetesDaemonSetSpec](#kubernetesdaemonsetspec)
- [KubernetesDeploymentSpec](#kubernetesdeploymentspec)
- [KubernetesHorizontalPodAutoscalerSpec](#kuberneteshorizontalpodautoscalerspec)
- [KubernetesPodDisruptionBudgetSpec](#kubernetespoddisruptionbudgetspec)
- [KubernetesServiceSpec](#kubernetesservicespec)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `type` | _[MergeType](#mergetype)_ |  false  |  | Type is the type of merge operation to perform<br />By default, StrategicMerge is used as the patch type. |
| `value` | _[JSON](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#json-v1-apiextensions-k8s-io)_ |  true  |  | Object contains the raw configuration for merged object |


#### KubernetesPodDisruptionBudgetSpec



KubernetesPodDisruptionBudgetSpec defines Kubernetes PodDisruptionBudget settings of Envoy Proxy Deployment.

_Appears in:_
- [EnvoyProxyKubernetesProvider](#envoyproxykubernetesprovider)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `minAvailable` | _[IntOrString](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#intorstring-intstr-util)_ |  false  |  | MinAvailable specifies the minimum amount of pods (can be expressed as integers or as a percentage) that must be available at all times during voluntary disruptions,<br />such as node drains or updates. This setting ensures that your envoy proxy maintains a certain level of availability<br />and resilience during maintenance operations. Cannot be combined with maxUnavailable. |
| `maxUnavailable` | _[IntOrString](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#intorstring-intstr-util)_ |  false  |  | MaxUnavailable specifies the maximum amount of pods (can be expressed as integers or as a percentage) that can be unavailable at all times during voluntary disruptions,<br />such as node drains or updates. This setting ensures that your envoy proxy maintains a certain level of availability<br />and resilience during maintenance operations. Cannot be combined with minAvailable. |
| `patch` | _[KubernetesPatchSpec](#kubernetespatchspec)_ |  false  |  | Patch defines how to perform the patch operation to the PodDisruptionBudget |
| `name` | _string_ |  false  |  | Name of the podDisruptionBudget.<br />When unset, this defaults to an autogenerated name. |


#### KubernetesPodSpec



KubernetesPodSpec defines the desired state of the Kubernetes pod resource.

_Appears in:_
- [KubernetesDaemonSetSpec](#kubernetesdaemonsetspec)
- [KubernetesDeploymentSpec](#kubernetesdeploymentspec)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `annotations` | _object (keys:string, values:string)_ |  false  |  | Annotations are the annotations that should be appended to the pods.<br />By default, no pod annotations are appended. |
| `labels` | _object (keys:string, values:string)_ |  false  |  | Labels are the additional labels that should be tagged to the pods.<br />By default, no additional pod labels are tagged. |
| `securityContext` | _[PodSecurityContext](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#podsecuritycontext-v1-core)_ |  false  |  | SecurityContext holds pod-level security attributes and common container settings.<br />Optional: Defaults to empty.  See type description for default values of each field. |
| `affinity` | _[Affinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#affinity-v1-core)_ |  false  |  | If specified, the pod's scheduling constraints. |
| `tolerations` | _[Toleration](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#toleration-v1-core) array_ |  false  |  | If specified, the pod's tolerations. |
| `volumes` | _[Volume](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#volume-v1-core) array_ |  false  |  | Volumes that can be mounted by containers belonging to the pod.<br />More info: https://kubernetes.io/docs/concepts/storage/volumes |
| `imagePullSecrets` | _[LocalObjectReference](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#localobjectreference-v1-core) array_ |  false  |  | ImagePullSecrets is an optional list of references to secrets<br />in the same namespace to use for pulling any of the images used by this PodSpec.<br />If specified, these secrets will be passed to individual puller implementations for them to use.<br />More info: https://kubernetes.io/docs/concepts/containers/images#specifying-imagepullsecrets-on-a-pod |
| `nodeSelector` | _object (keys:string, values:string)_ |  false  |  | NodeSelector is a selector which must be true for the pod to fit on a node.<br />Selector which must match a node's labels for the pod to be scheduled on that node.<br />More info: https://kubernetes.io/docs/concepts/configuration/assign-pod-node/ |
| `topologySpreadConstraints` | _[TopologySpreadConstraint](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#topologyspreadconstraint-v1-core) array_ |  false  |  | TopologySpreadConstraints describes how a group of pods ought to spread across topology<br />domains. Scheduler will schedule pods in a way which abides by the constraints.<br />All topologySpreadConstraints are ANDed. |


#### KubernetesServiceSpec



KubernetesServiceSpec defines the desired state of the Kubernetes service resource.

_Appears in:_
- [EnvoyProxyKubernetesProvider](#envoyproxykubernetesprovider)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `annotations` | _object (keys:string, values:string)_ |  false  |  | Annotations that should be appended to the service.<br />By default, no annotations are appended. |
| `labels` | _object (keys:string, values:string)_ |  false  |  | Labels that should be appended to the service.<br />By default, no labels are appended. |
| `type` | _[ServiceType](#servicetype)_ |  false  | LoadBalancer | Type determines how the Service is exposed. Defaults to LoadBalancer.<br />Valid options are ClusterIP, LoadBalancer and NodePort.<br />"LoadBalancer" means a service will be exposed via an external load balancer (if the cloud provider supports it).<br />"ClusterIP" means a service will only be accessible inside the cluster, via the cluster IP.<br />"NodePort" means a service will be exposed on a static Port on all Nodes of the cluster. |
| `loadBalancerClass` | _string_ |  false  |  | LoadBalancerClass, when specified, allows for choosing the LoadBalancer provider<br />implementation if more than one are available or is otherwise expected to be specified |
| `allocateLoadBalancerNodePorts` | _boolean_ |  false  |  | AllocateLoadBalancerNodePorts defines if NodePorts will be automatically allocated for<br />services with type LoadBalancer. Default is "true". It may be set to "false" if the cluster<br />load-balancer does not rely on NodePorts. If the caller requests specific NodePorts (by specifying a<br />value), those requests will be respected, regardless of this field. This field may only be set for<br />services with type LoadBalancer and will be cleared if the type is changed to any other type. |
| `loadBalancerSourceRanges` | _string array_ |  false  |  | LoadBalancerSourceRanges defines a list of allowed IP addresses which will be configured as<br />firewall rules on the platform providers load balancer. This is not guaranteed to be working as<br />it happens outside of kubernetes and has to be supported and handled by the platform provider.<br />This field may only be set for services with type LoadBalancer and will be cleared if the type<br />is changed to any other type. |
| `loadBalancerIP` | _string_ |  false  |  | LoadBalancerIP defines the IP Address of the underlying load balancer service. This field<br />may be ignored if the load balancer provider does not support this feature.<br />This field has been deprecated in Kubernetes, but it is still used for setting the IP Address in some cloud<br />providers such as GCP. |
| `externalTrafficPolicy` | _[ServiceExternalTrafficPolicy](#serviceexternaltrafficpolicy)_ |  false  | Local | ExternalTrafficPolicy determines the externalTrafficPolicy for the Envoy Service. Valid options<br />are Local and Cluster. Default is "Local". "Local" means traffic will only go to pods on the node<br />receiving the traffic. "Cluster" means connections are loadbalanced to all pods in the cluster. |
| `patch` | _[KubernetesPatchSpec](#kubernetespatchspec)_ |  false  |  | Patch defines how to perform the patch operation to the service |
| `name` | _string_ |  false  |  | Name of the service.<br />When unset, this defaults to an autogenerated name. |


#### KubernetesWatchMode



KubernetesWatchMode holds the configuration for which input resources to watch and reconcile.

_Appears in:_
- [EnvoyGatewayKubernetesProvider](#envoygatewaykubernetesprovider)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `type` | _[KubernetesWatchModeType](#kuberneteswatchmodetype)_ |  true  |  | Type indicates what watch mode to use. KubernetesWatchModeTypeNamespaces and<br />KubernetesWatchModeTypeNamespaceSelector are currently supported<br />By default, when this field is unset or empty, Envoy Gateway will watch for input namespaced resources<br />from all namespaces. |
| `namespaces` | _string array_ |  true  |  | Namespaces holds the list of namespaces that Envoy Gateway will watch for namespaced scoped<br />resources such as Gateway, HTTPRoute and Service.<br />Note that Envoy Gateway will continue to reconcile relevant cluster scoped resources such as<br />GatewayClass that it is linked to. Precisely one of Namespaces and NamespaceSelector must be set. |
| `namespaceSelector` | _[LabelSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#labelselector-v1-meta)_ |  true  |  | NamespaceSelector holds the label selector used to dynamically select namespaces.<br />Envoy Gateway will watch for namespaces matching the specified label selector.<br />Precisely one of Namespaces and NamespaceSelector must be set. |


#### KubernetesWatchModeType

_Underlying type:_ _string_

KubernetesWatchModeType defines the type of KubernetesWatchMode

_Appears in:_
- [KubernetesWatchMode](#kuberneteswatchmode)



#### LeaderElection



LeaderElection defines the desired leader election settings.

_Appears in:_
- [EnvoyGatewayKubernetesProvider](#envoygatewaykubernetesprovider)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `leaseDuration` | _[Duration](https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1.Duration)_ |  true  |  | LeaseDuration defines the time non-leader contenders will wait before attempting to claim leadership.<br />It's based on the timestamp of the last acknowledged signal. The default setting is 15 seconds. |
| `renewDeadline` | _[Duration](https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1.Duration)_ |  true  |  | RenewDeadline represents the time frame within which the current leader will attempt to renew its leadership<br />status before relinquishing its position. The default setting is 10 seconds. |
| `retryPeriod` | _[Duration](https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1.Duration)_ |  true  |  | RetryPeriod denotes the interval at which LeaderElector clients should perform action retries.<br />The default setting is 2 seconds. |
| `disable` | _boolean_ |  true  |  | Disable provides the option to turn off leader election, which is enabled by default. |


#### LiteralCustomTag



LiteralCustomTag adds hard-coded value to each span.

_Appears in:_
- [CustomTag](#customtag)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `value` | _string_ |  true  |  | Value defines the hard-coded value to add to each span. |


#### LoadBalancer



LoadBalancer defines the load balancer policy to be applied.

_Appears in:_
- [BackendTrafficPolicySpec](#backendtrafficpolicyspec)
- [ClusterSettings](#clustersettings)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `type` | _[LoadBalancerType](#loadbalancertype)_ |  true  |  | Type decides the type of Load Balancer policy.<br />Valid LoadBalancerType values are<br />"ConsistentHash",<br />"LeastRequest",<br />"Random",<br />"RoundRobin". |
| `consistentHash` | _[ConsistentHash](#consistenthash)_ |  false  |  | ConsistentHash defines the configuration when the load balancer type is<br />set to ConsistentHash |
| `slowStart` | _[SlowStart](#slowstart)_ |  false  |  | SlowStart defines the configuration related to the slow start load balancer policy.<br />If set, during slow start window, traffic sent to the newly added hosts will gradually increase.<br />Currently this is only supported for RoundRobin and LeastRequest load balancers |


#### LoadBalancerType

_Underlying type:_ _string_

LoadBalancerType specifies the types of LoadBalancer.

_Appears in:_
- [LoadBalancer](#loadbalancer)

| Value | Description |
| ----- | ----------- |
| `ConsistentHash` | ConsistentHashLoadBalancerType load balancer policy.<br /> | 
| `LeastRequest` | LeastRequestLoadBalancerType load balancer policy.<br /> | 
| `Random` | RandomLoadBalancerType load balancer policy.<br /> | 
| `RoundRobin` | RoundRobinLoadBalancerType load balancer policy.<br /> | 


#### LocalJWKS



LocalJWKS defines how to load a JSON Web Key Sets (JWKS) from a local source, either inline or from a reference to a ConfigMap.

_Appears in:_
- [JWTProvider](#jwtprovider)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `type` | _[LocalJWKSType](#localjwkstype)_ |  true  | Inline | Type is the type of method to use to read the body value.<br />Valid values are Inline and ValueRef, default is Inline. |
| `inline` | _string_ |  false  |  | Inline contains the value as an inline string. |
| `valueRef` | _[LocalObjectReference](#localobjectreference)_ |  false  |  | ValueRef is a reference to a local ConfigMap that contains the JSON Web Key Sets (JWKS).<br />The value of key `jwks` in the ConfigMap will be used.<br />If the key is not found, the first value in the ConfigMap will be used. |


#### LocalJWKSType

_Underlying type:_ _string_

LocalJWKSType defines the types of values for Local JWKS.

_Appears in:_
- [LocalJWKS](#localjwks)

| Value | Description |
| ----- | ----------- |
| `Inline` | LocalJWKSTypeInline defines the "Inline" LocalJWKS type.<br /> | 
| `ValueRef` | LocalJWKSTypeValueRef defines the "ValueRef" LocalJWKS type.<br /> | 


#### LocalRateLimit



LocalRateLimit defines local rate limit configuration.

_Appears in:_
- [RateLimitSpec](#ratelimitspec)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `rules` | _[RateLimitRule](#ratelimitrule) array_ |  false  |  | Rules are a list of RateLimit selectors and limits. If a request matches<br />multiple rules, the strictest limit is applied. For example, if a request<br />matches two rules, one with 10rps and one with 20rps, the final limit will<br />be based on the rule with 10rps. |


#### LogLevel

_Underlying type:_ _string_

LogLevel defines a log level for Envoy Gateway and EnvoyProxy system logs.

_Appears in:_
- [EnvoyGatewayLogging](#envoygatewaylogging)
- [ProxyLogging](#proxylogging)

| Value | Description |
| ----- | ----------- |
| `trace` | LogLevelTrace defines the "Trace" logging level.<br /> | 
| `debug` | LogLevelDebug defines the "debug" logging level.<br /> | 
| `info` | LogLevelInfo defines the "Info" logging level.<br /> | 
| `warn` | LogLevelWarn defines the "Warn" logging level.<br /> | 
| `error` | LogLevelError defines the "Error" logging level.<br /> | 


#### Lua



Lua defines a Lua extension
Only one of Inline or ValueRef must be set

_Appears in:_
- [EnvoyExtensionPolicySpec](#envoyextensionpolicyspec)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `type` | _[LuaValueType](#luavaluetype)_ |  true  | Inline | Type is the type of method to use to read the Lua value.<br />Valid values are Inline and ValueRef, default is Inline. |
| `inline` | _string_ |  false  |  | Inline contains the source code as an inline string. |
| `valueRef` | _[LocalObjectReference](#localobjectreference)_ |  false  |  | ValueRef has the source code specified as a local object reference.<br />Only a reference to ConfigMap is supported.<br />The value of key `lua` in the ConfigMap will be used.<br />If the key is not found, the first value in the ConfigMap will be used. |


#### LuaValueType

_Underlying type:_ _string_

LuaValueType defines the types of values for Lua supported by Envoy Gateway.

_Appears in:_
- [Lua](#lua)

| Value | Description |
| ----- | ----------- |
| `Inline` | LuaValueTypeInline defines the "Inline" Lua type.<br /> | 
| `ValueRef` | LuaValueTypeValueRef defines the "ValueRef" Lua type.<br /> | 


#### MergeType

_Underlying type:_ _string_

MergeType defines the type of merge operation

_Appears in:_
- [BackendTrafficPolicySpec](#backendtrafficpolicyspec)
- [KubernetesPatchSpec](#kubernetespatchspec)

| Value | Description |
| ----- | ----------- |
| `StrategicMerge` | StrategicMerge indicates a strategic merge patch type<br /> | 
| `JSONMerge` | JSONMerge indicates a JSON merge patch type<br /> | 


#### MetricSinkType

_Underlying type:_ _string_



_Appears in:_
- [EnvoyGatewayMetricSink](#envoygatewaymetricsink)
- [ProxyMetricSink](#proxymetricsink)

| Value | Description |
| ----- | ----------- |
| `OpenTelemetry` |  | 


#### OIDC



OIDC defines the configuration for the OpenID Connect (OIDC) authentication.

_Appears in:_
- [SecurityPolicySpec](#securitypolicyspec)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `provider` | _[OIDCProvider](#oidcprovider)_ |  true  |  | The OIDC Provider configuration. |
| `clientID` | _string_ |  true  |  | The client ID to be used in the OIDC<br />[Authentication Request](https://openid.net/specs/openid-connect-core-1_0.html#AuthRequest). |
| `clientSecret` | _[SecretObjectReference](https://gateway-api.sigs.k8s.io/references/spec/#gateway.networking.k8s.io/v1.SecretObjectReference)_ |  true  |  | The Kubernetes secret which contains the OIDC client secret to be used in the<br />[Authentication Request](https://openid.net/specs/openid-connect-core-1_0.html#AuthRequest).<br />This is an Opaque secret. The client secret should be stored in the key<br />"client-secret". |
| `cookieNames` | _[OIDCCookieNames](#oidccookienames)_ |  false  |  | The optional cookie name overrides to be used for Bearer and IdToken cookies in the<br />[Authentication Request](https://openid.net/specs/openid-connect-core-1_0.html#AuthRequest).<br />If not specified, uses a randomly generated suffix |
| `cookieConfig` | _[OIDCCookieConfig](#oidccookieconfig)_ |  false  |  | CookieConfigs allows overriding the SameSite attribute for OIDC cookies.<br />If a specific cookie is not configured, it will use the "Strict" SameSite policy by default. |
| `cookieDomain` | _string_ |  false  |  | The optional domain to set the access and ID token cookies on.<br />If not set, the cookies will default to the host of the request, not including the subdomains.<br />If set, the cookies will be set on the specified domain and all subdomains.<br />This means that requests to any subdomain will not require reauthentication after users log in to the parent domain. |
| `scopes` | _string array_ |  false  |  | The OIDC scopes to be used in the<br />[Authentication Request](https://openid.net/specs/openid-connect-core-1_0.html#AuthRequest).<br />The "openid" scope is always added to the list of scopes if not already<br />specified. |
| `resources` | _string array_ |  false  |  | The OIDC resources to be used in the<br />[Authentication Request](https://openid.net/specs/openid-connect-core-1_0.html#AuthRequest). |
| `redirectURL` | _string_ |  true  |  | The redirect URL to be used in the OIDC<br />[Authentication Request](https://openid.net/specs/openid-connect-core-1_0.html#AuthRequest).<br />If not specified, uses the default redirect URI "%REQ(x-forwarded-proto)%://%REQ(:authority)%/oauth2/callback" |
| `denyRedirect` | _[OIDCDenyRedirect](#oidcdenyredirect)_ |  false  |  | Any request that matches any of the provided matchers (with either tokens that are expired or missing tokens) will not be redirected to the OIDC Provider.<br />This behavior can be useful for AJAX or machine requests. |
| `logoutPath` | _string_ |  true  |  | The path to log a user out, clearing their credential cookies.<br />If not specified, uses a default logout path "/logout" |
| `forwardAccessToken` | _boolean_ |  false  |  | ForwardAccessToken indicates whether the Envoy should forward the access token<br />via the Authorization header Bearer scheme to the upstream.<br />If not specified, defaults to false. |
| `defaultTokenTTL` | _[Duration](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#duration-v1-meta)_ |  false  |  | DefaultTokenTTL is the default lifetime of the id token and access token.<br />Please note that Envoy will always use the expiry time from the response<br />of the authorization server if it is provided. This field is only used when<br />the expiry time is not provided by the authorization.<br />If not specified, defaults to 0. In this case, the "expires_in" field in<br />the authorization response must be set by the authorization server, or the<br />OAuth flow will fail. |
| `refreshToken` | _boolean_ |  false  |  | RefreshToken indicates whether the Envoy should automatically refresh the<br />id token and access token when they expire.<br />When set to true, the Envoy will use the refresh token to get a new id token<br />and access token when they expire.<br />If not specified, defaults to false. |
| `defaultRefreshTokenTTL` | _[Duration](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#duration-v1-meta)_ |  false  |  | DefaultRefreshTokenTTL is the default lifetime of the refresh token.<br />This field is only used when the exp (expiration time) claim is omitted in<br />the refresh token or the refresh token is not JWT.<br />If not specified, defaults to 604800s (one week).<br />Note: this field is only applicable when the "refreshToken" field is set to true. |
| `passThroughAuthHeader` | _boolean_ |  false  |  | Skips OIDC authentication when the request contains a header that will be extracted by the JWT filter. Unless<br />explicitly stated otherwise in the extractFrom field, this will be the "Authorization: Bearer ..." header.<br />The passThroughAuthHeader option is typically used for non-browser clients that may not be able to handle OIDC<br />redirects and wish to directly supply a token instead.<br />If not specified, defaults to false. |


#### OIDCCookieConfig





_Appears in:_
- [OIDC](#oidc)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `sameSite` | _string_ |  false  | Strict |  |


#### OIDCCookieNames



OIDCCookieNames defines the names of cookies to use in the Envoy OIDC filter.

_Appears in:_
- [OIDC](#oidc)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `accessToken` | _string_ |  false  |  | The name of the cookie used to store the AccessToken in the<br />[Authentication Request](https://openid.net/specs/openid-connect-core-1_0.html#AuthRequest).<br />If not specified, defaults to "AccessToken-(randomly generated uid)" |
| `idToken` | _string_ |  false  |  | The name of the cookie used to store the IdToken in the<br />[Authentication Request](https://openid.net/specs/openid-connect-core-1_0.html#AuthRequest).<br />If not specified, defaults to "IdToken-(randomly generated uid)" |


#### OIDCDenyRedirect



OIDCDenyRedirect defines headers to match against the request to deny redirect to the OIDC Provider.

_Appears in:_
- [OIDC](#oidc)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `headers` | _[OIDCDenyRedirectHeader](#oidcdenyredirectheader) array_ |  true  |  | Defines the headers to match against the request to deny redirect to the OIDC Provider. |


#### OIDCDenyRedirectHeader



OIDCDenyRedirectHeader defines how a header is matched

_Appears in:_
- [OIDCDenyRedirect](#oidcdenyredirect)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `name` | _string_ |  true  |  | Specifies the name of the header in the request. |
| `type` | _[StringMatchType](#stringmatchtype)_ |  false  | Exact | Type specifies how to match against a string. |
| `value` | _string_ |  true  |  | Value specifies the string value that the match must have. |


#### OIDCProvider



OIDCProvider defines the OIDC Provider configuration.

_Appears in:_
- [OIDC](#oidc)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `backendRef` | _[BackendObjectReference](https://gateway-api.sigs.k8s.io/references/spec/#gateway.networking.k8s.io/v1.BackendObjectReference)_ |  false  |  | BackendRef references a Kubernetes object that represents the<br />backend server to which the authorization request will be sent.<br />Deprecated: Use BackendRefs instead. |
| `backendRefs` | _[BackendRef](#backendref) array_ |  false  |  | BackendRefs references a Kubernetes object that represents the<br />backend server to which the authorization request will be sent. |
| `backendSettings` | _[ClusterSettings](#clustersettings)_ |  false  |  | BackendSettings holds configuration for managing the connection<br />to the backend. |
| `issuer` | _string_ |  true  |  | The OIDC Provider's [issuer identifier](https://openid.net/specs/openid-connect-discovery-1_0.html#IssuerDiscovery).<br />Issuer MUST be a URI RFC 3986 [RFC3986] with a scheme component that MUST<br />be https, a host component, and optionally, port and path components and<br />no query or fragment components. |
| `authorizationEndpoint` | _string_ |  false  |  | The OIDC Provider's [authorization endpoint](https://openid.net/specs/openid-connect-core-1_0.html#AuthorizationEndpoint).<br />If not provided, EG will try to discover it from the provider's [Well-Known Configuration Endpoint](https://openid.net/specs/openid-connect-discovery-1_0.html#ProviderConfigurationResponse). |
| `tokenEndpoint` | _string_ |  false  |  | The OIDC Provider's [token endpoint](https://openid.net/specs/openid-connect-core-1_0.html#TokenEndpoint).<br />If not provided, EG will try to discover it from the provider's [Well-Known Configuration Endpoint](https://openid.net/specs/openid-connect-discovery-1_0.html#ProviderConfigurationResponse). |


#### OpenTelemetryEnvoyProxyAccessLog



OpenTelemetryEnvoyProxyAccessLog defines the OpenTelemetry access log sink.

_Appears in:_
- [ProxyAccessLogSink](#proxyaccesslogsink)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `backendRef` | _[BackendObjectReference](https://gateway-api.sigs.k8s.io/references/spec/#gateway.networking.k8s.io/v1.BackendObjectReference)_ |  false  |  | BackendRef references a Kubernetes object that represents the<br />backend server to which the authorization request will be sent.<br />Deprecated: Use BackendRefs instead. |
| `backendRefs` | _[BackendRef](#backendref) array_ |  false  |  | BackendRefs references a Kubernetes object that represents the<br />backend server to which the authorization request will be sent. |
| `backendSettings` | _[ClusterSettings](#clustersettings)_ |  false  |  | BackendSettings holds configuration for managing the connection<br />to the backend. |
| `host` | _string_ |  false  |  | Host define the extension service hostname.<br />Deprecated: Use BackendRefs instead. |
| `port` | _integer_ |  false  | 4317 | Port defines the port the extension service is exposed on.<br />Deprecated: Use BackendRefs instead. |
| `resources` | _object (keys:string, values:string)_ |  false  |  | Resources is a set of labels that describe the source of a log entry, including envoy node info.<br />It's recommended to follow [semantic conventions](https://opentelemetry.io/docs/reference/specification/resource/semantic_conventions/). |


#### Operation



Operation specifies the operation of a request.

_Appears in:_
- [AuthorizationRule](#authorizationrule)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `methods` | _HTTPMethod array_ |  true  |  | Methods are the HTTP methods of the request.<br />If multiple methods are specified, all specified methods are allowed or denied, based on the action of the rule. |


#### Origin

_Underlying type:_ _string_

Origin is defined by the scheme (protocol), hostname (domain), and port of
the URL used to access it. The hostname can be "precise" which is just the
domain name or "wildcard" which is a domain name prefixed with a single
wildcard label such as "*.example.com".
In addition to that a single wildcard (with or without scheme) can be
configured to match any origin.

For example, the following are valid origins:
- https://foo.example.com
- https://*.example.com
- http://foo.example.com:8080
- http://*.example.com:8080
- https://*

_Appears in:_
- [CORS](#cors)



#### PassiveHealthCheck



PassiveHealthCheck defines the configuration for passive health checks in the context of Envoy's Outlier Detection,
see https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/upstream/outlier

_Appears in:_
- [HealthCheck](#healthcheck)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `splitExternalLocalOriginErrors` | _boolean_ |  false  | false | SplitExternalLocalOriginErrors enables splitting of errors between external and local origin. |
| `interval` | _[Duration](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#duration-v1-meta)_ |  false  | 3s | Interval defines the time between passive health checks. |
| `consecutiveLocalOriginFailures` | _integer_ |  false  | 5 | ConsecutiveLocalOriginFailures sets the number of consecutive local origin failures triggering ejection.<br />Parameter takes effect only when split_external_local_origin_errors is set to true. |
| `consecutiveGatewayErrors` | _integer_ |  false  | 0 | ConsecutiveGatewayErrors sets the number of consecutive gateway errors triggering ejection. |
| `consecutive5XxErrors` | _integer_ |  false  | 5 | Consecutive5xxErrors sets the number of consecutive 5xx errors triggering ejection. |
| `baseEjectionTime` | _[Duration](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#duration-v1-meta)_ |  false  | 30s | BaseEjectionTime defines the base duration for which a host will be ejected on consecutive failures. |
| `maxEjectionPercent` | _integer_ |  false  | 10 | MaxEjectionPercent sets the maximum percentage of hosts in a cluster that can be ejected. |


#### PathEscapedSlashAction

_Underlying type:_ _string_

PathEscapedSlashAction determines the action for requests that contain %2F, %2f, %5C, or %5c
sequences in the URI path.

_Appears in:_
- [PathSettings](#pathsettings)

| Value | Description |
| ----- | ----------- |
| `KeepUnchanged` | KeepUnchangedAction keeps escaped slashes as they arrive without changes<br /> | 
| `RejectRequest` | RejectRequestAction rejects client requests containing escaped slashes<br />with a 400 status. gRPC requests will be rejected with the INTERNAL (13)<br />error code.<br />The "httpN.downstream_rq_failed_path_normalization" counter is incremented<br />for each rejected request.<br /> | 
| `UnescapeAndRedirect` | UnescapeAndRedirect unescapes %2F and %5C sequences and redirects to the new path<br />if these sequences were present.<br />Redirect occurs after path normalization and merge slashes transformations if<br />they were configured. gRPC requests will be rejected with the INTERNAL (13)<br />error code.<br />This option minimizes possibility of path confusion exploits by forcing request<br />with unescaped slashes to traverse all parties: downstream client, intermediate<br />proxies, Envoy and upstream server.<br />The “httpN.downstream_rq_redirected_with_normalized_path” counter is incremented<br />for each redirected request.<br /> | 
| `UnescapeAndForward` | UnescapeAndForward unescapes %2F and %5C sequences and forwards the request.<br />Note: this option should not be enabled if intermediaries perform path based access<br />control as it may lead to path confusion vulnerabilities.<br /> | 


#### PathSettings



PathSettings provides settings that managing how the incoming path set by clients is handled.

_Appears in:_
- [ClientTrafficPolicySpec](#clienttrafficpolicyspec)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `escapedSlashesAction` | _[PathEscapedSlashAction](#pathescapedslashaction)_ |  false  |  | EscapedSlashesAction determines how %2f, %2F, %5c, or %5C sequences in the path URI<br />should be handled.<br />The default is UnescapeAndRedirect. |
| `disableMergeSlashes` | _boolean_ |  false  |  | DisableMergeSlashes allows disabling the default configuration of merging adjacent<br />slashes in the path.<br />Note that slash merging is not part of the HTTP spec and is provided for convenience. |


#### PerEndpointCircuitBreakers



PerEndpointCircuitBreakers defines Circuit Breakers that will apply per-endpoint for an upstream cluster

_Appears in:_
- [CircuitBreaker](#circuitbreaker)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `maxConnections` | _integer_ |  false  | 1024 | MaxConnections configures the maximum number of connections that Envoy will establish per-endpoint to the referenced backend defined within a xRoute rule. |


#### PerRetryPolicy





_Appears in:_
- [Retry](#retry)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `timeout` | _[Duration](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#duration-v1-meta)_ |  false  |  | Timeout is the timeout per retry attempt. |
| `backOff` | _[BackOffPolicy](#backoffpolicy)_ |  false  |  | Backoff is the backoff policy to be applied per retry attempt. gateway uses a fully jittered exponential<br />back-off algorithm for retries. For additional details,<br />see https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/router_filter#config-http-filters-router-x-envoy-max-retries |


#### PolicyTargetReferences





_Appears in:_
- [BackendTrafficPolicySpec](#backendtrafficpolicyspec)
- [ClientTrafficPolicySpec](#clienttrafficpolicyspec)
- [EnvoyExtensionPolicySpec](#envoyextensionpolicyspec)
- [SecurityPolicySpec](#securitypolicyspec)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `targetRef` | _[LocalPolicyTargetReferenceWithSectionName](https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1alpha2.LocalPolicyTargetReferenceWithSectionName)_ |  true  |  | TargetRef is the name of the resource this policy is being attached to.<br />This policy and the TargetRef MUST be in the same namespace for this<br />Policy to have effect<br />Deprecated: use targetRefs/targetSelectors instead |
| `targetRefs` | _[LocalPolicyTargetReferenceWithSectionName](https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1alpha2.LocalPolicyTargetReferenceWithSectionName) array_ |  true  |  | TargetRefs are the names of the Gateway resources this policy<br />is being attached to. |
| `targetSelectors` | _[TargetSelector](#targetselector) array_ |  true  |  | TargetSelectors allow targeting resources for this policy based on labels |


#### Principal



Principal specifies the client identity of a request.
A client identity can be a client IP, a JWT claim, username from the Authorization header,
or any other identity that can be extracted from a custom header.
If there are multiple principal types, all principals must match for the rule to match.

_Appears in:_
- [AuthorizationRule](#authorizationrule)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `clientCIDRs` | _[CIDR](#cidr) array_ |  false  |  | ClientCIDRs are the IP CIDR ranges of the client.<br />Valid examples are "192.168.1.0/24" or "2001:db8::/64"<br />If multiple CIDR ranges are specified, one of the CIDR ranges must match<br />the client IP for the rule to match.<br />The client IP is inferred from the X-Forwarded-For header, a custom header,<br />or the proxy protocol.<br />You can use the `ClientIPDetection` or the `EnableProxyProtocol` field in<br />the `ClientTrafficPolicy` to configure how the client IP is detected. |
| `jwt` | _[JWTPrincipal](#jwtprincipal)_ |  false  |  | JWT authorize the request based on the JWT claims and scopes.<br />Note: in order to use JWT claims for authorization, you must configure the<br />JWT authentication in the same `SecurityPolicy`. |
| `headers` | _[AuthorizationHeaderMatch](#authorizationheadermatch) array_ |  false  |  | Headers authorize the request based on user identity extracted from custom headers.<br />If multiple headers are specified, all headers must match for the rule to match. |


#### ProcessingModeOptions



ProcessingModeOptions defines if headers or body should be processed by the external service
and which attributes are sent to the processor

_Appears in:_
- [ExtProcProcessingMode](#extprocprocessingmode)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `body` | _[ExtProcBodyProcessingMode](#extprocbodyprocessingmode)_ |  false  |  | Defines body processing mode |
| `attributes` | _string array_ |  false  |  | Defines which attributes are sent to the external processor. Envoy Gateway currently<br />supports only the following attribute prefixes: connection, source, destination,<br />request, response, upstream and xds.route.<br />https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/advanced/attributes |


#### ProtocolUpgradeConfig



ProtocolUpgradeConfig specifies the configuration for protocol upgrades.

_Appears in:_
- [BackendTrafficPolicySpec](#backendtrafficpolicyspec)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `type` | _string_ |  true  |  | Type is the case-insensitive type of protocol upgrade.<br />e.g. `websocket`, `CONNECT`, `spdy/3.1` etc. |
| `connect` | _[ConnectConfig](#connectconfig)_ |  false  |  | Connect specifies the configuration for the CONNECT config.<br />This is allowed only when type is CONNECT. |


#### ProviderType

_Underlying type:_ _string_

ProviderType defines the types of providers supported by Envoy Gateway.

_Appears in:_
- [EnvoyGatewayProvider](#envoygatewayprovider)
- [EnvoyProxyProvider](#envoyproxyprovider)

| Value | Description |
| ----- | ----------- |
| `Kubernetes` | ProviderTypeKubernetes defines the "Kubernetes" provider.<br /> | 
| `Custom` | ProviderTypeCustom defines the "Custom" provider.<br /> | 


#### ProxyAccessLog





_Appears in:_
- [ProxyTelemetry](#proxytelemetry)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `disable` | _boolean_ |  false  |  | Disable disables access logging for managed proxies if set to true. |
| `settings` | _[ProxyAccessLogSetting](#proxyaccesslogsetting) array_ |  false  |  | Settings defines accesslog settings for managed proxies.<br />If unspecified, will send default format to stdout. |


#### ProxyAccessLogFormat



ProxyAccessLogFormat defines the format of accesslog.
By default accesslogs are written to standard output.

_Appears in:_
- [ProxyAccessLogSetting](#proxyaccesslogsetting)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `type` | _[ProxyAccessLogFormatType](#proxyaccesslogformattype)_ |  true  |  | Type defines the type of accesslog format. |
| `text` | _string_ |  false  |  | Text defines the text accesslog format, following Envoy accesslog formatting,<br />It's required when the format type is "Text".<br />Envoy [command operators](https://www.envoyproxy.io/docs/envoy/latest/configuration/observability/access_log/usage#command-operators) may be used in the format.<br />The [format string documentation](https://www.envoyproxy.io/docs/envoy/latest/configuration/observability/access_log/usage#config-access-log-format-strings) provides more information. |
| `json` | _object (keys:string, values:string)_ |  false  |  | JSON is additional attributes that describe the specific event occurrence.<br />Structured format for the envoy access logs. Envoy [command operators](https://www.envoyproxy.io/docs/envoy/latest/configuration/observability/access_log/usage#command-operators)<br />can be used as values for fields within the Struct.<br />It's required when the format type is "JSON". |


#### ProxyAccessLogFormatType

_Underlying type:_ _string_



_Appears in:_
- [ProxyAccessLogFormat](#proxyaccesslogformat)

| Value | Description |
| ----- | ----------- |
| `Text` | ProxyAccessLogFormatTypeText defines the text accesslog format.<br /> | 
| `JSON` | ProxyAccessLogFormatTypeJSON defines the JSON accesslog format.<br /> | 


#### ProxyAccessLogSetting





_Appears in:_
- [ProxyAccessLog](#proxyaccesslog)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `format` | _[ProxyAccessLogFormat](#proxyaccesslogformat)_ |  false  |  | Format defines the format of accesslog.<br />This will be ignored if sink type is ALS. |
| `matches` | _string array_ |  true  |  | Matches defines the match conditions for accesslog in CEL expression.<br />An accesslog will be emitted only when one or more match conditions are evaluated to true.<br />Invalid [CEL](https://www.envoyproxy.io/docs/envoy/latest/xds/type/v3/cel.proto.html#common-expression-language-cel-proto) expressions will be ignored. |
| `sinks` | _[ProxyAccessLogSink](#proxyaccesslogsink) array_ |  true  |  | Sinks defines the sinks of accesslog. |
| `type` | _[ProxyAccessLogType](#proxyaccesslogtype)_ |  false  |  | Type defines the component emitting the accesslog, such as Listener and Route.<br />If type not defined, the setting would apply to:<br />(1) All Routes.<br />(2) Listeners if and only if Envoy does not find a matching route for a request.<br />If type is defined, the accesslog settings would apply to the relevant component (as-is). |


#### ProxyAccessLogSink



ProxyAccessLogSink defines the sink of accesslog.

_Appears in:_
- [ProxyAccessLogSetting](#proxyaccesslogsetting)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `type` | _[ProxyAccessLogSinkType](#proxyaccesslogsinktype)_ |  true  |  | Type defines the type of accesslog sink. |
| `als` | _[ALSEnvoyProxyAccessLog](#alsenvoyproxyaccesslog)_ |  false  |  | ALS defines the gRPC Access Log Service (ALS) sink. |
| `file` | _[FileEnvoyProxyAccessLog](#fileenvoyproxyaccesslog)_ |  false  |  | File defines the file accesslog sink. |
| `openTelemetry` | _[OpenTelemetryEnvoyProxyAccessLog](#opentelemetryenvoyproxyaccesslog)_ |  false  |  | OpenTelemetry defines the OpenTelemetry accesslog sink. |


#### ProxyAccessLogSinkType

_Underlying type:_ _string_



_Appears in:_
- [ProxyAccessLogSink](#proxyaccesslogsink)

| Value | Description |
| ----- | ----------- |
| `ALS` | ProxyAccessLogSinkTypeALS defines the gRPC Access Log Service (ALS) sink.<br />The service must implement the Envoy gRPC Access Log Service streaming API:<br />https://www.envoyproxy.io/docs/envoy/latest/api-v3/service/accesslog/v3/als.proto<br /> | 
| `File` | ProxyAccessLogSinkTypeFile defines the file accesslog sink.<br /> | 
| `OpenTelemetry` | ProxyAccessLogSinkTypeOpenTelemetry defines the OpenTelemetry accesslog sink.<br />When the provider is Kubernetes, EnvoyGateway always sends `k8s.namespace.name`<br />and `k8s.pod.name` as additional attributes.<br /> | 


#### ProxyAccessLogType

_Underlying type:_ _string_



_Appears in:_
- [ProxyAccessLogSetting](#proxyaccesslogsetting)

| Value | Description |
| ----- | ----------- |
| `Listener` | ProxyAccessLogTypeListener defines the accesslog for Listeners.<br />https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/listener/v3/listener.proto#envoy-v3-api-field-config-listener-v3-listener-access-log<br /> | 
| `Route` | ProxyAccessLogTypeRoute defines the accesslog for HTTP, GRPC, UDP and TCP Routes.<br />https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/filters/udp/udp_proxy/v3/udp_proxy.proto#envoy-v3-api-field-extensions-filters-udp-udp-proxy-v3-udpproxyconfig-access-log<br />https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/filters/network/tcp_proxy/v3/tcp_proxy.proto#envoy-v3-api-field-extensions-filters-network-tcp-proxy-v3-tcpproxy-access-log<br />https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/filters/network/http_connection_manager/v3/http_connection_manager.proto#envoy-v3-api-field-extensions-filters-network-http-connection-manager-v3-httpconnectionmanager-access-log<br /> | 


#### ProxyBootstrap



ProxyBootstrap defines Envoy Bootstrap configuration.

_Appears in:_
- [EnvoyProxySpec](#envoyproxyspec)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `type` | _[BootstrapType](#bootstraptype)_ |  false  | Replace | Type is the type of the bootstrap configuration, it should be either **Replace**,  **Merge**, or **JSONPatch**.<br />If unspecified, it defaults to Replace. |
| `value` | _string_ |  false  |  | Value is a YAML string of the bootstrap. |
| `jsonPatches` | _[JSONPatchOperation](#jsonpatchoperation) array_ |  true  |  | JSONPatches is an array of JSONPatches to be applied to the default bootstrap. Patches are<br />applied in the order in which they are defined. |


#### ProxyLogComponent

_Underlying type:_ _string_

ProxyLogComponent defines a component that supports a configured logging level.

_Appears in:_
- [ProxyLogging](#proxylogging)

| Value | Description |
| ----- | ----------- |
| `default` | LogComponentDefault defines the default logging component.<br />See more details: https://www.envoyproxy.io/docs/envoy/latest/operations/cli#cmdoption-l<br /> | 
| `upstream` | LogComponentUpstream defines the "upstream" logging component.<br /> | 
| `http` | LogComponentHTTP defines the "http" logging component.<br /> | 
| `connection` | LogComponentConnection defines the "connection" logging component.<br /> | 
| `admin` | LogComponentAdmin defines the "admin" logging component.<br /> | 
| `client` | LogComponentClient defines the "client" logging component.<br /> | 
| `filter` | LogComponentFilter defines the "filter" logging component.<br /> | 
| `main` | LogComponentMain defines the "main" logging component.<br /> | 
| `router` | LogComponentRouter defines the "router" logging component.<br /> | 
| `runtime` | LogComponentRuntime defines the "runtime" logging component.<br /> | 


#### ProxyLogging



ProxyLogging defines logging parameters for managed proxies.

_Appears in:_
- [EnvoyProxySpec](#envoyproxyspec)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `level` | _object (keys:[ProxyLogComponent](#proxylogcomponent), values:[LogLevel](#loglevel))_ |  true  | \{ default:warn \} | Level is a map of logging level per component, where the component is the key<br />and the log level is the value. If unspecified, defaults to "default: warn". |


#### ProxyMetricSink



ProxyMetricSink defines the sink of metrics.
Default metrics sink is OpenTelemetry.

_Appears in:_
- [ProxyMetrics](#proxymetrics)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `type` | _[MetricSinkType](#metricsinktype)_ |  true  | OpenTelemetry | Type defines the metric sink type.<br />EG currently only supports OpenTelemetry. |
| `openTelemetry` | _[ProxyOpenTelemetrySink](#proxyopentelemetrysink)_ |  false  |  | OpenTelemetry defines the configuration for OpenTelemetry sink.<br />It's required if the sink type is OpenTelemetry. |


#### ProxyMetrics





_Appears in:_
- [ProxyTelemetry](#proxytelemetry)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `prometheus` | _[ProxyPrometheusProvider](#proxyprometheusprovider)_ |  true  |  | Prometheus defines the configuration for Admin endpoint `/stats/prometheus`. |
| `sinks` | _[ProxyMetricSink](#proxymetricsink) array_ |  true  |  | Sinks defines the metric sinks where metrics are sent to. |
| `matches` | _[StringMatch](#stringmatch) array_ |  true  |  | Matches defines configuration for selecting specific metrics instead of generating all metrics stats<br />that are enabled by default. This helps reduce CPU and memory overhead in Envoy, but eliminating some stats<br />may after critical functionality. Here are the stats that we strongly recommend not disabling:<br />`cluster_manager.warming_clusters`, `cluster.<cluster_name>.membership_total`,`cluster.<cluster_name>.membership_healthy`,<br />`cluster.<cluster_name>.membership_degraded`，reference  https://github.com/envoyproxy/envoy/issues/9856,<br />https://github.com/envoyproxy/envoy/issues/14610 |
| `enableVirtualHostStats` | _boolean_ |  false  |  | EnableVirtualHostStats enables envoy stat metrics for virtual hosts. |
| `enablePerEndpointStats` | _boolean_ |  false  |  | EnablePerEndpointStats enables per endpoint envoy stats metrics.<br />Please use with caution. |
| `enableRequestResponseSizesStats` | _boolean_ |  false  |  | EnableRequestResponseSizesStats enables publishing of histograms tracking header and body sizes of requests and responses. |
| `clusterStatName` | _string_ |  false  |  | ClusterStatName defines the value of cluster alt_stat_name, determining how cluster stats are named.<br />For more details, see envoy docs: https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/cluster/v3/cluster.proto.html<br />The supported operators for this pattern are:<br />%ROUTE_NAME%: name of Gateway API xRoute resource<br />%ROUTE_NAMESPACE%: namespace of Gateway API xRoute resource<br />%ROUTE_KIND%: kind of Gateway API xRoute resource<br />%ROUTE_RULE_NAME%: name of the Gateway API xRoute section<br />%ROUTE_RULE_NUMBER%: name of the Gateway API xRoute section<br />%BACKEND_REFS%: names of all backends referenced in <NAMESPACE>/<NAME>\|<NAMESPACE>/<NAME>\|... format<br />Only xDS Clusters created for HTTPRoute and GRPCRoute are currently supported.<br />Default: %ROUTE_KIND%/%ROUTE_NAMESPACE%/%ROUTE_NAME%/rule/%ROUTE_RULE_NUMBER%<br />Example: httproute/my-ns/my-route/rule/0 |


#### ProxyOpenTelemetrySink



ProxyOpenTelemetrySink defines the configuration for OpenTelemetry sink.

_Appears in:_
- [ProxyMetricSink](#proxymetricsink)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `backendRef` | _[BackendObjectReference](https://gateway-api.sigs.k8s.io/references/spec/#gateway.networking.k8s.io/v1.BackendObjectReference)_ |  false  |  | BackendRef references a Kubernetes object that represents the<br />backend server to which the authorization request will be sent.<br />Deprecated: Use BackendRefs instead. |
| `backendRefs` | _[BackendRef](#backendref) array_ |  false  |  | BackendRefs references a Kubernetes object that represents the<br />backend server to which the authorization request will be sent. |
| `backendSettings` | _[ClusterSettings](#clustersettings)_ |  false  |  | BackendSettings holds configuration for managing the connection<br />to the backend. |
| `host` | _string_ |  false  |  | Host define the service hostname.<br />Deprecated: Use BackendRefs instead. |
| `port` | _integer_ |  false  | 4317 | Port defines the port the service is exposed on.<br />Deprecated: Use BackendRefs instead. |


#### ProxyPrometheusProvider





_Appears in:_
- [ProxyMetrics](#proxymetrics)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `disable` | _boolean_ |  true  |  | Disable the Prometheus endpoint. |
| `compression` | _[Compression](#compression)_ |  false  |  | Configure the compression on Prometheus endpoint. Compression is useful in situations when bandwidth is scarce and large payloads can be effectively compressed at the expense of higher CPU load. |


#### ProxyProtocol



ProxyProtocol defines the configuration related to the proxy protocol
when communicating with the backend.

_Appears in:_
- [BackendTrafficPolicySpec](#backendtrafficpolicyspec)
- [ClusterSettings](#clustersettings)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `version` | _[ProxyProtocolVersion](#proxyprotocolversion)_ |  true  |  | Version of ProxyProtol<br />Valid ProxyProtocolVersion values are<br />"V1"<br />"V2" |


#### ProxyProtocolVersion

_Underlying type:_ _string_

ProxyProtocolVersion defines the version of the Proxy Protocol to use.

_Appears in:_
- [ProxyProtocol](#proxyprotocol)

| Value | Description |
| ----- | ----------- |
| `V1` | ProxyProtocolVersionV1 is the PROXY protocol version 1 (human readable format).<br /> | 
| `V2` | ProxyProtocolVersionV2 is the PROXY protocol version 2 (binary format).<br /> | 


#### ProxyTelemetry





_Appears in:_
- [EnvoyProxySpec](#envoyproxyspec)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `accessLog` | _[ProxyAccessLog](#proxyaccesslog)_ |  false  |  | AccessLogs defines accesslog parameters for managed proxies.<br />If unspecified, will send default format to stdout. |
| `tracing` | _[ProxyTracing](#proxytracing)_ |  false  |  | Tracing defines tracing configuration for managed proxies.<br />If unspecified, will not send tracing data. |
| `metrics` | _[ProxyMetrics](#proxymetrics)_ |  true  |  | Metrics defines metrics configuration for managed proxies. |


#### ProxyTracing



ProxyTracing defines the tracing configuration for a proxy.

_Appears in:_
- [ProxyTelemetry](#proxytelemetry)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `samplingRate` | _integer_ |  false  |  | SamplingRate controls the rate at which traffic will be<br />selected for tracing if no prior sampling decision has been made.<br />Defaults to 100, valid values [0-100]. 100 indicates 100% sampling.<br />Only one of SamplingRate or SamplingFraction may be specified.<br />If neither field is specified, all requests will be sampled. |
| `samplingFraction` | _[Fraction](https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1.Fraction)_ |  false  |  | SamplingFraction represents the fraction of requests that should be<br />selected for tracing if no prior sampling decision has been made.<br />Only one of SamplingRate or SamplingFraction may be specified.<br />If neither field is specified, all requests will be sampled. |
| `customTags` | _object (keys:string, values:[CustomTag](#customtag))_ |  false  |  | CustomTags defines the custom tags to add to each span.<br />If provider is kubernetes, pod name and namespace are added by default. |
| `provider` | _[TracingProvider](#tracingprovider)_ |  true  |  | Provider defines the tracing provider. |


#### RateLimit



RateLimit defines the configuration associated with the Rate Limit Service
used for Global Rate Limiting.

_Appears in:_
- [EnvoyGateway](#envoygateway)
- [EnvoyGatewaySpec](#envoygatewayspec)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `backend` | _[RateLimitDatabaseBackend](#ratelimitdatabasebackend)_ |  true  |  | Backend holds the configuration associated with the<br />database backend used by the rate limit service to store<br />state associated with global ratelimiting. |
| `timeout` | _[Duration](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#duration-v1-meta)_ |  false  |  | Timeout specifies the timeout period for the proxy to access the ratelimit server<br />If not set, timeout is 20ms. |
| `failClosed` | _boolean_ |  true  |  | FailClosed is a switch used to control the flow of traffic<br />when the response from the ratelimit server cannot be obtained.<br />If FailClosed is false, let the traffic pass,<br />otherwise, don't let the traffic pass and return 500.<br />If not set, FailClosed is False. |
| `telemetry` | _[RateLimitTelemetry](#ratelimittelemetry)_ |  false  |  | Telemetry defines telemetry configuration for RateLimit. |


#### RateLimitCost





_Appears in:_
- [RateLimitRule](#ratelimitrule)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `request` | _[RateLimitCostSpecifier](#ratelimitcostspecifier)_ |  false  |  | Request specifies the number to reduce the rate limit counters<br />on the request path. If this is not specified, the default behavior<br />is to reduce the rate limit counters by 1.<br />When Envoy receives a request that matches the rule, it tries to reduce the<br />rate limit counters by the specified number. If the counter doesn't have<br />enough capacity, the request is rate limited. |
| `response` | _[RateLimitCostSpecifier](#ratelimitcostspecifier)_ |  false  |  | Response specifies the number to reduce the rate limit counters<br />after the response is sent back to the client or the request stream is closed.<br />The cost is used to reduce the rate limit counters for the matching requests.<br />Since the reduction happens after the request stream is complete, the rate limit<br />won't be enforced for the current request, but for the subsequent matching requests.<br />This is optional and if not specified, the rate limit counters are not reduced<br />on the response path.<br />Currently, this is only supported for HTTP Global Rate Limits. |


#### RateLimitCostFrom

_Underlying type:_ _string_

RateLimitCostFrom specifies the source of the rate limit cost.
Valid RateLimitCostType values are "Number" and "Metadata".

_Appears in:_
- [RateLimitCostSpecifier](#ratelimitcostspecifier)

| Value | Description |
| ----- | ----------- |
| `Number` | RateLimitCostFromNumber specifies the rate limit cost to be a fixed number.<br /> | 
| `Metadata` | RateLimitCostFromMetadata specifies the rate limit cost to be retrieved from the per-request dynamic metadata.<br /> | 


#### RateLimitCostMetadata



RateLimitCostMetadata specifies the filter metadata to retrieve the usage number from.

_Appears in:_
- [RateLimitCostSpecifier](#ratelimitcostspecifier)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `namespace` | _string_ |  true  |  | Namespace is the namespace of the dynamic metadata. |
| `key` | _string_ |  true  |  | Key is the key to retrieve the usage number from the filter metadata. |


#### RateLimitCostSpecifier



RateLimitCostSpecifier specifies where the Envoy retrieves the number to reduce the rate limit counters.

_Appears in:_
- [RateLimitCost](#ratelimitcost)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `from` | _[RateLimitCostFrom](#ratelimitcostfrom)_ |  true  |  | From specifies where to get the rate limit cost. Currently, only "Number" and "Metadata" are supported. |
| `number` | _integer_ |  false  |  | Number specifies the fixed usage number to reduce the rate limit counters.<br />Using zero can be used to only check the rate limit counters without reducing them. |
| `metadata` | _[RateLimitCostMetadata](#ratelimitcostmetadata)_ |  false  |  | Refer to Kubernetes API documentation for fields of `metadata`. |


#### RateLimitDatabaseBackend



RateLimitDatabaseBackend defines the configuration associated with
the database backend used by the rate limit service.

_Appears in:_
- [RateLimit](#ratelimit)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `type` | _[RateLimitDatabaseBackendType](#ratelimitdatabasebackendtype)_ |  true  |  | Type is the type of database backend to use. Supported types are:<br />	* Redis: Connects to a Redis database. |
| `redis` | _[RateLimitRedisSettings](#ratelimitredissettings)_ |  false  |  | Redis defines the settings needed to connect to a Redis database. |


#### RateLimitDatabaseBackendType

_Underlying type:_ _string_

RateLimitDatabaseBackendType specifies the types of database backend
to be used by the rate limit service.

_Appears in:_
- [RateLimitDatabaseBackend](#ratelimitdatabasebackend)

| Value | Description |
| ----- | ----------- |
| `Redis` | RedisBackendType uses a redis database for the rate limit service.<br /> | 


#### RateLimitMetrics





_Appears in:_
- [RateLimitTelemetry](#ratelimittelemetry)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `prometheus` | _[RateLimitMetricsPrometheusProvider](#ratelimitmetricsprometheusprovider)_ |  true  |  | Prometheus defines the configuration for prometheus endpoint. |


#### RateLimitMetricsPrometheusProvider





_Appears in:_
- [RateLimitMetrics](#ratelimitmetrics)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `disable` | _boolean_ |  true  |  | Disable the Prometheus endpoint. |


#### RateLimitRedisSettings



RateLimitRedisSettings defines the configuration for connecting to redis database.

_Appears in:_
- [RateLimitDatabaseBackend](#ratelimitdatabasebackend)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `url` | _string_ |  true  |  | URL of the Redis Database. |
| `tls` | _[RedisTLSSettings](#redistlssettings)_ |  false  |  | TLS defines TLS configuration for connecting to redis database. |


#### RateLimitRule



RateLimitRule defines the semantics for matching attributes
from the incoming requests, and setting limits for them.

_Appears in:_
- [GlobalRateLimit](#globalratelimit)
- [LocalRateLimit](#localratelimit)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `clientSelectors` | _[RateLimitSelectCondition](#ratelimitselectcondition) array_ |  false  |  | ClientSelectors holds the list of select conditions to select<br />specific clients using attributes from the traffic flow.<br />All individual select conditions must hold True for this rule<br />and its limit to be applied.<br />If no client selectors are specified, the rule applies to all traffic of<br />the targeted Route.<br />If the policy targets a Gateway, the rule applies to each Route of the Gateway.<br />Please note that each Route has its own rate limit counters. For example,<br />if a Gateway has two Routes, and the policy has a rule with limit 10rps,<br />each Route will have its own 10rps limit. |
| `limit` | _[RateLimitValue](#ratelimitvalue)_ |  true  |  | Limit holds the rate limit values.<br />This limit is applied for traffic flows when the selectors<br />compute to True, causing the request to be counted towards the limit.<br />The limit is enforced and the request is ratelimited, i.e. a response with<br />429 HTTP status code is sent back to the client when<br />the selected requests have reached the limit. |
| `cost` | _[RateLimitCost](#ratelimitcost)_ |  false  |  | Cost specifies the cost of requests and responses for the rule.<br />This is optional and if not specified, the default behavior is to reduce the rate limit counters by 1 on<br />the request path and do not reduce the rate limit counters on the response path. |
| `shared` | _boolean_ |  false  |  | Shared determines whether this rate limit rule applies across all the policy targets.<br />If set to true, the rule is treated as a common bucket and is shared across all policy targets (xRoutes).<br />Default: false. |


#### RateLimitSelectCondition



RateLimitSelectCondition specifies the attributes within the traffic flow that can
be used to select a subset of clients to be ratelimited.
All the individual conditions must hold True for the overall condition to hold True.

_Appears in:_
- [RateLimitRule](#ratelimitrule)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `headers` | _[HeaderMatch](#headermatch) array_ |  false  |  | Headers is a list of request headers to match. Multiple header values are ANDed together,<br />meaning, a request MUST match all the specified headers.<br />At least one of headers or sourceCIDR condition must be specified. |
| `sourceCIDR` | _[SourceMatch](#sourcematch)_ |  false  |  | SourceCIDR is the client IP Address range to match on.<br />At least one of headers or sourceCIDR condition must be specified. |


#### RateLimitSpec



RateLimitSpec defines the desired state of RateLimitSpec.

_Appears in:_
- [BackendTrafficPolicySpec](#backendtrafficpolicyspec)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `type` | _[RateLimitType](#ratelimittype)_ |  true  |  | Type decides the scope for the RateLimits.<br />Valid RateLimitType values are "Global" or "Local". |
| `global` | _[GlobalRateLimit](#globalratelimit)_ |  false  |  | Global defines global rate limit configuration. |
| `local` | _[LocalRateLimit](#localratelimit)_ |  false  |  | Local defines local rate limit configuration. |


#### RateLimitTelemetry





_Appears in:_
- [RateLimit](#ratelimit)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `metrics` | _[RateLimitMetrics](#ratelimitmetrics)_ |  true  |  | Metrics defines metrics configuration for RateLimit. |
| `tracing` | _[RateLimitTracing](#ratelimittracing)_ |  true  |  | Tracing defines traces configuration for RateLimit. |


#### RateLimitTracing





_Appears in:_
- [RateLimitTelemetry](#ratelimittelemetry)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `samplingRate` | _integer_ |  false  |  | SamplingRate controls the rate at which traffic will be<br />selected for tracing if no prior sampling decision has been made.<br />Defaults to 100, valid values [0-100]. 100 indicates 100% sampling. |
| `provider` | _[RateLimitTracingProvider](#ratelimittracingprovider)_ |  true  |  | Provider defines the rateLimit tracing provider.<br />Only OpenTelemetry is supported currently. |


#### RateLimitTracingProvider



RateLimitTracingProvider defines the tracing provider configuration of RateLimit

_Appears in:_
- [RateLimitTracing](#ratelimittracing)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `type` | _[RateLimitTracingProviderType](#ratelimittracingprovidertype)_ |  true  |  | Type defines the tracing provider type.<br />Since to RateLimit Exporter currently using OpenTelemetry, only OpenTelemetry is supported |
| `url` | _string_ |  true  |  | URL is the endpoint of the trace collector that supports the OTLP protocol |


#### RateLimitTracingProviderType

_Underlying type:_ _string_



_Appears in:_
- [RateLimitTracingProvider](#ratelimittracingprovider)



#### RateLimitType

_Underlying type:_ _string_

RateLimitType specifies the types of RateLimiting.

_Appears in:_
- [RateLimitSpec](#ratelimitspec)

| Value | Description |
| ----- | ----------- |
| `Global` | GlobalRateLimitType allows the rate limits to be applied across all Envoy<br />proxy instances.<br /> | 
| `Local` | LocalRateLimitType allows the rate limits to be applied on a per Envoy<br />proxy instance basis.<br /> | 


#### RateLimitUnit

_Underlying type:_ _string_

RateLimitUnit specifies the intervals for setting rate limits.
Valid RateLimitUnit values are "Second", "Minute", "Hour", and "Day".

_Appears in:_
- [RateLimitValue](#ratelimitvalue)

| Value | Description |
| ----- | ----------- |
| `Second` | RateLimitUnitSecond specifies the rate limit interval to be 1 second.<br /> | 
| `Minute` | RateLimitUnitMinute specifies the rate limit interval to be 1 minute.<br /> | 
| `Hour` | RateLimitUnitHour specifies the rate limit interval to be 1 hour.<br /> | 
| `Day` | RateLimitUnitDay specifies the rate limit interval to be 1 day.<br /> | 


#### RateLimitValue



RateLimitValue defines the limits for rate limiting.

_Appears in:_
- [RateLimitRule](#ratelimitrule)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `requests` | _integer_ |  true  |  |  |
| `unit` | _[RateLimitUnit](#ratelimitunit)_ |  true  |  |  |


#### RedisTLSSettings



RedisTLSSettings defines the TLS configuration for connecting to redis database.

_Appears in:_
- [RateLimitRedisSettings](#ratelimitredissettings)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `certificateRef` | _[SecretObjectReference](https://gateway-api.sigs.k8s.io/references/spec/#gateway.networking.k8s.io/v1.SecretObjectReference)_ |  false  |  | CertificateRef defines the client certificate reference for TLS connections.<br />Currently only a Kubernetes Secret of type TLS is supported. |


#### RemoteJWKS



RemoteJWKS defines how to fetch and cache JSON Web Key Sets (JWKS) from a remote HTTP/HTTPS endpoint.

_Appears in:_
- [JWTProvider](#jwtprovider)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `backendRef` | _[BackendObjectReference](https://gateway-api.sigs.k8s.io/references/spec/#gateway.networking.k8s.io/v1.BackendObjectReference)_ |  false  |  | BackendRef references a Kubernetes object that represents the<br />backend server to which the authorization request will be sent.<br />Deprecated: Use BackendRefs instead. |
| `backendRefs` | _[BackendRef](#backendref) array_ |  false  |  | BackendRefs references a Kubernetes object that represents the<br />backend server to which the authorization request will be sent. |
| `backendSettings` | _[ClusterSettings](#clustersettings)_ |  false  |  | BackendSettings holds configuration for managing the connection<br />to the backend. |
| `uri` | _string_ |  true  |  | URI is the HTTPS URI to fetch the JWKS. Envoy's system trust bundle is used to validate the server certificate.<br />If a custom trust bundle is needed, it can be specified in a BackendTLSConfig resource and target the BackendRefs. |


#### ReplaceRegexMatch





_Appears in:_
- [HTTPPathModifier](#httppathmodifier)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `pattern` | _string_ |  true  |  | Pattern matches a regular expression against the value of the HTTP Path.The regex string must<br />adhere to the syntax documented in https://github.com/google/re2/wiki/Syntax. |
| `substitution` | _string_ |  true  |  | Substitution is an expression that replaces the matched portion.The expression may include numbered<br />capture groups that adhere to syntax documented in https://github.com/google/re2/wiki/Syntax. |


#### RequestBuffer





_Appears in:_
- [BackendTrafficPolicySpec](#backendtrafficpolicyspec)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |


#### RequestHeaderCustomTag



RequestHeaderCustomTag adds value from request header to each span.

_Appears in:_
- [CustomTag](#customtag)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `name` | _string_ |  true  |  | Name defines the name of the request header which to extract the value from. |
| `defaultValue` | _string_ |  false  |  | DefaultValue defines the default value to use if the request header is not set. |


#### RequestIDAction

_Underlying type:_ _string_

RequestIDAction configures Envoy's behavior for handling the `X-Request-ID` header.

_Appears in:_
- [HeaderSettings](#headersettings)

| Value | Description |
| ----- | ----------- |
| `PreserveOrGenerate` | Preserve `X-Request-ID` if already present or generate if empty<br /> | 
| `Preserve` | Preserve `X-Request-ID` if already present, do not generate when empty<br /> | 
| `Generate` | Always generate `X-Request-ID` header, do not preserve `X-Request-ID`<br />header if it exists. This is the default behavior.<br /> | 
| `Disable` | Do not preserve or generate `X-Request-ID` header<br /> | 


#### ResourceProviderType

_Underlying type:_ _string_

ResourceProviderType defines the types of custom resource providers supported by Envoy Gateway.

_Appears in:_
- [EnvoyGatewayResourceProvider](#envoygatewayresourceprovider)

| Value | Description |
| ----- | ----------- |
| `File` | ResourceProviderTypeFile defines the "File" provider.<br /> | 


#### ResponseOverride



ResponseOverride defines the configuration to override specific responses with a custom one.

_Appears in:_
- [BackendTrafficPolicySpec](#backendtrafficpolicyspec)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `match` | _[CustomResponseMatch](#customresponsematch)_ |  true  |  | Match configuration. |
| `response` | _[CustomResponse](#customresponse)_ |  true  |  | Response configuration. |
| `redirect` | _[CustomRedirect](#customredirect)_ |  true  |  | Redirect configuration |


#### ResponseValueType

_Underlying type:_ _string_

ResponseValueType defines the types of values for the response body supported by Envoy Gateway.

_Appears in:_
- [CustomResponseBody](#customresponsebody)

| Value | Description |
| ----- | ----------- |
| `Inline` | ResponseValueTypeInline defines the "Inline" response body type.<br /> | 
| `ValueRef` | ResponseValueTypeValueRef defines the "ValueRef" response body type.<br /> | 


#### Retry



Retry defines the retry strategy to be applied.

_Appears in:_
- [BackendTrafficPolicySpec](#backendtrafficpolicyspec)
- [ClusterSettings](#clustersettings)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `numRetries` | _integer_ |  false  | 2 | NumRetries is the number of retries to be attempted. Defaults to 2. |
| `numAttemptsPerPriority` | _integer_ |  false  |  | NumAttemptsPerPriority defines the number of requests (initial attempt + retries)<br />that should be sent to the same priority before switching to a different one.<br />If not specified or set to 0, all requests are sent to the highest priority that is healthy. |
| `retryOn` | _[RetryOn](#retryon)_ |  false  |  | RetryOn specifies the retry trigger condition.<br />If not specified, the default is to retry on connect-failure,refused-stream,unavailable,cancelled,retriable-status-codes(503). |
| `perRetry` | _[PerRetryPolicy](#perretrypolicy)_ |  false  |  | PerRetry is the retry policy to be applied per retry attempt. |


#### RetryOn





_Appears in:_
- [Retry](#retry)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `triggers` | _[TriggerEnum](#triggerenum) array_ |  false  |  | Triggers specifies the retry trigger condition(Http/Grpc). |
| `httpStatusCodes` | _[HTTPStatus](#httpstatus) array_ |  false  |  | HttpStatusCodes specifies the http status codes to be retried.<br />The retriable-status-codes trigger must also be configured for these status codes to trigger a retry. |


#### RetryableGRPCStatusCode

_Underlying type:_ _string_

GRPCStatus defines grpc status codes as defined in https://github.com/grpc/grpc/blob/master/doc/statuscodes.md.

_Appears in:_
- [ExtensionServiceRetry](#extensionserviceretry)



#### RoutingType

_Underlying type:_ _string_

RoutingType defines the type of routing of this Envoy proxy.

_Appears in:_
- [EnvoyProxySpec](#envoyproxyspec)

| Value | Description |
| ----- | ----------- |
| `Service` | ServiceRoutingType is the RoutingType for Service Cluster IP routing.<br /> | 
| `Endpoint` | EndpointRoutingType is the RoutingType for Endpoint routing.<br /> | 




#### SecurityPolicy



SecurityPolicy allows the user to configure various security settings for a
Gateway.



| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `apiVersion` | _string_ | |`gateway.envoyproxy.io/v1alpha1`
| `kind` | _string_ | |`SecurityPolicy`
| `metadata` | _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#objectmeta-v1-meta)_ |  true  |  | Refer to Kubernetes API documentation for fields of `metadata`. |
| `spec` | _[SecurityPolicySpec](#securitypolicyspec)_ |  true  |  | Spec defines the desired state of SecurityPolicy. |
| `status` | _[PolicyStatus](https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1alpha2.PolicyStatus)_ |  true  |  | Status defines the current status of SecurityPolicy. |


#### SecurityPolicySpec



SecurityPolicySpec defines the desired state of SecurityPolicy.

_Appears in:_
- [SecurityPolicy](#securitypolicy)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `targetRef` | _[LocalPolicyTargetReferenceWithSectionName](https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1alpha2.LocalPolicyTargetReferenceWithSectionName)_ |  true  |  | TargetRef is the name of the resource this policy is being attached to.<br />This policy and the TargetRef MUST be in the same namespace for this<br />Policy to have effect<br />Deprecated: use targetRefs/targetSelectors instead |
| `targetRefs` | _[LocalPolicyTargetReferenceWithSectionName](https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1alpha2.LocalPolicyTargetReferenceWithSectionName) array_ |  true  |  | TargetRefs are the names of the Gateway resources this policy<br />is being attached to. |
| `targetSelectors` | _[TargetSelector](#targetselector) array_ |  true  |  | TargetSelectors allow targeting resources for this policy based on labels |
| `apiKeyAuth` | _[APIKeyAuth](#apikeyauth)_ |  false  |  | APIKeyAuth defines the configuration for the API Key Authentication. |
| `cors` | _[CORS](#cors)_ |  false  |  | CORS defines the configuration for Cross-Origin Resource Sharing (CORS). |
| `basicAuth` | _[BasicAuth](#basicauth)_ |  false  |  | BasicAuth defines the configuration for the HTTP Basic Authentication. |
| `jwt` | _[JWT](#jwt)_ |  false  |  | JWT defines the configuration for JSON Web Token (JWT) authentication. |
| `oidc` | _[OIDC](#oidc)_ |  false  |  | OIDC defines the configuration for the OpenID Connect (OIDC) authentication. |
| `extAuth` | _[ExtAuth](#extauth)_ |  false  |  | ExtAuth defines the configuration for External Authorization. |
| `authorization` | _[Authorization](#authorization)_ |  false  |  | Authorization defines the authorization configuration. |


#### ServiceExternalTrafficPolicy

_Underlying type:_ _string_

ServiceExternalTrafficPolicy describes how nodes distribute service traffic they
receive on one of the Service's "externally-facing" addresses (NodePorts, ExternalIPs,
and LoadBalancer IPs.

_Appears in:_
- [KubernetesServiceSpec](#kubernetesservicespec)

| Value | Description |
| ----- | ----------- |
| `Cluster` | ServiceExternalTrafficPolicyCluster routes traffic to all endpoints.<br /> | 
| `Local` | ServiceExternalTrafficPolicyLocal preserves the source IP of the traffic by<br />routing only to endpoints on the same node as the traffic was received on<br />(dropping the traffic if there are no local endpoints).<br /> | 


#### ServiceType

_Underlying type:_ _string_

ServiceType string describes ingress methods for a service

_Appears in:_
- [KubernetesServiceSpec](#kubernetesservicespec)

| Value | Description |
| ----- | ----------- |
| `ClusterIP` | ServiceTypeClusterIP means a service will only be accessible inside the<br />cluster, via the cluster IP.<br /> | 
| `LoadBalancer` | ServiceTypeLoadBalancer means a service will be exposed via an<br />external load balancer (if the cloud provider supports it).<br /> | 
| `NodePort` | ServiceTypeNodePort means a service will be exposed on each Kubernetes Node<br />at a static Port, common across all Nodes.<br /> | 


#### Session



Session defines settings related to TLS session management.

_Appears in:_
- [ClientTLSSettings](#clienttlssettings)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `resumption` | _[SessionResumption](#sessionresumption)_ |  false  |  | Resumption determines the proxy's supported TLS session resumption option.<br />By default, Envoy Gateway does not enable session resumption. Use sessionResumption to<br />enable stateful and stateless session resumption. Users should consider security impacts<br />of different resumption methods. Performance gains from resumption are diminished when<br />Envoy proxy is deployed with more than one replica. |


#### SessionResumption



SessionResumption defines supported tls session resumption methods and their associated configuration.

_Appears in:_
- [Session](#session)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `stateless` | _[StatelessTLSSessionResumption](#statelesstlssessionresumption)_ |  false  |  | Stateless defines setting for stateless (session-ticket based) session resumption |
| `stateful` | _[StatefulTLSSessionResumption](#statefultlssessionresumption)_ |  false  |  | Stateful defines setting for stateful (session-id based) session resumption |


#### ShutdownConfig



ShutdownConfig defines configuration for graceful envoy shutdown process.

_Appears in:_
- [EnvoyProxySpec](#envoyproxyspec)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `drainTimeout` | _[Duration](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#duration-v1-meta)_ |  false  |  | DrainTimeout defines the graceful drain timeout. This should be less than the pod's terminationGracePeriodSeconds.<br />If unspecified, defaults to 60 seconds. |
| `minDrainDuration` | _[Duration](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#duration-v1-meta)_ |  false  |  | MinDrainDuration defines the minimum drain duration allowing time for endpoint deprogramming to complete.<br />If unspecified, defaults to 10 seconds. |


#### ShutdownManager



ShutdownManager defines the configuration for the shutdown manager.

_Appears in:_
- [EnvoyGatewayKubernetesProvider](#envoygatewaykubernetesprovider)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `image` | _string_ |  true  |  | Image specifies the ShutdownManager container image to be used, instead of the default image. |


#### SlowStart



SlowStart defines the configuration related to the slow start load balancer policy.

_Appears in:_
- [LoadBalancer](#loadbalancer)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `window` | _[Duration](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#duration-v1-meta)_ |  true  |  | Window defines the duration of the warm up period for newly added host.<br />During slow start window, traffic sent to the newly added hosts will gradually increase.<br />Currently only supports linear growth of traffic. For additional details,<br />see https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/cluster/v3/cluster.proto#config-cluster-v3-cluster-slowstartconfig |


#### SourceMatch





_Appears in:_
- [RateLimitSelectCondition](#ratelimitselectcondition)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `type` | _[SourceMatchType](#sourcematchtype)_ |  false  | Exact |  |
| `value` | _string_ |  true  |  | Value is the IP CIDR that represents the range of Source IP Addresses of the client.<br />These could also be the intermediate addresses through which the request has flown through and is part of the  `X-Forwarded-For` header.<br />For example, `192.168.0.1/32`, `192.168.0.0/24`, `001:db8::/64`. |


#### SourceMatchType

_Underlying type:_ _string_



_Appears in:_
- [SourceMatch](#sourcematch)

| Value | Description |
| ----- | ----------- |
| `Exact` | SourceMatchExact All IP Addresses within the specified Source IP CIDR are treated as a single client selector<br />and share the same rate limit bucket.<br /> | 
| `Distinct` | SourceMatchDistinct Each IP Address within the specified Source IP CIDR is treated as a distinct client selector<br />and uses a separate rate limit bucket/counter.<br /> | 


#### StatefulTLSSessionResumption



StatefulTLSSessionResumption defines the stateful (session-id based) type of TLS session resumption.
Note: When Envoy Proxy is deployed with more than one replica, session caches are not synchronized
between instances, possibly leading to resumption failures.
Envoy does not re-validate client certificates upon session resumption.
https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/route/v3/route_components.proto#config-route-v3-routematch-tlscontextmatchoptions

_Appears in:_
- [SessionResumption](#sessionresumption)



#### StatelessTLSSessionResumption



StatelessTLSSessionResumption defines the stateless (session-ticket based) type of TLS session resumption.
Note: When Envoy Proxy is deployed with more than one replica, session ticket encryption keys are not
synchronized between instances, possibly leading to resumption failures.
In-memory session ticket encryption keys are rotated every 48 hours.
https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/transport_sockets/tls/v3/common.proto#extensions-transport-sockets-tls-v3-tlssessionticketkeys
https://commondatastorage.googleapis.com/chromium-boringssl-docs/ssl.h.html#Session-tickets

_Appears in:_
- [SessionResumption](#sessionresumption)



#### StatusCodeMatch



StatusCodeMatch defines the configuration for matching a status code.

_Appears in:_
- [CustomResponseMatch](#customresponsematch)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `type` | _[StatusCodeValueType](#statuscodevaluetype)_ |  true  | Value | Type is the type of value.<br />Valid values are Value and Range, default is Value. |
| `value` | _integer_ |  false  |  | Value contains the value of the status code. |
| `range` | _[StatusCodeRange](#statuscoderange)_ |  false  |  | Range contains the range of status codes. |


#### StatusCodeRange



StatusCodeRange defines the configuration for define a range of status codes.

_Appears in:_
- [StatusCodeMatch](#statuscodematch)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `start` | _integer_ |  true  |  | Start of the range, including the start value. |
| `end` | _integer_ |  true  |  | End of the range, including the end value. |


#### StatusCodeValueType

_Underlying type:_ _string_

StatusCodeValueType defines the types of values for the status code match supported by Envoy Gateway.

_Appears in:_
- [StatusCodeMatch](#statuscodematch)

| Value | Description |
| ----- | ----------- |
| `Value` | StatusCodeValueTypeValue defines the "Value" status code match type.<br /> | 
| `Range` | StatusCodeValueTypeRange defines the "Range" status code match type.<br /> | 


#### StringMatch



StringMatch defines how to match any strings.
This is a general purpose match condition that can be used by other EG APIs
that need to match against a string.

_Appears in:_
- [OIDCDenyRedirectHeader](#oidcdenyredirectheader)
- [ProxyMetrics](#proxymetrics)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `type` | _[StringMatchType](#stringmatchtype)_ |  false  | Exact | Type specifies how to match against a string. |
| `value` | _string_ |  true  |  | Value specifies the string value that the match must have. |


#### StringMatchType

_Underlying type:_ _string_

StringMatchType specifies the semantics of how a string value should be compared.
Valid MatchType values are "Exact", "Prefix", "Suffix", "RegularExpression".

_Appears in:_
- [OIDCDenyRedirectHeader](#oidcdenyredirectheader)
- [StringMatch](#stringmatch)

| Value | Description |
| ----- | ----------- |
| `Exact` | StringMatchExact :the input string must match exactly the match value.<br /> | 
| `Prefix` | StringMatchPrefix :the input string must start with the match value.<br /> | 
| `Suffix` | StringMatchSuffix :the input string must end with the match value.<br /> | 
| `RegularExpression` | StringMatchRegularExpression :The input string must match the regular expression<br />specified in the match value.<br />The regex string must adhere to the syntax documented in<br />https://github.com/google/re2/wiki/Syntax.<br /> | 


#### TCPActiveHealthChecker



TCPActiveHealthChecker defines the settings of tcp health check.

_Appears in:_
- [ActiveHealthCheck](#activehealthcheck)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `send` | _[ActiveHealthCheckPayload](#activehealthcheckpayload)_ |  false  |  | Send defines the request payload. |
| `receive` | _[ActiveHealthCheckPayload](#activehealthcheckpayload)_ |  false  |  | Receive defines the expected response payload. |


#### TCPClientTimeout



TCPClientTimeout only provides timeout configuration on the listener whose protocol is TCP or TLS.

_Appears in:_
- [ClientTimeout](#clienttimeout)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `idleTimeout` | _[Duration](https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1.Duration)_ |  false  |  | IdleTimeout for a TCP connection. Idle time is defined as a period in which there are no<br />bytes sent or received on either the upstream or downstream connection.<br />Default: 1 hour. |


#### TCPKeepalive



TCPKeepalive define the TCP Keepalive configuration.

_Appears in:_
- [BackendTrafficPolicySpec](#backendtrafficpolicyspec)
- [ClientTrafficPolicySpec](#clienttrafficpolicyspec)
- [ClusterSettings](#clustersettings)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `probes` | _integer_ |  false  |  | The total number of unacknowledged probes to send before deciding<br />the connection is dead.<br />Defaults to 9. |
| `idleTime` | _[Duration](https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1.Duration)_ |  false  |  | The duration a connection needs to be idle before keep-alive<br />probes start being sent.<br />The duration format is<br />Defaults to `7200s`. |
| `interval` | _[Duration](https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1.Duration)_ |  false  |  | The duration between keep-alive probes.<br />Defaults to `75s`. |


#### TCPTimeout





_Appears in:_
- [Timeout](#timeout)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `connectTimeout` | _[Duration](https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1.Duration)_ |  false  |  | The timeout for network connection establishment, including TCP and TLS handshakes.<br />Default: 10 seconds. |


#### TLSSettings





_Appears in:_
- [BackendTLSConfig](#backendtlsconfig)
- [ClientTLSSettings](#clienttlssettings)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `minVersion` | _[TLSVersion](#tlsversion)_ |  false  |  | Min specifies the minimal TLS protocol version to allow.<br />The default is TLS 1.2 if this is not specified. |
| `maxVersion` | _[TLSVersion](#tlsversion)_ |  false  |  | Max specifies the maximal TLS protocol version to allow<br />The default is TLS 1.3 if this is not specified. |
| `ciphers` | _string array_ |  false  |  | Ciphers specifies the set of cipher suites supported when<br />negotiating TLS 1.0 - 1.2. This setting has no effect for TLS 1.3.<br />In non-FIPS Envoy Proxy builds the default cipher list is:<br />- [ECDHE-ECDSA-AES128-GCM-SHA256\|ECDHE-ECDSA-CHACHA20-POLY1305]<br />- [ECDHE-RSA-AES128-GCM-SHA256\|ECDHE-RSA-CHACHA20-POLY1305]<br />- ECDHE-ECDSA-AES256-GCM-SHA384<br />- ECDHE-RSA-AES256-GCM-SHA384<br />In builds using BoringSSL FIPS the default cipher list is:<br />- ECDHE-ECDSA-AES128-GCM-SHA256<br />- ECDHE-RSA-AES128-GCM-SHA256<br />- ECDHE-ECDSA-AES256-GCM-SHA384<br />- ECDHE-RSA-AES256-GCM-SHA384 |
| `ecdhCurves` | _string array_ |  false  |  | ECDHCurves specifies the set of supported ECDH curves.<br />In non-FIPS Envoy Proxy builds the default curves are:<br />- X25519<br />- P-256<br />In builds using BoringSSL FIPS the default curve is:<br />- P-256 |
| `signatureAlgorithms` | _string array_ |  false  |  | SignatureAlgorithms specifies which signature algorithms the listener should<br />support. |
| `alpnProtocols` | _[ALPNProtocol](#alpnprotocol) array_ |  false  |  | ALPNProtocols supplies the list of ALPN protocols that should be<br />exposed by the listener or used by the proxy to connect to the backend.<br />Defaults:<br />1. HTTPS Routes: h2 and http/1.1 are enabled in listener context.<br />2. Other Routes: ALPN is disabled.<br />3. Backends: proxy uses the appropriate ALPN options for the backend protocol.<br />When an empty list is provided, the ALPN TLS extension is disabled.<br />Supported values are:<br />- http/1.0<br />- http/1.1<br />- h2 |


#### TLSVersion

_Underlying type:_ _string_

TLSVersion specifies the TLS version

_Appears in:_
- [BackendTLSConfig](#backendtlsconfig)
- [ClientTLSSettings](#clienttlssettings)
- [TLSSettings](#tlssettings)

| Value | Description |
| ----- | ----------- |
| `Auto` | TLSAuto allows Envoy to choose the optimal TLS Version<br /> | 
| `1.0` | TLS1.0 specifies TLS version 1.0<br /> | 
| `1.1` | TLS1.1 specifies TLS version 1.1<br /> | 
| `1.2` | TLSv1.2 specifies TLS version 1.2<br /> | 
| `1.3` | TLSv1.3 specifies TLS version 1.3<br /> | 


#### TargetSelector





_Appears in:_
- [BackendTrafficPolicySpec](#backendtrafficpolicyspec)
- [ClientTrafficPolicySpec](#clienttrafficpolicyspec)
- [EnvoyExtensionPolicySpec](#envoyextensionpolicyspec)
- [PolicyTargetReferences](#policytargetreferences)
- [SecurityPolicySpec](#securitypolicyspec)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `group` | _[Group](#group)_ |  true  | gateway.networking.k8s.io | Group is the group that this selector targets. Defaults to gateway.networking.k8s.io |
| `kind` | _[Kind](#kind)_ |  true  |  | Kind is the resource kind that this selector targets. |
| `matchLabels` | _object (keys:string, values:string)_ |  false  |  | MatchLabels are the set of label selectors for identifying the targeted resource |
| `matchExpressions` | _[LabelSelectorRequirement](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#labelselectorrequirement-v1-meta) array_ |  false  |  | MatchExpressions is a list of label selector requirements. The requirements are ANDed. |


#### Timeout



Timeout defines configuration for timeouts related to connections.

_Appears in:_
- [BackendTrafficPolicySpec](#backendtrafficpolicyspec)
- [ClusterSettings](#clustersettings)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `tcp` | _[TCPTimeout](#tcptimeout)_ |  false  |  | Timeout settings for TCP. |
| `http` | _[HTTPTimeout](#httptimeout)_ |  false  |  | Timeout settings for HTTP. |


#### Tracing



Tracing defines the configuration for tracing.

_Appears in:_
- [BackendTelemetry](#backendtelemetry)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `samplingFraction` | _[Fraction](https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1.Fraction)_ |  false  |  | SamplingFraction represents the fraction of requests that should be<br />selected for tracing if no prior sampling decision has been made.<br />This will take precedence over sampling fraction on EnvoyProxy if set. |
| `customTags` | _object (keys:string, values:[CustomTag](#customtag))_ |  false  |  | CustomTags defines the custom tags to add to each span.<br />If provider is kubernetes, pod name and namespace are added by default. |


#### TracingProvider



TracingProvider defines the tracing provider configuration.

_Appears in:_
- [ProxyTracing](#proxytracing)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `backendRef` | _[BackendObjectReference](https://gateway-api.sigs.k8s.io/references/spec/#gateway.networking.k8s.io/v1.BackendObjectReference)_ |  false  |  | BackendRef references a Kubernetes object that represents the<br />backend server to which the authorization request will be sent.<br />Deprecated: Use BackendRefs instead. |
| `backendRefs` | _[BackendRef](#backendref) array_ |  false  |  | BackendRefs references a Kubernetes object that represents the<br />backend server to which the authorization request will be sent. |
| `backendSettings` | _[ClusterSettings](#clustersettings)_ |  false  |  | BackendSettings holds configuration for managing the connection<br />to the backend. |
| `type` | _[TracingProviderType](#tracingprovidertype)_ |  true  | OpenTelemetry | Type defines the tracing provider type. |
| `host` | _string_ |  false  |  | Host define the provider service hostname.<br />Deprecated: Use BackendRefs instead. |
| `port` | _integer_ |  false  | 4317 | Port defines the port the provider service is exposed on.<br />Deprecated: Use BackendRefs instead. |
| `zipkin` | _[ZipkinTracingProvider](#zipkintracingprovider)_ |  false  |  | Zipkin defines the Zipkin tracing provider configuration |


#### TracingProviderType

_Underlying type:_ _string_



_Appears in:_
- [TracingProvider](#tracingprovider)

| Value | Description |
| ----- | ----------- |
| `OpenTelemetry` |  | 
| `OpenTelemetry` |  | 
| `Zipkin` |  | 
| `Datadog` |  | 


#### TriggerEnum

_Underlying type:_ _string_

TriggerEnum specifies the conditions that trigger retries.

_Appears in:_
- [RetryOn](#retryon)

| Value | Description |
| ----- | ----------- |
| `5xx` | The upstream server responds with any 5xx response code, or does not respond at all (disconnect/reset/read timeout).<br />Includes connect-failure and refused-stream.<br /> | 
| `gateway-error` | The response is a gateway error (502,503 or 504).<br /> | 
| `reset` | The upstream server does not respond at all (disconnect/reset/read timeout.)<br /> | 
| `connect-failure` | Connection failure to the upstream server (connect timeout, etc.). (Included in *5xx*)<br /> | 
| `retriable-4xx` | The upstream server responds with a retriable 4xx response code.<br />Currently, the only response code in this category is 409.<br /> | 
| `refused-stream` | The upstream server resets the stream with a REFUSED_STREAM error code.<br /> | 
| `retriable-status-codes` | The upstream server responds with any response code matching one defined in the RetriableStatusCodes.<br /> | 
| `cancelled` | The gRPC status code in the response headers is “cancelled”.<br /> | 
| `deadline-exceeded` | The gRPC status code in the response headers is “deadline-exceeded”.<br /> | 
| `internal` | The gRPC status code in the response headers is “internal”.<br /> | 
| `resource-exhausted` | The gRPC status code in the response headers is “resource-exhausted”.<br /> | 
| `unavailable` | The gRPC status code in the response headers is “unavailable”.<br /> | 


#### UnixSocket



UnixSocket describes TCP/UDP unix domain socket address, corresponding to Envoy's Pipe
https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/core/v3/address.proto#config-core-v3-pipe

_Appears in:_
- [BackendEndpoint](#backendendpoint)
- [ExtensionService](#extensionservice)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `path` | _string_ |  true  |  | Path defines the unix domain socket path of the backend endpoint.<br />The path length must not exceed 108 characters. |


#### Wasm



Wasm defines a Wasm extension.

Note: at the moment, Envoy Gateway does not support configuring Wasm runtime.
v8 is used as the VM runtime for the Wasm extensions.

_Appears in:_
- [EnvoyExtensionPolicySpec](#envoyextensionpolicyspec)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `name` | _string_ |  false  |  | Name is a unique name for this Wasm extension. It is used to identify the<br />Wasm extension if multiple extensions are handled by the same vm_id and root_id.<br />It's also used for logging/debugging.<br />If not specified, EG will generate a unique name for the Wasm extension. |
| `rootID` | _string_ |  true  |  | RootID is a unique ID for a set of extensions in a VM which will share a<br />RootContext and Contexts if applicable (e.g., an Wasm HttpFilter and an Wasm AccessLog).<br />If left blank, all extensions with a blank root_id with the same vm_id will share Context(s).<br />Note: RootID must match the root_id parameter used to register the Context in the Wasm code. |
| `code` | _[WasmCodeSource](#wasmcodesource)_ |  true  |  | Code is the Wasm code for the extension. |
| `config` | _[JSON](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#json-v1-apiextensions-k8s-io)_ |  false  |  | Config is the configuration for the Wasm extension.<br />This configuration will be passed as a JSON string to the Wasm extension. |
| `failOpen` | _boolean_ |  false  | false | FailOpen is a switch used to control the behavior when a fatal error occurs<br />during the initialization or the execution of the Wasm extension.<br />If FailOpen is set to true, the system bypasses the Wasm extension and<br />allows the traffic to pass through. Otherwise, if it is set to false or<br />not set (defaulting to false), the system blocks the traffic and returns<br />an HTTP 5xx error. |
| `env` | _[WasmEnv](#wasmenv)_ |  false  |  | Env configures the environment for the Wasm extension |


#### WasmCodeSource



WasmCodeSource defines the source of the Wasm code.

_Appears in:_
- [Wasm](#wasm)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `type` | _[WasmCodeSourceType](#wasmcodesourcetype)_ |  true  |  | Type is the type of the source of the Wasm code.<br />Valid WasmCodeSourceType values are "HTTP" or "Image". |
| `http` | _[HTTPWasmCodeSource](#httpwasmcodesource)_ |  false  |  | HTTP is the HTTP URL containing the Wasm code.<br />Note that the HTTP server must be accessible from the Envoy proxy. |
| `image` | _[ImageWasmCodeSource](#imagewasmcodesource)_ |  false  |  | Image is the OCI image containing the Wasm code.<br />Note that the image must be accessible from the Envoy Gateway. |
| `pullPolicy` | _[ImagePullPolicy](#imagepullpolicy)_ |  false  |  | PullPolicy is the policy to use when pulling the Wasm module by either the HTTP or Image source.<br />This field is only applicable when the SHA256 field is not set.<br />If not specified, the default policy is IfNotPresent except for OCI images whose tag is latest.<br />Note: EG does not update the Wasm module every time an Envoy proxy requests<br />the Wasm module even if the pull policy is set to Always.<br />It only updates the Wasm module when the EnvoyExtension resource version changes. |


#### WasmCodeSourceTLSConfig



WasmCodeSourceTLSConfig defines the TLS configuration when connecting to the Wasm code source.

_Appears in:_
- [HTTPWasmCodeSource](#httpwasmcodesource)
- [ImageWasmCodeSource](#imagewasmcodesource)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `caCertificateRef` | _[SecretObjectReference](https://gateway-api.sigs.k8s.io/references/spec/#gateway.networking.k8s.io/v1.SecretObjectReference)_ |  true  |  | CACertificateRef contains a references to<br />Kubernetes objects that contain TLS certificates of<br />the Certificate Authorities that can be used<br />as a trust anchor to validate the certificates presented by the Wasm code source.<br />Kubernetes ConfigMap and Kubernetes Secret are supported.<br />Note: The ConfigMap or Secret must be in the same namespace as the EnvoyExtensionPolicy. |


#### WasmCodeSourceType

_Underlying type:_ _string_

WasmCodeSourceType specifies the types of sources for the Wasm code.

_Appears in:_
- [WasmCodeSource](#wasmcodesource)

| Value | Description |
| ----- | ----------- |
| `HTTP` | HTTPWasmCodeSourceType allows the user to specify the Wasm code in an HTTP URL.<br /> | 
| `Image` | ImageWasmCodeSourceType allows the user to specify the Wasm code in an OCI image.<br /> | 


#### WasmEnv



WasmEnv defines the environment variables for the VM of a Wasm extension

_Appears in:_
- [Wasm](#wasm)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `hostKeys` | _string array_ |  false  |  | HostKeys is a list of keys for environment variables from the host envoy process<br />that should be passed into the Wasm VM. This is useful for passing secrets to to Wasm extensions. |


#### WithUnderscoresAction

_Underlying type:_ _string_

WithUnderscoresAction configures the action to take when an HTTP header with underscores
is encountered.

_Appears in:_
- [HeaderSettings](#headersettings)

| Value | Description |
| ----- | ----------- |
| `Allow` | WithUnderscoresActionAllow allows headers with underscores to be passed through.<br /> | 
| `RejectRequest` | WithUnderscoresActionRejectRequest rejects the client request. HTTP/1 requests are rejected with<br />the 400 status. HTTP/2 requests end with the stream reset.<br /> | 
| `DropHeader` | WithUnderscoresActionDropHeader drops the client header with name containing underscores. The header<br />is dropped before the filter chain is invoked and as such filters will not see<br />dropped headers.<br /> | 


#### XDSTranslatorHook

_Underlying type:_ _string_

XDSTranslatorHook defines the types of hooks that an Envoy Gateway extension may support
for the xds-translator

_Appears in:_
- [XDSTranslatorHooks](#xdstranslatorhooks)

| Value | Description |
| ----- | ----------- |
| `VirtualHost` |  | 
| `Route` |  | 
| `HTTPListener` |  | 
| `Translation` |  | 


#### XDSTranslatorHooks



XDSTranslatorHooks contains all the pre and post hooks for the xds-translator runner.

_Appears in:_
- [ExtensionHooks](#extensionhooks)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `pre` | _[XDSTranslatorHook](#xdstranslatorhook) array_ |  true  |  |  |
| `post` | _[XDSTranslatorHook](#xdstranslatorhook) array_ |  true  |  |  |


#### XFCCCertData

_Underlying type:_ _string_

XFCCCertData specifies the fields in the client certificate to be forwarded in the XFCC header.

_Appears in:_
- [XForwardedClientCert](#xforwardedclientcert)

| Value | Description |
| ----- | ----------- |
| `Subject` | XFCCCertDataSubject is the Subject field of the current client certificate.<br /> | 
| `Cert` | XFCCCertDataCert is the entire client certificate in URL encoded PEM format.<br /> | 
| `Chain` | XFCCCertDataChain is the entire client certificate chain (including the leaf certificate) in URL encoded PEM format.<br /> | 
| `DNS` | XFCCCertDataDNS is the DNS type Subject Alternative Name field of the current client certificate.<br /> | 
| `URI` | XFCCCertDataURI is the URI type Subject Alternative Name field of the current client certificate.<br /> | 


#### XFCCForwardMode

_Underlying type:_ _string_

XFCCForwardMode defines how XFCC header is handled by Envoy Proxy.

_Appears in:_
- [XForwardedClientCert](#xforwardedclientcert)

| Value | Description |
| ----- | ----------- |
| `Sanitize` | XFCCForwardModeSanitize removes the XFCC header from the request. This is the default mode.<br /> | 
| `ForwardOnly` | XFCCForwardModeForwardOnly forwards the XFCC header in the request if the client connection is mTLS.<br /> | 
| `AppendForward` | XFCCForwardModeAppendForward appends the client certificate information to the request’s XFCC header and forward it if the client connection is mTLS.<br /> | 
| `SanitizeSet` | XFCCForwardModeSanitizeSet resets the XFCC header with the client certificate information and forward it if the client connection is mTLS.<br />The existing certificate information in the XFCC header is removed.<br /> | 
| `AlwaysForwardOnly` | XFCCForwardModeAlwaysForwardOnly always forwards the XFCC header in the request, regardless of whether the client connection is mTLS.<br /> | 


#### XForwardedClientCert



XForwardedClientCert configures how Envoy Proxy handle the x-forwarded-client-cert (XFCC) HTTP header.

_Appears in:_
- [HeaderSettings](#headersettings)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `mode` | _[XFCCForwardMode](#xfccforwardmode)_ |  false  |  | Mode defines how XFCC header is handled by Envoy Proxy.<br />If not set, the default mode is `Sanitize`. |
| `certDetailsToAdd` | _[XFCCCertData](#xfcccertdata) array_ |  false  |  | CertDetailsToAdd specifies the fields in the client certificate to be forwarded in the XFCC header.<br />Hash(the SHA 256 digest of the current client certificate) and By(the Subject Alternative Name)<br />are always included if the client certificate is forwarded.<br />This field is only applicable when the mode is set to `AppendForward` or<br />`SanitizeSet` and the client connection is mTLS. |


#### XForwardedForSettings



XForwardedForSettings provides configuration for using X-Forwarded-For headers for determining the client IP address.
Refer to https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_conn_man/headers#x-forwarded-for
for more details.

_Appears in:_
- [ClientIPDetectionSettings](#clientipdetectionsettings)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `numTrustedHops` | _integer_ |  false  |  | NumTrustedHops controls the number of additional ingress proxy hops from the right side of XFF HTTP<br />headers to trust when determining the origin client's IP address.<br />Only one of NumTrustedHops and TrustedCIDRs must be set. |
| `trustedCIDRs` | _[CIDR](#cidr) array_ |  false  |  | TrustedCIDRs is a list of CIDR ranges to trust when evaluating<br />the remote IP address to determine the original client’s IP address.<br />When the remote IP address matches a trusted CIDR and the x-forwarded-for header was sent,<br />each entry in the x-forwarded-for header is evaluated from right to left<br />and the first public non-trusted address is used as the original client address.<br />If all addresses in x-forwarded-for are within the trusted list, the first (leftmost) entry is used.<br />Only one of NumTrustedHops and TrustedCIDRs must be set. |


#### ZipkinTracingProvider



ZipkinTracingProvider defines the Zipkin tracing provider configuration.

_Appears in:_
- [TracingProvider](#tracingprovider)

| Field | Type | Required | Default | Description |
| ---   | ---  | ---      | ---     | ---         |
| `enable128BitTraceId` | _boolean_ |  false  |  | Enable128BitTraceID determines whether a 128bit trace id will be used<br />when creating a new trace instance. If set to false, a 64bit trace<br />id will be used. |
| `disableSharedSpanContext` | _boolean_ |  false  |  | DisableSharedSpanContext determines whether the default Envoy behaviour of<br />client and server spans sharing the same span context should be disabled. |


