// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gateway

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-logr/logr"
)

// defaultGatewayName is the parentRefs Gateway name the generated HTTPRoutes
// attach to. It matches the Gateway defined in setup/routes/gateway.yaml.
const defaultGatewayName = "gmp-gateway"

// Generator builds Envoy route files from the service and route
// tables and writes them to the routes directory watched by the file provider.
type Generator struct {
	repo        *Repository
	routesDir   string
	gatewayName string
	// publish, when non-nil, is called after files are written to notify other
	// tasks (e.g. via Redis pub/sub) that they should reload.
	publish func(cluster string)
	logger  logr.Logger
}

// GeneratorConfig configures a Generator.
type GeneratorConfig struct {
	Repo      *Repository
	RoutesDir string
	// GatewayName overrides the parentRefs Gateway name. Defaults to gmp-gateway.
	GatewayName string
	Publish     func(cluster string)
	Logger      logr.Logger
}

// NewGenerator creates a Generator.
func NewGenerator(cfg *GeneratorConfig) *Generator {
	name := cfg.GatewayName
	if name == "" {
		name = defaultGatewayName
	}
	return &Generator{
		repo:        cfg.Repo,
		routesDir:   cfg.RoutesDir,
		gatewayName: name,
		publish:     cfg.Publish,
		logger:      cfg.Logger,
	}
}

// Generate builds route files for all active services in the given cluster,
// writes one file per service to the routes directory, publishes a reload
// signal, and returns the combined YAML.
func (g *Generator) Generate(ctx context.Context, cluster string) (string, error) {
	cluster = strings.ToUpper(strings.TrimSpace(cluster))

	cl, err := g.repo.GetClusterByName(ctx, cluster)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return "", fmt.Errorf("%w: cluster %q does not exist", ErrInvalid, cluster)
		}
		return "", err
	}

	services, err := g.repo.ListServices(ctx, "")
	if err != nil {
		return "", err
	}
	if len(services) == 0 {
		return "", fmt.Errorf("%w: no active services found", ErrInvalid)
	}

	var combined strings.Builder
	for _, svc := range services {
		routes, err := g.repo.ListRoutesForService(ctx, svc.ID)
		if err != nil {
			return "", err
		}

		doc := buildServiceYAML(svc, routes, g.gatewayName, cl.InternalLB)

		if err := g.writeFile(svc.Name, doc); err != nil {
			return "", err
		}

		combined.WriteString(doc)
	}

	// Notify other tasks to reload their copy from the shared volume.
	if g.publish != nil {
		g.publish(cluster)
	}

	return combined.String(), nil
}

// writeFile atomically writes a service's YAML document to <routesDir>/<slug>.yaml.
func (g *Generator) writeFile(serviceName, content string) error {
	if g.routesDir == "" {
		return fmt.Errorf("routes directory is not configured")
	}
	if err := os.MkdirAll(g.routesDir, 0o755); err != nil {
		return fmt.Errorf("failed to create routes directory: %w", err)
	}

	target := filepath.Join(g.routesDir, slug(serviceName)+".yaml")
	tmp := target + ".tmp"

	if err := os.WriteFile(tmp, []byte(content), 0o644); err != nil {
		return fmt.Errorf("failed to write %s: %w", tmp, err)
	}
	// Rename is atomic within the same directory on POSIX filesystems, so the
	// file provider never observes a partially written file.
	if err := os.Rename(tmp, target); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("failed to move %s into place: %w", target, err)
	}

	g.logger.Info("wrote route file", "service", serviceName, "path", target)
	return nil
}

// buildServiceYAML renders one Backend plus one HTTPRoute (with Exact and
// PathPrefix match groups) for a service and its routes. The backend hostname
// (internalLB) comes from the cluster being generated; the port from the service.
func buildServiceYAML(svc serviceEntity, routes []routeEntity, gatewayName, internalLB string) string {
	name := slug(svc.Name)

	var b strings.Builder

	// Backend: fqdn endpoint (internalLB is a DNS hostname, http assumed).
	b.WriteString("# Generated from the cluster/service/route tables. Do not edit by hand.\n")
	b.WriteString("apiVersion: gateway.envoyproxy.io/v1alpha1\n")
	b.WriteString("kind: Backend\n")
	b.WriteString("metadata:\n")
	b.WriteString("  name: " + name + "\n")
	b.WriteString("spec:\n")
	b.WriteString("  endpoints:\n")
	b.WriteString("    - fqdn:\n")
	b.WriteString("        hostname: " + internalLB + "\n")
	b.WriteString(fmt.Sprintf("        port: %d\n", svc.Port))

	// Split routes into Exact and PathPrefix groups.
	var exact, prefix []routeEntity
	for _, r := range routes {
		switch r.MatchType {
		case MatchTypeExact:
			exact = append(exact, r)
		case MatchTypePathPrefix:
			prefix = append(prefix, r)
		}
	}

	// Only emit an HTTPRoute if the service has at least one route.
	if len(exact) == 0 && len(prefix) == 0 {
		return b.String()
	}

	b.WriteString("\n---\n")
	b.WriteString("apiVersion: gateway.networking.k8s.io/v1\n")
	b.WriteString("kind: HTTPRoute\n")
	b.WriteString("metadata:\n")
	b.WriteString("  name: " + name + "\n")
	b.WriteString("spec:\n")
	b.WriteString("  parentRefs:\n")
	b.WriteString("    - name: " + gatewayName + "\n")
	b.WriteString("  rules:\n")

	writeRule := func(group []routeEntity) {
		b.WriteString("    - backendRefs:\n")
		b.WriteString("        - group: gateway.envoyproxy.io\n")
		b.WriteString("          kind: Backend\n")
		b.WriteString("          name: " + name + "\n")
		b.WriteString("      matches:\n")
		for _, r := range group {
			b.WriteString("        - path:\n")
			b.WriteString("            type: " + r.MatchType + "\n")
			b.WriteString("            value: " + r.Path + "\n")
		}
	}

	if len(exact) > 0 {
		writeRule(exact)
	}
	if len(prefix) > 0 {
		writeRule(prefix)
	}

	return b.String()
}

// slug converts a service name into a DNS-safe identifier: lowercase, with any
// run of non-alphanumeric characters collapsed to a single hyphen and leading/
// trailing hyphens trimmed.
func slug(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	var b strings.Builder
	prevHyphen := false
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			prevHyphen = false
		} else if !prevHyphen {
			b.WriteRune('-')
			prevHyphen = true
		}
	}
	return strings.Trim(b.String(), "-")
}
