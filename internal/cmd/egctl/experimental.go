// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package egctl

import (
	"github.com/spf13/cobra"
)

func NewExperimentalCommand() *cobra.Command {

	experimentalCommand := &cobra.Command{
		Use:     "experimental",
		Aliases: []string{"x"},
		Short:   "Experimental features",
	}

	experimentalCommand.AddCommand(NewTranslateCommand())

	return experimentalCommand
}
