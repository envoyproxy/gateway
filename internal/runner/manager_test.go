// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package runner

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/envoyproxy/gateway/api/config/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/logging"
)

type mockRunner struct {
	*GenericRunner[struct{}]
}

func (r *mockRunner) Start(ctx context.Context) error {
	return nil
}

func (r *mockRunner) ShutDown(ctx context.Context) {
	// No-op
}

func (r *mockRunner) SubscribeAndTranslate(ctx context.Context) {
	// No-op
}

func TestManager(t *testing.T) {
	m := Manager()
	conf := config.Server{
		Logger: logging.DefaultLogger(v1alpha1.LogLevelInfo),
	} // Replace with actual config if needed
	m.Init(conf)

	runner1 := &mockRunner{GenericRunner: New("runner1", struct{}{}, conf)}
	runner2 := &mockRunner{GenericRunner: New("runner2", struct{}{}, conf)}

	m.Register(runner1, RootParentRunner)
	m.Register(runner2, "runner1")

	assert.Equal(t, 2, len(m.List()))
	assert.Equal(t, []string{"runner1", "runner2"}, m.ListNames())

	err := m.StartAll(context.Background())
	assert.Nil(t, err)

	err = m.Start(context.Background(), "runner1")
	assert.Nil(t, err)

	err = m.Start(context.Background(), "runner2")
	assert.Nil(t, err)

	m.ShutDown(context.Background(), "runner1")
	m.ShutDownAll(context.Background())
	assert.Equal(t, len(m.List()), 2)

	m.Remove(context.Background(), "runner2")

	assert.NotNil(t, m.Get("runner1"))
	assert.Nil(t, m.Get("runner2"))
	assert.Equal(t, len(m.List()), 1)

	m.RemoveAll(context.Background())
	assert.Nil(t, m.Get("runner1"))
	assert.Equal(t, len(m.List()), 0)
	assert.Nil(t, m.Get("nonexistent"))
}
