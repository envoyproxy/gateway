// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package app

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/envoyproxy/gateway/internal/cmd/version"
)

func NewVersionsCommand() *cobra.Command {
	versionCommand := &cobra.Command{
		Use:     "versions",
		Aliases: []string{"version"},
		Short:   "Show versions",
		RunE: func(cmd *cobra.Command, args []string) error {
			return versions()
		},
	}

	return versionCommand
}

type versionInfo struct {
	ClientVersion string
	// TODO: support display server version
}

func (v *versionInfo) String() string {
	return fmt.Sprintf("CLIENT_VERSION=%s", v.ClientVersion)
}

func Get() versionInfo {
	return versionInfo{
		ClientVersion: version.Get().EnvoyGatewayVersion,
	}
}

func versions() error {
	v := Get()
	fmt.Println(v)

	return nil
}
