Configuration API Design
===================

# Table of Contents
1. [Motivation](#motiviation)
2. [Goals](#goals)
3. [Non-Goals](#non-goals)
4. [Control Plane API](#control_plane_api)
   1. [Gateway Type](#gateway)
   2. [Provider Type](#provider)
   3. [Configuration Examples](#control_plane_configuration)
6. [Data Plane API](#data_plane_api)
   1. [Configuration Examples](#data_plane_configuration)

## Motivation
[Issue 51][issue_51] specifies the need to design an API for configuring Envoy Gateway. The control plane is configured
statically at startup and the data plane is configured dynamically through Kubernetes resources, primarily
[Gateway API][gw_api] objects. Refer to the Envoy Gateway [design doc][design_doc] for additional details regarding
Envoy Gateway terminology and configuration.

## Goals
* Define an __initial__ API to configure Envoy Gateway at startup.
* Define an __initial__ API for configuring the managed data plane, e.g. Envoy proxies.

## Non-Goals
* Implementation of the configuration APIs.
* Define the `status` subresource of the configuration APIs.
* Define a __complete__ set of APIs for configuring Envoy Gateway. As stated in the [Goals](#goals), this document
  defines the initial configuration APIs.
* Define an API for deploying/provisioning/operating Envoy Gateway. If needed, a future Envoy Gateway operator would be
  responsible for designing and implementing this type of API.
* Specify tooling for managing the API, e.g. generate protos, CRDs, controller RBAC, etc.

## Control Plane API
The `EnvoyGatewaySpec` API defines the control plane configuration, e.g. Envoy Gateway. Key points of this API are:

* The name `EnvoyGatewaySpec` is being used instead of `EnvoyGateway` since the API may be embedded in a higher-level
  API to surface control plane status.
* It will define Envoy Gateway's startup configuration file. If the file does not exist, Envoy Gateway will start up
  with default configuration parameters.
* If data plane static configuration is required in the future, Envoy Gateway will use a separate file for this purpose.

The `v1alpha1` version and `config.gateway.envoyproxy.io` API group get generated:
```go
// gateway/api/config/v1alpha1/doc.go

// Package v1alpha1 contains API Schema definitions for the config.gateway.envoyproxy.io API group.
//
// +groupName=config.gateway.envoyproxy.io
package v1alpha1
```

The initial `EnvoyGatewaySpec` API being proposed:
```go
// gateway/api/config/v1alpha1/envoygateway.go

package valpha1

import (
	"github.com/gateway/pkg/provider/file"
	"github.com/gateway/pkg/provider/kubernetes"
)

// EnvoyGatewaySpec defines the desired state of Envoy Gateway configuration.
type EnvoyGatewaySpec struct {
	// Gateway defines Gateway-API specific configuration. If unset, default
	// configuration parameters will apply.
	//
	// +optional
	Gateway *Gateway `json:"gateway,omitempty"`

	// Provider defines the desired provider configuration. If unspecified,
	// the Kubernetes provider is used with default parameters.
	//
	// +optional
	Provider *Provider `json:"provider,omitempty"`
}

// Gateway defines desired Gateway API configuration of Envoy Gateway.
type Gateway struct {
	// ControllerName defines the name of the Gateway API controller. If unspecified,
	// defaults to "gateway.envoyproxy.io/gatewayclass-controller". See the following
	// for additional details:
	//
	// https://gateway-api.sigs.k8s.io/v1alpha2/references/spec/#gateway.networking.k8s.io/v1alpha2.GatewayClass
	//
	// +optional
	ControllerName string `json:"controllerName,omitempty"`
}

// Provider defines the desired configuration of a provider.
// +union
type Provider struct {
	// Type is the type of provider to use. If unset, the Kubernetes provider is used.
	//
	// +unionDiscriminator
  Type ProviderType `json:"type,omitempty"`
	// Kubernetes defines the configuration of the Kubernetes provider. Kubernetes
	// provides runtime configuration via the Kubernetes API.
	//
	// +optional
	Kubernetes *kubernetes.Provider `json:"kubernetes,omitempty"`

	// File defines the configuration of the File provider. File provides runtime
	// configuration defined by one or more files.
	//
	// +optional
	File *file.Provider `json:"file,omitempty"`
}

// ProviderType defines the types of providers supported by Envoy Gateway.
type ProviderType string

const (
	// KubernetesProviderType defines the "Kubernetes" provider.
	KubernetesProviderType ProviderType = "Kubernetes"

	// FileProviderType defines the "File" provider.
	FileProviderType ProviderType = "File"
)
```
Note that a provider-specific configuration is defined in the provider package. The following is an example of the
Kubernetes provider:
```go
// gateway/pkg/kubernetes/kubernetes.go

package kubernetes

// Provider defines the configuration of the Kubernetes provider.
type Provider struct {
	// TODO: Define Kubernetes configuration fields, e.g. restrict namespaces to watch
	//       Gateway/HTTPRoute resources.
}
```

### Gateway
Gateway defines desired configuration of [Gateway API][gw_api] controllers that reconcile and translate Gateway API
resources into the Intermediate Representation (IR). Refer to the Envoy Gateway [design doc][design_doc] for additional
details.

### Provider
Provider defines the desired configuration of an Envoy Gateway provider. A provider is an infrastructure component that
Envoy Gateway calls to establish its runtime configuration. Provider is a [union type][union]. Therefore, Envoy Gateway
can be configured with only one provider based on the `type` discriminator field. Refer to the Envoy Gateway
[design doc][design_doc] for additional details.

### Configuration Examples
The configuration file is defined by the EnvoyGatewaySpec API type. At startup, Envoy Gateway searches for the
configuration at "/etc/envoy-gateway/config.yaml".

Start Envoy Gateway:
```shell
envoy-gateway start
```
Since the configuration file does not exist, Envoy Gateway will start with default configuration parameters.

The Kubernetes provider can be configured explicitly using `provider.kubernetes`:
```yaml
$ cat << EOF > /etc/envoy-gateway/config.yaml
provider:
  type: Kubernetes
  kubernetes: {}
EOF
```
This configuration will cause Envoy Gateway to use the Kubernetes provider with default configuration parameters.

The Kubernetes provider can be configured using the `provider` field. For example, the `foo` field can be set to "bar":
```yaml
$ cat << EOF > /etc/envoy-gateway/config.yaml
provider:
  type: Kubernetes
  kubernetes:
    foo: bar
EOF
```
__Note:__ The Provider API from the Kubernetes package is currently undefined and `foo: bar` is provided for
illustration purposes only.

The same API structure is followed for each supported provider. The following example causes Envoy Gateway to use the
File provider:
```yaml
$ cat << EOF > /etc/envoy-gateway/config.yaml
provider:
  type: File
  file:
    foo: bar
EOF
```
__Note:__ The Provider API from the File package is currently undefined and `foo: bar` is provided for illustration
purposes only.

Gateway API-related configuration is expressed through the `gateway` field. If unspecified, Envoy Gateway will use
default configuration parameters for `gateway`. The following example causes the [GatewayClass][gc] controller to
manage GatewayClasses with controllerName `foo` instead of the default `gateway.envoyproxy.io/gatewayclass-controller`:
```yaml
$ cat << EOF > /etc/envoy-gateway/config.yaml
gateway:
  controllerName: foo
```

With any of the above configuration examples, Envoy Gateway can be started without any additional arguments:
```shell
envoy-gateway start
```

## Data Plane API
The data plane is configured dynamically through Kubernetes resources, primarily [Gateway API][gw_api] objects.
Optionally, the data plane infrastructure can be configured by referencing a [custom resource (CR)][cr] through
`spec.parametersRef` of the managed GatewayClass. The `envoyproxies` API defines the data plane infrastructure
configuration and is represented as the CR referenced by the managed GatewayClass. Key points of this API are:

* If unreferenced by `spec.parametersRef`, default parameters will be used to configure the data plane infrastructure,
  e.g. expose Envoy network endpoints using a LoadBalancer service.
* Envoy Gateway will follow Gateway API [recommendations][gc] regarding updates to the EnvoyProxy CR:
  > It is recommended that this resource be used as a template for Gateways. This means that a Gateway is based on the
  > state of the GatewayClass at the time it was created and changes to the GatewayClass or associated parameters are
  > not propagated down to existing Gateways.

The initial `envoyproxies` API being proposed:
```go
// gateway/api/config/v1alpha1/envoyproxy.go

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EnvoyProxy is the Schema for the envoyproxies API.
type EnvoyProxy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   EnvoyProxySpec   `json:"spec,omitempty"`
	Status EnvoyProxyStatus `json:"status,omitempty"`
}

// EnvoyProxySpec defines the desired state of Envoy Proxy infrastructure
// configuration.
type EnvoyProxySpec struct {
	// Undefined by this design spec.
}

// EnvoyProxyStatus defines the observed state of EnvoyProxy.
type EnvoyProxyStatus struct {
	// Undefined by this design spec.
}
```
The EnvoyProxySpec and EnvoyProxyStatus fields will be defined in the future as proxy infrastructure configuration use
cases are better understood.

### Configuration Examples
GatewayClass and Gateway resources define the data plane infrastructure. Note that all examples assume Envoy Gateway is
running with the Kubernetes provider.
```yaml
apiVersion: gateway.networking.k8s.io/v1beta1
kind: GatewayClass
metadata:
  name: example-class
spec:
  controllerName: gateway.envoyproxy.io/gateway-controller
---
apiVersion: gateway.networking.k8s.io/v1beta1
kind: Gateway
metadata:
  name: example-gateway
spec:
  gatewayClassName: example-class
  listeners:
  - name: http
    protocol: HTTP
    port: 80
```
Since the GatewayClass does not define `spec.parametersRef`, the data plane is provisioned using default configuration
parameters. All Envoy proxies will be configured with a http listener and a Kubernetes LoadBalancer service listening
on port 80.

The following example will configure the data plane to use a ClusterIP service instead of the default LoadBalancer
service:
```yaml
apiVersion: gateway.networking.k8s.io/v1beta1
kind: GatewayClass
metadata:
  name: example-class
spec:
  controllerName: gateway.envoyproxy.io/gateway-controller
  parametersRef:
    name: example-config
    group: config.gateway.envoyproxy.io
    kind: EnvoyProxy
---
apiVersion: gateway.networking.k8s.io/v1beta1
kind: Gateway
metadata:
  name: example-gateway
spec:
  gatewayClassName: example-class
  listeners:
  - name: http
    protocol: HTTP
    port: 80
---
apiVersion: config.gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: example-config
spec:
  networkPublishing:
    type: NodePortService
```
__Note:__ The NetworkPublishing API is currently undefined and is provided here for illustration purposes only.

[issue_51]: https://github.com/envoyproxy/gateway/issues/51
[design_doc]: https://github.com/envoyproxy/gateway/blob/main/docs/design/SYSTEM_DESIGN.md
[xds]: https://github.com/cncf/xds
[gw_api]: https://gateway-api.sigs.k8s.io/
[config_guide]: https://github.com/envoyproxy/gateway/blob/main/docs/CONFIG.md
[gc]: https://gateway-api.sigs.k8s.io/v1alpha2/references/spec/#gateway.networking.k8s.io/v1alpha2.GatewayClass
[cr]: https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/
