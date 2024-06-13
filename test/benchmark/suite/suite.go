// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build benchmark
// +build benchmark

package suite

import (
	"context"
	"fmt"
	"os"
	"testing"

	batchv1 "k8s.io/api/batch/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/config"
	"sigs.k8s.io/yaml"
)

const (
	ScaledLabelKey        = "benchmark-test/scaled"
	DefaultControllerName = "gateway.envoyproxy.io/gatewayclass-controller"
)

type BenchmarkTestSuite struct {
	Client         client.Client
	TimeoutConfig  config.TimeoutConfig
	ControllerName string
	Options        BenchmarkOptions

	// Resources template for supported benchmark targets.
	GatewayTemplate   *gwapiv1.Gateway
	HTTPRouteTemplate *gwapiv1.HTTPRoute

	// Template for benchmark test client.
	BenchmarkClient *batchv1.Job

	// Indicates which resources are scaled.
	scaledLabel map[string]string
}

func NewBenchmarkTestSuite(client client.Client, options BenchmarkOptions,
	gatewayManifest, httpRouteManifest, benchmarkClientManifest string) (*BenchmarkTestSuite, error) {
	var (
		gateway         = new(gwapiv1.Gateway)
		httproute       = new(gwapiv1.HTTPRoute)
		benchmarkClient = new(batchv1.Job)
		timeoutConfig   = config.TimeoutConfig{}
	)

	data, err := os.ReadFile(gatewayManifest)
	if err != nil {
		return nil, err
	}
	if err = yaml.Unmarshal(data, gateway); err != nil {
		return nil, err
	}

	data, err = os.ReadFile(httpRouteManifest)
	if err != nil {
		return nil, err
	}
	if err = yaml.Unmarshal(data, httproute); err != nil {
		return nil, err
	}

	data, err = os.ReadFile(benchmarkClientManifest)
	if err != nil {
		return nil, err
	}
	if err = yaml.Unmarshal(data, benchmarkClient); err != nil {
		return nil, err
	}

	config.SetupTimeoutConfig(&timeoutConfig)

	// Prepare static options for benchmark client.
	srcArgs := prepareBenchmarkClientStaticArgs(options)
	dstArgs := benchmarkClient.Spec.Template.Spec.Containers[0].Args
	dstArgs = append(dstArgs, srcArgs...)

	return &BenchmarkTestSuite{
		Client:            client,
		Options:           options,
		TimeoutConfig:     timeoutConfig,
		ControllerName:    DefaultControllerName,
		GatewayTemplate:   gateway,
		HTTPRouteTemplate: httproute,
		BenchmarkClient:   benchmarkClient,
		scaledLabel: map[string]string{
			ScaledLabelKey: "true",
		},
	}, nil
}

func (b *BenchmarkTestSuite) Run(t *testing.T, tests []BenchmarkTest) {
	t.Logf("Running %d benchmark test", len(tests))

	for _, test := range tests {
		t.Logf("Running benchmark test: %s", test.ShortName)

		test.Test(t, b)
	}
}

// Benchmark prepares and runs benchmark test as a Kubernetes Job.
//
// TODO: currently running benchmark test via nighthawk-client,
// consider switching to gRPC nighthawk-service for benchmark test.
// ref: https://github.com/envoyproxy/nighthawk/blob/main/api/client/service.proto
func (b *BenchmarkTestSuite) Benchmark() {
	// TODO:
	//  1. prepare job
	//  2. create and run job
	//  3. wait job complete
	//  4. scrap job log as report
}

func prepareBenchmarkClientStaticArgs(options BenchmarkOptions) []string {
	staticArgs := []string{
		"--rps", options.RPS,
		"--connections", options.Connections,
		"--duration", options.Duration,
		"--concurrency", options.Concurrency,
	}
	if options.PrefetchConnections {
		staticArgs = append(staticArgs, "--prefetch-connections")
	}
	return staticArgs
}

func prepareBenchmarkClientArgs(gatewayHost string, requestHeaders ...string) []string {
	args := make([]string, 0, len(requestHeaders)*2+1)

	for _, reqHeader := range requestHeaders {
		args = append(args, "--request-header", reqHeader)
	}
	args = append(args, gatewayHost)

	return args
}

// ScaleHTTPRoutes scales HTTPRoutes that are all referenced to one Gateway according to
// the scale range: (begin, end]. The afterCreation is a callback function that only runs
// everytime after one HTTPRoutes has been created successfully.
//
// All scaled resources will be labeled with ScaledLabelKey.
func (b *BenchmarkTestSuite) ScaleHTTPRoutes(ctx context.Context, scaleRange [2]uint16, targetName, refGateway string, afterCreation func(route *gwapiv1.HTTPRoute)) error {
	var (
		i          uint16
		begin, end = scaleRange[0], scaleRange[1]
	)

	for i = begin + 1; i <= end; i++ {
		routeName := fmt.Sprintf(targetName, i)
		newRoute := b.HTTPRouteTemplate.DeepCopy()
		newRoute.SetName(routeName)
		newRoute.SetLabels(b.scaledLabel)
		newRoute.Spec.ParentRefs[0].Name = gwapiv1.ObjectName(refGateway)

		if err := b.CreateResource(ctx, newRoute); err != nil {
			return err
		}

		if afterCreation != nil {
			afterCreation(newRoute)
		}
	}

	return nil
}

func (b *BenchmarkTestSuite) CreateResource(ctx context.Context, object client.Object) error {
	if err := b.Client.Create(ctx, object); err != nil {
		if !kerrors.IsAlreadyExists(err) {
			return err
		} else {
			return nil
		}
	}
	return nil
}

func (b *BenchmarkTestSuite) CleanupResource(ctx context.Context, object client.Object) error {
	if err := b.Client.Delete(ctx, object); err != nil {
		if !kerrors.IsNotFound(err) {
			return err
		} else {
			return nil
		}
	}
	return nil
}

// CleanupScaledResources only cleanups all the resources with Scaled label under benchmark-test namespace.
func (b *BenchmarkTestSuite) CleanupScaledResources(ctx context.Context, object client.Object) error {
	if err := b.Client.DeleteAllOf(ctx, object,
		client.MatchingLabels{ScaledLabelKey: "true"}, client.InNamespace("benchmark-test")); err != nil {
		return err
	}
	return nil
}
