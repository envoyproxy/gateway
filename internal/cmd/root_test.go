// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetRootCommand(t *testing.T) {
	got := GetRootCommand()
	assert.Equal(t, "envoy-gateway", got.Use)
}
