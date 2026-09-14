// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package runner

import (
	"context"
	"os"
	"testing"

	endpointv3 "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	envoytypes "github.com/envoyproxy/go-control-plane/pkg/cache/types"
	resourcev3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/logging"
	"github.com/envoyproxy/gateway/internal/message"
	"github.com/envoyproxy/gateway/internal/xds/cache"
	"github.com/envoyproxy/gateway/internal/xds/translator"
	xdstypes "github.com/envoyproxy/gateway/internal/xds/types"
)

type patchCall struct {
	irKey       string
	assignments []envoytypes.Resource
}

// fakeCache records snapshot and patch calls; the embedded interface is nil and
// only the overridden methods may be called.
type fakeCache struct {
	cache.SnapshotCacheWithCallbacks
	snapshots         []string
	snapshotResources []xdstypes.XdsResources
	patches           []patchCall
	patchErr          error
	// patchCommitted is what UpdateEndpointResources reports alongside
	// patchErr: true models a snapshot that went live but failed to reach
	// some node.
	patchCommitted bool
}

func (f *fakeCache) GenerateNewSnapshot(irKey string, resources xdstypes.XdsResources, _ context.Context) (bool, error) {
	f.snapshots = append(f.snapshots, irKey)
	f.snapshotResources = append(f.snapshotResources, resources)
	return true, nil
}

func (f *fakeCache) UpdateEndpointResources(irKey string, assignments []envoytypes.Resource) error {
	if f.patchErr != nil {
		if f.patchCommitted {
			f.patches = append(f.patches, patchCall{irKey: irKey, assignments: assignments})
		}
		return f.patchErr
	}
	f.patches = append(f.patches, patchCall{irKey: irKey, assignments: assignments})
	return nil
}

func testEndpointSlice(svcName string, addresses ...string) *discoveryv1.EndpointSlice {
	return &discoveryv1.EndpointSlice{
		ObjectMeta:  metav1.ObjectMeta{Namespace: "default", Name: svcName + "-abc", Labels: map[string]string{discoveryv1.LabelServiceName: svcName}},
		AddressType: discoveryv1.AddressTypeIPv4,
		Endpoints: []discoveryv1.Endpoint{{
			Addresses:  addresses,
			Conditions: discoveryv1.EndpointConditions{Ready: new(true)},
		}},
		Ports: []discoveryv1.EndpointPort{{
			Name:     new("http"),
			Protocol: new(corev1.ProtocolTCP),
			Port:     new(int32(8080)),
		}},
	}
}

func testFastPathContext() *xdstypes.EndpointContext {
	const clusterName = "cluster-a"
	return &xdstypes.EndpointContext{
		ClusterName: clusterName,
		Settings: []*ir.DestinationSetting{{
			Name:        clusterName + "/backend/0",
			Weight:      new(uint32(1)),
			Protocol:    ir.HTTP,
			Endpoints:   []*ir.DestinationEndpoint{ir.NewDestEndpoint(nil, "1.1.1.1", 8080, false, nil)},
			AddressType: new(ir.IP),
			EndpointSource: &ir.EndpointSource{
				Kind:      "Service",
				Namespace: "default",
				Name:      "svc",
				PortName:  "http",
				Protocol:  corev1.ProtocolTCP,
			},
		}},
	}
}

func testFastPath(fc *fakeCache) *endpointFastPath {
	fp := newEndpointFastPath(fc, nil, logging.DefaultLogger(os.Stderr, egv1a1.LogLevelInfo))
	table := &xdstypes.ResourceVersionTable{}
	table.AddEndpointContext(testFastPathContext())
	fp.contexts["gw/eg"] = fp.buildContexts(&ir.Xds{}, table)
	return fp
}

func testUpdate(addresses ...string) *message.EndpointUpdate {
	return &message.EndpointUpdate{
		Kind:           "Service",
		Namespace:      "default",
		Name:           "svc",
		EndpointSlices: []*discoveryv1.EndpointSlice{testEndpointSlice("svc", addresses...)},
	}
}

