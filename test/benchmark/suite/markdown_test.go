// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build benchmark

package suite

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/envoyproxy/gateway/test/benchmark/proto"
)

func fakeCaseResult() *proto.Result {
	m := map[string]interface{}{
		"name": "worker_fake",
		"statistics": []interface{}{
			map[string]interface{}{
				"id":    "benchmark_http_client.request_to_response",
				"count": "1",
				"mean":  "0.001s",
				"min":   "0.0001s",
				"max":   "0.002s",
				"percentiles": []interface{}{
					map[string]interface{}{"duration": "0.0001s", "percentile": 0, "count": "1"},
					map[string]interface{}{"duration": "0.002s", "percentile": 1, "count": "1"},
				},
			},
		},
		"counters": []interface{}{
			map[string]interface{}{"name": "benchmark.http_2xx", "value": "1"},
			map[string]interface{}{"name": "upstream_cx_tx_bytes_total", "value": "100"},
		},
		"execution_duration": "1s",
		"execution_start":    time.Now().UTC().Format(time.RFC3339),
		"user_defined_outputs": []interface{}{
			map[string]interface{}{
				"name":  "fake_output",
				"value": "generated",
				"meta": map[string]interface{}{
					"note": "this is a fake result for tests",
				},
			},
		},
	}

	b, err := json.Marshal(m)
	if err != nil {
		return &proto.Result{}
	}

	var r proto.Result
	if err := protojson.Unmarshal(b, &r); err != nil {
		return &proto.Result{}
	}

	return &r
}

func TestToMarkdown(t *testing.T) {
	input := &BenchmarkSuiteReport{
		Settings: map[string]string{
			"rps":        "1000",
			"connection": "100",
		},
		Reports: []*BenchmarkTestReport{
			{
				Title:       "fake-title",
				Description: "fake-description",
				Reports: []*BenchmarkCaseReport{
					{
						Title:             "case-title",
						Routes:            100,
						RoutesPerHostname: 2,
						Result:            fakeCaseResult(),
						Phase:             "fake-phase",
						HeapProfileImage:  "fake-image",
					},
				},
			},
		},
	}
	out, err := ToMarkdown(input)
	require.NoError(t, err)
	require.NotEmpty(t, out)

	data, err := os.ReadFile("testdata/markdown_output.md")
	require.NoError(t, err)
	require.Equal(t, string(data), string(out))
}
