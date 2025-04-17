// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package status

import (
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
	RouteReasonPortNotFound       gwapiv1.RouteConditionReason = "PortNotFound"
	RouteReasonPortNotSpecified   gwapiv1.RouteConditionReason = "PortNotSpecified"
	RouteReasonInvalidAddressType gwapiv1.RouteConditionReason = "InvalidAddressType"
	RouteReasonInvalidAddress     gwapiv1.RouteConditionReason = "InvalidAddress"
)

// Error is an error interface that represents errors that need to be reflected
// in the status of a Kubernetes resource. It extends the standard error interface
// with a Reason method that returns the specific condition reason.
type Error interface {
	error
	Reason() gwapiv1.RouteConditionReason
}

// RouteStatusError represents an error that needs to be reflected in the status of an xRoute.
// It wraps an underlying error and provides a specific route condition reason.
type RouteStatusError struct {
	Wrapped              error
	RouteConditionReason gwapiv1.RouteConditionReason
}

// NewRouteStatusError creates a new RouteStatusError with the given wrapped error and route condition reason.
func NewRouteStatusError(wrapped error, reason gwapiv1.RouteConditionReason) *RouteStatusError {
	return &RouteStatusError{
		Wrapped:              wrapped,
		RouteConditionReason: reason,
	}
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
	return gwapiv1.RouteConditionReason(strings.Join(reasonList, ", "))
}
