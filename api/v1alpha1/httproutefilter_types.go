// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// KindHTTPRouteFilter is the name of the HTTPRouteFilter kind.
	KindHTTPRouteFilter = "HTTPRouteFilter"
)

// +kubebuilder:object:root=true
// +kubebuilder:resource:categories=envoy-gateway,shortName=hrf
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// HTTPRouteFilter is a custom Envoy Gateway HTTPRouteFilter which provides extended
// traffic processing options such as path regex rewrite, direct response and more.
type HTTPRouteFilter struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the desired state of HTTPRouteFilter.
	Spec HTTPRouteFilterSpec `json:"spec"`
}

// HTTPRouteFilterSpec defines the desired state of HTTPRouteFilter.
// +union
type HTTPRouteFilterSpec struct {
	// +optional
	URLRewrite *HTTPURLRewriteFilter `json:"urlRewrite,omitempty"`
	// +optional
	DirectResponse *HTTPDirectResponseFilter `json:"directResponse,omitempty"`
}

// HTTPURLRewriteFilter define rewrites of HTTP URL components such as path and host
type HTTPURLRewriteFilter struct {
	// Hostname is the value to be used to replace the Host header value during
	// forwarding.
	//
	// +optional
	// +notImplementedHide
	Hostname *HTTPHostnameModifier `json:"hostname,omitempty"`
	// Path defines a path rewrite.
	//
	// +optional
	Path *HTTPPathModifier `json:"path,omitempty"`
}

// HTTPDirectResponseFilter defines the configuration to return a fixed response.
type HTTPDirectResponseFilter struct {
	// Content Type of the response. This will be set in the Content-Type header.
	//
	// +optional
	ContentType *string `json:"contentType,omitempty"`

	// Body of the Response
	//
	// +optional
	Body *CustomResponseBody `json:"body,omitempty"`

	// Status Code of the HTTP response
	// If unset, defaults to 200.
	// +optional
	StatusCode *int `json:"statusCode,omitempty"`
}

// HTTPPathModifierType defines the type of path redirect or rewrite.
type HTTPPathModifierType string

const (
	// RegexHTTPPathModifier This type of modifier indicates that the portions of the path that match the specified
	//  regex would be substituted with the specified substitution value
	// https://www.envoyproxy.io/docs/envoy/latest/api-v3/type/matcher/v3/regex.proto#type-matcher-v3-regexmatchandsubstitute
	RegexHTTPPathModifier HTTPPathModifierType = "ReplaceRegexMatch"
)

// HTTPPathModifierType defines the type of Hostname rewrite.
type HTTPHostnameModifierType string

const (
	// HeaderHTTPHostnameModifier indicates that the Host header value would be replaced with the value of the header specified in setFromHeader.
	// https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/route/v3/route_components.proto#envoy-v3-api-field-config-route-v3-routeaction-host-rewrite-header
	HeaderHTTPHostnameModifier HTTPHostnameModifierType = "SetFromHeader"
	// BackendHTTPHostnameModifier indicates that the Host header value would be replaced by the DNS name of the backend if it exists.
	// https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/route/v3/route_components.proto#envoy-v3-api-field-config-route-v3-routeaction-auto-host-rewrite
	BackendHTTPHostnameModifier HTTPHostnameModifierType = "SetFromBackend"
)

type ReplaceRegexMatch struct {
	// Pattern matches a regular expression against the value of the HTTP Path.The regex string must
	// adhere to the syntax documented in https://github.com/google/re2/wiki/Syntax.
	// +kubebuilder:validation:MinLength=1
	Pattern string `json:"pattern"`
	// Substitution is an expression that replaces the matched portion.The expression may include numbered
	// capture groups that adhere to syntax documented in https://github.com/google/re2/wiki/Syntax.
	Substitution string `json:"substitution"`
}

// +kubebuilder:validation:XValidation:rule="self.type == 'ReplaceRegexMatch' ? has(self.replaceRegexMatch) : !has(self.replaceRegexMatch)",message="If HTTPPathModifier type is ReplaceRegexMatch, replaceRegexMatch field needs to be set."
type HTTPPathModifier struct {
	// +kubebuilder:validation:Enum=ReplaceRegexMatch
	// +kubebuilder:validation:Required
	Type HTTPPathModifierType `json:"type"`
	// ReplaceRegexMatch defines a path regex rewrite. The path portions matched by the regex pattern are replaced by the defined substitution.
	// https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/route/v3/route_components.proto#envoy-v3-api-field-config-route-v3-routeaction-regex-rewrite
	// Some examples:
	// (1) replaceRegexMatch:
	//       pattern: ^/service/([^/]+)(/.*)$
	//       substitution: \2/instance/\1
	//     Would transform /service/foo/v1/api into /v1/api/instance/foo.
	// (2) replaceRegexMatch:
	//       pattern: one
	//       substitution: two
	//     Would transform /xxx/one/yyy/one/zzz into /xxx/two/yyy/two/zzz.
	// (3) replaceRegexMatch:
	//       pattern: ^(.*?)one(.*)$
	//       substitution: \1two\2
	//     Would transform /xxx/one/yyy/one/zzz into /xxx/two/yyy/one/zzz.
	// (3) replaceRegexMatch:
	//       pattern: (?i)/xxx/
	//       substitution: /yyy/
	//     Would transform path /aaa/XxX/bbb into /aaa/yyy/bbb (case-insensitive).
	// +optional
	ReplaceRegexMatch *ReplaceRegexMatch `json:"replaceRegexMatch,omitempty"`
}

// +kubebuilder:validation:XValidation:message="setFromHeader must be nil if the type is not SetFromHeader",rule="!(has(self.setFromHeader) && self.type != 'SetFromHeader')"
// +kubebuilder:validation:XValidation:message="setFromHeader must be specified for SetFromHeader type",rule="!(!has(self.setFromHeader) && self.type == 'SetFromHeader')"
type HTTPHostnameModifier struct {
	// +kubebuilder:validation:Enum=SetFromHeader;SetFromBackend
	// +kubebuilder:validation:Required
	Type HTTPHostnameModifierType `json:"type"`

	// SetFromHeader is the name of the header whose value would be used to rewrite the Host header
	// +optional
	SetFromHeader *string `json:"setFromHeader,omitempty"`
}

//+kubebuilder:object:root=true

// HTTPRouteFilterList contains a list of HTTPRouteFilter resources.
type HTTPRouteFilterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HTTPRouteFilter `json:"items"`
}

func init() {
	SchemeBuilder.Register(&HTTPRouteFilter{}, &HTTPRouteFilterList{})
}
