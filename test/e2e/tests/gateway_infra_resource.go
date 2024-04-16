// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e
// +build e2e

package tests

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"

	"github.com/envoyproxy/gateway/internal/utils"
)

func init() {
	// nolint
	//ConformanceTests = append(ConformanceTests, GatewayInfraResourceTest)
}

// nolint
var GatewayInfraResourceTest = suite.ConformanceTest{
	ShortName:   "GatewayInfraResourceTest",
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

		labelSelector := labels.SelectorFromSet(labels.Set{"gateway.envoyproxy.io/owning-gateway-name": gatewayObjMeta.Name})

		var awaitOperation sync.WaitGroup

		t.Run("create gateway", func(t *testing.T) {
			awaitOperation.Add(1)

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

			err := suite.Client.Patch(context.TODO(), newGatewayObj, client.Apply, client.ForceOwnership, client.FieldOwner("e2e-test"))
			require.NoError(t, err)

			<-time.After(time.Millisecond * 300)

			var gatewayDeploymentList appsv1.DeploymentList
			err = suite.Client.List(context.TODO(), &gatewayDeploymentList, &client.ListOptions{
				LabelSelector: labelSelector,
				Namespace:     gatewayObjMeta.Namespace,
			})
			require.NoError(t, err)
			require.Len(t, gatewayDeploymentList.Items, 1)

			awaitOperation.Done()
		})

		awaitOperation.Wait()
		t.Run("update gateway - listener changes", func(t *testing.T) {
			awaitOperation.Add(1)

			newListenerTCPName := "custom-tcp"
			newListenerHTTPPort := int32(8001)

			changedGatewayObj := &gwapiv1.Gateway{
				TypeMeta:   gatewayTypeMeta,
				ObjectMeta: gatewayObjMeta,
				Spec: gwapiv1.GatewaySpec{
					GatewayClassName: gwapiv1.ObjectName(suite.GatewayClassName),
					Listeners: []gwapiv1.Listener{
						{
							Name:     "http",
							Port:     gwapiv1.PortNumber(newListenerHTTPPort),
							Protocol: "HTTP",
						},
						{
							Name:     gwapiv1.SectionName(newListenerTCPName),
							Port:     5432,
							Protocol: "TCP",
						},
					},
				},
			}

			err := suite.Client.Patch(context.TODO(), changedGatewayObj, client.Apply, client.ForceOwnership, client.FieldOwner("e2e-test"))
			require.NoError(t, err)

			<-time.After(time.Millisecond * 300)
			var gatewayDeploymentList appsv1.DeploymentList
			err = suite.Client.List(context.TODO(), &gatewayDeploymentList, &client.ListOptions{
				LabelSelector: labelSelector,
				Namespace:     gatewayObjMeta.Namespace,
			})
			require.NoError(t, err)
			require.Len(t, gatewayDeploymentList.Items, 1)

			gatewayDeployment := gatewayDeploymentList.Items[0]

			for _, container := range gatewayDeployment.Spec.Template.Spec.Containers {
				var isTCPPortNameMatch, isHTTPPortNumberMatch bool

				hashedPortName := utils.GetHashedName(newListenerTCPName, 6)
				if container.Name == "envoy" {
					for _, port := range container.Ports {
						if port.Name == hashedPortName {
							isTCPPortNameMatch = true
						}

						if port.ContainerPort == newListenerHTTPPort {
							isHTTPPortNumberMatch = true
						}
					}

					if !isTCPPortNameMatch {
						t.Errorf("container expected TCP port name '%v' is not found", hashedPortName)
					}

					if !isHTTPPortNumberMatch {
						t.Errorf("container expected HTTP port number '%v' is not found", hashedPortName)
					}
				}
			}

			awaitOperation.Done()
		})

		awaitOperation.Wait()
		t.Run("delete gateway", func(t *testing.T) {
			gwObj := &gwapiv1.Gateway{
				TypeMeta:   gatewayTypeMeta,
				ObjectMeta: gatewayObjMeta,
			}

			err := suite.Client.Delete(context.TODO(), gwObj)
			require.NoError(t, err)

			<-time.After(time.Millisecond * 300)

			var gatewayDeploymentList appsv1.DeploymentList
			err = suite.Client.List(context.TODO(), &gatewayDeploymentList, &client.ListOptions{
				LabelSelector: labelSelector,
				Namespace:     gatewayObjMeta.Namespace,
			})
			require.NoError(t, err)
			require.Empty(t, gatewayDeploymentList.Items)
		})
	},
}
