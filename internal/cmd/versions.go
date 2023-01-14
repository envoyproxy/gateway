// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package cmd

import (
	"github.com/spf13/cobra"

	"github.com/envoyproxy/gateway/internal/cmd/version"
)

// getVersionsCommand returns the version cobra command to be executed.
func getVersionsCommand() *cobra.Command {
	var output string

	cmd := &cobra.Command{
		Use:     "versions",
		Aliases: []string{"version"},
		Short:   "Show versions",
		RunE: func(cmd *cobra.Command, args []string) error {
			return version.Print(cmd.OutOrStdout(), output)
		},
	}

	cmd.PersistentFlags().StringVarP(&output, "output", "o", "", "One of 'yaml' or 'json'")

	return cmd
}
