// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/miekg/dns"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	gwapixv1a1 "sigs.k8s.io/gateway-api/apisx/v1alpha1"
	"sigs.k8s.io/gateway-api/conformance/echo-basic/grpcechoserver"
	"sigs.k8s.io/gateway-api/conformance/utils/config"
	"sigs.k8s.io/gateway-api/conformance/utils/grpc"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/conformance/utils/tlog"

	gatewayapi "github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
)

func init() {
	ConformanceTests = append(ConformanceTests, XListenerSetHTTPTest, XListenerSetHTTPSTest,
		XListenerSetGRPCTest, XListenerSetTCPTest, XListenerSetUDPTest, XListenerSetTLSTest)
}

// getListenerAddr extracts the host from a gateway address and joins it with a port
func getListenerAddr(gwAddrWithPort, port string) string {
	hostOnly := gwAddrWithPort
	if host, _, splitErr := net.SplitHostPort(gwAddrWithPort); splitErr == nil {
		hostOnly = host
	}
	return net.JoinHostPort(hostOnly, port)
}

// createXListenerSetParent creates a RouteParentStatus for an XListenerSet
func createXListenerSetParent(controllerName, xlistenerSetName, sectionName string) gwapiv1.RouteParentStatus {
	return gwapiv1.RouteParentStatus{
		ParentRef: gwapiv1.ParentReference{
			Group:       gatewayapi.GroupPtr(gwapixv1a1.GroupVersion.Group),
			Kind:        gatewayapi.KindPtr(resource.KindXListenerSet),
			Name:        gwapiv1.ObjectName(xlistenerSetName),
			Namespace:   gatewayapi.NamespacePtr("gateway-conformance-infra"),
			SectionName: gatewayapi.SectionNamePtr(sectionName),
		},
		ControllerName: gwapiv1.GatewayController(controllerName),
		Conditions: []metav1.Condition{
			{
				Type:   string(gwapiv1.RouteConditionAccepted),
				Status: metav1.ConditionTrue,
				Reason: string(gwapiv1.RouteReasonAccepted),
			},
			{
				Type:   string(gwapiv1.RouteConditionResolvedRefs),
				Status: metav1.ConditionTrue,
				Reason: string(gwapiv1.RouteReasonResolvedRefs),
			},
		},
	}
}

var XListenerSetHTTPTest = suite.ConformanceTest{
	ShortName:   "XListenerSetHTTP",
	Description: "HTTPRoute should attach to an XListenerSet listener and serve traffic",
	Manifests: []string{
		"testdata/xlistenerset-base.yaml",
		"testdata/xlistenerset-http.yaml",
	},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		gwNN := types.NamespacedName{Name: "xlistener-gateway", Namespace: ns}
		routeNN := types.NamespacedName{Name: "xlistener-httproute", Namespace: ns}

		gwAddrWithPort, err := kubernetes.WaitForGatewayAddress(t, suite.Client, suite.TimeoutConfig, kubernetes.NewGatewayRef(gwNN, "core"))
		require.NoError(t, err)

		parents := []gwapiv1.RouteParentStatus{
			createXListenerSetParent(suite.ControllerName, "xlistener-set-http", "extra-http"),
		}

		kubernetes.RouteMustHaveParents(t, suite.Client, suite.TimeoutConfig, routeNN, parents, false, &gwapiv1.HTTPRoute{})

		listenerAddr := getListenerAddr(gwAddrWithPort, "18081")
		expected := http.ExpectedResponse{
			Request: http.Request{
				Path: "/xlistener",
			},
			Response: http.Response{
				StatusCodes: []int{200},
			},
			Namespace: ns,
		}

		http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, listenerAddr, expected)
	},
}

