// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build benchmark
// +build benchmark

package suite

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/config"
	"sigs.k8s.io/yaml"
)

const (
	BenchmarkTestScaledKey = "benchmark-test/scaled"
	BenchmarkTestClientKey = "benchmark-test/client"
	DefaultControllerName  = "gateway.envoyproxy.io/gatewayclass-controller"
)

type BenchmarkTestSuite struct {
	Client         client.Client
	TimeoutConfig  config.TimeoutConfig
	ControllerName string
	Options        BenchmarkOptions
	ReportSavePath string

	// Resources template for supported benchmark targets.
	GatewayTemplate    *gwapiv1.Gateway
	HTTPRouteTemplate  *gwapiv1.HTTPRoute
	BenchmarkClientJob *batchv1.Job

	// Indicates which resources are scaled.
	scaledLabel map[string]string
}

func NewBenchmarkTestSuite(client client.Client, options BenchmarkOptions,
	gatewayManifest, httpRouteManifest, benchmarkClientManifest, reportPath string) (*BenchmarkTestSuite, error) {
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

	// Reset some timeout config for the benchmark test.
	config.SetupTimeoutConfig(&timeoutConfig)
	timeoutConfig.RouteMustHaveParents = 180 * time.Second

	// Prepare static options for benchmark client.
	staticArgs := prepareBenchmarkClientStaticArgs(options)
	container := &benchmarkClient.Spec.Template.Spec.Containers[0]
	container.Args = append(container.Args, staticArgs...)

	return &BenchmarkTestSuite{
		Client:             client,
		Options:            options,
		TimeoutConfig:      timeoutConfig,
		ControllerName:     DefaultControllerName,
		ReportSavePath:     reportPath,
		GatewayTemplate:    gateway,
		HTTPRouteTemplate:  httproute,
		BenchmarkClientJob: benchmarkClient,
		scaledLabel: map[string]string{
			BenchmarkTestScaledKey: "true",
		},
	}, nil
}

func (b *BenchmarkTestSuite) Run(t *testing.T, tests []BenchmarkTest) {
	t.Logf("Running %d benchmark test", len(tests))

	buf := make([]byte, 0)
	writer := bytes.NewBuffer(buf)

	writeSection(writer, "Benchmark Report", 1, "")
	renderEnvSettingsTable(writer)

	for _, test := range tests {
		t.Logf("Running benchmark test: %s", test.ShortName)

		reports := test.Test(t, b)
		if len(reports) == 0 {
			continue
		}

		// Generate a human-readable benchmark report for each test.
		t.Logf("Got %d reports for test: %s", len(reports), test.ShortName)

		if err := RenderReport(writer, "Test: "+test.ShortName, test.Description, reports, 2); err != nil {
			t.Errorf("Error generating report for %s: %v", test.ShortName, err)
		}
	}

	if len(b.ReportSavePath) > 0 {
		if err := os.WriteFile(b.ReportSavePath, writer.Bytes(), 0644); err != nil {
			t.Errorf("Error writing report to path '%s': %v", b.ReportSavePath, err)
		} else {
			t.Logf("Writing report to path '%s' successfully", b.ReportSavePath)
		}
	} else {
		t.Log(fmt.Sprintf("%s", writer.Bytes()))
	}
}

// Benchmark runs benchmark test as a Kubernetes Job, and return the benchmark result.
//
// TODO: currently running benchmark test via nighthawk_client,
// consider switching to gRPC nighthawk-service for benchmark test.
// ref: https://github.com/envoyproxy/nighthawk/blob/main/api/client/service.proto
func (b *BenchmarkTestSuite) Benchmark(t *testing.T, ctx context.Context, name, gatewayHostPort string, requestHeaders ...string) (*BenchmarkReport, error) {
	t.Logf("Running benchmark test: %s", name)

	jobNN, err := b.createBenchmarkClientJob(ctx, name, gatewayHostPort, requestHeaders...)
	if err != nil {
		return nil, err
	}

	duration, err := strconv.ParseInt(b.Options.Duration, 10, 64)
	if err != nil {
		return nil, err
	}

	// Wait from benchmark test job to complete.
	if err = wait.PollUntilContextTimeout(ctx, 10*time.Second, time.Duration(duration*10)*time.Second, true, func(ctx context.Context) (bool, error) {
		job := new(batchv1.Job)
		if err = b.Client.Get(ctx, *jobNN, job); err != nil {
			return false, err
		}

		for _, condition := range job.Status.Conditions {
			if condition.Type == batchv1.JobComplete && condition.Status == "True" {
				return true, nil
			}

			// Early return if job already failed.
			if condition.Type == batchv1.JobFailed && condition.Status == "True" &&
				condition.Reason == batchv1.JobReasonBackoffLimitExceeded {
				return false, fmt.Errorf("job already failed")
			}
		}

		t.Logf("Job %s still not complete", name)

		return false, nil
	}); err != nil {
		t.Errorf("Failed to run benchmark test: %v", err)

		return nil, err
	}

	t.Logf("Running benchmark test: %s successfully", name)

	report, err := NewBenchmarkReport(name)
	if err != nil {
		return nil, err
	}

	// Get all the reports from this benchmark test run.
	if err = report.Collect(t, ctx, jobNN); err != nil {
		return nil, err
	}

	report.Print(t, name)

	return report, nil
}

