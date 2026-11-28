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

	"github.com/quic-go/quic-go/http3"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

		for _, tc := range []struct {
			name, hostname, section, serverCertificateName string
			createPolicy, wildcard                         bool
		}{
			{name: "HTTPS", hostname: "www.awesome.org", section: "https-www-awesome", serverCertificateName: "www-awesome-org-tls", createPolicy: true},
			{name: "HTTPS wildcard", hostname: "admin.secure.awesome.org", section: "https-secure-awesome", serverCertificateName: "awesome-org-tls", createPolicy: true, wildcard: true},
			{name: "TLS termination", hostname: "secure.awesome.org", section: "tls-secure", serverCertificateName: "awesome-org-tls"},
		} {
			t.Run("QUIC wildcard rejects client-validated "+tc.name+" hostnames", func(t *testing.T) {
				policy := &egv1a1.ClientTrafficPolicy{
					ObjectMeta: metav1.ObjectMeta{Name: "http3-client-validation", Namespace: ConformanceInfraNamespace},
					Spec: egv1a1.ClientTrafficPolicySpec{
						PolicyTargetReferences: egv1a1.PolicyTargetReferences{TargetRefs: []gwapiv1.LocalPolicyTargetReferenceWithSectionName{{
							LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
								Group: gwapiv1.GroupName, Kind: "Gateway", Name: "http3-gtw",
							},
							SectionName: new(gwapiv1.SectionName(tc.section)),
						}}},
						HTTP3: &egv1a1.HTTP3Settings{},
						TLS: &egv1a1.ClientTLSSettings{ClientValidation: &egv1a1.ClientValidationContext{
							CACertificateRefs: []gwapiv1.SecretObjectReference{{Name: "www-awesome-org-tls"}},
						}},
					},
				}
				if tc.createPolicy {
					require.NoError(t, suite.Client.Create(t.Context(), policy))
					t.Cleanup(func() {
						ctx, cancel := context.WithTimeout(context.Background(), suite.TimeoutConfig.RequestTimeout)
						defer cancel()
						require.NoError(t, suite.Client.Delete(ctx, policy))
					})
				} else {
					policy.Name = "http3-tls-client-validation"
				}
				ClientTrafficPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: policy.Name, Namespace: policy.Namespace}, suite.ControllerName,
					gwapiv1.ParentReference{
						Group: ancestorRef.Group, Kind: ancestorRef.Kind, Name: ancestorRef.Name,
						Namespace: ancestorRef.Namespace, SectionName: new(gwapiv1.SectionName(tc.section)),
					})

				clientPEM, privateKey, _, err := GetTLSSecret(suite.Client, types.NamespacedName{Name: "www-awesome-org-tls", Namespace: ConformanceInfraNamespace})
				require.NoError(t, err)
				_, _, ca, err := GetTLSSecret(suite.Client, types.NamespacedName{Name: tc.serverCertificateName, Namespace: ConformanceInfraNamespace})
				require.NoError(t, err)
				roots := x509.NewCertPool()
				require.True(t, roots.AppendCertsFromPEM(ca))
				clientCertificate, err := tls.X509KeyPair(clientPEM, privateKey)
				require.NoError(t, err)
				tcpTransport := &stdhttp.Transport{TLSClientConfig: &tls.Config{
					RootCAs: roots, ServerName: tc.hostname, MinVersion: tls.VersionTLS12,
					Certificates: []tls.Certificate{clientCertificate},
				}}
				defer tcpTransport.CloseIdleConnections()
				// The fixture certificate covers the exact test hostnames. For the
				// nested wildcard, check QUIC via the public SNI below instead.
				if !tc.wildcard {
					// A successful authenticated TCP request without Alt-Svc proves that the
					// section's client-validation and HTTP/3 suppression reached Envoy.
					require.NoError(t, wait.PollUntilContextTimeout(t.Context(), time.Second, suite.TimeoutConfig.MaxTimeToConsistency, true, func(ctx context.Context) (bool, error) {
						ctx, cancel := context.WithTimeout(ctx, suite.TimeoutConfig.RequestTimeout)
						defer cancel()
						request, err := stdhttp.NewRequestWithContext(ctx, stdhttp.MethodGet, "https://"+gwAddr+"/", nil)
						if err != nil {
							return false, err
						}
						request.Host = tc.hostname
						response, err := tcpTransport.RoundTrip(request)
						if err != nil {
							return false, nil
						}
						defer response.Body.Close()
						return response.StatusCode == stdhttp.StatusOK && response.Header.Get("Alt-Svc") == "", nil
					}))
				}

				checkQUIC := func(transport *http3.Transport, authority string, status int) {
					t.Helper()
					require.NoError(t, wait.PollUntilContextTimeout(t.Context(), time.Second, suite.TimeoutConfig.MaxTimeToConsistency, true, func(ctx context.Context) (bool, error) {
						ctx, cancel := context.WithTimeout(ctx, suite.TimeoutConfig.RequestTimeout)
						defer cancel()
						request, err := stdhttp.NewRequestWithContext(ctx, stdhttp.MethodGet, "https://"+gwAddr+"/", nil)
						if err != nil {
							return false, err
						}
						request.Host = authority
						response, err := transport.RoundTrip(request)
						if err != nil {
							return false, nil
						}
						defer response.Body.Close()
						_, err = io.Copy(io.Discard, response.Body)
						return err == nil && response.ProtoMajor == 3 && response.StatusCode == status, nil
					}))
				}
				// Both QUIC SNIs select the wildcard chain because the protected
				// listener has no QUIC chain. Trust the wildcard's certificate.
				_, _, wildcardCA, err := GetTLSSecret(suite.Client, types.NamespacedName{Name: "awesome-org-tls", Namespace: ConformanceInfraNamespace})
				require.NoError(t, err)
				quicRoots := x509.NewCertPool()
				require.True(t, quicRoots.AppendCertsFromPEM(wildcardCA))
				serverNames := []string{tc.hostname, "api.awesome.org"}
				if tc.wildcard {
					serverNames = []string{"api.awesome.org"}
				}
				for _, sni := range serverNames {
					transport := &http3.Transport{TLSClientConfig: &tls.Config{
						RootCAs: quicRoots, ServerName: sni, MinVersion: tls.VersionTLS13,
					}}
					defer transport.Close()
					if sni == "api.awesome.org" {
						// Reuse the same QUIC transport/connection for a protected authority.
						checkQUIC(transport, "api.awesome.org", stdhttp.StatusOK)
					}
					authorities := []string{tc.hostname, tc.hostname + ":", tc.hostname + ":garbage"}
					if tc.wildcard {
						authorities = append(authorities, "x:"+tc.hostname+":443", "[x:admin].secure.awesome.org")
					}
					for _, authority := range authorities {
						checkQUIC(transport, authority, stdhttp.StatusForbidden)
					}
					if sni == tc.hostname {
						checkQUIC(transport, "api.awesome.org", stdhttp.StatusForbidden)
					} else {
						checkQUIC(transport, "api.awesome.org", stdhttp.StatusOK)
					}
				}
				testHTTP3(t, "bar.example.com", "bar-com-tls")
			})
		}

		t.Run("required Proxy Protocol and TLS fingerprints apply only to TCP", func(t *testing.T) {
			policyNN := types.NamespacedName{Name: "http3-ctp", Namespace: ConformanceInfraNamespace}
			require.NoError(t, wait.PollUntilContextTimeout(t.Context(), time.Second, suite.TimeoutConfig.MaxTimeToConsistency, true, func(ctx context.Context) (bool, error) {
				policy := &egv1a1.ClientTrafficPolicy{}
				if err := suite.Client.Get(ctx, policyNN, policy); err != nil {
					return false, err
				}
				policy.Spec.ProxyProtocol = &egv1a1.ProxyProtocolSettings{}
				policy.Spec.TLS = &egv1a1.ClientTLSSettings{TLSSettings: egv1a1.TLSSettings{
					Fingerprints: []egv1a1.TLSFingerprintType{egv1a1.TLSFingerprintTypeJA3, egv1a1.TLSFingerprintTypeJA4},
				}}
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
