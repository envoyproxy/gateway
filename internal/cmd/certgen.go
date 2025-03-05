// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package cmd

import (
	"context"
	"errors"
	"fmt"
	"path"

	"github.com/spf13/cobra"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	clicfg "sigs.k8s.io/controller-runtime/pkg/client/config"

	"github.com/envoyproxy/gateway/internal/crypto"
	"github.com/envoyproxy/gateway/internal/envoygateway"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/provider/kubernetes"
	"github.com/envoyproxy/gateway/internal/utils/file"
)

// cfgPath is the path to the EnvoyGateway configuration file.
var overwriteControlPlaneCerts bool

// TODO: make this path configurable or use server config directly.
const defaultLocalCertPath = "/tmp/envoy-gateway/certs"

// GetCertGenCommand returns the certGen cobra command to be executed.
func GetCertGenCommand() *cobra.Command {
	var local bool

	cmd := &cobra.Command{
		Use:   "certgen",
		Short: "Generate Control Plane Certificates",
		RunE: func(cmd *cobra.Command, args []string) error {
			return certGen(local)
		},
	}

	cmd.PersistentFlags().BoolVarP(&local, "local", "l", false,
		"Generate all the certificates locally.")
	cmd.PersistentFlags().BoolVarP(&overwriteControlPlaneCerts, "overwrite", "o", false,
		"Updates the secrets containing the control plane certs.")
	return cmd
}

// certGen generates control plane certificates.
func certGen(local bool) error {
	cfg, err := config.New()
	if err != nil {
		return err
	}
	log := cfg.Logger

	certs, err := crypto.GenerateCerts(cfg)
	if err != nil {
		return fmt.Errorf("failed to generate certificates: %w", err)
	}

	if !local {
		log.Info("generated certificates")
		cli, err := client.New(clicfg.GetConfigOrDie(), client.Options{Scheme: envoygateway.GetScheme()})
		if err != nil {
			return fmt.Errorf("failed to create controller-runtime client: %w", err)
		}

		if err = outputCertsForKubernetes(ctrl.SetupSignalHandler(), cli, cfg, overwriteControlPlaneCerts, certs); err != nil {
			return fmt.Errorf("failed to output certificates: %w", err)
		}
	} else {
		log.Info("generated certificates", "path", defaultLocalCertPath)
		if err = outputCertsForLocal(defaultLocalCertPath, certs); err != nil {
			return fmt.Errorf("failed to output certificates locally: %w", err)
		}
	}

	return nil
}

// outputCertsForKubernetes outputs the provided certs to a secret in namespace ns.
func outputCertsForKubernetes(ctx context.Context, cli client.Client, cfg *config.Server,
	updateSecrets bool, certs *crypto.Certificates,
) error {
	secrets, err := kubernetes.CreateOrUpdateSecrets(ctx, cli, kubernetes.CertsToSecret(cfg.Namespace, certs), updateSecrets)
	log := cfg.Logger

	if err != nil {
		if errors.Is(err, kubernetes.ErrSecretExists) {
			log.Info(err.Error())
		} else {
			return fmt.Errorf("failed to create or update secrets: %w", err)
		}
	}

	for i := range secrets {
		s := secrets[i]
		log.Info("created secret", "namespace", s.Namespace, "name", s.Name)
	}

	return nil
}

// outputCertsForLocal outputs the provided certs to the local directory as files.
func outputCertsForLocal(localPath string, certs *crypto.Certificates) (err error) {
	egDir := path.Join(localPath, "envoy-gateway")
	if err = file.WriteDir(certs.CACertificate, egDir, "ca.crt"); err != nil {
		return err
	}
	if err = file.WriteDir(certs.EnvoyGatewayCertificate, egDir, "tls.crt"); err != nil {
		return err
	}
	if err = file.WriteDir(certs.EnvoyGatewayPrivateKey, egDir, "tls.key"); err != nil {
		return err
	}

	envoyDir := path.Join(localPath, "envoy")
	if err = file.WriteDir(certs.CACertificate, envoyDir, "ca.crt"); err != nil {
		return err
	}
	if err = file.WriteDir(certs.EnvoyCertificate, envoyDir, "tls.crt"); err != nil {
		return err
	}
	if err = file.WriteDir(certs.EnvoyPrivateKey, envoyDir, "tls.key"); err != nil {
		return err
	}

	rlDir := path.Join(localPath, "envoy-rate-limit")
	if err = file.WriteDir(certs.CACertificate, rlDir, "ca.crt"); err != nil {
		return err
	}
	if err = file.WriteDir(certs.EnvoyRateLimitCertificate, rlDir, "tls.crt"); err != nil {
		return err
	}
	if err = file.WriteDir(certs.EnvoyRateLimitPrivateKey, rlDir, "tls.key"); err != nil {
		return err
	}

	if err = file.WriteDir(certs.OIDCHMACSecret, path.Join(localPath, "envoy-oidc-hmac"), "hmac-secret"); err != nil {
		return err
	}

	return
}
