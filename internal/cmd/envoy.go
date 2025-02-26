// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/envoyproxy/gateway/internal/cmd/envoy"
)

// GetEnvoyCommand returns the envoy cobra command to be executed.
func GetEnvoyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "envoy",
		Short: "Envoy proxy management",
	}

	cmd.AddCommand(getShutdownCommand())
	cmd.AddCommand(getShutdownManagerCommand())
	cmd.AddCommand(getEnvoyInitCommand())

	return cmd
}

// getShutdownCommand returns the shutdown cobra command to be executed.
func getShutdownCommand() *cobra.Command {
	var drainTimeout time.Duration
	var minDrainDuration time.Duration
	var exitAtConnections int

	cmd := &cobra.Command{
		Use:   "shutdown",
		Short: "Gracefully drain open connections prior to pod shutdown.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return envoy.Shutdown(drainTimeout, minDrainDuration, exitAtConnections)
		},
	}

	cmd.PersistentFlags().DurationVar(&drainTimeout, "drain-timeout", 60*time.Second,
		"Graceful shutdown timeout. This should be less than the pod's terminationGracePeriodSeconds.")

	cmd.PersistentFlags().DurationVar(&minDrainDuration, "min-drain-duration", 10*time.Second,
		"Minimum drain duration allowing time for endpoint deprogramming to complete.")

	cmd.PersistentFlags().IntVar(&exitAtConnections, "exit-at-connections", 0,
		"Number of connections to wait for when monitoring Envoy listener drain process.")

	return cmd
}

// getShutdownManagerCommand returns the shutdown manager cobra command to be executed.
func getShutdownManagerCommand() *cobra.Command {
	var readyTimeout time.Duration

	cmd := &cobra.Command{
		Use:   "shutdown-manager",
		Short: "Provides HTTP endpoint used in preStop hook to block until ready for pod shutdown.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return envoy.ShutdownManager(readyTimeout)
		},
	}

	cmd.PersistentFlags().DurationVar(&readyTimeout, "ready-timeout", 610*time.Second,
		"Shutdown ready timeout. This should be greater than shutdown's drain-timeout and less than the pod's terminationGracePeriodSeconds.")

	return cmd
}

// getEnvoyInitCommand returns the envoy init cobra command to be executed.
func getEnvoyInitCommand() *cobra.Command {
	var configPath string
	var discoverRegion bool
	var discoverZone bool
	var overrideRegion string
	var overrideZone string

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Provides HTTP endpoint used in preStop hook to block until ready for pod shutdown.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return envoy.EnvoyInit(configPath, discoverRegion, overrideRegion, discoverZone, overrideZone)
		},
	}
	cmd.PersistentFlags().StringVar(&configPath, "config-path", envoy.DefaultEnvoyInitConfigPath,
		fmt.Sprintf("File path to write discovered zone information. Defaults to %s", envoy.DefaultEnvoyInitConfigPath))
	cmd.PersistentFlags().BoolVar(&discoverRegion, "disable-region-discovery", false,
		"Whether to enable service region discovery via topology.kubernetes.io/region label on the underlying node.")
	cmd.PersistentFlags().StringVar(&overrideRegion, "override-region", "",
		"Override discovered region with custom string")
	cmd.PersistentFlags().BoolVar(&discoverZone, "disable-zone-discovery", false,
		"Whether to enable service zone discovery via topology.kubernetes.io/zone label on the underlying node.")
	cmd.PersistentFlags().StringVar(&overrideZone, "override-zone", "",
		"Override discovered region with custom string")

	return cmd
}
