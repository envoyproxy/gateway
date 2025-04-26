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

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/config"
	"sigs.k8s.io/yaml"

	opt "github.com/envoyproxy/gateway/internal/cmd/options"
	kube "github.com/envoyproxy/gateway/internal/kubernetes"
	prom "github.com/envoyproxy/gateway/test/utils/prometheus"
)

const (
	BenchmarkTestScaledKey                  = "benchmark-test/scaled"
	BenchmarkTestClientKey                  = "benchmark-test/client"
	BenchmarkTestServerDeploymentNameFormat = "nighthawk-test-server-%d"
	BenchmarkTestServerServiceFormat        = BenchmarkTestServerDeploymentNameFormat
	BenchmarkMetricsSampleTick              = 3 * time.Second
	DefaultControllerName                   = "gateway.envoyproxy.io/gatewayclass-controller"
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
	GatewayTemplate           *gwapiv1.Gateway
	HTTPRouteTemplate         *gwapiv1.HTTPRoute
	BenchmarkClientJob        *batchv1.Job
	BenchmarkServerDeployment *appsv1.Deployment
	BenchmarkServerService    *corev1.Service

	// Labels
	scaledLabels map[string]string // indicate which resources are scaled

	// Clients that for internal usage.
	kubeClient kube.CLIClient // required for getting logs from pod
	promClient *prom.Client
}

