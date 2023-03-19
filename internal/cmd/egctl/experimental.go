// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package egctl

import (
	"github.com/spf13/cobra"
)

func newExperimentalCommand() *cobra.Command {
	experimentalCommand := &cobra.Command{
		Use:     "experimental",
		Aliases: []string{"x"},
		Short:   "Experimental features",
		Example: `  # Use experimental features of egctl.
  egctl experimental [command]

  # Use experimental features of egctl with short syntax.
  egctl x [command]
	  `,
	}

	experimentalCommand.AddCommand(NewTranslateCommand())

	return experimentalCommand
}
