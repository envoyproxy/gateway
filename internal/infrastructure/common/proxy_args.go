// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package common

import (
	"fmt"

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

	bootstrapConfigOptions.ServiceClusterName = ptr.To(serviceCluster)

	// If IPFamily is not set, try to determine it from the infrastructure.
	if bootstrapConfigOptions != nil && bootstrapConfigOptions.IPFamily == nil {
		bootstrapConfigOptions.IPFamily = getIPFamily(infra)
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

	args := []string{
		fmt.Sprintf("--service-cluster %s", serviceCluster),
		fmt.Sprintf("--service-node %s", serviceNode),
		fmt.Sprintf("--config-yaml %s", bootstrapConfigurations),
		fmt.Sprintf("--log-level %s", logging.DefaultEnvoyProxyLoggingLevel()),
		"--cpuset-threads",
		"--drain-strategy immediate",
	}

	if infra.Config != nil &&
		infra.Config.Spec.Concurrency != nil {
		args = append(args, fmt.Sprintf("--concurrency %d", *infra.Config.Spec.Concurrency))
	}

	if componentsLogLevel := logging.GetEnvoyProxyComponentLevel(); componentsLogLevel != "" {
		args = append(args, fmt.Sprintf("--component-log-level %s", componentsLogLevel))
	}

	// Default drain timeout.
	drainTimeout := 60.0
	if shutdownConfig != nil && shutdownConfig.DrainTimeout != nil {
		drainTimeout = shutdownConfig.DrainTimeout.Seconds()
	}
	args = append(args, fmt.Sprintf("--drain-time-s %.0f", drainTimeout))

	if infra.Config != nil {
		args = append(args, infra.Config.Spec.ExtraArgs...)
	}

	return args, nil
}
