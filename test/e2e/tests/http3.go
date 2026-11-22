// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"bufio"
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"net"
	stdhttp "net/http"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	tlsutils "sigs.k8s.io/gateway-api/conformance/utils/tls"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/test/e2e/utils"
)

func init() {
	ConformanceTests = append(ConformanceTests, HTTP3Test)
}

var HTTP3Test = suite.ConformanceTest{
	ShortName:   "HTTP3",
	Description: "HTTP3 tests ensure that Envoy Gateway supports HTTP/3 features.",
	Manifests:   []string{"testdata/http3.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		routeNN := types.NamespacedName{Name: "http3-route-foo", Namespace: ConformanceInfraNamespace}
		gwNN := types.NamespacedName{Name: "http3-gtw", Namespace: ConformanceInfraNamespace}
		gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName,
			kubernetes.NewGatewayRef(gwNN), routeNN)

		ancestorRef := gwapiv1.ParentReference{
			Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
			Kind:      gatewayapi.KindPtr(resource.KindGateway),
			Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
			Name:      gwapiv1.ObjectName(gwNN.Name),
		}
		ClientTrafficPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "http3-ctp", Namespace: ConformanceInfraNamespace},
			suite.ControllerName, ancestorRef)

		t.Run("advertises configured HTTP/3 port", func(t *testing.T) {
			expected := http.ExpectedResponse{
				Request: http.Request{
					Host: "foo.example.com",
					Path: "/",
				},
				Response: http.Response{
					StatusCodes: []int{200},
					Headers: map[string]string{
						"alt-svc": `h3=":8443"; ma=86400`,
					},
				},
				Namespace: ConformanceInfraNamespace,
			}

			serverCertificate, _, _, err := GetTLSSecret(suite.Client, types.NamespacedName{Name: "foo-com-tls", Namespace: ConformanceInfraNamespace})
			if err != nil {
				t.Fatalf("unexpected error finding TLS secret: %v", err)
			}

			tlsutils.MakeTLSRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig,
				gwAddr, serverCertificate, nil, nil, "foo.example.com", expected)
		})

		testHTTP3 := func(t *testing.T, host, secretName string) {
			quicRoundTripper := &utils.QuicRoundTripper{
				Debug:         suite.Debug,
				TimeoutConfig: suite.TimeoutConfig,
			}

			expected := http.ExpectedResponse{
				Request: http.Request{
					Host: host,
					Path: "/",
				},
				Response: http.Response{
					StatusCodes: []int{200},
				},
				Namespace: ConformanceInfraNamespace,
			}

			serverCertificate, _, _, err := GetTLSSecret(suite.Client, types.NamespacedName{Name: secretName, Namespace: ConformanceInfraNamespace})
			if err != nil {
				t.Fatalf("unexpected error finding TLS secret: %v", err)
			}

			tlsutils.MakeTLSRequestAndExpectEventuallyConsistentResponse(t, quicRoundTripper, suite.TimeoutConfig,
				gwAddr, serverCertificate, nil, nil, host, expected)
		}

		testHTTP3(t, "foo.example.com", "foo-com-tls")
		testHTTP3(t, "bar.example.com", "bar-com-tls")
		testHTTP3(t, "www.awesome.org", "www-awesome-org-tls") // test with a domain overlapped with a wildcard domain
		testHTTP3(t, "api.awesome.org", "awesome-org-tls")     // test with a wildcard domain

		t.Run("required Proxy Protocol applies only to TCP", func(t *testing.T) {
			policyNN := types.NamespacedName{Name: "http3-ctp", Namespace: ConformanceInfraNamespace}
			require.NoError(t, wait.PollUntilContextTimeout(t.Context(), time.Second, suite.TimeoutConfig.MaxTimeToConsistency, true, func(ctx context.Context) (bool, error) {
				policy := &egv1a1.ClientTrafficPolicy{}
				if err := suite.Client.Get(ctx, policyNN, policy); err != nil {
					return false, err
				}
				policy.Spec.ProxyProtocol = &egv1a1.ProxyProtocolSettings{}
				if err := suite.Client.Update(ctx, policy); err != nil {
					return false, nil
				}
				return true, nil
			}))

			serverCertificate, _, _, err := GetTLSSecret(suite.Client, types.NamespacedName{Name: "foo-com-tls", Namespace: ConformanceInfraNamespace})
			require.NoError(t, err)
			roots := x509.NewCertPool()
			require.True(t, roots.AppendCertsFromPEM(serverCertificate))
			tlsConfig := &tls.Config{
				RootCAs: roots, ServerName: "foo.example.com", MinVersion: tls.VersionTLS12,
			}
			// A successful request with a PROXY header proves that the filter has
			// reached the data plane and that TCP/TLS traffic still works.
			require.NoError(t, wait.PollUntilContextTimeout(t.Context(), time.Second, suite.TimeoutConfig.MaxTimeToConsistency, true, func(ctx context.Context) (bool, error) {
				ctx, cancel := context.WithTimeout(ctx, suite.TimeoutConfig.RequestTimeout)
				defer cancel()
				rawConn, err := (&net.Dialer{}).DialContext(ctx, "tcp", gwAddr)
				if err != nil {
					return false, nil
				}
				defer rawConn.Close()
				deadline, _ := ctx.Deadline()
				if err := rawConn.SetDeadline(deadline); err != nil {
					return false, err
				}
				local := rawConn.LocalAddr().(*net.TCPAddr)
				remote := rawConn.RemoteAddr().(*net.TCPAddr)
				family := "TCP4"
				if remote.IP.To4() == nil {
					family = "TCP6"
				}
				if _, err := fmt.Fprintf(rawConn, "PROXY %s %s %s %d %d\r\n", family, local.IP, remote.IP, local.Port, remote.Port); err != nil {
					return false, nil
				}
				conn := tls.Client(rawConn, tlsConfig)
				if err := conn.HandshakeContext(ctx); err != nil {
					return false, nil
				}
				request, err := stdhttp.NewRequestWithContext(ctx, stdhttp.MethodGet, "https://foo.example.com/", nil)
				if err != nil {
					return false, err
				}
				request.Close = true
				if err := request.Write(conn); err != nil {
					return false, nil
				}
				response, err := stdhttp.ReadResponse(bufio.NewReader(conn), request)
				if err != nil {
					return false, nil
				}
				defer response.Body.Close()
				return response.StatusCode == stdhttp.StatusOK && response.Header.Get("Alt-Svc") == `h3=":8443"; ma=86400`, nil
			}))

			// Only an explicit close/reset during the TLS handshake counts as
			// rejection. Dial failures, timeouts, and certificate errors do not.
			require.NoError(t, wait.PollUntilContextTimeout(t.Context(), time.Second, suite.TimeoutConfig.MaxTimeToConsistency, true, func(ctx context.Context) (bool, error) {
				ctx, cancel := context.WithTimeout(ctx, suite.TimeoutConfig.RequestTimeout)
				defer cancel()
				rawConn, err := (&net.Dialer{}).DialContext(ctx, "tcp", gwAddr)
				if err != nil {
					return false, nil
				}
				defer rawConn.Close()
				conn := tls.Client(rawConn, tlsConfig)
				err = conn.HandshakeContext(ctx)
				return errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) || errors.Is(err, syscall.ECONNRESET), nil
			}))

			testHTTP3(t, "foo.example.com", "foo-com-tls")
			testHTTP3(t, "www.awesome.org", "www-awesome-org-tls")
			testHTTP3(t, "api.awesome.org", "awesome-org-tls")
		})
	},
}
