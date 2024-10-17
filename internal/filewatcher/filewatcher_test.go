// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package filewatcher

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"runtime"
	"sync"
	"testing"

	"github.com/fsnotify/fsnotify"
	"github.com/stretchr/testify/require"
)

func newWatchFile(t *testing.T) string {
	watchDir := t.TempDir()
	watchFile := path.Join(watchDir, "test.conf")
	err := os.WriteFile(watchFile, []byte("foo: bar\n"), 0o600)
	require.NoError(t, err)

	return watchFile
}

func newWatchFileThatDoesNotExist(t *testing.T) string {
	watchDir := t.TempDir()

	watchFile := path.Join(watchDir, "test.conf")

	return watchFile
}

// newTwoWatchFile returns with two watch files that exist in the same base dir.
func newTwoWatchFile(t *testing.T) (string, string) {
	watchDir := t.TempDir()

	watchFile1 := path.Join(watchDir, "test1.conf")
	err := os.WriteFile(watchFile1, []byte("foo: bar\n"), 0o600)
	require.NoError(t, err)

	watchFile2 := path.Join(watchDir, "test2.conf")
	err = os.WriteFile(watchFile2, []byte("foo: baz\n"), 0o600)
	require.NoError(t, err)

	return watchFile1, watchFile2
}

// newSymlinkedWatchFile simulates the behavior of k8s configmap/secret.
// Path structure looks like:
//
//	<watchDir>/test.conf
//	             ^
//	             |
//
// <watchDir>/data/test.conf
//
//	^
//	|
//
// <watchDir>/data1/test.conf
func newSymlinkedWatchFile(t *testing.T) (string, string) {
	watchDir := t.TempDir()

	dataDir1 := path.Join(watchDir, "data1")
	err := os.Mkdir(dataDir1, 0o777)
	require.NoError(t, err)

	realTestFile := path.Join(dataDir1, "test.conf")
	t.Logf("Real test file location: %s\n", realTestFile)
	err = os.WriteFile(realTestFile, []byte("foo: bar\n"), 0o600)
	require.NoError(t, err)

	// Now, symlink the tmp `data1` dir to `data` in the baseDir
	require.NoError(t, os.Symlink(dataDir1, path.Join(watchDir, "data")))
	// And link the `<watchdir>/datadir/test.conf` to `<watchdir>/test.conf`
	watchFile := path.Join(watchDir, "test.conf")
	require.NoError(t, os.Symlink(path.Join(watchDir, "data", "test.conf"), watchFile))
	t.Logf("Watch file location: %s\n", path.Join(watchDir, "test.conf"))
	return watchDir, watchFile
}

func TestWatchFile(t *testing.T) {
	t.Run("file content changed", func(t *testing.T) {
		// Given a file being watched
		watchFile := newWatchFile(t)
		_, err := os.Stat(watchFile)
		require.NoError(t, err)

		w := NewWatcher()
		require.NoError(t, w.Add(watchFile))
		events := w.Events(watchFile)

		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			<-events
			wg.Done()
		}()

		// Overwriting the file and waiting its event to be received.
		err = os.WriteFile(watchFile, []byte("foo: baz\n"), 0o600)
		require.NoError(t, err)
		wg.Wait()

		_ = w.Close()
	})

	t.Run("link to real file changed (for k8s configmap/secret path)", func(t *testing.T) {
		// skip if not executed on Linux
		if runtime.GOOS != "linux" {
			t.Skip("Skipping test as symlink replacements don't work on non-linux environment...")
		}

		watchDir, watchFile := newSymlinkedWatchFile(t)

		w := NewWatcher()
		require.NoError(t, w.Add(watchFile))
		events := w.Events(watchFile)

		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			<-events
			wg.Done()
		}()

		// Link to another `test.conf` file
		dataDir2 := path.Join(watchDir, "data2")
		err := os.Mkdir(dataDir2, 0o777)
		require.NoError(t, err)

		watchFile2 := path.Join(dataDir2, "test.conf")
		err = os.WriteFile(watchFile2, []byte("foo: baz\n"), 0o600)
		require.NoError(t, err)

		// change the symlink using the `ln -sfn` command
		err = exec.Command("ln", "-sfn", dataDir2, path.Join(watchDir, "data")).Run()
		require.NoError(t, err)

		// Wait its event to be received.
		wg.Wait()

		_ = w.Close()
	})

	t.Run("file added later", func(t *testing.T) {
		// Given a file being watched
		watchFile := newWatchFileThatDoesNotExist(t)

		w := NewWatcher()
		require.NoError(t, w.Add(watchFile))
		events := w.Events(watchFile)

		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			<-events
			wg.Done()
		}()

		// Overwriting the file and waiting its event to be received.
		err := os.WriteFile(watchFile, []byte("foo: baz\n"), 0o600)
		require.NoError(t, err)
		wg.Wait()

		_ = w.Close()
	})
}

