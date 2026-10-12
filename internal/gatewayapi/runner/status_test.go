// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package runner

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/logging"
	"github.com/envoyproxy/gateway/internal/message"
)

// acceptedStatus is the shape a Gateway's status has after translation: conditions with
// an observedGeneration but no transition timestamps, which the provider stamps later.
// Two generations of the same Gateway name and spec produce exactly this value.
func acceptedStatus() *gwapiv1.GatewayStatus {
	return &gwapiv1.GatewayStatus{
		Conditions: []metav1.Condition{{
			Type:               string(gwapiv1.GatewayConditionAccepted),
			Status:             metav1.ConditionTrue,
			Reason:             string(gwapiv1.GatewayReasonAccepted),
			ObservedGeneration: 1,
		}},
	}
}

// statusSubscriber drains a status map's subscription into a channel.
type statusSubscriber struct {
	updates chan message.Update[types.NamespacedName, *gwapiv1.GatewayStatus]
}

func subscribeToGatewayStatuses(t *testing.T, pr *message.ProviderResources) *statusSubscriber {
	t.Helper()

	s := &statusSubscriber{
		updates: make(chan message.Update[types.NamespacedName, *gwapiv1.GatewayStatus], 16),
	}
	sub := pr.GatewayStatuses.Subscribe(t.Context())
	go message.HandleSubscription(
		logging.NewLogger(t.Output(), egv1a1.DefaultEnvoyGatewayLogging()),
		message.Metadata{Runner: "test", Message: message.GatewayStatusMessageName},
		sub,
		func(update message.Update[types.NamespacedName, *gwapiv1.GatewayStatus], _ chan error) {
			s.updates <- update
		},
	)
	return s
}

// awaitStore waits for a store update for key, skipping the delete that a forced
// republish emits. A delete and the store that follows it may arrive as one coalesced
// update or as two, depending on how the subscriber and the coalescer interleave, so
// only the store is asserted on.
func (s *statusSubscriber) awaitStore(t *testing.T, key types.NamespacedName) *gwapiv1.GatewayStatus {
	t.Helper()

	deadline := time.After(10 * time.Second)
	for {
		select {
		case update := <-s.updates:
			if update.Key != key || update.Delete {
				continue
			}
			return update.Value
		case <-deadline:
			t.Fatalf("timed out waiting for a status update for %s", key)
			return nil
		}
	}
}

// requireNoStore asserts that no store update for key arrives.
func (s *statusSubscriber) requireNoStore(t *testing.T, key types.NamespacedName) {
	t.Helper()

	deadline := time.After(500 * time.Millisecond)
	for {
		select {
		case update := <-s.updates:
			if update.Key != key || update.Delete {
				continue
			}
			t.Fatalf("expected no status update for %s, got %+v", key, update.Value)
		case <-deadline:
			return
		}
	}
}

// TestStoreStatusRepublishesWhenObjectIsReplaced covers a Gateway deleted and recreated
// under the same name within one translation cycle. The translated status is identical,
// so the subscriber-side coalescer drops the store as unchanged and the recreated object
// is left with the default status its CRD gave it. See envoyproxy/gateway#9536.
func TestStoreStatusRepublishesWhenObjectIsReplaced(t *testing.T) {
	pr := new(message.ProviderResources)
	sub := subscribeToGatewayStatuses(t, pr)
	key := types.NamespacedName{Namespace: "default", Name: "eg"}
	seen := map[types.NamespacedName]types.UID{}

	// The Gateway is translated for the first time.
	storeStatus(&pr.GatewayStatuses, seen, key, "uid-1", acceptedStatus())
	require.Equal(t, acceptedStatus(), sub.awaitStore(t, key))

	// The same Gateway is translated again, unchanged. Dedup must still apply, or every
	// reconcile would cost a status write per object.
	storeStatus(&pr.GatewayStatuses, seen, key, "uid-1", acceptedStatus())
	sub.requireNoStore(t, key)

	// The Gateway is replaced: same name, same spec, new object. The value is identical,
	// so only the identity distinguishes this from the no-op above.
	storeStatus(&pr.GatewayStatuses, seen, key, "uid-2", acceptedStatus())
	require.Equal(t, acceptedStatus(), sub.awaitStore(t, key))
	require.Equal(t, types.UID("uid-2"), seen[key])
}

// TestStoreStatusFirstPublishIsNotAReplacement guards the presence check: a key absent
// from the cache has never been published, so it must not be treated as replaced. Without
// the check every key would emit a pointless delete on the first translation cycle.
func TestStoreStatusFirstPublishIsNotAReplacement(t *testing.T) {
	pr := new(message.ProviderResources)
	sub := subscribeToGatewayStatuses(t, pr)
	key := types.NamespacedName{Namespace: "default", Name: "eg"}

	storeStatus(&pr.GatewayStatuses, map[types.NamespacedName]types.UID{}, key, "uid-1", acceptedStatus())

	require.Equal(t, acceptedStatus(), sub.awaitStore(t, key))
	select {
	case update := <-sub.updates:
		require.Failf(t, "unexpected update", "expected only a store, got %+v", update)
	default:
	}
}

// TestStoreStatusWithoutUIDs covers the file provider, which builds objects with no UID.
// Replacement cannot be detected there, so this stays a plain store.
func TestStoreStatusWithoutUIDs(t *testing.T) {
	pr := new(message.ProviderResources)
	sub := subscribeToGatewayStatuses(t, pr)
	key := types.NamespacedName{Namespace: "default", Name: "eg"}
	seen := map[types.NamespacedName]types.UID{}

	storeStatus(&pr.GatewayStatuses, seen, key, "", acceptedStatus())
	require.Equal(t, acceptedStatus(), sub.awaitStore(t, key))

	storeStatus(&pr.GatewayStatuses, seen, key, "", acceptedStatus())
	sub.requireNoStore(t, key)
}

// TestMergeCarriesObjectIdentity guards the identity the aggregated stores feed to
// storeStatus: a route or policy is merged once per GatewayClass, and the UID has to
// survive every merge for a replacement to be detectable.
func TestMergeCarriesObjectIdentity(t *testing.T) {
	first := &gwapiv1.RouteStatus{Parents: []gwapiv1.RouteParentStatus{{ControllerName: "a"}}}
	second := &gwapiv1.RouteStatus{Parents: []gwapiv1.RouteParentStatus{{ControllerName: "b"}}}

	entry := mergeAggregatedRouteStatus(aggregatedRouteStatus{}, first, 1, "uid-1")
	require.Equal(t, types.UID("uid-1"), entry.uid)

	entry = mergeAggregatedRouteStatus(entry, second, 1, "uid-1")
	require.Equal(t, types.UID("uid-1"), entry.uid)
	require.Len(t, entry.status.Parents, 2)

	// A nil incoming status returns early; the identity is still recorded.
	require.Equal(t, types.UID("uid-2"), mergeAggregatedRouteStatus(entry, nil, 1, "uid-2").uid)
}
