// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package runner

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	listenerv3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tsaarni/certyaml"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/crypto"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/extension/types"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/message"
	"github.com/envoyproxy/gateway/internal/xds/bootstrap"
)

func TestTLSConfig(t *testing.T) {
	// Create trusted CA, server and client certs.
	trustedCACert := certyaml.Certificate{
		Subject: "cn=trusted-ca",
	}
	egCertBeforeRotation := certyaml.Certificate{
		Subject:         "cn=eg-before-rotation",
		SubjectAltNames: []string{"DNS:localhost"},
		Issuer:          &trustedCACert,
	}
	egCertAfterRotation := certyaml.Certificate{
		Subject:         "cn=eg-after-rotation",
		SubjectAltNames: []string{"DNS:localhost"},
		Issuer:          &trustedCACert,
	}
	trustedEnvoyCert := certyaml.Certificate{
		Subject: "cn=trusted-envoy",
		Issuer:  &trustedCACert,
	}

	// Create another CA and a client cert to test that untrusted clients are denied.
	untrustedCACert := certyaml.Certificate{
		Subject: "cn=untrusted-ca",
	}
	untrustedClientCert := certyaml.Certificate{
		Subject: "cn=untrusted-client",
		Issuer:  &untrustedCACert,
	}

	caCertPool := x509.NewCertPool()
	ca, err := trustedCACert.X509Certificate()
	require.NoError(t, err)
	caCertPool.AddCert(&ca)

	tests := map[string]struct {
		serverCredentials *certyaml.Certificate
		clientCredentials *certyaml.Certificate
		expectError       bool
	}{
		"successful TLS connection established": {
			serverCredentials: &egCertBeforeRotation,
			clientCredentials: &trustedEnvoyCert,
			expectError:       false,
		},
		"rotating server credentials returns new server cert": {
			serverCredentials: &egCertAfterRotation,
			clientCredentials: &trustedEnvoyCert,
			expectError:       false,
		},
		"rotating server credentials again to ensure rotation can be repeated": {
			serverCredentials: &egCertBeforeRotation,
			clientCredentials: &trustedEnvoyCert,
			expectError:       false,
		},
		"fail to connect with client certificate which is not signed by correct CA": {
			serverCredentials: &egCertBeforeRotation,
			clientCredentials: &untrustedClientCert,
			expectError:       true,
		},
	}

	// Create temporary directory to store certificates and key for the server.
	configDir, err := os.MkdirTemp("", "eg-testdata-")
	require.NoError(t, err)
	defer os.RemoveAll(configDir)

	caFile := filepath.Join(configDir, "ca.crt")
	certFile := filepath.Join(configDir, "tls.crt")
	keyFile := filepath.Join(configDir, "tls.key")

	// Initial set of credentials must be written into temp directory before
	// starting the tests to avoid error at server startup.
	err = trustedCACert.WritePEM(caFile, keyFile)
	require.NoError(t, err)
	err = egCertBeforeRotation.WritePEM(certFile, keyFile)
	require.NoError(t, err)

	// Start a dummy server.
	tlsCfg, err := crypto.LoadTLSConfig(certFile, keyFile, caFile)
	require.NoError(t, err)

	g := grpc.NewServer(grpc.Creds(credentials.NewTLS(tlsCfg)))
	if g == nil {
		t.Error("failed to create server")
	}

	address := "localhost:8001"
	l, err := net.Listen("tcp", address)
	require.NoError(t, err)

	go func() {
		err := g.Serve(l)
		require.NoError(t, err)
	}()
	defer g.GracefulStop()

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// Store certificate and key to temp dir used by serveContext.
			err = tc.serverCredentials.WritePEM(certFile, keyFile)
			require.NoError(t, err)
			clientCert, _ := tc.clientCredentials.TLSCertificate()
			receivedCert, err := tryConnect(address, &clientCert, caCertPool)
			gotError := err != nil
			if gotError != tc.expectError {
				t.Errorf("Unexpected result when connecting to the server: %s", err)
			}
			if err == nil {
				expectedCert, _ := tc.serverCredentials.X509Certificate()
				assert.Equal(t, &expectedCert, receivedCert)
			}
		})
	}
}

// tryConnect tries to establish TLS connection to the server.
// If successful, return the server certificate.
func tryConnect(address string, clientCert *tls.Certificate, caCertPool *x509.CertPool) (*x509.Certificate, error) {
	clientConfig := &tls.Config{
		ServerName:   "localhost",
		MinVersion:   tls.VersionTLS13,
		Certificates: []tls.Certificate{*clientCert},
		NextProtos:   []string{"h2"},
		RootCAs:      caCertPool,
	}
	conn, err := tls.Dial("tcp", address, clientConfig)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	err = peekError(conn)
	if err != nil {
		return nil, err
	}

	return conn.ConnectionState().PeerCertificates[0], nil
}

