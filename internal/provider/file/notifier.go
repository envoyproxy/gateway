// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package file

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/util/sets"
)

const (
	defaultCleanUpRemoveEventsPeriod = 300 * time.Millisecond
)

type Notifier struct {
	// Events record events used to update ResourcesStore,
	// which only include two types of events: Write/Remove.
	Events chan fsnotify.Event

	filesWatcher              *fsnotify.Watcher
	dirsWatcher               *fsnotify.Watcher
	cleanUpRemoveEventsPeriod time.Duration

	logger logr.Logger
}

func NewNotifier(logger logr.Logger) (*Notifier, error) {
	fw, err := fsnotify.NewBufferedWatcher(10)
	if err != nil {
		return nil, err
	}

	dw, err := fsnotify.NewBufferedWatcher(10)
	if err != nil {
		return nil, err
	}

	return &Notifier{
		Events:                    make(chan fsnotify.Event),
		filesWatcher:              fw,
		dirsWatcher:               dw,
		cleanUpRemoveEventsPeriod: defaultCleanUpRemoveEventsPeriod,
		logger:                    logger,
	}, nil
}

func (n *Notifier) Watch(ctx context.Context, dirs, files sets.Set[string]) {
	n.watchDirs(ctx, dirs)
	n.watchFiles(ctx, files)
}

func (n *Notifier) Close() error {
	if err := n.filesWatcher.Close(); err != nil {
		return err
	}
	if err := n.dirsWatcher.Close(); err != nil {
		return err
	}
	return nil
}

// watchFiles watches one or more files, but instead of watching the file directly,
// it watches its parent directory. This solves various issues where files are
// frequently renamed.
func (n *Notifier) watchFiles(ctx context.Context, files sets.Set[string]) {
	if len(files) < 1 {
		return
	}

	go n.runFilesWatcher(ctx, files)

	for p := range files {
		if err := n.filesWatcher.Add(filepath.Dir(p)); err != nil {
			n.logger.Error(err, "error adding file to notifier", "path", p)

			continue
		}
	}
}

func (n *Notifier) runFilesWatcher(ctx context.Context, files sets.Set[string]) {
	var (
		cleanUpTicker = time.NewTicker(n.cleanUpRemoveEventsPeriod)

		// This map records the exact previous Op of one event.
		preEventOp = make(map[string]fsnotify.Op)
		// This set records the name of event that related to Remove Op.
		curRemoveEvents = sets.NewString()
	)

	for {
		select {
		case <-ctx.Done():
			return

		case err, ok := <-n.filesWatcher.Errors:
			if !ok {
				return
			}
			n.logger.Error(err, "error from files watcher in notifier")

		case event, ok := <-n.filesWatcher.Events:
			if !ok {
				return
			}

			// Ignore file and operation the watcher not interested in.
			if !files.Has(event.Name) || event.Has(fsnotify.Chmod) {
				continue
			}

			// This logic is trying to avoid files be removed and then created
			// frequently by considering Remove/Rename and the follow Create
			// Op as one Write Notifier.Event.
			//
			// Actually, this approach is also suitable for commands like vi/vim.
			// It creates a temporary file, removes the existing one and replace
			// it with the temporary file when file is saved. So instead of Write
			// Op, the watcher will receive Rename and Create Op.

			var writeEvent bool
			switch event.Op {
			case fsnotify.Create:
				if op, ok := preEventOp[event.Name]; ok &&
					op.Has(fsnotify.Rename) || op.Has(fsnotify.Remove) {
					writeEvent = true
					// If the exact previous Op of Create is Rename/Remove,
					// then consider them as a Write Notifier.Event instead of Remove.
					curRemoveEvents.Delete(event.Name)
				}
			case fsnotify.Write:
				writeEvent = true
			case fsnotify.Remove, fsnotify.Rename:
				curRemoveEvents.Insert(event.Name)
			}

			if writeEvent {
				n.logger.Info("sending write event",
					"name", event.Name, "watcher", "files")

				n.Events <- fsnotify.Event{
					Name: event.Name,
					Op:   fsnotify.Write,
				}
			}
			preEventOp[event.Name] = event.Op

		case <-cleanUpTicker.C:
			// As for collected Remove Notifier.Event, clean them up
			// in a period of time to avoid neglect of dealing with
			// Remove/Rename Op.
			for e := range curRemoveEvents {
				n.logger.Info("sending remove event",
					"name", e, "watcher", "files")

				n.Events <- fsnotify.Event{
					Name: e,
					Op:   fsnotify.Remove,
				}
			}
			curRemoveEvents = sets.NewString()
		}
	}
}

