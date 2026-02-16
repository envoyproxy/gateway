// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package egctl

import (
	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
)

// GetRootCommand returns the root cobra command to be executed
// by egctl main.
func GetRootCommand() *cobra.Command {
	// discard klog logger
	klog.SetLogger(logr.Discard())
	rootCmd := &cobra.Command{
		Use:               "egctl",
		Long:              "A command line utility for operating Envoy Gateway",
		SilenceUsage:      true,
		DisableAutoGenTag: true,
	}

	rootCmd.AddCommand(newVersionCommand())
	rootCmd.AddCommand(newExperimentalCommand())
	rootCmd.AddCommand(newConfigCommand())

	return rootCmd
}
