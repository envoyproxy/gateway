// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package egctl

import (
	"bytes"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
)

func newValidateCommand() *cobra.Command {
	var inFile string

	validateCommand := &cobra.Command{
		Use:   "validate",
		Short: "Validate Gateway API Resources from the given file, return all the errors if got any.",
		Example: `  # Validate Gateway API Resources
  egctl x validate -f <input file>
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(inFile) == 0 {
				return fmt.Errorf("-f/--file must be specified")
			}

			return runValidate(cmd.OutOrStdout(), inFile)
		},
	}

	validateCommand.PersistentFlags().StringVarP(&inFile, "file", "f", "", "Location of input file.")
	if err := validateCommand.MarkPersistentFlagRequired("file"); err != nil {
		return nil
	}

	return validateCommand
}

func runValidate(w io.Writer, inFile string) error {
	inBytes, err := getInputBytes(inFile)
	if err != nil {
		return fmt.Errorf("unable to read input file: %w", err)
	}

	noErr := true
	_ = resource.IterYAMLBytes(inBytes, func(yamlByte []byte) error {
		// Passing each resource as YAML string and get all their errors from local validator.
		_, err = resource.LoadResourcesFromYAMLBytes(yamlByte, false)
		if err != nil {
			noErr = false
			yamlRows := bytes.Split(yamlByte, []byte("\n"))
			if len(yamlRows) > 6 {
				yamlRows = append(yamlRows[:6], []byte("..."))
			}
			_, err = fmt.Fprintf(w, "%s\n%s\n\n",
				bytes.Join(yamlRows, []byte("\n")), err.Error())
		}
		return nil
	})

	if noErr {
		_, err = fmt.Fprintln(w, "\033[32mOK\033[0m")
	}

	return err
}
