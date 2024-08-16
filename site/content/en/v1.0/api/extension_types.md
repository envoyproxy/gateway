+++
title = "API Reference"
+++


## Packages
- [gateway.envoyproxy.io/v1alpha1](#gatewayenvoyproxyiov1alpha1)


## gateway.envoyproxy.io/v1alpha1

Package v1alpha1 contains API schema definitions for the gateway.envoyproxy.io
API group.


### Resource Types
- [BackendTrafficPolicy](#backendtrafficpolicy)
- [BackendTrafficPolicyList](#backendtrafficpolicylist)
- [ClientTrafficPolicy](#clienttrafficpolicy)
- [ClientTrafficPolicyList](#clienttrafficpolicylist)
- [EnvoyGateway](#envoygateway)
- [EnvoyPatchPolicy](#envoypatchpolicy)
- [EnvoyPatchPolicyList](#envoypatchpolicylist)
- [EnvoyProxy](#envoyproxy)
- [SecurityPolicy](#securitypolicy)
- [SecurityPolicyList](#securitypolicylist)



#### ALPNProtocol

_Underlying type:_ _string_

ALPNProtocol specifies the protocol to be negotiated using ALPN

_Appears in:_
- [TLSSettings](#tlssettings)



#### ActiveHealthCheck



ActiveHealthCheck defines the active health check configuration. EG supports various types of active health checking including HTTP, TCP.

_Appears in:_
- [HealthCheck](#healthcheck)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `timeout` | _[Duration](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#duration-v1-meta)_ |  false  | Timeout defines the time to wait for a health check response. |
| `interval` | _[Duration](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#duration-v1-meta)_ |  false  | Interval defines the time between active health checks. |
| `unhealthyThreshold` | _integer_ |  false  | UnhealthyThreshold defines the number of unhealthy health checks required before a backend host is marked unhealthy. |
| `healthyThreshold` | _integer_ |  false  | HealthyThreshold defines the number of healthy health checks required before a backend host is marked healthy. |
| `type` | _[ActiveHealthCheckerType](#activehealthcheckertype)_ |  true  | Type defines the type of health checker. |
| `http` | _[HTTPActiveHealthChecker](#httpactivehealthchecker)_ |  false  | HTTP defines the configuration of http health checker. It's required while the health checker type is HTTP. |
| `tcp` | _[TCPActiveHealthChecker](#tcpactivehealthchecker)_ |  false  | TCP defines the configuration of tcp health checker. It's required while the health checker type is TCP. |


#### ActiveHealthCheckPayload



ActiveHealthCheckPayload defines the encoding of the payload bytes in the payload.

_Appears in:_
- [HTTPActiveHealthChecker](#httpactivehealthchecker)
- [TCPActiveHealthChecker](#tcpactivehealthchecker)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `type` | _[ActiveHealthCheckPayloadType](#activehealthcheckpayloadtype)_ |  true  | Type defines the type of the payload. |
| `text` | _string_ |  false  | Text payload in plain text. |
| `binary` | _integer array_ |  false  | Binary payload base64 encoded. |


#### ActiveHealthCheckPayloadType

_Underlying type:_ _string_

ActiveHealthCheckPayloadType is the type of the payload.

_Appears in:_
- [ActiveHealthCheckPayload](#activehealthcheckpayload)



#### ActiveHealthCheckerType

_Underlying type:_ _string_

ActiveHealthCheckerType is the type of health checker.

_Appears in:_
- [ActiveHealthCheck](#activehealthcheck)



#### BackOffPolicy





_Appears in:_
- [PerRetryPolicy](#perretrypolicy)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `baseInterval` | _[Duration](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#duration-v1-meta)_ |  true  | BaseInterval is the base interval between retries. |
| `maxInterval` | _[Duration](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#duration-v1-meta)_ |  false  | MaxInterval is the maximum interval between retries. This parameter is optional, but must be greater than or equal to the base_interval if set. The default is 10 times the base_interval |


#### BackendTrafficPolicy



BackendTrafficPolicy allows the user to configure the behavior of the connection between the Envoy Proxy listener and the backend service.

_Appears in:_
- [BackendTrafficPolicyList](#backendtrafficpolicylist)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `apiVersion` | _string_ | |`gateway.envoyproxy.io/v1alpha1`
| `kind` | _string_ | |`BackendTrafficPolicy`
| `metadata` | _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#objectmeta-v1-meta)_ |  true  | Refer to Kubernetes API documentation for fields of `metadata`. |
| `spec` | _[BackendTrafficPolicySpec](#backendtrafficpolicyspec)_ |  true  | spec defines the desired state of BackendTrafficPolicy. |


#### BackendTrafficPolicyList



BackendTrafficPolicyList contains a list of BackendTrafficPolicy resources.



| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `apiVersion` | _string_ | |`gateway.envoyproxy.io/v1alpha1`
| `kind` | _string_ | |`BackendTrafficPolicyList`
| `metadata` | _[ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#listmeta-v1-meta)_ |  true  | Refer to Kubernetes API documentation for fields of `metadata`. |
| `items` | _[BackendTrafficPolicy](#backendtrafficpolicy) array_ |  true  |  |


#### BackendTrafficPolicySpec



spec defines the desired state of BackendTrafficPolicy.

_Appears in:_
- [BackendTrafficPolicy](#backendtrafficpolicy)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `targetRef` | _[PolicyTargetReferenceWithSectionName](https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1alpha2.PolicyTargetReferenceWithSectionName)_ |  true  | targetRef is the name of the resource this policy is being attached to. This Policy and the TargetRef MUST be in the same namespace for this Policy to have effect and be applied to the Gateway. |
| `rateLimit` | _[RateLimitSpec](#ratelimitspec)_ |  false  | RateLimit allows the user to limit the number of incoming requests to a predefined value based on attributes within the traffic flow. |
| `loadBalancer` | _[LoadBalancer](#loadbalancer)_ |  false  | LoadBalancer policy to apply when routing traffic from the gateway to the backend endpoints |
| `proxyProtocol` | _[ProxyProtocol](#proxyprotocol)_ |  false  | ProxyProtocol enables the Proxy Protocol when communicating with the backend. |
| `tcpKeepalive` | _[TCPKeepalive](#tcpkeepalive)_ |  false  | TcpKeepalive settings associated with the upstream client connection. Disabled by default. |
| `healthCheck` | _[HealthCheck](#healthcheck)_ |  false  | HealthCheck allows gateway to perform active health checking on backends. |
| `faultInjection` | _[FaultInjection](#faultinjection)_ |  false  | FaultInjection defines the fault injection policy to be applied. This configuration can be used to inject delays and abort requests to mimic failure scenarios such as service failures and overloads |
| `circuitBreaker` | _[CircuitBreaker](#circuitbreaker)_ |  false  | Circuit Breaker settings for the upstream connections and requests. If not set, circuit breakers will be enabled with the default thresholds |
| `retry` | _[Retry](#retry)_ |  false  | Retry provides more advanced usage, allowing users to customize the number of retries, retry fallback strategy, and retry triggering conditions. If not set, retry will be disabled. |
| `timeout` | _[Timeout](#timeout)_ |  false  | Timeout settings for the backend connections. |
| `compression` | _[Compression](#compression) array_ |  false  | The compression config for the http streams. |


#### BasicAuth



BasicAuth defines the configuration for 	the HTTP Basic Authentication.

_Appears in:_
- [SecurityPolicySpec](#securitypolicyspec)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `users` | _[SecretObjectReference](https://gateway-api.sigs.k8s.io/references/spec/#gateway.networking.k8s.io/v1.SecretObjectReference)_ |  true  | The Kubernetes secret which contains the username-password pairs in htpasswd format, used to verify user credentials in the "Authorization" header. <br /><br /> This is an Opaque secret. The username-password pairs should be stored in the key ".htpasswd". As the key name indicates, the value needs to be the htpasswd format, for example: "user1:{SHA}hashed_user1_password". Right now, only SHA hash algorithm is supported. Reference to https://httpd.apache.org/docs/2.4/programs/htpasswd.html for more details. <br /><br /> Note: The secret must be in the same namespace as the SecurityPolicy. |


#### BootstrapType

_Underlying type:_ _string_

BootstrapType defines the types of bootstrap supported by Envoy Gateway.

_Appears in:_
- [ProxyBootstrap](#proxybootstrap)



#### CORS



CORS defines the configuration for Cross-Origin Resource Sharing (CORS).

_Appears in:_
- [SecurityPolicySpec](#securitypolicyspec)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `allowOrigins` | _[Origin](#origin) array_ |  true  | AllowOrigins defines the origins that are allowed to make requests. |
| `allowMethods` | _string array_ |  true  | AllowMethods defines the methods that are allowed to make requests. |
| `allowHeaders` | _string array_ |  true  | AllowHeaders defines the headers that are allowed to be sent with requests. |
| `exposeHeaders` | _string array_ |  true  | ExposeHeaders defines the headers that can be exposed in the responses. |
| `maxAge` | _[Duration](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#duration-v1-meta)_ |  true  | MaxAge defines how long the results of a preflight request can be cached. |
| `allowCredentials` | _boolean_ |  true  | AllowCredentials indicates whether a request can include user credentials like cookies, authentication headers, or TLS client certificates. |


#### CircuitBreaker



CircuitBreaker defines the Circuit Breaker configuration.

_Appears in:_
- [BackendTrafficPolicySpec](#backendtrafficpolicyspec)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `maxConnections` | _integer_ |  false  | The maximum number of connections that Envoy will establish to the referenced backend defined within a xRoute rule. |
| `maxPendingRequests` | _integer_ |  false  | The maximum number of pending requests that Envoy will queue to the referenced backend defined within a xRoute rule. |
| `maxParallelRequests` | _integer_ |  false  | The maximum number of parallel requests that Envoy will make to the referenced backend defined within a xRoute rule. |
| `maxParallelRetries` | _integer_ |  false  | The maximum number of parallel retries that Envoy will make to the referenced backend defined within a xRoute rule. |
| `maxRequestsPerConnection` | _integer_ |  false  | The maximum number of requests that Envoy will make over a single connection to the referenced backend defined within a xRoute rule. Default: unlimited. |


#### ClaimToHeader



ClaimToHeader defines a configuration to convert JWT claims into HTTP headers

_Appears in:_
- [JWTProvider](#jwtprovider)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `header` | _string_ |  true  | Header defines the name of the HTTP request header that the JWT Claim will be saved into. |
| `claim` | _string_ |  true  | Claim is the JWT Claim that should be saved into the header : it can be a nested claim of type (eg. "claim.nested.key", "sub"). The nested claim name must use dot "." to separate the JSON name path. |


#### ClientIPDetectionSettings



ClientIPDetectionSettings provides configuration for determining the original client IP address for requests.

_Appears in:_
- [ClientTrafficPolicySpec](#clienttrafficpolicyspec)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `xForwardedFor` | _[XForwardedForSettings](#xforwardedforsettings)_ |  false  | XForwardedForSettings provides configuration for using X-Forwarded-For headers for determining the client IP address. |
| `customHeader` | _[CustomHeaderExtensionSettings](#customheaderextensionsettings)_ |  false  | CustomHeader provides configuration for determining the client IP address for a request based on a trusted custom HTTP header. This uses the the custom_header original IP detection extension. Refer to https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/http/original_ip_detection/custom_header/v3/custom_header.proto for more details. |


#### ClientTimeout





_Appears in:_
- [ClientTrafficPolicySpec](#clienttrafficpolicyspec)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `http` | _[HTTPClientTimeout](#httpclienttimeout)_ |  false  | Timeout settings for HTTP. |


#### ClientTrafficPolicy



ClientTrafficPolicy allows the user to configure the behavior of the connection between the downstream client and Envoy Proxy listener.

_Appears in:_
- [ClientTrafficPolicyList](#clienttrafficpolicylist)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `apiVersion` | _string_ | |`gateway.envoyproxy.io/v1alpha1`
| `kind` | _string_ | |`ClientTrafficPolicy`
| `metadata` | _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#objectmeta-v1-meta)_ |  true  | Refer to Kubernetes API documentation for fields of `metadata`. |
| `spec` | _[ClientTrafficPolicySpec](#clienttrafficpolicyspec)_ |  true  | Spec defines the desired state of ClientTrafficPolicy. |


#### ClientTrafficPolicyList



ClientTrafficPolicyList contains a list of ClientTrafficPolicy resources.



| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `apiVersion` | _string_ | |`gateway.envoyproxy.io/v1alpha1`
| `kind` | _string_ | |`ClientTrafficPolicyList`
| `metadata` | _[ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#listmeta-v1-meta)_ |  true  | Refer to Kubernetes API documentation for fields of `metadata`. |
| `items` | _[ClientTrafficPolicy](#clienttrafficpolicy) array_ |  true  |  |


#### ClientTrafficPolicySpec



ClientTrafficPolicySpec defines the desired state of ClientTrafficPolicy.

_Appears in:_
- [ClientTrafficPolicy](#clienttrafficpolicy)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `targetRef` | _[PolicyTargetReferenceWithSectionName](https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1alpha2.PolicyTargetReferenceWithSectionName)_ |  true  | TargetRef is the name of the Gateway resource this policy is being attached to. This Policy and the TargetRef MUST be in the same namespace for this Policy to have effect and be applied to the Gateway. TargetRef |
| `tcpKeepalive` | _[TCPKeepalive](#tcpkeepalive)_ |  false  | TcpKeepalive settings associated with the downstream client connection. If defined, sets SO_KEEPALIVE on the listener socket to enable TCP Keepalives. Disabled by default. |
| `enableProxyProtocol` | _boolean_ |  false  | EnableProxyProtocol interprets the ProxyProtocol header and adds the Client Address into the X-Forwarded-For header. Note Proxy Protocol must be present when this field is set, else the connection is closed. |
| `clientIPDetection` | _[ClientIPDetectionSettings](#clientipdetectionsettings)_ |  false  | ClientIPDetectionSettings provides configuration for determining the original client IP address for requests. |
| `http3` | _[HTTP3Settings](#http3settings)_ |  false  | HTTP3 provides HTTP/3 configuration on the listener. |
| `tls` | _[TLSSettings](#tlssettings)_ |  false  | TLS settings configure TLS termination settings with the downstream client. |
| `path` | _[PathSettings](#pathsettings)_ |  false  | Path enables managing how the incoming path set by clients can be normalized. |
| `http1` | _[HTTP1Settings](#http1settings)_ |  false  | HTTP1 provides HTTP/1 configuration on the listener. |
| `headers` | _[HeaderSettings](#headersettings)_ |  false  | HeaderSettings provides configuration for header management. |
| `timeout` | _[ClientTimeout](#clienttimeout)_ |  false  | Timeout settings for the client connections. |


#### ClientValidationContext



ClientValidationContext holds configuration that can be used to validate the client initiating the TLS connection to the Gateway. By default, no client specific configuration is validated.

_Appears in:_
- [TLSSettings](#tlssettings)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `caCertificateRefs` | _[SecretObjectReference](https://gateway-api.sigs.k8s.io/references/spec/#gateway.networking.k8s.io/v1.SecretObjectReference) array_ |  false  | CACertificateRefs contains one or more references to Kubernetes objects that contain TLS certificates of the Certificate Authorities that can be used as a trust anchor to validate the certificates presented by the client. <br /><br /> A single reference to a Kubernetes ConfigMap or a Kubernetes Secret, with the CA certificate in a key named `ca.crt` is currently supported. <br /><br /> References to a resource in different namespace are invalid UNLESS there is a ReferenceGrant in the target namespace that allows the certificate to be attached. |


#### Compression



Compression defines the config of enabling compression. This can help reduce the bandwidth at the expense of higher CPU.

_Appears in:_
- [BackendTrafficPolicySpec](#backendtrafficpolicyspec)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `type` | _[CompressorType](#compressortype)_ |  true  | CompressorType defines the compressor type to use for compression. |
| `gzip` | _[GzipCompressor](#gzipcompressor)_ |  false  | The configuration for GZIP compressor. |


#### CompressorType

_Underlying type:_ _string_

CompressorType defines the types of compressor library supported by Envoy Gateway.

_Appears in:_
- [Compression](#compression)



#### ConsistentHash



ConsistentHash defines the configuration related to the consistent hash load balancer policy

_Appears in:_
- [LoadBalancer](#loadbalancer)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `type` | _[ConsistentHashType](#consistenthashtype)_ |  true  |  |


#### ConsistentHashType

_Underlying type:_ _string_

ConsistentHashType defines the type of input to hash on.

_Appears in:_
- [ConsistentHash](#consistenthash)



#### CustomHeaderExtensionSettings



CustomHeader provides configuration for determining the client IP address for a request based on a trusted custom HTTP header. This uses the the custom_header original IP detection extension. Refer to https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/http/original_ip_detection/custom_header/v3/custom_header.proto for more details.

_Appears in:_
- [ClientIPDetectionSettings](#clientipdetectionsettings)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `name` | _string_ |  true  | Name of the header containing the original downstream remote address, if present. |
| `failClosed` | _boolean_ |  false  | FailClosed is a switch used to control the flow of traffic when client IP detection fails. If set to true, the listener will respond with 403 Forbidden when the client IP address cannot be determined. |


#### CustomTag





_Appears in:_
- [ProxyTracing](#proxytracing)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `type` | _[CustomTagType](#customtagtype)_ |  true  | Type defines the type of custom tag. |
| `literal` | _[LiteralCustomTag](#literalcustomtag)_ |  true  | Literal adds hard-coded value to each span. It's required when the type is "Literal". |
| `environment` | _[EnvironmentCustomTag](#environmentcustomtag)_ |  true  | Environment adds value from environment variable to each span. It's required when the type is "Environment". |
| `requestHeader` | _[RequestHeaderCustomTag](#requestheadercustomtag)_ |  true  | RequestHeader adds value from request header to each span. It's required when the type is "RequestHeader". |


#### CustomTagType

_Underlying type:_ _string_



_Appears in:_
- [CustomTag](#customtag)



#### EnvironmentCustomTag



EnvironmentCustomTag adds value from environment variable to each span.

_Appears in:_
- [CustomTag](#customtag)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `name` | _string_ |  true  | Name defines the name of the environment variable which to extract the value from. |
| `defaultValue` | _string_ |  false  | DefaultValue defines the default value to use if the environment variable is not set. |


#### EnvoyGateway



EnvoyGateway is the schema for the envoygateways API.



| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `apiVersion` | _string_ | |`gateway.envoyproxy.io/v1alpha1`
| `kind` | _string_ | |`EnvoyGateway`
| `gateway` | _[Gateway](#gateway)_ |  false  | Gateway defines desired Gateway API specific configuration. If unset, default configuration parameters will apply. |
| `provider` | _[EnvoyGatewayProvider](#envoygatewayprovider)_ |  false  | Provider defines the desired provider and provider-specific configuration. If unspecified, the Kubernetes provider is used with default configuration parameters. |
| `logging` | _[EnvoyGatewayLogging](#envoygatewaylogging)_ |  false  | Logging defines logging parameters for Envoy Gateway. |
| `admin` | _[EnvoyGatewayAdmin](#envoygatewayadmin)_ |  false  | Admin defines the desired admin related abilities. If unspecified, the Admin is used with default configuration parameters. |
| `telemetry` | _[EnvoyGatewayTelemetry](#envoygatewaytelemetry)_ |  false  | Telemetry defines the desired control plane telemetry related abilities. If unspecified, the telemetry is used with default configuration. |
| `rateLimit` | _[RateLimit](#ratelimit)_ |  false  | RateLimit defines the configuration associated with the Rate Limit service deployed by Envoy Gateway required to implement the Global Rate limiting functionality. The specific rate limit service used here is the reference implementation in Envoy. For more details visit https://github.com/envoyproxy/ratelimit. This configuration is unneeded for "Local" rate limiting. |
| `extensionManager` | _[ExtensionManager](#extensionmanager)_ |  false  | ExtensionManager defines an extension manager to register for the Envoy Gateway Control Plane. |
| `extensionApis` | _[ExtensionAPISettings](#extensionapisettings)_ |  false  | ExtensionAPIs defines the settings related to specific Gateway API Extensions implemented by Envoy Gateway |


#### EnvoyGatewayAdmin



EnvoyGatewayAdmin defines the Envoy Gateway Admin configuration.

_Appears in:_
- [EnvoyGateway](#envoygateway)
- [EnvoyGatewaySpec](#envoygatewayspec)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `address` | _[EnvoyGatewayAdminAddress](#envoygatewayadminaddress)_ |  false  | Address defines the address of Envoy Gateway Admin Server. |
| `enableDumpConfig` | _boolean_ |  false  | EnableDumpConfig defines if enable dump config in Envoy Gateway logs. |
| `enablePprof` | _boolean_ |  false  | EnablePprof defines if enable pprof in Envoy Gateway Admin Server. |


#### EnvoyGatewayAdminAddress



EnvoyGatewayAdminAddress defines the Envoy Gateway Admin Address configuration.

_Appears in:_
- [EnvoyGatewayAdmin](#envoygatewayadmin)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `port` | _integer_ |  false  | Port defines the port the admin server is exposed on. |
| `host` | _string_ |  false  | Host defines the admin server hostname. |


#### EnvoyGatewayCustomProvider



EnvoyGatewayCustomProvider defines configuration for the Custom provider.

_Appears in:_
- [EnvoyGatewayProvider](#envoygatewayprovider)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `resource` | _[EnvoyGatewayResourceProvider](#envoygatewayresourceprovider)_ |  true  | Resource defines the desired resource provider. This provider is used to specify the provider to be used to retrieve the resource configurations such as Gateway API resources |
| `infrastructure` | _[EnvoyGatewayInfrastructureProvider](#envoygatewayinfrastructureprovider)_ |  true  | Infrastructure defines the desired infrastructure provider. This provider is used to specify the provider to be used to provide an environment to deploy the out resources like the Envoy Proxy data plane. |


#### EnvoyGatewayFileResourceProvider



EnvoyGatewayFileResourceProvider defines configuration for the File Resource provider.

_Appears in:_
- [EnvoyGatewayResourceProvider](#envoygatewayresourceprovider)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `paths` | _string array_ |  true  | Paths are the paths to a directory or file containing the resource configuration. Recursive sub directories are not currently supported. |


#### EnvoyGatewayHostInfrastructureProvider



EnvoyGatewayHostInfrastructureProvider defines configuration for the Host Infrastructure provider.

_Appears in:_
- [EnvoyGatewayInfrastructureProvider](#envoygatewayinfrastructureprovider)



#### EnvoyGatewayInfrastructureProvider



EnvoyGatewayInfrastructureProvider defines configuration for the Custom Infrastructure provider.

_Appears in:_
- [EnvoyGatewayCustomProvider](#envoygatewaycustomprovider)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `type` | _[InfrastructureProviderType](#infrastructureprovidertype)_ |  true  | Type is the type of infrastructure providers to use. Supported types are "Host". |
| `host` | _[EnvoyGatewayHostInfrastructureProvider](#envoygatewayhostinfrastructureprovider)_ |  false  | Host defines the configuration of the Host provider. Host provides runtime deployment of the data plane as a child process on the host environment. |


#### EnvoyGatewayKubernetesProvider



EnvoyGatewayKubernetesProvider defines configuration for the Kubernetes provider.

_Appears in:_
- [EnvoyGatewayProvider](#envoygatewayprovider)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `rateLimitDeployment` | _[KubernetesDeploymentSpec](#kubernetesdeploymentspec)_ |  false  | RateLimitDeployment defines the desired state of the Envoy ratelimit deployment resource. If unspecified, default settings for the managed Envoy ratelimit deployment resource are applied. |
| `watch` | _[KubernetesWatchMode](#kuberneteswatchmode)_ |  false  | Watch holds configuration of which input resources should be watched and reconciled. |
| `deploy` | _[KubernetesDeployMode](#kubernetesdeploymode)_ |  false  | Deploy holds configuration of how output managed resources such as the Envoy Proxy data plane should be deployed |
| `overwriteControlPlaneCerts` | _boolean_ |  false  | OverwriteControlPlaneCerts updates the secrets containing the control plane certs, when set. |


#### EnvoyGatewayLogComponent

_Underlying type:_ _string_

EnvoyGatewayLogComponent defines a component that supports a configured logging level.

_Appears in:_
- [EnvoyGatewayLogging](#envoygatewaylogging)



#### EnvoyGatewayLogging



EnvoyGatewayLogging defines logging for Envoy Gateway.

_Appears in:_
- [EnvoyGateway](#envoygateway)
- [EnvoyGatewaySpec](#envoygatewayspec)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `level` | _object (keys:[EnvoyGatewayLogComponent](#envoygatewaylogcomponent), values:[LogLevel](#loglevel))_ |  true  | Level is the logging level. If unspecified, defaults to "info". EnvoyGatewayLogComponent options: default/provider/gateway-api/xds-translator/xds-server/infrastructure/global-ratelimit. LogLevel options: debug/info/error/warn. |


#### EnvoyGatewayMetricSink



EnvoyGatewayMetricSink defines control plane metric sinks where metrics are sent to.

_Appears in:_
- [EnvoyGatewayMetrics](#envoygatewaymetrics)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `type` | _[MetricSinkType](#metricsinktype)_ |  true  | Type defines the metric sink type. EG control plane currently supports OpenTelemetry. |
| `openTelemetry` | _[EnvoyGatewayOpenTelemetrySink](#envoygatewayopentelemetrysink)_ |  true  | OpenTelemetry defines the configuration for OpenTelemetry sink. It's required if the sink type is OpenTelemetry. |


#### EnvoyGatewayMetrics



EnvoyGatewayMetrics defines control plane push/pull metrics configurations.

_Appears in:_
- [EnvoyGatewayTelemetry](#envoygatewaytelemetry)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `sinks` | _[EnvoyGatewayMetricSink](#envoygatewaymetricsink) array_ |  true  | Sinks defines the metric sinks where metrics are sent to. |
| `prometheus` | _[EnvoyGatewayPrometheusProvider](#envoygatewayprometheusprovider)_ |  true  | Prometheus defines the configuration for prometheus endpoint. |


#### EnvoyGatewayOpenTelemetrySink





_Appears in:_
- [EnvoyGatewayMetricSink](#envoygatewaymetricsink)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `host` | _string_ |  true  | Host define the sink service hostname. |
| `protocol` | _string_ |  true  | Protocol define the sink service protocol. |
| `port` | _integer_ |  false  | Port defines the port the sink service is exposed on. |


#### EnvoyGatewayPrometheusProvider



EnvoyGatewayPrometheusProvider will expose prometheus endpoint in pull mode.

_Appears in:_
- [EnvoyGatewayMetrics](#envoygatewaymetrics)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `disable` | _boolean_ |  true  | Disable defines if disables the prometheus metrics in pull mode. |


#### EnvoyGatewayProvider



EnvoyGatewayProvider defines the desired configuration of a provider.

_Appears in:_
- [EnvoyGateway](#envoygateway)
- [EnvoyGatewaySpec](#envoygatewayspec)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `type` | _[ProviderType](#providertype)_ |  true  | Type is the type of provider to use. Supported types are "Kubernetes". |
| `kubernetes` | _[EnvoyGatewayKubernetesProvider](#envoygatewaykubernetesprovider)_ |  false  | Kubernetes defines the configuration of the Kubernetes provider. Kubernetes provides runtime configuration via the Kubernetes API. |
| `custom` | _[EnvoyGatewayCustomProvider](#envoygatewaycustomprovider)_ |  false  | Custom defines the configuration for the Custom provider. This provider allows you to define a specific resource provider and a infrastructure provider. |


#### EnvoyGatewayResourceProvider



EnvoyGatewayResourceProvider defines configuration for the Custom Resource provider.

_Appears in:_
- [EnvoyGatewayCustomProvider](#envoygatewaycustomprovider)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `type` | _[ResourceProviderType](#resourceprovidertype)_ |  true  | Type is the type of resource provider to use. Supported types are "File". |
| `file` | _[EnvoyGatewayFileResourceProvider](#envoygatewayfileresourceprovider)_ |  false  | File defines the configuration of the File provider. File provides runtime configuration defined by one or more files. |


#### EnvoyGatewaySpec



EnvoyGatewaySpec defines the desired state of Envoy Gateway.

_Appears in:_
- [EnvoyGateway](#envoygateway)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `gateway` | _[Gateway](#gateway)_ |  false  | Gateway defines desired Gateway API specific configuration. If unset, default configuration parameters will apply. |
| `provider` | _[EnvoyGatewayProvider](#envoygatewayprovider)_ |  false  | Provider defines the desired provider and provider-specific configuration. If unspecified, the Kubernetes provider is used with default configuration parameters. |
| `logging` | _[EnvoyGatewayLogging](#envoygatewaylogging)_ |  false  | Logging defines logging parameters for Envoy Gateway. |
| `admin` | _[EnvoyGatewayAdmin](#envoygatewayadmin)_ |  false  | Admin defines the desired admin related abilities. If unspecified, the Admin is used with default configuration parameters. |
| `telemetry` | _[EnvoyGatewayTelemetry](#envoygatewaytelemetry)_ |  false  | Telemetry defines the desired control plane telemetry related abilities. If unspecified, the telemetry is used with default configuration. |
| `rateLimit` | _[RateLimit](#ratelimit)_ |  false  | RateLimit defines the configuration associated with the Rate Limit service deployed by Envoy Gateway required to implement the Global Rate limiting functionality. The specific rate limit service used here is the reference implementation in Envoy. For more details visit https://github.com/envoyproxy/ratelimit. This configuration is unneeded for "Local" rate limiting. |
| `extensionManager` | _[ExtensionManager](#extensionmanager)_ |  false  | ExtensionManager defines an extension manager to register for the Envoy Gateway Control Plane. |
| `extensionApis` | _[ExtensionAPISettings](#extensionapisettings)_ |  false  | ExtensionAPIs defines the settings related to specific Gateway API Extensions implemented by Envoy Gateway |


#### EnvoyGatewayTelemetry



EnvoyGatewayTelemetry defines telemetry configurations for envoy gateway control plane. Control plane will focus on metrics observability telemetry and tracing telemetry later.

_Appears in:_
- [EnvoyGateway](#envoygateway)
- [EnvoyGatewaySpec](#envoygatewayspec)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `metrics` | _[EnvoyGatewayMetrics](#envoygatewaymetrics)_ |  true  | Metrics defines metrics configuration for envoy gateway. |


#### EnvoyJSONPatchConfig



EnvoyJSONPatchConfig defines the configuration for patching a Envoy xDS Resource using JSONPatch semantic

_Appears in:_
- [EnvoyPatchPolicySpec](#envoypatchpolicyspec)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `type` | _[EnvoyResourceType](#envoyresourcetype)_ |  true  | Type is the typed URL of the Envoy xDS Resource |
| `name` | _string_ |  true  | Name is the name of the resource |
| `operation` | _[JSONPatchOperation](#jsonpatchoperation)_ |  true  | Patch defines the JSON Patch Operation |


#### EnvoyPatchPolicy



EnvoyPatchPolicy allows the user to modify the generated Envoy xDS resources by Envoy Gateway using this patch API

_Appears in:_
- [EnvoyPatchPolicyList](#envoypatchpolicylist)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `apiVersion` | _string_ | |`gateway.envoyproxy.io/v1alpha1`
| `kind` | _string_ | |`EnvoyPatchPolicy`
| `metadata` | _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#objectmeta-v1-meta)_ |  true  | Refer to Kubernetes API documentation for fields of `metadata`. |
| `spec` | _[EnvoyPatchPolicySpec](#envoypatchpolicyspec)_ |  true  | Spec defines the desired state of EnvoyPatchPolicy. |


#### EnvoyPatchPolicyList



EnvoyPatchPolicyList contains a list of EnvoyPatchPolicy resources.



| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `apiVersion` | _string_ | |`gateway.envoyproxy.io/v1alpha1`
| `kind` | _string_ | |`EnvoyPatchPolicyList`
| `metadata` | _[ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#listmeta-v1-meta)_ |  true  | Refer to Kubernetes API documentation for fields of `metadata`. |
| `items` | _[EnvoyPatchPolicy](#envoypatchpolicy) array_ |  true  |  |


#### EnvoyPatchPolicySpec



EnvoyPatchPolicySpec defines the desired state of EnvoyPatchPolicy.

_Appears in:_
- [EnvoyPatchPolicy](#envoypatchpolicy)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `type` | _[EnvoyPatchType](#envoypatchtype)_ |  true  | Type decides the type of patch. Valid EnvoyPatchType values are "JSONPatch". |
| `jsonPatches` | _[EnvoyJSONPatchConfig](#envoyjsonpatchconfig) array_ |  false  | JSONPatch defines the JSONPatch configuration. |
| `targetRef` | _[PolicyTargetReference](#policytargetreference)_ |  true  | TargetRef is the name of the Gateway API resource this policy is being attached to. By default attaching to Gateway is supported and when mergeGateways is enabled it should attach to GatewayClass. This Policy and the TargetRef MUST be in the same namespace for this Policy to have effect and be applied to the Gateway TargetRef |
| `priority` | _integer_ |  true  | Priority of the EnvoyPatchPolicy. If multiple EnvoyPatchPolicies are applied to the same TargetRef, they will be applied in the ascending order of the priority i.e. int32.min has the highest priority and int32.max has the lowest priority. Defaults to 0. |




#### EnvoyPatchType

_Underlying type:_ _string_

EnvoyPatchType specifies the types of Envoy patching mechanisms.

_Appears in:_
- [EnvoyPatchPolicySpec](#envoypatchpolicyspec)



#### EnvoyProxy



EnvoyProxy is the schema for the envoyproxies API.



| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `apiVersion` | _string_ | |`gateway.envoyproxy.io/v1alpha1`
| `kind` | _string_ | |`EnvoyProxy`
| `metadata` | _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#objectmeta-v1-meta)_ |  true  | Refer to Kubernetes API documentation for fields of `metadata`. |
| `spec` | _[EnvoyProxySpec](#envoyproxyspec)_ |  true  | EnvoyProxySpec defines the desired state of EnvoyProxy. |


#### EnvoyProxyKubernetesProvider



EnvoyProxyKubernetesProvider defines configuration for the Kubernetes resource provider.

_Appears in:_
- [EnvoyProxyProvider](#envoyproxyprovider)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `envoyDeployment` | _[KubernetesDeploymentSpec](#kubernetesdeploymentspec)_ |  false  | EnvoyDeployment defines the desired state of the Envoy deployment resource. If unspecified, default settings for the managed Envoy deployment resource are applied. |
| `envoyService` | _[KubernetesServiceSpec](#kubernetesservicespec)_ |  false  | EnvoyService defines the desired state of the Envoy service resource. If unspecified, default settings for the managed Envoy service resource are applied. |
| `envoyHpa` | _[KubernetesHorizontalPodAutoscalerSpec](#kuberneteshorizontalpodautoscalerspec)_ |  false  | EnvoyHpa defines the Horizontal Pod Autoscaler settings for Envoy Proxy Deployment. Once the HPA is being set, Replicas field from EnvoyDeployment will be ignored. |


#### EnvoyProxyProvider



EnvoyProxyProvider defines the desired state of a resource provider.

_Appears in:_
- [EnvoyProxySpec](#envoyproxyspec)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `type` | _[ProviderType](#providertype)_ |  true  | Type is the type of resource provider to use. A resource provider provides infrastructure resources for running the data plane, e.g. Envoy proxy, and optional auxiliary control planes. Supported types are "Kubernetes". |
| `kubernetes` | _[EnvoyProxyKubernetesProvider](#envoyproxykubernetesprovider)_ |  false  | Kubernetes defines the desired state of the Kubernetes resource provider. Kubernetes provides infrastructure resources for running the data plane, e.g. Envoy proxy. If unspecified and type is "Kubernetes", default settings for managed Kubernetes resources are applied. |


#### EnvoyProxySpec



EnvoyProxySpec defines the desired state of EnvoyProxy.

_Appears in:_
- [EnvoyProxy](#envoyproxy)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `provider` | _[EnvoyProxyProvider](#envoyproxyprovider)_ |  false  | Provider defines the desired resource provider and provider-specific configuration. If unspecified, the "Kubernetes" resource provider is used with default configuration parameters. |
| `logging` | _[ProxyLogging](#proxylogging)_ |  true  | Logging defines logging parameters for managed proxies. |
| `telemetry` | _[ProxyTelemetry](#proxytelemetry)_ |  false  | Telemetry defines telemetry parameters for managed proxies. |
| `bootstrap` | _[ProxyBootstrap](#proxybootstrap)_ |  false  | Bootstrap defines the Envoy Bootstrap as a YAML string. Visit https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/bootstrap/v3/bootstrap.proto#envoy-v3-api-msg-config-bootstrap-v3-bootstrap to learn more about the syntax. If set, this is the Bootstrap configuration used for the managed Envoy Proxy fleet instead of the default Bootstrap configuration set by Envoy Gateway. Some fields within the Bootstrap that are required to communicate with the xDS Server (Envoy Gateway) and receive xDS resources from it are not configurable and will result in the `EnvoyProxy` resource being rejected. Backward compatibility across minor versions is not guaranteed. We strongly recommend using `egctl x translate` to generate a `EnvoyProxy` resource with the `Bootstrap` field set to the default Bootstrap configuration used. You can edit this configuration, and rerun `egctl x translate` to ensure there are no validation errors. |
| `concurrency` | _integer_ |  false  | Concurrency defines the number of worker threads to run. If unset, it defaults to the number of cpuset threads on the platform. |
| `extraArgs` | _string array_ |  false  | ExtraArgs defines additional command line options that are provided to Envoy. More info: https://www.envoyproxy.io/docs/envoy/latest/operations/cli#command-line-options Note: some command line options are used internally(e.g. --log-level) so they cannot be provided here. |
| `mergeGateways` | _boolean_ |  false  | MergeGateways defines if Gateway resources should be merged onto the same Envoy Proxy Infrastructure. Setting this field to true would merge all Gateway Listeners under the parent Gateway Class. This means that the port, protocol and hostname tuple must be unique for every listener. If a duplicate listener is detected, the newer listener (based on timestamp) will be rejected and its status will be updated with a "Accepted=False" condition. |
| `shutdown` | _[ShutdownConfig](#shutdownconfig)_ |  false  | Shutdown defines configuration for graceful envoy shutdown process. |




#### EnvoyResourceType

_Underlying type:_ _string_

EnvoyResourceType specifies the type URL of the Envoy resource.

_Appears in:_
- [EnvoyJSONPatchConfig](#envoyjsonpatchconfig)



#### ExtAuth



ExtAuth defines the configuration for External Authorization.

_Appears in:_
- [SecurityPolicySpec](#securitypolicyspec)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `grpc` | _[GRPCExtAuthService](#grpcextauthservice)_ |  true  | GRPC defines the gRPC External Authorization service. Either GRPCService or HTTPService must be specified, and only one of them can be provided. |
| `http` | _[HTTPExtAuthService](#httpextauthservice)_ |  true  | HTTP defines the HTTP External Authorization service. Either GRPCService or HTTPService must be specified, and only one of them can be provided. |
| `headersToExtAuth` | _string array_ |  false  | HeadersToExtAuth defines the client request headers that will be included in the request to the external authorization service. Note: If not specified, the default behavior for gRPC and HTTP external authorization services is different due to backward compatibility reasons. All headers will be included in the check request to a gRPC authorization server. Only the following headers will be included in the check request to an HTTP authorization server: Host, Method, Path, Content-Length, and Authorization. And these headers will always be included to the check request to an HTTP authorization server by default, no matter whether they are specified in HeadersToExtAuth or not. |


#### ExtensionAPISettings



ExtensionAPISettings defines the settings specific to Gateway API Extensions.

_Appears in:_
- [EnvoyGateway](#envoygateway)
- [EnvoyGatewaySpec](#envoygatewayspec)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `enableEnvoyPatchPolicy` | _boolean_ |  true  | EnableEnvoyPatchPolicy enables Envoy Gateway to reconcile and implement the EnvoyPatchPolicy resources. |


#### ExtensionHooks



ExtensionHooks defines extension hooks across all supported runners

_Appears in:_
- [ExtensionManager](#extensionmanager)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `xdsTranslator` | _[XDSTranslatorHooks](#xdstranslatorhooks)_ |  true  | XDSTranslator defines all the supported extension hooks for the xds-translator runner |


#### ExtensionManager



ExtensionManager defines the configuration for registering an extension manager to the Envoy Gateway control plane.

_Appears in:_
- [EnvoyGateway](#envoygateway)
- [EnvoyGatewaySpec](#envoygatewayspec)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `resources` | _[GroupVersionKind](#groupversionkind) array_ |  false  | Resources defines the set of K8s resources the extension will handle. |
| `hooks` | _[ExtensionHooks](#extensionhooks)_ |  true  | Hooks defines the set of hooks the extension supports |
| `service` | _[ExtensionService](#extensionservice)_ |  true  | Service defines the configuration of the extension service that the Envoy Gateway Control Plane will call through extension hooks. |


#### ExtensionService



ExtensionService defines the configuration for connecting to a registered extension service.

_Appears in:_
- [ExtensionManager](#extensionmanager)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `host` | _string_ |  true  | Host define the extension service hostname. |
| `port` | _integer_ |  false  | Port defines the port the extension service is exposed on. |
| `tls` | _[ExtensionTLS](#extensiontls)_ |  false  | TLS defines TLS configuration for communication between Envoy Gateway and the extension service. |


#### ExtensionTLS



ExtensionTLS defines the TLS configuration when connecting to an extension service

_Appears in:_
- [ExtensionService](#extensionservice)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `certificateRef` | _[SecretObjectReference](https://gateway-api.sigs.k8s.io/references/spec/#gateway.networking.k8s.io/v1.SecretObjectReference)_ |  true  | CertificateRef contains a references to objects (Kubernetes objects or otherwise) that contains a TLS certificate and private keys. These certificates are used to establish a TLS handshake to the extension server. <br /><br /> CertificateRef can only reference a Kubernetes Secret at this time. |


#### FaultInjection



FaultInjection defines the fault injection policy to be applied. This configuration can be used to inject delays and abort requests to mimic failure scenarios such as service failures and overloads

_Appears in:_
- [BackendTrafficPolicySpec](#backendtrafficpolicyspec)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `delay` | _[FaultInjectionDelay](#faultinjectiondelay)_ |  false  | If specified, a delay will be injected into the request. |
| `abort` | _[FaultInjectionAbort](#faultinjectionabort)_ |  false  | If specified, the request will be aborted if it meets the configuration criteria. |


#### FaultInjectionAbort



FaultInjectionAbort defines the abort fault injection configuration

_Appears in:_
- [FaultInjection](#faultinjection)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `httpStatus` | _integer_ |  false  | StatusCode specifies the HTTP status code to be returned |
| `grpcStatus` | _integer_ |  false  | GrpcStatus specifies the GRPC status code to be returned |
| `percentage` | _float_ |  false  | Percentage specifies the percentage of requests to be aborted. Default 100%, if set 0, no requests will be aborted. Accuracy to 0.0001%. |


#### FaultInjectionDelay



FaultInjectionDelay defines the delay fault injection configuration

_Appears in:_
- [FaultInjection](#faultinjection)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `fixedDelay` | _[Duration](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#duration-v1-meta)_ |  true  | FixedDelay specifies the fixed delay duration |
| `percentage` | _float_ |  false  | Percentage specifies the percentage of requests to be delayed. Default 100%, if set 0, no requests will be delayed. Accuracy to 0.0001%. |


#### FileEnvoyProxyAccessLog





_Appears in:_
- [ProxyAccessLogSink](#proxyaccesslogsink)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `path` | _string_ |  true  | Path defines the file path used to expose envoy access log(e.g. /dev/stdout). |


#### GRPCExtAuthService



GRPCExtAuthService defines the gRPC External Authorization service The authorization request message is defined in https://www.envoyproxy.io/docs/envoy/latest/api-v3/service/auth/v3/external_auth.proto

_Appears in:_
- [ExtAuth](#extauth)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `backendRef` | _[BackendObjectReference](#backendobjectreference)_ |  true  | BackendRef references a Kubernetes object that represents the backend server to which the authorization request will be sent. Only service Kind is supported for now. |


#### Gateway



Gateway defines the desired Gateway API configuration of Envoy Gateway.

_Appears in:_
- [EnvoyGateway](#envoygateway)
- [EnvoyGatewaySpec](#envoygatewayspec)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `controllerName` | _string_ |  false  | ControllerName defines the name of the Gateway API controller. If unspecified, defaults to "gateway.envoyproxy.io/gatewayclass-controller". See the following for additional details: https://gateway-api.sigs.k8s.io/references/spec/#gateway.networking.k8s.io/v1.GatewayClass |


#### GlobalRateLimit



GlobalRateLimit defines global rate limit configuration.

_Appears in:_
- [RateLimitSpec](#ratelimitspec)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `rules` | _[RateLimitRule](#ratelimitrule) array_ |  true  | Rules are a list of RateLimit selectors and limits. Each rule and its associated limit is applied in a mutually exclusive way. If a request matches multiple rules, each of their associated limits get applied, so a single request might increase the rate limit counters for multiple rules if selected. The rate limit service will return a logical OR of the individual rate limit decisions of all matching rules. For example, if a request matches two rules, one rate limited and one not, the final decision will be to rate limit the request. |


#### GroupVersionKind



GroupVersionKind unambiguously identifies a Kind. It can be converted to k8s.io/apimachinery/pkg/runtime/schema.GroupVersionKind

_Appears in:_
- [ExtensionManager](#extensionmanager)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `group` | _string_ |  true  |  |
| `version` | _string_ |  true  |  |
| `kind` | _string_ |  true  |  |


#### GzipCompressor



GzipCompressor defines the config for the Gzip compressor. The default values can be found here: https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/compression/gzip/compressor/v3/gzip.proto#extension-envoy-compression-gzip-compressor

_Appears in:_
- [Compression](#compression)



#### HTTP10Settings



HTTP10Settings provides HTTP/1.0 configuration on the listener.

_Appears in:_
- [HTTP1Settings](#http1settings)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `useDefaultHost` | _boolean_ |  false  | UseDefaultHost defines if the HTTP/1.0 request is missing the Host header, then the hostname associated with the listener should be injected into the request. If this is not set and an HTTP/1.0 request arrives without a host, then it will be rejected. |


#### HTTP1Settings



HTTP1Settings provides HTTP/1 configuration on the listener.

_Appears in:_
- [ClientTrafficPolicySpec](#clienttrafficpolicyspec)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `enableTrailers` | _boolean_ |  false  | EnableTrailers defines if HTTP/1 trailers should be proxied by Envoy. |
| `preserveHeaderCase` | _boolean_ |  false  | PreserveHeaderCase defines if Envoy should preserve the letter case of headers. By default, Envoy will lowercase all the headers. |
| `http10` | _[HTTP10Settings](#http10settings)_ |  false  | HTTP10 turns on support for HTTP/1.0 and HTTP/0.9 requests. |


#### HTTP3Settings



HTTP3Settings provides HTTP/3 configuration on the listener.

_Appears in:_
- [ClientTrafficPolicySpec](#clienttrafficpolicyspec)



#### HTTPActiveHealthChecker



HTTPActiveHealthChecker defines the settings of http health check.

_Appears in:_
- [ActiveHealthCheck](#activehealthcheck)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `path` | _string_ |  true  | Path defines the HTTP path that will be requested during health checking. |
| `method` | _string_ |  false  | Method defines the HTTP method used for health checking. Defaults to GET |
| `expectedStatuses` | _[HTTPStatus](#httpstatus) array_ |  false  | ExpectedStatuses defines a list of HTTP response statuses considered healthy. Defaults to 200 only |
| `expectedResponse` | _[ActiveHealthCheckPayload](#activehealthcheckpayload)_ |  false  | ExpectedResponse defines a list of HTTP expected responses to match. |


#### HTTPClientTimeout





_Appears in:_
- [ClientTimeout](#clienttimeout)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `requestReceivedTimeout` | _[Duration](#duration)_ |  false  | The duration envoy waits for the complete request reception. This timer starts upon request initiation and stops when either the last byte of the request is sent upstream or when the response begins. |


#### HTTPExtAuthService



HTTPExtAuthService defines the HTTP External Authorization service

_Appears in:_
- [ExtAuth](#extauth)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `backendRef` | _[BackendObjectReference](#backendobjectreference)_ |  true  | BackendRef references a Kubernetes object that represents the backend server to which the authorization request will be sent. Only service Kind is supported for now. |
| `path` | _string_ |  true  | Path is the path of the HTTP External Authorization service. If path is specified, the authorization request will be sent to that path, or else the authorization request will be sent to the root path. |
| `headersToBackend` | _string array_ |  false  | HeadersToBackend are the authorization response headers that will be added to the original client request before sending it to the backend server. Note that coexisting headers will be overridden. If not specified, no authorization response headers will be added to the original client request. |


#### HTTPStatus

_Underlying type:_ _integer_

HTTPStatus defines the http status code.

_Appears in:_
- [HTTPActiveHealthChecker](#httpactivehealthchecker)
- [RetryOn](#retryon)



#### HTTPTimeout





_Appears in:_
- [Timeout](#timeout)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `connectionIdleTimeout` | _[Duration](#duration)_ |  false  | The idle timeout for an HTTP connection. Idle time is defined as a period in which there are no active requests in the connection. Default: 1 hour. |
| `maxConnectionDuration` | _[Duration](#duration)_ |  false  | The maximum duration of an HTTP connection. Default: unlimited. |




#### HeaderMatchType

_Underlying type:_ _string_

HeaderMatchType specifies the semantics of how HTTP header values should be compared. Valid HeaderMatchType values are "Exact", "RegularExpression", and "Distinct".

_Appears in:_
- [HeaderMatch](#headermatch)



#### HeaderSettings



HeaderSettings providess configuration options for headers on the listener.

_Appears in:_
- [ClientTrafficPolicySpec](#clienttrafficpolicyspec)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `enableEnvoyHeaders` | _boolean_ |  false  | EnableEnvoyHeaders configures Envoy Proxy to add the "X-Envoy-" headers to requests and responses. |


#### HealthCheck



HealthCheck configuration to decide which endpoints are healthy and can be used for routing.

_Appears in:_
- [BackendTrafficPolicySpec](#backendtrafficpolicyspec)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `active` | _[ActiveHealthCheck](#activehealthcheck)_ |  false  | Active health check configuration |
| `passive` | _[PassiveHealthCheck](#passivehealthcheck)_ |  false  | Passive passive check configuration |


#### InfrastructureProviderType

_Underlying type:_ _string_

InfrastructureProviderType defines the types of custom infrastructure providers supported by Envoy Gateway.

_Appears in:_
- [EnvoyGatewayInfrastructureProvider](#envoygatewayinfrastructureprovider)



#### JSONPatchOperation



JSONPatchOperation defines the JSON Patch Operation as defined in https://datatracker.ietf.org/doc/html/rfc6902

_Appears in:_
- [EnvoyJSONPatchConfig](#envoyjsonpatchconfig)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `op` | _[JSONPatchOperationType](#jsonpatchoperationtype)_ |  true  | Op is the type of operation to perform |
| `path` | _string_ |  true  | Path is the location of the target document/field where the operation will be performed Refer to https://datatracker.ietf.org/doc/html/rfc6901 for more details. |
| `from` | _string_ |  false  | From is the source location of the value to be copied or moved. Only valid for move or copy operations Refer to https://datatracker.ietf.org/doc/html/rfc6901 for more details. |
| `value` | _[JSON](#json)_ |  false  | Value is the new value of the path location. The value is only used by the `add` and `replace` operations. |


#### JSONPatchOperationType

_Underlying type:_ _string_

JSONPatchOperationType specifies the JSON Patch operations that can be performed.

_Appears in:_
- [JSONPatchOperation](#jsonpatchoperation)



#### JWT



JWT defines the configuration for JSON Web Token (JWT) authentication.

_Appears in:_
- [SecurityPolicySpec](#securitypolicyspec)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `providers` | _[JWTProvider](#jwtprovider) array_ |  true  | Providers defines the JSON Web Token (JWT) authentication provider type. When multiple JWT providers are specified, the JWT is considered valid if any of the providers successfully validate the JWT. For additional details, see https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/jwt_authn_filter.html. |


#### JWTExtractor



JWTExtractor defines a custom JWT token extraction from HTTP request. If specified, Envoy will extract the JWT token from the listed extractors (headers, cookies, or params) and validate each of them. If any value extracted is found to be an invalid JWT, a 401 error will be returned.

_Appears in:_
- [JWTProvider](#jwtprovider)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `headers` | _[JWTHeaderExtractor](#jwtheaderextractor) array_ |  false  | Headers represents a list of HTTP request headers to extract the JWT token from. |
| `cookies` | _string array_ |  false  | Cookies represents a list of cookie names to extract the JWT token from. |
| `params` | _string array_ |  false  | Params represents a list of query parameters to extract the JWT token from. |


#### JWTHeaderExtractor



JWTHeaderExtractor defines an HTTP header location to extract JWT token

_Appears in:_
- [JWTExtractor](#jwtextractor)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `name` | _string_ |  true  | Name is the HTTP header name to retrieve the token |
| `valuePrefix` | _string_ |  false  | ValuePrefix is the prefix that should be stripped before extracting the token. The format would be used by Envoy like "{ValuePrefix}<TOKEN>". For example, "Authorization: Bearer <TOKEN>", then the ValuePrefix="Bearer " with a space at the end. |


#### JWTProvider



JWTProvider defines how a JSON Web Token (JWT) can be verified.

_Appears in:_
- [JWT](#jwt)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `name` | _string_ |  true  | Name defines a unique name for the JWT provider. A name can have a variety of forms, including RFC1123 subdomains, RFC 1123 labels, or RFC 1035 labels. |
| `issuer` | _string_ |  false  | Issuer is the principal that issued the JWT and takes the form of a URL or email address. For additional details, see https://tools.ietf.org/html/rfc7519#section-4.1.1 for URL format and https://rfc-editor.org/rfc/rfc5322.html for email format. If not provided, the JWT issuer is not checked. |
| `audiences` | _string array_ |  false  | Audiences is a list of JWT audiences allowed access. For additional details, see https://tools.ietf.org/html/rfc7519#section-4.1.3. If not provided, JWT audiences are not checked. |
| `remoteJWKS` | _[RemoteJWKS](#remotejwks)_ |  true  | RemoteJWKS defines how to fetch and cache JSON Web Key Sets (JWKS) from a remote HTTP/HTTPS endpoint. |
| `claimToHeaders` | _[ClaimToHeader](#claimtoheader) array_ |  false  | ClaimToHeaders is a list of JWT claims that must be extracted into HTTP request headers For examples, following config: The claim must be of type; string, int, double, bool. Array type claims are not supported |
| `recomputeRoute` | _boolean_ |  false  | RecomputeRoute clears the route cache and recalculates the routing decision. This field must be enabled if the headers generated from the claim are used for route matching decisions. If the recomputation selects a new route, features targeting the new matched route will be applied. |
| `extractFrom` | _[JWTExtractor](#jwtextractor)_ |  false  | ExtractFrom defines different ways to extract the JWT token from HTTP request. If empty, it defaults to extract JWT token from the Authorization HTTP request header using Bearer schema or access_token from query parameters. |


#### KubernetesContainerSpec



KubernetesContainerSpec defines the desired state of the Kubernetes container resource.

_Appears in:_
- [KubernetesDeploymentSpec](#kubernetesdeploymentspec)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `env` | _[EnvVar](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#envvar-v1-core) array_ |  false  | List of environment variables to set in the container. |
| `resources` | _[ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#resourcerequirements-v1-core)_ |  false  | Resources required by this container. More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/ |
| `securityContext` | _[SecurityContext](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#securitycontext-v1-core)_ |  false  | SecurityContext defines the security options the container should be run with. If set, the fields of SecurityContext override the equivalent fields of PodSecurityContext. More info: https://kubernetes.io/docs/tasks/configure-pod-container/security-context/ |
| `image` | _string_ |  false  | Image specifies the EnvoyProxy container image to be used, instead of the default image. |
| `volumeMounts` | _[VolumeMount](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#volumemount-v1-core) array_ |  false  | VolumeMounts are volumes to mount into the container's filesystem. Cannot be updated. |


#### KubernetesDeployMode



KubernetesDeployMode holds configuration for how to deploy managed resources such as the Envoy Proxy data plane fleet.

_Appears in:_
- [EnvoyGatewayKubernetesProvider](#envoygatewaykubernetesprovider)



#### KubernetesDeploymentSpec



KubernetesDeploymentSpec defines the desired state of the Kubernetes deployment resource.

_Appears in:_
- [EnvoyGatewayKubernetesProvider](#envoygatewaykubernetesprovider)
- [EnvoyProxyKubernetesProvider](#envoyproxykubernetesprovider)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `patch` | _[KubernetesPatchSpec](#kubernetespatchspec)_ |  false  | Patch defines how to perform the patch operation to deployment |
| `replicas` | _integer_ |  false  | Replicas is the number of desired pods. Defaults to 1. |
| `strategy` | _[DeploymentStrategy](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#deploymentstrategy-v1-apps)_ |  false  | The deployment strategy to use to replace existing pods with new ones. |
| `pod` | _[KubernetesPodSpec](#kubernetespodspec)_ |  false  | Pod defines the desired specification of pod. |
| `container` | _[KubernetesContainerSpec](#kubernetescontainerspec)_ |  false  | Container defines the desired specification of main container. |
| `initContainers` | _[Container](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#container-v1-core) array_ |  false  | List of initialization containers belonging to the pod. More info: https://kubernetes.io/docs/concepts/workloads/pods/init-containers/ |


#### KubernetesHorizontalPodAutoscalerSpec



KubernetesHorizontalPodAutoscalerSpec defines Kubernetes Horizontal Pod Autoscaler settings of Envoy Proxy Deployment. When HPA is enabled, it is recommended that the value in `KubernetesDeploymentSpec.replicas` be removed, otherwise Envoy Gateway will revert back to this value every time reconciliation occurs. See k8s.io.autoscaling.v2.HorizontalPodAutoScalerSpec.

_Appears in:_
- [EnvoyProxyKubernetesProvider](#envoyproxykubernetesprovider)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `minReplicas` | _integer_ |  false  | minReplicas is the lower limit for the number of replicas to which the autoscaler can scale down. It defaults to 1 replica. |
| `maxReplicas` | _integer_ |  true  | maxReplicas is the upper limit for the number of replicas to which the autoscaler can scale up. It cannot be less that minReplicas. |
| `metrics` | _[MetricSpec](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#metricspec-v2-autoscaling) array_ |  false  | metrics contains the specifications for which to use to calculate the desired replica count (the maximum replica count across all metrics will be used). If left empty, it defaults to being based on CPU utilization with average on 80% usage. |
| `behavior` | _[HorizontalPodAutoscalerBehavior](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#horizontalpodautoscalerbehavior-v2-autoscaling)_ |  false  | behavior configures the scaling behavior of the target in both Up and Down directions (scaleUp and scaleDown fields respectively). If not set, the default HPAScalingRules for scale up and scale down are used. See k8s.io.autoscaling.v2.HorizontalPodAutoScalerBehavior. |


#### KubernetesPatchSpec



KubernetesPatchSpec defines how to perform the patch operation

_Appears in:_
- [KubernetesDeploymentSpec](#kubernetesdeploymentspec)
- [KubernetesServiceSpec](#kubernetesservicespec)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `type` | _[MergeType](#mergetype)_ |  false  | Type is the type of merge operation to perform <br /><br /> By default, StrategicMerge is used as the patch type. |
| `value` | _[JSON](#json)_ |  true  | Object contains the raw configuration for merged object |


#### KubernetesPodSpec



KubernetesPodSpec defines the desired state of the Kubernetes pod resource.

_Appears in:_
- [KubernetesDeploymentSpec](#kubernetesdeploymentspec)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `annotations` | _object (keys:string, values:string)_ |  false  | Annotations are the annotations that should be appended to the pods. By default, no pod annotations are appended. |
| `labels` | _object (keys:string, values:string)_ |  false  | Labels are the additional labels that should be tagged to the pods. By default, no additional pod labels are tagged. |
| `securityContext` | _[PodSecurityContext](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#podsecuritycontext-v1-core)_ |  false  | SecurityContext holds pod-level security attributes and common container settings. Optional: Defaults to empty.  See type description for default values of each field. |
| `affinity` | _[Affinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#affinity-v1-core)_ |  false  | If specified, the pod's scheduling constraints. |
| `tolerations` | _[Toleration](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#toleration-v1-core) array_ |  false  | If specified, the pod's tolerations. |
| `volumes` | _[Volume](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#volume-v1-core) array_ |  false  | Volumes that can be mounted by containers belonging to the pod. More info: https://kubernetes.io/docs/concepts/storage/volumes |
| `imagePullSecrets` | _[LocalObjectReference](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#localobjectreference-v1-core) array_ |  false  | ImagePullSecrets is an optional list of references to secrets in the same namespace to use for pulling any of the images used by this PodSpec. If specified, these secrets will be passed to individual puller implementations for them to use. More info: https://kubernetes.io/docs/concepts/containers/images#specifying-imagepullsecrets-on-a-pod |
| `nodeSelector` | _object (keys:string, values:string)_ |  false  | NodeSelector is a selector which must be true for the pod to fit on a node. Selector which must match a node's labels for the pod to be scheduled on that node. More info: https://kubernetes.io/docs/concepts/configuration/assign-pod-node/ |
| `topologySpreadConstraints` | _[TopologySpreadConstraint](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#topologyspreadconstraint-v1-core) array_ |  false  | TopologySpreadConstraints describes how a group of pods ought to spread across topology domains. Scheduler will schedule pods in a way which abides by the constraints. All topologySpreadConstraints are ANDed. |


#### KubernetesServiceSpec



KubernetesServiceSpec defines the desired state of the Kubernetes service resource.

_Appears in:_
- [EnvoyProxyKubernetesProvider](#envoyproxykubernetesprovider)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `annotations` | _object (keys:string, values:string)_ |  false  | Annotations that should be appended to the service. By default, no annotations are appended. |
| `type` | _[ServiceType](#servicetype)_ |  false  | Type determines how the Service is exposed. Defaults to LoadBalancer. Valid options are ClusterIP, LoadBalancer and NodePort. "LoadBalancer" means a service will be exposed via an external load balancer (if the cloud provider supports it). "ClusterIP" means a service will only be accessible inside the cluster, via the cluster IP. "NodePort" means a service will be exposed on a static Port on all Nodes of the cluster. |
| `loadBalancerClass` | _string_ |  false  | LoadBalancerClass, when specified, allows for choosing the LoadBalancer provider implementation if more than one are available or is otherwise expected to be specified |
| `allocateLoadBalancerNodePorts` | _boolean_ |  false  | AllocateLoadBalancerNodePorts defines if NodePorts will be automatically allocated for services with type LoadBalancer. Default is "true". It may be set to "false" if the cluster load-balancer does not rely on NodePorts. If the caller requests specific NodePorts (by specifying a value), those requests will be respected, regardless of this field. This field may only be set for services with type LoadBalancer and will be cleared if the type is changed to any other type. |
| `loadBalancerIP` | _string_ |  false  | LoadBalancerIP defines the IP Address of the underlying load balancer service. This field may be ignored if the load balancer provider does not support this feature. This field has been deprecated in Kubernetes, but it is still used for setting the IP Address in some cloud providers such as GCP. |
| `externalTrafficPolicy` | _[ServiceExternalTrafficPolicy](#serviceexternaltrafficpolicy)_ |  false  | ExternalTrafficPolicy determines the externalTrafficPolicy for the Envoy Service. Valid options are Local and Cluster. Default is "Local". "Local" means traffic will only go to pods on the node receiving the traffic. "Cluster" means connections are loadbalanced to all pods in the cluster. |
| `patch` | _[KubernetesPatchSpec](#kubernetespatchspec)_ |  false  | Patch defines how to perform the patch operation to the service |


#### KubernetesWatchMode



KubernetesWatchMode holds the configuration for which input resources to watch and reconcile.

_Appears in:_
- [EnvoyGatewayKubernetesProvider](#envoygatewaykubernetesprovider)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `type` | _[KubernetesWatchModeType](#kuberneteswatchmodetype)_ |  true  | Type indicates what watch mode to use. KubernetesWatchModeTypeNamespaces and KubernetesWatchModeTypeNamespaceSelector are currently supported By default, when this field is unset or empty, Envoy Gateway will watch for input namespaced resources from all namespaces. |
| `namespaces` | _string array_ |  true  | Namespaces holds the list of namespaces that Envoy Gateway will watch for namespaced scoped resources such as Gateway, HTTPRoute and Service. Note that Envoy Gateway will continue to reconcile relevant cluster scoped resources such as GatewayClass that it is linked to. Precisely one of Namespaces and NamespaceSelector must be set. |
| `namespaceSelector` | _[LabelSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#labelselector-v1-meta)_ |  true  | NamespaceSelector holds the label selector used to dynamically select namespaces. Envoy Gateway will watch for namespaces matching the specified label selector. Precisely one of Namespaces and NamespaceSelector must be set. |


#### KubernetesWatchModeType

_Underlying type:_ _string_

KubernetesWatchModeType defines the type of KubernetesWatchMode

_Appears in:_
- [KubernetesWatchMode](#kuberneteswatchmode)



#### LiteralCustomTag



LiteralCustomTag adds hard-coded value to each span.

_Appears in:_
- [CustomTag](#customtag)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `value` | _string_ |  true  | Value defines the hard-coded value to add to each span. |


#### LoadBalancer



LoadBalancer defines the load balancer policy to be applied.

_Appears in:_
- [BackendTrafficPolicySpec](#backendtrafficpolicyspec)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `type` | _[LoadBalancerType](#loadbalancertype)_ |  true  | Type decides the type of Load Balancer policy. Valid LoadBalancerType values are "ConsistentHash", "LeastRequest", "Random", "RoundRobin", |
| `consistentHash` | _[ConsistentHash](#consistenthash)_ |  false  | ConsistentHash defines the configuration when the load balancer type is set to ConsistentHash |
| `slowStart` | _[SlowStart](#slowstart)_ |  false  | SlowStart defines the configuration related to the slow start load balancer policy. If set, during slow start window, traffic sent to the newly added hosts will gradually increase. Currently this is only supported for RoundRobin and LeastRequest load balancers |


#### LoadBalancerType

_Underlying type:_ _string_

LoadBalancerType specifies the types of LoadBalancer.

_Appears in:_
- [LoadBalancer](#loadbalancer)



#### LocalRateLimit



LocalRateLimit defines local rate limit configuration.

_Appears in:_
- [RateLimitSpec](#ratelimitspec)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `rules` | _[RateLimitRule](#ratelimitrule) array_ |  false  | Rules are a list of RateLimit selectors and limits. If a request matches multiple rules, the strictest limit is applied. For example, if a request matches two rules, one with 10rps and one with 20rps, the final limit will be based on the rule with 10rps. |


#### LogLevel

_Underlying type:_ _string_

LogLevel defines a log level for Envoy Gateway and EnvoyProxy system logs.

_Appears in:_
- [EnvoyGatewayLogging](#envoygatewaylogging)
- [ProxyLogging](#proxylogging)





#### MetricSinkType

_Underlying type:_ _string_



_Appears in:_
- [EnvoyGatewayMetricSink](#envoygatewaymetricsink)
- [ProxyMetricSink](#proxymetricsink)



#### OIDC



OIDC defines the configuration for the OpenID Connect (OIDC) authentication.

_Appears in:_
- [SecurityPolicySpec](#securitypolicyspec)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `provider` | _[OIDCProvider](#oidcprovider)_ |  true  | The OIDC Provider configuration. |
| `clientID` | _string_ |  true  | The client ID to be used in the OIDC [Authentication Request](https://openid.net/specs/openid-connect-core-1_0.html#AuthRequest). |
| `clientSecret` | _[SecretObjectReference](https://gateway-api.sigs.k8s.io/references/spec/#gateway.networking.k8s.io/v1.SecretObjectReference)_ |  true  | The Kubernetes secret which contains the OIDC client secret to be used in the [Authentication Request](https://openid.net/specs/openid-connect-core-1_0.html#AuthRequest). <br /><br /> This is an Opaque secret. The client secret should be stored in the key "client-secret". |
| `scopes` | _string array_ |  false  | The OIDC scopes to be used in the [Authentication Request](https://openid.net/specs/openid-connect-core-1_0.html#AuthRequest). The "openid" scope is always added to the list of scopes if not already specified. |
| `redirectURL` | _string_ |  true  | The redirect URL to be used in the OIDC [Authentication Request](https://openid.net/specs/openid-connect-core-1_0.html#AuthRequest). If not specified, uses the default redirect URI "%REQ(x-forwarded-proto)%://%REQ(:authority)%/oauth2/callback" |
| `logoutPath` | _string_ |  true  | The path to log a user out, clearing their credential cookies. If not specified, uses a default logout path "/logout" |


#### OIDCProvider



OIDCProvider defines the OIDC Provider configuration.

_Appears in:_
- [OIDC](#oidc)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `issuer` | _string_ |  true  | The OIDC Provider's [issuer identifier](https://openid.net/specs/openid-connect-discovery-1_0.html#IssuerDiscovery). Issuer MUST be a URI RFC 3986 [RFC3986] with a scheme component that MUST be https, a host component, and optionally, port and path components and no query or fragment components. |
| `authorizationEndpoint` | _string_ |  false  | The OIDC Provider's [authorization endpoint](https://openid.net/specs/openid-connect-core-1_0.html#AuthorizationEndpoint). If not provided, EG will try to discover it from the provider's [Well-Known Configuration Endpoint](https://openid.net/specs/openid-connect-discovery-1_0.html#ProviderConfigurationResponse). |
| `tokenEndpoint` | _string_ |  false  | The OIDC Provider's [token endpoint](https://openid.net/specs/openid-connect-core-1_0.html#TokenEndpoint). If not provided, EG will try to discover it from the provider's [Well-Known Configuration Endpoint](https://openid.net/specs/openid-connect-discovery-1_0.html#ProviderConfigurationResponse). |


#### OpenTelemetryEnvoyProxyAccessLog



TODO: consider reuse ExtensionService?

_Appears in:_
- [ProxyAccessLogSink](#proxyaccesslogsink)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `host` | _string_ |  true  | Host define the extension service hostname. |
| `port` | _integer_ |  false  | Port defines the port the extension service is exposed on. |
| `resources` | _object (keys:string, values:string)_ |  false  | Resources is a set of labels that describe the source of a log entry, including envoy node info. It's recommended to follow [semantic conventions](https://opentelemetry.io/docs/reference/specification/resource/semantic_conventions/). |


#### Origin

_Underlying type:_ _string_

Origin is defined by the scheme (protocol), hostname (domain), and port of the URL used to access it. The hostname can be "precise" which is just the domain name or "wildcard" which is a domain name prefixed with a single wildcard label such as "*.example.com". In addition to that a single wildcard (with or without scheme) can be configured to match any origin. 
 For example, the following are valid origins: - https://foo.example.com - https://*.example.com - http://foo.example.com:8080 - http://*.example.com:8080 - https://*

_Appears in:_
- [CORS](#cors)



#### PassiveHealthCheck



PassiveHealthCheck defines the configuration for passive health checks in the context of Envoy's Outlier Detection, see https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/upstream/outlier

_Appears in:_
- [HealthCheck](#healthcheck)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `splitExternalLocalOriginErrors` | _boolean_ |  false  | SplitExternalLocalOriginErrors enables splitting of errors between external and local origin. |
| `interval` | _[Duration](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#duration-v1-meta)_ |  false  | Interval defines the time between passive health checks. |
| `consecutiveLocalOriginFailures` | _integer_ |  false  | ConsecutiveLocalOriginFailures sets the number of consecutive local origin failures triggering ejection. Parameter takes effect only when split_external_local_origin_errors is set to true. |
| `consecutiveGatewayErrors` | _integer_ |  false  | ConsecutiveGatewayErrors sets the number of consecutive gateway errors triggering ejection. |
| `consecutive5XxErrors` | _integer_ |  false  | Consecutive5xxErrors sets the number of consecutive 5xx errors triggering ejection. |
| `baseEjectionTime` | _[Duration](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#duration-v1-meta)_ |  false  | BaseEjectionTime defines the base duration for which a host will be ejected on consecutive failures. |
| `maxEjectionPercent` | _integer_ |  false  | MaxEjectionPercent sets the maximum percentage of hosts in a cluster that can be ejected. |


#### PathEscapedSlashAction

_Underlying type:_ _string_

PathEscapedSlashAction determines the action for requests that contain %2F, %2f, %5C, or %5c sequences in the URI path.

_Appears in:_
- [PathSettings](#pathsettings)



#### PathSettings



PathSettings provides settings that managing how the incoming path set by clients is handled.

_Appears in:_
- [ClientTrafficPolicySpec](#clienttrafficpolicyspec)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `escapedSlashesAction` | _[PathEscapedSlashAction](#pathescapedslashaction)_ |  false  | EscapedSlashesAction determines how %2f, %2F, %5c, or %5C sequences in the path URI should be handled. The default is UnescapeAndRedirect. |
| `disableMergeSlashes` | _boolean_ |  false  | DisableMergeSlashes allows disabling the default configuration of merging adjacent slashes in the path. Note that slash merging is not part of the HTTP spec and is provided for convenience. |


#### PerRetryPolicy





_Appears in:_
- [Retry](#retry)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `timeout` | _[Duration](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#duration-v1-meta)_ |  false  | Timeout is the timeout per retry attempt. |
| `backOff` | _[BackOffPolicy](#backoffpolicy)_ |  false  | Backoff is the backoff policy to be applied per retry attempt. gateway uses a fully jittered exponential back-off algorithm for retries. For additional details, see https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/router_filter#config-http-filters-router-x-envoy-max-retries |


#### ProviderType

_Underlying type:_ _string_

ProviderType defines the types of providers supported by Envoy Gateway.

_Appears in:_
- [EnvoyGatewayProvider](#envoygatewayprovider)
- [EnvoyProxyProvider](#envoyproxyprovider)



#### ProxyAccessLog





_Appears in:_
- [ProxyTelemetry](#proxytelemetry)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `disable` | _boolean_ |  true  | Disable disables access logging for managed proxies if set to true. |
| `settings` | _[ProxyAccessLogSetting](#proxyaccesslogsetting) array_ |  false  | Settings defines accesslog settings for managed proxies. If unspecified, will send default format to stdout. |


#### ProxyAccessLogFormat



ProxyAccessLogFormat defines the format of accesslog. By default accesslogs are written to standard output.

_Appears in:_
- [ProxyAccessLogSetting](#proxyaccesslogsetting)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `type` | _[ProxyAccessLogFormatType](#proxyaccesslogformattype)_ |  true  | Type defines the type of accesslog format. |
| `text` | _string_ |  false  | Text defines the text accesslog format, following Envoy accesslog formatting, It's required when the format type is "Text". Envoy [command operators](https://www.envoyproxy.io/docs/envoy/latest/configuration/observability/access_log/usage#command-operators) may be used in the format. The [format string documentation](https://www.envoyproxy.io/docs/envoy/latest/configuration/observability/access_log/usage#config-access-log-format-strings) provides more information. |
| `json` | _object (keys:string, values:string)_ |  false  | JSON is additional attributes that describe the specific event occurrence. Structured format for the envoy access logs. Envoy [command operators](https://www.envoyproxy.io/docs/envoy/latest/configuration/observability/access_log/usage#command-operators) can be used as values for fields within the Struct. It's required when the format type is "JSON". |


#### ProxyAccessLogFormatType

_Underlying type:_ _string_



_Appears in:_
- [ProxyAccessLogFormat](#proxyaccesslogformat)



#### ProxyAccessLogSetting





_Appears in:_
- [ProxyAccessLog](#proxyaccesslog)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `format` | _[ProxyAccessLogFormat](#proxyaccesslogformat)_ |  true  | Format defines the format of accesslog. |
| `sinks` | _[ProxyAccessLogSink](#proxyaccesslogsink) array_ |  true  | Sinks defines the sinks of accesslog. |


#### ProxyAccessLogSink



ProxyAccessLogSink defines the sink of accesslog.

_Appears in:_
- [ProxyAccessLogSetting](#proxyaccesslogsetting)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `type` | _[ProxyAccessLogSinkType](#proxyaccesslogsinktype)_ |  true  | Type defines the type of accesslog sink. |
| `file` | _[FileEnvoyProxyAccessLog](#fileenvoyproxyaccesslog)_ |  false  | File defines the file accesslog sink. |
| `openTelemetry` | _[OpenTelemetryEnvoyProxyAccessLog](#opentelemetryenvoyproxyaccesslog)_ |  false  | OpenTelemetry defines the OpenTelemetry accesslog sink. |


#### ProxyAccessLogSinkType

_Underlying type:_ _string_



_Appears in:_
- [ProxyAccessLogSink](#proxyaccesslogsink)



#### ProxyBootstrap



ProxyBootstrap defines Envoy Bootstrap configuration.

_Appears in:_
- [EnvoyProxySpec](#envoyproxyspec)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `type` | _[BootstrapType](#bootstraptype)_ |  false  | Type is the type of the bootstrap configuration, it should be either Replace or Merge. If unspecified, it defaults to Replace. |
| `value` | _string_ |  true  | Value is a YAML string of the bootstrap. |


#### ProxyLogComponent

_Underlying type:_ _string_

ProxyLogComponent defines a component that supports a configured logging level.

_Appears in:_
- [ProxyLogging](#proxylogging)



#### ProxyLogging



ProxyLogging defines logging parameters for managed proxies.

_Appears in:_
- [EnvoyProxySpec](#envoyproxyspec)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `level` | _object (keys:[ProxyLogComponent](#proxylogcomponent), values:[LogLevel](#loglevel))_ |  true  | Level is a map of logging level per component, where the component is the key and the log level is the value. If unspecified, defaults to "default: warn". |


#### ProxyMetricSink



ProxyMetricSink defines the sink of metrics. Default metrics sink is OpenTelemetry.

_Appears in:_
- [ProxyMetrics](#proxymetrics)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `type` | _[MetricSinkType](#metricsinktype)_ |  true  | Type defines the metric sink type. EG currently only supports OpenTelemetry. |
| `openTelemetry` | _[ProxyOpenTelemetrySink](#proxyopentelemetrysink)_ |  false  | OpenTelemetry defines the configuration for OpenTelemetry sink. It's required if the sink type is OpenTelemetry. |


#### ProxyMetrics





_Appears in:_
- [ProxyTelemetry](#proxytelemetry)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `prometheus` | _[ProxyPrometheusProvider](#proxyprometheusprovider)_ |  true  | Prometheus defines the configuration for Admin endpoint `/stats/prometheus`. |
| `sinks` | _[ProxyMetricSink](#proxymetricsink) array_ |  true  | Sinks defines the metric sinks where metrics are sent to. |
| `matches` | _[StringMatch](#stringmatch) array_ |  true  | Matches defines configuration for selecting specific metrics instead of generating all metrics stats that are enabled by default. This helps reduce CPU and memory overhead in Envoy, but eliminating some stats may after critical functionality. Here are the stats that we strongly recommend not disabling: `cluster_manager.warming_clusters`, `cluster.<cluster_name>.membership_total`,`cluster.<cluster_name>.membership_healthy`, `cluster.<cluster_name>.membership_degraded`，reference  https://github.com/envoyproxy/envoy/issues/9856, https://github.com/envoyproxy/envoy/issues/14610 |
| `enableVirtualHostStats` | _boolean_ |  true  | EnableVirtualHostStats enables envoy stat metrics for virtual hosts. |


#### ProxyOpenTelemetrySink





_Appears in:_
- [ProxyMetricSink](#proxymetricsink)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `host` | _string_ |  true  | Host define the service hostname. |
| `port` | _integer_ |  false  | Port defines the port the service is exposed on. |


#### ProxyPrometheusProvider





_Appears in:_
- [ProxyMetrics](#proxymetrics)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `disable` | _boolean_ |  true  | Disable the Prometheus endpoint. |


#### ProxyProtocol



ProxyProtocol defines the configuration related to the proxy protocol when communicating with the backend.

_Appears in:_
- [BackendTrafficPolicySpec](#backendtrafficpolicyspec)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `version` | _[ProxyProtocolVersion](#proxyprotocolversion)_ |  true  | Version of ProxyProtol Valid ProxyProtocolVersion values are "V1" "V2" |


#### ProxyProtocolVersion

_Underlying type:_ _string_

ProxyProtocolVersion defines the version of the Proxy Protocol to use.

_Appears in:_
- [ProxyProtocol](#proxyprotocol)



#### ProxyTelemetry





_Appears in:_
- [EnvoyProxySpec](#envoyproxyspec)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `accessLog` | _[ProxyAccessLog](#proxyaccesslog)_ |  false  | AccessLogs defines accesslog parameters for managed proxies. If unspecified, will send default format to stdout. |
| `tracing` | _[ProxyTracing](#proxytracing)_ |  false  | Tracing defines tracing configuration for managed proxies. If unspecified, will not send tracing data. |
| `metrics` | _[ProxyMetrics](#proxymetrics)_ |  true  | Metrics defines metrics configuration for managed proxies. |


#### ProxyTracing





_Appears in:_
- [ProxyTelemetry](#proxytelemetry)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `samplingRate` | _integer_ |  false  | SamplingRate controls the rate at which traffic will be selected for tracing if no prior sampling decision has been made. Defaults to 100, valid values [0-100]. 100 indicates 100% sampling. |
| `customTags` | _object (keys:string, values:[CustomTag](#customtag))_ |  true  | CustomTags defines the custom tags to add to each span. If provider is kubernetes, pod name and namespace are added by default. |
| `provider` | _[TracingProvider](#tracingprovider)_ |  true  | Provider defines the tracing provider. Only OpenTelemetry is supported currently. |


#### RateLimit



RateLimit defines the configuration associated with the Rate Limit Service used for Global Rate Limiting.

_Appears in:_
- [EnvoyGateway](#envoygateway)
- [EnvoyGatewaySpec](#envoygatewayspec)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `backend` | _[RateLimitDatabaseBackend](#ratelimitdatabasebackend)_ |  true  | Backend holds the configuration associated with the database backend used by the rate limit service to store state associated with global ratelimiting. |
| `timeout` | _[Duration](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#duration-v1-meta)_ |  false  | Timeout specifies the timeout period for the proxy to access the ratelimit server If not set, timeout is 20ms. |
| `failClosed` | _boolean_ |  true  | FailClosed is a switch used to control the flow of traffic when the response from the ratelimit server cannot be obtained. If FailClosed is false, let the traffic pass, otherwise, don't let the traffic pass and return 500. If not set, FailClosed is False. |
| `telemetry` | _[RateLimitTelemetry](#ratelimittelemetry)_ |  false  | Telemetry defines telemetry configuration for RateLimit. |


#### RateLimitDatabaseBackend



RateLimitDatabaseBackend defines the configuration associated with the database backend used by the rate limit service.

_Appears in:_
- [RateLimit](#ratelimit)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `type` | _[RateLimitDatabaseBackendType](#ratelimitdatabasebackendtype)_ |  true  | Type is the type of database backend to use. Supported types are: * Redis: Connects to a Redis database. |
| `redis` | _[RateLimitRedisSettings](#ratelimitredissettings)_ |  false  | Redis defines the settings needed to connect to a Redis database. |


#### RateLimitDatabaseBackendType

_Underlying type:_ _string_

RateLimitDatabaseBackendType specifies the types of database backend to be used by the rate limit service.

_Appears in:_
- [RateLimitDatabaseBackend](#ratelimitdatabasebackend)



#### RateLimitMetrics





_Appears in:_
- [RateLimitTelemetry](#ratelimittelemetry)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `prometheus` | _[RateLimitMetricsPrometheusProvider](#ratelimitmetricsprometheusprovider)_ |  true  | Prometheus defines the configuration for prometheus endpoint. |


#### RateLimitMetricsPrometheusProvider





_Appears in:_
- [RateLimitMetrics](#ratelimitmetrics)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `disable` | _boolean_ |  true  | Disable the Prometheus endpoint. |


#### RateLimitRedisSettings



RateLimitRedisSettings defines the configuration for connecting to redis database.

_Appears in:_
- [RateLimitDatabaseBackend](#ratelimitdatabasebackend)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `url` | _string_ |  true  | URL of the Redis Database. |
| `tls` | _[RedisTLSSettings](#redistlssettings)_ |  false  | TLS defines TLS configuration for connecting to redis database. |


#### RateLimitRule



RateLimitRule defines the semantics for matching attributes from the incoming requests, and setting limits for them.

_Appears in:_
- [GlobalRateLimit](#globalratelimit)
- [LocalRateLimit](#localratelimit)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `clientSelectors` | _[RateLimitSelectCondition](#ratelimitselectcondition) array_ |  false  | ClientSelectors holds the list of select conditions to select specific clients using attributes from the traffic flow. All individual select conditions must hold True for this rule and its limit to be applied. <br /><br /> If no client selectors are specified, the rule applies to all traffic of the targeted Route. <br /><br /> If the policy targets a Gateway, the rule applies to each Route of the Gateway. Please note that each Route has its own rate limit counters. For example, if a Gateway has two Routes, and the policy has a rule with limit 10rps, each Route will have its own 10rps limit. |
| `limit` | _[RateLimitValue](#ratelimitvalue)_ |  true  | Limit holds the rate limit values. This limit is applied for traffic flows when the selectors compute to True, causing the request to be counted towards the limit. The limit is enforced and the request is ratelimited, i.e. a response with 429 HTTP status code is sent back to the client when the selected requests have reached the limit. |


#### RateLimitSelectCondition



RateLimitSelectCondition specifies the attributes within the traffic flow that can be used to select a subset of clients to be ratelimited. All the individual conditions must hold True for the overall condition to hold True.

_Appears in:_
- [RateLimitRule](#ratelimitrule)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `headers` | _[HeaderMatch](#headermatch) array_ |  false  | Headers is a list of request headers to match. Multiple header values are ANDed together, meaning, a request MUST match all the specified headers. At least one of headers or sourceCIDR condition must be specified. |
| `sourceCIDR` | _[SourceMatch](#sourcematch)_ |  false  | SourceCIDR is the client IP Address range to match on. At least one of headers or sourceCIDR condition must be specified. |


#### RateLimitSpec



RateLimitSpec defines the desired state of RateLimitSpec.

_Appears in:_
- [BackendTrafficPolicySpec](#backendtrafficpolicyspec)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `type` | _[RateLimitType](#ratelimittype)_ |  true  | Type decides the scope for the RateLimits. Valid RateLimitType values are "Global" or "Local". |
| `global` | _[GlobalRateLimit](#globalratelimit)_ |  false  | Global defines global rate limit configuration. |
| `local` | _[LocalRateLimit](#localratelimit)_ |  false  | Local defines local rate limit configuration. |


#### RateLimitTelemetry





_Appears in:_
- [RateLimit](#ratelimit)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `metrics` | _[RateLimitMetrics](#ratelimitmetrics)_ |  true  | Metrics defines metrics configuration for RateLimit. |


#### RateLimitType

_Underlying type:_ _string_

RateLimitType specifies the types of RateLimiting.

_Appears in:_
- [RateLimitSpec](#ratelimitspec)



#### RateLimitUnit

_Underlying type:_ _string_

RateLimitUnit specifies the intervals for setting rate limits. Valid RateLimitUnit values are "Second", "Minute", "Hour", and "Day".

_Appears in:_
- [RateLimitValue](#ratelimitvalue)



#### RateLimitValue



RateLimitValue defines the limits for rate limiting.

_Appears in:_
- [RateLimitRule](#ratelimitrule)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `requests` | _integer_ |  true  |  |
| `unit` | _[RateLimitUnit](#ratelimitunit)_ |  true  |  |


#### RedisTLSSettings



RedisTLSSettings defines the TLS configuration for connecting to redis database.

_Appears in:_
- [RateLimitRedisSettings](#ratelimitredissettings)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `certificateRef` | _[SecretObjectReference](https://gateway-api.sigs.k8s.io/references/spec/#gateway.networking.k8s.io/v1.SecretObjectReference)_ |  false  | CertificateRef defines the client certificate reference for TLS connections. Currently only a Kubernetes Secret of type TLS is supported. |


#### RemoteJWKS



RemoteJWKS defines how to fetch and cache JSON Web Key Sets (JWKS) from a remote HTTP/HTTPS endpoint.

_Appears in:_
- [JWTProvider](#jwtprovider)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `uri` | _string_ |  true  | URI is the HTTPS URI to fetch the JWKS. Envoy's system trust bundle is used to validate the server certificate. |


#### RequestHeaderCustomTag



RequestHeaderCustomTag adds value from request header to each span.

_Appears in:_
- [CustomTag](#customtag)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `name` | _string_ |  true  | Name defines the name of the request header which to extract the value from. |
| `defaultValue` | _string_ |  false  | DefaultValue defines the default value to use if the request header is not set. |


#### ResourceProviderType

_Underlying type:_ _string_

ResourceProviderType defines the types of custom resource providers supported by Envoy Gateway.

_Appears in:_
- [EnvoyGatewayResourceProvider](#envoygatewayresourceprovider)



#### Retry



Retry defines the retry strategy to be applied.

_Appears in:_
- [BackendTrafficPolicySpec](#backendtrafficpolicyspec)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `numRetries` | _integer_ |  false  | NumRetries is the number of retries to be attempted. Defaults to 2. |
| `retryOn` | _[RetryOn](#retryon)_ |  false  | RetryOn specifies the retry trigger condition. <br /><br /> If not specified, the default is to retry on connect-failure,refused-stream,unavailable,cancelled,retriable-status-codes(503). |
| `perRetry` | _[PerRetryPolicy](#perretrypolicy)_ |  false  | PerRetry is the retry policy to be applied per retry attempt. |


#### RetryOn





_Appears in:_
- [Retry](#retry)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `triggers` | _[TriggerEnum](#triggerenum) array_ |  false  | Triggers specifies the retry trigger condition(Http/Grpc). |
| `httpStatusCodes` | _[HTTPStatus](#httpstatus) array_ |  false  | HttpStatusCodes specifies the http status codes to be retried. The retriable-status-codes trigger must also be configured for these status codes to trigger a retry. |


#### SecurityPolicy



SecurityPolicy allows the user to configure various security settings for a Gateway.

_Appears in:_
- [SecurityPolicyList](#securitypolicylist)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `apiVersion` | _string_ | |`gateway.envoyproxy.io/v1alpha1`
| `kind` | _string_ | |`SecurityPolicy`
| `metadata` | _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#objectmeta-v1-meta)_ |  true  | Refer to Kubernetes API documentation for fields of `metadata`. |
| `spec` | _[SecurityPolicySpec](#securitypolicyspec)_ |  true  | Spec defines the desired state of SecurityPolicy. |


#### SecurityPolicyList



SecurityPolicyList contains a list of SecurityPolicy resources.



| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `apiVersion` | _string_ | |`gateway.envoyproxy.io/v1alpha1`
| `kind` | _string_ | |`SecurityPolicyList`
| `metadata` | _[ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#listmeta-v1-meta)_ |  true  | Refer to Kubernetes API documentation for fields of `metadata`. |
| `items` | _[SecurityPolicy](#securitypolicy) array_ |  true  |  |


#### SecurityPolicySpec



SecurityPolicySpec defines the desired state of SecurityPolicy.

_Appears in:_
- [SecurityPolicy](#securitypolicy)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `targetRef` | _[PolicyTargetReferenceWithSectionName](https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1alpha2.PolicyTargetReferenceWithSectionName)_ |  true  | TargetRef is the name of the Gateway resource this policy is being attached to. This Policy and the TargetRef MUST be in the same namespace for this Policy to have effect and be applied to the Gateway. |
| `cors` | _[CORS](#cors)_ |  false  | CORS defines the configuration for Cross-Origin Resource Sharing (CORS). |
| `basicAuth` | _[BasicAuth](#basicauth)_ |  false  | BasicAuth defines the configuration for the HTTP Basic Authentication. |
| `jwt` | _[JWT](#jwt)_ |  false  | JWT defines the configuration for JSON Web Token (JWT) authentication. |
| `oidc` | _[OIDC](#oidc)_ |  false  | OIDC defines the configuration for the OpenID Connect (OIDC) authentication. |
| `extAuth` | _[ExtAuth](#extauth)_ |  false  | ExtAuth defines the configuration for External Authorization. |




#### ServiceExternalTrafficPolicy

_Underlying type:_ _string_

ServiceExternalTrafficPolicy describes how nodes distribute service traffic they receive on one of the Service's "externally-facing" addresses (NodePorts, ExternalIPs, and LoadBalancer IPs.

_Appears in:_
- [KubernetesServiceSpec](#kubernetesservicespec)



#### ServiceType

_Underlying type:_ _string_

ServiceType string describes ingress methods for a service

_Appears in:_
- [KubernetesServiceSpec](#kubernetesservicespec)



#### ShutdownConfig



ShutdownConfig defines configuration for graceful envoy shutdown process.

_Appears in:_
- [EnvoyProxySpec](#envoyproxyspec)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `drainTimeout` | _[Duration](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#duration-v1-meta)_ |  false  | DrainTimeout defines the graceful drain timeout. This should be less than the pod's terminationGracePeriodSeconds. If unspecified, defaults to 600 seconds. |
| `minDrainDuration` | _[Duration](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#duration-v1-meta)_ |  false  | MinDrainDuration defines the minimum drain duration allowing time for endpoint deprogramming to complete. If unspecified, defaults to 5 seconds. |


#### SlowStart



SlowStart defines the configuration related to the slow start load balancer policy.

_Appears in:_
- [LoadBalancer](#loadbalancer)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `window` | _[Duration](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#duration-v1-meta)_ |  true  | Window defines the duration of the warm up period for newly added host. During slow start window, traffic sent to the newly added hosts will gradually increase. Currently only supports linear growth of traffic. For additional details, see https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/cluster/v3/cluster.proto#config-cluster-v3-cluster-slowstartconfig |




#### SourceMatchType

_Underlying type:_ _string_



_Appears in:_
- [SourceMatch](#sourcematch)



#### StringMatch



StringMatch defines how to match any strings. This is a general purpose match condition that can be used by other EG APIs that need to match against a string.

_Appears in:_
- [ProxyMetrics](#proxymetrics)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `type` | _[StringMatchType](#stringmatchtype)_ |  false  | Type specifies how to match against a string. |
| `value` | _string_ |  true  | Value specifies the string value that the match must have. |


#### StringMatchType

_Underlying type:_ _string_

StringMatchType specifies the semantics of how a string value should be compared. Valid MatchType values are "Exact", "Prefix", "Suffix", "RegularExpression".

_Appears in:_
- [StringMatch](#stringmatch)



#### TCPActiveHealthChecker



TCPActiveHealthChecker defines the settings of tcp health check.

_Appears in:_
- [ActiveHealthCheck](#activehealthcheck)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `send` | _[ActiveHealthCheckPayload](#activehealthcheckpayload)_ |  false  | Send defines the request payload. |
| `receive` | _[ActiveHealthCheckPayload](#activehealthcheckpayload)_ |  false  | Receive defines the expected response payload. |


#### TCPKeepalive



TCPKeepalive define the TCP Keepalive configuration.

_Appears in:_
- [BackendTrafficPolicySpec](#backendtrafficpolicyspec)
- [ClientTrafficPolicySpec](#clienttrafficpolicyspec)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `probes` | _integer_ |  false  | The total number of unacknowledged probes to send before deciding the connection is dead. Defaults to 9. |
| `idleTime` | _[Duration](#duration)_ |  false  | The duration a connection needs to be idle before keep-alive probes start being sent. The duration format is Defaults to `7200s`. |
| `interval` | _[Duration](#duration)_ |  false  | The duration between keep-alive probes. Defaults to `75s`. |


#### TCPTimeout





_Appears in:_
- [Timeout](#timeout)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `connectTimeout` | _[Duration](#duration)_ |  false  | The timeout for network connection establishment, including TCP and TLS handshakes. Default: 10 seconds. |


#### TLSSettings





_Appears in:_
- [ClientTrafficPolicySpec](#clienttrafficpolicyspec)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `minVersion` | _[TLSVersion](#tlsversion)_ |  false  | Min specifies the minimal TLS protocol version to allow. The default is TLS 1.2 if this is not specified. |
| `maxVersion` | _[TLSVersion](#tlsversion)_ |  false  | Max specifies the maximal TLS protocol version to allow The default is TLS 1.3 if this is not specified. |
| `ciphers` | _string array_ |  false  | Ciphers specifies the set of cipher suites supported when negotiating TLS 1.0 - 1.2. This setting has no effect for TLS 1.3. In non-FIPS Envoy Proxy builds the default cipher list is: - [ECDHE-ECDSA-AES128-GCM-SHA256\|ECDHE-ECDSA-CHACHA20-POLY1305] - [ECDHE-RSA-AES128-GCM-SHA256\|ECDHE-RSA-CHACHA20-POLY1305] - ECDHE-ECDSA-AES256-GCM-SHA384 - ECDHE-RSA-AES256-GCM-SHA384 In builds using BoringSSL FIPS the default cipher list is: - ECDHE-ECDSA-AES128-GCM-SHA256 - ECDHE-RSA-AES128-GCM-SHA256 - ECDHE-ECDSA-AES256-GCM-SHA384 - ECDHE-RSA-AES256-GCM-SHA384 |
| `ecdhCurves` | _string array_ |  false  | ECDHCurves specifies the set of supported ECDH curves. In non-FIPS Envoy Proxy builds the default curves are: - X25519 - P-256 In builds using BoringSSL FIPS the default curve is: - P-256 |
| `signatureAlgorithms` | _string array_ |  false  | SignatureAlgorithms specifies which signature algorithms the listener should support. |
| `alpnProtocols` | _[ALPNProtocol](#alpnprotocol) array_ |  false  | ALPNProtocols supplies the list of ALPN protocols that should be exposed by the listener. By default h2 and http/1.1 are enabled. Supported values are: - http/1.0 - http/1.1 - h2 |
| `clientValidation` | _[ClientValidationContext](#clientvalidationcontext)_ |  false  | ClientValidation specifies the configuration to validate the client initiating the TLS connection to the Gateway listener. |


#### TLSVersion

_Underlying type:_ _string_

TLSVersion specifies the TLS version

_Appears in:_
- [TLSSettings](#tlssettings)



#### Timeout



Timeout defines configuration for timeouts related to connections.

_Appears in:_
- [BackendTrafficPolicySpec](#backendtrafficpolicyspec)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `tcp` | _[TCPTimeout](#tcptimeout)_ |  false  | Timeout settings for TCP. |
| `http` | _[HTTPTimeout](#httptimeout)_ |  false  | Timeout settings for HTTP. |


#### TracingProvider





_Appears in:_
- [ProxyTracing](#proxytracing)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `type` | _[TracingProviderType](#tracingprovidertype)_ |  true  | Type defines the tracing provider type. EG currently only supports OpenTelemetry. |
| `host` | _string_ |  true  | Host define the provider service hostname. |
| `port` | _integer_ |  false  | Port defines the port the provider service is exposed on. |


#### TracingProviderType

_Underlying type:_ _string_



_Appears in:_
- [TracingProvider](#tracingprovider)



#### TriggerEnum

_Underlying type:_ _string_

TriggerEnum specifies the conditions that trigger retries.

_Appears in:_
- [RetryOn](#retryon)



#### XDSTranslatorHook

_Underlying type:_ _string_

XDSTranslatorHook defines the types of hooks that an Envoy Gateway extension may support for the xds-translator

_Appears in:_
- [XDSTranslatorHooks](#xdstranslatorhooks)



#### XDSTranslatorHooks



XDSTranslatorHooks contains all the pre and post hooks for the xds-translator runner.

_Appears in:_
- [ExtensionHooks](#extensionhooks)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `pre` | _[XDSTranslatorHook](#xdstranslatorhook) array_ |  true  |  |
| `post` | _[XDSTranslatorHook](#xdstranslatorhook) array_ |  true  |  |


#### XForwardedForSettings



XForwardedForSettings provides configuration for using X-Forwarded-For headers for determining the client IP address.

_Appears in:_
- [ClientIPDetectionSettings](#clientipdetectionsettings)

| Field | Type | Required | Description |
| ---   | ---  | ---      | ---         |
| `numTrustedHops` | _integer_ |  false  | NumTrustedHops controls the number of additional ingress proxy hops from the right side of XFF HTTP headers to trust when determining the origin client's IP address. Refer to https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_conn_man/headers#x-forwarded-for for more details. |


