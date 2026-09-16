// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/conformance/utils/tlog"

	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
)

func init() {
	RemoteInfraTests = append(RemoteInfraTests, RemoteInfraTCPRouteTest)
}

// Prerequisites the make target sets up:
//   - test/config/envoy-gateaway-config/remote-infra.yaml configures EG to
//     dial the sidecar over a Unix domain socket.
//   - test/e2e/remote_infra/sidecar-patch.yaml injects the remote-infra
//     server as a native sidecar into the envoy-gateway Deployment.

// RemoteInfraTCPRouteTest applies the TCPRoute manifest used by the standard
// TCPRoute test and verifies traffic flows through the proxy fleet stood up
// by the remote infrastructure provider.
var RemoteInfraTCPRouteTest = suite.ConformanceTest{
	ShortName:   "RemoteInfraTCPRoute",
	Description: "Verify TCP traffic routes through proxies managed by the remote infrastructure provider",
	Manifests:   []string{"testdata/tcproute.yaml"},
	Test: func(t *testing.T, s *suite.ConformanceTestSuite) {
		ns := ConformanceInfraNamespace
		routeNN := types.NamespacedName{Name: "tcp-app-1", Namespace: ns}
		gwNN := types.NamespacedName{Name: "my-tcp-gateway", Namespace: ns}
		gatewayRef := NewGatewayRef(gwNN)
		var gwAddr string
		t.Run("Validate Gateway status", func(t *testing.T) {
			gwAddr = GatewayAndTCPRoutesMustBeAccepted(
				t, s.Client, &s.TimeoutConfig, s.ControllerName,
				gatewayRef, routeNN,
			)
		})

		t.Run("Remotely managed data plane serves traffic", func(t *testing.T) {
			tlog.Logf(t, "sending traffic to %s via remote-infra-managed proxy", gwAddr)
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, s.RoundTripper, s.TimeoutConfig, gwAddr,
				http.ExpectedResponse{
					Request:   http.Request{Path: "/"},
					Response:  http.Response{StatusCodes: []int{200}},
					Namespace: ns,
				},
			)
		})

		t.Run("Validate remotely managed data plane gets cleaned up", func(t *testing.T) {
			toDelete := &gwapiv1.Gateway{
				TypeMeta: metav1.TypeMeta{
					Kind:       resource.KindGateway,
					APIVersion: gwapiv1.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      gwNN.Name,
					Namespace: gwNN.Namespace,
				},
			}
			err := s.Client.Delete(context.Background(), toDelete)
			require.NoError(t, err, "failed to delete gateway %+v", gwNN)

			gwResourceNS := GetGatewayResourceNamespace()
			gwLabelSelector := labels.SelectorFromSet(map[string]string{
				"gateway.envoyproxy.io/owning-gateway-name":      gwNN.Name,
				"gateway.envoyproxy.io/owning-gateway-namespace": gwNN.Namespace,
			})

			tlog.Logf(t, "waiting for service with owning-gateway labels to be deleted in namespace %s", gwResourceNS)
			require.NoError(t, wait.PollUntilContextTimeout(context.Background(), time.Second, s.TimeoutConfig.DeleteTimeout, true, func(ctx context.Context) (bool, error) {
				svcList := &corev1.ServiceList{}
				if err := s.Client.List(ctx, svcList, &client.ListOptions{
					Namespace:     gwResourceNS,
					LabelSelector: gwLabelSelector,
				}); err != nil {
					return false, err
				}
				return len(svcList.Items) == 0, nil
			}), "service for gateway %+v was not cleaned up", gwNN)

			tlog.Logf(t, "waiting for deployment with owning-gateway labels to be deleted in namespace %s", gwResourceNS)
			require.NoError(t, wait.PollUntilContextTimeout(context.Background(), time.Second, s.TimeoutConfig.DeleteTimeout, true, func(ctx context.Context) (bool, error) {
				deployList := &appsv1.DeploymentList{}
				if err := s.Client.List(ctx, deployList, &client.ListOptions{
					Namespace:     gwResourceNS,
					LabelSelector: gwLabelSelector,
				}); err != nil {
					return false, err
				}
				return len(deployList.Items) == 0, nil
			}), "deployment for gateway %+v was not cleaned up", gwNN)
		})
	},
}
