// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package naming

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/types"
)

func TestServiceName(t *testing.T) {
	cases := []struct {
		nn       types.NamespacedName
		expected string
	}{
		{
			nn: types.NamespacedName{
				Name:      "foo",
				Namespace: "bar",
			},
			expected: "foo.bar",
		},
	}

	for _, c := range cases {
		t.Run("", func(t *testing.T) {
			got := ServiceName(c.nn)
			assert.Equal(t, c.expected, got)
		})
	}
}

func TestTruncateToBytes(t *testing.T) {
	cases := []struct {
		name     string
		in       string
		maxBytes int
		expected string
	}{
		{name: "shorter than budget", in: "abc", maxBytes: 10, expected: "abc"},
		{name: "exactly at budget", in: "abcde", maxBytes: 5, expected: "abcde"},
		{name: "ascii truncated", in: "abcdefgh", maxBytes: 3, expected: "abc"},
		{name: "zero budget", in: "abc", maxBytes: 0, expected: ""},
		// A 3-byte rune must be dropped whole rather than cut mid-sequence.
		{name: "multibyte rune not split", in: "ab世", maxBytes: 4, expected: "ab"},
		{name: "multibyte rune fits", in: "ab世", maxBytes: 5, expected: "ab世"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := TruncateToBytes(c.in, c.maxBytes)
			assert.Equal(t, c.expected, got)
			assert.True(t, utf8.ValidString(got))
		})
	}
}

func TestHashPrefix(t *testing.T) {
	// sha256("abc") = ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad
	assert.Equal(t, "ba7816bf", HashPrefix("abc", 4))
	assert.Equal(t, "ba7816bf8f01cfea", HashPrefix("abc", 8))
	assert.NotEqual(t, HashPrefix("abc", 8), HashPrefix("abd", 8))
}

func TestBounded(t *testing.T) {
	t.Run("under budget is returned verbatim", func(t *testing.T) {
		assert.Equal(t, "configmap/ns/ca-cert", Bounded("configmap/ns/ca-cert", 64))
	})

	t.Run("over budget is truncated and hashed", func(t *testing.T) {
		long := strings.Repeat("a", 100)
		got := Bounded(long, 32)
		assert.Len(t, got, 32)
		assert.Equal(t, strings.Repeat("a", 15)+"-"+HashPrefix(long, 8), got)
	})

	t.Run("names sharing a truncated prefix stay distinct", func(t *testing.T) {
		head := strings.Repeat("a", 100)
		assert.NotEqual(t, Bounded(head+"one", 32), Bounded(head+"two", 32))
	})

	t.Run("result never exceeds the budget", func(t *testing.T) {
		for _, n := range []int{24, 32, 64, 128} {
			assert.LessOrEqual(t, len(Bounded(strings.Repeat("x", 500), n)), n)
		}
	})
}
