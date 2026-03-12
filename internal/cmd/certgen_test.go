// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	"github.com/envoyproxy/gateway/internal/crypto"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
)

func TestGetCertgenCommand(t *testing.T) {
	got := GetCertGenCommand()
	assert.Equal(t, "certgen", got.Use)
}

func TestOutputCertsForLocal(t *testing.T) {
	cfg, err := config.New(os.Stdout, os.Stderr)
	require.NoError(t, err)

	certs, err := crypto.GenerateCerts(cfg)
	require.NoError(t, err)

	tmpDir := t.TempDir()
	err = outputCertsForLocal(tmpDir, certs)
	require.NoError(t, err)

	assert.FileExists(t, filepath.Join(tmpDir, "envoy-gateway", "ca.crt"))
	assert.FileExists(t, filepath.Join(tmpDir, "envoy-gateway", "tls.crt"))
	assert.FileExists(t, filepath.Join(tmpDir, "envoy-gateway", "tls.key"))
	assert.FileExists(t, filepath.Join(tmpDir, "envoy", "ca.crt"))
	assert.FileExists(t, filepath.Join(tmpDir, "envoy", "tls.crt"))
	assert.FileExists(t, filepath.Join(tmpDir, "envoy", "tls.key"))
	assert.FileExists(t, filepath.Join(tmpDir, "envoy-rate-limit", "ca.crt"))
	assert.FileExists(t, filepath.Join(tmpDir, "envoy-rate-limit", "tls.crt"))
	assert.FileExists(t, filepath.Join(tmpDir, "envoy-rate-limit", "tls.key"))
	assert.FileExists(t, filepath.Join(tmpDir, "envoy-oidc-hmac", "hmac-secret"))
}

func TestPatchTopologyWebhook(t *testing.T) {
	cfg, err := config.New(os.Stdout, os.Stderr)
	require.NoError(t, err)

	cases := []struct {
		caseName  string
		webhook   *admissionregistrationv1.MutatingWebhookConfiguration
		secret    *corev1.Secret
		wantPatch bool
	}{
		{
			caseName: "Update caBundle",
			webhook: &admissionregistrationv1.MutatingWebhookConfiguration{
				ObjectMeta: metav1.ObjectMeta{
					Name:   fmt.Sprintf("envoy-gateway-topology-injector.%s", cfg.ControllerNamespace),
					Labels: map[string]string{topologyInjectorComponentLabel: topologyInjectorComponentValue},
				},
				Webhooks: []admissionregistrationv1.MutatingWebhook{{ClientConfig: admissionregistrationv1.WebhookClientConfig{}}},
			},
			secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{Name: "envoy-gateway", Namespace: cfg.ControllerNamespace},
				Data:       map[string][]byte{"ca.crt": []byte("foo")},
			},
			wantPatch: true,
		},
		{
			caseName: "Update caBundle by label when name contains different namespace",
			webhook: &admissionregistrationv1.MutatingWebhookConfiguration{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "envoy-gateway-topology-injector.envoy-gateway-system",
					Labels: map[string]string{topologyInjectorComponentLabel: topologyInjectorComponentValue},
				},
				Webhooks: []admissionregistrationv1.MutatingWebhook{{ClientConfig: admissionregistrationv1.WebhookClientConfig{}}},
			},
			secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{Name: "envoy-gateway", Namespace: cfg.ControllerNamespace},
				Data:       map[string][]byte{"ca.crt": []byte("foo")},
			},
			wantPatch: true,
		},
		{
			caseName: "No-op",
			webhook: &admissionregistrationv1.MutatingWebhookConfiguration{
				ObjectMeta: metav1.ObjectMeta{
					Name:   fmt.Sprintf("envoy-gateway-topology-injector.%s", cfg.ControllerNamespace),
					Labels: map[string]string{topologyInjectorComponentLabel: topologyInjectorComponentValue},
				},
				Webhooks: []admissionregistrationv1.MutatingWebhook{{ClientConfig: admissionregistrationv1.WebhookClientConfig{CABundle: []byte("foo")}}},
			},
			secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{Name: "envoy-gateway", Namespace: cfg.ControllerNamespace},
				Data:       map[string][]byte{"ca.crt": []byte("foo")},
			},
			wantPatch: false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.caseName, func(t *testing.T) {
			topologyWebhookName = ""
			fakeClient := fake.NewClientBuilder().
				WithRuntimeObjects(tc.webhook, tc.secret).
				Build()
			beforeWebhook := &admissionregistrationv1.MutatingWebhookConfiguration{}
			require.NoError(t, fakeClient.Get(context.Background(), client.ObjectKey{Name: tc.webhook.Name}, beforeWebhook))

			err = patchTopologyInjectorWebhook(context.Background(), fakeClient, cfg)
			require.NoError(t, err)

			afterWebhook := &admissionregistrationv1.MutatingWebhookConfiguration{}
			require.NoError(t, fakeClient.Get(context.Background(), client.ObjectKey{Name: tc.webhook.Name}, afterWebhook))

			require.Equal(t, afterWebhook.Webhooks[0].ClientConfig.CABundle, tc.secret.Data["ca.crt"])
			assert.Equal(t, tc.wantPatch, beforeWebhook.GetResourceVersion() != afterWebhook.GetResourceVersion())
		})
	}

	t.Run("disabled topology injector", func(t *testing.T) {
		disableTopologyInjector = true
		defer func() { disableTopologyInjector = false }()
		fakeClient := fake.NewClientBuilder().Build()
		err = patchTopologyInjectorWebhook(context.Background(), fakeClient, cfg)
		require.NoError(t, err)
	})

	t.Run("list error", func(t *testing.T) {
		topologyWebhookName = ""
		fakeClient := fake.NewClientBuilder().WithInterceptorFuncs(interceptor.Funcs{
			List: func(_ context.Context, _ client.WithWatch, _ client.ObjectList, _ ...client.ListOption) error {
				return fmt.Errorf("list failed")
			},
		}).Build()
		err = patchTopologyInjectorWebhook(context.Background(), fakeClient, cfg)
		require.ErrorContains(t, err, "list failed")
	})

	t.Run("no webhooks found by label", func(t *testing.T) {
		topologyWebhookName = ""
		fakeClient := fake.NewClientBuilder().Build()
		err = patchTopologyInjectorWebhook(context.Background(), fakeClient, cfg)
		require.ErrorContains(t, err, "no mutating webhook configurations found")
	})

	t.Run("secret not found", func(t *testing.T) {
		topologyWebhookName = ""
		webhook := &admissionregistrationv1.MutatingWebhookConfiguration{
			ObjectMeta: metav1.ObjectMeta{
				Name:   fmt.Sprintf("envoy-gateway-topology-injector.%s", cfg.ControllerNamespace),
				Labels: map[string]string{topologyInjectorComponentLabel: topologyInjectorComponentValue},
			},
			Webhooks: []admissionregistrationv1.MutatingWebhook{{ClientConfig: admissionregistrationv1.WebhookClientConfig{}}},
		}
		fakeClient := fake.NewClientBuilder().WithRuntimeObjects(webhook).Build()
		err = patchTopologyInjectorWebhook(context.Background(), fakeClient, cfg)
		require.ErrorContains(t, err, "failed to get secret")
	})
}

