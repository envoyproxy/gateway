// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package deprecated

import (
	"net"
	"regexp"

	"go.yaml.in/yaml/v3"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/cmd/egctl/migrate"
)

const (
	// defaultTelemetryPort is the CRD default of the deprecated telemetry port fields.
	defaultTelemetryPort = 4317
	// consistentHashTypeHeader is the deprecated consistent hash type, replaced by Headers.
	consistentHashTypeHeader = "Header"
)

// migrateTargetRef rewrites spec.targetRef to spec.targetRefs.
//
// The translator uses targetRef alone when it is set, and ignores targetRefs, so targetRef can only
// be dropped when targetRefs holds exactly the same reference.
func migrateTargetRef(spec *yaml.Node, base string) migrate.Finding {
	field := base + ".targetRef"
	ref := migrate.Get(spec, "targetRef")
	refs := migrate.Get(spec, "targetRefs")
	switch {
	case migrate.IsEmpty(refs):
		migrate.Delete(spec, "targetRefs")
		migrate.Replace(spec, "targetRef", "targetRefs", migrate.Sequence(ref))
		return migrated(field, "moved to %s.targetRefs", base)
	case sameRef(refs, ref):
		migrate.Delete(spec, "targetRef")
		return migrated(field, "removed, %s.targetRefs already holds the same reference", base)
	}
	return errorf(field, "both targetRef and targetRefs are set: only targetRef is used today, "+
		"move it into targetRefs and decide which references to keep")
}

// sameRef reports whether refs is a sequence that holds only ref.
func sameRef(refs, ref *yaml.Node) bool {
	refs = migrate.Resolve(refs)
	return refs.Kind == yaml.SequenceNode && len(refs.Content) == 1 && migrate.Equal(refs.Content[0], ref)
}

// migrateBackendRef rewrites backendRef to backendRefs. When backendRefs is already set, backendRef
// is ignored by the translator, so it is removed.
func migrateBackendRef(cluster *yaml.Node, base string) migrate.Finding {
	field := base + ".backendRef"
	ref := migrate.Get(cluster, "backendRef")
	refs := migrate.Get(cluster, "backendRefs")
	if migrate.IsEmpty(refs) {
		migrate.Delete(cluster, "backendRefs")
		migrate.Replace(cluster, "backendRef", "backendRefs", migrate.Sequence(ref))
		return migrated(field, "moved to %s.backendRefs", base)
	}
	migrate.Delete(cluster, "backendRef")
	return migrated(field, "removed, it was ignored because %s.backendRefs is set", base)
}

// migrateProxyProtocol rewrites spec.enableProxyProtocol to spec.proxyProtocol. The legacy field
// maps to a non-optional proxy protocol, which is what an empty proxyProtocol gives.
func migrateProxyProtocol(spec *yaml.Node, base string) migrate.Finding {
	field := base + ".enableProxyProtocol"
	enable := migrate.Get(spec, "enableProxyProtocol")
	switch {
	case migrate.IsSet(spec, "proxyProtocol"):
		migrate.Delete(spec, "enableProxyProtocol")
		return migrated(field, "removed, it was ignored because %s.proxyProtocol is set", base)
	case migrate.IsTrue(enable):
		// A null proxyProtocol is unset, and the translator uses enableProxyProtocol.
		migrate.Delete(spec, "proxyProtocol")
		migrate.Replace(spec, "enableProxyProtocol", "proxyProtocol", migrate.Mapping())
		return migrated(field, "replaced with %s.proxyProtocol: {}", base)
	}
	migrate.Delete(spec, "enableProxyProtocol")
	return migrated(field, "removed, false is the default")
}

// migrateXRequestID rewrites spec.headers.preserveXRequestID to spec.headers.requestID.
func migrateXRequestID(headers *yaml.Node, base string) migrate.Finding {
	field := base + ".preserveXRequestID"
	preserve := migrate.Get(headers, "preserveXRequestID")
	switch {
	case migrate.IsSet(headers, "requestID"):
		migrate.Delete(headers, "preserveXRequestID")
		return migrated(field, "removed, it was ignored because %s.requestID is set", base)
	case migrate.IsTrue(preserve):
		migrate.Delete(headers, "requestID")
		migrate.Replace(headers, "preserveXRequestID", "requestID", migrate.String(string(egv1a1.RequestIDActionPreserveOrGenerate)))
		return migrated(field, "replaced with %s.requestID: PreserveOrGenerate", base)
	}
	migrate.Delete(headers, "preserveXRequestID")
	return migrated(field, "removed, false is the default")
}

