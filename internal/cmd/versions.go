// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package cmd

import (
	"fmt"

	"github.com/envoyproxy/gateway/internal/cmd/version"
	"github.com/spf13/cobra"
)

// getVersionsCommand returns the version cobra command to be executed.
func getVersionsCommand() *cobra.Command {
	// envOutput determines whether to output as environment settings
	var envOutput bool

	cmd := &cobra.Command{
		Use:     "versions",
		Aliases: []string{"version"},
		Short:   "Show versions",
		RunE: func(cmd *cobra.Command, args []string) error {
			return versions(envOutput)
		},
	}

	cmd.PersistentFlags().BoolVarP(&envOutput, "env", "e", false,
		"If set, output as environment variable settings.")

	return cmd
}

// versions shows the versions of the Envoy Gateway.
func versions(envOutput bool) error {
	if envOutput {
		fmt.Printf("ENVOY_VERSION=\"%s\"\n", version.EnvoyVersion)
		fmt.Printf("GATEWAYAPI_VERSION=\"%s\"\n", version.GatewayAPIVersion)
		fmt.Printf("ENVOY_GATEWAY_VERSION=\"%s\"\n", version.EnvoyGatewayVersion)
	} else {
		return version.Print()
	}

	return nil
}
