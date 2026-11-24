// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/envoyproxy/gateway/internal/cmd/egctl"
	"github.com/envoyproxy/gateway/internal/cmd/egctl/migrate"
)

func main() {
	if err := egctl.GetRootCommand().Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		// The migrate command exits with 1 or 2, see migrate.ExitError.
		if exitErr, ok := errors.AsType[*migrate.ExitError](err); ok {
			os.Exit(exitErr.Code)
		}
		os.Exit(1)
	}
}
