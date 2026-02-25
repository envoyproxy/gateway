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
	"sigs.k8s.io/gateway-api/conformance/echo-basic/grpcechoserver"
	"sigs.k8s.io/gateway-api/conformance/utils/config"
	"sigs.k8s.io/gateway-api/conformance/utils/grpc"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/conformance/utils/tlog"
	tlsutils "sigs.k8s.io/gateway-api/conformance/utils/tls"

	gatewayapi "github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
)

func init() {
	// TODO: move all ListenerSet tests to upstream
	ConformanceTests = append(ConformanceTests,
		ListenerSetHTTPTest,
		ListenerSetHTTPSTest,
		ListenerSetGRPCTest,
		ListenerSetTCPTest,
		ListenerSetUDPTest,
		ListenerSetTLSPassthroughTest,
		ListenerSetTLSTerminationTest)
}

// getListenerAddr extracts the host from a gateway address and joins it with a port
func getListenerAddr(gwAddrWithPort, port string) string {
	hostOnly := gwAddrWithPort
	if host, _, splitErr := net.SplitHostPort(gwAddrWithPort); splitErr == nil {
		hostOnly = host
	}
	return net.JoinHostPort(hostOnly, port)
}

// createListenerSetParent creates a RouteParentStatus for a ListenerSet
func createListenerSetParent(controllerName, listenerSetName, sectionName string) gwapiv1.RouteParentStatus {
	return gwapiv1.RouteParentStatus{
		ParentRef: gwapiv1.ParentReference{
			Group:       gatewayapi.GroupPtr(gwapiv1.GroupVersion.Group),
			Kind:        gatewayapi.KindPtr(resource.KindListenerSet),
			Name:        gwapiv1.ObjectName(listenerSetName),
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

var ListenerSetHTTPTest = suite.ConformanceTest{
	ShortName:   "ListenerSetHTTP",
	Description: "HTTPRoute should attach to a ListenerSet listener and serve traffic",
	Manifests: []string{
		"testdata/listenerset-base.yaml",
		"testdata/listenerset-http.yaml",
	},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		gwNN := types.NamespacedName{Name: "listener-gateway", Namespace: ns}
		routeNN := types.NamespacedName{Name: "listener-httproute", Namespace: ns}

		gwAddrWithPort, err := kubernetes.WaitForGatewayAddress(t, suite.Client, suite.TimeoutConfig, kubernetes.NewGatewayRef(gwNN, "core"))
		require.NoError(t, err)

		parents := []gwapiv1.RouteParentStatus{
			createListenerSetParent(suite.ControllerName, "listener-set-http", "extra-http"),
		}

		kubernetes.RouteMustHaveParents(t, suite.Client, suite.TimeoutConfig, routeNN, parents, false, &gwapiv1.HTTPRoute{})

		listenerAddr := getListenerAddr(gwAddrWithPort, "18081")
		expected := http.ExpectedResponse{
			Request: http.Request{
				Path: "/listener",
			},
			Response: http.Response{
				StatusCodes: []int{200},
			},
			Namespace: ns,
		}

		http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, listenerAddr, expected)
	},
}

var ListenerSetHTTPSTest = suite.ConformanceTest{
	ShortName:   "ListenerSetHTTPS",
	Description: "HTTPRoute should attach to an HTTPS ListenerSet listener and serve traffic",
	Manifests: []string{
		"testdata/listenerset-base.yaml",
		"testdata/listenerset-https.yaml",
	},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		gwNN := types.NamespacedName{Name: "listener-gateway", Namespace: ns}
		routeNN := types.NamespacedName{Name: "listener-httpsroute", Namespace: ns}

		gwAddrWithPort, err := kubernetes.WaitForGatewayAddress(t, suite.Client, suite.TimeoutConfig, kubernetes.NewGatewayRef(gwNN, "core"))
		require.NoError(t, err)

		listenerAddr := getListenerAddr(gwAddrWithPort, "18443")
		parents := []gwapiv1.RouteParentStatus{
			createListenerSetParent(suite.ControllerName, "listener-set-http", "extra-https"),
		}

		kubernetes.RouteMustHaveParents(t, suite.Client, suite.TimeoutConfig, routeNN, parents, false, &gwapiv1.HTTPRoute{})

		expected := http.ExpectedResponse{
			Request: http.Request{
				Host: "www.example.com",
				Path: "/listener-https",
			},
			Response: http.Response{
				StatusCodes: []int{200},
			},
			Namespace: ns,
		}

		certNN := types.NamespacedName{Name: "listener-https-certificate", Namespace: ns}
		serverCertificate, _, _, err := GetTLSSecret(suite.Client, certNN)
		require.NoError(t, err)

		tlsutils.MakeTLSRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig,
			listenerAddr, serverCertificate, nil, nil, "www.example.com", expected)
	},
}

