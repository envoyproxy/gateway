// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a3 "sigs.k8s.io/gateway-api/apis/v1alpha3"
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

	// IP defines an IP endpoint. Supports both IPv4 and IPv6 addresses.
	//
	// +optional
	IP *IPEndpoint `json:"ip,omitempty"`

	// Unix defines the unix domain socket endpoint
	//
	// +optional
	Unix *UnixSocket `json:"unix,omitempty"`

	// Zone defines the service zone of the backend endpoint.
	//
	// +optional
	Zone *string `json:"zone,omitempty"`
}

// IPEndpoint describes TCP/UDP socket address, corresponding to Envoy's Socket Address
// https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/core/v3/address.proto#config-core-v3-socketaddress
type IPEndpoint struct {
	// Address defines the IP address of the backend endpoint.
	// Supports both IPv4 and IPv6 addresses.
	//
	// +kubebuilder:validation:MinLength=3
	// +kubebuilder:validation:MaxLength=45
	// +kubebuilder:validation:Pattern=`^((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$|^(([0-9a-fA-F]{1,4}:){1,7}[0-9a-fA-F]{1,4}|::|(([0-9a-fA-F]{1,4}:){0,5})?(:[0-9a-fA-F]{1,4}){1,2})$`
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
	// +kubebuilder:validation:Pattern=`^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$`
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
	// The path length must not exceed 108 characters.
	//
	// +kubebuilder:validation:XValidation:rule="size(self) <= 108",message="unix domain socket path must not exceed 108 characters"
	Path string `json:"path"`
}

// BackendSpec describes the desired state of BackendSpec.
// +kubebuilder:validation:XValidation:rule="self.type != 'DynamicResolver' || !has(self.endpoints)",message="DynamicResolver type cannot have endpoints specified"
type BackendSpec struct {
	// Type defines the type of the backend. Defaults to "Endpoints"
	//
	// +kubebuilder:validation:Enum=Endpoints;DynamicResolver
	// +kubebuilder:default=Endpoints
	// +optional
	Type *BackendType `json:"type,omitempty"`

	// Endpoints defines the endpoints to be used when connecting to the backend.
	//
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=64
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
	Fallback *bool `json:"fallback,omitempty"`

	// TLS defines the TLS settings for the backend.
	// TLS.CACertificateRefs and TLS.WellKnownCACertificates can only be specified for DynamicResolver backends.
	// TLS.InsecureSkipVerify can be specified for any Backends
	//
	// +optional
	TLS *BackendTLSSettings `json:"tls,omitempty"`
}

// BackendTLSSettings holds the TLS settings for the backend.
// +kubebuilder:validation:XValidation:message="must not contain both CACertificateRefs and WellKnownCACertificates",rule="!(has(self.caCertificateRefs) && size(self.caCertificateRefs) > 0 && has(self.wellKnownCACertificates) && self.wellKnownCACertificates != \"\")"
// +kubebuilder:validation:XValidation:message="must not contain either CACertificateRefs or WellKnownCACertificates when InsecureSkipVerify is enabled",rule="!((has(self.insecureSkipVerify) && self.insecureSkipVerify) && ((has(self.caCertificateRefs) && size(self.caCertificateRefs) > 0) || (has(self.wellKnownCACertificates) && self.wellKnownCACertificates != \"\")))"
type BackendTLSSettings struct {
	// CACertificateRefs contains one or more references to Kubernetes objects that
	// contain TLS certificates of the Certificate Authorities that can be used
	// as a trust anchor to validate the certificates presented by the backend.
	//
	// A single reference to a Kubernetes ConfigMap or a Kubernetes Secret,
	// with the CA certificate in a key named `ca.crt` is currently supported.
	//
	// If CACertificateRefs is empty or unspecified, then WellKnownCACertificates must be
	// specified. Only one of CACertificateRefs or WellKnownCACertificates may be specified,
	// not both.
	//
	// Only used for DynamicResolver backends.
	//
	// +kubebuilder:validation:MaxItems=8
	// +optional
	CACertificateRefs []gwapiv1.LocalObjectReference `json:"caCertificateRefs,omitempty"`

	// WellKnownCACertificates specifies whether system CA certificates may be used in
	// the TLS handshake between the gateway and backend pod.
	//
	// If WellKnownCACertificates is unspecified or empty (""), then CACertificateRefs
	// must be specified with at least one entry for a valid configuration. Only one of
	// CACertificateRefs or WellKnownCACertificates may be specified, not both.
	//
	// Only used for DynamicResolver backends.
	//
	// +optional
	WellKnownCACertificates *gwapiv1a3.WellKnownCACertificatesType `json:"wellKnownCACertificates,omitempty"`

	// InsecureSkipVerify indicates whether the upstream's certificate verification
	// should be skipped. Defaults to "false".
	//
	// +kubebuilder:default=false
	// +optional
	InsecureSkipVerify *bool `json:"insecureSkipVerify,omitempty"`
}

// BackendType defines the type of the Backend.
type BackendType string

const (
	// BackendTypeEndpoints defines the type of the backend as Endpoints.
	BackendTypeEndpoints BackendType = "Endpoints"
	// BackendTypeDynamicResolver defines the type of the backend as DynamicResolver.
	//
	// When a backend is of type DynamicResolver, the Envoy will resolve the upstream
	// ip address and port from the host header of the incoming request. If the ip address
	// is directly set in the host header, the Envoy will use the ip address and port as the
	// upstream address. If the hostname is set in the host header, the Envoy will resolve the
	// ip address and port from the hostname using the DNS resolver.
	BackendTypeDynamicResolver BackendType = "DynamicResolver"
)

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
