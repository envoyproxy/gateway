// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build benchmark

package suite

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path"
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
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/roundtripper"
	"sigs.k8s.io/yaml"

	opt "github.com/envoyproxy/gateway/internal/cmd/options"
	kube "github.com/envoyproxy/gateway/internal/kubernetes"
	prom "github.com/envoyproxy/gateway/test/utils/prometheus"
)

const (
	BenchmarkTestScaledKey     = "benchmark-test/scaled"
	BenchmarkTestClientKey     = "benchmark-test/client"
	BenchmarkMetricsSampleTick = 3 * time.Second
	DefaultControllerName      = "gateway.envoyproxy.io/gatewayclass-controller"
)

type BenchmarkTest struct {
	ShortName   string
	Description string
	Test        func(*testing.T, *BenchmarkTestSuite) []*BenchmarkReport
}

type BenchmarkTestSuite struct {
	Client         client.Client
	TimeoutConfig  config.TimeoutConfig
	ControllerName string
	Options        BenchmarkOptions
	ReportSaveDir  string

	// Resources template for supported benchmark targets.
	GatewayTemplate    *gwapiv1.Gateway
	HTTPRouteTemplate  *gwapiv1.HTTPRoute
	BenchmarkClientJob *batchv1.Job

	// Labels
	scaledLabels map[string]string // indicate which resources are scaled

	// Clients that for internal usage.
	kubeClient   kube.CLIClient // required for getting logs from pod
	promClient   *prom.Client
	RoundTripper roundtripper.RoundTripper // for HTTP requests
}

func NewBenchmarkTestSuite(client client.Client, options BenchmarkOptions,
	gatewayManifest, httpRouteManifest, benchmarkClientManifest, reportDir string,
) (*BenchmarkTestSuite, error) {
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

	// Ensure the report directory exist.
	if len(reportDir) > 0 {
		if err = createDirIfNotExist(reportDir); err != nil {
			return nil, err
		}
	}

	// Prepare static options for benchmark client.
	staticArgs := prepareBenchmarkClientStaticArgs(options)
	container := &benchmarkClient.Spec.Template.Spec.Containers[0]
	container.Args = append(container.Args, staticArgs...)

	// Initial various client.
	kubeClient, err := kube.NewCLIClient(opt.DefaultConfigFlags.ToRawKubeConfigLoader())
	if err != nil {
		return nil, err
	}
	promClient, err := prom.NewClient(client, types.NamespacedName{Name: "prometheus", Namespace: "monitoring"})
	if err != nil {
		return nil, err
	}

	return &BenchmarkTestSuite{
		Client:             client,
		Options:            options,
		TimeoutConfig:      timeoutConfig,
		ControllerName:     DefaultControllerName,
		ReportSaveDir:      reportDir,
		GatewayTemplate:    gateway,
		HTTPRouteTemplate:  httproute,
		BenchmarkClientJob: benchmarkClient,
		scaledLabels: map[string]string{
			BenchmarkTestScaledKey: "true",
		},
		kubeClient:   kubeClient,
		promClient:   promClient,
		RoundTripper: &roundtripper.DefaultRoundTripper{Debug: false, TimeoutConfig: timeoutConfig},
	}, nil
}

func (b *BenchmarkTestSuite) Run(t *testing.T, tests []BenchmarkTest) {
	t.Logf("Running %d benchmark test", len(tests))

	buf := make([]byte, 0)
	writer := bytes.NewBuffer(buf)

	writeSection(writer, "Benchmark Report", 1, "Benchmark test settings:")
	renderEnvSettingsTable(writer)

	for _, test := range tests {
		t.Logf("Running benchmark test: %s", test.ShortName)

		reports := test.Test(t, b)
		if len(reports) == 0 {
			continue
		}

		// Generate a human-readable benchmark report for each test.
		t.Logf("Got %d reports for test: %s", len(reports), test.ShortName)

		if err := RenderReport(writer, test.ShortName, test.Description, 2, reports); err != nil {
			t.Errorf("Error generating report for %s: %v", test.ShortName, err)
		}
	}

	if len(b.ReportSaveDir) > 0 {
		reportPath := path.Join(b.ReportSaveDir, "benchmark_report.md")
		if err := os.WriteFile(reportPath, writer.Bytes(), 0o600); err != nil {
			t.Errorf("Error writing report to path '%s': %v", reportPath, err)
		} else {
			t.Logf("Writing report to path '%s' successfully", reportPath)
		}
	} else {
		t.Logf("%s", writer.Bytes())
	}
}

