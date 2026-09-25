// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package runner

import (
	"errors"
	"reflect"
	"slices"
	"sync"

	envoytypes "github.com/envoyproxy/go-control-plane/pkg/cache/types"
	cachev3 "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	resourcev3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/telepresenceio/watchable"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	extension "github.com/envoyproxy/gateway/internal/extension/types"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/logging"
	"github.com/envoyproxy/gateway/internal/message"
	"github.com/envoyproxy/gateway/internal/metrics"
	endpointsutil "github.com/envoyproxy/gateway/internal/utils/endpoints"
	"github.com/envoyproxy/gateway/internal/xds/cache"
	"github.com/envoyproxy/gateway/internal/xds/translator"
	xdstypes "github.com/envoyproxy/gateway/internal/xds/types"
)

var (
	endpointFastPathPatchTotal = metrics.NewCounter(
		"endpoint_fastpath_patch_total",
		"Total number of EDS-only snapshot patches applied by the endpoint fast path.",
	)

	endpointFastPathSkippedTotal = metrics.NewCounter(
		"endpoint_fastpath_skipped_total",
		"Total number of endpoint updates skipped by the endpoint fast path, by reason.",
	)

	skipReasonLabel = metrics.NewLabel("reason")
)

const (
	skipReasonNoContext         = "no_context"
	skipReasonDisabledForKey    = "disabled_for_key"
	skipReasonZeroTransition    = "zero_transition"
	skipReasonNoSnapshot        = "no_snapshot"
	skipReasonAddressTypeChange = "address_type_change"
)

// endpointFastPath propagates endpoint updates to Envoy as EDS-only snapshot patches,
// rebuilding the affected clusters' ClusterLoadAssignments from the contexts
// captured at the last successful full translation. Semantics are best-effort:
// fresh endpoints against last-known-good config, with full translations
// remaining authoritative.
type endpointFastPath struct {
	logger logging.Logger
	cache  cache.SnapshotCacheWithCallbacks

	// extensionModifiesEndpoints is true when a registered extension server
	// subscribes to a post-xDS hook that can modify clusters or endpoints,
	// in which case a regenerated CLA could lose the extension's changes and
	// the fast path stands down entirely.
	extensionModifiesEndpoints bool

	// mu serializes full-snapshot publishes and fast-path patches, so a patch
	// can never interleave with a snapshot rebuild or its context swap.
	mu sync.Mutex

	// edsContextsByIRKey maps each IR key to its backend index of ClusterLoadAssignment
	// contexts from the last successful full translation.
	edsContextsByIRKey map[string]map[string][]*xdstypes.ClusterLoadAssignmentContext

	// lastEndpoints holds the most recent endpoint update per backend key, so
	// it can be re-applied on top of a full snapshot that raced with it.
	lastEndpoints map[string]*message.EndpointUpdate
}

func newEndpointFastPath(c cache.SnapshotCacheWithCallbacks, extMgr extension.Manager, logger logging.Logger) *endpointFastPath {
	return &endpointFastPath{
		logger:                     logger,
		cache:                      c,
		extensionModifiesEndpoints: extensionModifiesEndpoints(extMgr),
		edsContextsByIRKey:         make(map[string]map[string][]*xdstypes.ClusterLoadAssignmentContext),
		lastEndpoints:              make(map[string]*message.EndpointUpdate),
	}
}

// extensionModifiesEndpoints reports whether a registered extension subscribes
// to a post-xDS hook that can modify clusters or endpoints. Extension mutations
// are applied in place and are invisible afterwards, so the decision must be
// made from hook registration, not from diffing results.
func extensionModifiesEndpoints(extMgr extension.Manager) bool {
	if extMgr == nil {
		return false
	}
	for _, hook := range []egv1a1.XDSTranslatorHook{egv1a1.XDSEndpoints, egv1a1.XDSCluster, egv1a1.XDSTranslation} {
		client, err := extMgr.GetPostXDSHookClient(hook)
		if err != nil || client != nil {
			return true
		}
	}
	return false
}

func endpointSourceKey(es *ir.EndpointSource) string {
	return message.BackendKey(es.Kind, es.Namespace, es.Name)
}

