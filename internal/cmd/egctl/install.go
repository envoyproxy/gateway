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

func newInstallCommand() *cobra.Command {

	packageFlags := &helm.PackageOptions{}
	pt := helm.NewPackageTool()

	installCmd := &cobra.Command{
		Use:   "install",
		Short: "install envoy gateway",
		Long:  installDescription(),
		Example: `  # Installed by default, this will install CRDs resources as well as envoy gateway instance resources
  egctl x install

  # Install envoy gateway instance resources only, skip installing CRDs
  egctl x install --skip-crds

  # Specify the envoy gateway version and enable debug logging
  egctl x install --version v0.6.0 --debug

  # Override the default values of the envoy gateway chart and install CRDs only
  egctl x install --set config.envoyGateway.logging.level.default=info --only-crds
`,
		PreRunE: func(_ *cobra.Command, _ []string) error {
			return pt.Setup()
		},
		RunE: func(_ *cobra.Command, _ []string) error {
			return pt.RunInstall(packageFlags)
		},
	}
	options.AddKubeConfigFlags(installCmd.Flags())
	pt.SetInstallEnvSettings(installCmd, packageFlags)
	pt.SetPreRun(installCmd)

	return installCmd
}

func installDescription() string {
	return `
This command installs envoy gateway, which uses the Helm library.
Will use oci://docker.io/envoyproxy/gateway-helm repo chart, default to the latest version, it is recommended to use a fixed.
The installation process uses a two-stage installation step: install CRDs first, and then install instance resources, which always ensures that CRDs resources are installed correctly.
`
}