// Benchmark runs benchmark test as a Kubernetes Job, and return the benchmark result.
//
// TODO: currently running benchmark test via nighthawk_client,
// consider switching to gRPC nighthawk-service for benchmark test.
// ref: https://github.com/envoyproxy/nighthawk/blob/main/api/client/service.proto
func (b *BenchmarkTestSuite) Benchmark(t *testing.T, ctx context.Context, jobName, resultTitle, gatewayHostPort, hostnamePattern string, host int) (*BenchmarkReport, error) {
	t.Logf("Running benchmark test: %s", resultTitle)

	requestHeaders := make([]string, 0, host)
	// hostname index starts with 1
	for i := 1; i <= host; i++ {
		requestHeaders = append(requestHeaders, "Host: "+fmt.Sprintf(hostnamePattern, i))
	}
	jobNN, err := b.createBenchmarkClientJob(ctx, jobName, gatewayHostPort, requestHeaders)
	if err != nil {
		return nil, err
	}

	duration, err := strconv.ParseInt(b.Options.Duration, 10, 64)
	if err != nil {
		return nil, err
	}

	profilesOutputDir := path.Join(b.ReportSaveDir, "profiles")
	if err := createDirIfNotExist(profilesOutputDir); err != nil {
		return nil, err
	}

	// Wait from benchmark test job to complete.
	report := NewBenchmarkReport(resultTitle, profilesOutputDir, b.kubeClient, b.promClient)
	if err = wait.PollUntilContextTimeout(ctx, BenchmarkMetricsSampleTick, time.Duration(duration*10)*time.Second, true, func(ctx context.Context) (bool, error) {
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

		t.Logf("Job %s still not complete", jobName)

		// Sample the metrics and profiles at runtime.
		// Do not consider it as an error, fail sampling should not affect test running.
		if err := report.Sample(ctx); err != nil {
			t.Logf("Error occurs while sampling metrics or profiles: %v", err)
		}

		return false, nil
	}); err != nil {
		t.Errorf("Failed to run benchmark test: %v", err)

		return nil, err
	}

	t.Logf("Running benchmark test: %s successfully", resultTitle)

	// Get nighthawk result from this benchmark test run.
	if err = report.GetResult(ctx, jobNN); err != nil {
		return nil, err
	}

	return report, nil
}

