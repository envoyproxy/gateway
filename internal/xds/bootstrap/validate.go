// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package bootstrap

import (
	"fmt"

	bootstrapv3 "github.com/envoyproxy/go-control-plane/envoy/config/bootstrap/v3"
	clusterv3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/utils/proto"
)

func fetchAndPatchBootstrap(boostrapConfig *egv1a1.ProxyBootstrap) (*bootstrapv3.Bootstrap, *bootstrapv3.Bootstrap, error) {
	// Use default configuration for validation
	defaultOpts := &RenderBootstrapConfigOptions{
		ServiceName: "envoy-gateway", // Use default service name for validation
	}
	defaultBootstrapStr, err := GetRenderedBootstrapConfig(defaultOpts)
	if err != nil {
		return nil, nil, err
	}
	defaultBootstrap := &bootstrapv3.Bootstrap{}
	if err := proto.FromYAML([]byte(defaultBootstrapStr), defaultBootstrap); err != nil {
		return nil, nil, fmt.Errorf("unable to unmarshal default bootstrap: %w", err)
	}
	if err := defaultBootstrap.Validate(); err != nil {
		return nil, nil, fmt.Errorf("default bootstrap validation failed: %w", err)
	}
	// Validate user bootstrap config
	patchedYaml, err := ApplyBootstrapConfig(boostrapConfig, defaultBootstrapStr)
	if err != nil {
		return nil, nil, err
	}
	patchedBootstrap := &bootstrapv3.Bootstrap{}
	if err := proto.FromYAML([]byte(patchedYaml), patchedBootstrap); err != nil {
		return nil, nil, fmt.Errorf("unable to unmarshal user bootstrap: %w", err)
	}
	if err := patchedBootstrap.Validate(); err != nil {
		return nil, nil, fmt.Errorf("validation failed for user bootstrap: %w", err)
	}
	return patchedBootstrap, defaultBootstrap, err
}

// Validate ensures that after applying the provided bootstrap configuration, the resulting
// bootstrap is still OK.
// This code previously was part of the validate logic in api/v1alpha1/validate, but was moved
// here to prevent code in the api packages from accessing code from the internal packages.
func Validate(boostrapConfig *egv1a1.ProxyBootstrap) error {
	if boostrapConfig == nil {
		return nil
	}
	// Validate user bootstrap config
	// TODO: need validate when enable prometheus?
	userBootstrap, defaultBootstrap, err := fetchAndPatchBootstrap(boostrapConfig)
	if err != nil {
		return err
	}

	// Ensure dynamic resources config is same
	if userBootstrap.DynamicResources == nil ||
		cmp.Diff(userBootstrap.DynamicResources, defaultBootstrap.DynamicResources, protocmp.Transform()) != "" {
		return fmt.Errorf("dynamic_resources cannot be modified")
	}

	// Ensure that the xds_cluster config is same
	var userXdsCluster, defaultXdsCluster *clusterv3.Cluster
	for _, cluster := range userBootstrap.StaticResources.Clusters {
		if cluster.Name == "xds_cluster" {
			userXdsCluster = cluster
			break
		}
	}
	for _, cluster := range defaultBootstrap.StaticResources.Clusters {
		if cluster.Name == "xds_cluster" {
			defaultXdsCluster = cluster
			break
		}
	}
	if userXdsCluster == nil ||
		cmp.Diff(userXdsCluster.LoadAssignment, defaultXdsCluster.LoadAssignment, protocmp.Transform()) != "" {
		return fmt.Errorf("xds_cluster's loadAssigntment cannot be modified")
	}

	return nil
}
