Configuration API Design
===================

## Motivation
[Issue 51][issue_51] specifies the need to design an API for configuring Envoy Gateway at runtime. This configuration is
referred to as the "static configuration" in the Envoy Gateway [design doc][design_doc].

## Goals
* Define an __initial__ API for configuring Envoy Gateway at runtime, independent of the runtime platform, e.g.
  Kubernetes.

## Non-Goals
* Implementation of the Envoy Gateway runtime configuration API.
* Define the `status` subresource of the runtime configuration API.
* Define a __complete__ API for configuring Envoy Gateway at runtime. As stated in the [Goals](#goals), this document
  defines the initial runtime configuration API for Envoy Gateway.
* Define an API for deploying/provisioning/operating Envoy Gateway.
* Define an __initial__ provider-independent API for configuring the managed Envoy proxy infrastructure.
* Specify tooling for managing the API, e.g. generate protos, CRDs, controller RBAC, etc.

## API Definition
`ControlPlaneSpec` defines the runtime configuration of Envoy Gateway. The name `ControlPlaneSpec` is being used
instead of `ControlPlane` since the API may be embedded in a higher-level API that is represented as a Kubernetes
resource.

The `v1alpha1` version and `gateway.envoyproxy.io` API group get generated:
```go
// gateway/api/v1alpha1/doc.go

// Package v1alpha1 contains API Schema definitions for the gateway.envoyproxy.io API group.
//
// +groupName=gateway.envoyproxy.io
package v1alpha1
```

The initial `ControlPlaneSpec` API being proposed:
```go
// gateway/api/v1alpha1/controlplane.go

package valpha1

import (
	"github.com/gateway/pkg/provider/file"
	"github.com/gateway/pkg/provider/kubernetes"
)

// ControlPlaneSpec defines the desired state of Envoy Gateway configuration.
type ControlPlaneSpec struct {
	// Gateway defines Gateway-API specific configuration. If unset, default
	// configuration parameters will apply.
	//
	// +optional
	Gateway *Gateway `json:"gateway,omitempty"`

	// Providers define provider configuration. If unspecified, the Kubernetes
	// provider is used with default parameters.
	//
	// +optional
	Providers *Providers `json:"providers,omitempty"`

	// XdsServer defines the xDS Server configuration parameters. If unspecified,
	// default configuration parameters are applied.
	//
	// +optional
	XdsServer XdsServer `json:"xdsServer,omitempty"`
}

// Gateway defines desired Gateway API configuration of Envoy Gateway.
type Gateway struct {
	// controllerName defines the name of the Gateway API controller. If unspecified,
	// defaults to "gateway.envoyproxy.io/gatewayclass-controller". See the following
	// for additional details:
	//
	// https://gateway-api.sigs.k8s.io/v1alpha2/references/spec/#gateway.networking.k8s.io/v1alpha2.GatewayClass
	//
	// +optional
	controllerName string `json:"controllerName,omitempty"`
}

// Providers define desired configuration of Envoy Gateway providers.
type Providers struct {
	// Kubernetes defines the configuration of the Kubernetes provider. Kubernetes
	// provides runtime configuration via the Kubernetes API.
	//
	// +support:alpha
	// +optional
	Kubernetes *kubernetes.Provider `json:"kubernetes,omitempty"`

	// File defines the configuration of the File provider. File provides runtime
	// configuration defined by one or more files.
	//
	// +support:alpha
	// +optional
	File *file.Provider `json:"file,omitempty"`
}

// XdsServer defines the desired configuration of the Envoy Gateway xDS server.
type XdsServer struct {
	// Defines the IP address for server to serve xDS over gRPC.
	//
	// If unspecified, defaults to "0.0.0.0", e.g. all IP addresses.
	//
	// +optional
	Address *string `json:"address,omitempty"`

	// Defines the network port for the xDS server to serve xDS over gRPC.
	//
	// If unspecified, defaults is 18000.
	//
	// +optional
	Port *int32 `json:"port,omitempty"`
}
```
Note that a provider-specific configuration is defined in the provider package. The following is an example of the
Kubernetes provider:
```go
// gateway/internal/kubernetes/kubernetes.go

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

### Providers
Providers define desired configuration of Envoy Gateway providers. Providers are infrastructure components that Envoy
Gateway calls to establish its runtime configuration. Providers are defined by a `+support` marker to indicate the
stability of the provider. Refer to the Envoy Gateway [design doc][design_doc] for additional details.

### XdsServer
XdsServer defines desired configuration of [Xds][xds] server configuration parameters. Refer to the Envoy Gateway
[design doc][design_doc] for additional details.

### Configuration Example
The configuration file is defined by the ControlPlaneSpec API type. At startup, Envoy Gateway searches for the
configuration at "/etc/envoy-gateway/config.yaml".

Start Envoy Gateway:
```shell
envoy-gateway start
```
Since the configuration file does not exist, Envoy Gateway will start with default configuration parameters.

The Kubernetes provider can be configured explicitly using `providers.kubernetes`:
```yaml
$ cat << EOF > /etc/envoy-gateway/config.yaml
providers:
  kubernetes: {}
EOF
```
This configuration will cause Envoy Gateway to use the Kubernetes provider with default configuration settings.

The Kubernetes provider can be configured using the `Providers` API. For example, the `foo` field can be set to "bar":
```yaml
$ cat << EOF > /etc/envoy-gateway/config.yaml
providers:
  kubernetes:
    foo: bar
EOF
```
__Note:__ The `Providers` API from the Kubernetes package is currently undefined and `foo: bar` is provided for
illustration purposes only.

The same API structure is followed for each supported provider. Multiple providers can be specified, for example:
```yaml
$ cat << EOF > /etc/envoy-gateway/config.yaml
providers:
  file:
    foo: bar
  futureProvider:
      foo: bar
EOF
```

Envoy Gateway will parse providers to ensure only supported provider combinations are specified. If providers are valid,
Envoy Gateway will merge the configuration from each provider into its IR. Envoy Gateway's Translator is then
responsible for translating the IR into xDS resources that are pushed to managed data plane instances.

__Note:__ To learn more about the Translator, refer to the Envoy Gateway [design doc][design_doc].

As with Providers, Gateway API and xDS control plane services can be configured using `spec.gateway` and
`spec.xdsServer`respectively. For example:
```yaml
$ cat << EOF > /etc/envoy-gateway/config.yaml
  gateway: {}
  xdsServer: {}
```
When unspecified, Envoy Gateway will use default configuration parameters for gateway and xdsServer fields.

The following example causes the GatewayClass controller to manage GatewayClasses with controllerName `foo` instead of
the default `gateway.envoyproxy.io/gatewayclass-controller`:
```yaml
$ cat << EOF > /etc/envoy-gateway/config.yaml
gateway:
  controllerName: foo
```

The following example causes the xDS Server to listen on localhost instead of all IP addresses (default):
```yaml
$ cat << EOF > /etc/envoy-gateway/config.yaml
xdsServer:
  address: 127.0.0.1
```

With any of the above configuration examples, you can Start Envoy Gateway without any additional arguments:
```shell
envoy-gateway start
```

### Outstanding Questions
- Should an invalid provider combination cause Envoy Gateway to not start or should providers be given priority. For
  example, should `providers=Kubernetes,File` cause Envoy Gateway to not start or start using the Kubernetes provider
  (assuming the Kubernetes provider configuration is valid)?
- Should we establish a maximum number of configured providers?

[issue_51]: https://github.com/envoyproxy/gateway/issues/51
[design_doc]: https://github.com/envoyproxy/gateway/blob/main/docs/design/SYSTEM_DESIGN.md
[xds]: https://github.com/cncf/xds
[gw_api]: https://gateway-api.sigs.k8s.io/
[config_guide]: https://github.com/envoyproxy/gateway/blob/main/docs/CONFIG.md
