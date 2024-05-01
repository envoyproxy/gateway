// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// KindBackend is the name of the Backend kind.
	KindBackend = "Backend"
)

type ProtocolType string

const (
	// ProtocolTypeTCP defines the TCP address protocol.
	ProtocolTypeTCP ProtocolType = "TCP"
	// ProtocolTypeUDP defines the UDP address protocol.
	ProtocolTypeUDP ProtocolType = "UDP"
)

type ApplicationProtocolType string

const (
	// ApplicationProtocolTypeH2C defines the HTTP/2 prior knowledge application protocol.
	ApplicationProtocolTypeH2C ApplicationProtocolType = "H2C"
	// ApplicationProtocolTypeWS defines the WebSocket over HTTP protocol.
	ApplicationProtocolTypeWS ApplicationProtocolType = "WS"
)

// +kubebuilder:object:root=true
// +kubebuilder:resource:categories=envoy-gateway,shortName=be
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.conditions[?(@.type=="Accepted")].reason`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
//
// Backend allows the user to configure the behavior of the connection
// between the Envoy Proxy listener and the backend service.
type Backend struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// spec defines the desired state of Backend.
	Spec BackendSpec `json:"spec"`

	// status defines the current status of Backend.
	Status BackendStatus `json:"status,omitempty"`
}

// BackendAddress describes are backend address, which is can be either a TCP/UDP socket or a Unix Domain Socket
// corresponding to Envoy's Address: https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/core/v3/address.proto#config-core-v3-address
//
// +kubebuilder:validation:XValidation:rule="(has(self.socketAddress) || has(self.unixDomainSocketAddress))",message="one of socketAddress or unixDomainSocketAddress must be specified"
// +kubebuilder:validation:XValidation:rule="(has(self.socketAddress) && !has(self.unixDomainSocketAddress)) || (!has(self.socketAddress) && has(self.unixDomainSocketAddress))",message="only one of socketAddress or unixDomainSocketAddress can be specified"
type BackendAddress struct {
	// Name is the unique name of the backend address
	Name string `json:"name,omitempty"`

	// SocketAddress is a [FQDN|IP]:[Port] address
	SocketAddress *SocketAddress `json:"socketAddress,omitempty"`

	// UnixDomainSocketAddress is a unix domain socket path
	UnixDomainSocketAddress *UnixDomainSocketAddress `json:"unixDomainSocketAddress,omitempty"`

	// ApplicationProtocol determines the application protocol to be used, e.g. HTTP2.
	ApplicationProtocol *ApplicationProtocolType `json:"applicationProtocol,omitempty"`
}

// SocketAddress describes TCP/UDP socket address, corresponding to Envoy's SocketAddress
// https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/core/v3/address.proto#config-core-v3-socketaddress
type SocketAddress struct {

	// Address refers to the FQDN or IP address of the backend service.
	Address string `json:"address"`

	// Address refers to the FQDN or IP address of the backend service.
	Port int32 `json:"port"`

	// +kubebuilder:validation:Enum=TCP;UDP
	Protocol *ProtocolType `json:"protocol,omitempty"`
}

// UnixDomainSocketAddress describes TCP/UDP unix domain socket address, corresponding to Envoy's Pipe
// https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/core/v3/address.proto#config-core-v3-pipe
type UnixDomainSocketAddress struct {
	Path string `json:"path"`
}

// BackendSpec describes the desired state of BackendSpec.
type BackendSpec struct {
	// +kubebuilder:validation:MaxItems=1
	BackendAddresses []BackendAddress `json:"addresses,omitempty"`
}

// BackendStatus defines the state of Backend
type BackendStatus struct {
	// Conditions describe the current conditions of the Backend.
	//
	// +optional
	// +listType=map
	// +listMapKey=type
	// +kubebuilder:validation:MaxItems=8
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}
