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



#### BackendTrafficPolicy



BackendTrafficPolicy allows the user to configure the behavior of the connection between the downstream client and Envoy Proxy listener.

_Appears in:_
- [BackendTrafficPolicyList](#backendtrafficpolicylist)

| Field | Description |
| --- | --- |
| `apiVersion` _string_ | `gateway.envoyproxy.io/v1alpha1`
| `kind` _string_ | `BackendTrafficPolicy`
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |
| `spec` _[BackendTrafficPolicySpec](#backendtrafficpolicyspec)_ | spec defines the desired state of BackendTrafficPolicy. |


#### BackendTrafficPolicyList



BackendTrafficPolicyList contains a list of BackendTrafficPolicy resources.



| Field | Description |
| --- | --- |
| `apiVersion` _string_ | `gateway.envoyproxy.io/v1alpha1`
| `kind` _string_ | `BackendTrafficPolicyList`
| `metadata` _[ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#listmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |
| `items` _[BackendTrafficPolicy](#backendtrafficpolicy) array_ |  |


#### BackendTrafficPolicySpec



spec defines the desired state of BackendTrafficPolicy.

_Appears in:_
- [BackendTrafficPolicy](#backendtrafficpolicy)

| Field | Description |
| --- | --- |
| `targetRef` _[PolicyTargetReferenceWithSectionName](#policytargetreferencewithsectionname)_ | targetRef is the name of the resource this policy is being attached to. This Policy and the TargetRef MUST be in the same namespace for this Policy to have effect and be applied to the Gateway. |
| `rateLimit` _[RateLimitSpec](#ratelimitspec)_ | RateLimit allows the user to limit the number of incoming requests to a predefined value based on attributes within the traffic flow. |
| `loadBalancer` _[LoadBalancer](#loadbalancer)_ | LoadBalancer policy to apply when routing traffic from the gateway to the backend endpoints |




#### BootstrapType

_Underlying type:_ `string`

BootstrapType defines the types of bootstrap supported by Envoy Gateway.

_Appears in:_
- [ProxyBootstrap](#proxybootstrap)



#### CORS



CORS defines the configuration for Cross-Origin Resource Sharing (CORS).

_Appears in:_
- [SecurityPolicySpec](#securitypolicyspec)

| Field | Description |
| --- | --- |
| `allowOrigins` _[StringMatch](#stringmatch) array_ | AllowOrigins defines the origins that are allowed to make requests. |
| `allowMethods` _string array_ | AllowMethods defines the methods that are allowed to make requests. |
| `allowHeaders` _string array_ | AllowHeaders defines the headers that are allowed to be sent with requests. |
| `exposeHeaders` _string array_ | ExposeHeaders defines the headers that can be exposed in the responses. |
| `maxAge` _[Duration](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#duration-v1-meta)_ | MaxAge defines how long the results of a preflight request can be cached. |


#### ClaimToHeader



ClaimToHeader defines a configuration to convert JWT claims into HTTP headers

_Appears in:_
- [JWTProvider](#jwtprovider)

| Field | Description |
| --- | --- |
| `header` _string_ | Header defines the name of the HTTP request header that the JWT Claim will be saved into. |
| `claim` _string_ | Claim is the JWT Claim that should be saved into the header : it can be a nested claim of type (eg. "claim.nested.key", "sub"). The nested claim name must use dot "." to separate the JSON name path. |


#### ClientTrafficPolicy



ClientTrafficPolicy allows the user to configure the behavior of the connection between the downstream client and Envoy Proxy listener.

_Appears in:_
- [ClientTrafficPolicyList](#clienttrafficpolicylist)

| Field | Description |
| --- | --- |
| `apiVersion` _string_ | `gateway.envoyproxy.io/v1alpha1`
| `kind` _string_ | `ClientTrafficPolicy`
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |
| `spec` _[ClientTrafficPolicySpec](#clienttrafficpolicyspec)_ | Spec defines the desired state of ClientTrafficPolicy. |


#### ClientTrafficPolicyList



ClientTrafficPolicyList contains a list of ClientTrafficPolicy resources.



| Field | Description |
| --- | --- |
| `apiVersion` _string_ | `gateway.envoyproxy.io/v1alpha1`
| `kind` _string_ | `ClientTrafficPolicyList`
| `metadata` _[ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#listmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |
| `items` _[ClientTrafficPolicy](#clienttrafficpolicy) array_ |  |


#### ClientTrafficPolicySpec



ClientTrafficPolicySpec defines the desired state of ClientTrafficPolicy.

_Appears in:_
- [ClientTrafficPolicy](#clienttrafficpolicy)

| Field | Description |
| --- | --- |
| `targetRef` _[PolicyTargetReferenceWithSectionName](#policytargetreferencewithsectionname)_ | TargetRef is the name of the Gateway resource this policy is being attached to. This Policy and the TargetRef MUST be in the same namespace for this Policy to have effect and be applied to the Gateway. TargetRef |
| `tcpKeepalive` _[TCPKeepalive](#tcpkeepalive)_ | TcpKeepalive settings associated with the downstream client connection. If defined, sets SO_KEEPALIVE on the listener socket to enable TCP Keepalives. Disabled by default. |




#### ConsistentHash



ConsistentHash defines the configuration related to the consistent hash load balancer policy

_Appears in:_
- [LoadBalancer](#loadbalancer)

| Field | Description |
| --- | --- |
| `type` _[ConsistentHashType](#consistenthashtype)_ |  |


#### ConsistentHashType

_Underlying type:_ `string`

ConsistentHashType defines the type of input to hash on.

_Appears in:_
- [ConsistentHash](#consistenthash)



#### CustomTag





_Appears in:_
- [ProxyTracing](#proxytracing)

| Field | Description |
| --- | --- |
| `type` _[CustomTagType](#customtagtype)_ | Type defines the type of custom tag. |
| `literal` _[LiteralCustomTag](#literalcustomtag)_ | Literal adds hard-coded value to each span. It's required when the type is "Literal". |
| `environment` _[EnvironmentCustomTag](#environmentcustomtag)_ | Environment adds value from environment variable to each span. It's required when the type is "Environment". |
| `requestHeader` _[RequestHeaderCustomTag](#requestheadercustomtag)_ | RequestHeader adds value from request header to each span. It's required when the type is "RequestHeader". |


#### CustomTagType

_Underlying type:_ `string`



_Appears in:_
- [CustomTag](#customtag)



#### EnvironmentCustomTag



EnvironmentCustomTag adds value from environment variable to each span.

_Appears in:_
- [CustomTag](#customtag)

| Field | Description |
| --- | --- |
| `name` _string_ | Name defines the name of the environment variable which to extract the value from. |
| `defaultValue` _string_ | DefaultValue defines the default value to use if the environment variable is not set. |


#### EnvoyGateway



EnvoyGateway is the schema for the envoygateways API.



| Field | Description |
| --- | --- |
| `apiVersion` _string_ | `gateway.envoyproxy.io/v1alpha1`
| `kind` _string_ | `EnvoyGateway`
| `gateway` _[Gateway](#gateway)_ | Gateway defines desired Gateway API specific configuration. If unset, default configuration parameters will apply. |
| `provider` _[EnvoyGatewayProvider](#envoygatewayprovider)_ | Provider defines the desired provider and provider-specific configuration. If unspecified, the Kubernetes provider is used with default configuration parameters. |
| `logging` _[EnvoyGatewayLogging](#envoygatewaylogging)_ | Logging defines logging parameters for Envoy Gateway. |
| `admin` _[EnvoyGatewayAdmin](#envoygatewayadmin)_ | Admin defines the desired admin related abilities. If unspecified, the Admin is used with default configuration parameters. |
| `telemetry` _[EnvoyGatewayTelemetry](#envoygatewaytelemetry)_ | Telemetry defines the desired control plane telemetry related abilities. If unspecified, the telemetry is used with default configuration. |
| `rateLimit` _[RateLimit](#ratelimit)_ | RateLimit defines the configuration associated with the Rate Limit service deployed by Envoy Gateway required to implement the Global Rate limiting functionality. The specific rate limit service used here is the reference implementation in Envoy. For more details visit https://github.com/envoyproxy/ratelimit. This configuration is unneeded for "Local" rate limiting. |
| `extensionManager` _[ExtensionManager](#extensionmanager)_ | ExtensionManager defines an extension manager to register for the Envoy Gateway Control Plane. |
| `extensionApis` _[ExtensionAPISettings](#extensionapisettings)_ | ExtensionAPIs defines the settings related to specific Gateway API Extensions implemented by Envoy Gateway |


#### EnvoyGatewayAdmin



EnvoyGatewayAdmin defines the Envoy Gateway Admin configuration.

_Appears in:_
- [EnvoyGateway](#envoygateway)
- [EnvoyGatewaySpec](#envoygatewayspec)

| Field | Description |
| --- | --- |
| `address` _[EnvoyGatewayAdminAddress](#envoygatewayadminaddress)_ | Address defines the address of Envoy Gateway Admin Server. |
| `enableDumpConfig` _boolean_ | EnableDumpConfig defines if enable dump config in Envoy Gateway logs. |
| `enablePprof` _boolean_ | EnablePprof defines if enable pprof in Envoy Gateway Admin Server. |


#### EnvoyGatewayAdminAddress



EnvoyGatewayAdminAddress defines the Envoy Gateway Admin Address configuration.

_Appears in:_
- [EnvoyGatewayAdmin](#envoygatewayadmin)

| Field | Description |
| --- | --- |
| `port` _integer_ | Port defines the port the admin server is exposed on. |
| `host` _string_ | Host defines the admin server hostname. |


#### EnvoyGatewayCustomProvider



EnvoyGatewayCustomProvider defines configuration for the Custom provider.

_Appears in:_
- [EnvoyGatewayProvider](#envoygatewayprovider)

| Field | Description |
| --- | --- |
| `resource` _[EnvoyGatewayResourceProvider](#envoygatewayresourceprovider)_ | Resource defines the desired resource provider. This provider is used to specify the provider to be used to retrieve the resource configurations such as Gateway API resources |
| `infrastructure` _[EnvoyGatewayInfrastructureProvider](#envoygatewayinfrastructureprovider)_ | Infrastructure defines the desired infrastructure provider. This provider is used to specify the provider to be used to provide an environment to deploy the out resources like the Envoy Proxy data plane. |


#### EnvoyGatewayFileResourceProvider



EnvoyGatewayFileResourceProvider defines configuration for the File Resource provider.

_Appears in:_
- [EnvoyGatewayResourceProvider](#envoygatewayresourceprovider)

| Field | Description |
| --- | --- |
| `paths` _string array_ | Paths are the paths to a directory or file containing the resource configuration. Recursive sub directories are not currently supported. |


#### EnvoyGatewayHostInfrastructureProvider



EnvoyGatewayHostInfrastructureProvider defines configuration for the Host Infrastructure provider.

_Appears in:_
- [EnvoyGatewayInfrastructureProvider](#envoygatewayinfrastructureprovider)



#### EnvoyGatewayInfrastructureProvider



EnvoyGatewayInfrastructureProvider defines configuration for the Custom Infrastructure provider.

_Appears in:_
- [EnvoyGatewayCustomProvider](#envoygatewaycustomprovider)

| Field | Description |
| --- | --- |
| `type` _[InfrastructureProviderType](#infrastructureprovidertype)_ | Type is the type of infrastructure providers to use. Supported types are "Host". |
| `host` _[EnvoyGatewayHostInfrastructureProvider](#envoygatewayhostinfrastructureprovider)_ | Host defines the configuration of the Host provider. Host provides runtime deployment of the data plane as a child process on the host environment. |


#### EnvoyGatewayKubernetesProvider



EnvoyGatewayKubernetesProvider defines configuration for the Kubernetes provider.

_Appears in:_
- [EnvoyGatewayProvider](#envoygatewayprovider)

| Field | Description |
| --- | --- |
| `rateLimitDeployment` _[KubernetesDeploymentSpec](#kubernetesdeploymentspec)_ | RateLimitDeployment defines the desired state of the Envoy ratelimit deployment resource. If unspecified, default settings for the managed Envoy ratelimit deployment resource are applied. |
| `watch` _[KubernetesWatchMode](#kuberneteswatchmode)_ | Watch holds configuration of which input resources should be watched and reconciled. |
| `deploy` _[KubernetesDeployMode](#kubernetesdeploymode)_ | Deploy holds configuration of how output managed resources such as the Envoy Proxy data plane should be deployed |
| `overwrite_control_plane_certs` _boolean_ | OverwriteControlPlaneCerts updates the secrets containing the control plane certs, when set. |


#### EnvoyGatewayLogComponent

_Underlying type:_ `string`

EnvoyGatewayLogComponent defines a component that supports a configured logging level.

_Appears in:_
- [EnvoyGatewayLogging](#envoygatewaylogging)



#### EnvoyGatewayLogging



EnvoyGatewayLogging defines logging for Envoy Gateway.

_Appears in:_
- [EnvoyGateway](#envoygateway)
- [EnvoyGatewaySpec](#envoygatewayspec)

| Field | Description |
| --- | --- |
| `level` _object (keys:[EnvoyGatewayLogComponent](#envoygatewaylogcomponent), values:[LogLevel](#loglevel))_ | Level is the logging level. If unspecified, defaults to "info". EnvoyGatewayLogComponent options: default/provider/gateway-api/xds-translator/xds-server/infrastructure/global-ratelimit. LogLevel options: debug/info/error/warn. |


#### EnvoyGatewayMetricSink



EnvoyGatewayMetricSink defines control plane metric sinks where metrics are sent to.

_Appears in:_
- [EnvoyGatewayMetrics](#envoygatewaymetrics)

| Field | Description |
| --- | --- |
| `type` _[MetricSinkType](#metricsinktype)_ | Type defines the metric sink type. EG control plane currently supports OpenTelemetry. |
| `openTelemetry` _[EnvoyGatewayOpenTelemetrySink](#envoygatewayopentelemetrysink)_ | OpenTelemetry defines the configuration for OpenTelemetry sink. It's required if the sink type is OpenTelemetry. |


#### EnvoyGatewayMetrics



EnvoyGatewayMetrics defines control plane push/pull metrics configurations.

_Appears in:_
- [EnvoyGatewayTelemetry](#envoygatewaytelemetry)

| Field | Description |
| --- | --- |
| `sinks` _[EnvoyGatewayMetricSink](#envoygatewaymetricsink) array_ | Sinks defines the metric sinks where metrics are sent to. |
| `prometheus` _[EnvoyGatewayPrometheusProvider](#envoygatewayprometheusprovider)_ | Prometheus defines the configuration for prometheus endpoint. |


#### EnvoyGatewayOpenTelemetrySink





_Appears in:_
- [EnvoyGatewayMetricSink](#envoygatewaymetricsink)

| Field | Description |
| --- | --- |
| `host` _string_ | Host define the sink service hostname. |
| `protocol` _string_ | Protocol define the sink service protocol. |
| `port` _integer_ | Port defines the port the sink service is exposed on. |


#### EnvoyGatewayPrometheusProvider



EnvoyGatewayPrometheusProvider will expose prometheus endpoint in pull mode.

_Appears in:_
- [EnvoyGatewayMetrics](#envoygatewaymetrics)

| Field | Description |
| --- | --- |
| `disable` _boolean_ | Disable defines if disables the prometheus metrics in pull mode. |


#### EnvoyGatewayProvider



EnvoyGatewayProvider defines the desired configuration of a provider.

_Appears in:_
- [EnvoyGateway](#envoygateway)
- [EnvoyGatewaySpec](#envoygatewayspec)

| Field | Description |
| --- | --- |
| `type` _[ProviderType](#providertype)_ | Type is the type of provider to use. Supported types are "Kubernetes". |
| `kubernetes` _[EnvoyGatewayKubernetesProvider](#envoygatewaykubernetesprovider)_ | Kubernetes defines the configuration of the Kubernetes provider. Kubernetes provides runtime configuration via the Kubernetes API. |
| `custom` _[EnvoyGatewayCustomProvider](#envoygatewaycustomprovider)_ | Custom defines the configuration for the Custom provider. This provider allows you to define a specific resource provider and a infrastructure provider. |


#### EnvoyGatewayResourceProvider



EnvoyGatewayResourceProvider defines configuration for the Custom Resource provider.

_Appears in:_
- [EnvoyGatewayCustomProvider](#envoygatewaycustomprovider)

| Field | Description |
| --- | --- |
| `type` _[ResourceProviderType](#resourceprovidertype)_ | Type is the type of resource provider to use. Supported types are "File". |
| `file` _[EnvoyGatewayFileResourceProvider](#envoygatewayfileresourceprovider)_ | File defines the configuration of the File provider. File provides runtime configuration defined by one or more files. |


#### EnvoyGatewaySpec



EnvoyGatewaySpec defines the desired state of Envoy Gateway.

_Appears in:_
- [EnvoyGateway](#envoygateway)

| Field | Description |
| --- | --- |
| `gateway` _[Gateway](#gateway)_ | Gateway defines desired Gateway API specific configuration. If unset, default configuration parameters will apply. |
| `provider` _[EnvoyGatewayProvider](#envoygatewayprovider)_ | Provider defines the desired provider and provider-specific configuration. If unspecified, the Kubernetes provider is used with default configuration parameters. |
| `logging` _[EnvoyGatewayLogging](#envoygatewaylogging)_ | Logging defines logging parameters for Envoy Gateway. |
| `admin` _[EnvoyGatewayAdmin](#envoygatewayadmin)_ | Admin defines the desired admin related abilities. If unspecified, the Admin is used with default configuration parameters. |
| `telemetry` _[EnvoyGatewayTelemetry](#envoygatewaytelemetry)_ | Telemetry defines the desired control plane telemetry related abilities. If unspecified, the telemetry is used with default configuration. |
| `rateLimit` _[RateLimit](#ratelimit)_ | RateLimit defines the configuration associated with the Rate Limit service deployed by Envoy Gateway required to implement the Global Rate limiting functionality. The specific rate limit service used here is the reference implementation in Envoy. For more details visit https://github.com/envoyproxy/ratelimit. This configuration is unneeded for "Local" rate limiting. |
| `extensionManager` _[ExtensionManager](#extensionmanager)_ | ExtensionManager defines an extension manager to register for the Envoy Gateway Control Plane. |
| `extensionApis` _[ExtensionAPISettings](#extensionapisettings)_ | ExtensionAPIs defines the settings related to specific Gateway API Extensions implemented by Envoy Gateway |


#### EnvoyGatewayTelemetry



EnvoyGatewayTelemetry defines telemetry configurations for envoy gateway control plane. Control plane will focus on metrics observability telemetry and tracing telemetry later.

_Appears in:_
- [EnvoyGateway](#envoygateway)
- [EnvoyGatewaySpec](#envoygatewayspec)

| Field | Description |
| --- | --- |
| `metrics` _[EnvoyGatewayMetrics](#envoygatewaymetrics)_ | Metrics defines metrics configuration for envoy gateway. |


#### EnvoyJSONPatchConfig



EnvoyJSONPatchConfig defines the configuration for patching a Envoy xDS Resource using JSONPatch semantic

_Appears in:_
- [EnvoyPatchPolicySpec](#envoypatchpolicyspec)

| Field | Description |
| --- | --- |
| `type` _[EnvoyResourceType](#envoyresourcetype)_ | Type is the typed URL of the Envoy xDS Resource |
| `name` _string_ | Name is the name of the resource |
| `operation` _[JSONPatchOperation](#jsonpatchoperation)_ | Patch defines the JSON Patch Operation |


#### EnvoyPatchPolicy



EnvoyPatchPolicy allows the user to modify the generated Envoy xDS resources by Envoy Gateway using this patch API

_Appears in:_
- [EnvoyPatchPolicyList](#envoypatchpolicylist)

| Field | Description |
| --- | --- |
| `apiVersion` _string_ | `gateway.envoyproxy.io/v1alpha1`
| `kind` _string_ | `EnvoyPatchPolicy`
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |
| `spec` _[EnvoyPatchPolicySpec](#envoypatchpolicyspec)_ | Spec defines the desired state of EnvoyPatchPolicy. |


#### EnvoyPatchPolicyList



EnvoyPatchPolicyList contains a list of EnvoyPatchPolicy resources.



| Field | Description |
| --- | --- |
| `apiVersion` _string_ | `gateway.envoyproxy.io/v1alpha1`
| `kind` _string_ | `EnvoyPatchPolicyList`
| `metadata` _[ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#listmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |
| `items` _[EnvoyPatchPolicy](#envoypatchpolicy) array_ |  |


#### EnvoyPatchPolicySpec



EnvoyPatchPolicySpec defines the desired state of EnvoyPatchPolicy.

_Appears in:_
- [EnvoyPatchPolicy](#envoypatchpolicy)

| Field | Description |
| --- | --- |
| `type` _[EnvoyPatchType](#envoypatchtype)_ | Type decides the type of patch. Valid EnvoyPatchType values are "JSONPatch". |
| `jsonPatches` _[EnvoyJSONPatchConfig](#envoyjsonpatchconfig) array_ | JSONPatch defines the JSONPatch configuration. |
| `targetRef` _[PolicyTargetReference](#policytargetreference)_ | TargetRef is the name of the Gateway API resource this policy is being attached to. Currently only attaching to Gateway is supported This Policy and the TargetRef MUST be in the same namespace for this Policy to have effect and be applied to the Gateway TargetRef |
| `priority` _integer_ | Priority of the EnvoyPatchPolicy. If multiple EnvoyPatchPolicies are applied to the same TargetRef, they will be applied in the ascending order of the priority i.e. int32.min has the highest priority and int32.max has the lowest priority. Defaults to 0. |




#### EnvoyPatchType

_Underlying type:_ `string`

EnvoyPatchType specifies the types of Envoy patching mechanisms.

_Appears in:_
- [EnvoyPatchPolicySpec](#envoypatchpolicyspec)



#### EnvoyProxy



EnvoyProxy is the schema for the envoyproxies API.



| Field | Description |
| --- | --- |
| `apiVersion` _string_ | `gateway.envoyproxy.io/v1alpha1`
| `kind` _string_ | `EnvoyProxy`
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |
| `spec` _[EnvoyProxySpec](#envoyproxyspec)_ | EnvoyProxySpec defines the desired state of EnvoyProxy. |


#### EnvoyProxyKubernetesProvider



EnvoyProxyKubernetesProvider defines configuration for the Kubernetes resource provider.

_Appears in:_
- [EnvoyProxyProvider](#envoyproxyprovider)

| Field | Description |
| --- | --- |
| `envoyDeployment` _[KubernetesDeploymentSpec](#kubernetesdeploymentspec)_ | EnvoyDeployment defines the desired state of the Envoy deployment resource. If unspecified, default settings for the managed Envoy deployment resource are applied. |
| `envoyService` _[KubernetesServiceSpec](#kubernetesservicespec)_ | EnvoyService defines the desired state of the Envoy service resource. If unspecified, default settings for the managed Envoy service resource are applied. |


#### EnvoyProxyProvider



EnvoyProxyProvider defines the desired state of a resource provider.

_Appears in:_
- [EnvoyProxySpec](#envoyproxyspec)

| Field | Description |
| --- | --- |
| `type` _[ProviderType](#providertype)_ | Type is the type of resource provider to use. A resource provider provides infrastructure resources for running the data plane, e.g. Envoy proxy, and optional auxiliary control planes. Supported types are "Kubernetes". |
| `kubernetes` _[EnvoyProxyKubernetesProvider](#envoyproxykubernetesprovider)_ | Kubernetes defines the desired state of the Kubernetes resource provider. Kubernetes provides infrastructure resources for running the data plane, e.g. Envoy proxy. If unspecified and type is "Kubernetes", default settings for managed Kubernetes resources are applied. |


#### EnvoyProxySpec



EnvoyProxySpec defines the desired state of EnvoyProxy.

_Appears in:_
- [EnvoyProxy](#envoyproxy)

| Field | Description |
| --- | --- |
| `provider` _[EnvoyProxyProvider](#envoyproxyprovider)_ | Provider defines the desired resource provider and provider-specific configuration. If unspecified, the "Kubernetes" resource provider is used with default configuration parameters. |
| `logging` _[ProxyLogging](#proxylogging)_ | Logging defines logging parameters for managed proxies. |
| `telemetry` _[ProxyTelemetry](#proxytelemetry)_ | Telemetry defines telemetry parameters for managed proxies. |
| `bootstrap` _[ProxyBootstrap](#proxybootstrap)_ | Bootstrap defines the Envoy Bootstrap as a YAML string. Visit https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/bootstrap/v3/bootstrap.proto#envoy-v3-api-msg-config-bootstrap-v3-bootstrap to learn more about the syntax. If set, this is the Bootstrap configuration used for the managed Envoy Proxy fleet instead of the default Bootstrap configuration set by Envoy Gateway. Some fields within the Bootstrap that are required to communicate with the xDS Server (Envoy Gateway) and receive xDS resources from it are not configurable and will result in the `EnvoyProxy` resource being rejected. Backward compatibility across minor versions is not guaranteed. We strongly recommend using `egctl x translate` to generate a `EnvoyProxy` resource with the `Bootstrap` field set to the default Bootstrap configuration used. You can edit this configuration, and rerun `egctl x translate` to ensure there are no validation errors. |
| `concurrency` _integer_ | Concurrency defines the number of worker threads to run. If unset, it defaults to the number of cpuset threads on the platform. |
| `mergeGateways` _boolean_ | MergeGateways defines if Gateway resources should be merged onto the same Envoy Proxy Infrastructure. Setting this field to true would merge all Gateway Listeners under the parent Gateway Class. This means that the port, protocol and hostname tuple must be unique for every listener. If a duplicate listener is detected, the newer listener (based on timestamp) will be rejected and its status will be updated with a "Accepted=False" condition. |




#### EnvoyResourceType

_Underlying type:_ `string`

EnvoyResourceType specifies the type URL of the Envoy resource.

_Appears in:_
- [EnvoyJSONPatchConfig](#envoyjsonpatchconfig)



#### ExtensionAPISettings



ExtensionAPISettings defines the settings specific to Gateway API Extensions.

_Appears in:_
- [EnvoyGateway](#envoygateway)
- [EnvoyGatewaySpec](#envoygatewayspec)

| Field | Description |
| --- | --- |
| `enableEnvoyPatchPolicy` _boolean_ | EnableEnvoyPatchPolicy enables Envoy Gateway to reconcile and implement the EnvoyPatchPolicy resources. |


#### ExtensionHooks



ExtensionHooks defines extension hooks across all supported runners

_Appears in:_
- [ExtensionManager](#extensionmanager)

| Field | Description |
| --- | --- |
| `xdsTranslator` _[XDSTranslatorHooks](#xdstranslatorhooks)_ | XDSTranslator defines all the supported extension hooks for the xds-translator runner |


#### ExtensionManager



ExtensionManager defines the configuration for registering an extension manager to the Envoy Gateway control plane.

_Appears in:_
- [EnvoyGateway](#envoygateway)
- [EnvoyGatewaySpec](#envoygatewayspec)

| Field | Description |
| --- | --- |
| `resources` _[GroupVersionKind](#groupversionkind) array_ | Resources defines the set of K8s resources the extension will handle. |
| `hooks` _[ExtensionHooks](#extensionhooks)_ | Hooks defines the set of hooks the extension supports |
| `service` _[ExtensionService](#extensionservice)_ | Service defines the configuration of the extension service that the Envoy Gateway Control Plane will call through extension hooks. |


#### ExtensionService



ExtensionService defines the configuration for connecting to a registered extension service.

_Appears in:_
- [ExtensionManager](#extensionmanager)

| Field | Description |
| --- | --- |
| `host` _string_ | Host define the extension service hostname. |
| `port` _integer_ | Port defines the port the extension service is exposed on. |
| `tls` _[ExtensionTLS](#extensiontls)_ | TLS defines TLS configuration for communication between Envoy Gateway and the extension service. |


#### ExtensionTLS



ExtensionTLS defines the TLS configuration when connecting to an extension service

_Appears in:_
- [ExtensionService](#extensionservice)

| Field | Description |
| --- | --- |
| `certificateRef` _[SecretObjectReference](https://gateway-api.sigs.k8s.io/references/spec/#gateway.networking.k8s.io/v1.SecretObjectReference)_ | CertificateRef contains a references to objects (Kubernetes objects or otherwise) that contains a TLS certificate and private keys. These certificates are used to establish a TLS handshake to the extension server. 
 CertificateRef can only reference a Kubernetes Secret at this time. |


#### FileEnvoyProxyAccessLog





_Appears in:_
- [ProxyAccessLogSink](#proxyaccesslogsink)

| Field | Description |
| --- | --- |
| `path` _string_ | Path defines the file path used to expose envoy access log(e.g. /dev/stdout). |


#### Gateway



Gateway defines the desired Gateway API configuration of Envoy Gateway.

_Appears in:_
- [EnvoyGateway](#envoygateway)
- [EnvoyGatewaySpec](#envoygatewayspec)

| Field | Description |
| --- | --- |
| `controllerName` _string_ | ControllerName defines the name of the Gateway API controller. If unspecified, defaults to "gateway.envoyproxy.io/gatewayclass-controller". See the following for additional details: https://gateway-api.sigs.k8s.io/v1alpha2/references/spec/#gateway.networking.k8s.io/v1.GatewayClass |


#### GlobalRateLimit



GlobalRateLimit defines global rate limit configuration.

_Appears in:_
- [RateLimitSpec](#ratelimitspec)

| Field | Description |
| --- | --- |
| `rules` _[RateLimitRule](#ratelimitrule) array_ | Rules are a list of RateLimit selectors and limits. Each rule and its associated limit is applied in a mutually exclusive way i.e. if multiple rules get selected, each of their associated limits get applied, so a single traffic request might increase the rate limit counters for multiple rules if selected. |


#### GroupVersionKind



GroupVersionKind unambiguously identifies a Kind. It can be converted to k8s.io/apimachinery/pkg/runtime/schema.GroupVersionKind

_Appears in:_
- [ExtensionManager](#extensionmanager)

| Field | Description |
| --- | --- |
| `group` _string_ |  |
| `version` _string_ |  |
| `kind` _string_ |  |


#### HeaderMatch



HeaderMatch defines the match attributes within the HTTP Headers of the request.

_Appears in:_
- [RateLimitSelectCondition](#ratelimitselectcondition)



#### InfrastructureProviderType

_Underlying type:_ `string`

InfrastructureProviderType defines the types of custom infrastructure providers supported by Envoy Gateway.

_Appears in:_
- [EnvoyGatewayInfrastructureProvider](#envoygatewayinfrastructureprovider)



#### JSONPatchOperation



JSONPatchOperation defines the JSON Patch Operation as defined in https://datatracker.ietf.org/doc/html/rfc6902

_Appears in:_
- [EnvoyJSONPatchConfig](#envoyjsonpatchconfig)

| Field | Description |
| --- | --- |
| `op` _[JSONPatchOperationType](#jsonpatchoperationtype)_ | Op is the type of operation to perform |
| `path` _string_ | Path is the location of the target document/field where the operation will be performed Refer to https://datatracker.ietf.org/doc/html/rfc6901 for more details. |
| `value` _[JSON](#json)_ | Value is the new value of the path location. |


#### JSONPatchOperationType

_Underlying type:_ `string`

JSONPatchOperationType specifies the JSON Patch operations that can be performed.

_Appears in:_
- [JSONPatchOperation](#jsonpatchoperation)



#### JWT



JWT defines the configuration for JSON Web Token (JWT) authentication.

_Appears in:_
- [SecurityPolicySpec](#securitypolicyspec)

| Field | Description |
| --- | --- |
| `providers` _[JWTProvider](#jwtprovider) array_ | Providers defines the JSON Web Token (JWT) authentication provider type. 
 When multiple JWT providers are specified, the JWT is considered valid if any of the providers successfully validate the JWT. For additional details, see https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/jwt_authn_filter.html. |


#### JWTProvider



JWTProvider defines how a JSON Web Token (JWT) can be verified.

_Appears in:_
- [JWT](#jwt)

| Field | Description |
| --- | --- |
| `name` _string_ | Name defines a unique name for the JWT provider. A name can have a variety of forms, including RFC1123 subdomains, RFC 1123 labels, or RFC 1035 labels. |
| `issuer` _string_ | Issuer is the principal that issued the JWT and takes the form of a URL or email address. For additional details, see https://tools.ietf.org/html/rfc7519#section-4.1.1 for URL format and https://rfc-editor.org/rfc/rfc5322.html for email format. If not provided, the JWT issuer is not checked. |
| `audiences` _string array_ | Audiences is a list of JWT audiences allowed access. For additional details, see https://tools.ietf.org/html/rfc7519#section-4.1.3. If not provided, JWT audiences are not checked. |
| `remoteJWKS` _[RemoteJWKS](#remotejwks)_ | RemoteJWKS defines how to fetch and cache JSON Web Key Sets (JWKS) from a remote HTTP/HTTPS endpoint. |
| `claimToHeaders` _[ClaimToHeader](#claimtoheader) array_ | ClaimToHeaders is a list of JWT claims that must be extracted into HTTP request headers For examples, following config: The claim must be of type; string, int, double, bool. Array type claims are not supported |


#### KubernetesContainerSpec



KubernetesContainerSpec defines the desired state of the Kubernetes container resource.

_Appears in:_
- [KubernetesDeploymentSpec](#kubernetesdeploymentspec)

| Field | Description |
| --- | --- |
| `env` _[EnvVar](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#envvar-v1-core) array_ | List of environment variables to set in the container. |
| `resources` _[ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#resourcerequirements-v1-core)_ | Resources required by this container. More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/ |
| `securityContext` _[SecurityContext](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#securitycontext-v1-core)_ | SecurityContext defines the security options the container should be run with. If set, the fields of SecurityContext override the equivalent fields of PodSecurityContext. More info: https://kubernetes.io/docs/tasks/configure-pod-container/security-context/ |
| `image` _string_ | Image specifies the EnvoyProxy container image to be used, instead of the default image. |
| `volumeMounts` _[VolumeMount](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#volumemount-v1-core) array_ | VolumeMounts are volumes to mount into the container's filesystem. Cannot be updated. |


#### KubernetesDeployMode



KubernetesDeployMode holds configuration for how to deploy managed resources such as the Envoy Proxy data plane fleet.

_Appears in:_
- [EnvoyGatewayKubernetesProvider](#envoygatewaykubernetesprovider)



#### KubernetesDeploymentSpec



KubernetesDeploymentSpec defines the desired state of the Kubernetes deployment resource.

_Appears in:_
- [EnvoyGatewayKubernetesProvider](#envoygatewaykubernetesprovider)
- [EnvoyProxyKubernetesProvider](#envoyproxykubernetesprovider)

| Field | Description |
| --- | --- |
| `replicas` _integer_ | Replicas is the number of desired pods. Defaults to 1. |
| `strategy` _[DeploymentStrategy](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#deploymentstrategy-v1-apps)_ | The deployment strategy to use to replace existing pods with new ones. |
| `pod` _[KubernetesPodSpec](#kubernetespodspec)_ | Pod defines the desired specification of pod. |
| `container` _[KubernetesContainerSpec](#kubernetescontainerspec)_ | Container defines the desired specification of main container. |
| `initContainers` _[Container](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#container-v1-core) array_ | List of initialization containers belonging to the pod. More info: https://kubernetes.io/docs/concepts/workloads/pods/init-containers/ |


#### KubernetesPodSpec



KubernetesPodSpec defines the desired state of the Kubernetes pod resource.

_Appears in:_
- [KubernetesDeploymentSpec](#kubernetesdeploymentspec)

| Field | Description |
| --- | --- |
| `annotations` _object (keys:string, values:string)_ | Annotations are the annotations that should be appended to the pods. By default, no pod annotations are appended. |
| `labels` _object (keys:string, values:string)_ | Labels are the additional labels that should be tagged to the pods. By default, no additional pod labels are tagged. |
| `securityContext` _[PodSecurityContext](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#podsecuritycontext-v1-core)_ | SecurityContext holds pod-level security attributes and common container settings. Optional: Defaults to empty.  See type description for default values of each field. |
| `affinity` _[Affinity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#affinity-v1-core)_ | If specified, the pod's scheduling constraints. |
| `tolerations` _[Toleration](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#toleration-v1-core) array_ | If specified, the pod's tolerations. |
| `volumes` _[Volume](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#volume-v1-core) array_ | Volumes that can be mounted by containers belonging to the pod. More info: https://kubernetes.io/docs/concepts/storage/volumes |


#### KubernetesServiceSpec



KubernetesServiceSpec defines the desired state of the Kubernetes service resource.

_Appears in:_
- [EnvoyProxyKubernetesProvider](#envoyproxykubernetesprovider)

| Field | Description |
| --- | --- |
| `annotations` _object (keys:string, values:string)_ | Annotations that should be appended to the service. By default, no annotations are appended. |
| `type` _[ServiceType](#servicetype)_ | Type determines how the Service is exposed. Defaults to LoadBalancer. Valid options are ClusterIP, LoadBalancer and NodePort. "LoadBalancer" means a service will be exposed via an external load balancer (if the cloud provider supports it). "ClusterIP" means a service will only be accessible inside the cluster, via the cluster IP. "NodePort" means a service will be exposed on a static Port on all Nodes of the cluster. |
| `loadBalancerClass` _string_ | LoadBalancerClass, when specified, allows for choosing the LoadBalancer provider implementation if more than one are available or is otherwise expected to be specified |
| `allocateLoadBalancerNodePorts` _boolean_ | AllocateLoadBalancerNodePorts defines if NodePorts will be automatically allocated for services with type LoadBalancer. Default is "true". It may be set to "false" if the cluster load-balancer does not rely on NodePorts. If the caller requests specific NodePorts (by specifying a value), those requests will be respected, regardless of this field. This field may only be set for services with type LoadBalancer and will be cleared if the type is changed to any other type. |
| `loadBalancerIP` _string_ | LoadBalancerIP defines the IP Address of the underlying load balancer service. This field may be ignored if the load balancer provider does not support this feature. This field has been deprecated in Kubernetes, but it is still used for setting the IP Address in some cloud providers such as GCP. |


#### KubernetesWatchMode



KubernetesWatchMode holds the configuration for which input resources to watch and reconcile.

_Appears in:_
- [EnvoyGatewayKubernetesProvider](#envoygatewaykubernetesprovider)

| Field | Description |
| --- | --- |
| `Type` _[KubernetesWatchModeType](#kuberneteswatchmodetype)_ | Type indicates what watch mode to use. KubernetesWatchModeTypeNamespaces and KubernetesWatchModeTypeNamespaceSelectors are currently supported By default, when this field is unset or empty, Envoy Gateway will watch for input namespaced resources from all namespaces. |
| `Namespaces` _string array_ | Namespaces holds the list of namespaces that Envoy Gateway will watch for namespaced scoped resources such as Gateway, HTTPRoute and Service. Note that Envoy Gateway will continue to reconcile relevant cluster scoped resources such as GatewayClass that it is linked to. Precisely one of Namespaces and NamespaceSelectors must be set |
| `namespaces` _string array_ | NamespaceSelectors holds a list of labels that namespaces have to have in order to be watched. Note this doesn't set the informer to watch the namespaces with the given labels. Informer still watches all namespaces. But the events for objects whois namespce have no given labels will be filtered out. Precisely one of Namespaces and NamespaceSelectors must be set |


#### KubernetesWatchModeType

_Underlying type:_ `string`

KubernetesWatchModeType defines the type of KubernetesWatchMode

_Appears in:_
- [KubernetesWatchMode](#kuberneteswatchmode)



#### LiteralCustomTag



LiteralCustomTag adds hard-coded value to each span.

_Appears in:_
- [CustomTag](#customtag)

| Field | Description |
| --- | --- |
| `value` _string_ | Value defines the hard-coded value to add to each span. |


#### LoadBalancer



LoadBalancer defines the load balancer policy to be applied.

_Appears in:_
- [BackendTrafficPolicySpec](#backendtrafficpolicyspec)

| Field | Description |
| --- | --- |
| `type` _[LoadBalancerType](#loadbalancertype)_ | Type decides the type of Load Balancer policy. Valid LoadBalancerType values are "ConsistentHash", "LeastRequest", "Random", "RoundRobin", |
| `consistentHash` _[ConsistentHash](#consistenthash)_ | ConsistentHash defines the configuration when the load balancer type is set to ConsistentHash |


#### LoadBalancerType

_Underlying type:_ `string`

LoadBalancerType specifies the types of LoadBalancer.

_Appears in:_
- [LoadBalancer](#loadbalancer)



#### LogLevel

_Underlying type:_ `string`

LogLevel defines a log level for Envoy Gateway and EnvoyProxy system logs.

_Appears in:_
- [EnvoyGatewayLogging](#envoygatewaylogging)
- [ProxyLogging](#proxylogging)



#### Match



Match defines the stats match configuration.

_Appears in:_
- [ProxyMetrics](#proxymetrics)

| Field | Description |
| --- | --- |
| `type` _[MatcherType](#matchertype)_ | MatcherType defines the stats matcher type |
| `value` _string_ |  |


#### MatchType

_Underlying type:_ `string`

MatchType specifies the semantics of how a string value should be compared. Valid MatchType values are "Exact", "Prefix", "Suffix", "RegularExpression".

_Appears in:_
- [StringMatch](#stringmatch)



#### MatcherType

_Underlying type:_ `string`



_Appears in:_
- [Match](#match)



#### MetricSinkType

_Underlying type:_ `string`



_Appears in:_
- [EnvoyGatewayMetricSink](#envoygatewaymetricsink)
- [ProxyMetricSink](#proxymetricsink)



#### OpenTelemetryEnvoyProxyAccessLog



TODO: consider reuse ExtensionService?

_Appears in:_
- [ProxyAccessLogSink](#proxyaccesslogsink)

| Field | Description |
| --- | --- |
| `host` _string_ | Host define the extension service hostname. |
| `port` _integer_ | Port defines the port the extension service is exposed on. |
| `resources` _object (keys:string, values:string)_ | Resources is a set of labels that describe the source of a log entry, including envoy node info. It's recommended to follow [semantic conventions](https://opentelemetry.io/docs/reference/specification/resource/semantic_conventions/). |


#### ProviderType

_Underlying type:_ `string`

ProviderType defines the types of providers supported by Envoy Gateway.

_Appears in:_
- [EnvoyGatewayProvider](#envoygatewayprovider)
- [EnvoyProxyProvider](#envoyproxyprovider)



#### ProxyAccessLog





_Appears in:_
- [ProxyTelemetry](#proxytelemetry)

| Field | Description |
| --- | --- |
| `disable` _boolean_ | Disable disables access logging for managed proxies if set to true. |
| `settings` _[ProxyAccessLogSetting](#proxyaccesslogsetting) array_ | Settings defines accesslog settings for managed proxies. If unspecified, will send default format to stdout. |


#### ProxyAccessLogFormat



ProxyAccessLogFormat defines the format of accesslog. By default accesslogs are written to standard output.

_Appears in:_
- [ProxyAccessLogSetting](#proxyaccesslogsetting)

| Field | Description |
| --- | --- |
| `type` _[ProxyAccessLogFormatType](#proxyaccesslogformattype)_ | Type defines the type of accesslog format. |
| `text` _string_ | Text defines the text accesslog format, following Envoy accesslog formatting, It's required when the format type is "Text". Envoy [command operators](https://www.envoyproxy.io/docs/envoy/latest/configuration/observability/access_log/usage#command-operators) may be used in the format. The [format string documentation](https://www.envoyproxy.io/docs/envoy/latest/configuration/observability/access_log/usage#config-access-log-format-strings) provides more information. |
| `json` _object (keys:string, values:string)_ | JSON is additional attributes that describe the specific event occurrence. Structured format for the envoy access logs. Envoy [command operators](https://www.envoyproxy.io/docs/envoy/latest/configuration/observability/access_log/usage#command-operators) can be used as values for fields within the Struct. It's required when the format type is "JSON". |


#### ProxyAccessLogFormatType

_Underlying type:_ `string`



_Appears in:_
- [ProxyAccessLogFormat](#proxyaccesslogformat)



#### ProxyAccessLogSetting





_Appears in:_
- [ProxyAccessLog](#proxyaccesslog)

| Field | Description |
| --- | --- |
| `format` _[ProxyAccessLogFormat](#proxyaccesslogformat)_ | Format defines the format of accesslog. |
| `sinks` _[ProxyAccessLogSink](#proxyaccesslogsink) array_ | Sinks defines the sinks of accesslog. |


#### ProxyAccessLogSink





_Appears in:_
- [ProxyAccessLogSetting](#proxyaccesslogsetting)

| Field | Description |
| --- | --- |
| `type` _[ProxyAccessLogSinkType](#proxyaccesslogsinktype)_ | Type defines the type of accesslog sink. |
| `file` _[FileEnvoyProxyAccessLog](#fileenvoyproxyaccesslog)_ | File defines the file accesslog sink. |
| `openTelemetry` _[OpenTelemetryEnvoyProxyAccessLog](#opentelemetryenvoyproxyaccesslog)_ | OpenTelemetry defines the OpenTelemetry accesslog sink. |


#### ProxyAccessLogSinkType

_Underlying type:_ `string`



_Appears in:_
- [ProxyAccessLogSink](#proxyaccesslogsink)



#### ProxyBootstrap



ProxyBootstrap defines Envoy Bootstrap configuration.

_Appears in:_
- [EnvoyProxySpec](#envoyproxyspec)

| Field | Description |
| --- | --- |
| `type` _[BootstrapType](#bootstraptype)_ | Type is the type of the bootstrap configuration, it should be either Replace or Merge. If unspecified, it defaults to Replace. |
| `value` _string_ | Value is a YAML string of the bootstrap. |


#### ProxyLogComponent

_Underlying type:_ `string`

ProxyLogComponent defines a component that supports a configured logging level.

_Appears in:_
- [ProxyLogging](#proxylogging)



#### ProxyLogging



ProxyLogging defines logging parameters for managed proxies.

_Appears in:_
- [EnvoyProxySpec](#envoyproxyspec)

| Field | Description |
| --- | --- |
| `level` _object (keys:[ProxyLogComponent](#proxylogcomponent), values:[LogLevel](#loglevel))_ | Level is a map of logging level per component, where the component is the key and the log level is the value. If unspecified, defaults to "default: warn". |


#### ProxyMetricSink





_Appears in:_
- [ProxyMetrics](#proxymetrics)

| Field | Description |
| --- | --- |
| `type` _[MetricSinkType](#metricsinktype)_ | Type defines the metric sink type. EG currently only supports OpenTelemetry. |
| `openTelemetry` _[ProxyOpenTelemetrySink](#proxyopentelemetrysink)_ | OpenTelemetry defines the configuration for OpenTelemetry sink. It's required if the sink type is OpenTelemetry. |


#### ProxyMetrics





_Appears in:_
- [ProxyTelemetry](#proxytelemetry)

| Field | Description |
| --- | --- |
| `prometheus` _[ProxyPrometheusProvider](#proxyprometheusprovider)_ | Prometheus defines the configuration for Admin endpoint `/stats/prometheus`. |
| `sinks` _[ProxyMetricSink](#proxymetricsink) array_ | Sinks defines the metric sinks where metrics are sent to. |
| `matches` _[Match](#match) array_ | Matches defines configuration for selecting specific metrics instead of generating all metrics stats that are enabled by default. This helps reduce CPU and memory overhead in Envoy, but eliminating some stats may after critical functionality. Here are the stats that we strongly recommend not disabling: `cluster_manager.warming_clusters`, `cluster.<cluster_name>.membership_total`,`cluster.<cluster_name>.membership_healthy`, `cluster.<cluster_name>.membership_degraded`，reference  https://github.com/envoyproxy/envoy/issues/9856, https://github.com/envoyproxy/envoy/issues/14610 |
| `enableVirtualHostStats` _boolean_ | EnableVirtualHostStats enables envoy stat metrics for virtual hosts. |


#### ProxyOpenTelemetrySink





_Appears in:_
- [ProxyMetricSink](#proxymetricsink)

| Field | Description |
| --- | --- |
| `host` _string_ | Host define the service hostname. |
| `port` _integer_ | Port defines the port the service is exposed on. |


#### ProxyPrometheusProvider





_Appears in:_
- [ProxyMetrics](#proxymetrics)

| Field | Description |
| --- | --- |
| `disable` _boolean_ | Disable the Prometheus endpoint. |


#### ProxyTelemetry





_Appears in:_
- [EnvoyProxySpec](#envoyproxyspec)

| Field | Description |
| --- | --- |
| `accessLog` _[ProxyAccessLog](#proxyaccesslog)_ | AccessLogs defines accesslog parameters for managed proxies. If unspecified, will send default format to stdout. |
| `tracing` _[ProxyTracing](#proxytracing)_ | Tracing defines tracing configuration for managed proxies. If unspecified, will not send tracing data. |
| `metrics` _[ProxyMetrics](#proxymetrics)_ | Metrics defines metrics configuration for managed proxies. |


#### ProxyTracing





_Appears in:_
- [ProxyTelemetry](#proxytelemetry)

| Field | Description |
| --- | --- |
| `samplingRate` _integer_ | SamplingRate controls the rate at which traffic will be selected for tracing if no prior sampling decision has been made. Defaults to 100, valid values [0-100]. 100 indicates 100% sampling. |
| `customTags` _object (keys:string, values:[CustomTag](#customtag))_ | CustomTags defines the custom tags to add to each span. If provider is kubernetes, pod name and namespace are added by default. |
| `provider` _[TracingProvider](#tracingprovider)_ | Provider defines the tracing provider. Only OpenTelemetry is supported currently. |


#### RateLimit



RateLimit defines the configuration associated with the Rate Limit Service used for Global Rate Limiting.

_Appears in:_
- [EnvoyGateway](#envoygateway)
- [EnvoyGatewaySpec](#envoygatewayspec)

| Field | Description |
| --- | --- |
| `backend` _[RateLimitDatabaseBackend](#ratelimitdatabasebackend)_ | Backend holds the configuration associated with the database backend used by the rate limit service to store state associated with global ratelimiting. |
| `timeout` _[Duration](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#duration-v1-meta)_ | Timeout specifies the timeout period for the proxy to access the ratelimit server If not set, timeout is 20ms. |
| `failClosed` _boolean_ | FailClosed is a switch used to control the flow of traffic when the response from the ratelimit server cannot be obtained. If FailClosed is false, let the traffic pass, otherwise, don't let the traffic pass and return 500. If not set, FailClosed is False. |


#### RateLimitDatabaseBackend



RateLimitDatabaseBackend defines the configuration associated with the database backend used by the rate limit service.

_Appears in:_
- [RateLimit](#ratelimit)

| Field | Description |
| --- | --- |
| `type` _[RateLimitDatabaseBackendType](#ratelimitdatabasebackendtype)_ | Type is the type of database backend to use. Supported types are: * Redis: Connects to a Redis database. |
| `redis` _[RateLimitRedisSettings](#ratelimitredissettings)_ | Redis defines the settings needed to connect to a Redis database. |


#### RateLimitDatabaseBackendType

_Underlying type:_ `string`

RateLimitDatabaseBackendType specifies the types of database backend to be used by the rate limit service.

_Appears in:_
- [RateLimitDatabaseBackend](#ratelimitdatabasebackend)



#### RateLimitRedisSettings



RateLimitRedisSettings defines the configuration for connecting to redis database.

_Appears in:_
- [RateLimitDatabaseBackend](#ratelimitdatabasebackend)

| Field | Description |
| --- | --- |
| `url` _string_ | URL of the Redis Database. |
| `tls` _[RedisTLSSettings](#redistlssettings)_ | TLS defines TLS configuration for connecting to redis database. |


#### RateLimitRule



RateLimitRule defines the semantics for matching attributes from the incoming requests, and setting limits for them.

_Appears in:_
- [GlobalRateLimit](#globalratelimit)

| Field | Description |
| --- | --- |
| `clientSelectors` _[RateLimitSelectCondition](#ratelimitselectcondition) array_ | ClientSelectors holds the list of select conditions to select specific clients using attributes from the traffic flow. All individual select conditions must hold True for this rule and its limit to be applied. If this field is empty, it is equivalent to True, and the limit is applied. |
| `limit` _[RateLimitValue](#ratelimitvalue)_ | Limit holds the rate limit values. This limit is applied for traffic flows when the selectors compute to True, causing the request to be counted towards the limit. The limit is enforced and the request is ratelimited, i.e. a response with 429 HTTP status code is sent back to the client when the selected requests have reached the limit. |


#### RateLimitSelectCondition



RateLimitSelectCondition specifies the attributes within the traffic flow that can be used to select a subset of clients to be ratelimited. All the individual conditions must hold True for the overall condition to hold True.

_Appears in:_
- [RateLimitRule](#ratelimitrule)

| Field | Description |
| --- | --- |
| `headers` _[HeaderMatch](#headermatch) array_ | Headers is a list of request headers to match. Multiple header values are ANDed together, meaning, a request MUST match all the specified headers. |
| `sourceCIDR` _[SourceMatch](#sourcematch)_ | SourceCIDR is the client IP Address range to match on. |


#### RateLimitSpec



RateLimitSpec defines the desired state of RateLimitSpec.

_Appears in:_
- [BackendTrafficPolicySpec](#backendtrafficpolicyspec)

| Field | Description |
| --- | --- |
| `type` _[RateLimitType](#ratelimittype)_ | Type decides the scope for the RateLimits. Valid RateLimitType values are "Global". |
| `global` _[GlobalRateLimit](#globalratelimit)_ | Global defines global rate limit configuration. |


#### RateLimitType

_Underlying type:_ `string`

RateLimitType specifies the types of RateLimiting.

_Appears in:_
- [RateLimitSpec](#ratelimitspec)



#### RateLimitUnit

_Underlying type:_ `string`

RateLimitUnit specifies the intervals for setting rate limits. Valid RateLimitUnit values are "Second", "Minute", "Hour", and "Day".

_Appears in:_
- [RateLimitValue](#ratelimitvalue)



#### RateLimitValue



RateLimitValue defines the limits for rate limiting.

_Appears in:_
- [RateLimitRule](#ratelimitrule)

| Field | Description |
| --- | --- |
| `requests` _integer_ |  |
| `unit` _[RateLimitUnit](#ratelimitunit)_ |  |


#### RedisTLSSettings



RedisTLSSettings defines the TLS configuration for connecting to redis database.

_Appears in:_
- [RateLimitRedisSettings](#ratelimitredissettings)

| Field | Description |
| --- | --- |
| `certificateRef` _[SecretObjectReference](https://gateway-api.sigs.k8s.io/references/spec/#gateway.networking.k8s.io/v1.SecretObjectReference)_ | CertificateRef defines the client certificate reference for TLS connections. Currently only a Kubernetes Secret of type TLS is supported. |


#### RemoteJWKS



RemoteJWKS defines how to fetch and cache JSON Web Key Sets (JWKS) from a remote HTTP/HTTPS endpoint.

_Appears in:_
- [JWTProvider](#jwtprovider)

| Field | Description |
| --- | --- |
| `uri` _string_ | URI is the HTTPS URI to fetch the JWKS. Envoy's system trust bundle is used to validate the server certificate. |


#### RequestHeaderCustomTag



RequestHeaderCustomTag adds value from request header to each span.

_Appears in:_
- [CustomTag](#customtag)

| Field | Description |
| --- | --- |
| `name` _string_ | Name defines the name of the request header which to extract the value from. |
| `defaultValue` _string_ | DefaultValue defines the default value to use if the request header is not set. |


#### ResourceProviderType

_Underlying type:_ `string`

ResourceProviderType defines the types of custom resource providers supported by Envoy Gateway.

_Appears in:_
- [EnvoyGatewayResourceProvider](#envoygatewayresourceprovider)



#### SecurityPolicy



SecurityPolicy allows the user to configure various security settings for a Gateway.

_Appears in:_
- [SecurityPolicyList](#securitypolicylist)

| Field | Description |
| --- | --- |
| `apiVersion` _string_ | `gateway.envoyproxy.io/v1alpha1`
| `kind` _string_ | `SecurityPolicy`
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |
| `spec` _[SecurityPolicySpec](#securitypolicyspec)_ | Spec defines the desired state of SecurityPolicy. |


#### SecurityPolicyList



SecurityPolicyList contains a list of SecurityPolicy resources.



| Field | Description |
| --- | --- |
| `apiVersion` _string_ | `gateway.envoyproxy.io/v1alpha1`
| `kind` _string_ | `SecurityPolicyList`
| `metadata` _[ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#listmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |
| `items` _[SecurityPolicy](#securitypolicy) array_ |  |


#### SecurityPolicySpec



SecurityPolicySpec defines the desired state of SecurityPolicy.

_Appears in:_
- [SecurityPolicy](#securitypolicy)

| Field | Description |
| --- | --- |
| `targetRef` _[PolicyTargetReferenceWithSectionName](#policytargetreferencewithsectionname)_ | TargetRef is the name of the Gateway resource this policy is being attached to. This Policy and the TargetRef MUST be in the same namespace for this Policy to have effect and be applied to the Gateway. TargetRef |
| `cors` _[CORS](#cors)_ | CORS defines the configuration for Cross-Origin Resource Sharing (CORS). |
| `jwt` _[JWT](#jwt)_ | JWT defines the configuration for JSON Web Token (JWT) authentication. |




#### ServiceType

_Underlying type:_ `string`

ServiceType string describes ingress methods for a service

_Appears in:_
- [KubernetesServiceSpec](#kubernetesservicespec)



#### SourceMatch





_Appears in:_
- [RateLimitSelectCondition](#ratelimitselectcondition)



#### StringMatch



StringMatch defines how to match any strings. This is a general purpose match condition that can be used by other EG APIs that need to match against a string.

_Appears in:_
- [CORS](#cors)

| Field | Description |
| --- | --- |
| `type` _[MatchType](#matchtype)_ | Type specifies how to match against a string. |
| `value` _string_ | Value specifies the string value that the match must have. |


#### TCPKeepalive



TCPKeepalive define the TCP Keepalive configuration.

_Appears in:_
- [ClientTrafficPolicySpec](#clienttrafficpolicyspec)

| Field | Description |
| --- | --- |
| `probes` _integer_ | The total number of unacknowledged probes to send before deciding the connection is dead. Defaults to 9. |
| `idleTime` _Duration_ | The duration a connection needs to be idle before keep-alive probes start being sent. The duration format is Defaults to `7200s`. |
| `interval` _Duration_ | The duration between keep-alive probes. Defaults to `75s`. |


#### TracingProvider





_Appears in:_
- [ProxyTracing](#proxytracing)

| Field | Description |
| --- | --- |
| `type` _[TracingProviderType](#tracingprovidertype)_ | Type defines the tracing provider type. EG currently only supports OpenTelemetry. |
| `host` _string_ | Host define the provider service hostname. |
| `port` _integer_ | Port defines the port the provider service is exposed on. |


#### TracingProviderType

_Underlying type:_ `string`



_Appears in:_
- [TracingProvider](#tracingprovider)



#### XDSTranslatorHook

_Underlying type:_ `string`

XDSTranslatorHook defines the types of hooks that an Envoy Gateway extension may support for the xds-translator

_Appears in:_
- [XDSTranslatorHooks](#xdstranslatorhooks)



#### XDSTranslatorHooks



XDSTranslatorHooks contains all the pre and post hooks for the xds-translator runner.

_Appears in:_
- [ExtensionHooks](#extensionhooks)

| Field | Description |
| --- | --- |
| `pre` _[XDSTranslatorHook](#xdstranslatorhook) array_ |  |
| `post` _[XDSTranslatorHook](#xdstranslatorhook) array_ |  |


