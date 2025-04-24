// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package cmd

import (
	"context"
	"fmt"
	v1 "k8s.io/api/admissionregistration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"path/filepath"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/envoyproxy/gateway/internal/crypto"
)

func TestGetCertgenCommand(t *testing.T) {
	got := GetCertGenCommand()
	assert.Equal(t, "certgen", got.Use)
}

func TestOutputCertsForLocal(t *testing.T) {
	cfg, err := getConfig(os.Stdout)
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
	cfg, err := getConfig(os.Stdout)
	require.NoError(t, err)

	cases := []struct {
		caseName  string
		webhook   *v1.MutatingWebhookConfiguration
		caBundle  []byte
		wantErr   error
		wantPatch bool
	}{
		{
			caseName: "Update caBundle",
			webhook: &v1.MutatingWebhookConfiguration{
				ObjectMeta: metav1.ObjectMeta{
					Name: fmt.Sprintf("%s.%s", topologyWebhookNamePrefix, cfg.Namespace),
				},
				Webhooks: []v1.MutatingWebhook{{ClientConfig: v1.WebhookClientConfig{}}},
			},
			caBundle:  []byte("foo"),
			wantErr:   nil,
			wantPatch: true,
		},
		{
			caseName: "No-op",
			webhook: &v1.MutatingWebhookConfiguration{
				ObjectMeta: metav1.ObjectMeta{
					Name: fmt.Sprintf("%s.%s", topologyWebhookNamePrefix, cfg.Namespace),
				},
				Webhooks: []v1.MutatingWebhook{{ClientConfig: v1.WebhookClientConfig{CABundle: []byte("foo")}}},
			},
			caBundle:  []byte("foo"),
			wantPatch: false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.caseName, func(t *testing.T) {
			fakeClient := fake.NewClientBuilder().
				WithRuntimeObjects(tc.webhook).
				Build()
			beforeWebhook := &v1.MutatingWebhookConfiguration{}
			require.NoError(t, fakeClient.Get(context.Background(), client.ObjectKey{Name: tc.webhook.Name}, beforeWebhook))
			err = patchTopologyInjectorWebhook(context.Background(), fakeClient, cfg, tc.caBundle)

			require.NoError(t, err)

			afterWebhook := &v1.MutatingWebhookConfiguration{}
			require.NoError(t, fakeClient.Get(context.Background(), client.ObjectKey{Name: tc.webhook.Name}, afterWebhook))

			require.Equal(t, afterWebhook.Webhooks[0].ClientConfig.CABundle, tc.caBundle)
			assert.Equal(t, tc.wantPatch, beforeWebhook.GetResourceVersion() != afterWebhook.GetResourceVersion())
		})
	}
}
