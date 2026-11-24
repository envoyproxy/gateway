// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

// Package deprecated migrates the deprecated Envoy Gateway API fields to their replacements.
package deprecated

import (
	"cmp"
	"context"
	"fmt"
	"maps"
	"slices"
	"strings"

	"go.yaml.in/yaml/v3"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/cmd/egctl/migrate"
)

const (
	egAPIVersion = egv1a1.GroupName + "/v1alpha1"

	// envoyGatewayConfigMap is the ConfigMap that holds the EnvoyGateway configuration.
	envoyGatewayConfigMap = "envoy-gateway-config"
	// envoyGatewayConfigKey is the ConfigMap key that holds the EnvoyGateway configuration.
	envoyGatewayConfigKey = "envoy-gateway.yaml"

	helmValuesHint = "update config.envoyGateway in the Helm values"
)

// Path migrates the deprecated API fields tracked in
// https://github.com/envoyproxy/gateway/issues/10211.
type Path struct {
	// EnvoyGatewayNamespace is the namespace of the envoy-gateway-config ConfigMap, used with
	// --from-cluster.
	EnvoyGatewayNamespace string
}

var _ migrate.Path = (*Path)(nil)

// Fetch lists the policies and EnvoyProxy resources in namespace (all namespaces if empty), and
// the envoy-gateway-config ConfigMap.
func (p *Path) Fetch(ctx context.Context, cli client.Client, namespace string) ([]client.Object, error) {
	var objs []client.Object
	// The objects are read as unstructured, so that the deprecated fields are kept even when this
	// egctl is built from an API that no longer has them.
	for _, kind := range migratedKinds {
		list := &unstructured.UnstructuredList{}
		list.SetGroupVersionKind(egv1a1.GroupVersion.WithKind(kind + "List"))
		if err := cli.List(ctx, list, client.InNamespace(namespace)); err != nil {
			return nil, fmt.Errorf("listing %s resources: %w", kind, err)
		}
		for i := range list.Items {
			objs = append(objs, &list.Items[i])
		}
	}

	if namespace == "" || namespace == p.EnvoyGatewayNamespace {
		cm := &corev1.ConfigMap{}
		err := cli.Get(ctx, client.ObjectKey{Namespace: p.EnvoyGatewayNamespace, Name: envoyGatewayConfigMap}, cm)
		switch {
		case err == nil:
			objs = append(objs, cm)
		case !kerrors.IsNotFound(err):
			return nil, fmt.Errorf("getting ConfigMap %s/%s: %w", p.EnvoyGatewayNamespace, envoyGatewayConfigMap, err)
		}
	}
	return objs, nil
}

// migratedKinds are the kinds whose deprecated fields are rewritten.
var migratedKinds = []string{
	egv1a1.KindClientTrafficPolicy,
	egv1a1.KindBackendTrafficPolicy,
	egv1a1.KindSecurityPolicy,
	egv1a1.KindEnvoyExtensionPolicy,
	egv1a1.KindEnvoyProxy,
}

// Migrate rewrites the deprecated fields of obj.
func (p *Path) Migrate(obj *migrate.Object) ([]migrate.Finding, error) {
	switch {
	case obj.APIVersion == egAPIVersion:
		return migrateEnvoyGatewayResource(obj)
	case obj.APIVersion == "v1" && obj.Kind == "ConfigMap":
		return reportConfigMap(obj)
	case obj.APIVersion == "" && obj.Kind == "":
		// Helm values of the gateway-helm chart.
		if config := migrate.Lookup(obj.Root, "config.envoyGateway"); config != nil && config.Kind == yaml.MappingNode {
			return reportEnvoyGateway("config.envoyGateway.", config, "update it in the Helm values"), nil
		}
	}
	return nil, nil
}