// peekError is a workaround for TLS 1.3: due to shortened handshake, TLS alert
// from server is received at first read from the socket. To receive alert for
// bad certificate, this function tries to read one byte.
// Adapted from https://golang.org/src/crypto/tls/handshake_client_test.go
func peekError(conn net.Conn) error {
	_ = conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
	_, err := conn.Read(make([]byte, 1))
	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil
		}

		var netErr net.Error
		if !errors.As(netErr, &netErr) || !netErr.Timeout() {
			return err
		}
	}
	return nil
}

// setupTLSCerts creates temporary TLS certificates for testing
func setupTLSCerts(t *testing.T) (caFile, certFile, keyFile string, cleanup func()) {
	configDir, err := os.MkdirTemp("", "eg-runner-test-")
	require.NoError(t, err)

	caFile = filepath.Join(configDir, "ca.crt")
	certFile = filepath.Join(configDir, "tls.crt")
	keyFile = filepath.Join(configDir, "tls.key")

	// Create certificates
	trustedCACert := certyaml.Certificate{
		Subject: "cn=test-ca",
	}
	serverCert := certyaml.Certificate{
		Subject:         "cn=test-server",
		SubjectAltNames: []string{"DNS:localhost"},
		Issuer:          &trustedCACert,
	}

	err = trustedCACert.WritePEM(caFile, keyFile)
	require.NoError(t, err)
	err = serverCert.WritePEM(certFile, keyFile)
	require.NoError(t, err)

	return caFile, certFile, keyFile, func() {
		os.RemoveAll(configDir)
	}
}

func TestServeXdsServerListenFailed(t *testing.T) {
	// Occupy the address to make listening failed
	addr := net.JoinHostPort(XdsServerAddress, strconv.Itoa(bootstrap.DefaultXdsServerPort))
	l, err := net.Listen("tcp", addr)
	require.NoError(t, err)
	defer l.Close()

	cfg, _ := config.New(os.Stdout)
	r := New(&Config{
		Server: *cfg,
	})
	r.Logger = r.Logger.WithName(r.Name()).WithValues("runner", r.Name())
	// Don't crash in this function
	r.serveXdsServer(context.Background())
}

