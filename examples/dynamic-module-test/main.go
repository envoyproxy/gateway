// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package main

import (
	sdk "github.com/envoyproxy/envoy/source/extensions/dynamic_modules/sdk/go"
	_ "github.com/envoyproxy/envoy/source/extensions/dynamic_modules/sdk/go/abi"
	"github.com/envoyproxy/envoy/source/extensions/dynamic_modules/sdk/go/shared"
)

func init() {
	sdk.RegisterHttpFilterConfigFactories(map[string]shared.HttpFilterConfigFactory{
		"header_mutation": &headerMutationConfigFactory{},
	})
}

func main() {}

type headerMutationConfigFactory struct {
	shared.EmptyHttpFilterConfigFactory
}

func (f *headerMutationConfigFactory) Create(_ shared.HttpFilterConfigHandle, _ []byte) (shared.HttpFilterFactory, error) {
	return &headerMutationFilterFactory{}, nil
}

type headerMutationFilterFactory struct{}

func (f *headerMutationFilterFactory) Create(handle shared.HttpFilterHandle) shared.HttpFilter {
	return &headerMutationFilter{shared.EmptyHttpFilter{}}
}

type headerMutationFilter struct {
	shared.EmptyHttpFilter
}

func (f *headerMutationFilter) OnResponseHeaders(headers shared.HeaderMap, _ bool) shared.HeadersStatus {
	headers.Set("x-dynamic-module", "true")
	return shared.HeadersStatusContinue
}
