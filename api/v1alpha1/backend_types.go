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

// AppProtocolType defines various backend applications protocols supported by Envoy Gateway
//
// +kubebuilder:validation:Enum=gateway.envoyproxy.io/h2c;gateway.envoyproxy.io/ws;gateway.envoyproxy.io/wss
type AppProtocolType string

const (
	// AppProtocolTypeH2C defines the HTTP/2 application protocol.
	AppProtocolTypeH2C AppProtocolType = "gateway.envoyproxy.io/h2c"
	// AppProtocolTypeWS defines the WebSocket over HTTP protocol.
	AppProtocolTypeWS AppProtocolType = "gateway.envoyproxy.io/ws"
	// AppProtocolTypeWSS defines the WebSocket over HTTPS protocol.
	AppProtocolTypeWSS AppProtocolType = "gateway.envoyproxy.io/wss"
)

// Backend allows the user to configure the endpoints of a backend and
// the behavior of the connection from Envoy Proxy to the backend.
//
// +kubebuilder:object:root=true
// +kubebuilder:resource:categories=envoy-gateway,shortName=be
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.conditions[?(@.type=="Accepted")].reason`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
type Backend struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the desired state of Backend.
	Spec BackendSpec `json:"spec"`

	// Status defines the current status of Backend.
	Status BackendStatus `json:"status,omitempty"`
}

// BackendEndpoint describes a backend endpoint, which can be either a fully-qualified domain name, IP address or unix domain socket
// corresponding to Envoy's Address: https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/core/v3/address.proto#config-core-v3-address
//
// +kubebuilder:validation:XValidation:rule="(has(self.fqdn) || has(self.ip) || has(self.unix))",message="one of fqdn, ip or unix must be specified"
// +kubebuilder:validation:XValidation:rule="((has(self.fqdn) && !(has(self.ip) || has(self.unix))) || (has(self.ip) && !(has(self.fqdn) || has(self.unix))) || (has(self.unix) && !(has(self.ip) || has(self.fqdn))))",message="only one of fqdn, ip or unix can be specified"
type BackendEndpoint struct {
	// FQDN defines a FQDN endpoint
	//
	// +optional
	FQDN *FQDNEndpoint `json:"fqdn,omitempty"`

	// IP defines an IP endpoint. Currently, only IPv4 Addresses are supported.
	//
	// +optional
	IP *IPEndpoint `json:"ip,omitempty"`

	// Unix defines the unix domain socket endpoint
	//
	// +optional
	Unix *UnixSocket `json:"unix,omitempty"`
}

// IPEndpoint describes TCP/UDP socket address, corresponding to Envoy's Socket Address
// https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/core/v3/address.proto#config-core-v3-socketaddress
type IPEndpoint struct {
	// Address defines the IP address of the backend endpoint.
	//
	// +kubebuilder:validation:MinLength=7
	// +kubebuilder:validation:MaxLength=15
	// +kubebuilder:validation:Pattern=`^((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$`
	Address string `json:"address"`

	// Port defines the port of the backend endpoint.
	//
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=65535
	Port int32 `json:"port"`
}

// FQDNEndpoint describes TCP/UDP socket address, corresponding to Envoy's Socket Address
// https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/core/v3/address.proto#config-core-v3-socketaddress
type FQDNEndpoint struct {
	// Hostname defines the FQDN hostname of the backend endpoint.
	//
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=253
	// +kubebuilder:validation:Pattern=`^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9]))*$`
	Hostname string `json:"hostname"`

	// Port defines the port of the backend endpoint.
	//
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=65535
	Port int32 `json:"port"`
}

// UnixSocket describes TCP/UDP unix domain socket address, corresponding to Envoy's Pipe
// https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/core/v3/address.proto#config-core-v3-pipe
type UnixSocket struct {
	// Path defines the unix domain socket path of the backend endpoint.
	Path string `json:"path"`
}

// BackendSpec describes the desired state of BackendSpec.
type BackendSpec struct {
	// Endpoints defines the endpoints to be used when connecting to the backend.
	//
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=4
	// +kubebuilder:validation:XValidation:rule="self.all(f, has(f.fqdn)) || !self.exists(f, has(f.fqdn))",message="fqdn addresses cannot be mixed with other address types"
	Endpoints []BackendEndpoint `json:"endpoints,omitempty"`

	// AppProtocols defines the application protocols to be supported when connecting to the backend.
	//
	// +optional
	AppProtocols []AppProtocolType `json:"appProtocols,omitempty"`

	// Fallback indicates whether the backend is designated as a fallback.
	// It is highly recommended to configure active or passive health checks to ensure that failover can be detected
	// when the active backends become unhealthy and to automatically readjust once the primary backends are healthy again.
	// The overprovisioning factor is set to 1.4, meaning the fallback backends will only start receiving traffic when
	// the health of the active backends falls below 72%.
	//
	// +optional
	// +notImplementedHide
	Fallback *bool `json:"fallback,omitempty"`
}

// BackendConditionType is a type of condition for a backend. This type should be
// used with a Backend resource Status.Conditions field.
type BackendConditionType string

// BackendConditionReason is a reason for a backend condition.
type BackendConditionReason string

const (
	// BackendConditionAccepted indicates whether the backend has been accepted or
	// rejected by a targeted resource, and why.
	//
	// Possible reasons for this condition to be True are:
	//
	// * "Accepted"
	//
	// Possible reasons for this condition to be False are:
	//
	// * "Invalid"
	//
	BackendConditionAccepted BackendConditionType = "Accepted"

	// BackendReasonAccepted is used with the "Accepted" condition when the backend
	// has been accepted by the targeted resource.
	BackendReasonAccepted BackendConditionReason = "Accepted"

	// BackendReasonInvalid is used with the "Accepted" condition when the backend
	// is syntactically or semantically invalid.
	BackendReasonInvalid BackendConditionReason = "Invalid"
)

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

// BackendList contains a list of Backend resources.
//
// +kubebuilder:object:root=true
type BackendList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Backend `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Backend{}, &BackendList{})
}
