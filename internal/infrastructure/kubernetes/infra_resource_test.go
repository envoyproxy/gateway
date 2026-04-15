// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	policyv1 "k8s.io/api/policy/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/ir"
	egmetrics "github.com/envoyproxy/gateway/internal/metrics"
)

// newDeleteTrackingClient builds a fake client whose interceptor counts
// DeleteAllOf invocations per concrete Go type (e.g. "*v1.DaemonSet").
func newDeleteTrackingClient(mu *sync.Mutex, counts map[string]int) client.Client {
	return fakeclient.NewClientBuilder().
		WithScheme(envoygateway.GetScheme()).
		WithInterceptorFuncs(interceptor.Funcs{
			Patch: interceptorFunc.Patch,
			DeleteAllOf: func(ctx context.Context, clnt client.WithWatch, obj client.Object, opts ...client.DeleteAllOfOption) error {
				kind := fmt.Sprintf("%T", obj)
				mu.Lock()
				counts[kind]++
				mu.Unlock()
				return clnt.DeleteAllOf(ctx, obj, opts...)
			},
		}).
		Build()
}

func setupDeleteMetricsRecorder(t *testing.T) *sdkmetric.ManualReader {
	t.Helper()

	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))

	previousProvider := otel.GetMeterProvider()
	previousDeleteTotal := resourceDeleteTotal
	previousDeleteDuration := resourceDeleteDurationSeconds

	otel.SetMeterProvider(provider)
	resourceDeleteTotal = egmetrics.NewCounter(
		"resource_delete_total",
		"Total number of deleted resources.",
	)
	resourceDeleteDurationSeconds = egmetrics.NewHistogram(
		"resource_delete_duration_seconds",
		"How long in seconds a resource be deleted successfully.",
		[]float64{0.001, 0.01, 0.1, 1, 5, 10},
	)

	t.Cleanup(func() {
		resourceDeleteTotal = previousDeleteTotal
		resourceDeleteDurationSeconds = previousDeleteDuration
		otel.SetMeterProvider(previousProvider)
	})

	return reader
}

func collectDeleteSuccessTotalForKind(t *testing.T, reader *sdkmetric.ManualReader, kind string) float64 {
	t.Helper()

	var rm metricdata.ResourceMetrics
	require.NoError(t, reader.Collect(context.Background(), &rm))

	var total float64
	for _, scopeMetric := range rm.ScopeMetrics {
		for _, metric := range scopeMetric.Metrics {
			if metric.Name != "resource_delete_total" {
				continue
			}

			sum, ok := metric.Data.(metricdata.Sum[float64])
			require.True(t, ok, "resource_delete_total should export a float64 sum")

			for _, point := range sum.DataPoints {
				if attributeValue(point.Attributes, "kind") == kind && attributeValue(point.Attributes, "status") == egmetrics.StatusSuccess {
					total += point.Value
				}
			}
		}
	}

	return total
}

func attributeValue(set attribute.Set, key string) string {
	for _, kv := range set.ToSlice() {
		if string(kv.Key) == key {
			return kv.Value.AsString()
		}
	}

	return ""
}

var sharedTestLabels = map[string]string{
	"app.kubernetes.io/name":                         "envoy",
	"app.kubernetes.io/component":                    "proxy",
	"app.kubernetes.io/managed-by":                   "envoy-gateway",
	"gateway.envoyproxy.io/owning-gateway-namespace": "default",
	"gateway.envoyproxy.io/owning-gateway-name":      "test-gw",
}

// standardDeploymentInfra returns an ir.Infra in Deployment mode
// (no DaemonSet, no HPA, no PDB configured).
func standardDeploymentInfra() *ir.Infra {
	infra := ir.NewInfra()
	infra.GetProxyInfra().GetProxyMetadata().Labels = sharedTestLabels
	infra.GetProxyInfra().GetProxyMetadata().OwnerReference = &ir.ResourceMetadata{
		Kind: resource.KindGateway,
		Name: testGatewayClass,
	}
	return infra
}

// daemonSetModeInfra returns an ir.Infra in DaemonSet mode
// (EnvoyDaemonSet configured, EnvoyDeployment is nil, no HPA, no PDB).
func daemonSetModeInfra() *ir.Infra {
	infra := ir.NewInfra()
	infra.GetProxyInfra().GetProxyMetadata().Labels = sharedTestLabels
	infra.GetProxyInfra().GetProxyMetadata().OwnerReference = &ir.ResourceMetadata{
		Kind: resource.KindGateway,
		Name: testGatewayClass,
	}
	infra.GetProxyInfra().Config = &egv1a1.EnvoyProxy{
		Spec: egv1a1.EnvoyProxySpec{
			Provider: &egv1a1.EnvoyProxyProvider{
				Type: egv1a1.EnvoyProxyProviderTypeKubernetes,
				Kubernetes: &egv1a1.EnvoyProxyKubernetesProvider{
					// DaemonSet mode: Deployment() returns nil, DaemonSet() returns non-nil.
					EnvoyDaemonSet: egv1a1.DefaultKubernetesDaemonSet(egv1a1.DefaultEnvoyProxyImage),
					EnvoyService:   egv1a1.DefaultKubernetesService(),
				},
			},
		},
	}
	return infra
}

