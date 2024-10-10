// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package runner

import (
	"context"

	"github.com/envoyproxy/gateway/internal/envoygateway/config"
)

type Runner interface {
	Name() string

	Start(ctx context.Context) error
	// Reload will be called when the config file changes,
	// runner should take care of reloading the config.
	Reload(serverCfg *config.Server) error
}
