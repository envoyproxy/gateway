// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/envoyproxy/gateway/internal/ir"
)

func TestRemoveAddedHeaders(t *testing.T) {
	from := []ir.AddHeader{
		{Name: "Content-Type", Value: []string{"text/plain"}},
		{Name: "Set-Cookie", Value: []string{"dropped=1"}},
		{Name: "X-Keep", Value: []string{"yes"}},
	}
	remove := []ir.AddHeader{
		{Name: "Content-Type", Value: []string{"text/plain"}},
		{Name: "Set-Cookie", Value: []string{"dropped=1"}},
	}

	got := removeAddedHeaders(from, remove)
	require.Equal(t, []ir.AddHeader{{Name: "X-Keep", Value: []string{"yes"}}}, got)
	require.Nil(t, removeAddedHeaders(from[:2], remove))
	require.Equal(t, from, removeAddedHeaders(from, nil))
}