func TestPatchTopologyWebhookByName(t *testing.T) {
	cfg, err := config.New(os.Stdout, os.Stderr)
	require.NoError(t, err)

	cases := []struct {
		caseName    string
		webhookName string
		wantErr     bool
	}{
		{
			caseName:    "patch by explicit name",
			webhookName: "envoy-gateway-topology-injector.envoy-gateway-system",
			wantErr:     false,
		},
		{
			caseName:    "explicit name not found",
			webhookName: "envoy-gateway-topology-injector.does-not-exist",
			wantErr:     true,
		},
	}

	webhook := &admissionregistrationv1.MutatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "envoy-gateway-topology-injector.envoy-gateway-system",
			Labels: map[string]string{topologyInjectorComponentLabel: topologyInjectorComponentValue},
		},
		Webhooks: []admissionregistrationv1.MutatingWebhook{{ClientConfig: admissionregistrationv1.WebhookClientConfig{}}},
	}
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "envoy-gateway", Namespace: cfg.ControllerNamespace},
		Data:       map[string][]byte{"ca.crt": []byte("foo")},
	}
	fakeClient := fake.NewClientBuilder().WithRuntimeObjects(webhook, secret).Build()

	for _, tc := range cases {
		t.Run(tc.caseName, func(t *testing.T) {
			topologyWebhookName = tc.webhookName
			defer func() { topologyWebhookName = "" }()

			err = patchTopologyInjectorWebhook(context.Background(), fakeClient, cfg)
			if tc.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.webhookName)
			} else {
				require.NoError(t, err)
				after := &admissionregistrationv1.MutatingWebhookConfiguration{}
				require.NoError(t, fakeClient.Get(context.Background(), client.ObjectKey{Name: tc.webhookName}, after))
				require.Equal(t, []byte("foo"), after.Webhooks[0].ClientConfig.CABundle)
			}
		})
	}
}
