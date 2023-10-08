// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package egctl

import (
	"github.com/spf13/cobra"
)

func statsCommand() *cobra.Command {
	c := &cobra.Command{
		Use:   "stats",
		Long:  "Retrieve Prometheus Stats from envoy proxy.",
		Short: "Retrieve stats from envoy proxy.",
	}

	c.AddCommand(newEnvoyStatsCmd())

	return c
}