var XListenerSetHTTPSTest = suite.ConformanceTest{
	ShortName:   "XListenerSetHTTPS",
	Description: "HTTPRoute should attach to an HTTPS XListenerSet listener and serve traffic",
	Manifests: []string{
		"testdata/xlistenerset-base.yaml",
		"testdata/xlistenerset-https.yaml",
	},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		gwNN := types.NamespacedName{Name: "xlistener-gateway", Namespace: ns}
		routeNN := types.NamespacedName{Name: "xlistener-httpsroute", Namespace: ns}

		gwAddrWithPort, err := kubernetes.WaitForGatewayAddress(t, suite.Client, suite.TimeoutConfig, kubernetes.NewGatewayRef(gwNN, "core"))
		require.NoError(t, err)

		listenerAddr := getListenerAddr(gwAddrWithPort, "18443")
		parents := []gwapiv1.RouteParentStatus{
			createXListenerSetParent(suite.ControllerName, "xlistener-set-http", "extra-https"),
		}

		kubernetes.RouteMustHaveParents(t, suite.Client, suite.TimeoutConfig, routeNN, parents, false, &gwapiv1.HTTPRoute{})

		expected := http.ExpectedResponse{
			Request: http.Request{
				Host: "www.example.com",
				Path: "/xlistener-https",
			},
			Response: http.Response{
				StatusCodes: []int{200},
			},
			Namespace: ns,
		}

		req := http.MakeRequest(t, &expected, listenerAddr, "HTTPS", "https")

		certNN := types.NamespacedName{Name: "xlistener-https-certificate", Namespace: ns}
		cPem, keyPem, caPem, err := GetTLSSecret(suite.Client, certNN)
		require.NoError(t, err)

		combined := string(cPem)
		if len(caPem) > 0 {
			combined += "\n" + string(caPem)
		}

		WaitForConsistentMTLSResponse(t, suite.RoundTripper, &req, &expected, suite.TimeoutConfig.RequiredConsecutiveSuccesses, suite.TimeoutConfig.MaxTimeToConsistency,
			[]byte(combined), keyPem, "www.example.com")
	},
}

var XListenerSetGRPCTest = suite.ConformanceTest{
	ShortName:   "XListenerSetGRPC",
	Description: "GRPCRoute should attach to an XListenerSet listener and serve traffic",
	Manifests: []string{
		"testdata/xlistenerset-base.yaml",
		"testdata/xlistenerset-grpc.yaml",
	},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		gwNN := types.NamespacedName{Name: "xlistener-gateway", Namespace: ns}
		routeNN := types.NamespacedName{Name: "xlistener-grpcroute", Namespace: ns}

		gwAddrWithPort, err := kubernetes.WaitForGatewayAddress(t, suite.Client, suite.TimeoutConfig, kubernetes.NewGatewayRef(gwNN, "core"))
		require.NoError(t, err)

		listenerAddr := getListenerAddr(gwAddrWithPort, "18082")
		parents := []gwapiv1.RouteParentStatus{
			createXListenerSetParent(suite.ControllerName, "xlistener-set-grpc", "extra-grpc"),
		}

		kubernetes.RouteMustHaveParents(t, suite.Client, suite.TimeoutConfig, routeNN, parents, false, &gwapiv1.GRPCRoute{})

		expected := grpc.ExpectedResponse{
			EchoRequest: &grpcechoserver.EchoRequest{},
			Backend:     "grpc-xlistener-backend",
			Namespace:   ns,
		}

		grpc.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.GRPCClient, suite.TimeoutConfig, listenerAddr, expected)
	},
}

var XListenerSetTCPTest = suite.ConformanceTest{
	ShortName:   "XListenerSetTCP",
	Description: "TCPRoute should attach to an XListenerSet listener and serve traffic",
	Manifests: []string{
		"testdata/xlistenerset-base.yaml",
		"testdata/xlistenerset-tcproute.yaml",
	},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		gwNN := types.NamespacedName{Name: "xlistener-gateway", Namespace: ns}
		routeNN := types.NamespacedName{Name: "xlistener-tcproute", Namespace: ns}

		gwAddrWithPort, err := kubernetes.WaitForGatewayAddress(t, suite.Client, suite.TimeoutConfig, kubernetes.NewGatewayRef(gwNN, "core"))
		require.NoError(t, err)

		listenerAddr := getListenerAddr(gwAddrWithPort, "18083")
		parents := []gwapiv1.RouteParentStatus{
			createXListenerSetParent(suite.ControllerName, "xlistener-set-tcp", "extra-tcp"),
		}

		TCPRouteMustHaveParents(t, suite.Client, &suite.TimeoutConfig, routeNN, parents, false)

		expected := http.ExpectedResponse{
			Request: http.Request{
				Path: "/xlistener",
			},
			Response: http.Response{
				StatusCodes: []int{200},
			},
			Namespace: ns,
		}

		http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, listenerAddr, expected)
	},
}

