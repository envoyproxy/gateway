// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package common

import (
	"fmt"
	"time"

	"k8s.io/utils/ptr"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/xds/bootstrap"
)

func getIPFamily(infra *ir.ProxyInfra) *egv1a1.IPFamily {
	if infra == nil || infra.Config == nil {
		return nil
	}

	return infra.Config.Spec.IPFamily
}

// BuildProxyArgs builds command arguments for proxy infrastructure.
func BuildProxyArgs(
	infra *ir.ProxyInfra,
	shutdownConfig *egv1a1.ShutdownConfig,
	bootstrapConfigOptions *bootstrap.RenderBootstrapConfigOptions,
	serviceNode string,
	gatewayNamespaceMode bool,
) ([]string, error) {
	serviceCluster := infra.Name
	if gatewayNamespaceMode {
		serviceCluster = fmt.Sprintf("%s/%s", infra.Namespace, infra.Name)
	}

	if bootstrapConfigOptions != nil {
		// Configure local Envoy ServiceCluster
		bootstrapConfigOptions.ServiceClusterName = ptr.To(serviceCluster)

		// If IPFamily is not set, try to determine it from the infrastructure.
		if bootstrapConfigOptions.IPFamily == nil {
			bootstrapConfigOptions.IPFamily = getIPFamily(infra)
		}
	}

	bootstrapConfigOptions.GatewayNamespaceMode = gatewayNamespaceMode
	bootstrapConfigurations, err := bootstrap.GetRenderedBootstrapConfig(bootstrapConfigOptions)
	if err != nil {
		return nil, err
	}

	// Apply Bootstrap from EnvoyProxy API if set by the user
	// The config should have been validated already.
	if infra.Config != nil && infra.Config.Spec.Bootstrap != nil {
		bootstrapConfigurations, err = bootstrap.ApplyBootstrapConfig(infra.Config.Spec.Bootstrap, bootstrapConfigurations)
		if err != nil {
			return nil, err
		}
	}

	logging := infra.Config.Spec.Logging

	// The func-e library used by the infrastructure provider parses the arguments and does not support the '=' syntax.
	// It is important to make sure each element of the array is a single string to make sure arguments are properly
	// processed.
	args := []string{
		"--service-cluster", serviceCluster,
		"--service-node", serviceNode,
		"--config-yaml", bootstrapConfigurations,
		"--log-level", string(logging.DefaultEnvoyProxyLoggingLevel()),
		"--cpuset-threads",
		"--drain-strategy", "immediate",
	}

	if infra.Config != nil &&
		infra.Config.Spec.Concurrency != nil {
		args = append(args, "--concurrency", fmt.Sprintf("%d", *infra.Config.Spec.Concurrency))
	}

	if componentsLogLevel := logging.GetEnvoyProxyComponentLevel(); componentsLogLevel != "" {
		args = append(args, "--component-log-level", componentsLogLevel)
	}

	// Default drain timeout.
	drainTimeout := 60.0
	if shutdownConfig != nil && shutdownConfig.DrainTimeout != nil {
		d, err := time.ParseDuration(string(*shutdownConfig.DrainTimeout))
		if err != nil {
			return nil, err
		}
		drainTimeout = d.Seconds()
	}
	args = append(args, "--drain-time-s", fmt.Sprintf("%.0f", drainTimeout))

	if infra.Config != nil {
		args = append(args, infra.Config.Spec.ExtraArgs...)
	}

	return args, nil
}
