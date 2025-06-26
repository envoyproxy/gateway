// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package message

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/telepresenceio/watchable"

	"github.com/envoyproxy/gateway/internal/ir"
)

func TestHandleSubscriptionAlreadyClosed(t *testing.T) {
	ch := make(chan watchable.Snapshot[string, any])
	close(ch)

	var calls int
	HandleSubscription[string, any](
		Metadata{Runner: "demo", Message: "demo"},
		ch,
		func(update Update[string, any], errChans chan error) { calls++ },
	)
	assert.Equal(t, 0, calls)
}

func TestPanicInSubscriptionHandler(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			assert.Fail(t, "recovered from an unexpected panic")
		}
	}()
	m := newPreSubscribedWatchableMap[string, any](context.Background(), 1)
	m.Store("foo", "bar")

	go func() {
		time.Sleep(100 * time.Millisecond)
		m.Store("baz", "qux")
		time.Sleep(100 * time.Millisecond)
		m.Close()
	}()

	numCalls := 0
	HandleSubscription[string, any](
		Metadata{Runner: "demo", Message: "demo"},
		m.GetSubscription(),
		func(update Update[string, any], errChans chan error) {
			numCalls++
			panic("oops " + update.Key)
		},
	)
	assert.Equal(t, 2, numCalls)
}

func TestHandleSubscriptionAlreadyInitialized(t *testing.T) {
	m := newPreSubscribedWatchableMap[string, any](context.Background(), 1)
	m.Store("foo", "bar")

	endCtx, end := context.WithCancel(context.Background())
	go func() {
		<-endCtx.Done()
		m.Store("baz", "qux")
		m.Delete("qux")       // no-op
		m.Store("foo", "bar") // no-op
		m.Delete("baz")
		time.Sleep(100 * time.Millisecond)
		m.Close()
	}()

	var storeCalls int
	var deleteCalls int
	HandleSubscription[string, any](
		Metadata{Runner: "demo", Message: "demo"},
		m.GetSubscription(),
		func(update Update[string, any], errChans chan error) {
			end()
			if update.Delete {
				deleteCalls++
			} else {
				storeCalls++
			}
		},
	)
	assert.Equal(t, 2, storeCalls)
	assert.Equal(t, 1, deleteCalls)
}

func TestHandleStore(t *testing.T) {
	m := newPreSubscribedWatchableMap[string, any](context.Background(), 1)
	HandleStore(Metadata{Runner: "demo", Message: "demo"}, "foo", "bar", m)

	endCtx, end := context.WithCancel(context.Background())
	go func() {
		<-endCtx.Done()
		HandleStore(Metadata{Runner: "demo", Message: "demo"}, "baz", "qux", m)
		m.Delete("qux")                                                         // no-op
		HandleStore(Metadata{Runner: "demo", Message: "demo"}, "foo", "bar", m) // no-op
		m.Delete("baz")
		time.Sleep(100 * time.Millisecond)
		m.Close()
	}()

	var storeCalls int
	var deleteCalls int
	HandleSubscription[string, any](
		Metadata{Runner: "demo", Message: "demo"},
		m.GetSubscription(),
		func(update Update[string, any], errChans chan error) {
			end()
			if update.Delete {
				deleteCalls++
			} else {
				storeCalls++
			}
		},
	)
	assert.Equal(t, 2, storeCalls)
	assert.Equal(t, 1, deleteCalls)
}

func TestXdsIRUpdates(t *testing.T) {
	tests := []struct {
		desc    string
		xx      []*ir.Xds
		updates int
	}{
		{
			desc: "HTTP listener order change skips update",
			xx: []*ir.Xds{
				{
					HTTP: []*ir.HTTPListener{
						{CoreListenerDetails: ir.CoreListenerDetails{Name: "listener-1"}},
						{CoreListenerDetails: ir.CoreListenerDetails{Name: "listener-2"}},
					},
				},
				{
					HTTP: []*ir.HTTPListener{
						{CoreListenerDetails: ir.CoreListenerDetails{Name: "listener-2"}},
						{CoreListenerDetails: ir.CoreListenerDetails{Name: "listener-1"}},
					},
				},
			},
			updates: 1,
		},
		{
			desc: "Additional HTTP listener triggers update",
			xx: []*ir.Xds{
				{
					HTTP: []*ir.HTTPListener{
						{CoreListenerDetails: ir.CoreListenerDetails{Name: "listener-1"}},
					},
				},
				{
					HTTP: []*ir.HTTPListener{
						{CoreListenerDetails: ir.CoreListenerDetails{Name: "listener-1"}},
						{CoreListenerDetails: ir.CoreListenerDetails{Name: "listener-2"}},
					},
				},
			},
			updates: 2,
		},
	}
	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			ctx := context.Background()
			m := NewSubscribedXdsIR(ctx)

			snapshotC := m.GetSubscription()
			endCtx, end := context.WithCancel(ctx)
			m.Store("start", &ir.Xds{})

			go func() {
				<-endCtx.Done()
				for _, x := range tc.xx {
					m.Store("test", x)
				}
				m.Store("end", &ir.Xds{})
			}()

			updates := 0
			HandleSubscription(Metadata{Runner: "demo", Message: "demo"}, snapshotC, func(u Update[string, *ir.Xds], errChans chan error) {
				end()
				if u.Key == "test" {
					updates += 1
				}
				if u.Key == "end" {
					m.Close()
				}
			})
			assert.Equal(t, tc.updates, updates)
		})
	}
}
