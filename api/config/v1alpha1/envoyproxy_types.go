package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// KindEnvoyProxy is the name of the EnvoyProxy kind.
	KindEnvoyProxy = "EnvoyProxy"
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// EnvoyProxy is the Schema for the envoyproxies API
type EnvoyProxy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   EnvoyProxySpec   `json:"spec,omitempty"`
	Status EnvoyProxyStatus `json:"status,omitempty"`
}

// EnvoyProxySpec defines the desired state of EnvoyProxy.
type EnvoyProxySpec struct {
	// XDSServer defines the xDS Server configuration of EnvoyProxy.
	// If unspecified, default configuration parameters will be used.
	//
	// +optional
	XDSServer *XDSServer `json:"xdsServer,omitempty"`
}

// XDSServer defines the desired xDS Server configuration of EnvoyProxy.
type XDSServer struct {
	// Address defines the IPv4 address or DNS name of the server used for xDS
	// configuration. If unspecified, defaults to the DNS name "envoy-gateway".
	// Address must be a valid IPv4 host address or RFC 1123 Label Name.
	//
	// +optional
	Address string `json:"address,omitempty"`
}

// EnvoyProxyStatus defines the observed state of EnvoyProxy
type EnvoyProxyStatus struct {
	// INSERT ADDITIONAL STATUS FIELDS - define observed state of cluster.
	// Important: Run "make" to regenerate code after modifying this file.
}

//+kubebuilder:object:root=true

// EnvoyProxyList contains a list of EnvoyProxy
type EnvoyProxyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []EnvoyProxy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&EnvoyProxy{}, &EnvoyProxyList{})
}
