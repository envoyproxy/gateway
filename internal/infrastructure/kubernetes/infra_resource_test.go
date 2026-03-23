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

// TestDeleteNoopReconcileDeploymentMode verifies that in Deployment mode,
// DeleteAllOf is NOT called for DaemonSet/HPA/PDB since they never existed.
// This is the primary scenario reported in https://github.com/envoyproxy/gateway/issues/8438.
func TestDeleteNoopReconcileDeploymentMode(t *testing.T) {
	var (
		mu     sync.Mutex
		counts = make(map[string]int)
	)

	cli := newDeleteTrackingClient(&mu, counts)
	kube := newTestInfraWithClient(t, cli)
	require.NoError(t, setupOwnerReferenceResources(context.Background(), kube.Client))

	err := kube.CreateOrUpdateProxyInfra(context.Background(), standardDeploymentInfra())
	require.NoError(t, err)

	mu.Lock()
	defer mu.Unlock()

	assert.Equal(t, 0, counts["*v1.DaemonSet"],
		"DeleteAllOf should not be called for DaemonSet when none exist")
	assert.Equal(t, 0, counts["*v2.HorizontalPodAutoscaler"],
		"DeleteAllOf should not be called for HPA when none exist")
	assert.Equal(t, 0, counts["*v1.PodDisruptionBudget"],
		"DeleteAllOf should not be called for PDB when none exist")
}

// TestDeleteNoopReconcileDaemonSetMode verifies that in DaemonSet mode,
// DeleteAllOf is NOT called for Deployment/HPA/PDB since they never existed.
// This exercises the deleteDeployment fix (Deployment is nil in DaemonSet mode).
func TestDeleteNoopReconcileDaemonSetMode(t *testing.T) {
	var (
		mu     sync.Mutex
		counts = make(map[string]int)
	)

	cli := newDeleteTrackingClient(&mu, counts)
	kube := newTestInfraWithClient(t, cli)
	require.NoError(t, setupOwnerReferenceResources(context.Background(), kube.Client))

	err := kube.CreateOrUpdateProxyInfra(context.Background(), daemonSetModeInfra())
	require.NoError(t, err)

	mu.Lock()
	defer mu.Unlock()

	assert.Equal(t, 0, counts["*v1.Deployment"],
		"DeleteAllOf should not be called for Deployment when none exist (DaemonSet mode)")
	assert.Equal(t, 0, counts["*v2.HorizontalPodAutoscaler"],
		"DeleteAllOf should not be called for HPA when none exist")
	assert.Equal(t, 0, counts["*v1.PodDisruptionBudget"],
		"DeleteAllOf should not be called for PDB when none exist")
}

// TestDeleteIsCalledWhenResourcesExistDeploymentMode verifies that in Deployment mode,
// DeleteAllOf IS called for DaemonSet/HPA/PDB when they actually exist
// (e.g. switching from DaemonSet to Deployment mode).
func TestDeleteIsCalledWhenResourcesExistDeploymentMode(t *testing.T) {
	var (
		mu     sync.Mutex
		counts = make(map[string]int)
	)

	cli := newDeleteTrackingClient(&mu, counts)
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

	mu.Lock()
	defer mu.Unlock()

	assert.Positive(t, counts["*v1.DaemonSet"],
		"DeleteAllOf should be called for DaemonSet when it exists")
	assert.Positive(t, counts["*v2.HorizontalPodAutoscaler"],
		"DeleteAllOf should be called for HPA when it exists")
	assert.Positive(t, counts["*v1.PodDisruptionBudget"],
		"DeleteAllOf should be called for PDB when it exists")
}

// TestDeleteIsCalledWhenResourcesExistDaemonSetMode verifies that in DaemonSet mode,
// DeleteAllOf IS called for a stale Deployment when it actually exists
// (e.g. switching from Deployment to DaemonSet mode).
func TestDeleteIsCalledWhenResourcesExistDaemonSetMode(t *testing.T) {
	var (
		mu     sync.Mutex
		counts = make(map[string]int)
	)

	cli := newDeleteTrackingClient(&mu, counts)
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

	mu.Lock()
	defer mu.Unlock()

	assert.Positive(t, counts["*v1.Deployment"],
		"DeleteAllOf should be called for Deployment when it exists (DaemonSet mode)")
}

// TestReconcileIdempotencyDeploymentMode is a regression test that runs
// CreateOrUpdateProxyInfra multiple times in Deployment mode and verifies
// that DeleteAllOf is never called for optional resources.
func TestReconcileIdempotencyDeploymentMode(t *testing.T) {
	var (
		mu     sync.Mutex
		counts = make(map[string]int)
	)

	cli := newDeleteTrackingClient(&mu, counts)
	kube := newTestInfraWithClient(t, cli)
	ctx := context.Background()
	require.NoError(t, setupOwnerReferenceResources(ctx, kube.Client))

	infra := standardDeploymentInfra()

	// Simulate 3 consecutive reconcile loops.
	for i := 0; i < 3; i++ {
		require.NoError(t, kube.CreateOrUpdateProxyInfra(ctx, infra))
	}

	mu.Lock()
	defer mu.Unlock()

	assert.Equal(t, 0, counts["*v1.DaemonSet"],
		"DaemonSet DeleteAllOf must never be called across multiple no-op reconciles")
	assert.Equal(t, 0, counts["*v2.HorizontalPodAutoscaler"],
		"HPA DeleteAllOf must never be called across multiple no-op reconciles")
	assert.Equal(t, 0, counts["*v1.PodDisruptionBudget"],
		"PDB DeleteAllOf must never be called across multiple no-op reconciles")
}

// TestReconcileIdempotencyDaemonSetMode is a regression test that runs
// CreateOrUpdateProxyInfra multiple times in DaemonSet mode and verifies
// that DeleteAllOf is never called for optional resources.
func TestReconcileIdempotencyDaemonSetMode(t *testing.T) {
	var (
		mu     sync.Mutex
		counts = make(map[string]int)
	)

	cli := newDeleteTrackingClient(&mu, counts)
	kube := newTestInfraWithClient(t, cli)
	ctx := context.Background()
	require.NoError(t, setupOwnerReferenceResources(ctx, kube.Client))

	infra := daemonSetModeInfra()

	// Simulate 3 consecutive reconcile loops.
	for i := 0; i < 3; i++ {
		require.NoError(t, kube.CreateOrUpdateProxyInfra(ctx, infra))
	}

	mu.Lock()
	defer mu.Unlock()

	assert.Equal(t, 0, counts["*v1.Deployment"],
		"Deployment DeleteAllOf must never be called across multiple no-op reconciles (DaemonSet mode)")
	assert.Equal(t, 0, counts["*v2.HorizontalPodAutoscaler"],
		"HPA DeleteAllOf must never be called across multiple no-op reconciles")
	assert.Equal(t, 0, counts["*v1.PodDisruptionBudget"],
		"PDB DeleteAllOf must never be called across multiple no-op reconciles")
}