func TestRunner(t *testing.T) {
	// Setup TLS certificates
	caFile, certFile, keyFile, cleanup := setupTLSCerts(t)
	defer cleanup()

	// Setup
	xdsIR := new(message.XdsIR)
	pResource := new(message.ProviderResources)
	cfg, err := config.New(os.Stdout)
	require.NoError(t, err)
	r := New(&Config{
		Server:            *cfg,
		ProviderResources: pResource,
		XdsIR:             xdsIR,
		TLSCertPath:       certFile,
		TLSKeyPath:        keyFile,
		TLSCaPath:         caFile,
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start
	err = r.Start(ctx)
	require.NoError(t, err)
	defer func() {
		cancel()
		time.Sleep(100 * time.Millisecond) // Allow graceful shutdown
	}()

	// xDS is nil at start
	require.Equal(t, map[string]*ir.Xds{}, xdsIR.LoadAll())

	// test translation
	path := "example"
	res := ir.Xds{
		HTTP: []*ir.HTTPListener{
			{
				CoreListenerDetails: ir.CoreListenerDetails{
					Name:    "test",
					Address: "0.0.0.0",
					Port:    80,
				},
				Hostnames: []string{"example.com"},
				Routes: []*ir.HTTPRoute{
					{
						Name: "test-route",
						PathMatch: &ir.StringMatch{
							Exact: &path,
						},
						Destination: &ir.RouteDestination{
							Name: "test-dest",
							Settings: []*ir.DestinationSetting{
								{
									Endpoints: []*ir.DestinationEndpoint{
										{
											Host: "10.11.12.13",
											Port: 8080,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	xdsIR.Store("test", &res)
	require.Eventually(t, func() bool {
		// Check that the cache has the snapshot for our test key
		return r.cache.SnapshotHasIrKey("test")
	}, time.Second*5, time.Millisecond*50)

	// Delete the IR triggering an xds delete
	xdsIR.Delete("test")
	require.Eventually(t, func() bool {
		// Wait for the IR to be empty after deletion
		return len(xdsIR.LoadAll()) == 0
	}, time.Second*5, time.Millisecond*50)
}

func TestRunner_withExtensionManager_FailOpen(t *testing.T) {
	// Setup TLS certificates
	caFile, certFile, keyFile, cleanup := setupTLSCerts(t)
	defer cleanup()

	// Setup
	xdsIR := new(message.XdsIR)
	pResource := new(message.ProviderResources)

	cfg, err := config.New(os.Stdout)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	extMgr := &extManagerMock{}
	extMgr.ShouldFailOpen = true

	r := New(&Config{
		Server:            *cfg,
		ProviderResources: pResource,
		XdsIR:             xdsIR,
		ExtensionManager:  extMgr,
		TLSCertPath:       certFile,
		TLSKeyPath:        keyFile,
		TLSCaPath:         caFile,
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start
	err = r.Start(ctx)
	require.NoError(t, err)
	defer func() {
		cancel()
		time.Sleep(100 * time.Millisecond) // Allow graceful shutdown
	}()

	// xDS is nil at start
	require.Equal(t, map[string]*ir.Xds{}, xdsIR.LoadAll())

	// test translation
	path := "example"
	res := ir.Xds{
		HTTP: []*ir.HTTPListener{
			{
				CoreListenerDetails: ir.CoreListenerDetails{
					Name:    "test",
					Address: "0.0.0.0",
					Port:    80,
				},
				Hostnames: []string{"example.com"},
				Routes: []*ir.HTTPRoute{
					{
						Name: "test-route",
						PathMatch: &ir.StringMatch{
							Exact: &path,
						},
						Destination: &ir.RouteDestination{
							Name: "test-dest",
							Settings: []*ir.DestinationSetting{
								{
									Endpoints: []*ir.DestinationEndpoint{
										{
											Host: "10.11.12.13",
											Port: 8080,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	xdsIR.Store("test", &res)
	require.Eventually(t, func() bool {
		// Since the extension manager is configured to fail open, in an event of an error
		// from the extension manager hooks, xds update should be published.
		return r.cache.SnapshotHasIrKey("test")
	}, time.Second*5, time.Millisecond*50)
}

func TestRunner_withExtensionManager_FailClosed(t *testing.T) {
	// Setup TLS certificates
	caFile, certFile, keyFile, cleanup := setupTLSCerts(t)
	defer cleanup()

	// Setup
	xdsIR := new(message.XdsIR)
	pResource := new(message.ProviderResources)

	cfg, err := config.New(os.Stdout)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	extMgr := &extManagerMock{}

	r := New(&Config{
		Server:            *cfg,
		ProviderResources: pResource,
		XdsIR:             xdsIR,
		ExtensionManager:  extMgr,
		TLSCertPath:       certFile,
		TLSKeyPath:        keyFile,
		TLSCaPath:         caFile,
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start
	err = r.Start(ctx)
	require.NoError(t, err)
	defer func() {
		cancel()
		time.Sleep(100 * time.Millisecond) // Allow graceful shutdown
	}()

	// xDS is nil at start
	require.Equal(t, map[string]*ir.Xds{}, xdsIR.LoadAll())

	// test translation
	path := "example"
	res := ir.Xds{
		HTTP: []*ir.HTTPListener{
			{
				CoreListenerDetails: ir.CoreListenerDetails{
					Name:    "test",
					Address: "0.0.0.0",
					Port:    80,
				},
				Hostnames: []string{"example.com"},
				Routes: []*ir.HTTPRoute{
					{
						Name: "test-route",
						PathMatch: &ir.StringMatch{
							Exact: &path,
						},
						Destination: &ir.RouteDestination{
							Name: "test-dest",
							Settings: []*ir.DestinationSetting{
								{
									Endpoints: []*ir.DestinationEndpoint{
										{
											Host: "10.11.12.13",
											Port: 8080,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	xdsIR.Store("test", &res)
	require.Never(t, func() bool {
		// Since the extension manager is configured to fail closed, in an event of an error
		// from the extension manager hooks, xds update should not be published.
		return r.cache.SnapshotHasIrKey("test")
	}, time.Second*5, time.Millisecond*50)
}

type extManagerMock struct {
	types.Manager
	ShouldFailOpen bool
}

func (m *extManagerMock) GetPostXDSHookClient(xdsHookType egv1a1.XDSTranslatorHook) (types.XDSHookClient, error) {
	if xdsHookType == egv1a1.XDSHTTPListener {
		return &xdsHookClientMock{}, nil
	}

	return nil, nil
}

func (m *extManagerMock) FailOpen() bool {
	return m.ShouldFailOpen
}

type xdsHookClientMock struct {
	types.XDSHookClient
}

func (c *xdsHookClientMock) PostHTTPListenerModifyHook(*listenerv3.Listener, []*unstructured.Unstructured) (*listenerv3.Listener, error) {
	return nil, fmt.Errorf("assuming a network error during the call")
}

func TestGetRandomMaxConnectionAge(t *testing.T) {
	// Counter to track how many times each value is returned
	counts := make(map[time.Duration]int)

	// Call the function 100 times
	for i := 0; i < 100; i++ {
		value := getRandomMaxConnectionAge()

		// Verify the value is one of the expected values from maxConnectionAgeValues
		found := false
		for _, expected := range maxConnectionAgeValues {
			if value == expected {
				found = true
				break
			}
		}
		assert.True(t, found, "Unexpected value returned: %v", value)

		// Track counts
		counts[value]++
	}

	// Verify each known duration gets called at least 10 times
	for _, expectedValue := range maxConnectionAgeValues {
		count := counts[expectedValue]
		assert.GreaterOrEqual(t, count, 10, "Expected value %v to be called at least 10 times, got %d", expectedValue, count)
	}

	// Verify we got different values (randomness check)
	assert.Len(t, counts, len(maxConnectionAgeValues), "Should see all possible values")
}
