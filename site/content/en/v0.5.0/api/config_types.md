---
title: "Config APIs"
---

## Packages
- [config.gateway.envoyproxy.io/v1alpha1](#configgatewayenvoyproxyiov1alpha1)


## config.gateway.envoyproxy.io/v1alpha1

Package v1alpha1 contains API schema definitions for the config.gateway.envoyproxy.io
API group.


### Resource Types
- [EnvoyGateway](#envoygateway)
- [EnvoyProxy](#envoyproxy)



## CustomTag





_Appears in:_
- [ProxyTracing](#proxytracing)

| Field | Description |
| --- | --- |
| `type` _[CustomTagType](#customtagtype)_ | Type defines the type of custom tag. |
| `literal` _[LiteralCustomTag](#literalcustomtag)_ | Literal adds hard-coded value to each span. It's required when the type is "Literal". |
| `environment` _[EnvironmentCustomTag](#environmentcustomtag)_ | Environment adds value from environment variable to each span. It's required when the type is "Environment". |
| `requestHeader` _[RequestHeaderCustomTag](#requestheadercustomtag)_ | RequestHeader adds value from request header to each span. It's required when the type is "RequestHeader". |


## CustomTagType

_Underlying type:_ `string`



_Appears in:_
- [CustomTag](#customtag)



## EnvironmentCustomTag



EnvironmentCustomTag adds value from environment variable to each span.

_Appears in:_
- [CustomTag](#customtag)

| Field | Description |
| --- | --- |
| `name` _string_ | Name defines the name of the environment variable which to extract the value from. |
| `defaultValue` _string_ | DefaultValue defines the default value to use if the environment variable is not set. |


## EnvoyGateway



EnvoyGateway is the schema for the envoygateways API.



| Field | Description |
| --- | --- |
| `apiVersion` _string_ | `config.gateway.envoyproxy.io/v1alpha1`
| `kind` _string_ | `EnvoyGateway`
| `gateway` _[Gateway](#gateway)_ | Gateway defines desired Gateway API specific configuration. If unset, default configuration parameters will apply. |
| `provider` _[EnvoyGatewayProvider](#envoygatewayprovider)_ | Provider defines the desired provider and provider-specific configuration. If unspecified, the Kubernetes provider is used with default configuration parameters. |
| `logging` _[EnvoyGatewayLogging](#envoygatewaylogging)_ | Logging defines logging parameters for Envoy Gateway. |
| `admin` _[EnvoyGatewayAdmin](#envoygatewayadmin)_ | Admin defines the desired admin related abilities. If unspecified, the Admin is used with default configuration parameters. |
| `rateLimit` _[RateLimit](#ratelimit)_ | RateLimit defines the configuration associated with the Rate Limit service deployed by Envoy Gateway required to implement the Global Rate limiting functionality. The specific rate limit service used here is the reference implementation in Envoy. For more details visit https://github.com/envoyproxy/ratelimit. This configuration is unneeded for "Local" rate limiting. |
| `extensionManager` _[ExtensionManager](#extensionmanager)_ | ExtensionManager defines an extension manager to register for the Envoy Gateway Control Plane. |
| `extensionApis` _[ExtensionAPISettings](#extensionapisettings)_ | ExtensionAPIs defines the settings related to specific Gateway API Extensions implemented by Envoy Gateway |


## EnvoyGatewayAdmin



EnvoyGatewayAdmin defines the Envoy Gateway Admin configuration.

_Appears in:_
- [EnvoyGateway](#envoygateway)
- [EnvoyGatewaySpec](#envoygatewayspec)

| Field | Description |
| --- | --- |
| `address` _[EnvoyGatewayAdminAddress](#envoygatewayadminaddress)_ | Address defines the address of Envoy Gateway Admin Server. |
| `debug` _boolean_ | Debug defines if enable the /debug endpoint of Envoy Gateway. |


## EnvoyGatewayAdminAddress



EnvoyGatewayAdminAddress defines the Envoy Gateway Admin Address configuration.

_Appears in:_
- [EnvoyGatewayAdmin](#envoygatewayadmin)

| Field | Description |
| --- | --- |
| `port` _integer_ | Port defines the port the admin server is exposed on. |
| `host` _string_ | Host defines the admin server hostname. |


## EnvoyGatewayCustomProvider



EnvoyGatewayCustomProvider defines configuration for the Custom provider.

_Appears in:_
- [EnvoyGatewayProvider](#envoygatewayprovider)

| Field | Description |
| --- | --- |
| `resource` _[EnvoyGatewayResourceProvider](#envoygatewayresourceprovider)_ | Resource defines the desired resource provider. This provider is used to specify the provider to be used to retrieve the resource configurations such as Gateway API resources |
| `infrastructure` _[EnvoyGatewayInfrastructureProvider](#envoygatewayinfrastructureprovider)_ | Infrastructure defines the desired infrastructure provider. This provider is used to specify the provider to be used to provide an environment to deploy the out resources like the Envoy Proxy data plane. |


## EnvoyGatewayFileResourceProvider



EnvoyGatewayFileResourceProvider defines configuration for the File Resource provider.

_Appears in:_
- [EnvoyGatewayResourceProvider](#envoygatewayresourceprovider)

| Field | Description |
| --- | --- |
| `paths` _string array_ | Paths are the paths to a directory or file containing the resource configuration. Recursive sub directories are not currently supported. |


## EnvoyGatewayHostInfrastructureProvider



EnvoyGatewayHostInfrastructureProvider defines configuration for the Host Infrastructure provider.

_Appears in:_
- [EnvoyGatewayInfrastructureProvider](#envoygatewayinfrastructureprovider)



## EnvoyGatewayInfrastructureProvider



EnvoyGatewayInfrastructureProvider defines configuration for the Custom Infrastructure provider.

_Appears in:_
- [EnvoyGatewayCustomProvider](#envoygatewaycustomprovider)

| Field | Description |
| --- | --- |
| `type` _[InfrastructureProviderType](#infrastructureprovidertype)_ | Type is the type of infrastructure providers to use. Supported types are "Host". |
| `host` _[EnvoyGatewayHostInfrastructureProvider](#envoygatewayhostinfrastructureprovider)_ | Host defines the configuration of the Host provider. Host provides runtime deployment of the data plane as a child process on the host environment. |


## EnvoyGatewayKubernetesProvider



EnvoyGatewayKubernetesProvider defines configuration for the Kubernetes provider.

_Appears in:_
- [EnvoyGatewayProvider](#envoygatewayprovider)

| Field | Description |
| --- | --- |
| `rateLimitDeployment` _[KubernetesDeploymentSpec](#kubernetesdeploymentspec)_ | RateLimitDeployment defines the desired state of the Envoy ratelimit deployment resource. If unspecified, default settings for the managed Envoy ratelimit deployment resource are applied. |
| `watch` _[KubernetesWatchMode](#kuberneteswatchmode)_ | Watch holds configuration of which input resources should be watched and reconciled. |
| `deploy` _[KubernetesDeployMode](#kubernetesdeploymode)_ | Deploy holds configuration of how output managed resources such as the Envoy Proxy data plane should be deployed |
| `overwrite_control_plane_certs` _boolean_ | OverwriteControlPlaneCerts updates the secrets containing the control plane certs, when set. |


## EnvoyGatewayLogComponent

_Underlying type:_ `string`

EnvoyGatewayLogComponent defines a component that supports a configured logging level.

_Appears in:_
- [EnvoyGatewayLogging](#envoygatewaylogging)



## EnvoyGatewayLogging



EnvoyGatewayLogging defines logging for Envoy Gateway.

_Appears in:_
- [EnvoyGateway](#envoygateway)
- [EnvoyGatewaySpec](#envoygatewayspec)

| Field | Description |
| --- | --- |
| `level` _object (keys:[EnvoyGatewayLogComponent](#envoygatewaylogcomponent), values:[LogLevel](#loglevel))_ | Level is the logging level. If unspecified, defaults to "info". EnvoyGatewayLogComponent options: default/provider/gateway-api/xds-translator/xds-server/infrastructure/global-ratelimit. LogLevel options: debug/info/error/warn. |


## EnvoyGatewayProvider



EnvoyGatewayProvider defines the desired configuration of a provider.

_Appears in:_
- [EnvoyGateway](#envoygateway)
- [EnvoyGatewaySpec](#envoygatewayspec)

| Field | Description |
| --- | --- |
| `type` _[ProviderType](#providertype)_ | Type is the type of provider to use. Supported types are "Kubernetes". |
| `kubernetes` _[EnvoyGatewayKubernetesProvider](#envoygatewaykubernetesprovider)_ | Kubernetes defines the configuration of the Kubernetes provider. Kubernetes provides runtime configuration via the Kubernetes API. |
| `custom` _[EnvoyGatewayCustomProvider](#envoygatewaycustomprovider)_ | Custom defines the configuration for the Custom provider. This provider allows you to define a specific resource provider and a infrastructure provider. |


## EnvoyGatewayResourceProvider



EnvoyGatewayResourceProvider defines configuration for the Custom Resource provider.

_Appears in:_
- [EnvoyGatewayCustomProvider](#envoygatewaycustomprovider)

| Field | Description |
| --- | --- |
| `type` _[ResourceProviderType](#resourceprovidertype)_ | Type is the type of resource provider to use. Supported types are "File". |
| `file` _[EnvoyGatewayFileResourceProvider](#envoygatewayfileresourceprovider)_ | File defines the configuration of the File provider. File provides runtime configuration defined by one or more files. |


## EnvoyGatewaySpec



EnvoyGatewaySpec defines the desired state of Envoy Gateway.

_Appears in:_
- [EnvoyGateway](#envoygateway)

| Field | Description |
| --- | --- |
| `gateway` _[Gateway](#gateway)_ | Gateway defines desired Gateway API specific configuration. If unset, default configuration parameters will apply. |
| `provider` _[EnvoyGatewayProvider](#envoygatewayprovider)_ | Provider defines the desired provider and provider-specific configuration. If unspecified, the Kubernetes provider is used with default configuration parameters. |
| `logging` _[EnvoyGatewayLogging](#envoygatewaylogging)_ | Logging defines logging parameters for Envoy Gateway. |
| `admin` _[EnvoyGatewayAdmin](#envoygatewayadmin)_ | Admin defines the desired admin related abilities. If unspecified, the Admin is used with default configuration parameters. |
| `rateLimit` _[RateLimit](#ratelimit)_ | RateLimit defines the configuration associated with the Rate Limit service deployed by Envoy Gateway required to implement the Global Rate limiting functionality. The specific rate limit service used here is the reference implementation in Envoy. For more details visit https://github.com/envoyproxy/ratelimit. This configuration is unneeded for "Local" rate limiting. |
| `extensionManager` _[ExtensionManager](#extensionmanager)_ | ExtensionManager defines an extension manager to register for the Envoy Gateway Control Plane. |
| `extensionApis` _[ExtensionAPISettings](#extensionapisettings)_ | ExtensionAPIs defines the settings related to specific Gateway API Extensions implemented by Envoy Gateway |


## EnvoyProxy



EnvoyProxy is the schema for the envoyproxies API.



| Field | Description |
| --- | --- |
| `apiVersion` _string_ | `config.gateway.envoyproxy.io/v1alpha1`
| `kind` _string_ | `EnvoyProxy`
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |
| `spec` _[EnvoyProxySpec](#envoyproxyspec)_ | EnvoyProxySpec defines the desired state of EnvoyProxy. |


## EnvoyProxyKubernetesProvider



EnvoyProxyKubernetesProvider defines configuration for the Kubernetes resource provider.

_Appears in:_
- [EnvoyProxyProvider](#envoyproxyprovider)

| Field | Description |
| --- | --- |
| `envoyDeployment` _[KubernetesDeploymentSpec](#kubernetesdeploymentspec)_ | EnvoyDeployment defines the desired state of the Envoy deployment resource. If unspecified, default settings for the managed Envoy deployment resource are applied. |
| `envoyService` _[KubernetesServiceSpec](#kubernetesservicespec)_ | EnvoyService defines the desired state of the Envoy service resource. If unspecified, default settings for the managed Envoy service resource are applied. |


## EnvoyProxyProvider



EnvoyProxyProvider defines the desired state of a resource provider.

_Appears in:_
- [EnvoyProxySpec](#envoyproxyspec)

| Field | Description |
| --- | --- |
| `type` _[ProviderType](#providertype)_ | Type is the type of resource provider to use. A resource provider provides infrastructure resources for running the data plane, e.g. Envoy proxy, and optional auxiliary control planes. Supported types are "Kubernetes". |
| `kubernetes` _[EnvoyProxyKubernetesProvider](#envoyproxykubernetesprovider)_ | Kubernetes defines the desired state of the Kubernetes resource provider. Kubernetes provides infrastructure resources for running the data plane, e.g. Envoy proxy. If unspecified and type is "Kubernetes", default settings for managed Kubernetes resources are applied. |


## EnvoyProxySpec



EnvoyProxySpec defines the desired state of EnvoyProxy.

_Appears in:_
- [EnvoyProxy](#envoyproxy)

| Field | Description |
| --- | --- |
| `provider` _[EnvoyProxyProvider](#envoyproxyprovider)_ | Provider defines the desired resource provider and provider-specific configuration. If unspecified, the "Kubernetes" resource provider is used with default configuration parameters. |
| `logging` _[ProxyLogging](#proxylogging)_ | Logging defines logging parameters for managed proxies. |
| `telemetry` _[ProxyTelemetry](#proxytelemetry)_ | Telemetry defines telemetry parameters for managed proxies. |
| `bootstrap` _string_ | Bootstrap defines the Envoy Bootstrap as a YAML string. Visit https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/bootstrap/v3/bootstrap.proto#envoy-v3-api-msg-config-bootstrap-v3-bootstrap to learn more about the syntax. If set, this is the Bootstrap configuration used for the managed Envoy Proxy fleet instead of the default Bootstrap configuration set by Envoy Gateway. Some fields within the Bootstrap that are required to communicate with the xDS Server (Envoy Gateway) and receive xDS resources from it are not configurable and will result in the `EnvoyProxy` resource being rejected. Backward compatibility across minor versions is not guaranteed. We strongly recommend using `egctl x translate` to generate a `EnvoyProxy` resource with the `Bootstrap` field set to the default Bootstrap configuration used. You can edit this configuration, and rerun `egctl x translate` to ensure there are no validation errors. |




## ExtensionAPISettings



ExtensionAPISettings defines the settings specific to Gateway API Extensions.

_Appears in:_
- [EnvoyGateway](#envoygateway)
- [EnvoyGatewaySpec](#envoygatewayspec)

| Field | Description |
| --- | --- |
| `enableEnvoyPatchPolicy` _boolean_ | EnableEnvoyPatchPolicy enables Envoy Gateway to reconcile and implement the EnvoyPatchPolicy resources. |


## ExtensionHooks



ExtensionHooks defines extension hooks across all supported runners

_Appears in:_
- [ExtensionManager](#extensionmanager)

| Field | Description |
| --- | --- |
| `xdsTranslator` _[XDSTranslatorHooks](#xdstranslatorhooks)_ | XDSTranslator defines all the supported extension hooks for the xds-translator runner |


## ExtensionManager



ExtensionManager defines the configuration for registering an extension manager to the Envoy Gateway control plane.

_Appears in:_
- [EnvoyGateway](#envoygateway)
- [EnvoyGatewaySpec](#envoygatewayspec)

| Field | Description |
| --- | --- |
| `resources` _[GroupVersionKind](#groupversionkind) array_ | Resources defines the set of K8s resources the extension will handle. |
| `hooks` _[ExtensionHooks](#extensionhooks)_ | Hooks defines the set of hooks the extension supports |
| `service` _[ExtensionService](#extensionservice)_ | Service defines the configuration of the extension service that the Envoy Gateway Control Plane will call through extension hooks. |


## ExtensionService



ExtensionService defines the configuration for connecting to a registered extension service.

_Appears in:_
- [ExtensionManager](#extensionmanager)

| Field | Description |
| --- | --- |
| `host` _string_ | Host define the extension service hostname. |
| `port` _integer_ | Port defines the port the extension service is exposed on. |
| `tls` _[ExtensionTLS](#extensiontls)_ | TLS defines TLS configuration for communication between Envoy Gateway and the extension service. |


## ExtensionTLS



ExtensionTLS defines the TLS configuration when connecting to an extension service

_Appears in:_
- [ExtensionService](#extensionservice)

| Field | Description |
| --- | --- |
| `certificateRef` _[SecretObjectReference](https://gateway-api.sigs.k8s.io/references/spec/#gateway.networking.k8s.io/v1beta1.SecretObjectReference)_ | CertificateRef contains a references to objects (Kubernetes objects or otherwise) that contains a TLS certificate and private keys. These certificates are used to establish a TLS handshake to the extension server. 
 CertificateRef can only reference a Kubernetes Secret at this time. |


## FileEnvoyProxyAccessLog





_Appears in:_
- [ProxyAccessLogSink](#proxyaccesslogsink)

| Field | Description |
| --- | --- |
| `path` _string_ | Path defines the file path used to expose envoy access log(e.g. /dev/stdout). Empty value disables accesslog. |


## Gateway



Gateway defines the desired Gateway API configuration of Envoy Gateway.

_Appears in:_
- [EnvoyGateway](#envoygateway)
- [EnvoyGatewaySpec](#envoygatewayspec)

| Field | Description |
| --- | --- |
| `controllerName` _string_ | ControllerName defines the name of the Gateway API controller. If unspecified, defaults to "gateway.envoyproxy.io/gatewayclass-controller". See the following for additional details: https://gateway-api.sigs.k8s.io/v1alpha2/references/spec/#gateway.networking.k8s.io/v1alpha2.GatewayClass |


## GroupVersionKind



GroupVersionKind unambiguously identifies a Kind. It can be converted to k8s.io/apimachinery/pkg/runtime/schema.GroupVersionKind

_Appears in:_
- [ExtensionManager](#extensionmanager)

| Field | Description |
| --- | --- |
| `group` _string_ |  |
| `version` _string_ |  |
| `kind` _string_ |  |


## InfrastructureProviderType

_Underlying type:_ `string`

InfrastructureProviderType defines the types of custom infrastructure providers supported by Envoy Gateway.

_Appears in:_
- [EnvoyGatewayInfrastructureProvider](#envoygatewayinfrastructureprovider)



## KubernetesContainerSpec



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


## KubernetesDeployMode



KubernetesDeployMode holds configuration for how to deploy managed resources such as the Envoy Proxy data plane fleet.

_Appears in:_
- [EnvoyGatewayKubernetesProvider](#envoygatewaykubernetesprovider)



## KubernetesDeploymentSpec



KubernetesDeploymentSpec defines the desired state of the Kubernetes deployment resource.

_Appears in:_
- [EnvoyGatewayKubernetesProvider](#envoygatewaykubernetesprovider)
- [EnvoyProxyKubernetesProvider](#envoyproxykubernetesprovider)

| Field | Description |
| --- | --- |
| `replicas` _integer_ | Replicas is the number of desired pods. Defaults to 1. |
| `strategy` _[DeploymentStrategy](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#deploymentstrategy-v1-apps)_ | The deployment strategy to use to replace existing pods with new ones. |
| `pod` _[KubernetesPodSpec](#kubernetespodspec)_ | Pod defines the desired annotations and securityContext of container. |
| `container` _[KubernetesContainerSpec](#kubernetescontainerspec)_ | Container defines the resources and securityContext of container. |


## KubernetesPodSpec



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


## KubernetesServiceSpec



KubernetesServiceSpec defines the desired state of the Kubernetes service resource.

_Appears in:_
- [EnvoyProxyKubernetesProvider](#envoyproxykubernetesprovider)

| Field | Description |
| --- | --- |
| `annotations` _object (keys:string, values:string)_ | Annotations that should be appended to the service. By default, no annotations are appended. |
| `type` _[ServiceType](#servicetype)_ | Type determines how the Service is exposed. Defaults to LoadBalancer. Valid options are ClusterIP, LoadBalancer and NodePort. "LoadBalancer" means a service will be exposed via an external load balancer (if the cloud provider supports it). "ClusterIP" means a service will only be accessible inside the cluster, via the cluster IP. "NodePort" means a service will be exposed on a static Port on all Nodes of the cluster. |


## KubernetesWatchMode



KubernetesWatchMode holds the configuration for which input resources to watch and reconcile.

_Appears in:_
- [EnvoyGatewayKubernetesProvider](#envoygatewaykubernetesprovider)

| Field | Description |
| --- | --- |
| `Namespaces` _string array_ | Namespaces holds the list of namespaces that Envoy Gateway will watch for namespaced scoped resources such as Gateway, HTTPRoute and Service. Note that Envoy Gateway will continue to reconcile relevant cluster scoped resources such as GatewayClass that it is linked to. By default, when this field is unset or empty, Envoy Gateway will watch for input namespaced resources from all namespaces. |


## LiteralCustomTag



LiteralCustomTag adds hard-coded value to each span.

_Appears in:_
- [CustomTag](#customtag)

| Field | Description |
| --- | --- |
| `value` _string_ | Value defines the hard-coded value to add to each span. |


## LogComponent

_Underlying type:_ `string`

LogComponent defines a component that supports a configured logging level.

_Appears in:_
- [ProxyLogging](#proxylogging)



## LogLevel

_Underlying type:_ `string`

LogLevel defines a log level for Envoy Gateway and EnvoyProxy system logs. This type is not implemented for EnvoyProxy until https://github.com/envoyproxy/gateway/issues/280 is fixed.

_Appears in:_
- [EnvoyGatewayLogging](#envoygatewaylogging)
- [ProxyLogging](#proxylogging)



## MetricSink





_Appears in:_
- [ProxyMetrics](#proxymetrics)

| Field | Description |
| --- | --- |
| `type` _[MetricSinkType](#metricsinktype)_ | Type defines the metric sink type. EG currently only supports OpenTelemetry. |
| `openTelemetry` _[OpenTelemetrySink](#opentelemetrysink)_ | OpenTelemetry defines the configuration for OpenTelemetry sink. It's required if the sink type is OpenTelemetry. |


## MetricSinkType

_Underlying type:_ `string`



_Appears in:_
- [MetricSink](#metricsink)



## OpenTelemetryEnvoyProxyAccessLog



TODO: consider reuse ExtensionService?

_Appears in:_
- [ProxyAccessLogSink](#proxyaccesslogsink)

| Field | Description |
| --- | --- |
| `host` _string_ | Host define the extension service hostname. |
| `port` _integer_ | Port defines the port the extension service is exposed on. |
| `resources` _object (keys:string, values:string)_ | Resources is a set of labels that describe the source of a log entry, including envoy node info. It's recommended to follow [semantic conventions](https://opentelemetry.io/docs/reference/specification/resource/semantic_conventions/). |


## OpenTelemetrySink





_Appears in:_
- [MetricSink](#metricsink)

| Field | Description |
| --- | --- |
| `host` _string_ | Host define the service hostname. |
| `port` _integer_ | Port defines the port the service is exposed on. |


## PrometheusProvider





_Appears in:_
- [ProxyMetrics](#proxymetrics)



## ProviderType

_Underlying type:_ `string`

ProviderType defines the types of providers supported by Envoy Gateway.

_Appears in:_
- [EnvoyGatewayProvider](#envoygatewayprovider)
- [EnvoyProxyProvider](#envoyproxyprovider)



## ProxyAccessLog





_Appears in:_
- [ProxyTelemetry](#proxytelemetry)

| Field | Description |
| --- | --- |
| `disable` _boolean_ | Disable disables access logging for managed proxies if set to true. |
| `settings` _[ProxyAccessLogSetting](#proxyaccesslogsetting) array_ | Settings defines accesslog settings for managed proxies. If unspecified, will send default format to stdout. |


## ProxyAccessLogFormat



ProxyAccessLogFormat defines the format of accesslog.

_Appears in:_
- [ProxyAccessLogSetting](#proxyaccesslogsetting)

| Field | Description |
| --- | --- |
| `type` _[ProxyAccessLogFormatType](#proxyaccesslogformattype)_ | Type defines the type of accesslog format. |
| `text` _string_ | Text defines the text accesslog format, following Envoy accesslog formatting, empty value results in proxy's default access log format. It's required when the format type is "Text". Envoy [command operators](https://www.envoyproxy.io/docs/envoy/latest/configuration/observability/access_log/usage#command-operators) may be used in the format. The [format string documentation](https://www.envoyproxy.io/docs/envoy/latest/configuration/observability/access_log/usage#config-access-log-format-strings) provides more information. |
| `json` _object (keys:string, values:string)_ | JSON is additional attributes that describe the specific event occurrence. Structured format for the envoy access logs. Envoy [command operators](https://www.envoyproxy.io/docs/envoy/latest/configuration/observability/access_log/usage#command-operators) can be used as values for fields within the Struct. It's required when the format type is "JSON". |


## ProxyAccessLogFormatType

_Underlying type:_ `string`



_Appears in:_
- [ProxyAccessLogFormat](#proxyaccesslogformat)



## ProxyAccessLogSetting





_Appears in:_
- [ProxyAccessLog](#proxyaccesslog)

| Field | Description |
| --- | --- |
| `format` _[ProxyAccessLogFormat](#proxyaccesslogformat)_ | Format defines the format of accesslog. |
| `sinks` _[ProxyAccessLogSink](#proxyaccesslogsink) array_ | Sinks defines the sinks of accesslog. |


## ProxyAccessLogSink





_Appears in:_
- [ProxyAccessLogSetting](#proxyaccesslogsetting)

| Field | Description |
| --- | --- |
| `type` _[ProxyAccessLogSinkType](#proxyaccesslogsinktype)_ | Type defines the type of accesslog sink. |
| `file` _[FileEnvoyProxyAccessLog](#fileenvoyproxyaccesslog)_ | File defines the file accesslog sink. |
| `openTelemetry` _[OpenTelemetryEnvoyProxyAccessLog](#opentelemetryenvoyproxyaccesslog)_ | OpenTelemetry defines the OpenTelemetry accesslog sink. |


## ProxyAccessLogSinkType

_Underlying type:_ `string`



_Appears in:_
- [ProxyAccessLogSink](#proxyaccesslogsink)



## ProxyLogging



ProxyLogging defines logging parameters for managed proxies.

_Appears in:_
- [EnvoyProxySpec](#envoyproxyspec)

| Field | Description |
| --- | --- |
| `level` _object (keys:[LogComponent](#logcomponent), values:[LogLevel](#loglevel))_ | Level is a map of logging level per component, where the component is the key and the log level is the value. If unspecified, defaults to "default: warn". |


## ProxyMetrics





_Appears in:_
- [ProxyTelemetry](#proxytelemetry)

| Field | Description |
| --- | --- |
| `prometheus` _[PrometheusProvider](#prometheusprovider)_ | Prometheus defines the configuration for Admin endpoint `/stats/prometheus`. |
| `sinks` _[MetricSink](#metricsink) array_ | Sinks defines the metric sinks where metrics are sent to. |


## ProxyTelemetry





_Appears in:_
- [EnvoyProxySpec](#envoyproxyspec)

| Field | Description |
| --- | --- |
| `accessLog` _[ProxyAccessLog](#proxyaccesslog)_ | AccessLogs defines accesslog parameters for managed proxies. If unspecified, will send default format to stdout. |
| `tracing` _[ProxyTracing](#proxytracing)_ | Tracing defines tracing configuration for managed proxies. If unspecified, will not send tracing data. |
| `metrics` _[ProxyMetrics](#proxymetrics)_ | Metrics defines metrics configuration for managed proxies. |


## ProxyTracing





_Appears in:_
- [ProxyTelemetry](#proxytelemetry)

| Field | Description |
| --- | --- |
| `samplingRate` _integer_ | SamplingRate controls the rate at which traffic will be selected for tracing if no prior sampling decision has been made. Defaults to 100, valid values [0-100]. 100 indicates 100% sampling. |
| `customTags` _object (keys:string, values:[CustomTag](#customtag))_ | CustomTags defines the custom tags to add to each span. If provider is kubernetes, pod name and namespace are added by default. |
| `provider` _[TracingProvider](#tracingprovider)_ | Provider defines the tracing provider. Only OpenTelemetry is supported currently. |


## RateLimit



RateLimit defines the configuration associated with the Rate Limit Service used for Global Rate Limiting.

_Appears in:_
- [EnvoyGateway](#envoygateway)
- [EnvoyGatewaySpec](#envoygatewayspec)

| Field | Description |
| --- | --- |
| `backend` _[RateLimitDatabaseBackend](#ratelimitdatabasebackend)_ | Backend holds the configuration associated with the database backend used by the rate limit service to store state associated with global ratelimiting. |


## RateLimitDatabaseBackend



RateLimitDatabaseBackend defines the configuration associated with the database backend used by the rate limit service.

_Appears in:_
- [RateLimit](#ratelimit)

| Field | Description |
| --- | --- |
| `type` _[RateLimitDatabaseBackendType](#ratelimitdatabasebackendtype)_ | Type is the type of database backend to use. Supported types are: * Redis: Connects to a Redis database. |
| `redis` _[RateLimitRedisSettings](#ratelimitredissettings)_ | Redis defines the settings needed to connect to a Redis database. |


## RateLimitDatabaseBackendType

_Underlying type:_ `string`

RateLimitDatabaseBackendType specifies the types of database backend to be used by the rate limit service.

_Appears in:_
- [RateLimitDatabaseBackend](#ratelimitdatabasebackend)



## RateLimitRedisSettings



RateLimitRedisSettings defines the configuration for connecting to redis database.

_Appears in:_
- [RateLimitDatabaseBackend](#ratelimitdatabasebackend)

| Field | Description |
| --- | --- |
| `url` _string_ | URL of the Redis Database. |
| `tls` _[RedisTLSSettings](#redistlssettings)_ | TLS defines TLS configuration for connecting to redis database. |


## RedisTLSSettings



RedisTLSSettings defines the TLS configuration for connecting to redis database.

_Appears in:_
- [RateLimitRedisSettings](#ratelimitredissettings)

| Field | Description |
| --- | --- |
| `certificateRef` _[SecretObjectReference](https://gateway-api.sigs.k8s.io/references/spec/#gateway.networking.k8s.io/v1beta1.SecretObjectReference)_ | CertificateRef defines the client certificate reference for TLS connections. Currently only a Kubernetes Secret of type TLS is supported. |


## RequestHeaderCustomTag



RequestHeaderCustomTag adds value from request header to each span.

_Appears in:_
- [CustomTag](#customtag)

| Field | Description |
| --- | --- |
| `name` _string_ | Name defines the name of the request header which to extract the value from. |
| `defaultValue` _string_ | DefaultValue defines the default value to use if the request header is not set. |


## ResourceProviderType

_Underlying type:_ `string`

ResourceProviderType defines the types of custom resource providers supported by Envoy Gateway.

_Appears in:_
- [EnvoyGatewayResourceProvider](#envoygatewayresourceprovider)



## ServiceType

_Underlying type:_ `string`

ServiceType string describes ingress methods for a service

_Appears in:_
- [KubernetesServiceSpec](#kubernetesservicespec)



## TracingProvider





_Appears in:_
- [ProxyTracing](#proxytracing)

| Field | Description |
| --- | --- |
| `type` _[TracingProviderType](#tracingprovidertype)_ | Type defines the tracing provider type. EG currently only supports OpenTelemetry. |
| `host` _string_ | Host define the provider service hostname. |
| `port` _integer_ | Port defines the port the provider service is exposed on. |


## TracingProviderType

_Underlying type:_ `string`



_Appears in:_
- [TracingProvider](#tracingprovider)



## XDSTranslatorHook

_Underlying type:_ `string`

XDSTranslatorHook defines the types of hooks that an Envoy Gateway extension may support for the xds-translator

_Appears in:_
- [XDSTranslatorHooks](#xdstranslatorhooks)



## XDSTranslatorHooks



XDSTranslatorHooks contains all the pre and post hooks for the xds-translator runner.

_Appears in:_
- [ExtensionHooks](#extensionhooks)

| Field | Description |
| --- | --- |
| `pre` _[XDSTranslatorHook](#xdstranslatorhook) array_ |  |
| `post` _[XDSTranslatorHook](#xdstranslatorhook) array_ |  |


