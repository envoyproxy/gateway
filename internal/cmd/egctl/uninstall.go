// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package egctl

import (
	"github.com/spf13/cobra"

	"github.com/envoyproxy/gateway/internal/cmd/options"
	"github.com/envoyproxy/gateway/internal/utils/helm"
)

func newUnInstallCommand() *cobra.Command {

	packageFlags := &helm.PackageOptions{}
	pt := helm.NewPackageTool()

	uninstallCmd := &cobra.Command{
		Use:   "uninstall",
		Short: "uninstall envoy gateway",
		Long:  uninstallDescription(),
		Example: `  # Uninstall envoy gateway by default, this only offloads envoy gateway instance resources
  egctl x uninstall

  # uninstall all envoy gateway resources, this includes CRDs
  egctl x uninstall --with-crds

  # Uninstall the envoy gateway with the specified release-name
  egctl x uninstall --release-name eg
`,
		PreRunE: func(_ *cobra.Command, _ []string) error {
			return pt.Setup()
		},
		RunE: func(_ *cobra.Command, _ []string) error {
			return pt.RunUninstall(packageFlags)
		},
	}
	options.AddKubeConfigFlags(uninstallCmd.Flags())
	pt.SetUninstallEnvSetting(uninstallCmd, packageFlags)
	pt.SetPreRun(uninstallCmd)

	return uninstallCmd
}

func uninstallDescription() string {
	return `
This command uninstalls envoy gateway.
Since egctl install uses a multi-stage installation, the default uninstall will only uninstall envoy gateway instance resources and not CRDs.
We can specify '--with-crds' to also unload CRDs.
`
}
