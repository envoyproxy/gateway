// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package e2e

import (
	"flag"
	"io/fs"
	"os"
	"testing"

	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/gateway-api/conformance"
	"sigs.k8s.io/gateway-api/conformance/utils/flags"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/conformance/utils/tlog"
	"sigs.k8s.io/gateway-api/pkg/features"

	"github.com/envoyproxy/gateway/test/e2e/tests"
	kubetest "github.com/envoyproxy/gateway/test/utils/kubernetes"
)

func TestE2E(t *testing.T) {
	flag.Parse()
	log.SetLogger(zap.New(zap.WriteTo(os.Stderr), zap.UseDevMode(true)))

	c, cfg := kubetest.NewClient(t)

	if flags.RunTest != nil && *flags.RunTest != "" {
		tlog.Logf(t, "Running E2E test %s with %s GatewayClass\n cleanup: %t\n debug: %t",
			*flags.RunTest, *flags.GatewayClassName, *flags.CleanupBaseResources, *flags.ShowDebug)
	} else {
		tlog.Logf(t, "Running E2E tests with %s GatewayClass\n cleanup: %t\n debug: %t",
			*flags.GatewayClassName, *flags.CleanupBaseResources, *flags.ShowDebug)
	}

	skipTests := []string{
		tests.GatewayInfraResourceTest.ShortName, // https://github.com/envoyproxy/gateway/issues/3191
	}

	// Skip test only work on DualStack cluster
	if tests.IPFamily != "dual" {
		skipTests = append(skipTests,
			tests.BackendDualStackTest.ShortName,
			tests.HTTPRouteDualStackTest.ShortName,
		)
	}

	// Skip Dynamic Resolver test because DNS resolver doesn't work properly in IPV6 Github worker
	if tests.IPFamily == "ipv6" {
		skipTests = append(skipTests,
			tests.DynamicResolverBackendTest.ShortName,
			tests.DynamicResolverBackendWithTLSTest.ShortName,
			tests.RateLimitCIDRMatchTest.ShortName,
			tests.RateLimitMultipleListenersTest.ShortName,
			tests.RateLimitGlobalSharedCidrMatchTest.ShortName,
		)
	}

	if tests.XDSNameSchemeV2() {
		skipTests = append(skipTests,
			tests.EnvoyPatchPolicyTest.ShortName,
		)
	} else {
		skipTests = append(skipTests,
			tests.EnvoyPatchPolicyXDSNameSchemeV2Test.ShortName,
		)
	}

	enabledFeatures := sets.New(features.SupportGateway)
	if tests.EnabledClusterTrustBundle() {
		tlog.Logf(t, "ClusterTrustBundle feature is enabled")
		enabledFeatures.Insert(tests.ClusterTrustBundleFeature)
	}
	// If focusing on a single test, clear the skip list to ensure it runs.
	if *flags.RunTest != "" {
		skipTests = nil
	}
	opts := conformance.DefaultOptions(t)

	opts.Client = c
	opts.RestConfig = cfg
	opts.GatewayClassName = *flags.GatewayClassName
	opts.Debug = *flags.ShowDebug
	opts.CleanupBaseResources = *flags.CleanupBaseResources
	opts.ManifestFS = []fs.FS{Manifests}
	opts.RunTest = *flags.RunTest
	// SupportedFeatures cannot be empty, so we set it to SupportGateway
	// All e2e tests should leave Features empty.
	opts.SupportedFeatures = enabledFeatures
	opts.SkipTests = skipTests
	opts.AllowCRDsMismatch = *flags.AllowCRDsMismatch
	opts.Hook = Hook
	opts.FailFast = true

	cSuite, err := suite.NewConformanceTestSuite(opts)
	if err != nil {
		t.Fatalf("Failed to create ConformanceTestSuite: %v", err)
	}
	tlog.Logf(t, "DefaultPollInterval: %v", cSuite.TimeoutConfig.DefaultPollInterval)
	recorder := NewTimingRecorder()
	t.Cleanup(func() {
		recorder.Report(t)
	})
	timedTests := WrapConformanceTestsWithTiming(tests.ConformanceTests, recorder)
	cSuite.Setup(t, timedTests)
	if cSuite.RunTest != "" {
		tlog.Logf(t, "Running E2E test %s", cSuite.RunTest)
	} else {
		tlog.Logf(t, "Running %d E2E tests", len(tests.ConformanceTests))
	}
	err = cSuite.Run(t, timedTests)
	if err != nil {
		tlog.Fatalf(t, "Failed to run E2E tests: %v", err)
	}
}
