// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package naming

import (
	"crypto/sha256"
	"encoding/hex"
	"unicode/utf8"

	"k8s.io/apimachinery/pkg/types"
)

func ServiceName(nn types.NamespacedName) string {
	return nn.Name + "." + nn.Namespace
}

// TruncateToBytes shortens s to at most maxBytes, cutting whole runes off the end so the
// result stays valid UTF-8.
func TruncateToBytes(s string, maxBytes int) string {
	for len(s) > maxBytes {
		_, size := utf8.DecodeLastRuneInString(s)
		s = s[:len(s)-size]
	}
	return s
}

// HashPrefix returns the hex encoding of the first n bytes of sha256(s).
func HashPrefix(s string, n int) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:n])
}

// Bounded returns name unchanged when it fits in maxBytes, and otherwise a truncated head
// joined to a hash of the full name, so names sharing a prefix stay distinct. Callers whose
// readable part is lossy on its own must hash unconditionally instead.
func Bounded(name string, maxBytes int) string {
	if len(name) <= maxBytes {
		return name
	}
	const hashBytes = 8
	suffix := "-" + HashPrefix(name, hashBytes)
	return TruncateToBytes(name, maxBytes-len(suffix)) + suffix
}
