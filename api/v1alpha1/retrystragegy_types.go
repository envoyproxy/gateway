// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RetryStrategy defines the retry strategy to be applied.
type RetryStrategy struct {
	// NumRetries is the number of retries to be attempted. Defaults to 0. If nonzero, maxBudget is ignored.
	//
	// +optional
	MaxRetries *int `json:"maxRetries,omitempty"`

	// MaxBudget is specifies the limit on concurrent retries as a percentage of the sum of active requests and active pending requests.
	// For example, if there are 100 active requests and the MaxBudget is set to 25, there may be 25 active retries.
	// This parameter is optional. Defaults to 20%.
	//
	// +optional
	MaxBudget *int `json:"maxBudget,omitempty"`

	// Minconcurrent specifies the minimum retry concurrency allowed for the retry budget. The limit on the number of active retries may never go below this number.
	// This parameter is optional. Defaults to 3.
	//
	// +optional
	MinConcurrent *int `json:"minConcurrent,omitempty"`

	// MaxParallel is the maximum number of parallel retries. If not specified, the default is 3. Priority lower than retry budget.
	//
	// +optional
	MaxParallel *int `json:"maxParallel,omitempty"`

	// RetryOn specifies the retry trigger condition.
	//
	// +optional
	RetryOn *RetryOn `json:"retryOn,omitempty"`

	// PerRetry is the retry policy to be applied per retry attempt.
	//
	// +optional
	PerRetry *PerRetryPolicy `json:"perRetry,omitempty"`
}

type RetryOn struct {
	// Triggers specifies the retry trigger condition(Http/Grpc).
	//
	// +optional
	Triggers []TriggerEnum `json:"triggers,omitempty"`

	// HttpStatusCodes specifies the http status codes to be retried.
	//
	// +optional
	HTTPStatusCodes []int `json:"httpStatusCodes,omitempty"`
}

// TriggerEnum specifies the conditions that trigger retries.
type TriggerEnum string

const (
	// HTTP events.
	// For additional details, see https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/router_filter#x-envoy-retry-on

	// The upstream server responds with any 5xx response code, or does not respond at all (disconnect/reset/read timeout).
	// Includes connect-failure and refused-stream.
	Error5XX TriggerEnum = "5xx"
	// The response is a gateway error (502,503 or 504).
	GatewayError TriggerEnum = "gateway-error"
	// The upstream server does not respond at all (disconnect/reset/read timeout.)
	DisconnectRest TriggerEnum = "disconnect-reset"
	// Connection failure to the upstream server (connect timeout, etc.). (Included in *5xx*)
	ConnectFailure TriggerEnum = "connect-failure"
	// The upstream server responds with a retriable 4xx response code.
	// Currently, the only response code in this category is 409.
	Retriable4XX TriggerEnum = "retriable-4xx"
	// The upstream server resets the stream with a REFUSED_STREAM error code.
	RefusedStream TriggerEnum = "refused-stream"
	// The upstream server responds with any response code matching one defined in the RetriableStatusCodes.
	RetriableStatusCodes TriggerEnum = "retriable-status-codes"

	// GRPC events, currently only supported for gRPC status codes in response headers.
	// For additional details, see https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/router_filter#x-envoy-retry-grpc-on

	// The gRPC status code in the response headers is “cancelled”.
	Cancelled TriggerEnum = "cancelled"
	// The gRPC status code in the response headers is “deadline-exceeded”.
	DeadlineExceeded TriggerEnum = "deadline-exceeded"
	// The gRPC status code in the response headers is “internal”.
	Internal TriggerEnum = "internal"
	// The gRPC status code in the response headers is “resource-exhausted”.
	ResourceExhausted TriggerEnum = "resource-exhausted"
	// The gRPC status code in the response headers is “unavailable”.
	Unavailable TriggerEnum = "unavailable"
)

type PerRetryPolicy struct {
	// Timeout is the timeout per retry attempt.
	//
	// +optional
	// +kubebuilder:validation:Format=duration
	Timeout *metav1.Duration `json:"timeout,omitempty"`
	// IdleTimeout is the upstream idle timeout per retry attempt.This parameter is optional and if absent there is no per try idle timeout.
	//
	// +optional
	// +kubebuilder:validation:Format=duration
	IdleTimeout *metav1.Duration `json:"idleTimeout,omitempty"`
	// Backoff is the backoff policy to be applied per retry attempt. gateway uses a fully jittered exponential
	// back-off algorithm for retries. For additional details,
	// see https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/router_filter#config-http-filters-router-x-envoy-max-retries
	//
	// +optional
	BackOff *BackOffPolicy `json:"backOff,omitempty"`
}

type BackOffPolicy struct {
	// BaseInterval is the base interval between retries.
	//
	// +kubebuilder:validation:Format=duration
	BaseInterval *metav1.Duration `json:"baseInterval,omitempty"`
	// MaxInterval is the maximum interval between retries. This parameter is optional, but must be greater than or equal to the base_interval if set.
	// The default is 10 times the base_interval
	//
	// +optional
	// +kubebuilder:validation:Format=duration
	MaxInterval *metav1.Duration `json:"maxInterval,omitempty"`
	// we can add rate limited based backoff config here if we want to.
}
