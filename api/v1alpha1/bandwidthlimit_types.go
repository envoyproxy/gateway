// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/api/resource"
)

// BandwidthLimitSpec defines the desired state of BandwidthLimit.
//
// +kubebuilder:validation:XValidation:rule="has(self.request) || has(self.response)",message="at least one of request or response must be specified"
type BandwidthLimitSpec struct {
	// Request configures the bandwidth limit for client-to-upstream (ingress) traffic.
	//
	// +optional
	Request *BandwidthLimitRequestConfig `json:"request,omitempty"`

	// Response configures the bandwidth limit for upstream-to-client (egress) traffic.
	//
	// +optional
	Response *BandwidthLimitResponseConfig `json:"response,omitempty"`
}

// BandwidthLimitRequestConfig defines the bandwidth limit configuration for the request direction.
type BandwidthLimitRequestConfig struct {
	// Limit specifies the bandwidth limit as a bytes-per-second throughput rate.
	//
	// +kubebuilder:validation:XIntOrString
	// +kubebuilder:validation:Pattern="^[1-9]+[0-9]*([EPTGMK]i|[EPTGMk])?$"
	Limit resource.Quantity `json:"limit"`
}

// BandwidthLimitResponseConfig defines the bandwidth limit configuration for the response direction.
type BandwidthLimitResponseConfig struct {
	// Limit specifies the bandwidth limit as a bytes-per-second throughput rate.
	//
	// +kubebuilder:validation:XIntOrString
	// +kubebuilder:validation:Pattern="^[1-9]+[0-9]*([EPTGMK]i|[EPTGMk])?$"
	Limit resource.Quantity `json:"limit"`

	// ResponseTrailers con figures the trailer headers appended to responses
	// when bandwidth limiting introduces delays.
	//
	// +optional
	ResponseTrailers *BandwidthLimitResponseTrailers `json:"responseTrailers,omitempty"`
}

type BandwidthLimitResponseTrailers struct {
	// Prefix is prepended to each trailer header name with delay metrics.
	// For example, setting "x-eg" produces trailers such as "x-eg-bandwidth-request-delay-ms".
	//
	// The following four trailers can be added:
	// "bandwidth-request-delay-ms" is delay time in milliseconds it took for the request stream transfer
	// including request body transfer time and the time added by the filter.
	// "bandwidth-response-delay-ms" is delay time in milliseconds it took for the response stream transfer
	// including response body transfer time and the time added by the filter.
	// "bandwidth-request-filter-delay-ms" is delay time in milliseconds in request stream transfer added by the filter.
	// "bandwidth-response-filter-delay-ms" is delay time in milliseconds that added by the filter.
	//
	// +optional
	Prefix *string `json:"prefix,omitempty"`
}
