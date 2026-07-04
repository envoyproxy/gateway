// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package utils

// CopyStringMap copies all key/value pairs from src into dst, converting the
// string-based key and value types of src to plain strings. It mirrors
// maps.Copy for maps whose key and value types have an underlying string type
// (e.g. Gateway API's LabelKey/LabelValue), which maps.Copy cannot handle
// because it requires identical key/value types.
func CopyStringMap[K, V ~string](dst map[string]string, src map[K]V) {
	for k, v := range src {
		dst[string(k)] = string(v)
	}
}