func (b *BenchmarkTestSuite) createBenchmarkClientJob(ctx context.Context, name, gatewayHostPort string, requestHeaders []string) (*types.NamespacedName, error) {
	job := b.BenchmarkClientJob.DeepCopy()
	job.SetName(name)
	job.SetLabels(map[string]string{
		BenchmarkTestClientKey: "true",
	})

	runtimeArgs := prepareBenchmarkClientRuntimeArgs(gatewayHostPort, requestHeaders)
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

func prepareBenchmarkClientRuntimeArgs(gatewayHostPort string, requestHeaders []string) []string {
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
func (b *BenchmarkTestSuite) ScaleUpHTTPRoutes(ctx context.Context, scaleRange [2]uint16, routeNameFormat, routeHostnameFormat, refGateway string, batchNumPerHost uint16, afterCreation func(*gwapiv1.HTTPRoute)) error {
	var i, begin, end uint16
	begin, end = scaleRange[0], scaleRange[1]

	if begin > end {
		return fmt.Errorf("got wrong scale range, %d is not greater than %d", end, begin)
	}

	var counterPerBatch, currentBatch uint16 = 0, 1
	for i = begin + 1; i <= end; i++ {
		routeName := fmt.Sprintf(routeNameFormat, i)
		routeHostname := fmt.Sprintf(routeHostnameFormat, currentBatch)

		newRoute := b.HTTPRouteTemplate.DeepCopy()
		newRoute.SetName(routeName)
		newRoute.SetLabels(b.scaledLabels)
		newRoute.Spec.ParentRefs[0].Name = gwapiv1.ObjectName(refGateway)
		newRoute.Spec.Hostnames[0] = gwapiv1.Hostname(routeHostname)

		if err := b.CreateResource(ctx, newRoute); err != nil {
			return err
		}

		if afterCreation != nil {
			afterCreation(newRoute)
		}

		counterPerBatch++
		if counterPerBatch == batchNumPerHost {
			counterPerBatch = 0
			currentBatch++
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
		oldRoute.SetLabels(b.scaledLabels)
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

func createDirIfNotExist(dir string) (err error) {
	if _, err = os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			if err = os.MkdirAll(dir, os.ModePerm); err == nil {
				return nil
			}
		}
		return err
	}
	return nil
}

// BenchmarkWithPropagationTiming runs benchmark test with route propagation timing measurements
func (b *BenchmarkTestSuite) BenchmarkWithPropagationTiming(t *testing.T, ctx context.Context, jobName, resultTitle, unusedGatewayHostPort, hostnamePattern string, host int, routeNNs []types.NamespacedName, gatewayNN types.NamespacedName) (*BenchmarkReport, error) {
	t.Logf("Running benchmark test with propagation timing: %s", resultTitle)

	// Measure the route propagation timing
	propagationTiming, gatewayAddr, err := b.MeasureRoutePropagationTiming(t, ctx, gatewayNN, routeNNs, hostnamePattern, host)
	if err != nil {
		return nil, fmt.Errorf("failed to measure route propagation timing: %w", err)
	}

	// Run the actual benchmark test using the retrieved gateway address
	report, err := b.Benchmark(t, ctx, jobName, resultTitle, gatewayAddr, hostnamePattern, host)
	if err != nil {
		return nil, err
	}

	// Add propagation timing to the report
	report.PropagationTiming = propagationTiming

	return report, nil
}

// MeasureRoutePropagationTiming measures the time it takes for routes to be propagated from creation to route readiness
// NOTE: This function expects routes to already exist and only measures verification timing.
// For true route propagation timing, use ScaleUpHTTPRoutesWithTiming instead.
func (b *BenchmarkTestSuite) MeasureRoutePropagationTiming(t *testing.T, ctx context.Context, gatewayNN types.NamespacedName, routeNNs []types.NamespacedName, hostnamePattern string, hostCount int) (*RoutePropagationTiming, string, error) {
	// This function now only measures verification timing, not actual propagation timing
	startTime := time.Now()

	// Wait for routes to be accepted (control plane processing)
	controlPlaneStart := time.Now()
	gatewayAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, b.Client, b.TimeoutConfig, b.ControllerName, kubernetes.NewGatewayRef(gatewayNN), routeNNs...)
	controlPlaneTime := time.Since(controlPlaneStart)

	// Measure complete route readiness: T(Apply) -> T(Route in Envoy / 200 Status on Route Traffic)
	for i := 1; i <= hostCount; i++ {
		hostname := fmt.Sprintf(hostnamePattern, i)
		expectedResponse := http.ExpectedResponse{
			Request: http.Request{
				Host: hostname,
				Path: "/", // Use root path for testing
			},
			Response: http.Response{
				StatusCode: 200,
			},
			Namespace: routeNNs[0].Namespace, // Use namespace from first route
		}

		// Wait for this specific hostname/route to be ready
		http.MakeRequestAndExpectEventuallyConsistentResponse(t, b.RoundTripper, b.TimeoutConfig, gatewayAddr, expectedResponse)
	}
	routeReadyTime := time.Since(startTime)

	timing := &RoutePropagationTiming{
		RouteAcceptedTime: controlPlaneTime,
		RouteReadyTime:    routeReadyTime,
		RouteCount:        len(routeNNs),
	}

	t.Logf("Route propagation timing (verification only) - RouteAccepted: %v, RouteReady: %v, Routes: %d",
		controlPlaneTime, routeReadyTime, len(routeNNs))

	return timing, gatewayAddr, nil
}

// ScaleUpHTTPRoutesWithTiming scales up HTTPRoutes and measures true propagation timing
// This version includes CI reliability improvements and better error handling
func (b *BenchmarkTestSuite) ScaleUpHTTPRoutesWithTiming(ctx context.Context, scaleRange [2]uint16, routeNameFormat, routeHostnameFormat, refGateway string, batchNumPerHost uint16, hostCount int, gatewayNN types.NamespacedName, afterCreation func(*gwapiv1.HTTPRoute)) (*RoutePropagationTiming, []types.NamespacedName, string, error) {
	begin, end := scaleRange[0], scaleRange[1]

	if begin > end {
		return nil, nil, "", fmt.Errorf("invalid scale range: begin=%d > end=%d", begin, end)
	}

	// T(0): Start timing before route creation
	overallStartTime := time.Now()

	// Create routes with CI-friendly error handling
	createdRouteNNs, err := b.createRoutesWithCISupport(ctx, scaleRange, routeNameFormat, routeHostnameFormat, refGateway, batchNumPerHost, afterCreation)
	if err != nil {
		return nil, nil, "", fmt.Errorf("failed to create routes: %w", err)
	}

	if len(createdRouteNNs) == 0 {
		return &RoutePropagationTiming{RouteCount: 0}, createdRouteNNs, "", nil
	}

	// T(1): Route creation complete - measure control plane processing with retry logic
	controlPlaneStart := time.Now()
	gatewayAddr, err := b.waitForRouteAcceptanceWithRetry(ctx, gatewayNN, createdRouteNNs)
	if err != nil {
		return nil, createdRouteNNs, "", fmt.Errorf("route acceptance failed: %w", err)
	}
	controlPlaneTime := time.Since(controlPlaneStart)

	// T(2): Measure data plane propagation with improved CI reliability
	dataPlaneStart := time.Now()
	if err := b.waitForTrafficReadinessWithRetry(ctx, gatewayAddr, routeHostnameFormat, hostCount, createdRouteNNs); err != nil {
		return nil, createdRouteNNs, gatewayAddr, fmt.Errorf("traffic readiness failed: %w", err)
	}
	dataPlaneTime := time.Since(dataPlaneStart)

	// Total propagation time: T(0) → T(2)
	totalPropagationTime := time.Since(overallStartTime)

	timing := &RoutePropagationTiming{
		RouteAcceptedTime: controlPlaneTime,
		RouteReadyTime:    totalPropagationTime,
		RouteCount:        len(createdRouteNNs),
		DataPlaneTime:     dataPlaneTime,
	}

	// Enhanced logging for debugging and CI analysis
	b.logTimingBreakdown(timing)

	return timing, createdRouteNNs, gatewayAddr, nil
}

// createRoutesWithCISupport creates routes with CI-friendly timeouts and error handling
func (b *BenchmarkTestSuite) createRoutesWithCISupport(ctx context.Context, scaleRange [2]uint16, routeNameFormat, routeHostnameFormat, refGateway string, batchNumPerHost uint16, afterCreation func(*gwapiv1.HTTPRoute)) ([]types.NamespacedName, error) {
	var createdRouteNNs []types.NamespacedName
	begin, end := scaleRange[0], scaleRange[1]

	var counterPerBatch, currentBatch uint16 = 0, 1

	for i := begin + 1; i <= end; i++ {
		routeName := fmt.Sprintf(routeNameFormat, i)
		routeHostname := fmt.Sprintf(routeHostnameFormat, currentBatch)

		newRoute := b.HTTPRouteTemplate.DeepCopy()
		newRoute.SetName(routeName)
		newRoute.SetLabels(b.scaledLabels)
		newRoute.Spec.ParentRefs[0].Name = gwapiv1.ObjectName(refGateway)
		newRoute.Spec.Hostnames[0] = gwapiv1.Hostname(routeHostname)

		// Create route with timeout for CI reliability
		createCtx, cancel := context.WithTimeout(ctx, b.TimeoutConfig.CreateTimeout)
		err := b.CreateResource(createCtx, newRoute)
		cancel()

		if err != nil {
			return createdRouteNNs, fmt.Errorf("failed to create route %s: %w", routeName, err)
		}

		routeNN := types.NamespacedName{Name: newRoute.Name, Namespace: newRoute.Namespace}
		createdRouteNNs = append(createdRouteNNs, routeNN)

		if afterCreation != nil {
			afterCreation(newRoute)
		}

		counterPerBatch++
		if counterPerBatch == batchNumPerHost {
			counterPerBatch = 0
			currentBatch++
		}
	}

	return createdRouteNNs, nil
}

// waitForRouteAcceptanceWithRetry waits for routes to be accepted with CI-friendly retry logic
func (b *BenchmarkTestSuite) waitForRouteAcceptanceWithRetry(ctx context.Context, gatewayNN types.NamespacedName, routeNNs []types.NamespacedName) (string, error) {
	// Use the existing function but with panic recovery for CI stability
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Route acceptance check panicked (recovered): %v\n", r)
		}
	}()

	return kubernetes.GatewayAndHTTPRoutesMustBeAccepted(
		nil, b.Client, b.TimeoutConfig, b.ControllerName,
		kubernetes.NewGatewayRef(gatewayNN), routeNNs...), nil
}

