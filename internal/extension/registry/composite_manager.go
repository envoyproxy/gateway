// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package registry

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	extTypes "github.com/envoyproxy/gateway/internal/extension/types"
)

var _ extTypes.Manager = (*CompositeManager)(nil)

// namedManager pairs a Manager with its name and declared resource/policy
// Group/Kinds. Version is intentionally dropped: the rest of the pipeline
// (runner.ExtensionGroupKinds, Manager.HasExtension, Gateway API
// extensionRef) matches extensions by group+kind only.
type namedManager struct {
	name            string
	manager         extTypes.Manager
	resourceGKSet   sets.Set[schema.GroupKind] // Resources + BackendResources GKs
	policyGKSet     sets.Set[schema.GroupKind] // PolicyResources GKs
	cleanupHookConn func()
}

// CompositeManager wraps multiple Manager instances and implements the Manager interface.
// It chains extension calls sequentially: each extension's output becomes the next extension's input.
type CompositeManager struct {
	managers []namedManager
}

// NewCompositeManager creates a CompositeManager from a list of named managers.
func NewCompositeManager(managers []namedManager) *CompositeManager {
	return &CompositeManager{managers: managers}
}

// HasExtension returns true if any child manager has the extension (union semantics).
func (c *CompositeManager) HasExtension(g gwapiv1.Group, k gwapiv1.Kind) bool {
	for _, nm := range c.managers {
		if nm.manager.HasExtension(g, k) {
			return true
		}
	}
	return false
}

// FailOpen returns true only if all children are fail-open (conservative default).
func (c *CompositeManager) FailOpen() bool {
	for _, nm := range c.managers {
		if !nm.manager.FailOpen() {
			return false
		}
	}
	return true
}

// GetTranslationHookConfig merges configs using OR semantics:
// a resource type is included if any manager enables it.
// The ShouldInclude* helpers handle nil/default values correctly
// (clusters/secrets default to true; listeners/routes default to false).
func (c *CompositeManager) GetTranslationHookConfig() *egv1a1.TranslationConfig {
	hasAny := false
	for _, nm := range c.managers {
		if nm.manager.GetTranslationHookConfig() != nil {
			hasAny = true
			break
		}
	}
	if !hasAny {
		return nil
	}

	includeListeners := false
	includeRoutes := false
	includeClusters := false
	includeSecrets := false

	for _, nm := range c.managers {
		tc := nm.manager.GetTranslationHookConfig()
		if tc.ShouldIncludeListeners() {
			includeListeners = true
		}
		if tc.ShouldIncludeRoutes() {
			includeRoutes = true
		}
		if tc.ShouldIncludeClusters() {
			includeClusters = true
		}
		if tc.ShouldIncludeSecrets() {
			includeSecrets = true
		}
	}

	merged := &egv1a1.TranslationConfig{}
	if includeListeners {
		merged.Listener = &egv1a1.ListenerTranslationConfig{IncludeAll: new(true)}
	}
	if includeRoutes {
		merged.Route = &egv1a1.RouteTranslationConfig{IncludeAll: new(true)}
	}
	if includeClusters {
		merged.Cluster = &egv1a1.ClusterTranslationConfig{IncludeAll: new(true)}
	}
	if includeSecrets {
		merged.Secret = &egv1a1.SecretTranslationConfig{IncludeAll: new(true)}
	}

	return merged
}

// hookClientGetter abstracts GetPreXDSHookClient/GetPostXDSHookClient to allow shared collection logic.
type hookClientGetter func(extTypes.Manager, egv1a1.XDSTranslatorHook) (extTypes.XDSHookClient, error)

// GetPreXDSHookClient returns a compositeXDSHookClient that chains all child clients
// for the given hook type.
func (c *CompositeManager) GetPreXDSHookClient(xdsHookType egv1a1.XDSTranslatorHook) (extTypes.XDSHookClient, error) {
	return c.collectHookClients(xdsHookType, func(m extTypes.Manager, h egv1a1.XDSTranslatorHook) (extTypes.XDSHookClient, error) {
		return m.GetPreXDSHookClient(h)
	}, false)
}

// GetPostXDSHookClient returns a compositeXDSHookClient that chains all child clients
// for the given hook type.
func (c *CompositeManager) GetPostXDSHookClient(xdsHookType egv1a1.XDSTranslatorHook) (extTypes.XDSHookClient, error) {
	return c.collectHookClients(xdsHookType, func(m extTypes.Manager, h egv1a1.XDSTranslatorHook) (extTypes.XDSHookClient, error) {
		return m.GetPostXDSHookClient(h)
	}, true)
}

// collectHookClients iterates over all child managers, collects hook clients via the given getter,
// and returns a compositeXDSHookClient. If includeTranslationConfig is true, each entry's
// translationConfig is populated from the manager.
func (c *CompositeManager) collectHookClients(
	xdsHookType egv1a1.XDSTranslatorHook,
	getter hookClientGetter,
	includeTranslationConfig bool,
) (extTypes.XDSHookClient, error) {
	var entries []hookClientEntry
	for _, nm := range c.managers {
		client, err := getter(nm.manager, xdsHookType)
		if err != nil {
			if nm.manager.FailOpen() {
				continue
			}
			return nil, err
		}
		if client != nil {
			entry := hookClientEntry{
				name:          nm.name,
				client:        client,
				failOpen:      nm.manager.FailOpen(),
				resourceGKSet: nm.resourceGKSet,
				policyGKSet:   nm.policyGKSet,
			}
			if includeTranslationConfig {
				entry.translationConfig = nm.manager.GetTranslationHookConfig()
			}
			entries = append(entries, entry)
		}
	}
	if len(entries) == 0 {
		return nil, nil
	}
	return &compositeXDSHookClient{entries: entries}, nil
}

// CleanupHookConns closes all gRPC connections for all child managers.
func (c *CompositeManager) CleanupHookConns() {
	for _, nm := range c.managers {
		if nm.cleanupHookConn != nil {
			nm.cleanupHookConn()
		}
	}
}
