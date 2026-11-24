// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package deprecated

import (
	"fmt"
	"maps"
	"slices"
	"strconv"

	"go.yaml.in/yaml/v3"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/cmd/egctl/migrate"
)

// The detection below works on the YAML document, not on the egv1a1 types, so that it keeps
// finding the deprecated fields after they are removed from the API.

// removalRelease is the release that removes the deprecated fields tracked in
// https://github.com/envoyproxy/gateway/issues/10211.
const removalRelease = "v1.11"

// field describes a deprecated field in use.
type field struct {
	// replacement is the path of the field to use instead.
	replacement string
	// scheduled reports whether the field is removed in removalRelease.
	scheduled bool
}

// message returns the warning message for the deprecated field at path.
func (f field) message(path string) string {
	if f.scheduled {
		return fmt.Sprintf("%s is deprecated and will be removed in %s, use %s instead", path, removalRelease, f.replacement)
	}
	return fmt.Sprintf("%s is deprecated, use %s instead", path, f.replacement)
}

// fields maps the path of each deprecated field in use to its description.
type fields map[string]field

func (fs fields) scheduled(path, replacement string) {
	fs[path] = field{replacement: replacement, scheduled: true}
}

// set reports whether key is set to a non-null value in the mapping node m.
func set(m *yaml.Node, key string) bool {
	v := migrate.Get(m, key)
	return v != nil && (v.Kind != yaml.ScalarNode || v.Tag != "!!null")
}

// each calls fn with every item of the sequence at key in the mapping node m, and its path.
func each(m *yaml.Node, key, path string, fn func(item *yaml.Node, path string)) {
	seq := migrate.Get(m, key)
	if seq == nil || seq.Kind != yaml.SequenceNode {
		return
	}
	for i, item := range seq.Content {
		fn(item, fmt.Sprintf("%s.%s[%d]", path, key, i))
	}
}

// detect returns the deprecated fields used in the resource root of the given kind.
func detect(kind string, root *yaml.Node) fields {
	fs := fields{}
	spec := migrate.GetMapping(root, "spec")
	if spec == nil {
		return fs
	}
	switch kind {
	case egv1a1.KindClientTrafficPolicy:
		detectClientTrafficPolicy(fs, spec)
	case egv1a1.KindBackendTrafficPolicy:
		detectTargetRef(fs, spec)
		if c := migrate.Get(spec, "compression"); c != nil && c.Kind == yaml.SequenceNode && len(c.Content) > 0 {
			fs.scheduled("spec.compression", "spec.compressor")
		}
		detectLoadBalancer(fs, spec, "spec")
	case egv1a1.KindSecurityPolicy:
		detectSecurityPolicy(fs, spec)
	case egv1a1.KindEnvoyExtensionPolicy:
		detectTargetRef(fs, spec)
		each(spec, "extProc", "spec", func(ep *yaml.Node, path string) {
			detectBackendCluster(fs, ep, path, notSink)
		})
		each(spec, "backends", "spec", func(b *yaml.Node, path string) {
			detectLoadBalancer(fs, migrate.GetMapping(b, "backendSettings"), path+".backendSettings")
		})
	case egv1a1.KindEnvoyProxy:
		detectEnvoyProxySpec(fs, spec, "spec", defaultedSink)
	}
	return fs
}

func detectTargetRef(fs fields, spec *yaml.Node) {
	if set(spec, "targetRef") {
		fs.scheduled("spec.targetRef", "spec.targetRefs")
	}
}

func detectClientTrafficPolicy(fs fields, spec *yaml.Node) {
	detectTargetRef(fs, spec)
	if set(spec, "enableProxyProtocol") {
		fs.scheduled("spec.enableProxyProtocol", "spec.proxyProtocol")
	}
	if set(migrate.GetMapping(spec, "headers"), "preserveXRequestID") {
		fs.scheduled("spec.headers.preserveXRequestID", "spec.headers.requestID")
	}
	// optional: false is the default, so only true is reported.
	if cv := migrate.Lookup(spec, "tls.clientValidation"); migrate.IsTrue(migrate.Get(cv, "optional")) {
		fs["spec.tls.clientValidation.optional"] = field{replacement: "spec.tls.clientValidation.mode"}
	}
}

func detectSecurityPolicy(fs fields, spec *yaml.Node) {
	detectTargetRef(fs, spec)
	if extAuth := migrate.GetMapping(spec, "extAuth"); extAuth != nil {
		detectBackendCluster(fs, migrate.GetMapping(extAuth, "grpc"), "spec.extAuth.grpc", notSink)
		detectBackendCluster(fs, migrate.GetMapping(extAuth, "http"), "spec.extAuth.http", notSink)
	}
	each(migrate.GetMapping(spec, "jwt"), "providers", "spec.jwt", func(p *yaml.Node, path string) {
		detectBackendCluster(fs, migrate.GetMapping(p, "remoteJWKS"), path+".remoteJWKS", notSink)
	})
	detectBackendCluster(fs, migrate.Lookup(spec, "oidc.provider"), "spec.oidc.provider", notSink)
}

// sinkKind says whether a BackendCluster is a telemetry sink with the deprecated host and port.
type sinkKind int

const (
	// notSink is a BackendCluster without host and port.
	notSink sinkKind = iota
	// defaultedSink is a telemetry sink of an EnvoyProxy resource. The CRD defaults its port to
	// defaultTelemetryPort, so only another port is known to be set by the user.
	defaultedSink
	// explicitSink is a telemetry sink of the EnvoyGateway configuration, which isn't defaulted, so
	// any nonzero port is set by the user.
	explicitSink
)

