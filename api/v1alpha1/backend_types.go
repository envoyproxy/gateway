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

// +kubebuilder:validation:Enum=FQDN;UDS;IPv4
// +notImplementedHide
type AddressType string

const (
	// AddressTypeFQDN defines the RFC-1123 compliant fully qualified domain name address type.
	AddressTypeFQDN ProtocolType = "FQDN"
	// AddressTypeUDS defines the unix domain socket address type.
	AddressTypeUDS ProtocolType = "UDS"
	// AddressTypeIPv4 defines the IPv4 address type.
	AddressTypeIPv4 ProtocolType = "IPv4"
)

// +kubebuilder:validation:Enum=TCP;UDP
// +notImplementedHide
type ProtocolType string

const (
	// ProtocolTypeTCP defines the TCP address protocol.
	ProtocolTypeTCP ProtocolType = "TCP"
	// ProtocolTypeUDP defines the UDP address protocol.
	ProtocolTypeUDP ProtocolType = "UDP"
)

// +kubebuilder:validation:Enum=HTTP2;WS
// +notImplementedHide
type ApplicationProtocolType string

const (
	// ApplicationProtocolTypeHTTP2 defines the HTTP/2 application protocol.
	ApplicationProtocolTypeHTTP2 ApplicationProtocolType = "HTTP2"
	// ApplicationProtocolTypeWS defines the WebSocket over HTTP protocol.
	ApplicationProtocolTypeWS ApplicationProtocolType = "WS"
)

// +kubebuilder:object:root=true
// +kubebuilder:resource:categories=envoy-gateway,shortName=be
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.conditions[?(@.type=="Accepted")].reason`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
// +notImplementedHide
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
// corresponding to Envoy's Host: https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/core/v3/address.proto#config-core-v3-address
//
// +kubebuilder:validation:XValidation:rule="(has(self.socketAddress) || has(self.unixDomainSocketAddress))",message="one of socketAddress or unixDomainSocketAddress must be specified"
// +kubebuilder:validation:XValidation:rule="(has(self.socketAddress) && !has(self.unixDomainSocketAddress)) || (!has(self.socketAddress) && has(self.unixDomainSocketAddress))",message="only one of socketAddress or unixDomainSocketAddress can be specified"
// +kubebuilder:validation:XValidation:rule="((has(self.socketAddress) && (self.type == 'FQDN' || self.type == 'IPv4')) || has(self.unixDomainSocketAddress) && self.type == 'UDS')",message="if type is FQDN or IPv4, socketAddress must be set; if type is UDS, unixDomainSocketAddress must be set"
// +notImplementedHide
type BackendAddress struct {
	// Type is the the type name of the backend address: FQDN, UDS, IPv4
	Type AddressType `json:"type"`

	// NetworkSocket defines a FQDN or IPv4 address
	//
	// +optional
	NetworkSocket *NetworkSocket `json:"networkSocket,omitempty"`

	// UnixSocket defines the unix domain socket path
	//
	// +optional
	UnixSocket *UnixSocket `json:"unixSocket,omitempty"`
}

// NetworkSocket describes TCP/UDP socket address, corresponding to Envoy's NetworkSocket
// https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/core/v3/address.proto#config-core-v3-socketaddress
// +notImplementedHide
type NetworkSocket struct {
	// Host defines to the FQDN or IP address of the backend service.
	Host string `json:"address"`

	// Port defines to the port of of the backend service.
	//
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=65535
	Port int32 `json:"port"`
}

// UnixSocket describes TCP/UDP unix domain socket address, corresponding to Envoy's Pipe
// https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/core/v3/address.proto#config-core-v3-pipe
// +notImplementedHide
type UnixSocket struct {
	Path string `json:"path"`
}

// BackendSpec describes the desired state of BackendSpec.
// +notImplementedHide
type BackendSpec struct {
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=4
	// +kubebuilder:validation:XValidation:rule="self.all(f, f.type == 'FQDN') || !self.exists(f, f.type == 'FQDN')",message="FQDN addresses cannot be mixed with other address types"
	BackendAddresses []BackendAddress `json:"addresses,omitempty"`

	// ApplicationProtocol defines the application protocol to be used, e.g. HTTP2.
	//
	// +optional
	ApplicationProtocol *ApplicationProtocolType `json:"applicationProtocol,omitempty"`
}

// BackendStatus defines the state of Backend
// +notImplementedHide
type BackendStatus struct {
	// Conditions describe the current conditions of the Backend.
	//
	// +optional
	// +listType=map
	// +listMapKey=type
	// +kubebuilder:validation:MaxItems=8
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +notImplementedHide
// BackendList contains a list of Backend resources.
type BackendList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Backend `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Backend{}, &BackendList{})
}
