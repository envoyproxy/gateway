// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package validation

import (
	"errors"
	"fmt"
	"reflect"

	bootstrapv3 "github.com/envoyproxy/go-control-plane/envoy/config/bootstrap/v3"
	clusterv3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	"google.golang.org/protobuf/encoding/protojson"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"sigs.k8s.io/yaml"

	egcfgv1a1 "github.com/envoyproxy/gateway/api/config/v1alpha1"
	"github.com/envoyproxy/gateway/internal/xds/bootstrap"
	_ "github.com/envoyproxy/gateway/internal/xds/extensions" // register the generated types to support protojson unmarshalling
)

// ValidateEnvoyProxy validates the provided EnvoyProxy.
func ValidateEnvoyProxy(ep *egcfgv1a1.EnvoyProxy) error {
	var errs []error
	if ep == nil {
		return errors.New("envoyproxy is nil")
	}
	if err := validateEnvoyProxySpec(&ep.Spec); err != nil {
		errs = append(errs, err)
	}

	return utilerrors.NewAggregate(errs)
}

// validateEnvoyProxySpec validates the provided EnvoyProxy spec.
func validateEnvoyProxySpec(spec *egcfgv1a1.EnvoyProxySpec) error {
	var errs []error

	if spec == nil {
		errs = append(errs, errors.New("spec is nil"))
	}
	if spec != nil && spec.Provider != nil && spec.Provider.Type != egcfgv1a1.ProviderTypeKubernetes {
		errs = append(errs, fmt.Errorf("unsupported provider type %v", spec.Provider.Type))
	}
	if spec.Bootstrap != nil {
		if err := validateBootstrap(spec.Bootstrap); err != nil {
			errs = append(errs, err)
		}
	}
	return utilerrors.NewAggregate(errs)
}

func validateBootstrap(bstrap *string) error {
	userBootstrap := &bootstrapv3.Bootstrap{}
	jsonData, err := yaml.YAMLToJSON([]byte(*bstrap))
	if err != nil {
		return fmt.Errorf("unable to convert user bootstrap to json: %w", err)
	}

	if err := protojson.Unmarshal(jsonData, userBootstrap); err != nil {
		return fmt.Errorf("unable to unmarshal user bootstrap: %w", err)
	}

	// Call Validate method
	if err := userBootstrap.Validate(); err != nil {
		return fmt.Errorf("validation failed for user bootstrap: %w", err)
	}
	defaultBootstrap := &bootstrapv3.Bootstrap{}
	defaultBootstrapStr, err := bootstrap.GetRenderedBootstrapConfig()
	if err != nil {
		return err
	}

	jsonData, err = yaml.YAMLToJSON([]byte(defaultBootstrapStr))
	if err != nil {
		return fmt.Errorf("unable to convert default bootstrap to json: %w", err)
	}

	if err := protojson.Unmarshal(jsonData, defaultBootstrap); err != nil {
		return fmt.Errorf("unable to unmarshal default bootstrap: %w", err)
	}

	// Ensure dynamic resources config is same
	// nolint // Circumvents this error "Error: copylocks: call of reflect.DeepEqual copies lock value:"
	if userBootstrap.DynamicResources == nil || !reflect.DeepEqual(*userBootstrap.DynamicResources, *defaultBootstrap.DynamicResources) {
		return fmt.Errorf("dynamic_resources cannot be modified")
	}
	// Ensure layered runtime resources config is same
	// nolint // Circumvents this error "Error: copylocks: call of reflect.DeepEqual copies lock value:"
	if userBootstrap.LayeredRuntime == nil || !reflect.DeepEqual(*userBootstrap.LayeredRuntime, *defaultBootstrap.LayeredRuntime) {
		return fmt.Errorf("layered_runtime cannot be modified")
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

	// nolint // Circumvents this error "Error: copylocks: call of reflect.DeepEqual copies lock value:"
	if userXdsCluster == nil || !reflect.DeepEqual(*userXdsCluster.LoadAssignment, *defaultXdsCluster.LoadAssignment) {
		return fmt.Errorf("xds_cluster's loadAssigntment cannot be modified")
	}

	return nil
}
