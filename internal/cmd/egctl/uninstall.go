// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package egctl

import (
	"github.com/spf13/cobra"

	"github.com/envoyproxy/gateway/internal/cmd/options"
)

func newUnInstallCommand() *cobra.Command {

	htFlags := &HelmOptions{}
	ht := NewHelmTool()

	uninstallCmd := &cobra.Command{
		Use:   "uninstall",
		Short: "uninstall envoy gateway",
		Long:  uninstallDescription(),
		Example: `  # Uninstall envoy gateway by default, this only offloads envoy gateway instance resources
  egctl uninstall

  # uninstall all envoy gateway resources, this includes CRDs
  egctl uninstall --with-crd

  # Uninstall the envoy gateway with the specified release-name
  egctl uninstall --release-name eg
`,
		PreRunE: func(_ *cobra.Command, _ []string) error {
			return ht.setup()
		},
		RunE: func(_ *cobra.Command, _ []string) error {
			return ht.runUninstall(htFlags)
		},
	}
	options.AddKubeConfigFlags(uninstallCmd.Flags())
	ht.setUninstallEnvSetting(uninstallCmd, htFlags)
	ht.setPrinter(uninstallCmd)

	return uninstallCmd
}

func uninstallDescription() string {
	return `
This command uninstalls envoy gateway.
Since egctl install uses a multi-stage installation, the default uninstall will only uninstall envoy gateway instance resources and not CRDs.
We can specify '--with-crd' to also unload CRDs.
`
}
