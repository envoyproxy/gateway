// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package main

import (
	"github.com/proxy-wasm/proxy-wasm-go-sdk/proxywasm"
	"github.com/proxy-wasm/proxy-wasm-go-sdk/proxywasm/types"
)

func main() {}

func init() {
	proxywasm.SetHttpContext(func(uint32) types.HttpContext {
		return &calloutFilter{}
	})
}

type calloutFilter struct {
	types.DefaultHttpContext
	calloutSucceeded bool
}

//nolint:staticcheck // Match the SDK callback name.
func (f *calloutFilter) OnHttpRequestHeaders(int, bool) types.Action {
	// Another route can bind this alias differently. Resolve it for each request.
	cluster, err := proxywasm.GetProperty([]string{
		"xds", "route_metadata", "filter_metadata",
		"gateway.envoyproxy.io/extension-backends", "resolver",
	})
	if err != nil {
		sendFailure()
		return types.ActionPause
	}
	_, err = proxywasm.DispatchHttpCall(string(cluster), [][2]string{
		{":method", "GET"},
		{":path", "/"},
		{":authority", "callout"},
	}, nil, nil, 2000, f.onCalloutResponse)
	if err != nil {
		sendFailure()
	}
	return types.ActionPause
}

func (f *calloutFilter) onCalloutResponse(int, int, int) {
	// The SDK restores the request context before invoking this callback.
	headers, err := proxywasm.GetHttpCallResponseHeaders()
	if err == nil {
		for _, header := range headers {
			if header[0] == ":status" && header[1] == "200" {
				f.calloutSucceeded = true
				if err := proxywasm.ResumeHttpRequest(); err != nil {
					sendFailure()
				}
				return
			}
		}
	}
	sendFailure()
}

//nolint:staticcheck // Match the SDK callback name.
func (f *calloutFilter) OnHttpResponseHeaders(int, bool) types.Action {
	if f.calloutSucceeded {
		_ = proxywasm.AddHttpResponseHeader("x-wasm-callout", "true")
	}
	return types.ActionContinue
}

func sendFailure() {
	_ = proxywasm.SendHttpResponse(502, nil, nil, -1)
}