func (b *BenchmarkTestSuite) createBenchmarkClientJob(ctx context.Context, name, gatewayHostPort string, requestHeaders ...string) (*types.NamespacedName, error) {
	job := b.BenchmarkClientJob.DeepCopy()
	job.SetName(name)

	runtimeArgs := prepareBenchmarkClientRuntimeArgs(gatewayHostPort, requestHeaders...)
	container := &job.Spec.Template.Spec.Containers[0]
	container.Args = append(container.Args, runtimeArgs...)

	if err := b.CreateResource(ctx, job); err != nil {
		return nil, err
	}

	return &types.NamespacedName{Name: job.Name, Namespace: job.Namespace}, nil
}

func prepareBenchmarkClientStaticArgs(options BenchmarkOptions) []string {
	staticArgs := []string{
		"--rps", options.RPS,
		"--connections", options.Connections,
		"--duration", options.Duration,
		"--concurrency", options.Concurrency,
	}
	return staticArgs
}

func prepareBenchmarkClientRuntimeArgs(gatewayHostPort string, requestHeaders ...string) []string {
	args := make([]string, 0, len(requestHeaders)*2+1)

	for _, reqHeader := range requestHeaders {
		args = append(args, "--request-header", reqHeader)
	}
	args = append(args, "http://"+gatewayHostPort)

	return args
}

// ScaleUpHTTPRoutes scales up HTTPRoutes that are all referenced to one Gateway according to
// the scale range: (a, b], which scales up from a to b with a <= b.
//
// The `afterCreation` is a callback function that only runs every time after one HTTPRoutes
// has been created successfully.
//
// All created scaled resources will be labeled with BenchmarkTestScaledKey.
func (b *BenchmarkTestSuite) ScaleUpHTTPRoutes(ctx context.Context, scaleRange [2]uint16, routeNameFormat, refGateway string, afterCreation func(*gwapiv1.HTTPRoute)) error {
	var i, begin, end uint16
	begin, end = scaleRange[0], scaleRange[1]

	if begin > end {
		return fmt.Errorf("got wrong scale range, %d is not greater than %d", end, begin)
	}

	for i = begin + 1; i <= end; i++ {
		routeName := fmt.Sprintf(routeNameFormat, i)
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

// ScaleDownHTTPRoutes scales down HTTPRoutes that are all referenced to one Gateway according to
// the scale range: [a, b), which scales down from a to b with a > b.
//
// The `afterDeletion` is a callback function that only runs every time after one HTTPRoutes has
// been deleted successfully.
func (b *BenchmarkTestSuite) ScaleDownHTTPRoutes(ctx context.Context, scaleRange [2]uint16, routeNameFormat, refGateway string, afterDeletion func(*gwapiv1.HTTPRoute)) error {
	var i, begin, end uint16
	begin, end = scaleRange[0], scaleRange[1]

	if begin <= end {
		return fmt.Errorf("got wrong scale range, %d is not less than %d", end, begin)
	}

	if end == 0 {
		return fmt.Errorf("cannot scale routes down to zero")
	}

	for i = begin; i > end; i-- {
		routeName := fmt.Sprintf(routeNameFormat, i)
		oldRoute := b.HTTPRouteTemplate.DeepCopy()
		oldRoute.SetName(routeName)
		oldRoute.SetLabels(b.scaledLabel)
		oldRoute.Spec.ParentRefs[0].Name = gwapiv1.ObjectName(refGateway)

		if err := b.DeleteResource(ctx, oldRoute); err != nil {
			return err
		}

		if afterDeletion != nil {
			afterDeletion(oldRoute)
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

func (b *BenchmarkTestSuite) DeleteResource(ctx context.Context, object client.Object) error {
	if err := b.Client.Delete(ctx, object); err != nil {
		if !kerrors.IsNotFound(err) {
			return err
		} else {
			return nil
		}
	}
	return nil
}

// DeleteScaledResources only cleanups all the resources under benchmark-test namespace.
func (b *BenchmarkTestSuite) DeleteScaledResources(ctx context.Context, object client.Object) error {
	if err := b.Client.DeleteAllOf(ctx, object,
		client.MatchingLabels{BenchmarkTestScaledKey: "true"}, client.InNamespace("benchmark-test")); err != nil {
		return err
	}
	return nil
}

// RegisterCleanup registers cleanup functions for all benchmark test resources.
func (b *BenchmarkTestSuite) RegisterCleanup(t *testing.T, ctx context.Context, object, scaledObject client.Object) {
	t.Cleanup(func() {
		t.Logf("Start to cleanup benchmark test resources")

		_ = b.DeleteResource(ctx, object)
		_ = b.DeleteScaledResources(ctx, scaledObject)

		t.Logf("Clean up complete!")
	})
}
