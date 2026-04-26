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
	"time"

	"github.com/spf13/cobra"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	clicfg "sigs.k8s.io/controller-runtime/pkg/client/config"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/crypto"
	"github.com/envoyproxy/gateway/internal/envoygateway"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/infrastructure/host"
	"github.com/envoyproxy/gateway/internal/infrastructure/kubernetes/ratelimit"
	"github.com/envoyproxy/gateway/internal/provider/kubernetes"
	"github.com/envoyproxy/gateway/internal/utils/file"
)

// cfgPath is the path to the EnvoyGateway configuration file.
var overwriteControlPlaneCerts bool

var disableTopologyInjector bool

const (
	topologyWebhookNamePrefix = "envoy-gateway-topology-injector"
)

// GetCertGenCommand returns the certGen cobra command to be executed.
func GetCertGenCommand() *cobra.Command {
	var (
		local      bool
		configHome string
	)

	cmd := &cobra.Command{
		Use:   "certgen",
		Short: "Generate Control Plane Certificates",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return certGen(cmd.Context(), cmd.OutOrStdout(), local, configHome)
		},
	}

	cmd.PersistentFlags().BoolVarP(&local, "local", "l", false,
		"Generate all the certificates locally.")
	cmd.PersistentFlags().StringVar(&configHome, "config-home", "",
		"Directory for certificates (defaults to ~/.config/envoy-gateway")
	cmd.PersistentFlags().BoolVarP(&overwriteControlPlaneCerts, "overwrite", "o", false,
		"Updates the secrets containing the control plane certs.")
	cmd.PersistentFlags().BoolVar(&disableTopologyInjector, "disable-topology-injector", false,
		"Disables patching caBundle for injector MutatingWebhookConfiguration.")
	return cmd
}

// certGen generates control plane certificates.
func certGen(ctx context.Context, logOut io.Writer, local bool, configHome string) error {
	cfg, err := config.New(logOut, io.Discard)
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
		if err = patchTopologyInjectorWebhook(ctx, cli, cfg); err != nil {
			return fmt.Errorf("failed to patch webhook: %w", err)
		}
		if overwriteControlPlaneCerts {
			if err = rolloutRestartRateLimitDeployment(ctx, cli, cfg); err != nil {
				return fmt.Errorf("failed to restart rate limit deployment: %w", err)
			}
		}
	} else {
		// Use provided configHome or default
		hostCfg := &egv1a1.EnvoyGatewayHostInfrastructureProvider{}
		if configHome != "" {
			hostCfg.ConfigHome = &configHome
		}

		paths, err := host.GetPaths(hostCfg)
		if err != nil {
			return fmt.Errorf("failed to determine paths: %w", err)
		}

		certPath := paths.CertDir("")
		log.Info("generated certificates", "path", certPath)

		if err = outputCertsForLocal(certPath, certs); err != nil {
			return fmt.Errorf("failed to output certificates locally: %w", err)
		}
	}

	return nil
}

