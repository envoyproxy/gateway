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
// +kubebuilder:resource:categories=envoy-gateway,shortName=hf
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
	URLRewrite HTTPURLRewriteFilter `json:"urlRewrite,omitempty"`
}

// HTTPURLRewriteFilter define rewrites of HTTP URL components such as path and host
type HTTPURLRewriteFilter struct {
	// Path defines a path rewrite.
	//
	// +optional
	Path *HTTPPathModifier `json:"path,omitempty"`
}

// HTTPPathModifierType defines the type of path redirect or rewrite.
type HTTPPathModifierType string

const (
	// RegexHTTPPathModifier This type of modifier indicates that the portions of the path that match the specified
	//  regex would be substituted with the specified substitution value
	// https://www.envoyproxy.io/docs/envoy/latest/api-v3/type/matcher/v3/regex.proto#type-matcher-v3-regexmatchandsubstitute
	RegexHTTPPathModifier HTTPPathModifierType = "ReplaceRegexMatch"

	// TemplateHTTPPathModifier This type of modifier indicates that the portions of the path that match the specified
	// pattern would be rewritten according to the specified template
	// https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/path/rewrite/uri_template/v3/uri_template_rewrite.proto#extension-envoy-path-rewrite-uri-template-uri-template-rewriter
	TemplateHTTPPathModifier HTTPPathModifierType = "ReplaceTemplate"
)

type ReplaceRegexMatch struct {
	// Regex matches a regular expression against the value of the HTTP Path.The regex string must
	// adhere to the syntax documented in https://github.com/google/re2/wiki/Syntax.
	Regex string `json:"regex"`
	// Replacement is an expression that replaces the matched portion.The expression may include numbered
	// capture groups that adhere to syntax documented in https://github.com/google/re2/wiki/Syntax.
	Replacement string `json:"replacement"`
}

type HTTPPathModifier struct {
	// +kubebuilder:validation:Enum=RegexHTTPPathModifier
	Type HTTPPathModifierType `json:"type"`

	ReplaceRegexMatch *ReplaceRegexMatch `json:"replaceRegexMatch,omitempty"`
}

//+kubebuilder:object:root=true

// HTTPFilterList contains a list of HTTPRouteFilter resources.
type HTTPFilterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HTTPRouteFilter `json:"items"`
}

func init() {
	SchemeBuilder.Register(&HTTPRouteFilter{}, &HTTPFilterList{})
}