// updateEDSForFullSnapshot applies retained endpoint updates to a completed
// translation's EDS resources and replaces the stored EDS contexts. The caller
// must hold f.mu through snapshot publication and clear the contexts if publication fails.
func (f *endpointFastPath) updateEDSForFullSnapshot(irKey string, xdsIR *ir.Xds, table *xdstypes.ResourceVersionTable) {
	ctxs := table.EDSContexts
	// Exclude clusters targeted by endpoint patches before replaying retained
	// updates so neither replay nor later fast-path updates overwrite policy changes.
	if clustersWithPatches := clustersWithEndpointPatches(xdsIR); len(clustersWithPatches) > 0 {
		for backendKey, contexts := range ctxs {
			contexts = slices.DeleteFunc(contexts, func(ec *xdstypes.ClusterLoadAssignmentContext) bool {
				_, hasEndpointPatch := clustersWithPatches[ec.ClusterName]
				return hasEndpointPatch
			})
			if len(contexts) == 0 {
				delete(ctxs, backendKey)
			} else {
				ctxs[backendKey] = contexts
			}
		}
	}
	if !f.extensionModifiesEndpoints {
		for backendKey := range ctxs {
			update, ok := f.lastEndpoints[backendKey]
			if !ok {
				continue
			}
			// Mutations target the private contexts captured by the build; skips (zero
			// transition, address-type change) leave the build's own endpoint
			// state in place.
			assignments, skipReason := f.patchContexts(ctxs, update)
			if skipReason != "" {
				// Counted here too: this build publishes endpoint state the
				// fast path knows is already superseded.
				endpointFastPathSkippedTotal.With(skipReasonLabel.Value(skipReason)).Increment()
				continue
			}
			replaceEndpointResources(table, assignments)
		}
	}

	f.edsContextsByIRKey[irKey] = ctxs
}

// replaceEndpointResources swaps the given ClusterLoadAssignments into the
// translation table's EDS resources by cluster name.
func replaceEndpointResources(table *xdstypes.ResourceVersionTable, assignments []envoytypes.Resource) {
	if len(assignments) == 0 {
		return
	}
	byName := make(map[string]envoytypes.Resource, len(assignments))
	for _, assignment := range assignments {
		byName[cachev3.GetResourceName(assignment)] = assignment
	}
	eds := table.XdsResources[resourcev3.EndpointType]
	for i, res := range eds {
		if replacement, ok := byName[cachev3.GetResourceName(res)]; ok {
			eds[i] = replacement
		}
	}
}

// OnDelete drops the ClusterLoadAssignment contexts for a deleted IR key.
func (f *endpointFastPath) OnDelete(irKey string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.edsContextsByIRKey, irKey)
}

// clustersWithEndpointPatches returns the cluster names targeted by
// ClusterLoadAssignment patches in the IR's EnvoyPatchPolicies.
func clustersWithEndpointPatches(xdsIR *ir.Xds) map[string]struct{} {
	clusters := make(map[string]struct{})
	for _, policy := range xdsIR.EnvoyPatchPolicies {
		for _, patch := range policy.JSONPatches {
			if patch.Type == resourcev3.EndpointType && patch.Name != "" {
				clusters[patch.Name] = struct{}{}
			}
		}
	}
	return clusters
}

// subscribe consumes endpoint updates until the subscription context is
// cancelled. The EndpointUpdates map itself is never closed, because the
// provider writes it from informer handlers that Close cannot synchronize with.
func (f *endpointFastPath) subscribe(sub <-chan watchable.Snapshot[string, *message.EndpointUpdate], runnerName string) {
	message.HandleSubscription(
		f.logger,
		message.Metadata{Runner: runnerName, Message: message.EndpointSlicesMessageName}, sub,
		func(update message.Update[string, *message.EndpointUpdate], _ chan error) {
			message.PublishRunnerEventMetric(runnerName, update.Delete)
			if update.Delete || update.Value == nil {
				// The provider deletes an entry when the backend's last
				// EndpointSlice is gone; drop the retained state and leave the
				// transition to the full path.
				f.mu.Lock()
				delete(f.lastEndpoints, update.Key)
				f.mu.Unlock()
				return
			}
			f.handleUpdate(update.Value)
		},
	)
	f.logger.Info("endpoint fast path subscriber shutting down")
}

