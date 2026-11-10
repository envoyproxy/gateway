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
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/infrastructure/kubernetes/ratelimit"
	"github.com/envoyproxy/gateway/internal/message"
)

func createRateLimitTLSSecret(t *testing.T, client client.Client) {
	secretErr := client.Create(context.Background(), &corev1.Secret{
		Type: corev1.SecretTypeTLS,
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ratelimit-cert",
			Namespace: "envoy-gateway-system",
			Labels: map[string]string{
				"control-plane": "envoy-gateway",
			},
		},
	})
	require.NoError(t, secretErr)
}

func createEnvoyGatewayService(t *testing.T, client client.Client, ns string) {
	err := client.Create(context.Background(), &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "envoy-gateway",
			Namespace: ns,
		},
	})
	require.NoError(t, err)
}

func createEnvoyGatewayDeployment(t *testing.T, client client.Client, ns string) {
	err := client.Create(context.Background(), &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "envoy-gateway",
			Namespace: ns,
		},
	})
	require.NoError(t, err)
}

func createEnvoyGatewayServiceAccount(t *testing.T, client client.Client, ns string) {
	err := client.Create(context.Background(), &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceAccount",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "envoy-gateway",
			Namespace: ns,
		},
	})
	require.NoError(t, err)
}

func TestCreateRateLimitInfra(t *testing.T) {
	testCases := []struct {
		name            string
		ownerReferences []string
		expect          bool
	}{
		{
			name: "default infra",
			ownerReferences: []string{
				ratelimit.ResourceKindService,
				ratelimit.ResourceKindDeployment,
				ratelimit.ResourceKindServiceAccount,
			},
			expect: true,
		},
		{
			name: "default infra but missing service owner reference",
			ownerReferences: []string{
				ratelimit.ResourceKindDeployment,
				ratelimit.ResourceKindServiceAccount,
			},
			expect: false,
		},
		{
			name: "default infra but missing deployment owner reference",
			ownerReferences: []string{
				ratelimit.ResourceKindService,
				ratelimit.ResourceKindServiceAccount,
			},
			expect: false,
		},
		{
			name: "default infra but missing service account owner reference",
			ownerReferences: []string{
				ratelimit.ResourceKindService,
				ratelimit.ResourceKindDeployment,
			},
			expect: false,
		},
		{
			name:   "default infra but missing all owner reference",
			expect: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			kube := newTestInfra(t)

			// Ratelimit infra creation require Gateway infra as owner reference.
			for _, ref := range tc.ownerReferences {
				switch ref {
				case ratelimit.ResourceKindService:
					createEnvoyGatewayService(t, kube.Client.Client, kube.ControllerNamespace)
				case ratelimit.ResourceKindDeployment:
					createEnvoyGatewayDeployment(t, kube.Client.Client, kube.ControllerNamespace)
				case ratelimit.ResourceKindServiceAccount:
					createEnvoyGatewayServiceAccount(t, kube.Client.Client, kube.ControllerNamespace)
				}
			}

			createRateLimitTLSSecret(t, kube.Client.Client)

			err := kube.CreateOrUpdateRateLimitInfra(context.Background())
			if !tc.expect {
				require.Error(t, err)
			} else {
				require.NoError(t, err)

				// Verify all resources were created via the fake kube client.
				sa := &corev1.ServiceAccount{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: kube.ControllerNamespace,
						Name:      ratelimit.InfraName,
					},
				}
				require.NoError(t, kube.Client.Get(context.Background(), client.ObjectKeyFromObject(sa), sa))

				deploy := &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: kube.ControllerNamespace,
						Name:      ratelimit.InfraName,
					},
				}
				require.NoError(t, kube.Client.Get(context.Background(), client.ObjectKeyFromObject(deploy), deploy))

				svc := &corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: kube.ControllerNamespace,
						Name:      ratelimit.InfraName,
					},
				}
				require.NoError(t, kube.Client.Get(context.Background(), client.ObjectKeyFromObject(svc), svc))
			}
		})
	}
}

func TestDeleteRateLimitInfra(t *testing.T) {
	testCases := []struct {
		name   string
		expect bool
	}{
		{
			name:   "default infra",
			expect: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			kube := newTestInfra(t)

			createRateLimitTLSSecret(t, kube.Client.Client)

			err := kube.DeleteRateLimitInfra(context.Background())
			if !tc.expect {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestCoLocatedRateLimitInfra verifies the shared-namespace co-location
// guarantee for the rate limit infra: the resources are scoped by the
// per-instance resource prefix, so an instance that does not enable rate
// limiting deletes only its own resources and never the rate limit
// Deployment or Service of a co-located instance.
func TestCoLocatedRateLimitInfra(t *testing.T) {
	t.Parallel()
	cli := fakeclient.NewClientBuilder().
		WithScheme(envoygateway.GetScheme()).
		WithInterceptorFuncs(interceptorFunc).
		Build()

	newInfra := func(prefix string, enableRateLimit bool) *Infra {
		cfg, err := config.New(os.Stdout, os.Stderr)
		require.NoError(t, err)
		cfg.ControllerResourcePrefix = prefix
		cfg.EnvoyGateway = &egv1a1.EnvoyGateway{
			TypeMeta: metav1.TypeMeta{},
			EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{
				RateLimit: &egv1a1.RateLimit{
					Backend: egv1a1.RateLimitDatabaseBackend{
						Type: egv1a1.RedisBackendType,
						Redis: &egv1a1.RateLimitRedisSettings{
							URL: new(""),
							TLS: &egv1a1.RedisTLSSettings{
								CertificateRef: &gwapiv1.SecretObjectReference{
									Name: "ratelimit-cert",
								},
							},
						},
					},
				},
			},
		}
		if !enableRateLimit {
			cfg.EnvoyGateway.RateLimit = nil
		}
		errorNotifier := message.RunnerErrorNotifier{RunnerName: t.Name(), RunnerErrors: &message.RunnerErrors{}}
		return NewInfra(cli, cfg, errorNotifier)
	}

	// Instance A has global rate limiting enabled and provisions its own
	// rate limit infra, owned by its own (prefixed) control-plane resources.
	infraA := newInfra("tenant-a", true)
	owner := config.InstanceResourceName("tenant-a", config.EnvoyGatewayServiceName)
	require.NoError(t, cli.Create(context.Background(), &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: owner, Namespace: infraA.ControllerNamespace},
	}))
	require.NoError(t, cli.Create(context.Background(), &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: owner, Namespace: infraA.ControllerNamespace},
	}))
	require.NoError(t, cli.Create(context.Background(), &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{Name: owner, Namespace: infraA.ControllerNamespace},
	}))
	createRateLimitTLSSecret(t, cli)
	require.NoError(t, infraA.CreateOrUpdateRateLimitInfra(context.Background()))

	infraName := config.InstanceResourceName("tenant-a", ratelimit.InfraName)
	deploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: infraA.ControllerNamespace,
			Name:      infraName,
		},
	}
	require.NoError(t, cli.Get(context.Background(), client.ObjectKeyFromObject(deploy), deploy))

	// Instance B has rate limiting disabled (the default) and, on startup,
	// deletes its own (absent) rate limit infra. It must not touch the
	// resources of co-located instance A.
	infraB := newInfra("tenant-b", false)
	require.NoError(t, infraB.DeleteRateLimitInfra(context.Background()))

	require.NoError(t, cli.Get(context.Background(), client.ObjectKeyFromObject(deploy), deploy))
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: infraA.ControllerNamespace,
			Name:      infraName,
		},
	}
	require.NoError(t, cli.Get(context.Background(), client.ObjectKeyFromObject(svc), svc))
}
