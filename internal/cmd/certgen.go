// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	clicfg "sigs.k8s.io/controller-runtime/pkg/client/config"

	"github.com/envoyproxy/gateway/internal/crypto"
	"github.com/envoyproxy/gateway/internal/envoygateway"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/provider/kubernetes"
)

// getCertGenCommand returns the certGen cobra command to be executed.
func getCertGenCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "certgen",
		Short: "Generate Control Plane Certificates",
		RunE: func(cmd *cobra.Command, args []string) error {
			return certGen()
		},
	}

	return cmd
}

// certGen generates control plane certificates.
func certGen() error {
	cfg, err := getConfig()
	if err != nil {
		return err
	}
	log := cfg.Logger

	certs, err := crypto.GenerateCerts(cfg)
	if err != nil {
		return fmt.Errorf("failed to generate certificates: %v", err)
	}
	log.Info("generated certificates")

	cli, err := client.New(clicfg.GetConfigOrDie(), client.Options{Scheme: envoygateway.GetScheme()})
	if err != nil {
		return fmt.Errorf("failed to create controller-runtime client: %v", err)
	}

	if err := outputCerts(ctrl.SetupSignalHandler(), cli, cfg, certs); err != nil {
		return fmt.Errorf("failed to output certificates: %v", err)
	}

	return nil
}

// outputCerts outputs the provided certs to a secret in namespace ns.
func outputCerts(ctx context.Context, cli client.Client, cfg *config.Server, certs *crypto.Certificates) error {
	secrets, err := kubernetes.CreateOrUpdateSecrets(ctx, cli, cfg, kubernetes.CertsToSecret(cfg.Namespace, certs))
	log := cfg.Logger

	if err != nil {
		return fmt.Errorf("failed to create or update secrets: %v", err)
	}

	for i := range secrets {
		s := secrets[i]
		log.Info("created secret", "namespace", s.Namespace, "name", s.Name)
	}

	return nil
}
