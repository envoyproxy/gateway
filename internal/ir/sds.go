// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package ir

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

// SDSClusterNameFromURL returns the canonical xDS cluster name for an SDS URL.
func SDSClusterNameFromURL(url string) string {
	hash := sha256.Sum256([]byte(url))
	if strings.HasPrefix(url, "/") {
		const maxReadablePrefixLength = 48

		hashSuffix := hex.EncodeToString(hash[:16])
		readablePrefix := strings.Trim(strings.ReplaceAll(url, "/", "_"), "_")
		if len(readablePrefix) > maxReadablePrefixLength {
			readablePrefix = readablePrefix[:maxReadablePrefixLength]
		}
		if readablePrefix != "" {
			return fmt.Sprintf("sds_%s_%s", readablePrefix, hashSuffix)
		}
	}

	return fmt.Sprintf("sds_%s", hex.EncodeToString(hash[:8]))
}
