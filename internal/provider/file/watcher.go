// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package file

import (
	"context"
	"time"
)

const (
	// defaultWatchTicker defines default ticker (in seconds) for watcher.
	// TODO(sh2): make it configurable
	defaultWatchTicker = 3
)

type watcher struct {
	paths []string

	ticker *time.Ticker
}

func newWatcher(paths []string) *watcher {
	return &watcher{
		paths:  paths,
		ticker: time.NewTicker(defaultWatchTicker * time.Second),
	}
}

// Watch watches and loads files from paths.
func (w *watcher) Watch(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			w.ticker.Stop()
			return nil
		case <-w.ticker.C:
			w.watch()
		default:
		}
	}
}

func (w *watcher) watch() {
	// TODO: implement watch logic
}
