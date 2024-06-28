// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

// Copyright Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package wasm

import "github.com/envoyproxy/gateway/internal/metrics"

// Const strings for label value.
const (
	// For remote fetch metric.
	fetchSuccess     = "success"
	fetchFailure     = "fetch_failure"
	downloadFailure  = "download_failure"
	manifestFailure  = "manifest_failure"
	checksumMismatch = "checksum_mismatched"
)

var (
	hitTag    = metrics.NewLabel("hit")
	resultTag = metrics.NewLabel("result")

	wasmCacheEntries = metrics.NewGauge(
		"wasm_cache_entries",
		"number of Wasm remote fetch cache entries.",
	)

	wasmCacheLookupCount = metrics.NewCounter(
		"wasm_cache_lookup_count",
		"number of Wasm remote fetch cache lookups.",
	)

	wasmRemoteFetchCount = metrics.NewCounter(
		"wasm_remote_fetch_count",
		"number of Wasm remote fetches and results, including success, download failure, and checksum mismatch.",
	)
)

// TODO zhaohuabing export metrics to control plane dashboard.
