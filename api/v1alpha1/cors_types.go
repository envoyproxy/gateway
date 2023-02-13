package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

const (
	// KindCorsFilterKind is the name of the CorsFilter kind.
	KindCorsFilter = "CorsFilter"
)

//+kubebuilder:object:root=true

type CorsFilter struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the desired state of the CorsFilter type.
	Spec CorsFilterSpec `json:"spec"`

	// Note: The status sub-resource has been excluded but may be added in the future.
}

// CorsFilterSpec defines the desired state of the CorsFilter type.
// +union

type CorsFilterSpec struct {
	// AllowOriginStringMatch defines a list of origins that will be allowed to
	// do CORS requests. An origin is allowed if it matches either
	// an item in this list or the special "*" wildcard.
	//
	// +optional
	AllowOriginStringMatch []StringMatch `json:"allow_origin_string_match,omitempty"`
	// AllowMethods defines a list of HTTP methods that will be allowed to do
	// CORS requests. An HTTP method is allowed if it matches either
	// an item in this list or the special "*" wildcard.
	//
	// +optional
	AllowMethods string `json:"allow_methods,omitempty"`
	// AllowHeaders defines a list of HTTP headers that will be allowed to do
	// CORS requests. An HTTP header is allowed if it matches either
	// an item in this list or the special "*" wildcard.
	//
	// +optional
	AllowHeaders string `json:"allow_headers,omitempty"`
	// MaxAge defines the duration that the results of a preflight request
	// can be cached.
	//
	// +optional
	MaxAge string `json:"max_age,omitempty"`
	// ExposeHeaders defines a list of HTTP headers that will be exposed as
	// part of the CORS response.
	//
	// +optional
	ExposeHeaders string `json:"expose_headers,omitempty"`
}

// StringMatch defines a string match.
type StringMatch struct {
	// Prefix defines a string prefix match.
	//
	// +optional
	Prefix string `json:"prefix,omitempty"`
}

type CorsFilterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CorsFilter `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CorsFilter{}, &CorsFilterList{})
}
