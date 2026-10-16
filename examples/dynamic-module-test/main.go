// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package main

import (
	"encoding/json"
	"fmt"

	sdk "github.com/envoyproxy/envoy/source/extensions/dynamic_modules/sdk/go"
	_ "github.com/envoyproxy/envoy/source/extensions/dynamic_modules/sdk/go/abi"
	"github.com/envoyproxy/envoy/source/extensions/dynamic_modules/sdk/go/shared"
)

func init() {
	sdk.RegisterHttpFilterConfigFactories(map[string]shared.HttpFilterConfigFactory{
		"header_mutation": &headerMutationConfigFactory{},
		"http_callout":    &httpCalloutConfigFactory{},
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

func (f *headerMutationFilterFactory) OnDestroy() {}

type headerMutationFilter struct {
	shared.EmptyHttpFilter
}

func (f *headerMutationFilter) OnResponseHeaders(headers shared.HeaderMap, _ bool) shared.HeadersStatus {
	headers.Set("x-dynamic-module", "true")
	return shared.HeadersStatusContinue
}

type httpCalloutConfig struct {
	Cluster string `json:"cluster"`
}

type httpCalloutConfigFactory struct {
	shared.EmptyHttpFilterConfigFactory
}

func (f *httpCalloutConfigFactory) Create(_ shared.HttpFilterConfigHandle, config []byte) (shared.HttpFilterFactory, error) {
	var calloutConfig httpCalloutConfig
	if err := json.Unmarshal(config, &calloutConfig); err != nil {
		return nil, fmt.Errorf("invalid http callout config: %w", err)
	}
	if calloutConfig.Cluster == "" {
		return nil, fmt.Errorf("http callout config requires cluster")
	}
	return &httpCalloutFilterFactory{config: calloutConfig}, nil
}

type httpCalloutFilterFactory struct {
	config httpCalloutConfig
}

func (f *httpCalloutFilterFactory) Create(handle shared.HttpFilterHandle) shared.HttpFilter {
	return &httpCalloutFilter{
		EmptyHttpFilter: shared.EmptyHttpFilter{},
		handle:          handle,
		cluster:         f.config.Cluster,
	}
}

func (f *httpCalloutFilterFactory) OnDestroy() {}

type httpCalloutFilter struct {
	shared.EmptyHttpFilter
	handle           shared.HttpFilterHandle
	cluster          string
	calloutSucceeded bool
}

func (f *httpCalloutFilter) OnRequestHeaders(_ shared.HeaderMap, _ bool) shared.HeadersStatus {
	result, _ := f.handle.HttpCallout(f.cluster, [][2]string{
		{":method", "GET"},
		{":path", "/"},
		{":authority", "callout"},
		{":scheme", "http"},
	}, nil, 2000, f)
	if result != shared.HttpCalloutInitSuccess {
		f.sendFailure()
	}
	return shared.HeadersStatusStopAllAndBuffer
}

func (f *httpCalloutFilter) OnHttpCalloutDone(_ uint64, result shared.HttpCalloutResult, headers [][2]shared.UnsafeEnvoyBuffer, _ []shared.UnsafeEnvoyBuffer) {
	if result != shared.HttpCalloutSuccess || !hasOKStatus(headers) {
		f.sendFailure()
		return
	}
	f.calloutSucceeded = true
	f.handle.ContinueRequest()
}

func (f *httpCalloutFilter) OnResponseHeaders(headers shared.HeaderMap, _ bool) shared.HeadersStatus {
	if f.calloutSucceeded {
		headers.Set("x-module-callout", "true")
	}
	return shared.HeadersStatusContinue
}

func (f *httpCalloutFilter) sendFailure() {
	f.handle.SendLocalResponse(502, nil, nil, "dynamic module callout failed")
}

func hasOKStatus(headers [][2]shared.UnsafeEnvoyBuffer) bool {
	for _, header := range headers {
		if header[0].ToString() == ":status" && header[1].ToString() == "200" {
			return true
		}
	}
	return false
}
