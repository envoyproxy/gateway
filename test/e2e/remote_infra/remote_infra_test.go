// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package remoteinfra

import (
	"encoding/json"
	"flag"
	"io/fs"
	"testing"

	"sigs.k8s.io/gateway-api/conformance/utils/flags"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/conformance/utils/tlog"
	"sigs.k8s.io/gateway-api/pkg/features"

	"github.com/envoyproxy/gateway/test/e2e"
	"github.com/envoyproxy/gateway/test/e2e/tests"
	kubetest "github.com/envoyproxy/gateway/test/utils/kubernetes"
)

func TestRemoteInfra(t *testing.T) {
	if !tests.IsRemoteInfraMode() {
		t.Skip("Remote Infra Mode not specified, skipping tests.")
	}
	flag.Parse()

	c, cfg := kubetest.NewClient(t)
	suiteOpts := suite.ConfigurableOptions{}
	flags.ApplyAll(&suiteOpts)
	data, _ := json.MarshalIndent(suiteOpts, "", "  ")
	tlog.Logf(t, "Running Remote Infra tests with options: %s\n", string(data))
	suiteOpts.TimeoutConfig = tests.TimeoutConfig()
	// SupportedFeatures cannot be empty, so we set it to SupportGateway
	// All e2e tests should leave Features empty.
	suiteOpts.SupportedFeatures = []features.FeatureName{features.SupportGateway}
	suiteOpts.SkipTests = []string{}
	suiteOpts.FailFast = true
	suiteOpts.CleanupTestResources = true
	cSuite, err := suite.NewConformanceTestSuite(suite.ConformanceOptions{
		Client:              c,
		RestConfig:          cfg,
		ConfigurableOptions: suiteOpts,
		ManifestFS:          []fs.FS{e2e.Manifests},
		Hook:                e2e.Hook,
	})
	if err != nil {
		t.Fatalf("Failed to create ConformanceTestSuite: %v", err)
	}

	recorder := e2e.NewTimingRecorder()
	t.Cleanup(func() {
		recorder.Report(t)
	})
	timedTests := e2e.WrapConformanceTestsWithTiming(tests.RemoteInfraTests, recorder)

	// Apply the base conformance resources (namespaces, backends, secrets, etc.) and
	// wait for them to be ready. Unlike the merge_gateways/multiple_gc suites, the
	// remote-infra suite runs standalone and cannot rely on a previous suite having
	// already created the base resources.
	cSuite.Setup(t, timedTests)

	tlog.Logf(t, "Running %d Remote Infra tests", len(tests.RemoteInfraTests))
	err = cSuite.Run(t, timedTests)
	if err != nil {
		t.Fatalf("Failed to run Remote Infra tests: %v", err)
	}
}