// watchDirs watches one or more directories.
func (n *Notifier) watchDirs(ctx context.Context, dirs sets.Set[string]) {
	if len(dirs) < 1 {
		return
	}

	// This map maintains the subdirectories ignored by each directory.
	ignoredSubDirs := make(map[string]sets.Set[string])

	for p := range dirs {
		if err := n.dirsWatcher.Add(p); err != nil {
			n.logger.Error(err, "error adding dir to notifier", "path", p)

			continue
		}

		// Find current exist subdirectories to init ignored subdirectories set.
		entries, err := os.ReadDir(p)
		if err != nil {
			n.logger.Error(err, "error reading dir in notifier", "path", p)

			if err = n.dirsWatcher.Remove(p); err != nil {
				n.logger.Error(err, "error removing dir from notifier", "path", p)
			}

			continue
		}

		ignoredSubDirs[p] = sets.New[string]()
		for _, entry := range entries {
			if entry.IsDir() {
				// The entry name is dir name, not dir path.
				ignoredSubDirs[p].Insert(entry.Name())
			}
		}
	}

	go n.runDirsWatcher(ctx, ignoredSubDirs)
}

func (n *Notifier) runDirsWatcher(ctx context.Context, ignoredSubDirs map[string]sets.Set[string]) {
	var (
		cleanUpTicker = time.NewTicker(n.cleanUpRemoveEventsPeriod)

		// This map records the exact previous Op of one event.
		preEventOp = make(map[string]fsnotify.Op)
		// This set records the name of event that related to Remove Op.
		curRemoveEvents = sets.NewString()
	)

	for {
		select {
		case <-ctx.Done():
			return

		case err, ok := <-n.dirsWatcher.Errors:
			if !ok {
				return
			}
			n.logger.Error(err, "error from dirs watcher in notifier")

		case event, ok := <-n.dirsWatcher.Events:
			if !ok {
				return
			}

			// Ignore the hidden or temporary file related event.
			_, name := filepath.Split(event.Name)
			if event.Has(fsnotify.Chmod) ||
				strings.HasPrefix(name, ".") ||
				strings.HasSuffix(name, "~") {
				continue
			}

			// Ignore any subdirectory related event.
			switch event.Op {
			case fsnotify.Create:
				if fi, err := os.Lstat(event.Name); err == nil && fi.IsDir() {
					parentDir := filepath.Dir(event.Name)
					if _, ok := ignoredSubDirs[parentDir]; ok {
						ignoredSubDirs[parentDir].Insert(name)
						continue
					}
				}
			case fsnotify.Remove, fsnotify.Rename:
				parentDir := filepath.Dir(event.Name)
				if sub, ok := ignoredSubDirs[parentDir]; ok && sub.Has(name) {
					ignoredSubDirs[parentDir].Delete(name)
					continue
				}
			}

			// Share the similar logic as in files watcher.
			var writeEvent bool
			switch event.Op {
			case fsnotify.Create:
				if op, ok := preEventOp[event.Name]; ok &&
					op.Has(fsnotify.Rename) || op.Has(fsnotify.Remove) {
					curRemoveEvents.Delete(event.Name)
				}
				// Since the watcher watches the whole dir, the creation of file
				// should also be able to trigger the Write event.
				writeEvent = true
			case fsnotify.Write:
				writeEvent = true
			case fsnotify.Remove, fsnotify.Rename:
				curRemoveEvents.Insert(event.Name)
			}

			if writeEvent {
				n.logger.Info("sending write event",
					"name", event.Name, "watcher", "dirs")

				n.Events <- fsnotify.Event{
					Name: event.Name,
					Op:   fsnotify.Write,
				}
			}
			preEventOp[event.Name] = event.Op

		case <-cleanUpTicker.C:
			// Merge files to be removed in the same parent directory
			// to suppress events, because the file has already been
			// removed and is unnecessary to send event for each of them.
			parentDirs := sets.NewString()
			for e := range curRemoveEvents {
				parentDirs.Insert(filepath.Dir(e))
			}

			for parentDir := range parentDirs {
				n.logger.Info("sending remove event",
					"name", parentDir, "watcher", "dirs")

				n.Events <- fsnotify.Event{
					Name: parentDir,
					Op:   fsnotify.Remove,
				}
			}
			curRemoveEvents = sets.NewString()
		}
	}
}
