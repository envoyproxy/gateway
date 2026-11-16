// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	certificatesv1 "k8s.io/api/certificates/v1"
	certificatesv1b1 "k8s.io/api/certificates/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/logging"
)

func TestGetClusterTrustBundle(t *testing.T) {
	const (
		name       = "ca-bundle"
		signerName = "example.com/signer"
		bundle     = "-----BEGIN CERTIFICATE-----\nfoo\n-----END CERTIFICATE-----\n"
	)

	testCases := []struct {
		name      string
		gv        schema.GroupVersion
		objs      []client.Object
		expectErr string
	}{
		{
			name: "certificates.k8s.io/v1",
			gv:   certificatesv1.SchemeGroupVersion,
			objs: []client.Object{&certificatesv1.ClusterTrustBundle{
				ObjectMeta: metav1.ObjectMeta{Name: name},
				Spec:       certificatesv1.ClusterTrustBundleSpec{SignerName: signerName, TrustBundle: bundle},
			}},
		},
		{
			name: "certificates.k8s.io/v1beta1 is converted to v1",
			gv:   certificatesv1b1.SchemeGroupVersion,
			objs: []client.Object{&certificatesv1b1.ClusterTrustBundle{
				ObjectMeta: metav1.ObjectMeta{Name: name},
				Spec:       certificatesv1b1.ClusterTrustBundleSpec{SignerName: signerName, TrustBundle: bundle},
			}},
		},
		{
			name:      "not found",
			gv:        certificatesv1.SchemeGroupVersion,
			expectErr: "not found",
		},
		{
			name:      "API not served",
			expectErr: "ClusterTrustBundle API is not served by the cluster",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := &gatewayAPIReconciler{
				log:                  logging.DefaultLogger(os.Stdout, egv1a1.LogLevelInfo),
				clusterTrustBundleGV: tc.gv,
				client: fakeclient.NewClientBuilder().
					WithScheme(envoygateway.GetScheme()).
					WithObjects(tc.objs...).
					Build(),
			}

			ctb, err := r.getClusterTrustBundle(context.Background(), name)
			if tc.expectErr != "" {
				require.ErrorContains(t, err, tc.expectErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, name, ctb.Name)
			require.Equal(t, signerName, ctb.Spec.SignerName)
			require.Equal(t, bundle, ctb.Spec.TrustBundle)
			if tc.gv == certificatesv1b1.SchemeGroupVersion {
				require.Equal(t, certificatesv1.SchemeGroupVersion.String(), ctb.APIVersion)
				require.Equal(t, resource.KindClusterTrustBundle, ctb.Kind)
			}
		})
	}
}