// TestNoopDeleteMetricsSuppressedInDeploymentMode verifies that in Deployment mode,
// optional-resource cleanup does not record success delete metrics when those
// resources do not exist.
func TestNoopDeleteMetricsSuppressedInDeploymentMode(t *testing.T) {
	reader := setupDeleteMetricsRecorder(t)
	cli := newDeleteTrackingClient(&sync.Mutex{}, make(map[string]int))
	kube := newTestInfraWithClient(t, cli)
	require.NoError(t, setupOwnerReferenceResources(context.Background(), kube.Client))

	err := kube.CreateOrUpdateProxyInfra(context.Background(), standardDeploymentInfra())
	require.NoError(t, err)

	assert.Zero(t, collectDeleteSuccessTotalForKind(t, reader, "DaemonSet"),
		"no-op reconcile should not record a successful delete metric for DaemonSet")
	assert.Zero(t, collectDeleteSuccessTotalForKind(t, reader, "HPA"),
		"no-op reconcile should not record a successful delete metric for HPA")
	assert.Zero(t, collectDeleteSuccessTotalForKind(t, reader, "PDB"),
		"no-op reconcile should not record a successful delete metric for PDB")
}

// TestNoopDeleteMetricsSuppressedInDaemonSetMode verifies that in DaemonSet mode,
// optional-resource cleanup does not record success delete metrics when those
// resources do not exist.
func TestNoopDeleteMetricsSuppressedInDaemonSetMode(t *testing.T) {
	reader := setupDeleteMetricsRecorder(t)
	cli := newDeleteTrackingClient(&sync.Mutex{}, make(map[string]int))
	kube := newTestInfraWithClient(t, cli)
	require.NoError(t, setupOwnerReferenceResources(context.Background(), kube.Client))

	err := kube.CreateOrUpdateProxyInfra(context.Background(), daemonSetModeInfra())
	require.NoError(t, err)

	assert.Zero(t, collectDeleteSuccessTotalForKind(t, reader, "Deployment"),
		"no-op reconcile should not record a successful delete metric for Deployment")
	assert.Zero(t, collectDeleteSuccessTotalForKind(t, reader, "HPA"),
		"no-op reconcile should not record a successful delete metric for HPA")
	assert.Zero(t, collectDeleteSuccessTotalForKind(t, reader, "PDB"),
		"no-op reconcile should not record a successful delete metric for PDB")
}

// TestDeleteMetricsRecordedForStaleResourcesInDeploymentMode verifies that in
// Deployment mode, stale DaemonSet/HPA/PDB cleanup records successful delete
// metrics when those resources actually exist.
func TestDeleteMetricsRecordedForStaleResourcesInDeploymentMode(t *testing.T) {
	reader := setupDeleteMetricsRecorder(t)
	cli := newDeleteTrackingClient(&sync.Mutex{}, make(map[string]int))
	kube := newTestInfraWithClient(t, cli)
	ctx := context.Background()
	require.NoError(t, setupOwnerReferenceResources(ctx, kube.Client))

	// Pre-create stale resources from a previous DaemonSet-mode configuration.
	ds := &appsv1.DaemonSet{}
	ds.Name = "envoy-default-test-gw-abc123"
	ds.Namespace = kube.ControllerNamespace
	ds.Labels = sharedTestLabels
	require.NoError(t, cli.Create(ctx, ds))

	hpa := &autoscalingv2.HorizontalPodAutoscaler{}
	hpa.Name = "envoy-default-test-gw-abc123"
	hpa.Namespace = kube.ControllerNamespace
	hpa.Labels = sharedTestLabels
	require.NoError(t, cli.Create(ctx, hpa))

	pdb := &policyv1.PodDisruptionBudget{}
	pdb.Name = "envoy-default-test-gw-abc123"
	pdb.Namespace = kube.ControllerNamespace
	pdb.Labels = sharedTestLabels
	require.NoError(t, cli.Create(ctx, pdb))

	// Reconcile in Deployment mode → should DELETE the stale DaemonSet/HPA/PDB.
	err := kube.CreateOrUpdateProxyInfra(ctx, standardDeploymentInfra())
	require.NoError(t, err)

	assert.Positive(t, collectDeleteSuccessTotalForKind(t, reader, "DaemonSet"),
		"stale DaemonSet cleanup should record a successful delete metric")
	assert.Positive(t, collectDeleteSuccessTotalForKind(t, reader, "HPA"),
		"stale HPA cleanup should record a successful delete metric")
	assert.Positive(t, collectDeleteSuccessTotalForKind(t, reader, "PDB"),
		"stale PDB cleanup should record a successful delete metric")
}

