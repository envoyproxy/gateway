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

const (
	reasonFetchError       = "fetch_error"
	reasonDownloadError    = "download_error"
	reasonManifestError    = "manifest_error"
	reasonChecksumMismatch = "checksum_mismatched"
)

var (
	hitTag = metrics.NewLabel("hit")

	wasmCacheEntries = metrics.NewGauge(
		"wasm_cache_entries",
		"Number of Wasm remote fetch cache entries.",
	)

	wasmCacheLookupTotal = metrics.NewCounter(
		"wasm_cache_lookup_total",
		"Total number of Wasm remote fetch cache lookups.",
	)

	wasmRemoteFetchTotal = metrics.NewCounter(
		"wasm_remote_fetch_total",
		"Total number of Wasm remote fetches and results.",
	)
)