func migrateEnvoyGatewayResource(obj *migrate.Object) ([]migrate.Finding, error) {
	if obj.Kind == egv1a1.KindEnvoyGateway {
		return reportEnvoyGateway("", obj.Root, helmValuesHint), nil
	}
	if !slices.Contains(migratedKinds, obj.Kind) {
		return nil, nil
	}

	// Null deprecated keys have no value to transfer. Remove them before applying rules,
	// keeping replacement values intact. Repeat because a host finding can hide its port.
	var findings []migrate.Finding
	for {
		removed := false
		for _, path := range slices.Sorted(maps.Keys(detect(obj.Kind, obj.Root))) {
			i := strings.LastIndex(path, ".")
			parent, key := path[:i], path[i+1:]
			n := migrate.Lookup(obj.Root, parent)
			if migrate.Has(n, key) && !migrate.IsSet(n, key) {
				migrate.Delete(n, key)
				findings = append(findings, migrated(path, "removed null deprecated field"))
				removed = true
			}
		}
		if !removed {
			break
		}
	}
	fields := detect(obj.Kind, obj.Root)
	rewritten, unmigrated := rewrite(obj.Root, slices.Collect(maps.Keys(fields)))

	findings = append(findings, rewritten...)
	slices.SortStableFunc(findings, func(a, b migrate.Finding) int { return cmp.Compare(a.Field, b.Field) })

	// Safety net: every field that was reported as migrated must be gone.
	after := detect(obj.Kind, obj.Root)
	for _, path := range slices.Sorted(maps.Keys(after)) {
		if !unmigrated[path] {
			return nil, fmt.Errorf("%s is still deprecated after migration, please report this as a bug", path)
		}
	}
	return findings, nil
}

func reportConfigMap(obj *migrate.Object) ([]migrate.Finding, error) {
	data := migrate.GetMapping(obj.Root, "data")
	v := migrate.Get(data, envoyGatewayConfigKey)
	if v == nil || v.Kind != yaml.ScalarNode {
		return nil, nil
	}
	var doc yaml.Node
	if err := yaml.Unmarshal([]byte(v.Value), &doc); err != nil {
		return nil, fmt.Errorf("decoding %s: %w", envoyGatewayConfigKey, err)
	}
	if len(doc.Content) == 0 {
		return nil, nil
	}
	eg := doc.Content[0]
	if kind := migrate.Get(eg, "kind"); kind == nil || kind.Value != egv1a1.KindEnvoyGateway {
		return nil, nil
	}
	return reportEnvoyGateway("data."+envoyGatewayConfigKey+":", eg, helmValuesHint), nil
}

// reportEnvoyGateway reports every deprecated field of the EnvoyGateway configuration as an error.
// The configuration is usually rendered from the Helm values, so a rewrite here would be lost.
func reportEnvoyGateway(prefix string, eg *yaml.Node, hint string) []migrate.Finding {
	warnings := detectEnvoyGateway(eg)
	findings := make([]migrate.Finding, 0, len(warnings))
	for _, w := range warnings {
		findings = append(findings, migrate.Finding{
			Field:  prefix + w.path,
			Result: migrate.ResultError,
			Note:   fmt.Sprintf("%s; EnvoyGateway configuration is not migrated automatically, %s", w.message, hint),
		})
	}
	return findings
}

// rule rewrites the node at a base path.
type rule int

// The order of the rules is the order they're applied in: backendRef is converted to backendRefs
// before the telemetry rules look at backendRefs.
const (
	ruleBackendRef rule = iota
	ruleTargetRef
	ruleProxyProtocol
	ruleXRequestID
	ruleCompression
	ruleClientValidation
	ruleLuaValidation
	ruleConsistentHash
	ruleTelemetryHost
	ruleTelemetryPort
	ruleUnknown
)

type group struct {
	rule rule
	base string
	keys []string
}