func NewBenchmarkTestSuite(
	client client.Client,
	options BenchmarkOptions,
	gatewayManifest, httpRouteManifest, benchmarkClientManifest, benchmarkServerManifest, reportDir string,
) (*BenchmarkTestSuite, error) {
	var (
		gateway                   = new(gwapiv1.Gateway)
		httproute                 = new(gwapiv1.HTTPRoute)
		benchmarkClient           = new(batchv1.Job)
		benchmarkServerDeployment = new(appsv1.Deployment)
		benchmarkServerService    = new(corev1.Service)
		timeoutConfig             = config.TimeoutConfig{}
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

	data, err = os.ReadFile(benchmarkServerManifest)
	if err != nil {
		return nil, err
	}
	splitedData := bytes.SplitN(data, []byte("---"), 2)
	if err = yaml.Unmarshal(splitedData[0], benchmarkServerDeployment); err != nil {
		return nil, err
	}
	if err = yaml.Unmarshal(splitedData[1], benchmarkServerService); err != nil {
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
		Client:                    client,
		Options:                   options,
		TimeoutConfig:             timeoutConfig,
		ControllerName:            DefaultControllerName,
		ReportSaveDir:             reportDir,
		GatewayTemplate:           gateway,
		HTTPRouteTemplate:         httproute,
		BenchmarkClientJob:        benchmarkClient,
		BenchmarkServerDeployment: benchmarkServerDeployment,
		BenchmarkServerService:    benchmarkServerService,
		scaledLabels: map[string]string{
			BenchmarkTestScaledKey: "true",
		},
		kubeClient: kubeClient,
		promClient: promClient,
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

/*
 * HTTPRoutes resource processing
 */

func (b *BenchmarkTestSuite) getExpectHTTPRoute(name, hostname, gatewayName, serviceName string) *gwapiv1.HTTPRoute {
	newRoute := b.HTTPRouteTemplate.DeepCopy()
	newRoute.SetName(name)
	newRoute.SetLabels(b.scaledLabels)
	newRoute.Spec.ParentRefs[0].Name = gwapiv1.ObjectName(gatewayName)
	if len(hostname) > 0 {
		newRoute.Spec.Hostnames[0] = gwapiv1.Hostname(hostname)
	}
	if len(serviceName) > 0 {
		newRoute.Spec.Rules[0].BackendRefs[0].Name = gwapiv1.ObjectName(serviceName)
	}
	return newRoute
}

// ScaleUpHTTPRoutes scales up HTTPRoutes that are all referenced to one Gateway according to
// the scale range: (a, b], which scales up from a to b with a <= b.
//
// The `afterCreation` is a callback function that only runs every time after one HTTPRoute
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
		serviceName := fmt.Sprintf(BenchmarkTestServerServiceFormat, currentBatch)

		newRoute := b.getExpectHTTPRoute(routeName, routeHostname, refGateway, serviceName)
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
// The `afterDeletion` is a callback function that only runs every time after one HTTPRoute has
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
		oldRoute := b.getExpectHTTPRoute(routeName, "", refGateway, "")
		if err := b.DeleteResource(ctx, oldRoute); err != nil {
			return err
		}

		if afterDeletion != nil {
			afterDeletion(oldRoute)
		}
	}

	return nil
}

/*
 * Deployments and Services resource processing
 */

func (b *BenchmarkTestSuite) getExpectDeployment(name string, replicas int32) *appsv1.Deployment {
	matchingLabels := map[string]string{
		"app": name,
	}
	newDeployment := b.BenchmarkServerDeployment.DeepCopy()
	newDeployment.SetName(name)
	newDeployment.SetLabels(b.scaledLabels)
	newDeployment.Spec.Selector.MatchLabels = matchingLabels
	newDeployment.Spec.Template.Labels = matchingLabels
	if replicas > 0 {
		newDeployment.Spec.Replicas = &replicas
	}
	return newDeployment
}

func (b *BenchmarkTestSuite) getExpectService(name string) *corev1.Service {
	matchingLabels := map[string]string{
		"app": name,
	}
	newService := b.BenchmarkServerService.DeepCopy()
	newService.SetName(name)
	newService.SetLabels(b.scaledLabels)
	newService.Spec.Selector = matchingLabels
	return newService
}

// ScaleUpDeployments scales up Deployments and Services according to
// the scale range: (a, b], which scales up from a to b with a <= b.
//
// The `afterCreation` is a callback function that only runs every time after one
// Deployment and Service has been created successfully.
func (b *BenchmarkTestSuite) ScaleUpDeployments(ctx context.Context, scaleRange [2]uint16, replicas int32, afterCreation func(*appsv1.Deployment, *corev1.Service)) error {
	var i, begin, end uint16
	begin, end = scaleRange[0], scaleRange[1]

	if begin > end {
		return fmt.Errorf("got wrong scale range, %d is not greater than %d", end, begin)
	}

	for i = begin + 1; i <= end; i++ {
		deploymentName := fmt.Sprintf(BenchmarkTestServerDeploymentNameFormat, i)
		serviceName := fmt.Sprintf(BenchmarkTestServerServiceFormat, i)

		newDeployment := b.getExpectDeployment(deploymentName, replicas)
		if err := b.CreateResource(ctx, newDeployment); err != nil {
			return err
		}

		newService := b.getExpectService(serviceName)
		if err := b.CreateResource(ctx, newService); err != nil {
			return err
		}

		if afterCreation != nil {
			afterCreation(newDeployment, newService)
		}
	}

	return nil
}

// ScaleDownDeployments scales down Deployments and Services according to
// the scale range: [a, b], which scales down from a to b with a > b.
//
// The `afterDeletion` is a callback function that only runs every time after one
// Deployment and Service has been deleted successfully.
func (b *BenchmarkTestSuite) ScaleDownDeployments(ctx context.Context, scaleRange [2]uint16, afterDeletion func(*appsv1.Deployment, *corev1.Service)) error {
	var i, begin, end uint16
	begin, end = scaleRange[0], scaleRange[1]

	if begin <= end {
		return fmt.Errorf("got wrong scale range, %d is not less than %d", end, begin)
	}

	for i = begin; i > end; i-- {
		deploymentName := fmt.Sprintf(BenchmarkTestServerDeploymentNameFormat, i)
		serviceName := fmt.Sprintf(BenchmarkTestServerServiceFormat, i)

		oldService := b.getExpectService(serviceName)
		if err := b.DeleteResource(ctx, oldService); err != nil {
			return err
		}

		oldDeployment := b.getExpectDeployment(deploymentName, 0)
		if err := b.DeleteResource(ctx, oldDeployment); err != nil {
			return err
		}

		if afterDeletion != nil {
			afterDeletion(oldDeployment, oldService)
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