// outputCertsForKubernetes outputs the provided certs to a secret in namespace ns.
func outputCertsForKubernetes(ctx context.Context, cli client.Client, cfg *config.Server,
	updateSecrets bool, certs *crypto.Certificates,
) error {
	secrets, err := kubernetes.CreateOrUpdateSecrets(ctx, cli, kubernetes.CertsToSecret(cfg.ControllerNamespace, certs), updateSecrets)
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

// rolloutRestartRateLimitDeployment triggers a rolling restart of the Rate Limit
// deployment by patching its pod-template annotation. This is necessary because
// the Rate Limit process loads its CA certificate once at startup and does not
// watch the mounted secret for changes. After a cert rotation the kubelet will
// update the secret volume on disk, but the running process continues to verify
// client certs against the old CA in memory, causing mTLS failures for any Envoy
// pod that has already reloaded its new leaf cert via SDS.
//
// A rolling restart ensures every Rate Limit pod starts fresh with the updated
// CA bundle. The restart respects any PodDisruptionBudget the operator has
// configured and uses the deployment's existing RollingUpdate strategy, so no
// replica is terminated until a replacement is healthy.
//
// If the Rate Limit deployment does not exist (Rate Limit is not enabled) the
// function returns nil without error.
func rolloutRestartRateLimitDeployment(ctx context.Context, cli client.Client, cfg *config.Server) error {
	log := cfg.Logger

	deployment := &appsv1.Deployment{}
	key := types.NamespacedName{
		Namespace: cfg.ControllerNamespace,
		Name:      ratelimit.InfraName,
	}
	if err := cli.Get(ctx, key, deployment); err != nil {
		if kerrors.IsNotFound(err) {
			// Rate Limit is not deployed; nothing to restart.
			log.Info("rate limit deployment not found, skipping restart")
			return nil
		}
		return fmt.Errorf("failed to get rate limit deployment %s/%s: %w", key.Namespace, key.Name, err)
	}

	patched := deployment.DeepCopy()
	if patched.Spec.Template.Annotations == nil {
		patched.Spec.Template.Annotations = make(map[string]string)
	}
	patched.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"] = metav1.Now().UTC().Format(time.RFC3339)

	if err := cli.Patch(ctx, patched, client.MergeFrom(deployment)); err != nil {
		return fmt.Errorf("failed to patch rate limit deployment %s/%s: %w", key.Namespace, key.Name, err)
	}

	log.Info("triggered rolling restart of rate limit deployment",
		"namespace", key.Namespace, "name", key.Name)
	return nil
}

func patchTopologyInjectorWebhook(ctx context.Context, cli client.Client, cfg *config.Server) error {
	if disableTopologyInjector {
		return nil
	}

	webhookConfigName := fmt.Sprintf("%s.%s", topologyWebhookNamePrefix, cfg.ControllerNamespace)
	webhookCfg := &admissionregistrationv1.MutatingWebhookConfiguration{}
	if err := cli.Get(ctx, client.ObjectKey{Name: webhookConfigName}, webhookCfg); err != nil {
		return fmt.Errorf("failed to get mutating webhook configuration: %w", err)
	}

	secretName := types.NamespacedName{Name: "envoy-gateway", Namespace: cfg.ControllerNamespace}
	current := &corev1.Secret{}
	if err := cli.Get(ctx, secretName, current); err != nil {
		return fmt.Errorf("failed to get secret %s/%s: %w", current.Namespace, current.Name, err)
	}

	var updated bool
	desiredBundle := current.Data["ca.crt"]
	for i := range webhookCfg.Webhooks {
		if !bytes.Equal(desiredBundle, webhookCfg.Webhooks[i].ClientConfig.CABundle) {
			webhookCfg.Webhooks[i].ClientConfig.CABundle = desiredBundle
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
func outputCertsForLocal(localPath string, certs *crypto.Certificates) error {
	egDir := path.Join(localPath, "envoy-gateway")
	if err := file.WriteDir(certs.CACertificate, egDir, "ca.crt"); err != nil {
		return err
	}
	if err := file.WriteDir(certs.EnvoyGatewayCertificate, egDir, "tls.crt"); err != nil {
		return err
	}
	if err := file.WriteDir(certs.EnvoyGatewayPrivateKey, egDir, "tls.key"); err != nil {
		return err
	}

	envoyDir := path.Join(localPath, "envoy")
	if err := file.WriteDir(certs.CACertificate, envoyDir, "ca.crt"); err != nil {
		return err
	}
	if err := file.WriteDir(certs.EnvoyCertificate, envoyDir, "tls.crt"); err != nil {
		return err
	}
	if err := file.WriteDir(certs.EnvoyPrivateKey, envoyDir, "tls.key"); err != nil {
		return err
	}

	rlDir := path.Join(localPath, "envoy-rate-limit")
	if err := file.WriteDir(certs.CACertificate, rlDir, "ca.crt"); err != nil {
		return err
	}
	if err := file.WriteDir(certs.EnvoyRateLimitCertificate, rlDir, "tls.crt"); err != nil {
		return err
	}
	if err := file.WriteDir(certs.EnvoyRateLimitPrivateKey, rlDir, "tls.key"); err != nil {
		return err
	}

	if err := file.WriteDir(certs.OIDCHMACSecret, path.Join(localPath, "envoy-oidc-hmac"), "hmac-secret"); err != nil {
		return err
	}

	return nil
}
