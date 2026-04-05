// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build resilience

package resilience

import (
	"flag"
	"io/fs"
	"os"
	"testing"

	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/gateway-api/conformance/utils/flags"

	"github.com/envoyproxy/gateway/test/resilience/suite"
	"github.com/envoyproxy/gateway/test/resilience/tests"
	kubetest "github.com/envoyproxy/gateway/test/utils/kubernetes"
)

func TestResilience(t *testing.T) {
	cli, _ := kubetest.NewClient(t)
	// Parse benchmark options.
	flag.Parse()
	log.SetLogger(zap.New(zap.WriteTo(os.Stderr), zap.UseDevMode(true)))
	bSuite, err := suite.NewResilienceTestSuite(
		cli,
		*suite.ReportSaveDir,
		[]fs.FS{Manifests},
		*flags.GatewayClassName,
	)
	if err != nil {
		t.Fatalf("Failed to create the resilience test suite: %v", err)
	}
	t.Logf("Running %d resilience tests", len(tests.ResilienceTests))
	bSuite.Run(t, tests.ResilienceTests)
}
