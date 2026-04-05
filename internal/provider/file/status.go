// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package file

import (
	"context"
	"fmt"
	"sync"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/yaml"

	"github.com/envoyproxy/gateway/internal/provider/kubernetes"
)

type StatusHandler struct {
	logger        logr.Logger
	updateChannel chan kubernetes.Update
	wg            *sync.WaitGroup
}

func NewStatusHandler(log logr.Logger) *StatusHandler {
	u := &StatusHandler{
		logger:        log,
		updateChannel: make(chan kubernetes.Update, 1000),
		wg:            new(sync.WaitGroup),
	}

	u.wg.Add(1)

	return u
}

// Start runs the goroutine to perform status writes.
func (u *StatusHandler) Start(ctx context.Context, ready *sync.WaitGroup) {
	u.logger.Info("started status update handler")
	defer u.logger.Info("stopped status update handler")

	// Enable Updaters to start sending updates to this handler.
	u.wg.Done()
	ready.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case update := <-u.updateChannel:
			u.logger.Info("received a status update",
				"kind", kubernetes.KindOf(update.Resource),
				"namespace", update.NamespacedName.Namespace,
				"name", update.NamespacedName.Name,
			)

			u.logStatus(update)
		}
	}
}

func (u *StatusHandler) logStatus(update kubernetes.Update) {
	obj := update.Resource
	newObj := update.Mutator.Mutate(obj)
	log := u.logger.WithValues("key", update.NamespacedName.String())

	// Log the resource status.
	raw, err := runtime.DefaultUnstructuredConverter.ToUnstructured(newObj)
	if err != nil {
		log.Error(err, "failed to convert object")
		return
	}

	rawStatus, ok := raw["status"]
	if !ok {
		log.Error(fmt.Errorf("no status field"), "failed to log status")
		return
	}

	byteStatus, err := yaml.Marshal(rawStatus)
	if err != nil {
		log.Error(err, "failed to marshal object")
		return
	}

	log.Info(fmt.Sprintf("Got new status for %s\n%s", kubernetes.KindOf(obj), string(byteStatus)))
}

// Writer retrieves the interface that should be used to write to the StatusHandler.
func (u *StatusHandler) Writer() kubernetes.Updater {
	return &StatusWriter{
		updateChannel: u.updateChannel,
		wg:            u.wg,
	}
}

// StatusWriter takes status updates and sends these to the StatusHandler via a channel.
type StatusWriter struct {
	updateChannel chan<- kubernetes.Update
	wg            *sync.WaitGroup
}

// Send sends the given Update off to the update channel for writing by the StatusHandler.
func (u *StatusWriter) Send(update kubernetes.Update) {
	// Wait until updater is ready
	u.wg.Wait()
	u.updateChannel <- update
}
