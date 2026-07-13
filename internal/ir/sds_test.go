// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package ir

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/require"
)

func TestSDSClusterNameFromURLDistinguishesUnixSocketPaths(t *testing.T) {
	first := SDSClusterNameFromURL("/run/a/b/socket")
	second := SDSClusterNameFromURL("/run/a_b/socket")

	require.Equal(t, "sds_run_a_b_socket_73917e80488448b0df63fc687c91913f", first)
	require.NotEqual(t, first, second)
	require.Contains(t, first, "run_a_b_socket")
	require.Contains(t, second, "run_a_b_socket")
}

func TestSDSClusterNameFromURLPreservesValidUTF8(t *testing.T) {
	url := "unix:///" + strings.Repeat("a", 47) + "é/socket"

	name := SDSClusterNameFromURL(url)

	require.True(t, utf8.ValidString(name))
}

func TestSDSClusterNameFromURLUsesStrongHashWithoutReadablePrefix(t *testing.T) {
	name := SDSClusterNameFromURL("unix:///")

	require.Len(t, strings.TrimPrefix(name, "sds_"), 32)
}
