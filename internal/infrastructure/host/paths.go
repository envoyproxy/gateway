// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package host

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

// Paths contains all XDG-style directory paths for host infrastructure.
type Paths struct {
	ConfigHome string
	DataHome   string
	StateHome  string
	RuntimeDir string
}

// GetPaths returns directory paths from config or XDG defaults.
// This follows the same pattern as func-e's ConfigHome(), DataHome(), StateHome(), RuntimeDir()
// but uses configuration fields instead of RunOptions.
func GetPaths(cfg *egv1a1.EnvoyGatewayHostInfrastructureProvider) (*Paths, error) {
	u, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("failed to get current user: %w", err)
	}

	paths := &Paths{}

	// ConfigHome
	if cfg != nil && cfg.ConfigHome != nil {
		paths.ConfigHome = *cfg.ConfigHome
	} else {
		paths.ConfigHome = filepath.Join(u.HomeDir, ".config", "envoy-gateway")
	}

	// DataHome (Envoy binaries, shared application data)
	if cfg != nil && cfg.DataHome != nil {
		paths.DataHome = *cfg.DataHome
	} else {
		paths.DataHome = filepath.Join(u.HomeDir, ".local", "share", "envoy-gateway")
	}

	// StateHome (logs, persistent state)
	if cfg != nil && cfg.StateHome != nil {
		paths.StateHome = *cfg.StateHome
	} else {
		paths.StateHome = filepath.Join(u.HomeDir, ".local", "state", "envoy-gateway")
	}

	// RuntimeDir (ephemeral files)
	if cfg != nil && cfg.RuntimeDir != nil {
		paths.RuntimeDir = *cfg.RuntimeDir
	} else {
		// Use UID for multi-user safety, like func-e does
		paths.RuntimeDir = filepath.Join(os.TempDir(), fmt.Sprintf("envoy-gateway-%s", u.Uid))
	}

	return paths, nil
}

// CertDir returns the certificate directory path (under ConfigHome).
func (p *Paths) CertDir(component string) string {
	return filepath.Join(p.ConfigHome, "certs", component)
}
