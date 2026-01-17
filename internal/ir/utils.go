// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package ir

import (
	"cmp"
	"slices"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

// MapToSlice converts a map[string]string to a slice of MapEntry, sorted by key.
func MapToSlice(m map[string]string) []MapEntry {
	if len(m) == 0 {
		return nil
	}
	res := make([]MapEntry, 0, len(m))
	for k, v := range m {
		res = append(res, MapEntry{Key: k, Value: v})
	}
	slices.SortFunc(res, func(a, b MapEntry) int {
		return cmp.Compare(a.Key, b.Key)
	})
	return res
}

// CustomTagMapToSlice converts a map[string]CustomTag to a slice of CustomTagMapEntry, sorted by key.
func CustomTagMapToSlice(m map[string]egv1a1.CustomTag) []CustomTagMapEntry {
	if len(m) == 0 {
		return nil
	}
	res := make([]CustomTagMapEntry, 0, len(m))
	for k, v := range m {
		res = append(res, CustomTagMapEntry{Key: k, Value: v})
	}
	slices.SortFunc(res, func(a, b CustomTagMapEntry) int {
		return cmp.Compare(a.Key, b.Key)
	})
	return res
}
