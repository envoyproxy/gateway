// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package ir

import (
	"sort"

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
	sort.Slice(res, func(i, j int) bool {
		return res[i].Key < res[j].Key
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
	sort.Slice(res, func(i, j int) bool {
		return res[i].Key < res[j].Key
	})
	return res
}

// SliceToMap converts a slice of MapEntry to a map[string]string.
func SliceToMap(s []MapEntry) map[string]string {
	if len(s) == 0 {
		return nil
	}
	res := make(map[string]string, len(s))
	for _, entry := range s {
		res[entry.Key] = entry.Value
	}
	return res
}
