// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/api/resource"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

// BandwidthLimitDirection specifies which direction of traffic the bandwidth limit applies to.
//
// +kubebuilder:validation:Enum=Request;Response;Both
type BandwidthLimitDirection string

const (
	// BandwidthLimitDirectionRequest limits traffic from the client to the upstream.
	BandwidthLimitDirectionRequest BandwidthLimitDirection = "Request"

	// BandwidthLimitDirectionResponse limits traffic from the upstream to the client.
	BandwidthLimitDirectionResponse BandwidthLimitDirection = "Response"

	// BandwidthLimitDirectionBoth limits traffic in both directions.
	BandwidthLimitDirectionBoth BandwidthLimitDirection = "Both"
)

// BandwidthLimitSpec defines the desired state of BandwidthLimit.
//
// +kubebuilder:validation:XValidation:rule="!has(self.fillInterval) || (duration(self.fillInterval) >= duration('20ms'))",message="fillInterval must be at least 20ms"
// +kubebuilder:validation:XValidation:rule="!has(self.responseTrailers) || self.direction == 'Response' || self.direction == 'Both'",message="responseTrailers can only be specified when direction is Response or Both"
type BandwidthLimitSpec struct {
	// Limit specifies the bandwidth limit as a bytes-per-second throughput rate.
	//
	// +kubebuilder:validation:XIntOrString
	// +kubebuilder:validation:Pattern="^[1-9]+[0-9]*([EPTGMK]i|[EPTGMk])?$"
	Limit resource.Quantity `json:"limit"`

	// Direction controls which traffic direction the bandwidth limit applies to.
	// Request limits traffic from the client to the upstream (ingress).
	// Response limits traffic from the upstream to the client (egress).
	// Both limits traffic in both directions.
	//
	// +kubebuilder:default=Both
	Direction BandwidthLimitDirection `json:"direction"`

	// FillInterval is the token bucket refill interval.
	// Minimum allowed value is 20ms. Defaults to 50ms if not specified.
	//
	// +optional
	FillInterval *gwapiv1.Duration `json:"fillInterval,omitempty"`

	// BandwidthLimitResponseTrailers configures the trailer headers appended to responses
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
	// Only effective when Direction is Response or Both.
	//
	// +optional
	Prefix *string `json:"prefix,omitempty"`
}
