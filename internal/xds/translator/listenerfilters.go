// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"slices"
	"strings"

	listenerv3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
)

type OrderedListenerFilter struct {
	filter        *listenerv3.ListenerFilter
	priority      int
	originalIndex int
}

type OrderedListenerFilters []*OrderedListenerFilter

// newOrderedListenerFilter gives each listener filter a rational order.
// This is needed because the order of the filters is important.
// For example, the proxy_protocol filter must come before other listener
// filters such as the tls_inpsector filter, as PPv2 injects data
// into the bytestream and may cause other listeners to fail if the
// proxy_protocol filter doesn't consume this injected data and
// restore the original bytestream.

func newOrderedListenerFilter(filter *listenerv3.ListenerFilter, originalIndex int) *OrderedListenerFilter {
	priority := 50

	switch {
	case isListenerFilterType(filter, wellknown.ProxyProtocol):
		priority = 1
	}

	return &OrderedListenerFilter{
		filter,
		priority,
		originalIndex,
	}
}

func isListenerFilterType(filter *listenerv3.ListenerFilter, filterTypeStr string) bool {
	return strings.HasPrefix(filter.Name, filterTypeStr)
}

func compareOrderedListenerFilters(filterA *OrderedListenerFilter, filterB *OrderedListenerFilter) int {
	diff := filterA.priority - filterB.priority

	if diff == 0 {
		diff = filterA.originalIndex - filterB.originalIndex
	}

	return diff
}

func sortListenerFilters(filters []*listenerv3.ListenerFilter) []*listenerv3.ListenerFilter {
	// Assign ordering weights to each filter
	filterLen := len(filters)
	orderedFilters := make(OrderedListenerFilters, filterLen)
	for idx, filter := range filters {
		orderedFilters[idx] = newOrderedListenerFilter(filter, idx)
	}

	// Sort filters
	slices.SortStableFunc(orderedFilters, compareOrderedListenerFilters)

	// Make result slice, populate it, and return
	resultFilters := make([]*listenerv3.ListenerFilter, filterLen)
	for idx, orderedFilter := range orderedFilters {
		resultFilters[idx] = orderedFilter.filter
	}

	return resultFilters
}
