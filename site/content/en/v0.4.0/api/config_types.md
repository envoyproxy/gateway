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



## EnvoyGateway



EnvoyGateway is the schema for the envoygateways API.



| Field | Description |
| --- | --- |
| `apiVersion` _string_ | `config.gateway.envoyproxy.io/v1alpha1`
| `kind` _string_ | `EnvoyGateway`
| `EnvoyGatewaySpec` _[EnvoyGatewaySpec](#envoygatewayspec)_ | EnvoyGatewaySpec defines the desired state of EnvoyGateway. |


## EnvoyGatewayFileProvider



EnvoyGatewayFileProvider defines configuration for the File provider.

_Appears in:_
- [EnvoyGatewayProvider](#envoygatewayprovider)



## EnvoyGatewayKubernetesProvider



EnvoyGatewayKubernetesProvider defines configuration for the Kubernetes provider.

_Appears in:_
- [EnvoyGatewayProvider](#envoygatewayprovider)

| Field | Description |
| --- | --- |
| `rateLimitDeployment` _[KubernetesDeploymentSpec](#kubernetesdeploymentspec)_ | RateLimitDeployment defines the desired state of the Envoy ratelimit deployment resource. If unspecified, default settings for the managed Envoy ratelimit deployment resource are applied. |


## EnvoyGatewayProvider



EnvoyGatewayProvider defines the desired configuration of a provider.

_Appears in:_
- [EnvoyGatewaySpec](#envoygatewayspec)

| Field | Description |
| --- | --- |
| `type` _[ProviderType](#providertype)_ | Type is the type of provider to use. Supported types are "Kubernetes". |
| `kubernetes` _[EnvoyGatewayKubernetesProvider](#envoygatewaykubernetesprovider)_ | Kubernetes defines the configuration of the Kubernetes provider. Kubernetes provides runtime configuration via the Kubernetes API. |
| `file` _[EnvoyGatewayFileProvider](#envoygatewayfileprovider)_ | File defines the configuration of the File provider. File provides runtime configuration defined by one or more files. This type is not implemented until https://github.com/envoyproxy/gateway/issues/1001 is fixed. |


## EnvoyGatewaySpec



EnvoyGatewaySpec defines the desired state of Envoy Gateway.

_Appears in:_
- [EnvoyGateway](#envoygateway)

| Field | Description |
| --- | --- |
| `gateway` _[Gateway](#gateway)_ | Gateway defines desired Gateway API specific configuration. If unset, default configuration parameters will apply. |
| `provider` _[EnvoyGatewayProvider](#envoygatewayprovider)_ | Provider defines the desired provider and provider-specific configuration. If unspecified, the Kubernetes provider is used with default configuration parameters. |
| `rateLimit` _[RateLimit](#ratelimit)_ | RateLimit defines the configuration associated with the Rate Limit service deployed by Envoy Gateway required to implement the Global Rate limiting functionality. The specific rate limit service used here is the reference implementation in Envoy. For more details visit https://github.com/envoyproxy/ratelimit. This configuration is unneeded for "Local" rate limiting. |
| `extension` _[Extension](#extension)_ | Extension defines an extension to register for the Envoy Gateway Control Plane. |


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
| `logging` _[ProxyLogging](#proxylogging)_ | Logging defines logging parameters for managed proxies. If unspecified, default settings apply. This type is not implemented until https://github.com/envoyproxy/gateway/issues/280 is fixed. |
| `bootstrap` _string_ | Bootstrap defines the Envoy Bootstrap as a YAML string. Visit https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/bootstrap/v3/bootstrap.proto#envoy-v3-api-msg-config-bootstrap-v3-bootstrap to learn more about the syntax. If set, this is the Bootstrap configuration used for the managed Envoy Proxy fleet instead of the default Bootstrap configuration set by Envoy Gateway. Some fields within the Bootstrap that are required to communicate with the xDS Server (Envoy Gateway) and receive xDS resources from it are not configurable and will result in the `EnvoyProxy` resource being rejected. Backward compatibility across minor versions is not guaranteed. We strongly recommend using `egctl x translate` to generate a `EnvoyProxy` resource with the `Bootstrap` field set to the default Bootstrap configuration used. You can edit this configuration, and rerun `egctl x translate` to ensure there are no validation errors. |




## Extension



Extension defines the configuration for registering an extension to the Envoy Gateway control plane.

_Appears in:_
- [EnvoyGatewaySpec](#envoygatewayspec)

| Field | Description |
| --- | --- |
| `resources` _[GroupVersionKind](#groupversionkind) array_ | Resources defines the set of K8s resources the extension will handle. |
| `hooks` _[ExtensionHooks](#extensionhooks)_ | Hooks defines the set of hooks the extension supports |
| `service` _[ExtensionService](#extensionservice)_ | Service defines the configuration of the extension service that the Envoy Gateway Control Plane will call through extension hooks. |


## ExtensionHooks



ExtensionHooks defines extension hooks across all supported runners

_Appears in:_
- [Extension](#extension)

| Field | Description |
| --- | --- |
| `xdsTranslator` _[XDSTranslatorHooks](#xdstranslatorhooks)_ | XDSTranslator defines all the supported extension hooks for the xds-translator runner |


## ExtensionService



ExtensionService defines the configuration for connecting to a registered extension service.

_Appears in:_
- [Extension](#extension)

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
| `certificateRef` _[SecretObjectReference](#secretobjectreference)_ | CertificateRef contains a references to objects (Kubernetes objects or otherwise) that contains a TLS certificate and private keys. These certificates are used to establish a TLS handshake to the extension server. 
 CertificateRef can only reference a Kubernetes Secret at this time. |


## Gateway



Gateway defines the desired Gateway API configuration of Envoy Gateway.

_Appears in:_
- [EnvoyGatewaySpec](#envoygatewayspec)

| Field | Description |
| --- | --- |
| `controllerName` _string_ | ControllerName defines the name of the Gateway API controller. If unspecified, defaults to "gateway.envoyproxy.io/gatewayclass-controller". See the following for additional details: https://gateway-api.sigs.k8s.io/v1alpha2/references/spec/#gateway.networking.k8s.io/v1alpha2.GatewayClass |


## GroupVersionKind



GroupVersionKind unambiguously identifies a Kind. It can be converted to k8s.io/apimachinery/pkg/runtime/schema.GroupVersionKind

_Appears in:_
- [Extension](#extension)

| Field | Description |
| --- | --- |
| `group` _string_ |  |
| `version` _string_ |  |
| `kind` _string_ |  |


## KubernetesContainerSpec



KubernetesContainerSpec defines the desired state of the Kubernetes container resource.

_Appears in:_
- [KubernetesDeploymentSpec](#kubernetesdeploymentspec)

| Field | Description |
| --- | --- |
| `resources` _[ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#resourcerequirements-v1-core)_ | Resources required by this container. More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/ |
| `securityContext` _[SecurityContext](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#securitycontext-v1-core)_ | SecurityContext defines the security options the container should be run with. If set, the fields of SecurityContext override the equivalent fields of PodSecurityContext. More info: https://kubernetes.io/docs/tasks/configure-pod-container/security-context/ |
| `image` _string_ | Image specifies the EnvoyProxy container image to be used, instead of the default image. |


## KubernetesDeploymentSpec



KubernetesDeploymentSpec defines the desired state of the Kubernetes deployment resource.

_Appears in:_
- [EnvoyGatewayKubernetesProvider](#envoygatewaykubernetesprovider)
- [EnvoyProxyKubernetesProvider](#envoyproxykubernetesprovider)

| Field | Description |
| --- | --- |
| `replicas` _integer_ | Replicas is the number of desired pods. Defaults to 1. |
| `pod` _[KubernetesPodSpec](#kubernetespodspec)_ | Pod defines the desired annotations and securityContext of container. |
| `container` _[KubernetesContainerSpec](#kubernetescontainerspec)_ | Container defines the resources and securityContext of container. |


## KubernetesPodSpec



KubernetesPodSpec defines the desired state of the Kubernetes pod resource.

_Appears in:_
- [KubernetesDeploymentSpec](#kubernetesdeploymentspec)

| Field | Description |
| --- | --- |
| `annotations` _object (keys:string, values:string)_ | Annotations are the annotations that should be appended to the pods. By default, no pod annotations are appended. |
| `securityContext` _[PodSecurityContext](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#podsecuritycontext-v1-core)_ | SecurityContext holds pod-level security attributes and common container settings. Optional: Defaults to empty.  See type description for default values of each field. |


## KubernetesServiceSpec



KubernetesServiceSpec defines the desired state of the Kubernetes service resource.

_Appears in:_
- [EnvoyProxyKubernetesProvider](#envoyproxykubernetesprovider)

| Field | Description |
| --- | --- |
| `annotations` _object (keys:string, values:string)_ | Annotations that should be appended to the service. By default, no annotations are appended. |
| `type` _[ServiceType](#servicetype)_ | Type determines how the Service is exposed. Defaults to LoadBalancer. Valid options are ClusterIP and LoadBalancer. "LoadBalancer" means a service will be exposed via an external load balancer (if the cloud provider supports it). "ClusterIP" means a service will only be accessible inside the cluster, via the cluster IP. |


## LogComponent

_Underlying type:_ `string`

LogComponent defines a component that supports a configured logging level. This type is not implemented until https://github.com/envoyproxy/gateway/issues/280 is fixed.

_Appears in:_
- [ProxyLogging](#proxylogging)



## LogLevel

_Underlying type:_ `string`

LogLevel defines a log level for system logs. This type is not implemented until https://github.com/envoyproxy/gateway/issues/280 is fixed.

_Appears in:_
- [ProxyLogging](#proxylogging)



## ProviderType

_Underlying type:_ `string`

ProviderType defines the types of providers supported by Envoy Gateway.

_Appears in:_
- [EnvoyGatewayProvider](#envoygatewayprovider)
- [EnvoyProxyProvider](#envoyproxyprovider)



## ProxyLogging



ProxyLogging defines logging parameters for managed proxies. This type is not implemented until https://github.com/envoyproxy/gateway/issues/280 is fixed.

_Appears in:_
- [EnvoyProxySpec](#envoyproxyspec)

| Field | Description |
| --- | --- |
| `level` _object (keys:[LogComponent](#logcomponent), values:[LogLevel](#loglevel))_ | Level is a map of logging level per component, where the component is the key and the log level is the value. If unspecified, defaults to "System: Info". |


## RateLimit



RateLimit defines the configuration associated with the Rate Limit Service used for Global Rate Limiting.

_Appears in:_
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



RateLimitRedisSettings defines the configuration for connecting to a Redis database.

_Appears in:_
- [RateLimitDatabaseBackend](#ratelimitdatabasebackend)

| Field | Description |
| --- | --- |
| `url` _string_ | URL of the Redis Database. |


## ServiceType

_Underlying type:_ `string`

ServiceType string describes ingress methods for a service

_Appears in:_
- [KubernetesServiceSpec](#kubernetesservicespec)



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


