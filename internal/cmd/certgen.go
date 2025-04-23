// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package cmd

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"path"

	"github.com/spf13/cobra"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
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

var disableTopologyInjector bool

// TODO: make this path configurable or use server config directly.
const (
	defaultLocalCertPath      = "/tmp/envoy-gateway/certs"
	topologyWebhookNamePrefix = "envoy-gateway-topology-injector"
)

// GetCertGenCommand returns the certGen cobra command to be executed.
func GetCertGenCommand() *cobra.Command {
	var local bool

	cmd := &cobra.Command{
		Use:   "certgen",
		Short: "Generate Control Plane Certificates",
		RunE: func(cmd *cobra.Command, args []string) error {
			return certGen(cmd.Context(), cmd.OutOrStdout(), local)
		},
	}

	cmd.PersistentFlags().BoolVarP(&local, "local", "l", false,
		"Generate all the certificates locally.")
	cmd.PersistentFlags().BoolVarP(&overwriteControlPlaneCerts, "overwrite", "o", false,
		"Updates the secrets containing the control plane certs.")
	cmd.PersistentFlags().BoolVar(&disableTopologyInjector, "disable-topology-injector", false,
		"Disables patching caBundle for injector MutatingWebhookConfiguration.")
	return cmd
}

// certGen generates control plane certificates.
func certGen(ctx context.Context, logOut io.Writer, local bool) error {
	cfg, err := config.New(logOut)
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

		if err = outputCertsForKubernetes(ctx, cli, cfg, overwriteControlPlaneCerts, certs); err != nil {
			return fmt.Errorf("failed to output certificates: %w", err)
		}
		if err = patchTopologyWebhook(ctx, cli, cfg, certs.CACertificate); err != nil {
			return fmt.Errorf("failed to patch webhook: %w", err)
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

func patchTopologyWebhook(ctx context.Context, cli client.Client, cfg *config.Server, caBundle []byte) error {
	if disableTopologyInjector {
		return nil
	}

	webhookConfigName := fmt.Sprintf("%s.%s", topologyWebhookNamePrefix, cfg.Namespace)
	webhookCfg := &admissionregistrationv1.MutatingWebhookConfiguration{}
	if err := cli.Get(ctx, client.ObjectKey{Name: webhookConfigName}, webhookCfg); err != nil {
		return fmt.Errorf("failed to get mutating webhook configuration: %w", err)
	}

	var updated bool
	for i, webhook := range webhookCfg.Webhooks {
		if !bytes.Equal(caBundle, webhook.ClientConfig.CABundle) {
			webhookCfg.Webhooks[i].ClientConfig.CABundle = caBundle
			updated = true
		}
	}
	if updated {
		if err := cli.Update(ctx, webhookCfg); err != nil {
			return fmt.Errorf("failed to update mutating webhook configuration: %w", err)
		}
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
