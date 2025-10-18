// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package root

import (
	"github.com/spf13/cobra"

	"github.com/envoyproxy/gateway/internal/cmd"
)

// GetRootCommand returns the root cobra command to be executed by main.
// This command receives an async error handler to let the main process decide how to
// handle critical errors that may happen in the runners that may prevent Envoy Gateway from
// functioning properly.
// The Envoy AI Gateway CLI is an example use case of this function, where it needs to terminate
// if the infra runner fails to start the Envoy process.
func GetRootCommand(asyncErrHandler func(error)) *cobra.Command {
	errChan := make(chan error)
	if asyncErrHandler == nil {
		asyncErrHandler = func(error) {}
	}

	c := &cobra.Command{
		Use:   "envoy-gateway",
		Short: "Envoy Gateway",
		Long:  "Manages Envoy Proxy as a standalone or Kubernetes-based application gateway",
		PersistentPreRun: func(cmd *cobra.Command, _ []string) {
			go func() {
				for {
					select {
					case <-cmd.Context().Done():
						close(errChan)
						return
					case err, ok := <-errChan:
						if ok {
							asyncErrHandler(err)
						}
					}
				}
			}()
		},
	}

	c.AddCommand(cmd.GetServerCommand(errChan))
	c.AddCommand(cmd.GetEnvoyCommand())
	c.AddCommand(cmd.GetVersionCommand())
	c.AddCommand(cmd.GetCertGenCommand())

	return c
}
