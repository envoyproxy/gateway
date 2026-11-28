// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"fmt"
	"os"
	"testing"

	endpointv3 "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	resourcev3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	mcsapiv1a1 "sigs.k8s.io/mcs-api/pkg/apis/v1alpha1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/infrastructure/kubernetes/proxy"
	"github.com/envoyproxy/gateway/internal/message"
	"github.com/envoyproxy/gateway/internal/provider/kubernetes/test"
	xdstranslator "github.com/envoyproxy/gateway/internal/xds/translator"
)

func TestClusterPlatform(t *testing.T) {
	platform := &clusterPlatform{}
	node := &corev1.Node{Spec: corev1.NodeSpec{ProviderID: "aws:///us-east-1d/i-123"}}
	require.Empty(t, platform.zoneID(node))
	require.True(t, platform.aws.Load())

	// Nodes supply their own IDs after the controller detects AWS.
	for _, zoneID := range []string{"use1-az6", "use1-az2", ""} {
		node = &corev1.Node{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{awsZoneIDLabel: zoneID}}}
		require.Equal(t, zoneID, platform.zoneID(node))
	}
}

func TestApplyAWSZoneIDs(t *testing.T) {
	local := &discoveryv1.EndpointSlice{
		ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{discoveryv1.LabelManagedBy: endpointSliceControllerName}},
		Endpoints: []discoveryv1.Endpoint{
			{NodeName: new("node"), Zone: new("us-east-1d")},
			{NodeName: new("node")},
			{NodeName: new("no-id"), Zone: new("us-east-1b")},
			{NodeName: new("empty-id"), Zone: new("us-east-1b")},
			{NodeName: new("missing"), Zone: new("us-east-1c")},
			{Zone: new("explicit")},
		},
	}
	local.Labels[mcsapiv1a1.LabelServiceName] = "also-exported"
	imported := local.DeepCopy()
	imported.Labels[mcsapiv1a1.LabelServiceName] = "remote"
	imported.Labels[discoveryv1.LabelManagedBy] = "mcs-controller"
	custom := local.DeepCopy()
	custom.Labels[discoveryv1.LabelManagedBy] = "custom-controller"
	mirrored := local.DeepCopy()
	mirrored.Labels[discoveryv1.LabelManagedBy] = endpointSliceMirroringControllerName
	original := local.DeepCopy()
	resources := resource.ControllerResources{
		{EndpointSlices: []*discoveryv1.EndpointSlice{local, imported, custom, mirrored}},
		{EndpointSlices: []*discoveryv1.EndpointSlice{local}},
	}
	store := newProviderStore()
	r := &gatewayAPIReconciler{store: store}
	r.applyAWSZoneIDs(resources)
	require.Same(t, local, resources[0].EndpointSlices[0])
	for _, node := range []*corev1.Node{
		{ObjectMeta: metav1.ObjectMeta{Name: "node", Labels: map[string]string{awsZoneIDLabel: "use1-az6"}}},
		{ObjectMeta: metav1.ObjectMeta{Name: "no-id"}},
		{ObjectMeta: metav1.ObjectMeta{Name: "empty-id", Labels: map[string]string{awsZoneIDLabel: ""}}},
	} {
		store.addNode(node)
	}
	r.applyAWSZoneIDs(resources)
	for _, res := range resources {
		require.Equal(t, "use1-az6", *res.EndpointSlices[0].Endpoints[0].Zone)
		require.Equal(t, "use1-az6", *res.EndpointSlices[0].Endpoints[1].Zone)
		require.Equal(t, original.Endpoints[2:], res.EndpointSlices[0].Endpoints[2:])
	}
	require.Equal(t, original, local)
	require.Equal(t, "use1-az6", *resources[0].EndpointSlices[3].Endpoints[0].Zone)
	require.Same(t, imported, resources[0].EndpointSlices[1])
	require.Same(t, custom, resources[0].EndpointSlices[2])
	require.Equal(t, original.Endpoints, imported.Endpoints)
	require.Equal(t, original.Endpoints, custom.Endpoints)
}

func TestAWSZoneIDNodeUpdates(t *testing.T) {
	r := &gatewayAPIReconciler{}
	old := &corev1.Node{ObjectMeta: metav1.ObjectMeta{Name: "node"}}
	updated := old.DeepCopy()
	updated.Labels = map[string]string{awsZoneIDLabel: "use1-az6"}
	for _, nodes := range [][2]*corev1.Node{{old, updated}, {updated, old}} {
		require.True(t, r.nodePredicate().Update(event.TypedUpdateEvent[*corev1.Node]{ObjectOld: nodes[0], ObjectNew: nodes[1]}))
	}
	require.False(t, r.nodePredicate().Update(event.TypedUpdateEvent[*corev1.Node]{ObjectOld: updated, ObjectNew: updated}))
	old = updated.DeepCopy()
	updated.Spec.ProviderID = "aws:///us-east-1d/i-123"
	require.True(t, r.nodePredicate().Update(event.TypedUpdateEvent[*corev1.Node]{ObjectOld: old, ObjectNew: updated}))
	old = updated.DeepCopy()
	updated.Generation++
	require.True(t, r.nodePredicate().Update(event.TypedUpdateEvent[*corev1.Node]{ObjectOld: old, ObjectNew: updated}))
}

