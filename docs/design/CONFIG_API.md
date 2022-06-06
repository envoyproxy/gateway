Config API Design
===================

## Motivation

[Issue 51][issue_51] specifies the need to design an API for configuring Envoy Gateway. This configuration is referred
to as the "Bootstrap Config" in the Envoy Gateway high-level [design doc][design_doc].

## Goals

* Define an API for configuring Envoy Gateway.
* Specify tooling for managing the API, e.g. generate protos, CRDs, controller RBAC, etc.

## Non-Goals

* Implementation of the API.
* Define the `status` subresource of the configuration API.

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
kubebuilder create api --group envoygateway.io --version v1alpha1 --kind BootstrapConfig
```

The `v1alpha1` version and `envoygateway.io` API group get generated:
```go
// gateway/api/v1alpha1/doc.go

// Package v1alpha1 contains API Schema definitions for the envoygateway.io API group.
//
// +groupName=config.envoygateway.io
package v1alpha1
```

The initial `BootstrapConfig` API being proposed:
```go
// gateway/api/v1alpha1/bootstrapconfig_types.go

// BootstrapConfigSpec defines the desired state of BootstrapConfig.
type BootstrapConfigSpec struct {
	// EnabledServices defines the services to enable. If unspecified,
	// defaults to enabling all supported services.
	//
	// +optional
	// +kubebuilder:validation:MaxItems=12
	// +kubebuilder:validation:Enum=XdsServer;Provisioner
	// +kubebuilder:default=XdsServer;Provisioner
	EnabledServices []ServiceType `json:"enabledServices,omitempty"`

	// XDSServer contains parameters of the xDS server. If unspecified,
	// the xDS server will use default configuration settings.
	//
	// +optional
	XDSServer *XDSServerConfig `json:"xdsServer,omitempty"`

	// Provisioner contains parameters of the provisioner server. If
	// unspecified, the provisioner server will use default configuration
	// settings.
	//
	// +optional
	Provisioner *ProvisionerConfig `json:"Provisioner,omitempty"`
}

// ServiceType defines the type of services supported by Envoy Gateway.
//
// +kubebuilder:validation:Enum=XdsServer;Provisioner
type ServiceType string

const (
	// XdsServerServiceType is the name of the "XdsServer" service type.
    XdsServerServiceType ServiceType = "XdsServer"

	// ProvisionerServiceType is the name of the "Provisioner" service type.
    ProvisionerServiceType ServiceType = "Provisioner"
)

// XDSServerConfig contains the configuration of the xDS server.
type XDSServerConfig struct {
	// Defines the IP address of the server to serve xDS over gRPC.
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

// ProvisionerConfig defines the provisioner server configuration.
type ProvisionerConfig struct {
	// ControlPlane specifies deployment configuration of the control plane,
	// e.g. the xDS server.
	//
	// +optional
	ControlPlane *ControlPlaneConfig `json:"controlPlane,omitempty"`

	// Envoy specifies deployment configuration of the managed Envoy data
	// plane.
	//
	// +optional
	Envoy *EnvoyConfig `json:"envoy,omitempty"`
}

// ControlPlaneConfig defines the configuration of the control plane deployment.
type ControlPlaneConfig struct {
	// Replicas is the desired number of Contour replicas. If unspecified,
	// defaults to 2.
	//
	// +optional
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:default=2
	Replicas *int32 `json:"replicas,omitempty"`

	// NetworkPublishing defines how to publish network endpoint(s) to a network.
	// If unspecified, defaults to "KubeLoadBalancerService" which defines a
	// Kubernetes service of type LoadBalancer.
	//
	// +optional.
	NetworkPublishing *NetworkPublishing `json:"networkPublishing,omitempty"`
}

// EnvoyConfig defines the configuration of the Envoy data plane deployment.
type EnvoyConfig struct {
	// Replicas is the desired number of Envoy replicas. If unspecified,
	// defaults to 2.
	//
	// +optional
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:default=2
	Replicas *int32 `json:"replicas,omitempty"`

	// NetworkPublishing defines how to publish network endpoint(s) to a network.
	// If unspecified, defaults to "KubeLoadBalancerService" which defines a
	// Kubernetes service of type LoadBalancer.
	//
	// +optional
	// +kubebuilder:default={type: KubeLoadBalancerService}
	NetworkPublishing *NetworkPublishing `json:"networkPublishing,omitempty"`
}

// NetworkPublishing defines how to publish network endpoint(s) to a network.
type NetworkPublishing struct {
	// NetworkPublishingType is the type of publishing strategy to use. Valid values are:
	//
	// * KubeLoadBalancerService
	//
	// In this configuration, network endpoints for Envoy use container networking.
	// A Kubernetes LoadBalancer Service is created to publish Envoy network
	// endpoints.
	//
	// See: https://kubernetes.io/docs/concepts/services-networking/service/#loadbalancer
	//
	// * KubeNodePortService
	//
	// Publishes Envoy network endpoints using a Kubernetes NodePort Service.
	//
	// In this configuration, Envoy network endpoints use container networking. A Kubernetes
	// NodePort Service is created to publish the network endpoints.
	//
	// See: https://kubernetes.io/docs/concepts/services-networking/service/#nodeport
	//
	// * KubeClusterIPService
	//
	// Publishes Envoy network endpoints using a Kubernetes ClusterIP Service.
	//
	// In this configuration, Envoy network endpoints use container networking. A Kubernetes
	// ClusterIP Service is created to publish the network endpoints.
	//
	// See: https://kubernetes.io/docs/concepts/services-networking/service/#publishing-services-service-types
	//
	// If unset, defaults to KubeLoadBalancerService.
	//
	// +optional
	// +kubebuilder:default=KubeLoadBalancerService
	Type *NetworkPublishingType `json:"type,omitempty"`
}

// NetworkPublishingType is a way to publish Envoy network endpoints to a network.
type NetworkPublishingType string

const (
	// KubeLoadBalancerPublishingType publishes a network endpoint using a Kubernetes
	// LoadBalancer Service.
	KubeLoadBalancerPublishingType NetworkPublishingType = "KubeLoadBalancerService"

	// KubeNodePortPublishingType publishes a network endpoint using a Kubernetes
	// NodePort Service.
	KubeNodePortPublishingType NetworkPublishingType = "KubeNodePortService"

	// KubeClusterIPPublishingType publishes a network endpoint using a Kubernetes
	// ClusterIP Service.
	KubeClusterIPPublishingType NetworkPublishingType = "KubeClusterIPService"
)

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
