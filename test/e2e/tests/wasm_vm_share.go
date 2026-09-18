// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"context"
	"runtime"
	"testing"
	"time"

	"github.com/prometheus/common/model"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/conformance/utils/tlog"

	"github.com/envoyproxy/gateway/test/utils/prometheus"
)

func init() {
	ConformanceTests = append(ConformanceTests, WasmVMShareTest)
}

// WasmVMShareTest verifies that opted-in policies reuse Wasm VMs across routes.
var WasmVMShareTest = suite.ConformanceTest{
	ShortName:   "WasmVMShare",
	Description: "Test Wasm VM sharing across policies with identical code and different configurations",
	Manifests:   []string{"testdata/wasm-vm-share.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("first route with shared wasm vm", func(t *testing.T) {
			testWasmHTTPCodeSource(t, suite, "http-with-http-wasm-source-shared-1", "http-wasm-source-test-shared-1", "/wasm-http-shared-1")
		})

		t.Run("second route with shared wasm vm", func(t *testing.T) {
			testWasmHTTPCodeSource(t, suite, "http-with-http-wasm-source-shared-2", "http-wasm-source-test-shared-2", "/wasm-http-shared-2")
		})

		// Both policies opt into sharing and use identical code, despite having
		// different plugin names and configurations. Together they contribute one
		// VM per worker plus two base VMs to the process-wide gauge. Without
		// sharing, the count would be twice this value.
		tlog.Logf(t, "concurrency: %d", runtime.NumCPU())
		t.Run("wasm vm count is shared across policies", func(t *testing.T) {
			promQL := `sum(envoy_wasm_wasm_vm_count{app_kubernetes_io_component="proxy", app_kubernetes_io_managed_by="envoy-gateway", app_kubernetes_io_name="envoy", gateway_envoyproxy_io_owning_gateway_name="same-namespace"})`
			expectedCount := model.SampleValue(runtime.NumCPU() + 2)
			tlog.Logf(t, "expected wasm_vm_count: %v", expectedCount)
			if err := wait.PollUntilContextTimeout(context.TODO(), time.Second, time.Minute, true,
				func(_ context.Context) (done bool, err error) {
					v, err := prometheus.QueryPrometheus(suite.Client, promQL)
					if err != nil {
						tlog.Logf(t, "failed to query prometheus: %v", err)
						return false, nil
					}
					if v != nil {
						vectorVal := v.(model.Vector)
						if len(vectorVal) == 1 && vectorVal[0].Value == expectedCount {
							tlog.Logf(t, "got expected wasm_vm_count value: %v", vectorVal[0].Value)
							return true, nil
						} else {
							tlog.Logf(t, "got metric: %v", vectorVal)
						}
					}

					return false, nil
				}); err != nil {
				t.Errorf("failed to get expected wasm_vm_count metric: %v", err)
			}
		})
	},
}
