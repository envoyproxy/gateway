// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package options

import (
	"github.com/spf13/pflag"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

var DefaultConfigFlags = genericclioptions.NewConfigFlags(true).
	WithDeprecatedPasswordFlag().
	WithDiscoveryBurst(300).
	WithDiscoveryQPS(50.0)

func AddKubeConfigFlags(flags *pflag.FlagSet) {
	flags.StringVar(DefaultConfigFlags.KubeConfig, "kubeconfig", *DefaultConfigFlags.KubeConfig,
		"Path to the kubeconfig file to use for CLI requests.")
	flags.StringVar(DefaultConfigFlags.Context, "context", *DefaultConfigFlags.Context,
		"The name of the kubeconfig context to use.")
}