func claAddresses(t *testing.T, res envoytypes.Resource) []string {
	t.Helper()
	cla, ok := res.(*endpointv3.ClusterLoadAssignment)
	require.True(t, ok)
	var addrs []string
	for _, locality := range cla.GetEndpoints() {
		for _, ep := range locality.GetLbEndpoints() {
			addrs = append(addrs, ep.GetEndpoint().GetAddress().GetSocketAddress().GetAddress())
		}
	}
	return addrs
}

func TestFastPathPatchesChangedEndpoints(t *testing.T) {
	fc := &fakeCache{}
	fp := testFastPath(fc)

	fp.handleUpdate(testUpdate("2.2.2.2", "3.3.3.3"))

	require.Len(t, fc.patches, 1)
	require.Equal(t, "gw/eg", fc.patches[0].irKey)
	require.Len(t, fc.patches[0].assignments, 1)
	require.ElementsMatch(t, []string{"2.2.2.2", "3.3.3.3"}, claAddresses(t, fc.patches[0].assignments[0]))

	// The cached context tracks the applied endpoint state, so replaying the
	// same update is a no-op.
	fp.handleUpdate(testUpdate("2.2.2.2", "3.3.3.3"))
	require.Len(t, fc.patches, 1)
}

func TestFastPathSkipsZeroEndpointTransition(t *testing.T) {
	fc := &fakeCache{}
	fp := testFastPath(fc)

	update := testUpdate()
	update.EndpointSlices = nil
	fp.handleUpdate(update)

	require.Empty(t, fc.patches)
	// The cached context still holds the last-applied endpoints.
	require.Len(t, fp.contexts["gw/eg"].byBackend[update.Key()][0].Settings[0].Endpoints, 1)
}

func TestFastPathSkipsAddressTypeChange(t *testing.T) {
	fc := &fakeCache{}
	fp := testFastPath(fc)

	update := testUpdate("foo.example.com")
	update.EndpointSlices[0].AddressType = discoveryv1.AddressTypeFQDN
	fp.handleUpdate(update)

	require.Empty(t, fc.patches)
}

func TestFastPathSkipsUnknownBackend(t *testing.T) {
	fc := &fakeCache{}
	fp := testFastPath(fc)

	update := testUpdate("2.2.2.2")
	update.Name = "other-svc"
	fp.handleUpdate(update)

	require.Empty(t, fc.patches)
}

func TestFastPathDisabledForKeyWithEnvoyPatchPolicies(t *testing.T) {
	fc := &fakeCache{}
	fp := testFastPath(fc)

	// A full build whose IR carries EnvoyPatchPolicies disables the fast path
	// for the key: patches could target ClusterLoadAssignments.
	table := &xdstypes.ResourceVersionTable{}
	table.AddEndpointContext(testFastPathContext())
	fp.contexts["gw/eg"] = fp.buildContexts(&ir.Xds{EnvoyPatchPolicies: []*ir.EnvoyPatchPolicy{{}}}, table)

	fp.handleUpdate(testUpdate("2.2.2.2"))
	require.Empty(t, fc.patches)
}

// publishFullBuildTable builds a translation table whose EDS resources match
// the endpoint context captured by the build (as a real translation would).
func publishFullBuildTable(ec *xdstypes.EndpointContext) *xdstypes.ResourceVersionTable {
	table := &xdstypes.ResourceVersionTable{}
	table.AddEndpointContext(ec)
	table.XdsResources = xdstypes.XdsResources{
		resourcev3.EndpointType: {translator.BuildClusterLoadAssignment(ec)},
	}
	return table
}

