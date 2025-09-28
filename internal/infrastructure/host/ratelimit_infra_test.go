// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package host

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeleteRateLimitInfra(t *testing.T) {
	// Create a host infrastructure instance
	infra := &Infra{}

	// Test that DeleteRateLimitInfra returns nil (no error) in standalone mode
	err := infra.DeleteRateLimitInfra(context.Background())
	assert.NoError(t, err, "DeleteRateLimitInfra should return nil in standalone mode")
}

func TestCreateOrUpdateRateLimitInfra(t *testing.T) {
	// Create a host infrastructure instance
	infra := &Infra{}

	// Test that CreateOrUpdateRateLimitInfra returns an error (not implemented yet)
	err := infra.CreateOrUpdateRateLimitInfra(context.Background())
	assert.Error(t, err, "CreateOrUpdateRateLimitInfra should return an error as it's not implemented yet")
	assert.Contains(t, err.Error(), "create/update ratelimit infrastructure is not supported yet for host infrastructure")
}