// handleUpdate applies one endpoint update: for every IR key whose contexts
// reference the updated backend, the affected clusters' CLAs are rebuilt and
// patched into the current snapshot with only the EDS version bumped.
func (f *endpointFastPath) handleUpdate(update *message.EndpointUpdate) {
	logger := f.logger

	f.mu.Lock()
	defer f.mu.Unlock()

	backendKey := update.Key()
	f.lastEndpoints[backendKey] = update

	found := false
	for irKey, ctxs := range f.edsContextsByIRKey {
		if _, ok := ctxs[backendKey]; !ok {
			continue
		}
		found = true
		if f.extensionModifiesEndpoints {
			endpointFastPathSkippedTotal.With(skipReasonLabel.Value(skipReasonDisabledForKey)).Increment()
			continue
		}
		assignments, skipReason := f.patchContexts(ctxs, update)
		if skipReason != "" {
			endpointFastPathSkippedTotal.With(skipReasonLabel.Value(skipReason)).Increment()
			continue
		}
		if len(assignments) == 0 {
			// The endpoint state is already reflected in the snapshot (e.g. the
			// full path got there first); nothing to patch.
			continue
		}
		// The full reconcile enqueued by the same endpoint event is the backstop
		// for both failure paths, so the contexts are left describing the new
		// endpoints either way.
		switch err := f.cache.UpdateEndpointResources(irKey, assignments); {
		case errors.Is(err, cache.ErrNoSnapshot):
			endpointFastPathSkippedTotal.With(skipReasonLabel.Value(skipReasonNoSnapshot)).Increment()
		case err != nil:
			logger.Error(err, "endpoint fast path: failed to patch endpoint resources", "key", irKey, "backend", backendKey)
		default:
			endpointFastPathPatchTotal.Increment()
			logger.Info("endpoint fast path: patched endpoints", "key", irKey, "backend", backendKey, "clusters", len(assignments))
		}
	}
	if !found {
		endpointFastPathSkippedTotal.With(skipReasonLabel.Value(skipReasonNoContext)).Increment()
	}
}

// patchContexts recomputes the endpoints fed by the updated backend and returns
// the rebuilt CLAs of the clusters that changed. A non-empty skip reason means
// the contexts were left untouched: the change is not CLA-only and belongs to
// the full path. Must be called with f.mu held.
func (f *endpointFastPath) patchContexts(ctxs map[string][]*xdstypes.ClusterLoadAssignmentContext, update *message.EndpointUpdate) ([]envoytypes.Resource, string) {
	backendKey := update.Key()

	type pendingChange struct {
		ds        *ir.DestinationSetting
		endpoints []*ir.DestinationEndpoint
		addrType  *ir.DestinationAddressType
	}
	var (
		pending         []pendingChange
		changedClusters []*xdstypes.ClusterLoadAssignmentContext
		seen            = make(map[*xdstypes.ClusterLoadAssignmentContext]struct{})
	)

	for _, ec := range ctxs[backendKey] {
		if _, ok := seen[ec]; ok {
			continue
		}
		seen[ec] = struct{}{}
		changed := false
		for _, ds := range ec.Settings {
			if ds.EndpointSource == nil || endpointSourceKey(ds.EndpointSource) != backendKey {
				continue
			}
			endpoints, addrType := endpointsutil.EndpointsFromSlices(update.EndpointSlices, ds.EndpointSource.PortName, ds.EndpointSource.Protocol)
			// A setting captured in a context always had endpoints at full
			// translation time; dropping to zero restructures clusters, routes
			// and statuses, so it must take the full path.
			if len(endpoints) == 0 {
				return nil, skipReasonZeroTransition
			}
			if !equalAddressType(addrType, ds.AddressType) {
				return nil, skipReasonAddressTypeChange
			}
			if reflect.DeepEqual(endpoints, ds.Endpoints) {
				continue
			}
			pending = append(pending, pendingChange{
				ds:        ds,
				endpoints: endpoints,
				addrType:  addrType,
			})
			changed = true
		}
		if changed {
			changedClusters = append(changedClusters, ec)
		}
	}

	// All settings validated; apply the changes to the cached contexts so they
	// track the last-applied endpoint state, then rebuild the affected CLAs.
	for _, p := range pending {
		p.ds.Endpoints = p.endpoints
		p.ds.AddressType = p.addrType
	}
	assignments := make([]envoytypes.Resource, 0, len(changedClusters))
	for _, ec := range changedClusters {
		assignments = append(assignments, translator.BuildClusterLoadAssignment(ec))
	}
	return assignments, ""
}

func equalAddressType(a, b *ir.DestinationAddressType) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}