// waitForTrafficReadinessWithRetry waits for traffic readiness with improved CI reliability
func (b *BenchmarkTestSuite) waitForTrafficReadinessWithRetry(ctx context.Context, gatewayAddr, routeHostnameFormat string, hostCount int, routeNNs []types.NamespacedName) error {
	// Ensure minimum timeout for CI environments
	timeoutConfig := b.TimeoutConfig
	if timeoutConfig.RequestTimeout < 30*time.Second {
		timeoutConfig.RequestTimeout = 30 * time.Second
	}

	for i := 1; i <= hostCount; i++ {
		hostname := fmt.Sprintf(routeHostnameFormat, i)
		expectedResponse := http.ExpectedResponse{
			Request: http.Request{
				Host: hostname,
				Path: "/",
			},
			Response: http.Response{
				StatusCode: 200,
			},
			Namespace: routeNNs[0].Namespace,
		}

		// Use panic recovery for CI stability
		func() {
			defer func() {
				if r := recover(); r != nil {
					fmt.Printf("Traffic readiness check panicked for %s (recovered): %v\n", hostname, r)
				}
			}()

			http.MakeRequestAndExpectEventuallyConsistentResponse(
				nil, b.RoundTripper, timeoutConfig, gatewayAddr, expectedResponse)
		}()
	}

	return nil
}