// detectBackendCluster records the deprecated fields of the BackendCluster at path: backendRef, the
// host and port of a telemetry sink, and the consistent hash fields of backendSettings, which is
// inlined into the BackendCluster.
func detectBackendCluster(fs fields, cluster *yaml.Node, path string, sink sinkKind) {
	if cluster == nil || cluster.Kind != yaml.MappingNode {
		return
	}
	if sink != notSink {
		port := migrate.Get(cluster, "port")
		switch {
		case set(cluster, "host"):
			fs.scheduled(path+".host", path+".backendRefs")
		case !set(cluster, "port") || port.Value == "0":
		case sink == defaultedSink && port.Value == strconv.Itoa(defaultTelemetryPort):
		default:
			// The host warning covers the port when both are set.
			fs.scheduled(path+".port", path+".backendRefs")
		}
	}
	if set(cluster, "backendRef") {
		fs.scheduled(path+".backendRef", path+".backendRefs")
	}
	detectLoadBalancer(fs, migrate.GetMapping(cluster, "backendSettings"), path+".backendSettings")
}

// detectLoadBalancer records the deprecated consistent hash fields of the loadBalancer of settings,
// a ClusterSettings mapping at path.
func detectLoadBalancer(fs fields, settings *yaml.Node, path string) {
	ch := migrate.Lookup(settings, "loadBalancer.consistentHash")
	if ch == nil || ch.Kind != yaml.MappingNode {
		return
	}
	path += ".loadBalancer.consistentHash"
	if t := migrate.Get(ch, "type"); t != nil && t.Value == consistentHashTypeHeader {
		fs.scheduled(path+".type: Header", path+".type: Headers")
	}
	if set(ch, "header") {
		fs.scheduled(path+".header", path+".headers")
	}
}

// detectEnvoyProxySpec records the deprecated fields of the EnvoyProxy spec at path, which is
// "spec" for an EnvoyProxy resource and "envoyProxy" for the EnvoyGateway default. sink says how
// to treat the port of its telemetry sinks.
func detectEnvoyProxySpec(fs fields, spec *yaml.Node, path string, sink sinkKind) {
	if set(spec, "luaValidation") {
		fs[path+".luaValidation"] = field{replacement: path + ".lua.validationType"}
	}
	telemetry := migrate.GetMapping(spec, "telemetry")
	if telemetry == nil {
		return
	}
	detectBackendCluster(fs, migrate.Lookup(telemetry, "tracing.provider"), path+".telemetry.tracing.provider", sink)
	each(migrate.GetMapping(telemetry, "metrics"), "sinks", path+".telemetry.metrics", func(s *yaml.Node, p string) {
		detectBackendCluster(fs, migrate.GetMapping(s, "openTelemetry"), p+".openTelemetry", sink)
	})
	each(migrate.GetMapping(telemetry, "accessLog"), "settings", path+".telemetry.accessLog", func(setting *yaml.Node, p string) {
		each(setting, "sinks", p, func(s *yaml.Node, p string) {
			detectBackendCluster(fs, migrate.GetMapping(s, "openTelemetry"), p+".openTelemetry", sink)
			detectBackendCluster(fs, migrate.GetMapping(s, "als"), p+".als", notSink)
		})
	})
}

// warning is a deprecation warning for the field at path of the EnvoyGateway configuration.
type warning struct {
	path    string
	message string
}

// detectEnvoyGateway returns the deprecated fields used in the EnvoyGateway configuration eg.
func detectEnvoyGateway(eg *yaml.Node) []warning {
	var warnings []warning
	if set(migrate.GetMapping(eg, "extensionApis"), "disableLua") {
		warnings = append(warnings, warning{"extensionApis.disableLua", "disableLua is deprecated, use enableLua instead"})
	}
	warnings = append(warnings, detectExtensionService(migrate.Lookup(eg, "extensionManager.service"), "extensionManager.service")...)
	each(eg, "extensionManagers", "", func(em *yaml.Node, path string) {
		path += ".service"
		warnings = append(warnings, detectExtensionService(migrate.GetMapping(em, "service"), path[1:])...)
	})
	warnings = append(warnings, detectExtensionService(migrate.Lookup(eg, "provider.custom.infrastructure.remote.service"),
		"provider.custom.infrastructure.remote.service")...)

	// The default EnvoyProxy spec isn't an EnvoyProxy resource, so it has no status to carry a
	// DeprecatedField warning.
	fs := fields{}
	if ep := migrate.GetMapping(eg, "envoyProxy"); ep != nil {
		detectEnvoyProxySpec(fs, ep, "envoyProxy", explicitSink)
	}
	for _, path := range slices.Sorted(maps.Keys(fs)) {
		warnings = append(warnings, warning{path, fs[path].message(path)})
	}
	return warnings
}

// detectExtensionService returns the deprecated host and port fields of the ExtensionService svc
// at path. A port is deprecated even when the target has been moved to fqdn, ip or unix.
func detectExtensionService(svc *yaml.Node, path string) []warning {
	if svc == nil || svc.Kind != yaml.MappingNode {
		return nil
	}
	if host := migrate.Get(svc, "host"); host != nil && host.Kind == yaml.ScalarNode && host.Value != "" && host.Tag != "!!null" {
		return []warning{{path + ".host", fmt.Sprintf("%[1]s.host and %[1]s.port are deprecated and will be removed in %[2]s, "+
			"use %[1]s.fqdn, %[1]s.ip or %[1]s.unix instead", path, removalRelease)}}
	}
	if port := migrate.Get(svc, "port"); port != nil && port.Value != "0" && port.Tag != "!!null" {
		return []warning{{path + ".port", fmt.Sprintf("%[1]s.port is deprecated and will be removed in %[2]s, "+
			"use the port of %[1]s.fqdn or %[1]s.ip instead", path, removalRelease)}}
	}
	return nil
}
