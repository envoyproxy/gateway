package message_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/telepresenceio/watchable"

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
