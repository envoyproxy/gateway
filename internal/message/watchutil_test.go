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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/message"
)

func TestHandleSubscriptionAlreadyClosed(t *testing.T) {
	ch := make(chan watchable.Snapshot[string, any])
	close(ch)

	var calls int
	message.HandleSubscription[string, any](
		message.Metadata{Runner: "demo", Message: "demo"},
		ch,
		func(update message.Update[string, any], errChans chan error) { calls++ },
	)
	assert.Equal(t, 0, calls)
}

func TestPanicInSubscriptionHandler(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			assert.Fail(t, "recovered from an unexpected panic")
		}
	}()
	var m watchable.Map[string, any]
	m.Store("foo", "bar")

	go func() {
		time.Sleep(100 * time.Millisecond)
		m.Store("baz", "qux")
		time.Sleep(100 * time.Millisecond)
		m.Close()
	}()

	numCalls := 0
	message.HandleSubscription[string, any](
		message.Metadata{Runner: "demo", Message: "demo"},
		m.Subscribe(context.Background()),
		func(update message.Update[string, any], errChans chan error) {
			numCalls++
			panic("oops " + update.Key)
		},
	)
	assert.Equal(t, 2, numCalls)
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
		message.Metadata{Runner: "demo", Message: "demo"},
		m.Subscribe(context.Background()),
		func(update message.Update[string, any], errChans chan error) {
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

func TestControllerResourceUpdate(t *testing.T) {
	tests := []struct {
		desc      string
		resources []*resource.ControllerResources
		updates   int
	}{
		{
			desc: "Resource order change skips update",
			resources: []*resource.ControllerResources{
				{
					{
						GatewayClass: &gwapiv1.GatewayClass{ObjectMeta: metav1.ObjectMeta{Name: "class-1"}},
					},
					{
						GatewayClass: &gwapiv1.GatewayClass{ObjectMeta: metav1.ObjectMeta{Name: "class-2"}},
					},
				},
				{
					{
						GatewayClass: &gwapiv1.GatewayClass{ObjectMeta: metav1.ObjectMeta{Name: "class-2"}},
					},
					{
						GatewayClass: &gwapiv1.GatewayClass{ObjectMeta: metav1.ObjectMeta{Name: "class-1"}},
					},
				},
			},
			updates: 1,
		},
		{
			desc: "Additional resource triggers update",
			resources: []*resource.ControllerResources{
				{
					{
						GatewayClass: &gwapiv1.GatewayClass{ObjectMeta: metav1.ObjectMeta{Name: "class-1"}},
					},
				},
				{
					{
						GatewayClass: &gwapiv1.GatewayClass{ObjectMeta: metav1.ObjectMeta{Name: "class-1"}},
					},
					{
						GatewayClass: &gwapiv1.GatewayClass{ObjectMeta: metav1.ObjectMeta{Name: "class-2"}},
					},
				},
			},
			updates: 2,
		},
		{
			desc: "Multiple Gateways in Resources struct with order change skips update",
			resources: []*resource.ControllerResources{
				{
					{
						GatewayClass: &gwapiv1.GatewayClass{ObjectMeta: metav1.ObjectMeta{Name: "class-1"}},
						Gateways: []*gwapiv1.Gateway{
							{ObjectMeta: metav1.ObjectMeta{Name: "gateway-1", Namespace: "default"}},
							{ObjectMeta: metav1.ObjectMeta{Name: "gateway-2", Namespace: "default"}},
						},
					},
					{
						GatewayClass: &gwapiv1.GatewayClass{ObjectMeta: metav1.ObjectMeta{Name: "class-2"}},
						Gateways: []*gwapiv1.Gateway{
							{ObjectMeta: metav1.ObjectMeta{Name: "gateway-3", Namespace: "system"}},
						},
					},
				},
				{
					{
						GatewayClass: &gwapiv1.GatewayClass{ObjectMeta: metav1.ObjectMeta{Name: "class-2"}},
						Gateways: []*gwapiv1.Gateway{
							{ObjectMeta: metav1.ObjectMeta{Name: "gateway-3", Namespace: "system"}},
						},
					},
					{
						GatewayClass: &gwapiv1.GatewayClass{ObjectMeta: metav1.ObjectMeta{Name: "class-1"}},
						Gateways: []*gwapiv1.Gateway{
							{ObjectMeta: metav1.ObjectMeta{Name: "gateway-2", Namespace: "default"}},
							{ObjectMeta: metav1.ObjectMeta{Name: "gateway-1", Namespace: "default"}},
						},
					},
				},
			},
			updates: 1,
		},
		{
			desc: "Multiple Gateways with Gateway change triggers update",
			resources: []*resource.ControllerResources{
				{
					{
						GatewayClass: &gwapiv1.GatewayClass{ObjectMeta: metav1.ObjectMeta{Name: "class-1"}},
						Gateways: []*gwapiv1.Gateway{
							{ObjectMeta: metav1.ObjectMeta{Name: "gateway-1", Namespace: "default"}},
						},
					},
				},
				{
					{
						GatewayClass: &gwapiv1.GatewayClass{ObjectMeta: metav1.ObjectMeta{Name: "class-1"}},
						Gateways: []*gwapiv1.Gateway{
							{ObjectMeta: metav1.ObjectMeta{Name: "gateway-1", Namespace: "default"}},
							{ObjectMeta: metav1.ObjectMeta{Name: "gateway-2", Namespace: "default"}},
						},
					},
				},
			},
			updates: 2,
		},
		{
			desc: "Multiple Resources with varying Gateway counts",
			resources: []*resource.ControllerResources{
				{
					{
						GatewayClass: &gwapiv1.GatewayClass{ObjectMeta: metav1.ObjectMeta{Name: "class-1"}},
						Gateways: []*gwapiv1.Gateway{
							{ObjectMeta: metav1.ObjectMeta{Name: "gateway-1", Namespace: "default"}},
							{ObjectMeta: metav1.ObjectMeta{Name: "gateway-2", Namespace: "default"}},
							{ObjectMeta: metav1.ObjectMeta{Name: "gateway-3", Namespace: "test"}},
						},
					},
					{
						GatewayClass: &gwapiv1.GatewayClass{ObjectMeta: metav1.ObjectMeta{Name: "class-2"}},
						Gateways: []*gwapiv1.Gateway{
							{ObjectMeta: metav1.ObjectMeta{Name: "gateway-4", Namespace: "system"}},
							{ObjectMeta: metav1.ObjectMeta{Name: "gateway-5", Namespace: "system"}},
						},
					},
				},
				{
					{
						GatewayClass: &gwapiv1.GatewayClass{ObjectMeta: metav1.ObjectMeta{Name: "class-1"}},
						Gateways: []*gwapiv1.Gateway{
							{ObjectMeta: metav1.ObjectMeta{Name: "gateway-1", Namespace: "default"}},
							{ObjectMeta: metav1.ObjectMeta{Name: "gateway-2", Namespace: "default"}},
							{ObjectMeta: metav1.ObjectMeta{Name: "gateway-3", Namespace: "test"}},
							{ObjectMeta: metav1.ObjectMeta{Name: "gateway-6", Namespace: "test"}},
						},
					},
					{
						GatewayClass: &gwapiv1.GatewayClass{ObjectMeta: metav1.ObjectMeta{Name: "class-2"}},
						Gateways: []*gwapiv1.Gateway{
							{ObjectMeta: metav1.ObjectMeta{Name: "gateway-4", Namespace: "system"}},
						},
					},
				},
			},
			updates: 2,
		},
	}
	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			ctx := context.Background()
			m := &message.ProviderResources{}

			snapshotC := m.GatewayAPIResources.Subscribe(ctx)
			endCtx, end := context.WithCancel(ctx)
			m.GatewayAPIResources.Store("start", &resource.ControllerResources{})

			go func() {
				<-endCtx.Done()
				for _, r := range tc.resources {
					r.Sort()
					m.GatewayAPIResources.Store("test", r)
				}
				m.GatewayAPIResources.Store("end", &resource.ControllerResources{})
			}()

			updates := 0
			message.HandleSubscription(message.Metadata{Runner: "demo", Message: "demo"}, snapshotC, func(u message.Update[string, *resource.ControllerResources], errChans chan error) {
				end()
				if u.Key == "test" {
					updates += 1
				}
				if u.Key == "end" {
					m.GatewayAPIResources.Close()
				}
			})
			assert.Equal(t, tc.updates, updates)
		})
	}
}