// compressorBlocks maps a compressor type to the field holding its settings.
var compressorBlocks = map[string]string{
	string(egv1a1.GzipCompressorType):   "gzip",
	string(egv1a1.BrotliCompressorType): "brotli",
	string(egv1a1.ZstdCompressorType):   "zstd",
}

// migrateCompression rewrites spec.compression to spec.compressor. An entry of compressor is
// skipped unless the settings block of its type is set, which compression doesn't require, so an
// empty block is added where it is missing.
func migrateCompression(spec *yaml.Node, base string) migrate.Finding {
	field := base + ".compression"
	if compressor := migrate.Resolve(migrate.Get(spec, "compressor")); compressor != nil && compressor.Kind == yaml.SequenceNode && len(compressor.Content) > 0 {
		migrate.Delete(spec, "compression")
		return migrated(field, "removed, it was ignored because %s.compressor is set", base)
	}
	compression := migrate.Get(spec, "compression")
	entries := migrate.Resolve(compression)
	if entries.Kind != yaml.SequenceNode {
		return errorf(field, "unexpected shape, not migrated")
	}
	added := 0
	for _, entry := range entries.Content {
		block, ok := compressorBlocks[migrate.Scalar(migrate.Get(entry, "type"))]
		if ok && migrate.Get(entry, block) == nil {
			migrate.Set(entry, block, migrate.Mapping())
			added++
		}
	}
	migrate.Delete(spec, "compressor")
	migrate.Replace(spec, "compression", "compressor", compression)
	if added > 0 {
		return migrated(field, "moved to %s.compressor, with an empty settings block added for %d compressor type(s)", base, added)
	}
	return migrated(field, "moved to %s.compressor", base)
}

// migrateClientValidation rewrites clientValidation.optional: true to mode: VerifyIfGiven.
func migrateClientValidation(cv *yaml.Node, base string) migrate.Finding {
	field := base + ".optional"
	switch {
	case migrate.IsSet(cv, "mode"):
		migrate.Delete(cv, "optional")
		return migrated(field, "removed, it was ignored because %s.mode is set", base)
	case migrate.IsTrue(migrate.Get(cv, "optional")):
		migrate.Delete(cv, "mode")
		migrate.Replace(cv, "optional", "mode", migrate.String(string(egv1a1.ClientValidationVerifyIfGiven)))
		return migrated(field, "replaced with %s.mode: VerifyIfGiven", base)
	}
	migrate.Delete(cv, "optional")
	return migrated(field, "removed, false is the default")
}

// migrateLuaValidation rewrites spec.luaValidation to spec.lua.validationType.
func migrateLuaValidation(spec *yaml.Node, base string) migrate.Finding {
	field := base + ".luaValidation"
	value := migrate.Get(spec, "luaValidation")
	lua := migrate.Resolve(migrate.Get(spec, "lua"))
	switch {
	case migrate.IsSet(lua, "validationType"):
		migrate.Delete(spec, "luaValidation")
		return migrated(field, "removed, it was ignored because %s.lua.validationType is set", base)
	case !migrate.IsSet(spec, "lua"):
		migrate.Delete(spec, "lua")
		migrate.Replace(spec, "luaValidation", "lua", migrate.Mapping("validationType", value))
	case lua.Kind != yaml.MappingNode:
		return errorf(field, "unexpected shape of %s.lua, not migrated", base)
	case migrate.Scalar(value) == string(egv1a1.LuaValidationStrict):
		// lua.validationType defaults to Strict, so the result is the same either way.
		migrate.Delete(spec, "luaValidation")
		return migrated(field, "removed, %s.lua.validationType defaults to Strict", base)
	default:
		// The API server defaults lua.validationType to Strict when lua is set, which takes
		// precedence, so luaValidation is already ignored in a cluster but not in standalone mode.
		return errorf(field, "%s.lua is set without validationType, which defaults to Strict and overrides "+
			"luaValidation: %s; set %s.lua.validationType to the mode you want and remove luaValidation",
			base, migrate.Scalar(value), base)
	}
	return migrated(field, "moved to %s.lua.validationType", base)
}

