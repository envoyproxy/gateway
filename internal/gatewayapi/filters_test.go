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

func TestRemoveAddedHeadersKeepsEarlierSameNameModifier(t *testing.T) {
	from := []ir.AddHeader{
		{Name: "Set-Cookie", Value: []string{"keep=1"}},
		{Name: "Set-Cookie", Value: []string{"dropped=1"}},
		{Name: "Content-Type", Value: []string{"text/plain"}},
	}
	remove := []ir.AddHeader{
		{Name: "Set-Cookie", Value: []string{"dropped=1"}},
		{Name: "Content-Type", Value: []string{"text/plain"}},
	}

	got := removeAddedHeaders(from, remove)
	require.Equal(t, []ir.AddHeader{{Name: "Set-Cookie", Value: []string{"keep=1"}}}, got)
}

func TestCloneHTTPFiltersContextDeepCopiesURLRewrite(t *testing.T) {
	prefix := "/v2"
	orig := &HTTPFiltersContext{
		HTTPFilterIR: &HTTPFilterIR{
			DirectResponse: &ir.CustomResponse{StatusCode: new(uint32(403))},
			URLRewrite: &ir.URLRewrite{
				Path: &ir.ExtendedHTTPPathModifier{
					HTTPPathModifier: ir.HTTPPathModifier{PrefixMatchReplace: &prefix},
				},
			},
		},
	}

	cloned := cloneHTTPFiltersContext(orig)
	cloned.URLRewrite.Host = &ir.HTTPHostModifier{
		PathRegex: &ir.RegexMatchReplace{Pattern: "^/v1/([^/]+)/.*$", Substitution: "\\1.svc"},
	}

	require.Nil(t, orig.URLRewrite.Host)
	require.NotNil(t, cloned.URLRewrite.Host)
	require.NotSame(t, orig.URLRewrite, cloned.URLRewrite)
	require.Equal(t, "/v2", *orig.URLRewrite.Path.PrefixMatchReplace)
}

func TestConfiguredDirectResponseIs2xx(t *testing.T) {
	ok := uint32(200)
	created := uint32(201)
	denied := uint32(404)
	serverErr := uint32(500)

	require.False(t, configuredDirectResponseIs2xx(nil))
	require.False(t, configuredDirectResponseIs2xx(&HTTPFiltersContext{HTTPFilterIR: &HTTPFilterIR{}}))
	require.False(t, configuredDirectResponseIs2xx(&HTTPFiltersContext{
		HTTPFilterIR: &HTTPFilterIR{
			DirectResponse: &ir.CustomResponse{StatusCode: &ok},
		},
	}))
	require.True(t, configuredDirectResponseIs2xx(&HTTPFiltersContext{
		HTTPFilterIR: &HTTPFilterIR{
			DirectResponse:           &ir.CustomResponse{StatusCode: &ok},
			configuredDirectResponse: true,
		},
	}))
	require.True(t, configuredDirectResponseIs2xx(&HTTPFiltersContext{
		HTTPFilterIR: &HTTPFilterIR{
			DirectResponse:           &ir.CustomResponse{StatusCode: &created},
			configuredDirectResponse: true,
		},
	}))
	require.False(t, configuredDirectResponseIs2xx(&HTTPFiltersContext{
		HTTPFilterIR: &HTTPFilterIR{
			DirectResponse:           &ir.CustomResponse{StatusCode: &denied},
			configuredDirectResponse: true,
		},
	}))
	require.False(t, configuredDirectResponseIs2xx(&HTTPFiltersContext{
		HTTPFilterIR: &HTTPFilterIR{
			DirectResponse:           &ir.CustomResponse{StatusCode: &serverErr},
			configuredDirectResponse: true,
		},
	}))
}

func TestRemoveAddedHeadersRemovesLastIdenticalOccurrence(t *testing.T) {
	from := []ir.AddHeader{
		{Name: "X-Foo", Value: []string{"same"}},
		{Name: "X-Foo", Value: []string{"same"}},
	}
	remove := []ir.AddHeader{
		{Name: "X-Foo", Value: []string{"same"}},
	}

	got := removeAddedHeaders(from, remove)
	require.Equal(t, []ir.AddHeader{{Name: "X-Foo", Value: []string{"same"}}}, got)
}
