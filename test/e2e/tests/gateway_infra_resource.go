// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/conformance/utils/tlog"
)

func init() {
	ConformanceTests = append(ConformanceTests, GatewayInfraResource)
}

var GatewayInfraResource = suite.ConformanceTest{
	ShortName:   "GatewayInfraResource",
	Description: "Gateway Infra Resource E2E Test",
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		gatewayTypeMeta := metav1.TypeMeta{
			Kind:       "Gateway",
			APIVersion: "gateway.networking.k8s.io/v1",
		}
		gatewayObjMeta := metav1.ObjectMeta{
			Name:      "e2e-test-infra",
			Namespace: "envoy-gateway-system",
		}

		t.Cleanup(func() {
			// delete the gateway if it still exists after the test is done
			gwObj := &gwapiv1.Gateway{
				TypeMeta:   gatewayTypeMeta,
				ObjectMeta: gatewayObjMeta,
			}
			_ = suite.Client.Delete(context.TODO(), gwObj)
		})

		labelSelector := labels.SelectorFromSet(labels.Set{
			"gateway.envoyproxy.io/owning-gateway-name": gatewayObjMeta.Name,
			"app.kubernetes.io/managed-by":              "envoy-gateway",
		})

		tlog.Logf(t, "creating gateway")
		{
			newGatewayObj := &gwapiv1.Gateway{
				TypeMeta:   gatewayTypeMeta,
				ObjectMeta: gatewayObjMeta,
				Spec: gwapiv1.GatewaySpec{
					GatewayClassName: gwapiv1.ObjectName(suite.GatewayClassName),
					Listeners: []gwapiv1.Listener{
						{
							Name:     "http",
							Port:     8000,
							Protocol: "HTTP",
						},
						{
							Name:     "my-tcp",
							Port:     5432,
							Protocol: "TCP",
						},
					},
				},
			}
			err := suite.Client.Create(t.Context(), newGatewayObj, &client.CreateOptions{
				FieldManager: "e2e-test",
			})
			require.NoError(t, err)

			require.NoError(t, wait.PollUntilContextTimeout(t.Context(), time.Second, suite.TimeoutConfig.MaxTimeToConsistency, true, func(ctx context.Context) (bool, error) {
				gatewayDeploymentList := &appsv1.DeploymentList{}
				err = suite.Client.List(ctx, gatewayDeploymentList, &client.ListOptions{
					LabelSelector: labelSelector,
					Namespace:     gatewayObjMeta.Namespace,
				})
				if err != nil {
					tlog.Logf(t, "error listing gateway deployment: %v", err)
					return false, nil
				}
				return len(gatewayDeploymentList.Items) == 1, nil
			}))
		}

		tlog.Logf(t, "updating listeners")
		{
			newListenerTCPName := "custom-tcp"
			newListenerHTTPPort := int32(8001)

			require.NoError(t, wait.PollUntilContextTimeout(t.Context(), time.Second, suite.TimeoutConfig.MaxTimeToConsistency, true, func(ctx context.Context) (bool, error) {
				gateway := &gwapiv1.Gateway{
					TypeMeta:   gatewayTypeMeta,
					ObjectMeta: gatewayObjMeta,
				}
				err := suite.Client.Get(ctx, client.ObjectKeyFromObject(gateway), gateway)
				if err != nil {
					tlog.Logf(t, "error getting gateway: %v", err)
					return false, nil
				}

				gateway.Spec.Listeners = []gwapiv1.Listener{
					{
						Name:     "http",
						Port:     newListenerHTTPPort,
						Protocol: "HTTP",
					},
					{
						Name:     gwapiv1.SectionName(newListenerTCPName),
						Port:     5432,
						Protocol: "TCP",
					},
				}

				err = suite.Client.Update(ctx, gateway)
				if err != nil {
					tlog.Logf(t, "failed to update gateway: %v", err)
					return false, nil
				}

				return true, nil
			}))

			require.NoError(t, wait.PollUntilContextTimeout(t.Context(), time.Second, suite.TimeoutConfig.MaxTimeToConsistency, true, func(ctx context.Context) (bool, error) {
				svcList := &corev1.ServiceList{}
				err := suite.Client.List(ctx, svcList, &client.ListOptions{
					LabelSelector: labelSelector,
					Namespace:     gatewayObjMeta.Namespace,
				})
				if err != nil {
					tlog.Logf(t, "error listing gateway deployment: %v", err)
					return false, nil
				}
				if len(svcList.Items) == 0 {
					return false, nil
				}

				svc := svcList.Items[0]

				var isTCPPortNameMatch, isHTTPPortNumberMatch bool
				for _, port := range svc.Spec.Ports {
					tlog.Logf(t, "service port name: %v, service port number: %d", port.Name, port.Port)
					if port.Name == fmt.Sprintf("%s-%d", "tcp", 5432) {
						isTCPPortNameMatch = true
					}

					if port.Name == fmt.Sprintf("%s-%d", "http", newListenerHTTPPort) {
						isHTTPPortNumberMatch = true
					}
				}

				if !isTCPPortNameMatch {
					tlog.Logf(t, "container expected TCP port name '%v' is not found", fmt.Sprintf("%s-%d", newListenerTCPName, 5432))
					return false, nil
				}

				if !isHTTPPortNumberMatch {
					tlog.Logf(t, "container expected HTTP port name '%v' is not found", fmt.Sprintf("%s-%d", "http", newListenerHTTPPort))
					return false, nil
				}

				return true, nil
			}))
		}

		tlog.Logf(t, "deleting gateway")
		{
			gwObj := &gwapiv1.Gateway{
				TypeMeta:   gatewayTypeMeta,
				ObjectMeta: gatewayObjMeta,
			}
			require.NoError(t, suite.Client.Delete(t.Context(), gwObj))
			require.NoError(t, wait.PollUntilContextTimeout(t.Context(), time.Second, suite.TimeoutConfig.MaxTimeToConsistency, true, func(ctx context.Context) (bool, error) {
				gatewayDeploymentList := &appsv1.DeploymentList{}
				err := suite.Client.List(ctx, gatewayDeploymentList, &client.ListOptions{
					LabelSelector: labelSelector,
					Namespace:     gatewayObjMeta.Namespace,
				})
				if err != nil {
					tlog.Logf(t, "error listing gateway deployment: %v", err)
					return false, nil
				}
				return len(gatewayDeploymentList.Items) == 0, nil
			}))
		}
	},
}