// migrateConsistentHash rewrites consistentHash type Header and header to type Headers and headers.
func migrateConsistentHash(ch *yaml.Node, base string) migrate.Finding {
	typ := migrate.Get(ch, "type")
	header := migrate.Get(ch, "header")
	headers := migrate.Get(ch, "headers")

	if typ == nil {
		return errorf(base+".type", "unexpected shape, consistentHash has no type, not migrated")
	}
	if t := migrate.Scalar(typ); t != consistentHashTypeHeader {
		// Only header is set, with another type: it is unused.
		migrate.Delete(ch, "header")
		return migrated(base+".header", "removed, it is unused with type %s", t)
	}
	if !migrate.IsSet(ch, "header") {
		return errorf(base+".type", "type Header is set without header, set type Headers and headers instead")
	}
	// Replace the node rather than its value, which may be an anchor shared with other fields.
	migrate.Replace(ch, "type", "type", migrate.String(string(egv1a1.HeadersConsistentHashType)))
	switch {
	case migrate.IsEmpty(headers):
		migrate.Delete(ch, "headers")
		migrate.Replace(ch, "header", "headers", migrate.Sequence(header))
	case migrate.Resolve(headers).Kind == yaml.SequenceNode:
		// Headers is unused with type Header; only header is hashed today.
		migrate.Delete(ch, "headers")
		migrate.Replace(ch, "header", "headers", migrate.Sequence(header))
	default:
		return errorf(base+".header", "unexpected shape of %s.headers, not migrated", base)
	}
	return migrated(base+".header", "replaced type Header and header with type Headers and headers")
}

// serviceHost matches the in-cluster DNS name of a Service.
var serviceHost = regexp.MustCompile(`^([a-z0-9]([-a-z0-9]*[a-z0-9])?)\.([a-z0-9]([-a-z0-9]*[a-z0-9])?)\.svc(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*\.?$`)

// migrateTelemetryHost rewrites the deprecated host and port of a telemetry sink to backendRefs, if
// host is the DNS name of a Service. Other hosts need a Backend resource or a Service, which is up
// to the user, so they're reported as errors.
func migrateTelemetryHost(sink *yaml.Node, base string) migrate.Finding {
	field := base + ".host"
	host := migrate.Get(sink, "host")
	port := telemetryPort(sink)

	if !migrate.IsEmpty(migrate.Get(sink, "backendRefs")) {
		// The translator only falls back to host and port without backendRefs.
		migrate.Delete(sink, "host")
		migrate.Delete(sink, "port")
		return migrated(field, "removed host and port, they were ignored because %s.backendRefs is set", base)
	}

	m := serviceHost.FindStringSubmatch(migrate.Scalar(host))
	if m == nil {
		kind := "fqdn"
		if net.ParseIP(migrate.Scalar(host)) != nil {
			kind = "ip"
		}
		return errorf(field, "host %s:%d is not a Service DNS name, not migrated: set backendRefs to a Service, "+
			"or to a Backend with an %s endpoint (requires extensionApis.enableBackend)", migrate.Scalar(host), port, kind)
	}

	ref := migrate.Mapping("name", migrate.String(m[1]), "namespace", migrate.String(m[3]), "port", migrate.Int(int64(port)))
	migrate.Replace(sink, "host", "backendRefs", migrate.Sequence(ref))
	migrate.Delete(sink, "port")
	return migrated(field, "replaced host and port with a backendRef to Service %s/%s port %d; "+
		"Envoy now uses the Service endpoints instead of DNS, and a cross-namespace reference needs a ReferenceGrant",
		m[3], m[1], port)
}

// migrateTelemetryPort removes an explicit telemetry port that is set without host.
func migrateTelemetryPort(sink *yaml.Node, base string) migrate.Finding {
	field := base + ".port"
	if migrate.IsEmpty(migrate.Get(sink, "backendRefs")) {
		return errorf(field, "port is set without host or backendRefs, not migrated: set backendRefs instead")
	}
	migrate.Delete(sink, "port")
	return migrated(field, "removed, it was ignored because %s.backendRefs is set", base)
}

// telemetryPort returns the port of the telemetry sink, or the CRD default if it isn't set. It
// decodes the node, so that an alias resolves to the port of its anchor.
func telemetryPort(sink *yaml.Node) int {
	var v int
	if p := migrate.Get(sink, "port"); migrate.IsSet(sink, "port") && p.Decode(&v) == nil {
		return v
	}
	return defaultTelemetryPort
}
