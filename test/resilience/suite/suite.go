// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build resilience

package suite

import (
	"context"
	"github.com/envoyproxy/gateway/test/utils/kubernetes"
	"io/fs"
	"testing"
	"time"

	opt "github.com/envoyproxy/gateway/internal/cmd/options"
	kube "github.com/envoyproxy/gateway/internal/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/gateway-api/conformance/utils/config"
)

const (
	BenchmarkTestScaledKey = "benchmark-test/scaled"
	BenchmarkTestClientKey = "benchmark-test/client"
	DefaultControllerName  = "gateway.envoyproxy.io/gatewayclass-controller"
)

type ResilienceTest struct {
	ShortName   string
	Description string
	Test        func(*testing.T, *ResilienceTestSuite)
}

type ResilienceTestSuite struct {
	Client         client.Client
	TimeoutConfig  config.TimeoutConfig
	ControllerName string
	ReportSaveDir  string
	KubeActions    *kubernetes.KubeActions
	// Labels
	scaledLabels map[string]string // indicate which resources are scaled

	// Clients that for internal usage.
	kubeClient       kube.CLIClient // required for getting logs from pod\
	ManifestFS       []fs.FS
	GatewayClassName string
}

func NewResilienceTestSuite(client client.Client, reportDir string, manifestFS []fs.FS, gcn string) (*ResilienceTestSuite, error) {
	var (
		timeoutConfig = config.TimeoutConfig{}
	)

	// Reset some timeout config for the benchmark test.
	config.SetupTimeoutConfig(&timeoutConfig)
	timeoutConfig.RouteMustHaveParents = 180 * time.Second

	// Initial various client.
	kubeClient, err := kube.NewCLIClient(opt.DefaultConfigFlags.ToRawKubeConfigLoader())
	if err != nil {
		return nil, err
	}
	KubeActions := kubernetes.NewKubeHelper(client, kubeClient)
	return &ResilienceTestSuite{
		Client:           client,
		ManifestFS:       manifestFS,
		TimeoutConfig:    timeoutConfig,
		ControllerName:   DefaultControllerName,
		ReportSaveDir:    reportDir,
		GatewayClassName: gcn,
		scaledLabels: map[string]string{
			BenchmarkTestScaledKey: "true",
		},
		KubeActions: KubeActions,
		kubeClient:  kubeClient,
	}, nil
}

func (rts *ResilienceTestSuite) WithResCleanUp(ctx context.Context, t *testing.T, f func() (client.Object, error)) error {
	res, err := f()
	t.Cleanup(func() {
		t.Logf("Start to cleanup resilsence test resources")
		if deleteErr := rts.Client.Delete(ctx, res); deleteErr != nil {
		}

		t.Logf("Clean up complete!")
	})
	return err
}

func (rts *ResilienceTestSuite) Kube() *kubernetes.KubeActions {
	return rts.KubeActions
}

func (rts *ResilienceTestSuite) Run(t *testing.T, tests []ResilienceTest) {
	t.Logf("Running %d resilence tests", len(tests))
	for _, test := range tests {
		t.Logf("Running resilence test: %s", test.ShortName)
		test.Test(t, rts)
	}
}

func (rts *ResilienceTestSuite) RegisterCleanup(t *testing.T, ctx context.Context, object client.Object) {
	t.Cleanup(func() {
		t.Logf("Start to cleanup resilsence test resources")
		if deleteErr := rts.Client.Delete(ctx, object); deleteErr != nil {
		}

		t.Logf("Clean up complete!")
	})
}
