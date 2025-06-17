// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package status

import (
	"sort"
	"strings"

	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

// Route condition reasons for various error scenarios
const (
	// Route configuration related reasons
	RouteReasonUnsupportedSetting  gwapiv1.RouteConditionReason = "UnsupportedValue"
	RouteReasonUnsupportedRefValue gwapiv1.RouteConditionReason = "UnsupportedRefValue"

	// Backend reference related reasons
	RouteReasonInvalidBackendRef     gwapiv1.RouteConditionReason = "InvalidBackendRef"
	RouteReasonInvalidBackendTLS     gwapiv1.RouteConditionReason = "InvalidBackendTLS"
	RouteReasonInvalidBackendFilters gwapiv1.RouteConditionReason = "InvalidBackendFilters"

	// Network configuration related reasons
	RouteReasonPortNotFound           gwapiv1.RouteConditionReason = "PortNotFound"
	RouteReasonPortNotSpecified       gwapiv1.RouteConditionReason = "PortNotSpecified"
	RouteReasonUnsupportedAddressType gwapiv1.RouteConditionReason = "UnsupportedAddressType"
	RouteReasonInvalidAddress         gwapiv1.RouteConditionReason = "InvalidAddress"
	RouteReasonEndpointSliceNotFound  gwapiv1.RouteConditionReason = "EndpointSliceNotFound"
)

// Error is an error interface that represents errors that need to be reflected
// in the status of a Kubernetes resource. It extends the standard error interface
// with a Reason method that returns the specific condition reason.
type Error interface {
	error
	Reason() gwapiv1.RouteConditionReason
	Type() gwapiv1.RouteConditionType
}

// RouteStatusError represents an error that needs to be reflected in the status of an xRoute.
// It wraps an underlying error and provides a specific route condition reason.
type RouteStatusError struct {
	Wrapped              error
	RouteConditionReason gwapiv1.RouteConditionReason
	RouteConditionType   gwapiv1.RouteConditionType
}

// NewRouteStatusError creates a new RouteStatusError with the given wrapped error and route condition reason.
func NewRouteStatusError(wrapped error, reason gwapiv1.RouteConditionReason) *RouteStatusError {
	return &RouteStatusError{
		Wrapped:              wrapped,
		RouteConditionReason: reason,
	}
}

func (s *RouteStatusError) WithType(t gwapiv1.RouteConditionType) *RouteStatusError {
	s.RouteConditionType = t
	return s
}

// Error returns the string representation of the error.
// If Wrapped is nil, it returns the string representation of the RouteConditionReason.
func (s *RouteStatusError) Error() string {
	if s == nil {
		return ""
	}
	if s.Wrapped != nil {
		return s.Wrapped.Error()
	}
	return string(s.RouteConditionReason)
}

// Reason returns the route condition reason associated with this error.
func (s *RouteStatusError) Reason() gwapiv1.RouteConditionReason {
	if s == nil {
		return ""
	}
	return s.RouteConditionReason
}

// Type returns the route condition type associated with this error.
func (s *RouteStatusError) Type() gwapiv1.RouteConditionType {
	// Default to ResolvedRefs because it's the most common type.
	if s == nil {
		return gwapiv1.RouteConditionResolvedRefs
	}
	return s.RouteConditionType
}

// MultiStatusError represents a collection of status errors that occurred during processing.
// It implements the StatusError interface and provides methods to manage multiple errors.
type MultiStatusError struct {
	errs []Error
}

// Empty returns true if there are no errors in the collection.
func (m *MultiStatusError) Empty() bool {
	return m == nil || len(m.errs) == 0
}

// Add appends a new status error to the collection.
// If the error is nil, it is ignored.
func (m *MultiStatusError) Add(err Error) {
	if err == nil {
		return
	}
	if m == nil {
		m = &MultiStatusError{}
	}
	m.errs = append(m.errs, err)
}

// Error returns a string representation of all the wrapped errors.
// If there are no errors, it returns an empty string.
func (m *MultiStatusError) Error() string {
	if m == nil || len(m.errs) == 0 {
		return ""
	}
	if len(m.errs) == 1 {
		return m.errs[0].Error()
	}

	var b strings.Builder
	b.WriteString(Error2ConditionMsg(m.errs[0]))
	for _, err := range m.errs[1:] {
		b.WriteByte('\n')
		b.WriteString(Error2ConditionMsg(err))
	}
	return b.String()
}

// Reason returns a string representation of all unique reasons from the wrapped errors.
// If there are no errors or no reasons, it returns an empty string.
func (m *MultiStatusError) Reason() gwapiv1.RouteConditionReason {
	if m == nil || len(m.errs) == 0 {
		return ""
	}

	reasons := make(map[string]struct{})
	for _, err := range m.errs {
		if reason := err.Reason(); reason != "" {
			reasons[string(reason)] = struct{}{}
		}
	}

	if len(reasons) == 0 {
		return ""
	}

	reasonList := make([]string, 0, len(reasons))
	for reason := range reasons {
		reasonList = append(reasonList, reason)
	}
	sort.Strings(reasonList)
	return gwapiv1.RouteConditionReason(strings.Join(reasonList, ", "))
}

func (m *MultiStatusError) Type() gwapiv1.RouteConditionType {
	if m == nil || len(m.errs) == 0 {
		return gwapiv1.RouteConditionResolvedRefs
	}
	for _, err := range m.errs {
		if err.Type() == gwapiv1.RouteConditionAccepted {
			return gwapiv1.RouteConditionAccepted
		}
	}
	return gwapiv1.RouteConditionResolvedRefs
}

func isAcceptedReason(reason gwapiv1.RouteConditionReason) bool {
	return reason == gwapiv1.RouteReasonNotAllowedByListeners ||
		reason == gwapiv1.RouteReasonNoMatchingListenerHostname ||
		reason == gwapiv1.RouteReasonNoMatchingParent ||
		reason == gwapiv1.RouteReasonUnsupportedValue
}

// ConvertToAcceptedReason converts ResolvedRefs reasons to Accepted condition reasons
// This is used to make the reasons compatible with the Gateway API spec.
// For example, the BackendRefs validation may return a InvalidBackendRef reason for a Mirror filter validation,
// but this error should be reflected in the Accepted condition as UnsupportedValue.
func ConvertToAcceptedReason(reason gwapiv1.RouteConditionReason) gwapiv1.RouteConditionReason {
	if isAcceptedReason(reason) {
		return reason
	}
	// Return UnsupportedValue as the default reason for ResolvedRefs reasons, which is kind of vague and can be used for
	// any other reasons.
	return gwapiv1.RouteReasonUnsupportedValue
}