func TestWatcherLifecycle(t *testing.T) {
	watchFile1, watchFile2 := newTwoWatchFile(t)

	w := NewWatcher()

	// Validate Add behavior
	err := w.Add(watchFile1)
	require.NoError(t, err)
	err = w.Add(watchFile2)
	require.NoError(t, err)

	// Validate events and errors channel are fulfilled.
	events1 := w.Events(watchFile1)
	require.NotNil(t, events1)
	events2 := w.Events(watchFile2)
	require.NotNil(t, events2)

	errors1 := w.Errors(watchFile1)
	require.NotNil(t, errors1)
	errors2 := w.Errors(watchFile2)
	require.NotNil(t, errors2)

	// Validate Remove behavior
	err = w.Remove(watchFile1)
	require.NoError(t, err)
	events1 = w.Events(watchFile1)
	require.Nil(t, events1)
	errors1 = w.Errors(watchFile1)
	require.Nil(t, errors1)
	events2 = w.Events(watchFile2)
	require.NotNil(t, events2)
	errors2 = w.Errors(watchFile2)
	require.NotNil(t, errors2)

	fmt.Printf("2\n")
	// Validate Close behavior
	err = w.Close()
	require.NoError(t, err)
	events1 = w.Events(watchFile1)
	require.Nil(t, events1)
	errors1 = w.Errors(watchFile1)
	require.Nil(t, errors1)
	events2 = w.Events(watchFile2)
	require.Nil(t, events2)
	errors2 = w.Errors(watchFile2)
	require.Nil(t, errors2)
}

func TestErrors(t *testing.T) {
	w := NewWatcher()

	if ch := w.Errors("XYZ"); ch != nil {
		t.Error("Expected no channel")
	}

	if ch := w.Events("XYZ"); ch != nil {
		t.Error("Expected no channel")
	}

	name := newWatchFile(t)
	_ = w.Add(name)
	_ = w.Remove(name)

	if ch := w.Errors("XYZ"); ch != nil {
		t.Error("Expected no channel")
	}

	if ch := w.Events(name); ch != nil {
		t.Error("Expected no channel")
	}

	_ = w.Close()

	if err := w.Add(name); err == nil {
		t.Error("Expecting error")
	}

	if err := w.Remove(name); err == nil {
		t.Error("Expecting error")
	}

	if ch := w.Errors(name); ch != nil {
		t.Error("Expecting nil")
	}

	if ch := w.Events(name); ch != nil {
		t.Error("Expecting nil")
	}
}

func TestBadWatcher(t *testing.T) {
	w := NewWatcher()
	w.(*fileWatcher).funcs.newWatcher = func() (*fsnotify.Watcher, error) {
		return nil, errors.New("FOOBAR")
	}

	name := newWatchFile(t)
	if err := w.Add(name); err == nil {
		t.Errorf("Expecting error, got nil")
	}
	if err := w.Close(); err != nil {
		t.Errorf("Expecting nil, got %v", err)
	}
}

func TestBadAddWatcher(t *testing.T) {
	w := NewWatcher()
	w.(*fileWatcher).funcs.addWatcherPath = func(*fsnotify.Watcher, string) error {
		return errors.New("FOOBAR")
	}

	name := newWatchFile(t)
	if err := w.Add(name); err == nil {
		t.Errorf("Expecting error, got nil")
	}
	if err := w.Close(); err != nil {
		t.Errorf("Expecting nil, got %v", err)
	}
}

func TestDuplicateAdd(t *testing.T) {
	w := NewWatcher()

	name := newWatchFile(t)

	if err := w.Add(name); err != nil {
		t.Errorf("Expecting nil, got %v", err)
	}

	if err := w.Add(name); err == nil {
		t.Errorf("Expecting error, got nil")
	}

	_ = w.Close()
}

func TestBogusRemove(t *testing.T) {
	w := NewWatcher()

	name := newWatchFile(t)
	if err := w.Remove(name); err == nil {
		t.Errorf("Expecting error, got nil")
	}

	_ = w.Close()
}
