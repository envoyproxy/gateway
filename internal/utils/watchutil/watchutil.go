package watchutil

import (
	"github.com/telepresenceio/watchable"
)

// HandleSubscription takes a channel returned by
// watchable.Map.Subscribe() (or .SubscribeSubset()), and calls the
// given function for each initial value in the map, and for any
// updates.
//
// This is better than simply iterating over snapshot.Updates because
// it handles the case where the the watchable.Map already contains
// entries before .Subscribe is called.
func HandleSubscription[K comparable, V any](
	subscription <-chan watchable.Snapshot[K, V],
	handle func(watchable.Update[K, V]),
) {
	if snapshot, ok := <-subscription; ok {
		for k, v := range snapshot.State {
			handle(watchable.Update[K, V]{
				Key:   k,
				Value: v,
			})
		}
	}
	for snapshot := range subscription {
		for _, update := range snapshot.Updates {
			handle(update)
		}
	}
}