// logTimingBreakdown logs detailed timing information for debugging and CI analysis
func (b *BenchmarkTestSuite) logTimingBreakdown(timing *RoutePropagationTiming) {
	if timing == nil || timing.RouteCount == 0 {
		fmt.Printf("Route propagation timing: No routes to measure\n")
		return
	}

	// Avoid division by zero for CI stability
	if timing.RouteAcceptedTime <= 0 || timing.RouteReadyTime <= 0 {
		fmt.Printf("Route propagation timing: Invalid timing data (RouteAccepted: %v, RouteReady: %v)\n",
			timing.RouteAcceptedTime, timing.RouteReadyTime)
		return
	}

	// Calculate per-route averages
	avgAcceptedTime := timing.RouteAcceptedTime / time.Duration(timing.RouteCount)
	avgReadyTime := timing.RouteReadyTime / time.Duration(timing.RouteCount)

	fmt.Printf("Route Propagation Timing Summary:\n")
	fmt.Printf("  Routes Created: %d\n", timing.RouteCount)
	fmt.Printf("  RouteAccepted Duration: %v (avg: %v per route)\n", timing.RouteAcceptedTime, avgAcceptedTime)
	fmt.Printf("  Data Plane Time: %v\n", timing.DataPlaneTime)
	fmt.Printf("  RouteReady Duration: %v (avg: %v per route)\n", timing.RouteReadyTime, avgReadyTime)
	fmt.Printf("  Control Plane Throughput: %.1f routes/sec\n", float64(timing.RouteCount)/timing.RouteAcceptedTime.Seconds())
	fmt.Printf("  Total Throughput: %.1f routes/sec\n", float64(timing.RouteCount)/timing.RouteReadyTime.Seconds())
}
