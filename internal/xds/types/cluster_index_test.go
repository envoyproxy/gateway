// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package types

import (
	"testing"

	clusterv3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	cachetypes "github.com/envoyproxy/go-control-plane/pkg/cache/types"
	resourcev3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/stretchr/testify/require"
)

func TestClusterIndexFollowsResourceChanges(t *testing.T) {
	table := new(ResourceVersionTable)
	table.EnableClusterIndex()
	require.Nil(t, table.FindCluster("shared"))
	first := &clusterv3.Cluster{Name: "shared"}
	second := &clusterv3.Cluster{Name: "shared"}
	require.NoError(t, table.AddXdsResource(resourcev3.ClusterType, first))
	require.NoError(t, table.AddXdsResource(resourcev3.ClusterType, second))
	require.Same(t, first, table.FindCluster("shared"), "preserve first-match lookup semantics")

	renamed := &clusterv3.Cluster{Name: "renamed"}
	require.NoError(t, table.AddOrReplaceXdsResource(resourcev3.ClusterType, renamed,
		func(existing, _ cachetypes.Resource) bool { return existing == first }))
	require.Same(t, second, table.FindCluster("shared"))
	require.Same(t, renamed, table.FindCluster("renamed"))

	replacement := &clusterv3.Cluster{Name: "replacement"}
	table.SetResources(resourcev3.ClusterType, []cachetypes.Resource{replacement})
	require.Nil(t, table.FindCluster("shared"))
	require.Nil(t, table.FindCluster("renamed"))
	require.Same(t, replacement, table.FindCluster("replacement"))
	require.Error(t, table.AddXdsResource(resourcev3.ClusterType, &clusterv3.Cluster{}))
	require.Nil(t, table.FindCluster(""), "a rejected resource must not enter the index")
	require.Len(t, table.XdsResources[resourcev3.ClusterType], 1)

	// Patching and extension phases may modify resource names in place.
	table.DisableClusterIndex()
	replacement.Name = "patched"
	require.Nil(t, table.FindCluster("replacement"))
	require.Same(t, replacement, table.FindCluster("patched"))

	// Enabling an index on existing resources must include them without changing order.
	table.EnableClusterIndex()
	require.Same(t, replacement, table.FindCluster("patched"))
	table.SetResources(resourcev3.ClusterType, nil)
	require.Nil(t, table.FindCluster("patched"))
}
