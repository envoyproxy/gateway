// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package message_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/telepresenceio/watchable"

	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/message"
)

func TestHandleSubscriptionAlreadyClosed(t *testing.T) {
	ch := make(chan watchable.Snapshot[string, any])
	close(ch)

	var calls int
	message.HandleSubscription[string, any](
		ch,
		func(message.Update[string, any]) { calls++ },
	)
	assert.Equal(t, 0, calls)
}

func TestHandleSubscriptionAlreadyInitialized(t *testing.T) {
	var m watchable.Map[string, any]
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
	message.HandleSubscription[string, any](
		m.Subscribe(context.Background()),
		func(update message.Update[string, any]) {
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
						{Name: "listener-1"},
						{Name: "listener-2"},
					},
				},
				{
					HTTP: []*ir.HTTPListener{
						{Name: "listener-2"},
						{Name: "listener-1"},
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
						{Name: "listener-1"},
					},
				},
				{
					HTTP: []*ir.HTTPListener{
						{Name: "listener-1"},
						{Name: "listener-2"},
					},
				},
			},
			updates: 2,
		},
	}
	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			ctx := context.Background()
			m := new(message.XdsIR)

			snapshotC := m.Subscribe(ctx)
			go func() {
				for _, x := range tc.xx {
					m.Store("test", x)
				}
				m.Close()
			}()

			updates := 0
			message.HandleSubscription(snapshotC, func(u message.Update[string, *ir.Xds]) {
				updates += 1
			})
			assert.Equal(t, tc.updates, updates)
		})
	}
}
