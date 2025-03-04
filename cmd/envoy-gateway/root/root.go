// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package root

import (
	"github.com/spf13/cobra"

	"github.com/envoyproxy/gateway/internal/cmd"
)

// GetRootCommand returns the root cobra command to be executed
// by main.
func GetRootCommand() *cobra.Command {
	c := &cobra.Command{
		Use:   "envoy-gateway",
		Short: "Envoy Gateway",
		Long:  "Manages Envoy Proxy as a standalone or Kubernetes-based application gateway",
	}

	c.AddCommand(cmd.GetServerCommand())
	c.AddCommand(cmd.GetEnvoyCommand())
	c.AddCommand(cmd.GetVersionCommand())
	c.AddCommand(cmd.GetCertGenCommand())

	return c
}
