// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package app

import (
	"testing"
)

func TestNewVersionsCommand(t *testing.T) {
	c := NewVersionsCommand()
	if err := c.Execute(); err != nil {
		t.Errorf("Cannot execute version command: %v", err)
	}
}
