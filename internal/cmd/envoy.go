// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package cmd

import (
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
	cmd.AddCommand(getTopologyWebhookCommand())

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

// getTopologyWebhookCommand returns the topology webhook cobra command to be executed.
func getTopologyWebhookCommand() *cobra.Command {
	var (
		certDir                string
		certName               string
		keyName                string
		port                   int
		healthProbeBindAddress string
	)

	cmd := &cobra.Command{
		Use:   "topology-webhook",
		Short: "Provides HTTP endpoint used in MutatingWebhookConfiguration to inject topology information to EnvoyProxy pods.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return envoy.TopologyWebhook(certDir, certName, keyName, healthProbeBindAddress, port)
		},
	}
	cmd.PersistentFlags().StringVar(&certDir, "tls-dir", "/certs",
		"Directory with TLS certificates for the webhook server")
	cmd.PersistentFlags().StringVar(&certName, "tls-crt", "tls.crt",
		"Filename for server TLS certificate")
	cmd.PersistentFlags().StringVar(&keyName, "tls-key", "tls.key",
		"Filename for server TLS key")
	cmd.PersistentFlags().IntVar(&port, "port", 8443,
		"Listener port for webhook server")
	cmd.PersistentFlags().StringVar(&healthProbeBindAddress, "health-probe-bind-address", ":8081",
		"Bind address for health probes")
	return cmd
}
