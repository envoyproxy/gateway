// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package cmd

import (
	"github.com/spf13/cobra"

	"github.com/envoyproxy/gateway/internal/cmd/version"
)

// getVersionCommand returns the version cobra command to be executed.
func getVersionCommand() *cobra.Command {
	var output string

	cmd := &cobra.Command{
		Use:     "version",
		Aliases: []string{"versions", "v"},
		Short:   "Show versions",
		RunE: func(cmd *cobra.Command, args []string) error {
			return version.Print(cmd.OutOrStdout(), output)
		},
	}

	cmd.PersistentFlags().StringVarP(&output, "output", "o", "", "One of 'yaml' or 'json'")

	return cmd
}
