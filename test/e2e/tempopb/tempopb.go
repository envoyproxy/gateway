// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e
// +build e2e

// this is a package copied from tempo project,
// EG only use pb struct for testing purpose.
package tempopb

type SearchResponse struct {
	Traces  []*TraceSearchMetadata `protobuf:"bytes,1,rep,name=traces,proto3" json:"traces,omitempty"`
	Metrics *SearchMetrics         `protobuf:"bytes,2,opt,name=metrics,proto3" json:"metrics,omitempty"`
}

type TraceSearchMetadata struct {
	TraceID           string `protobuf:"bytes,1,opt,name=traceID,proto3" json:"traceID,omitempty"`
	RootServiceName   string `protobuf:"bytes,2,opt,name=rootServiceName,proto3" json:"rootServiceName,omitempty"`
	RootTraceName     string `protobuf:"bytes,3,opt,name=rootTraceName,proto3" json:"rootTraceName,omitempty"`
	StartTimeUnixNano uint64 `protobuf:"varint,4,opt,name=startTimeUnixNano,proto3" json:"startTimeUnixNano,omitempty"`
	DurationMs        uint32 `protobuf:"varint,5,opt,name=durationMs,proto3" json:"durationMs,omitempty"`
}

type SearchMetrics struct {
	InspectedTraces uint32 `protobuf:"varint,1,opt,name=inspectedTraces,proto3" json:"inspectedTraces,omitempty"`
	InspectedBytes  uint64 `protobuf:"varint,2,opt,name=inspectedBytes,proto3" json:"inspectedBytes,omitempty"`
	InspectedBlocks uint32 `protobuf:"varint,3,opt,name=inspectedBlocks,proto3" json:"inspectedBlocks,omitempty"`
	SkippedBlocks   uint32 `protobuf:"varint,4,opt,name=skippedBlocks,proto3" json:"skippedBlocks,omitempty"`
	SkippedTraces   uint32 `protobuf:"varint,5,opt,name=skippedTraces,proto3" json:"skippedTraces,omitempty"`
	TotalBlockBytes uint64 `protobuf:"varint,6,opt,name=totalBlockBytes,proto3" json:"totalBlockBytes,omitempty"`
}
