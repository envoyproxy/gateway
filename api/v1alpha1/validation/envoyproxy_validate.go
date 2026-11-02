// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package validation

import (
	"errors"
	"fmt"
	"math"
	"net"
	"net/netip"
	"regexp"
	"slices"
	"strings"

	"github.com/dominikbraun/graph"
	"k8s.io/apimachinery/pkg/api/resource"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/utils/ptr"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

var statNameRegex = regexp.MustCompile("%[^%]*%")

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

// InvalidField is an EnvoyProxy field that failed validation but can be unset without
// affecting the proxy infrastructure or the rest of the EnvoyProxy configuration.
type InvalidField struct {
	// Path is the path of the field relative to the EnvoyProxy spec.
	Path string
	Err  error
}

func (f InvalidField) Error() string {
	return fmt.Sprintf("%s: %v", f.Path, f.Err)
}

// ValidateEnvoyProxyFields validates the provided EnvoyProxy like ValidateEnvoyProxy, but
// returns the errors scoped to a field that DropInvalidFields can unset separately from the
// errors that make the whole EnvoyProxy unusable.
func ValidateEnvoyProxyFields(proxy *egv1a1.EnvoyProxy) ([]InvalidField, error) {
	if proxy == nil {
		return nil, errors.New("envoyproxy is nil")
	}
	invalid := append(invalidTelemetryFields(&proxy.Spec), invalidFilterOrderFields(&proxy.Spec)...)
	return invalid, utilerrors.NewAggregate(validateEnvoyProxySpecFatal(&proxy.Spec))
}

// DropInvalidFields returns a copy of proxy with every field reported by
// ValidateEnvoyProxyFields as an InvalidField unset. It returns proxy itself when there is
// nothing to drop.
//
// It is intentionally separate from ValidateEnvoyProxyFields rather than a single
// validate-and-sanitize function: the translator validates each source EnvoyProxy to report
// status on it, and separately sanitizes each layer before merging it into the EnvoyProxy that
// reaches the IR, so a combined function would deep copy every EnvoyProxy on the status path
// just to throw the copy away. The cost is that both functions must cover the same fields; when adding an
// InvalidField check, clear the field here too. TestValidateEnvoyProxyFields asserts that
// nothing is reported after DropInvalidFields.
func DropInvalidFields(proxy *egv1a1.EnvoyProxy) *egv1a1.EnvoyProxy {
	if proxy == nil {
		return nil
	}
	if len(invalidTelemetryFields(&proxy.Spec)) == 0 && len(invalidFilterOrderFields(&proxy.Spec)) == 0 {
		return proxy
	}

	out := proxy.DeepCopy()
	spec := &out.Spec
	if spec.FilterOrder != nil && validateFilterOrder(spec.FilterOrder) != nil {
		spec.FilterOrder = nil
	}
	if spec.Telemetry == nil {
		return out
	}
	if al := spec.Telemetry.AccessLog; al != nil && !ptr.Deref(al.Disable, false) {
		// An invalid format makes the whole setting unusable, while an invalid sink only
		// affects itself.
		al.Settings = slices.DeleteFunc(al.Settings, func(setting egv1a1.ProxyAccessLogSetting) bool {
			return validateProxyAccessLogFormat(setting.Format) != nil
		})
		for i := range al.Settings {
			al.Settings[i].Sinks = slices.DeleteFunc(al.Settings[i].Sinks, func(sink egv1a1.ProxyAccessLogSink) bool {
				return validateProxyAccessLogSink(&sink) != nil
			})
		}
		// A setting needs at least one sink, so one whose sinks were all invalid is dropped
		// too. When no setting is left, settings is unset rather than empty, so that a merged
		// EnvoyProxy keeps the settings of the less specific layer.
		al.Settings = slices.DeleteFunc(al.Settings, func(setting egv1a1.ProxyAccessLogSetting) bool {
			return len(setting.Sinks) == 0
		})
		if len(al.Settings) == 0 {
			al.Settings = nil
		}
	}
	if validateProxyTracing(spec.Telemetry.Tracing) != nil {
		spec.Telemetry.Tracing = nil
	}
	if m := spec.Telemetry.Metrics; m != nil {
		m.Sinks = slices.DeleteFunc(m.Sinks, func(sink egv1a1.ProxyMetricSink) bool {
			return len(validateProxyMetricSink(&sink)) > 0
		})
		if m.ClusterStatName != nil && ValidateClusterStatName(*m.ClusterStatName) != nil {
			m.ClusterStatName = nil
		}
		if m.Prometheus != nil && m.Prometheus.Compression != nil &&
			validatePrometheusCompression(m.Prometheus.Compression) != nil {
			m.Prometheus.Compression = nil
		}
	}
	return out
}

// validateEnvoyProxySpec validates the provided EnvoyProxy spec.
// This method validates everything except for the bootstrap section, because validating the bootstrap
// section in this method would require calling into the internal apis, and would cause an import cycle.
func validateEnvoyProxySpec(spec *egv1a1.EnvoyProxySpec) error {
	errs := validateEnvoyProxySpecFatal(spec)
	for _, f := range invalidTelemetryFields(spec) {
		errs = append(errs, f.Err)
	}
	for _, f := range invalidFilterOrderFields(spec) {
		errs = append(errs, f.Err)
	}

	return utilerrors.NewAggregate(errs)
}

// validateEnvoyProxySpecFatal returns the errors that can't be scoped to a field that is
// safe to unset, because they affect the proxy infrastructure.
func validateEnvoyProxySpecFatal(spec *egv1a1.EnvoyProxySpec) []error {
	var errs []error

	if spec == nil {
		errs = append(errs, errors.New("spec is nil"))
	}

	// validate provider
	validateProviderErrs := validateProvider(spec)
	if len(validateProviderErrs) != 0 {
		errs = append(errs, validateProviderErrs...)
	}

	// Mirrors the EnvoyProxySpec CEL rule "mergeGateways and mergeBackends cannot
	// both be enabled". The CEL rule only runs through Kubernetes admission, which
	// the embedded default EnvoyProxy spec loaded from an EnvoyGateway file never
	// goes through, so this method has to enforce it explicitly too.
	if spec != nil && ptr.Deref(spec.MergeGateways, false) && spec.MergeBackends != nil {
		errs = append(errs, errors.New("mergeGateways and mergeBackends cannot both be enabled"))
	}

	return errs
}

func invalidFilterOrderFields(spec *egv1a1.EnvoyProxySpec) []InvalidField {
	if spec == nil || spec.FilterOrder == nil {
		return nil
	}
	if err := validateFilterOrder(spec.FilterOrder); err != nil {
		return []InvalidField{{Path: "filterOrder", Err: err}}
	}
	return nil
}

// TODO: remove this function if CEL validation became stable
func validateProvider(spec *egv1a1.EnvoyProxySpec) []error {
	var errs []error
	if spec != nil && spec.Provider != nil {
		switch spec.Provider.Type {
		case egv1a1.EnvoyProxyProviderTypeKubernetes:
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
		case egv1a1.EnvoyProxyProviderTypeHost:
			if spec.Provider.Host == nil {
				errs = append(errs, fmt.Errorf("field 'host' should be specified when provider type is 'Host'"))
			}
		default:
			errs = append(errs, fmt.Errorf("unsupported provider type %v", spec.Provider.Type))
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
				if _, _, err := net.ParseCIDR(serviceLoadBalancerSourceRange); err != nil {
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

func invalidTelemetryFields(spec *egv1a1.EnvoyProxySpec) []InvalidField {
	if spec == nil || spec.Telemetry == nil {
		return nil
	}
	var invalid []InvalidField

	if al := spec.Telemetry.AccessLog; al != nil && !ptr.Deref(al.Disable, false) {
		for i, setting := range al.Settings {
			if err := validateProxyAccessLogFormat(setting.Format); err != nil {
				invalid = append(invalid, InvalidField{Path: fmt.Sprintf("telemetry.accessLog.settings[%d].format", i), Err: err})
			}
			for j := range setting.Sinks {
				if err := validateProxyAccessLogSink(&setting.Sinks[j]); err != nil {
					invalid = append(invalid, InvalidField{Path: fmt.Sprintf("telemetry.accessLog.settings[%d].sinks[%d]", i, j), Err: err})
				}
			}
		}
	}

	if err := validateProxyTracing(spec.Telemetry.Tracing); err != nil {
		invalid = append(invalid, InvalidField{Path: "telemetry.tracing", Err: err})
	}

	if m := spec.Telemetry.Metrics; m != nil {
		for i := range m.Sinks {
			for _, err := range validateProxyMetricSink(&m.Sinks[i]) {
				invalid = append(invalid, InvalidField{Path: fmt.Sprintf("telemetry.metrics.sinks[%d]", i), Err: err})
			}
		}

		if m.ClusterStatName != nil {
			if err := ValidateClusterStatName(*m.ClusterStatName); err != nil {
				invalid = append(invalid, InvalidField{Path: "telemetry.metrics.clusterStatName", Err: err})
			}
		}

		if m.Prometheus != nil && m.Prometheus.Compression != nil {
			if err := validatePrometheusCompression(m.Prometheus.Compression); err != nil {
				invalid = append(invalid, InvalidField{Path: "telemetry.metrics.prometheus.compression", Err: err})
			}
		}
	}

	return invalid
}

func validateProxyTracing(tracing *egv1a1.ProxyTracing) error {
	if tracing == nil {
		return nil
	}
	p := tracing.Provider
	if p.Type == egv1a1.TracingProviderTypeOpenTelemetry &&
		p.BackendSettings != nil && p.BackendSettings.HTTP1 != nil {
		return fmt.Errorf("http1 is not supported for OpenTelemetry tracing backends, which use gRPC")
	}
	return nil
}

func validateProxyMetricSink(sink *egv1a1.ProxyMetricSink) []error {
	var errs []error
	if sink.Type == egv1a1.MetricSinkTypeOpenTelemetry {
		if sink.OpenTelemetry == nil {
			err := fmt.Errorf("opentelemetry is required if the sink type is OpenTelemetry")
			errs = append(errs, err)
		}
	}
	// Check regardless of sink.Type: when sink.Type is omitted, CRD defaults are not
	// applied to embedded EnvoyProxy templates, so processMetrics processes any non-nil
	// sink.OpenTelemetry as gRPC, making http1 silently ignored.
	if sink.OpenTelemetry != nil && sink.OpenTelemetry.BackendSettings != nil && sink.OpenTelemetry.BackendSettings.HTTP1 != nil {
		errs = append(errs, fmt.Errorf("http1 is not supported for OpenTelemetry metrics backends, which use gRPC"))
	}
	return errs
}

func validatePrometheusCompression(compression *egv1a1.Compression) error {
	if err := ValidateMinContentLength(compression.MinContentLength); err != nil {
		return fmt.Errorf("invalid prometheus compression: %w", err)
	}
	return nil
}

// ValidateMinContentLength validates that the compression minContentLength fits the uint32 Envoy field.
func ValidateMinContentLength(minContentLength *resource.Quantity) error {
	if minContentLength == nil {
		return nil
	}

	value, ok := minContentLength.AsInt64()
	if !ok || value < 1 || value > math.MaxUint32 {
		return fmt.Errorf("minContentLength must be between 1 and %d bytes, got %s", uint32(math.MaxUint32), minContentLength.String())
	}
	return nil
}

func validateProxyAccessLog(accessLog *egv1a1.ProxyAccessLog) []error {
	if ptr.Deref(accessLog.Disable, false) {
		return nil
	}

	var errs []error
	for i := range accessLog.Settings {
		errs = append(errs, validateProxyAccessLogSetting(&accessLog.Settings[i])...)
	}

	return errs
}

func validateProxyAccessLogSetting(setting *egv1a1.ProxyAccessLogSetting) []error {
	var errs []error
	if err := validateProxyAccessLogFormat(setting.Format); err != nil {
		errs = append(errs, err)
	}
	for i := range setting.Sinks {
		if err := validateProxyAccessLogSink(&setting.Sinks[i]); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

func validateProxyAccessLogFormat(format *egv1a1.ProxyAccessLogFormat) error {
	if format == nil || format.Type == nil {
		return nil
	}
	switch *format.Type {
	case egv1a1.ProxyAccessLogFormatTypeText:
		if format.Text == nil {
			return fmt.Errorf("unable to configure access log when using Text format but \"text\" field being empty")
		}
	case egv1a1.ProxyAccessLogFormatTypeJSON:
		// TODO: add JSON format validation if needed
	}
	return nil
}

func validateProxyAccessLogSink(sink *egv1a1.ProxyAccessLogSink) error {
	switch sink.Type {
	case egv1a1.ProxyAccessLogSinkTypeFile:
		if sink.File == nil {
			return fmt.Errorf("unable to configure access log when using File sink type but \"file\" field being empty")
		}
	case egv1a1.ProxyAccessLogSinkTypeOpenTelemetry:
		if sink.OpenTelemetry == nil {
			return fmt.Errorf("unable to configure access log when using OpenTelemetry sink type but \"openTelemetry\" field being empty")
		} else if sink.OpenTelemetry.BackendSettings != nil && sink.OpenTelemetry.BackendSettings.HTTP1 != nil {
			return fmt.Errorf("http1 is not supported for OpenTelemetry access log backends, which use gRPC")
		}
	case egv1a1.ProxyAccessLogSinkTypeALS:
		if sink.ALS == nil {
			return fmt.Errorf("unable to configure access log when using ALS sink type but \"als\" field being empty")
		} else if sink.ALS.BackendSettings != nil && sink.ALS.BackendSettings.HTTP1 != nil {
			return fmt.Errorf("http1 is not supported for ALS backends, which use gRPC")
		}
	}
	return nil
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

var (
	routeStatSupportedOperators = map[string]bool{
		egv1a1.StatFormatterRouteName:      true,
		egv1a1.StatFormatterRouteNamespace: true,
		egv1a1.StatFormatterRouteKind:      true,
		egv1a1.StatFormatterRouteRuleName:  true,
	}

	clusterStatSupportedOperators = map[string]bool{
		egv1a1.StatFormatterRouteName:       true,
		egv1a1.StatFormatterRouteNamespace:  true,
		egv1a1.StatFormatterRouteKind:       true,
		egv1a1.StatFormatterRouteRuleName:   true,
		egv1a1.StatFormatterRouteRuleNumber: true,
		egv1a1.StatFormatterBackendRefs:     true,
	}
)

func ValidateRouteStatName(routeStatName string) error {
	if err := validateStatName(routeStatName, routeStatSupportedOperators); err != nil {
		return fmt.Errorf("unable to configure Route Stat Name: %w", err)
	}

	return nil
}

func ValidateClusterStatName(clusterStatName string) error {
	if err := validateStatName(clusterStatName, clusterStatSupportedOperators); err != nil {
		return fmt.Errorf("unable to configure Cluster Stat Name: %w", err)
	}

	return nil
}

func validateStatName(statName string, supportedOperators map[string]bool) error {
	var unsupportedOperators []string
	matches := statNameRegex.FindAllString(statName, -1)
	if len(matches) == 0 && strings.Contains(statName, "%") {
		return fmt.Errorf("unable to configure stat name with invalid operator")
	}
	for _, operator := range matches {
		if _, ok := supportedOperators[operator]; !ok {
			unsupportedOperators = append(unsupportedOperators, operator)
		}
	}

	if len(unsupportedOperators) > 0 {
		return fmt.Errorf("unsupported operators: %v", unsupportedOperators)
	}

	return nil
}
