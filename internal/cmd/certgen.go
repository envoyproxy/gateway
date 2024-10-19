// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/spf13/cobra"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	clicfg "sigs.k8s.io/controller-runtime/pkg/client/config"

	"github.com/envoyproxy/gateway/internal/crypto"
	"github.com/envoyproxy/gateway/internal/envoygateway"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/provider/kubernetes"
)

// TODO: make this path configurable.
const defaultLocalCertPath = "/tmp/envoy-gateway/certs"

// getCertGenCommand returns the certGen cobra command to be executed.
func getCertGenCommand() *cobra.Command {
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

	return cmd
}

// certGen generates control plane certificates.
func certGen(local bool) error {
	cfg, err := getConfig()
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

		if err = outputCertsForKubernetes(ctrl.SetupSignalHandler(), cli, cfg, certs); err != nil {
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
func outputCertsForKubernetes(ctx context.Context, cli client.Client, cfg *config.Server, certs *crypto.Certificates) error {
	var updateSecrets bool
	if cfg.EnvoyGateway != nil &&
		cfg.EnvoyGateway.Provider != nil &&
		cfg.EnvoyGateway.Provider.Kubernetes != nil &&
		cfg.EnvoyGateway.Provider.Kubernetes.OverwriteControlPlaneCerts != nil &&
		*cfg.EnvoyGateway.Provider.Kubernetes.OverwriteControlPlaneCerts {
		updateSecrets = true
	}
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

func outputCertsForLocal(localPath string, certs *crypto.Certificates) (errs error) {
	writeCerts := func(dir, filename string, cert []byte) error {
		err := os.MkdirAll(dir, 0o750)
		if err != nil {
			return err
		}

		var f *os.File
		f, err = os.Create(path.Join(dir, filename))
		if err != nil {
			return err
		}
		defer f.Close()

		_, err = f.Write(cert)
		return err
	}

	errs = errors.Join(writeCerts(path.Join(localPath, "envoy-gateway"), "ca.crt", certs.CACertificate))
	errs = errors.Join(writeCerts(path.Join(localPath, "envoy-gateway"), "tls.crt", certs.EnvoyGatewayCertificate))
	errs = errors.Join(writeCerts(path.Join(localPath, "envoy-gateway"), "tls.key", certs.EnvoyGatewayPrivateKey))

	errs = errors.Join(writeCerts(path.Join(localPath, "envoy"), "ca.crt", certs.CACertificate))
	errs = errors.Join(writeCerts(path.Join(localPath, "envoy"), "tls.crt", certs.EnvoyCertificate))
	errs = errors.Join(writeCerts(path.Join(localPath, "envoy"), "tls.key", certs.EnvoyPrivateKey))

	errs = errors.Join(writeCerts(path.Join(localPath, "envoy-rate-limit"), "ca.crt", certs.CACertificate))
	errs = errors.Join(writeCerts(path.Join(localPath, "envoy-rate-limit"), "tls.crt", certs.EnvoyRateLimitCertificate))
	errs = errors.Join(writeCerts(path.Join(localPath, "envoy-rate-limit"), "tls.key", certs.EnvoyRateLimitPrivateKey))

	errs = errors.Join(writeCerts(path.Join(localPath, "envoy-oidc-hmac"), "hmac-secret", certs.OIDCHMACSecret))

	return
}
