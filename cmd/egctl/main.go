// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/envoyproxy/gateway/internal/cmd/egctl"
)

func main() {
	if err := rootCommand().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func rootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:               "egctl",
		Long:              "A command line utility for operating Envoy Gateway",
		SilenceUsage:      true,
		DisableAutoGenTag: true,
	}

	rootCmd.AddCommand(egctl.NewVersionCommand())
	rootCmd.AddCommand(egctl.NewExperimentalCommand())
	rootCmd.AddCommand(egctl.NewConfigCommand())

	return rootCmd
}
