// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

// This file contains code derived from Contour,
// https://github.com/projectcontour/contour
// from the source file
// https://github.com/projectcontour/contour/blob/main/internal/status/gatewayclassconditions.go
// and is provided here subject to the following:
// Copyright Project Contour Authors
// SPDX-License-Identifier: Apache-2.0

package gatewayapi

import (
	"strings"
	"unsafe"

	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/envoyproxy/gateway/internal/gatewayapi/status"
)

const (
	RouteReasonUnsupportedSetting    gwapiv1.RouteConditionReason = "UnsupportedValue"
	RouteReasonInvalidBackendRef     gwapiv1.RouteConditionReason = "InvalidBackendRef"
	RouteReasonInvalidBackendTLS     gwapiv1.RouteConditionReason = "InvalidBackendTLS"
	RouteReasonInvalidPortNotFound   gwapiv1.RouteConditionReason = "PortNotFound"
	RouteReasonPortNotSpecified      gwapiv1.RouteConditionReason = "PortNotSpecified"
	RouteReasonInvalidAddressType    gwapiv1.RouteConditionReason = "InvalidAddressType"
	RouteReasonInvalidBackendFilters gwapiv1.RouteConditionReason = "InvalidBackendFilters"
	RouteReasonUnsupportedRefValue   gwapiv1.RouteConditionReason = "UnsupportedRefValue"
)

// StatusError is an error that needs to be reflected in the status of the resource.
type StatusError interface {
	error
	Reason() gwapiv1.RouteConditionReason
}

// RouteStatusError is an error that needs to be reflected in the status of xRoute.
type RouteStatusError struct {
	Wrapped              error
	RouteConditionReason gwapiv1.RouteConditionReason
}

func (s *RouteStatusError) Error() string {
	if s.Wrapped != nil {
		return s.Wrapped.Error()
	}
	return string(s.RouteConditionReason)
}

func (s *RouteStatusError) Reason() gwapiv1.RouteConditionReason {
	return s.RouteConditionReason
}

// MultiStatusError is an error that contains multiple status errors.
type MultiStatusError struct {
	errs []StatusError
}

func (m *MultiStatusError) Empty() bool {
	return len(m.errs) == 0
}

func (m *MultiStatusError) Add(err StatusError) {
	if err == nil {
		return
	}
	m.errs = append(m.errs, err)
}

// Error returns a string representation of the error of all the wrapped errors.
func (m *MultiStatusError) Error() string {
	if len(m.errs) > 0 {
		if len(m.errs) == 1 {
			return m.errs[0].Error()
		}

		b := []byte(status.Error2ConditionMsg(m.errs[0]))
		for _, err := range m.errs[1:] {
			b = append(b, '\n')
			b = append(b, status.Error2ConditionMsg(err)...)
		}
		// At this point, b has at least one byte '\n'.
		return unsafe.String(&b[0], len(b))
	}
	return ""
}

// Reason returns a string representation of the reason of all the wrapped errors.
func (m *MultiStatusError) Reason() gwapiv1.RouteConditionReason {
	reasons := make([]string, 0)
	for _, err := range m.errs {
		if err.Reason() == "" {
			continue
		}
		existing := false
		for _, existingReason := range reasons {
			if existingReason == string(err.Reason()) {
				existing = true
				break
			}
		}
		if !existing {
			reasons = append(reasons, string(err.Reason()))
		}
	}

	if len(reasons) == 0 {
		return ""
	}

	return gwapiv1.RouteConditionReason(strings.Join(reasons, ", "))
}