var ListenerSetGRPCTest = suite.ConformanceTest{
	ShortName:   "ListenerSetGRPC",
	Description: "GRPCRoute should attach to a ListenerSet listener and serve traffic",
	Manifests: []string{
		"testdata/listenerset-base.yaml",
		"testdata/listenerset-grpc.yaml",
	},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		gwNN := types.NamespacedName{Name: "listener-gateway", Namespace: ns}
		routeNN := types.NamespacedName{Name: "listener-grpcroute", Namespace: ns}

		gwAddrWithPort, err := kubernetes.WaitForGatewayAddress(t, suite.Client, suite.TimeoutConfig, kubernetes.NewGatewayRef(gwNN, "core"))
		require.NoError(t, err)

		listenerAddr := getListenerAddr(gwAddrWithPort, "18082")
		parents := []gwapiv1.RouteParentStatus{
			createListenerSetParent(suite.ControllerName, "listener-set-grpc", "extra-grpc"),
		}

		kubernetes.RouteMustHaveParents(t, suite.Client, suite.TimeoutConfig, routeNN, parents, false, &gwapiv1.GRPCRoute{})

		expected := grpc.ExpectedResponse{
			EchoRequest: &grpcechoserver.EchoRequest{},
			Backend:     "grpc-listener-backend",
			Namespace:   ns,
		}

		grpc.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.GRPCClient, suite.TimeoutConfig, listenerAddr, expected)
	},
}

var ListenerSetTCPTest = suite.ConformanceTest{
	ShortName:   "ListenerSetTCP",
	Description: "TCPRoute should attach to a ListenerSet listener and serve traffic",
	Manifests: []string{
		"testdata/listenerset-base.yaml",
		"testdata/listenerset-tcp.yaml",
	},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		gwNN := types.NamespacedName{Name: "listener-gateway", Namespace: ns}
		routeNN := types.NamespacedName{Name: "listener-tcproute", Namespace: ns}

		gwAddrWithPort, err := kubernetes.WaitForGatewayAddress(t, suite.Client, suite.TimeoutConfig, kubernetes.NewGatewayRef(gwNN, "core"))
		require.NoError(t, err)

		listenerAddr := getListenerAddr(gwAddrWithPort, "18083")
		parents := []gwapiv1.RouteParentStatus{
			createListenerSetParent(suite.ControllerName, "listener-set-tcp", "extra-tcp"),
		}

		TCPRouteMustHaveParents(t, suite.Client, &suite.TimeoutConfig, routeNN, parents, false)

		expected := http.ExpectedResponse{
			Request: http.Request{
				Path: "/listener",
			},
			Response: http.Response{
				StatusCodes: []int{200},
			},
			Namespace: ns,
		}

		http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, listenerAddr, expected)
	},
}

