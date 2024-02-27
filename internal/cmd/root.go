// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package cmd

import (
	"github.com/spf13/cobra"
)

// GetRootCommand returns the root cobra command to be executed
// by main.
func GetRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "envoy-gateway",
		Short: "Envoy Gateway",
		Long:  "Manages Envoy Proxy as a standalone or Kubernetes-based application gateway",
	}

	cmd.AddCommand(getServerCommand())
	cmd.AddCommand(getEnvoyCommand())
	cmd.AddCommand(getVersionCommand())
	cmd.AddCommand(getCertGenCommand())

	return cmd
}
