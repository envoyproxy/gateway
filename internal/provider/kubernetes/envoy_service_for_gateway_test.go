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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway"
	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/logging"
)

// TestEnvoyServiceForGatewayUncachedFallback ensures a transient empty result from the
// cached client does not make the Envoy Service look absent (which would wipe the Gateway
// status addresses): the lookup confirms via the uncached API reader before giving up.
func TestEnvoyServiceForGatewayUncachedFallback(t *testing.T) {
	const (
		gwNS    = "default"
		gwName  = "gw-1"
		envoyNS = "envoy-gateway-system"
		svcName = "envoy-default-gw-1-abcd"
	)

	gw := &gwapiv1.Gateway{
		ObjectMeta: metav1.ObjectMeta{Namespace: gwNS, Name: gwName},
		Spec:       gwapiv1.GatewaySpec{GatewayClassName: "gc"},
	}
	newEnvoySvc := func() *corev1.Service {
		return &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: envoyNS,
				Name:      svcName,
				Labels:    gatewayapi.GatewayOwnerLabels(gwNS, gwName),
			},
			Spec: corev1.ServiceSpec{Type: corev1.ServiceTypeLoadBalancer},
		}
	}
	newReconciler := func(cached, uncached []client.Object) *gatewayAPIReconciler {
		return &gatewayAPIReconciler{
			log:           logging.DefaultLogger(os.Stdout, egv1a1.LogLevelInfo),
			namespace:     envoyNS,
			mergeGateways: sets.New[string](),
			client: fakeclient.NewClientBuilder().
				WithScheme(envoygateway.GetScheme()).WithObjects(cached...).Build(),
			apiReader: fakeclient.NewClientBuilder().
				WithScheme(envoygateway.GetScheme()).WithObjects(uncached...).Build(),
		}
	}

	t.Run("empty in cache but present via uncached reader is returned", func(t *testing.T) {
		r := newReconciler(nil, []client.Object{newEnvoySvc()})
		svc, err := r.envoyServiceForGateway(context.Background(), gw)
		require.NoError(t, err)
		require.NotNil(t, svc)
		require.Equal(t, svcName, svc.Name)
	})

	t.Run("absent in both cache and API server returns nil", func(t *testing.T) {
		r := newReconciler(nil, nil)
		svc, err := r.envoyServiceForGateway(context.Background(), gw)
		require.NoError(t, err)
		require.Nil(t, svc)
	})

	t.Run("present in cache is returned", func(t *testing.T) {
		r := newReconciler([]client.Object{newEnvoySvc()}, nil)
		svc, err := r.envoyServiceForGateway(context.Background(), gw)
		require.NoError(t, err)
		require.NotNil(t, svc)
		require.Equal(t, svcName, svc.Name)
	})
}
