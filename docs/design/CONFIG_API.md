Configuration API Design
===================

## Motivation

[Issue 51][issue_51] specifies the need to design an API for configuring Envoy Gateway at runtime. This configuration is
referred to as the "Bootstrap Config" in the Envoy Gateway high-level [design doc][design_doc]. This document replaces
the term "bootstrap" with "runtime" as some users may confuse bootstrap with Envoy's [bootstrap][envoy_boot]
configuration.

## Goals

* Define an __initial__ API for configuring Envoy Gateway at runtime, independent of the runtime platform, e.g.
  Kubernetes.
* Specify tooling for managing the API, e.g. generate protos, CRDs, controller RBAC, etc.

## Non-Goals

* Implementation of the Envoy Gateway runtime configuration API.
* Define the `status` subresource of the runtime configuration API.
* Define a __complete__ API for configuring Envoy Gateway at runtime. As stated in the [Goals](#goals), this document
  defines the initial runtime configuration API for Envoy Gateway.
* Define an API for deploying/provisioning/operating Envoy Gateway. A separate design document should be used to define
  the Envoy Gateway operator API.

## Proposal

Utilize the [Kubebuilder][kubebuilder] framework to build and manage Envoy Gateway APIs. Kubebuilder facilitates the
following developer workflow for building APIs:

- Create one or more resource APIs as CRDs and then add fields to the resources.
- Implement reconcile loops in controllers and watch additional resources.
- Test by running against a cluster (self-installs CRDs and starts controllers automatically).
- Update bootstrapped integration tests to test new fields and business logic.
- Build and publish a container from a Dockerfile.

After installing the latest Kubebuilder release, create the `RuntimeConfig` API:
```shell
kubebuilder create api --group gateway.envoyproxy.io --version v1alpha1 --kind RuntimeConfig
```

The `v1alpha1` version and `gateway.envoyproxy.io` API group get generated:
```go
// gateway/api/v1alpha1/doc.go

// Package v1alpha1 contains API Schema definitions for the gateway.envoyproxy.io API group.
//
// +groupName=gateway.envoyproxy.io
package v1alpha1
```

The initial `RuntimeConfig` API being proposed:
```go
package valpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/gateway/pkg/provider/file"
	kube "github.com/gateway/pkg/provider/kubernetes"
)

// RuntimeConfigSpec defines the desired state of RuntimeConfig.
type RuntimeConfigSpec struct {
	// Providers define provider configuration. If unspecified, the Kubernetes
	// provider is used with default parameters.
	//
	// +optional
	// Note: the following default is provided for illustration purposes only.
	// +kubebuilder:default={kubernetes: {foo: bar}}
	Providers *Providers `json:"providers,omitempty"`

	// ControlPlane defines control plane configuration parameters. If unspecified,
	// the control plane will use default configuration parameters.
	//
	// +optional
	ControlPlane *ControlPlane `json:"controlPlane,omitempty"`

	// DataPlane defines data plane configuration parameters. If unspecified,
	// the data plane will use default configuration parameters.
	//
	// +optional
	DataPlane *DataPlane `json:"dataPlane,omitempty"`
}

// Providers defines configuration of Envoy Gateway providers.
type Providers struct {
	// Kubernetes defines the configuration of the Kubernetes provider. Kubernetes
	// provides runtime configuration via the Kubernetes API.
	//
	// +optional
	Kubernetes *kube.Provider `json:"kubernetes,omitempty"`

	// File defines the configuration of the File provider. File provides runtime
	// configuration defined by one or more files.
	//
	// +optional
	File *file.Provider `json:"file,omitempty"`
}

// ControlPlane defines configuration of the Envoy Gateway control plane.
type ControlPlane struct {
	// XdsServer defines the xDS Server configuration parameters. If unspecified,
	// default configuration parameters are used for the xDS server.
	//
	// +optional
	// +kubebuilder:default={address: 0.0.0.0, port: 18000}
	XdsServer XdsServer `json:"xdsServer,omitempty"`
}

// XdsServer defines configuration of the Envoy Gateway xDS server.
type XdsServer struct {
	// Defines the IP address for server to serve xDS over gRPC.
	//
	// If unspecified, defaults to "0.0.0.0", e.g. all IP addresses.
	//
	// +optional
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:default=0.0.0.0
	Address *string `json:"address,omitempty"`

	// Defines the network port for the xDS server to serve xDS over gRPC.
	//
	// If unspecified, defaults is 18000.
	//
	// +optional
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	// +kubebuilder:default=18000
	Port *int32 `json:"port,omitempty"`
}

// DataPlane defines configuration of the Envoy Gateway data plane.
type DataPlane struct {
	// TODO: Define data plane configuration fields.
}

// RuntimeConfigStatus defines the observed state of RuntimeConfig.
type RuntimeConfigStatus struct {
	// TODO: Define status fields.
}

type RuntimeConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RuntimeConfigSpec   `json:"spec,omitempty"`
	Status RuntimeConfigStatus `json:"status,omitempty"`
}
```

Provider-specific configuration is defined in the provider package. The following is an example of the Kubernetes
provider:
```go
package kubernetes

// Provider defines the configuration of the Kubernetes provider.
type Provider struct {
	// TODO: Define Kubernetes configuration fields, e.g. restrict namespaces to watch
	//       Gateway/HTTPRoute resources.
}
```

The following is an example of the File provider:
```go
package file

// Provider defines the configuration of the File provider.
type Provider struct {
	// TODO: Define Kubernetes configuration fields, e.g. path to GatewayClass, Gateway,
	//       and HTTPRoute manifests.
}
```

### Spec

Top-level configuration is exposed through `spec` fields. The only spec fields specified by this design are `providers`,
`controlplane`, and `dataplane`. However, additional fields can be introduced in the future to expose system-wide
configuration. The following sections provide details of these top-level fields.

### Providers

Envoy Gateway runtime configuration is achieved through Providers. The providers are infrastructure components that
Envoy Gateway calls to establish its runtime configuration. Refer to the Envoy Gateway [design doc][design_doc] for
additional details.

### Control Plane

Envoy Gateway control plane configuration is exposed through the `ControlPlane` API. This API allows users to
specify provider-agnostic control plane configuration parameters. Currently, `ControlPlane` only exposes
[Xds][xds] server configuration parameters but the API supports adding configuration management of other Envoy Gateway
control plane components in the future.

### Data Plane

Envoy Gateway data plane configuration is exposed through the `DataPlane` API. This API allows users to specify
provider-agnostic data plane configuration settings. This API is provided in this design doc for completeness, but
should not be a defined type until data plane configuration management use cases are better understood.

### Configuration Examples

Run Envoy Gateway using the Kubernetes provider:
```yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: RuntimeConfig
metadata:
  name: example
spec: {}
```
Since Kubernetes is the default provider, no configuration is required.

The Kubernetes provider can be configured explicitly using `spec.providers.kubernetes`:
```yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: RuntimeConfig
metadata:
  name: example
spec:
  providers:
    kubernetes: {}
```
This configuration will cause Envoy Gateway to use the Kubernetes provider with default configuration settings.

The Kubernetes provider can be configured using the `Provider` API from the Kubernetes provider package. For example,
the `foo` field can be set to "bar":
```yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: RuntimeConfig
metadata:
  name: example
spec:
  providers:
    kubernetes:
      foo: bar
```
__Note:__ The Kubernetes `Provider` API is currently undefined and `foo: bar` is provided for illustration purposes
only.

The same API structure is followed for each supported provider. Multiple providers can be specified, for example:
```yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: RuntimeConfig
metadata:
  name: example
spec:
  providers:
  - file:
      foo: bar
  - futureProvider:
      foo: bar
```

Envoy Gateway will parse the list of providers to ensure only supported provider combinations are specified. If the
provider list is valid, Envoy Gateway will merge the configuration from each provider into its IR. Envoy Gateway's
Translator is then responsible for translating the IR into xDS resources that are pushed to managed data plane
instances.

__Note:__ To learn more about the Translator, refer to the Envoy Gateway high-level [design doc][design_doc].

As with Providers, the control plane and data plane can be configured using `spec.controlplane` and `spec.dataplane`
respectively. For example:
```yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: RuntimeConfig
metadata:
  name: example
spec:
  controlplane: {}
  dataplane: {}
```
When unspecified, Envoy Gateway will use default configuration parameters for control and data planes.

The following example causes the control plane to listen on `127.0.0.1:1234` instead of the default `0.0.0.0:18000`
```yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: RuntimeConfig
metadata:
  name: example
spec:
  controlplane:
    type: Xds
    xdsServer:
      address: 127.0.0.1
      port: 1234
```

### Outstanding Questions
- Should an invalid provider list cause Envoy Gateway to not start or should providers be given priority. For example,
  should `providers=Kubernetes,Docker` cause Envoy Gateway to not start or start using the Kubernetes provider
  (assuming the Kubernetes provider configuration is valid)?
- Should we establish a maximum number of configured providers?

[issue_51]: https://github.com/envoyproxy/gateway/issues/51
[design_doc]: https://github.com/envoyproxy/gateway/blob/main/docs/design/SYSTEM_DESIGN.md
[envoy_boot]: https://www.envoyproxy.io/docs/envoy/latest/configuration/overview/bootstrap
[kubebuilder]: https://book-v2.book.kubebuilder.io/
[docker]: https://docs.docker.com/engine/api/v1.41/
[xds]: https://github.com/cncf/xds
[gw_api]: https://gateway-api.sigs.k8s.io/