func TestFastPathReappliesLatestEndpointsOnFullBuild(t *testing.T) {
	fc := &fakeCache{}
	fp := testFastPath(fc)

	// The fast path has seen endpoint state newer than what the (racing) full
	// build translated.
	fp.handleUpdate(testUpdate("9.9.9.9"))
	require.Len(t, fc.patches, 1)

	// The full build still carries the old endpoints; publishing it must patch
	// the state the fast path already published into the build's own EDS
	// resources, so the snapshot goes out fresh in a single push.
	table := publishFullBuildTable(testFastPathContext())
	require.NoError(t, fp.OnFullBuild("gw/eg", &ir.Xds{}, table, context.Background()))

	require.Equal(t, []string{"gw/eg"}, fc.snapshots)
	eds := fc.snapshotResources[0][resourcev3.EndpointType]
	require.Len(t, eds, 1)
	require.ElementsMatch(t, []string{"9.9.9.9"}, claAddresses(t, eds[0]))
	// No second push happened.
	require.Len(t, fc.patches, 1)
}

func TestFastPathKeepsContextWhenPatchCommittedWithError(t *testing.T) {
	fc := &fakeCache{}
	fp := testFastPath(fc)

	// The patched snapshot went live but publishing to a node failed. The
	// cached context must keep describing the live snapshot, so a later rebuild
	// of the same cluster (e.g. triggered by another backend it serves) does not
	// regress these endpoints.
	fc.patchErr = context.DeadlineExceeded
	fc.patchCommitted = true
	fp.handleUpdate(testUpdate("2.2.2.2"))
	require.Len(t, fc.patches, 1)

	applied := fp.contexts["gw/eg"].byBackend[testUpdate().Key()][0].Settings[0].Endpoints
	require.Len(t, applied, 1)
	require.Equal(t, "2.2.2.2", applied[0].Host)
}

func TestFastPathLeavesUncommittedPatchToTheFullPath(t *testing.T) {
	fc := &fakeCache{}
	fp := testFastPath(fc)

	// Nothing was committed. The contexts still track the new endpoints, so the
	// fast path does not retry; the full build triggered by the same event is
	// the backstop and re-applies the retained endpoints.
	fc.patchErr = context.DeadlineExceeded
	fp.handleUpdate(testUpdate("2.2.2.2"))
	require.Empty(t, fc.patches)

	table := publishFullBuildTable(testFastPathContext())
	require.NoError(t, fp.OnFullBuild("gw/eg", &ir.Xds{}, table, context.Background()))
	require.ElementsMatch(t, []string{"2.2.2.2"},
		claAddresses(t, fc.snapshotResources[0][resourcev3.EndpointType][0]))
}

func TestFastPathRetainsEndpointsForBackendsNotYetInAnyContext(t *testing.T) {
	fc := &fakeCache{}
	fp := testFastPath(fc)

	// An endpoint update arrives for a backend no committed context references
	// yet, because the build that will introduce it is still in flight.
	pending := testUpdate("7.7.7.7")
	pending.Name = "pending-svc"
	fp.handleUpdate(pending)
	require.Empty(t, fc.patches)

	// An unrelated IR key's build commits. It must not drop the retained
	// update: the in-flight build still needs it.
	other := publishFullBuildTable(testFastPathContext())
	require.NoError(t, fp.OnFullBuild("gw/other", &ir.Xds{}, other, context.Background()))
	require.Contains(t, fp.lastEndpoints, pending.Key())

	// When the in-flight build lands carrying older endpoints, the retained
	// update is patched into it rather than being lost.
	ec := testFastPathContext()
	ec.Settings[0].EndpointSource.Name = "pending-svc"
	table := publishFullBuildTable(ec)
	require.NoError(t, fp.OnFullBuild("gw/pending", &ir.Xds{}, table, context.Background()))

	eds := fc.snapshotResources[len(fc.snapshotResources)-1][resourcev3.EndpointType]
	require.Len(t, eds, 1)
	require.ElementsMatch(t, []string{"7.7.7.7"}, claAddresses(t, eds[0]))
}

func TestFastPathOnDelete(t *testing.T) {
	fc := &fakeCache{}
	fp := testFastPath(fc)

	fp.OnDelete("gw/eg")
	fp.handleUpdate(testUpdate("2.2.2.2"))
	require.Empty(t, fc.patches)
}
