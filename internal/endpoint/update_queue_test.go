// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package endpoint

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/envoyproxy/gateway/internal/ir"
)

func TestUpdateQueueKeepsNewestRevision(t *testing.T) {
	var queue UpdateQueue
	backend := ir.EndpointBackend{Kind: "Service", Namespace: "default", Name: "orders", UID: "one"}
	older := &BackendEndpoints{Backend: backend, Revision: 1}
	newer := &BackendEndpoints{Backend: backend, Revision: 2}
	queue.Notify(newer)
	queue.Notify(older)
	require.Len(t, queue.Ready(), 1)
	require.Equal(t, map[Key]*BackendEndpoints{BackendKey(backend): newer}, queue.Drain())
	require.Empty(t, queue.Drain())
}
