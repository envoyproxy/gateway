// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package validation

import (
	"errors"
	"fmt"
	"net"
	"net/netip"
	"regexp"

	"github.com/dominikbraun/graph"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/utils/ptr"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

// ValidateEnvoyProxy validates the provided EnvoyProxy.
func ValidateEnvoyProxy(proxy *egv1a1.EnvoyProxy) error {
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
// This method validates everything except for the bootstrap section, because validating the bootstrap
// section in this method would require calling into the internal apis, and would cause an import cycle.
func validateEnvoyProxySpec(spec *egv1a1.EnvoyProxySpec) error {
	var errs []error

	if spec == nil {
		errs = append(errs, errors.New("spec is nil"))
	}

	// validate provider
	validateProviderErrs := validateProvider(spec)
	if len(validateProviderErrs) != 0 {
		errs = append(errs, validateProviderErrs...)
	}

	validateProxyTelemetryErrs := validateProxyTelemetry(spec)
	if len(validateProxyTelemetryErrs) != 0 {
		errs = append(errs, validateProxyTelemetryErrs...)
	}

	// validate filter order
	if spec != nil && spec.FilterOrder != nil {
		if err := validateFilterOrder(spec.FilterOrder); err != nil {
			errs = append(errs, err)
		}
	}

	return utilerrors.NewAggregate(errs)
}

// TODO: remove this function if CEL validation became stable
func validateProvider(spec *egv1a1.EnvoyProxySpec) []error {
	var errs []error
	if spec != nil && spec.Provider != nil {
		if spec.Provider.Type != egv1a1.ProviderTypeKubernetes {
			errs = append(errs, fmt.Errorf("unsupported provider type %v", spec.Provider.Type))
		}
		validateDeploymentErrs := validateDeployment(spec)
		if len(validateDeploymentErrs) != 0 {
			errs = append(errs, validateDeploymentErrs...)
		}
		validateHpaErrors := validateHpa(spec)
		if len(validateHpaErrors) != 0 {
			errs = append(errs, validateHpaErrors...)
		}
		validatePdbErrors := validatePdb(spec)
		if len(validatePdbErrors) != 0 {
			errs = append(errs, validatePdbErrors...)
		}
		validateServiceErrs := validateService(spec)
		if len(validateServiceErrs) != 0 {
			errs = append(errs, validateServiceErrs...)
		}
	}
	return errs
}

func validateDeployment(spec *egv1a1.EnvoyProxySpec) []error {
	var errs []error
	if spec.Provider.Kubernetes != nil && spec.Provider.Kubernetes.EnvoyDeployment != nil {
		if patch := spec.Provider.Kubernetes.EnvoyDeployment.Patch; patch != nil {
			if patch.Value.Raw == nil {
				errs = append(errs, fmt.Errorf("envoy deployment patch object cannot be empty"))
			}
			if patch.Type != nil && *patch.Type != egv1a1.JSONMerge && *patch.Type != egv1a1.StrategicMerge {
				errs = append(errs, fmt.Errorf("unsupported envoy deployment patch type %s", *patch.Type))
			}
		}
	}
	return errs
}

func validateHpa(spec *egv1a1.EnvoyProxySpec) []error {
	var errs []error
	if spec.Provider.Kubernetes != nil && spec.Provider.Kubernetes.EnvoyHpa != nil {
		if patch := spec.Provider.Kubernetes.EnvoyHpa.Patch; patch != nil {
			if patch.Value.Raw == nil {
				errs = append(errs, fmt.Errorf("envoy hpa patch object cannot be empty"))
			}
			if patch.Type != nil && *patch.Type != egv1a1.JSONMerge && *patch.Type != egv1a1.StrategicMerge {
				errs = append(errs, fmt.Errorf("unsupported envoy hpa patch type %s", *patch.Type))
			}
		}
	}
	return errs
}

func validatePdb(spec *egv1a1.EnvoyProxySpec) []error {
	var errs []error
	if spec.Provider.Kubernetes != nil && spec.Provider.Kubernetes.EnvoyPDB != nil {
		if patch := spec.Provider.Kubernetes.EnvoyPDB.Patch; patch != nil {
			if patch.Value.Raw == nil {
				errs = append(errs, fmt.Errorf("envoy pdb patch object cannot be empty"))
			}
			if patch.Type != nil && *patch.Type != egv1a1.JSONMerge && *patch.Type != egv1a1.StrategicMerge {
				errs = append(errs, fmt.Errorf("unsupported envoy pdb patch type %s", *patch.Type))
			}
		}
	}
	return errs
}

// TODO: remove this function if CEL validation became stable
func validateService(spec *egv1a1.EnvoyProxySpec) []error {
	var errs []error
	if spec.Provider.Kubernetes != nil && spec.Provider.Kubernetes.EnvoyService != nil {
		if serviceType := spec.Provider.Kubernetes.EnvoyService.Type; serviceType != nil {
			if *serviceType != egv1a1.ServiceTypeLoadBalancer &&
				*serviceType != egv1a1.ServiceTypeClusterIP &&
				*serviceType != egv1a1.ServiceTypeNodePort {
				errs = append(errs, fmt.Errorf("unsupported envoy service type %v", serviceType))
			}
		}
		if serviceType, serviceAllocateLoadBalancerNodePorts := spec.Provider.Kubernetes.EnvoyService.Type, spec.Provider.Kubernetes.EnvoyService.AllocateLoadBalancerNodePorts; serviceType != nil && serviceAllocateLoadBalancerNodePorts != nil {
			if *serviceType != egv1a1.ServiceTypeLoadBalancer {
				errs = append(errs, fmt.Errorf("allocateLoadBalancerNodePorts can only be set for %v type", egv1a1.ServiceTypeLoadBalancer))
			}
		}
		if serviceType, serviceLoadBalancerSourceRanges := spec.Provider.Kubernetes.EnvoyService.Type, spec.Provider.Kubernetes.EnvoyService.LoadBalancerSourceRanges; serviceType != nil && serviceLoadBalancerSourceRanges != nil {
			if *serviceType != egv1a1.ServiceTypeLoadBalancer {
				errs = append(errs, fmt.Errorf("loadBalancerSourceRanges can only be set for %v type", egv1a1.ServiceTypeLoadBalancer))
			}

			for _, serviceLoadBalancerSourceRange := range serviceLoadBalancerSourceRanges {
				if ip, _, err := net.ParseCIDR(serviceLoadBalancerSourceRange); err != nil || ip.To4() == nil {
					errs = append(errs, fmt.Errorf("loadBalancerSourceRange:%s is an invalid IP subnet", serviceLoadBalancerSourceRange))
				}
			}
		}
		if serviceType, serviceLoadBalancerIP := spec.Provider.Kubernetes.EnvoyService.Type, spec.Provider.Kubernetes.EnvoyService.LoadBalancerIP; serviceType != nil && serviceLoadBalancerIP != nil {
			if *serviceType != egv1a1.ServiceTypeLoadBalancer {
				errs = append(errs, fmt.Errorf("loadBalancerIP can only be set for %v type", egv1a1.ServiceTypeLoadBalancer))
			}

			if ip, err := netip.ParseAddr(*serviceLoadBalancerIP); err != nil || !ip.Unmap().Is4() {
				errs = append(errs, fmt.Errorf("loadBalancerIP:%s is an invalid IP address", *serviceLoadBalancerIP))
			}
		}
		if patch := spec.Provider.Kubernetes.EnvoyService.Patch; patch != nil {
			if patch.Value.Raw == nil {
				errs = append(errs, fmt.Errorf("envoy service patch object cannot be empty"))
			}
			if patch.Type != nil && *patch.Type != egv1a1.JSONMerge && *patch.Type != egv1a1.StrategicMerge {
				errs = append(errs, fmt.Errorf("unsupported envoy service patch type %s", *patch.Type))
			}
		}

	}
	return errs
}

func validateProxyTelemetry(spec *egv1a1.EnvoyProxySpec) []error {
	var errs []error

	if spec != nil &&
		spec.Telemetry != nil &&
		spec.Telemetry.AccessLog != nil {
		accessLogErrs := validateProxyAccessLog(spec.Telemetry.AccessLog)
		if len(accessLogErrs) > 0 {
			errs = append(errs, accessLogErrs...)
		}
	}

	if spec != nil && spec.Telemetry != nil && spec.Telemetry.Metrics != nil {
		for _, sink := range spec.Telemetry.Metrics.Sinks {
			if sink.Type == egv1a1.MetricSinkTypeOpenTelemetry {
				if sink.OpenTelemetry == nil {
					err := fmt.Errorf("opentelemetry is required if the sink type is OpenTelemetry")
					errs = append(errs, err)
				}
			}
		}

		if spec.Telemetry.Metrics.ClusterStatName != nil {
			if clusterStatErrs := validateClusterStatName(*spec.Telemetry.Metrics.ClusterStatName); clusterStatErrs != nil {
				errs = append(errs, clusterStatErrs...)
			}
		}
	}

	return errs
}

func validateProxyAccessLog(accessLog *egv1a1.ProxyAccessLog) []error {
	if ptr.Deref(accessLog.Disable, false) {
		return nil
	}

	var errs []error

	for _, setting := range accessLog.Settings {
		if setting.Format != nil {
			switch setting.Format.Type {
			case egv1a1.ProxyAccessLogFormatTypeText:
				if setting.Format.Text == nil {
					err := fmt.Errorf("unable to configure access log when using Text format but \"text\" field being empty")
					errs = append(errs, err)
				}
			case egv1a1.ProxyAccessLogFormatTypeJSON:
				if setting.Format.JSON == nil {
					err := fmt.Errorf("unable to configure access log when using JSON format but \"json\" field being empty")
					errs = append(errs, err)
				}
			}
		}

		for _, sink := range setting.Sinks {
			switch sink.Type {
			case egv1a1.ProxyAccessLogSinkTypeFile:
				if sink.File == nil {
					err := fmt.Errorf("unable to configure access log when using File sink type but \"file\" field being empty")
					errs = append(errs, err)
				}
			case egv1a1.ProxyAccessLogSinkTypeOpenTelemetry:
				if sink.OpenTelemetry == nil {
					err := fmt.Errorf("unable to configure access log when using OpenTelemetry sink type but \"openTelemetry\" field being empty")
					errs = append(errs, err)
				}
			}
		}
	}

	return errs
}

func validateFilterOrder(filterOrder []egv1a1.FilterPosition) error {
	g := graph.New(graph.StringHash, graph.Directed(), graph.PreventCycles())

	for _, filter := range filterOrder {
		// Ignore the error since the same filter can be added multiple times
		_ = g.AddVertex(string(filter.Name))
		if filter.Before != nil {
			_ = g.AddVertex(string(*filter.Before))
		}
		if filter.After != nil {
			_ = g.AddVertex(string(*filter.After))
		}
	}

	for _, filter := range filterOrder {
		var from, to string
		if filter.Before != nil {
			from = string(filter.Name)
			to = string(*filter.Before)
		} else {
			from = string(*filter.After)
			to = string(filter.Name)
		}
		if err := g.AddEdge(from, to); err != nil {
			if errors.Is(err, graph.ErrEdgeCreatesCycle) {
				return fmt.Errorf("there is a cycle in the filter order: %s -> %s", from, to)
			}
		}
	}

	return nil
}

func validateClusterStatName(clusterStatName string) []error {
	supportedOperators := map[string]bool{
		egv1a1.StatFormatterRouteName:       true,
		egv1a1.StatFormatterRouteNamespace:  true,
		egv1a1.StatFormatterRouteKind:       true,
		egv1a1.StatFormatterRouteRuleName:   true,
		egv1a1.StatFormatterRouteRuleNumber: true,
		egv1a1.StatFormatterBackendRefs:     true,
	}

	var errs []error
	re := regexp.MustCompile("%[^%]*%")
	matches := re.FindAllString(clusterStatName, -1)
	for _, operator := range matches {
		if _, ok := supportedOperators[operator]; !ok {
			err := fmt.Errorf("unable to configure Cluster Stat Name with unsupported operator: %s", operator)
			errs = append(errs, err)
		}
	}

	return errs
}