var XListenerSetUDPTest = suite.ConformanceTest{
	ShortName:   "XListenerSetUDP",
	Description: "UDPRoute should attach to an XListenerSet listener and serve traffic",
	Manifests: []string{
		"testdata/xlistenerset-base.yaml",
		"testdata/xlistenerset-udp.yaml",
	},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		gwNN := types.NamespacedName{Name: "xlistener-gateway", Namespace: ns}
		routeNN := types.NamespacedName{Name: "xlistener-udproute", Namespace: ns}

		gwAddrWithPort, err := kubernetes.WaitForGatewayAddress(t, suite.Client, suite.TimeoutConfig, kubernetes.NewGatewayRef(gwNN, "core"))
		require.NoError(t, err)

		listenerAddr := getListenerAddr(gwAddrWithPort, "5300")
		parents := []gwapiv1.RouteParentStatus{
			createXListenerSetParent(suite.ControllerName, "xlistener-set-udp", "extra-udp"),
		}

		UDPRouteMustHaveParents(t, suite.Client, &suite.TimeoutConfig, routeNN, parents, false)

		domain := "foo.bar.com."
		msg := new(dns.Msg)
		msg.SetQuestion(domain, dns.TypeA)

		if err := wait.PollUntilContextTimeout(context.TODO(), time.Second, time.Minute, true,
			func(_ context.Context) (done bool, err error) {
				tlog.Logf(t, "performing DNS query %s on %s", domain, listenerAddr)
				r, err := dns.Exchange(msg, listenerAddr)
				if err != nil {
					tlog.Logf(t, "failed to perform a UDP query: %v", err)
					return false, nil
				}
				tlog.Logf(t, "got DNS response: %s", r.String())
				return true, nil
			}); err != nil {
			t.Errorf("failed to perform DNS query: %v", err)
		}
	},
}

var XListenerSetTLSTest = suite.ConformanceTest{
	ShortName:   "XListenerSetTLS",
	Description: "TLSRoute should attach to an XListenerSet TLS listener and serve traffic",
	Manifests: []string{
		"testdata/xlistenerset-base.yaml",
		"testdata/xlistenerset-tls.yaml",
	},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		gwNN := types.NamespacedName{Name: "xlistener-gateway", Namespace: ns}
		routeNN := types.NamespacedName{Name: "xlistener-tlsroute", Namespace: ns}

		gwAddrWithPort, err := kubernetes.WaitForGatewayAddress(t, suite.Client, suite.TimeoutConfig, kubernetes.NewGatewayRef(gwNN, "core"))
		require.NoError(t, err)

		listenerAddr := getListenerAddr(gwAddrWithPort, "18444")
		parents := []gwapiv1.RouteParentStatus{
			createXListenerSetParent(suite.ControllerName, "xlistener-set-tls", "extra-tls"),
		}

		TLSRouteMustHaveParents(t, suite.Client, &suite.TimeoutConfig, routeNN, parents)

		expected := http.ExpectedResponse{
			Request: http.Request{
				Host: "tls.example.com",
				Path: "/",
			},
			Response: http.Response{
				StatusCodes: []int{200},
			},
			Namespace: ns,
		}

		req := http.MakeRequest(t, &expected, listenerAddr, "HTTPS", "https")

		certNN := types.NamespacedName{Name: "backend-tls-certificate", Namespace: ns}
		cPem, keyPem, _, err := GetTLSSecret(suite.Client, certNN)
		require.NoError(t, err)

		WaitForConsistentMTLSResponse(
			t,
			suite.RoundTripper,
			&req,
			&expected,
			suite.TimeoutConfig.RequiredConsecutiveSuccesses,
			suite.TimeoutConfig.MaxTimeToConsistency,
			cPem,
			keyPem,
			"tls.example.com")
	},
}

// TLSRouteMustHaveParents waits for the TLSRoute to have parents matching the expected parents
func TLSRouteMustHaveParents(t *testing.T, client client.Client, timeoutConfig *config.TimeoutConfig, routeName types.NamespacedName, parents []gwapiv1.RouteParentStatus) {
	t.Helper()
	var actual []gwapiv1.RouteParentStatus
	waitErr := wait.PollUntilContextTimeout(context.Background(), 1*time.Second, timeoutConfig.RouteMustHaveParents, true, func(ctx context.Context) (bool, error) {
		route := &gwapiv1a2.TLSRoute{}
		err := client.Get(ctx, routeName, route)
		if err != nil {
			return false, fmt.Errorf("error fetching TLSRoute: %w", err)
		}

		actual = route.Status.Parents
		return parentsForRouteMatch(t, routeName, parents, actual, false), nil
	})
	require.NoErrorf(t, waitErr, "error waiting for TLSRoute to have parents matching expectations")
}
