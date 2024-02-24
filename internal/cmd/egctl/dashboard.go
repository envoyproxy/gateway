// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package egctl

import (
	"github.com/spf13/cobra"
)

func newDashboardCommand() *cobra.Command {
	c := &cobra.Command{
		Use:     "dashboard",
		Aliases: []string{"d"},
		Long:    "Retrieve the dashboard.",
		Short:   "Retrieve the dashboard.",
	}

	c.AddCommand(newEnvoyDashboardCmd())

	return c
}
