Config API Design
===================

## Motivation

[Issue 51][issue_51] specifies the need to design an API for configuring Envoy Gateway. This configuration is referred
to as the "Bootstrap Config" in the Envoy Gateway high-level [design doc][design_doc].

## Goals

* Define an API for configuring Envoy Gateway, independent of the runtime platform.
* Specify tooling for managing the API, e.g. generate protos, CRDs, controller RBAC, etc.

## Non-Goals

* Implementation of the API.
* Define the `status` subresource of the configuration API.
* Define provisioning configuration for the control/data planes.

## Proposal

Utilize the [Kubebuilder][kubebuilder] framework to build and manage Envoy Gateway APIs. Kubebuilder facilitates the
following developer workflow for building APIs:

- Create one or more resource APIs as CRDs and then add fields to the resources.
- Implement reconcile loops in controllers and watch additional resources.
- Test by running against a cluster (self-installs CRDs and starts controllers automatically).
- Update bootstrapped integration tests to test new fields and business logic.
- Build and publish a container from a Dockerfile.

After installing the latest Kubebuilder release, create the `BootstrapConfig` API:
```shell
kubebuilder create api --group gateway.envoyproxy.io --version v1alpha1 --kind BootstrapConfig
```

The `v1alpha1` version and `gateway.envoyproxy.io` API group get generated:
```go
// gateway/api/v1alpha1/doc.go

// Package v1alpha1 contains API Schema definitions for the gateway.envoyproxy.io API group.
//
// +groupName=gateway.envoyproxy.io
package v1alpha1
```

The initial `BootstrapConfig` API being proposed:
```go
// gateway/api/v1alpha1/bootstrapconfig_types.go

// BootstrapConfigSpec defines the desired state of BootstrapConfig.
type BootstrapConfigSpec struct {
    // ControlPlane contains parameters of the Envoy Gateway control plane.
    // If unspecified, the control plane will use default configuration settings.
    //
    // +optional
    ControlPlane *ControlPlaneConfig `json:"controlPlane,omitempty"`
}

// ControlPlaneConfig defines configuration of the Envoy Gateway control plane.
type ControlPlaneConfig struct {
    // Type defines the type of control plane used by Envoy Gateway. Valid values are:
    //
    // * Xds
    //
    // In this configuration, xDS is used to manage the data plane.
    //
    // See: https://github.com/cncf/xds
    //
    // +unionDiscriminator
    // +kubebuilder:default=Xds
    Type ControlPlaneType `json:"type,omitempty"`

    // XdsServer defines the xDS Server configuration parameters. Present only if
    // type is "Xds".
    //
    // If unspecified, default configuration parameters are used for the xDS server.
    //
    // +optional
    // +kubebuilder:default={address: 0.0.0.0, port: 18000}
    XdsServer XdsServerConfig `json:"xdsServer,omitempty"`
}

// ControlPlaneType defines a type of control plane used by Envoy Gateway.
// +kubebuilder:validation:Enum=Xds
type ControlPlaneType string

const (
    // XdsControlPlaneType is a control plane that uses xDS to manage a data plane.
    // See https://github.com/cncf/xds for additional details.
    XdsControlPlaneType ControlPlaneType = "Xds"
)

// XdsServerConfig defines configuration of the Envoy Gateway control plane.
type XdsServerConfig struct {
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

// BootstrapConfigStatus defines the observed state of BootstrapConfig.
type BootstrapConfigStatus struct {
    // TODO: Define status fields.
}

type BootstrapConfig struct {
    metav1.TypeMeta   `json:",inline"`
    metav1.ObjectMeta `json:"metadata,omitempty"`

    Spec   BootstrapConfigSpec   `json:"spec,omitempty"`
    Status BootstrapConfigStatus `json:"status,omitempty"`
}
```

[issue_51]: https://github.com/envoyproxy/gateway/issues/51
[design_doc]: https://github.com/envoyproxy/gateway/blob/main/docs/design/SYSTEM_DESIGN.md
[kubebuilder]: https://book-v2.book.kubebuilder.io/