// TestDeleteMetricsRecordedForStaleResourcesInDaemonSetMode verifies that in
// DaemonSet mode, stale Deployment cleanup records a successful delete metric
// when the resource actually exists.
func TestDeleteMetricsRecordedForStaleResourcesInDaemonSetMode(t *testing.T) {
	reader := setupDeleteMetricsRecorder(t)
	cli := newDeleteTrackingClient(&sync.Mutex{}, make(map[string]int))
	kube := newTestInfraWithClient(t, cli)
	ctx := context.Background()
	require.NoError(t, setupOwnerReferenceResources(ctx, kube.Client))

	// Pre-create a stale Deployment from a previous Deployment-mode configuration.
	deploy := &appsv1.Deployment{}
	deploy.Name = "envoy-default-test-gw-abc123"
	deploy.Namespace = kube.ControllerNamespace
	deploy.Labels = sharedTestLabels
	require.NoError(t, cli.Create(ctx, deploy))

	// Reconcile in DaemonSet mode → should DELETE the stale Deployment.
	err := kube.CreateOrUpdateProxyInfra(ctx, daemonSetModeInfra())
	require.NoError(t, err)

	assert.Positive(t, collectDeleteSuccessTotalForKind(t, reader, "Deployment"),
		"stale Deployment cleanup should record a successful delete metric")
}

// TestNoopDeleteMetricsStaySuppressedAcrossDeploymentReconciles is a regression test that runs
// CreateOrUpdateProxyInfra multiple times in Deployment mode and verifies
// that repeated no-op reconciles still do not emit successful delete metrics
// for optional resources that were already absent.
func TestNoopDeleteMetricsStaySuppressedAcrossDeploymentReconciles(t *testing.T) {
	reader := setupDeleteMetricsRecorder(t)
	cli := newDeleteTrackingClient(&sync.Mutex{}, make(map[string]int))
	kube := newTestInfraWithClient(t, cli)
	ctx := context.Background()
	require.NoError(t, setupOwnerReferenceResources(ctx, kube.Client))

	infra := standardDeploymentInfra()

	// Simulate 3 consecutive reconcile loops.
	for i := 0; i < 3; i++ {
		require.NoError(t, kube.CreateOrUpdateProxyInfra(ctx, infra))
	}

	assert.Zero(t, collectDeleteSuccessTotalForKind(t, reader, "DaemonSet"),
		"repeated no-op reconciles should not record successful delete metrics for DaemonSet")
	assert.Zero(t, collectDeleteSuccessTotalForKind(t, reader, "HPA"),
		"repeated no-op reconciles should not record successful delete metrics for HPA")
	assert.Zero(t, collectDeleteSuccessTotalForKind(t, reader, "PDB"),
		"repeated no-op reconciles should not record successful delete metrics for PDB")
}

// TestNoopDeleteMetricsStaySuppressedAcrossDaemonSetReconciles is a regression test that runs
// CreateOrUpdateProxyInfra multiple times in DaemonSet mode and verifies
// that repeated no-op reconciles still do not emit successful delete metrics
// for optional resources that were already absent.
func TestNoopDeleteMetricsStaySuppressedAcrossDaemonSetReconciles(t *testing.T) {
	reader := setupDeleteMetricsRecorder(t)
	cli := newDeleteTrackingClient(&sync.Mutex{}, make(map[string]int))
	kube := newTestInfraWithClient(t, cli)
	ctx := context.Background()
	require.NoError(t, setupOwnerReferenceResources(ctx, kube.Client))

	infra := daemonSetModeInfra()

	// Simulate 3 consecutive reconcile loops.
	for i := 0; i < 3; i++ {
		require.NoError(t, kube.CreateOrUpdateProxyInfra(ctx, infra))
	}

	assert.Zero(t, collectDeleteSuccessTotalForKind(t, reader, "Deployment"),
		"repeated no-op reconciles should not record successful delete metrics for Deployment")
	assert.Zero(t, collectDeleteSuccessTotalForKind(t, reader, "HPA"),
		"repeated no-op reconciles should not record successful delete metrics for HPA")
	assert.Zero(t, collectDeleteSuccessTotalForKind(t, reader, "PDB"),
		"repeated no-op reconciles should not record successful delete metrics for PDB")
}
