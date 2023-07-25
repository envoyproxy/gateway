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
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/testing/protocmp"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"sigs.k8s.io/yaml"

	egcfgv1a1 "github.com/envoyproxy/gateway/api/config/v1alpha1"
	"github.com/envoyproxy/gateway/internal/xds/bootstrap"
	_ "github.com/envoyproxy/gateway/internal/xds/extensions" // register the generated types to support protojson unmarshalling
)

// Validate validates the provided EnvoyProxy.
func ValidateEnvoyProxy(proxy *egcfgv1a1.EnvoyProxy) error {
	var errs []error
	if proxy == nil {
		return errors.New("envoyproxy is nil")
	}
	if err := validateEnvoyProxySpec(&proxy.Spec); err != nil {
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

	// validate provider
	validateProviderErrs := validateProvider(spec)
	if len(validateProviderErrs) != 0 {
		errs = append(errs, validateProviderErrs...)
	}

	// validate bootstrap
	if spec != nil && spec.Bootstrap != nil {
		if err := validateBootstrap(spec.Bootstrap); err != nil {
			errs = append(errs, err)
		}
	}
	return utilerrors.NewAggregate(errs)
}

func validateProvider(spec *egcfgv1a1.EnvoyProxySpec) []error {
	var errs []error
	if spec != nil && spec.Provider != nil {
		if spec.Provider.Type != egcfgv1a1.ProviderTypeKubernetes {
			errs = append(errs, fmt.Errorf("unsupported provider type %v", spec.Provider.Type))
		}
		validateServiceTypeErrs := validateServiceType(spec)
		if len(validateServiceTypeErrs) != 0 {
			errs = append(errs, validateServiceTypeErrs...)
		}
	}
	return errs
}

func validateServiceType(spec *egcfgv1a1.EnvoyProxySpec) []error {
	var errs []error
	if spec.Provider.Kubernetes != nil && spec.Provider.Kubernetes.EnvoyService != nil {
		if serviceType := spec.Provider.Kubernetes.EnvoyService.Type; serviceType != nil {
			if *serviceType != egcfgv1a1.ServiceTypeLoadBalancer &&
				*serviceType != egcfgv1a1.ServiceTypeClusterIP &&
				*serviceType != egcfgv1a1.ServiceTypeNodePort {
				errs = append(errs, fmt.Errorf("unsupported envoy service type %v", serviceType))
			}
		}
	}
	return errs
}

func validateBootstrap(boostrapConfig *string) error {
	userBootstrap := &bootstrapv3.Bootstrap{}
	jsonData, err := yaml.YAMLToJSON([]byte(*boostrapConfig))
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
	if userBootstrap.DynamicResources == nil ||
		cmp.Diff(userBootstrap.DynamicResources, defaultBootstrap.DynamicResources, protocmp.Transform()) != "" {
		return fmt.Errorf("dynamic_resources cannot be modified")
	}
	// Ensure layered runtime resources config is same
	if userBootstrap.LayeredRuntime == nil ||
		cmp.Diff(userBootstrap.LayeredRuntime, defaultBootstrap.LayeredRuntime, protocmp.Transform()) != "" {
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
