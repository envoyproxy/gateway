package cmd

import (
	"fmt"
	"runtime/debug"
	"strings"

	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/spf13/cobra"
)

var (
	// envOutput determines whether to output as environment settings
	envOutput bool
)

// getVersionsCommand returns the server cobra command to be executed.
func getVersionsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "versions",
		Aliases: []string{"version"},
		Short:   "Show versions",
		RunE: func(cmd *cobra.Command, args []string) error {
			return versions()
		},
	}

	cmd.PersistentFlags().BoolVarP(&envOutput, "env", "e", false,
		"If set, output as environment variable settings.")

	return cmd
}

// versions shows the versions of the Envoy Gateway.
func versions() error {
	envoyVersion := strings.Split(ir.DefaultProxyImage, ":")[1]

	if envOutput {
		fmt.Printf("ENVOY_VERSION=\"%s\"\n", envoyVersion)
	} else {
		fmt.Printf("Envoy:       %s\n", envoyVersion)
	}

	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return fmt.Errorf("could not read build info")
	}

	foundGatewayAPI := false

	for _, dep := range bi.Deps {
		if dep.Path == "sigs.k8s.io/gateway-api" {
			if envOutput {
				fmt.Printf("GATEWAYAPI_VERSION=\"%s\"\n", dep.Version)
			} else {
				fmt.Printf("Gateway API: %s\n", dep.Version)
			}

			foundGatewayAPI = true
			break
		}
	}

	if !foundGatewayAPI {
		return fmt.Errorf("could not find Gateway API version")
	}

	return nil
}
