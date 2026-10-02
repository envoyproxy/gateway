// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"context"
	"fmt"
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
	Description: "Test Wasm VM sharing and isolation with HTTP and local modules",
	Manifests:   []string{"testdata/wasm-vm-share.yaml", "testdata/wasm-vm-share-local.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("first route with shared wasm vm", func(t *testing.T) {
			testWasmCodeSource(t, suite, "same-namespace", "http-with-http-wasm-source-shared-1", "http-wasm-source-test-shared-1", "/wasm-http-shared-1", "FOO")
		})

		t.Run("second route with shared wasm vm", func(t *testing.T) {
			testWasmCodeSource(t, suite, "same-namespace", "http-with-http-wasm-source-shared-2", "http-wasm-source-test-shared-2", "/wasm-http-shared-2", "FOO")
		})

		// Both policies specify the Namespace sharing scope and use identical code,
		// despite having different plugin names and configurations. Together they
		// contribute one VM per worker plus two base VMs to the process-wide gauge.
		// Without sharing, the count would be twice this value.
		tlog.Logf(t, "concurrency: %d", runtime.NumCPU())
		t.Run("wasm vm count is shared across policies", func(t *testing.T) {
			testWasmVMCount(t, suite, "same-namespace", model.SampleValue(runtime.NumCPU()+2))
		})

		t.Run("wasm-local-shared", func(t *testing.T) {
			testWasmCodeSource(t, suite, "wasm-local-shared", "wasm-local-shared-1", "wasm-local-shared-1", "/wasm-local-shared-1", "FOO")
			testWasmCodeSource(t, suite, "wasm-local-shared", "wasm-local-shared-2", "wasm-local-shared-2", "/wasm-local-shared-2", "FOO")
			// One worker VM plus two base VMs with sharing enabled.
			testWasmVMCount(t, suite, "wasm-local-shared", 3)
		})

		t.Run("wasm-local-policy", func(t *testing.T) {
			testWasmCodeSource(t, suite, "wasm-local-policy", "wasm-local-policy-1", "wasm-local-policy-1", "/wasm-local-policy-1", "FOO")
			testWasmCodeSource(t, suite, "wasm-local-policy", "wasm-local-policy-2", "wasm-local-policy-2", "/wasm-local-policy-2", "FOO")
			// Policy scope keeps identical local modules in separate VMs.
			// Each policy has one worker VM plus two base VMs.
			testWasmVMCount(t, suite, "wasm-local-policy", 6)
		})

		t.Run("wasm-local-distinct", func(t *testing.T) {
			testWasmCodeSource(t, suite, "wasm-local-distinct", "wasm-local-distinct-1", "wasm-local-distinct-1", "/wasm-local-distinct-1", "FOO")
			testWasmCodeSource(t, suite, "wasm-local-distinct", "wasm-local-distinct-2", "wasm-local-distinct-2", "/wasm-local-distinct-2", "BAR")
			// Distinct modules each have one worker VM plus two base VMs.
			testWasmVMCount(t, suite, "wasm-local-distinct", 6)
		})
	},
}

func testWasmVMCount(t *testing.T, suite *suite.ConformanceTestSuite, gateway string, expectedCount model.SampleValue) {
	t.Helper()
	promQL := fmt.Sprintf(`sum(envoy_wasm_wasm_vm_count{app_kubernetes_io_component="proxy", app_kubernetes_io_managed_by="envoy-gateway", app_kubernetes_io_name="envoy", gateway_envoyproxy_io_owning_gateway_name="%s"})`, gateway)
	tlog.Logf(t, "expected wasm_vm_count: %v", expectedCount)
	if err := wait.PollUntilContextTimeout(context.TODO(), time.Second, time.Minute, true,
		func(_ context.Context) (bool, error) {
			v, err := prometheus.QueryPrometheus(suite.Client, promQL)
			if err != nil {
				tlog.Logf(t, "failed to query prometheus: %v", err)
				return false, nil
			}
			tlog.Logf(t, "got metric: %v; expected: %v", v, expectedCount)
			samples, ok := v.(model.Vector)
			return ok && len(samples) == 1 && samples[0].Value == expectedCount, nil
		}); err != nil {
		t.Errorf("failed to get expected wasm_vm_count metric: %v", err)
	}
}