// classify maps a deprecated field path reported by detect to its rule and the path
// of the node the rule rewrites.
func classify(path string) (rule, string) {
	switch path {
	case "spec.targetRef":
		return ruleTargetRef, "spec"
	case "spec.enableProxyProtocol":
		return ruleProxyProtocol, "spec"
	case "spec.headers.preserveXRequestID":
		return ruleXRequestID, "spec.headers"
	case "spec.compression":
		return ruleCompression, "spec"
	case "spec.tls.clientValidation.optional":
		return ruleClientValidation, "spec.tls.clientValidation"
	case "spec.luaValidation":
		return ruleLuaValidation, "spec"
	}
	for suffix, r := range map[string]rule{
		".consistentHash.type: Header": ruleConsistentHash,
		".consistentHash.header":       ruleConsistentHash,
	} {
		if base, ok := strings.CutSuffix(path, suffix); ok {
			return r, base + ".consistentHash"
		}
	}
	for suffix, r := range map[string]rule{
		".backendRef": ruleBackendRef,
		".host":       ruleTelemetryHost,
		".port":       ruleTelemetryPort,
	} {
		if base, ok := strings.CutSuffix(path, suffix); ok {
			return r, base
		}
	}
	return ruleUnknown, path
}

// rewrite applies the rules for the deprecated field paths to root. It returns the findings, and
// the paths that were left unchanged because they couldn't be migrated.
func rewrite(root *yaml.Node, paths []string) ([]migrate.Finding, map[string]bool) {
	groups := map[string]*group{}
	for _, p := range paths {
		r, base := classify(p)
		k := fmt.Sprintf("%d|%s", r, base)
		if groups[k] == nil {
			groups[k] = &group{rule: r, base: base}
		}
		groups[k].keys = append(groups[k].keys, p)
	}
	ordered := slices.SortedFunc(maps.Values(groups), func(a, b *group) int {
		return cmp.Or(cmp.Compare(a.rule, b.rule), cmp.Compare(a.base, b.base))
	})

	findings := make([]migrate.Finding, 0, len(ordered))
	unmigrated := map[string]bool{}
	for _, g := range ordered {
		f := apply(root, g)
		if f.Result == migrate.ResultError {
			for _, k := range g.keys {
				unmigrated[k] = true
			}
		}
		findings = append(findings, f)
	}
	slices.SortStableFunc(findings, func(a, b migrate.Finding) int { return cmp.Compare(a.Field, b.Field) })
	return findings, unmigrated
}

func apply(root *yaml.Node, g *group) migrate.Finding {
	n := migrate.Lookup(root, g.base)
	if g.rule == ruleUnknown {
		return errorf(g.base, "deprecated, but there is no automatic migration for it")
	}
	if n == nil || n.Kind != yaml.MappingNode {
		return errorf(g.keys[0], "unexpected shape, not migrated")
	}
	switch g.rule {
	case ruleBackendRef:
		return migrateBackendRef(n, g.base)
	case ruleTargetRef:
		return migrateTargetRef(n, g.base)
	case ruleProxyProtocol:
		return migrateProxyProtocol(n, g.base)
	case ruleXRequestID:
		return migrateXRequestID(n, g.base)
	case ruleCompression:
		return migrateCompression(n, g.base)
	case ruleClientValidation:
		return migrateClientValidation(n, g.base)
	case ruleLuaValidation:
		return migrateLuaValidation(n, g.base)
	case ruleConsistentHash:
		return migrateConsistentHash(n, g.base)
	case ruleTelemetryHost:
		return migrateTelemetryHost(n, g.base)
	case ruleTelemetryPort:
		return migrateTelemetryPort(n, g.base)
	}
	return errorf(g.base, "deprecated, but there is no automatic migration for it")
}

func migrated(field, note string, args ...any) migrate.Finding {
	return migrate.Finding{Field: field, Result: migrate.ResultMigrated, Note: fmt.Sprintf(note, args...)}
}

func errorf(field, note string, args ...any) migrate.Finding {
	return migrate.Finding{Field: field, Result: migrate.ResultError, Note: fmt.Sprintf(note, args...)}
}