var ListenerSetUDPTest = suite.ConformanceTest{
	ShortName:   "ListenerSetUDP",
	Description: "UDPRoute should attach to a ListenerSet listener and serve traffic",
	Manifests: []string{
		"testdata/listenerset-base.yaml",
		"testdata/listenerset-udp.yaml",
	},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		gwNN := types.NamespacedName{Name: "listener-gateway", Namespace: ns}
		routeNN := types.NamespacedName{Name: "listener-udproute", Namespace: ns}

		gwAddrWithPort, err := kubernetes.WaitForGatewayAddress(t, suite.Client, suite.TimeoutConfig, kubernetes.NewGatewayRef(gwNN, "core"))
		require.NoError(t, err)

		listenerAddr := getListenerAddr(gwAddrWithPort, "5300")
		parents := []gwapiv1.RouteParentStatus{
			createListenerSetParent(suite.ControllerName, "listener-set-udp", "extra-udp"),
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

var ListenerSetTLSPassthroughTest = suite.ConformanceTest{
	ShortName:   "ListenerSetTLSPassthrough",
	Description: "TLSRoute should attach to a ListenerSet TLS passthrough listener and serve traffic",
	Manifests: []string{
		"testdata/listenerset-base.yaml",
		"testdata/listenerset-tls-passthrough.yaml",
	},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		gwNN := types.NamespacedName{Name: "listener-gateway", Namespace: ns}
		routeNN := types.NamespacedName{Name: "listener-tlsroute-passthrough", Namespace: ns}

		gwAddrWithPort, err := kubernetes.WaitForGatewayAddress(t, suite.Client, suite.TimeoutConfig, kubernetes.NewGatewayRef(gwNN, "core"))
		require.NoError(t, err)

		listenerAddr := getListenerAddr(gwAddrWithPort, "18444")
		parents := []gwapiv1.RouteParentStatus{
			createListenerSetParent(suite.ControllerName, "listener-set-tls-passthrough", "extra-tls"),
		}

		TLSRouteMustHaveParents(t, suite.Client, &suite.TimeoutConfig, routeNN, parents)

		expected := http.ExpectedResponse{
			Request: http.Request{
				Host: "example.com",
				Path: "/",
			},
			Response: http.Response{
				StatusCodes: []int{200},
			},
			Namespace: ns,
		}

		certNN := types.NamespacedName{Name: "backend-tls-certificate", Namespace: ns}
		serverCertificate, _, _, err := GetTLSSecret(suite.Client, certNN)
		require.NoError(t, err)

		tlsutils.MakeTLSRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig,
			listenerAddr, serverCertificate, nil, nil, "example.com", expected)
	},
}

var ListenerSetTLSTerminationTest = suite.ConformanceTest{
	ShortName:   "ListenerSetTLSTermination",
	Description: "HTTPRoute should attach to a ListenerSet TLS termination listener and serve traffic",
	Manifests: []string{
		"testdata/listenerset-base.yaml",
		"testdata/listenerset-tls-termination.yaml",
	},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		gwNN := types.NamespacedName{Name: "listener-gateway", Namespace: ns}
		routeNN := types.NamespacedName{Name: "listener-httproute-tls-termination", Namespace: ns}

		gwAddrWithPort, err := kubernetes.WaitForGatewayAddress(t, suite.Client, suite.TimeoutConfig, kubernetes.NewGatewayRef(gwNN, "core"))
		require.NoError(t, err)

		listenerAddr := getListenerAddr(gwAddrWithPort, "19443")
		parents := []gwapiv1.RouteParentStatus{
			createListenerSetParent(suite.ControllerName, "listener-set-tls-termination", "extra-https-tls-termination"),
		}

		kubernetes.RouteMustHaveParents(t, suite.Client, suite.TimeoutConfig, routeNN, parents, false, &gwapiv1.HTTPRoute{})

		expected := http.ExpectedResponse{
			Request: http.Request{
				Host: "www.example.com",
				Path: "/",
			},
			Response: http.Response{
				StatusCodes: []int{200},
			},
			Namespace: ns,
		}

		certNN := types.NamespacedName{Name: "listener-https-certificate", Namespace: "listenerset-tls-termination-secret"}
		serverCertificate, _, _, err := GetTLSSecret(suite.Client, certNN)
		require.NoError(t, err)

		tlsutils.MakeTLSRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig,
			listenerAddr, serverCertificate, nil, nil, "www.example.com", expected)
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