func TestAWSZoneIDLocalityXDS(t *testing.T) {
	cfg, err := config.New(os.Stdout, os.Stderr)
	require.NoError(t, err)
	cfg.EnvoyGateway.Provider = &egv1a1.EnvoyGatewayProvider{Type: egv1a1.ProviderTypeCustom}
	updates := new(message.ProviderResources)
	r, err := NewOfflineGatewayAPIController(t.Context(), cfg, nil, updates)
	require.NoError(t, err)
	defer updates.Close()
	gateway := test.GetGateway(types.NamespacedName{Namespace: "default", Name: "gateway"}, "eg", 80)
	fleet := test.GetService(types.NamespacedName{
		Namespace: cfg.ControllerNamespace, Name: proxy.ExpectedResourceHashedName("default/gateway"),
	}, gatewayapi.OwnerLabels(gateway, false), map[string]int32{"dummy": 8080})
	backend := test.GetService(types.NamespacedName{Namespace: "default", Name: "backend"}, nil, map[string]int32{"dummy": 8080})
	backend.Spec.TrafficDistribution = new("PreferClose")
	fleetSlice := test.GetEndpointSlice(types.NamespacedName{Namespace: fleet.Namespace, Name: "fleet"}, fleet.Name, false)
	backendSlice := test.GetEndpointSlice(types.NamespacedName{Namespace: backend.Namespace, Name: "backend"}, backend.Name, false)
	for i, slice := range []*discoveryv1.EndpointSlice{fleetSlice, backendSlice} {
		slice.AddressType = discoveryv1.AddressTypeIPv4
		slice.Labels[discoveryv1.LabelManagedBy] = endpointSliceControllerName
		slice.Endpoints[0].NodeName = new("node")
		slice.Endpoints[0].Zone = new("us-east-1d")
		slice.Endpoints[0].Addresses = []string{fmt.Sprintf("192.0.2.%d", i+1)}
	}
	remoteSlice := backendSlice.DeepCopy()
	remoteSlice.Name = "remote"
	remoteSlice.Labels[discoveryv1.LabelManagedBy] = "remote-controller"
	remoteSlice.Endpoints[0].Zone = new("use1-az2")
	remoteSlice.Endpoints[0].Addresses = []string{"192.0.2.3"}
	node := &corev1.Node{ObjectMeta: metav1.ObjectMeta{Name: "node", Labels: map[string]string{
		corev1.LabelTopologyZone: "us-east-1d", awsZoneIDLabel: "use1-az6",
	}}}
	for _, obj := range []client.Object{
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "default"}},
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: cfg.ControllerNamespace}},
		test.GetGatewayClass("eg", egv1a1.GatewayControllerName, nil), gateway, fleet, backend,
		fleetSlice, backendSlice, remoteSlice, node,
		test.GetHTTPRoute(types.NamespacedName{Namespace: "default", Name: "route"}, gateway.Name,
			test.GetServiceBackendRef(types.NamespacedName{Name: backend.Name}, 8080), ""),
	} {
		require.NoError(t, r.Client.Create(t.Context(), obj))
	}
	r.store.addNode(node)
	require.NoError(t, r.Reconcile(t.Context()))
	published, ok := updates.GatewayAPIResources.Load(cfg.EnvoyGateway.Gateway.ControllerName)
	require.True(t, ok)
	require.Len(t, *published.Resources, 1)
	gt := &gatewayapi.Translator{
		GatewayControllerName: egv1a1.GatewayControllerName, GatewayClassName: "eg",
		ControllerNamespace: cfg.ControllerNamespace, Logger: cfg.Logger,
	}
	translated, err := gt.Translate(t.Context(), (*published.Resources)[0])
	require.NoError(t, err)
	require.Len(t, translated.XdsIR, 1)
	zones := map[string]string{}
	for _, xdsIR := range translated.XdsIR {
		xt := &xdstranslator.Translator{Logger: cfg.Logger}
		table, err := xt.Translate(t.Context(), xdsIR)
		require.NoError(t, err)
		for _, entry := range table.XdsResources[resourcev3.EndpointType] {
			for _, locality := range entry.(*endpointv3.ClusterLoadAssignment).Endpoints {
				for _, endpoint := range locality.LbEndpoints {
					zones[endpoint.GetEndpoint().Address.GetSocketAddress().Address] = locality.Locality.GetZone()
				}
			}
		}
	}
	require.Equal(t, map[string]string{
		"192.0.2.1": "use1-az6", "192.0.2.2": "use1-az6", "192.0.2.3": "use1-az2",
	}, zones)
	var stored discoveryv1.EndpointSlice
	require.NoError(t, r.Client.Get(t.Context(), client.ObjectKeyFromObject(backendSlice), &stored))
	require.Equal(t, "us-east-1d", *stored.Endpoints[0].Zone)
}
