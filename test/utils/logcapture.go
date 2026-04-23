// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package utils

import (
	"bytes"
	"fmt"
	"io"
	"sync"
	"testing"
)

// OutBuffer combines io.Writer with fmt.Stringer for buffer access
type OutBuffer interface {
	io.Writer
	fmt.Stringer
	Reset()
	Len() int
}

// OutBuffers allows you to reset all the buffers easily.
type OutBuffers []OutBuffer

func (s OutBuffers) Reset() {
	for _, buf := range s {
		buf.Reset()
	}
}

// outBuffer is a thread-safe buffer implementing OutBuffer
type outBuffer struct {
	mu sync.RWMutex
	b  *bytes.Buffer
}

func (s *outBuffer) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.b.Reset()
}

func (s *outBuffer) Write(p []byte) (n int, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.Write(p)
}

func (s *outBuffer) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.b.Len()
}

func (s *outBuffer) String() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.b.String()
}

// CaptureOutput creates labeled OutBuffers in the same order as labels.
func CaptureOutput(labels ...string) OutBuffers {
	buffers := make([]OutBuffer, len(labels))

	for i := range labels {
		buffers[i] = &outBuffer{b: bytes.NewBuffer(nil)}
	}

	return buffers
}

// DumpLogsOnFail creates labeled OutBuffers in the same order as labels.
// The difference between this and CaptureOutput is that when the test fails,
// these are dumped for diagnosis
func DumpLogsOnFail(t testing.TB, labels ...string) OutBuffers {
	buffers := CaptureOutput(labels...)

	t.Cleanup(func() {
		if t.Failed() {
			for i, label := range labels {
				out := buffers[i].String()
				if len(out) == 0 {
					continue
				}
				t.Logf("=== %s ===\n%s", label, out)
			}
		}
	})

	return buffers
}
