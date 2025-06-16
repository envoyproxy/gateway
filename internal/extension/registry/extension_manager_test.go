// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package registry

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"math"
	"net"
	"os"
	"reflect"
	"sync"
	"testing"

	clusterv3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	v3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/utils/ptr"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/proto/extension"
)

func TestGetExtensionServerAddress(t *testing.T) {
	tests := []struct {
		Name     string
		Service  *egv1a1.ExtensionService
		Expected string
	}{
		{
			Name: "has an FQDN",
			Service: &egv1a1.ExtensionService{
				BackendEndpoint: egv1a1.BackendEndpoint{
					FQDN: &egv1a1.FQDNEndpoint{
						Hostname: "extserver.svc.cluster.local",
						Port:     5050,
					},
				},
			},
			Expected: "extserver.svc.cluster.local:5050",
		},
		{
			Name: "has an IP",
			Service: &egv1a1.ExtensionService{
				BackendEndpoint: egv1a1.BackendEndpoint{
					IP: &egv1a1.IPEndpoint{
						Address: "10.10.10.10",
						Port:    5050,
					},
				},
			},
			Expected: "10.10.10.10:5050",
		},
		{
			Name: "has a Unix path",
			Service: &egv1a1.ExtensionService{
				BackendEndpoint: egv1a1.BackendEndpoint{
					Unix: &egv1a1.UnixSocket{
						Path: "/some/path",
					},
				},
			},
			Expected: "unix:///some/path",
		},
		{
			Name: "has a Unix path",
			Service: &egv1a1.ExtensionService{
				Host: "foo.bar",
				Port: 5050,
			},
			Expected: "foo.bar:5050",
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			out := getExtensionServerAddress(tc.Service)
			require.Equal(t, tc.Expected, out)
		})
	}
}

func Test_setupGRPCOpts(t *testing.T) {
	type args struct {
		ext *egv1a1.ExtensionManager
	}
	tests := []struct {
		name    string
		args    args
		want    []grpc.DialOption
		wantErr bool
	}{
		{
			args: args{
				ext: &egv1a1.ExtensionManager{
					MaxMessageSize: ptr.To(resource.MustParse(fmt.Sprintf("%dM", math.MaxInt))),
					Service: &egv1a1.ExtensionService{
						BackendEndpoint: egv1a1.BackendEndpoint{
							FQDN: &egv1a1.FQDNEndpoint{
								Hostname: "foo.bar",
								Port:     44344,
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			args: args{
				ext: &egv1a1.ExtensionManager{
					MaxMessageSize: ptr.To(resource.MustParse(fmt.Sprintf("%dM", 0))),
					Service: &egv1a1.ExtensionService{
						BackendEndpoint: egv1a1.BackendEndpoint{
							FQDN: &egv1a1.FQDNEndpoint{
								Hostname: "foo.bar",
								Port:     44344,
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			args: args{
				ext: &egv1a1.ExtensionManager{
					MaxMessageSize: ptr.To(resource.MustParse(fmt.Sprintf("%dM", 10))),
					Service: &egv1a1.ExtensionService{
						BackendEndpoint: egv1a1.BackendEndpoint{
							FQDN: &egv1a1.FQDNEndpoint{
								Hostname: "foo.bar",
								Port:     44344,
							},
						},
						Retry: &egv1a1.ExtensionServiceRetry{
							MaxAttempts:    ptr.To(20),
							InitialBackoff: ptr.To(gwapiv1.Duration("500ms")),
							MaxBackoff:     ptr.To(gwapiv1.Duration("5s")),
							BackoffMultiplier: &gwapiv1.Fraction{
								Numerator: 50,
							},
							RetryableStatusCodes: []egv1a1.RetryableGRPCStatusCode{
								"CANCELLED",
								"UNKNOWN",
								"INVALID_ARGUMENT",
								"DEADLINE_EXCEEDED",
								"NOT_FOUND",
								"ALREADY_EXISTS",
								"PERMISSION_DENIED",
								"RESOURCE_EXHAUSTED",
								"FAILED_PRECONDITION",
								"ABORTED",
								"OUT_OF_RANGE",
								"UNIMPLEMENTED",
								"INTERNAL",
								"UNAVAILABLE",
								"DATA_LOSS",
								"UNAUTHENTICATED",
							},
						},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fc := fakeclient.NewClientBuilder().WithScheme(envoygateway.GetScheme()).WithObjects().Build()
			_, err := setupGRPCOpts(context.TODO(), fc, tt.args.ext, "envoy-gateway-system")
			if (err != nil) != tt.wantErr {
				t.Errorf("setupGRPCOpts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

type testServer struct {
	extension.UnimplementedEnvoyGatewayExtensionServer
}

func (s *testServer) PostRouteModify(ctx context.Context, req *extension.PostRouteModifyRequest) (*extension.PostRouteModifyResponse, error) {
	return &extension.PostRouteModifyResponse{
		Route: req.Route,
	}, nil
}

func Test_TLS(t *testing.T) {
	testDir := "testdata"
	caFile := testDir + "/ca.pem"
	certFile := testDir + "/cert.pem"
	keyFile := testDir + "/key.pem"

	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	require.NoError(t, err)

	caCert, err := os.ReadFile(caFile)
	require.NoError(t, err)
	caPool := x509.NewCertPool()
	ok := caPool.AppendCertsFromPEM(caCert)
	require.True(t, ok)

	lis, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)
	defer lis.Close()

	port := lis.Addr().(*net.TCPAddr).Port
	server := grpc.NewServer(grpc.Creds(credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.NoClientCert,
		MinVersion:   tls.VersionTLS12,
	})))
	extension.RegisterEnvoyGatewayExtensionServer(server, &testServer{})
	go func() {
		_ = server.Serve(lis)
		defer server.GracefulStop()
	}()

	extManager := &egv1a1.ExtensionManager{
		Service: &egv1a1.ExtensionService{
			BackendEndpoint: egv1a1.BackendEndpoint{
				IP: &egv1a1.IPEndpoint{
					Address: "localhost",
					Port:    int32(port),
				},
			},
			TLS: &egv1a1.ExtensionTLS{
				CertificateRef: gwapiv1.SecretObjectReference{
					Name:      "cert",
					Namespace: ptr.To(gwapiv1.Namespace("default")),
				},
			},
		},
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cert",
			Namespace: "default",
		},
		Type: corev1.SecretTypeTLS,
		Data: map[string][]byte{
			corev1.TLSCertKey: caCert,
		},
	}

	fakeClient := fakeclient.NewClientBuilder().WithScheme(envoygateway.GetScheme()).WithObjects(secret).Build()

	opts, err := setupGRPCOpts(context.Background(), fakeClient, extManager, "test-ns")
	require.NoError(t, err)
	require.NotEmpty(t, opts)

	conn, err := grpc.DialContext(context.Background(), fmt.Sprintf("localhost:%d", port),
		opts...,
	)
	require.NoError(t, err)
	defer conn.Close()

	client := extension.NewEnvoyGatewayExtensionClient(conn)
	require.NotNil(t, client)

	response, err := client.PostRouteModify(context.Background(), &extension.PostRouteModifyRequest{
		Route: &v3.Route{
			Name: "test-route",
		},
	})
	require.NoError(t, err)
	require.Equal(t, "test-route", response.Route.Name)
}

func Test_buildServiceConfig(t *testing.T) {
	type args struct {
		ext *egv1a1.ExtensionManager
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "default",
			args: args{
				ext: &egv1a1.ExtensionManager{
					Service: &egv1a1.ExtensionService{
						BackendEndpoint: egv1a1.BackendEndpoint{
							FQDN: &egv1a1.FQDNEndpoint{
								Hostname: "foo.bar",
								Port:     44344,
							},
						},
					},
				},
			},
			want: `{
"methodConfig": [{
	"name": [{"service": "envoygateway.extension.EnvoyGatewayExtension"}],
	"waitForReady": true,
	"retryPolicy": {
		"MaxAttempts": 4,
		"InitialBackoff": "0.100000s",
		"MaxBackoff": "1.000000s",
		"BackoffMultiplier": 2.000000,
		"RetryableStatusCodes": [ "UNAVAILABLE" ]
	}
}]}`,
		},
		{
			name: "valid",
			args: args{
				ext: &egv1a1.ExtensionManager{
					Service: &egv1a1.ExtensionService{
						BackendEndpoint: egv1a1.BackendEndpoint{
							FQDN: &egv1a1.FQDNEndpoint{
								Hostname: "foo.bar",
								Port:     44344,
							},
						},
						Retry: &egv1a1.ExtensionServiceRetry{
							MaxAttempts:    ptr.To(20),
							InitialBackoff: ptr.To(gwapiv1.Duration("500ms")),
							MaxBackoff:     ptr.To(gwapiv1.Duration("5s")),
							BackoffMultiplier: &gwapiv1.Fraction{
								Numerator: 50,
							},
							RetryableStatusCodes: []egv1a1.RetryableGRPCStatusCode{
								"CANCELLED",
								"UNKNOWN",
								"INVALID_ARGUMENT",
								"DEADLINE_EXCEEDED",
								"NOT_FOUND",
								"ALREADY_EXISTS",
								"PERMISSION_DENIED",
								"RESOURCE_EXHAUSTED",
								"FAILED_PRECONDITION",
								"ABORTED",
								"OUT_OF_RANGE",
								"UNIMPLEMENTED",
								"INTERNAL",
								"UNAVAILABLE",
								"DATA_LOSS",
								"UNAUTHENTICATED",
							},
						},
					},
				},
			},
			want: `{
"methodConfig": [{
	"name": [{"service": "envoygateway.extension.EnvoyGatewayExtension"}],
	"waitForReady": true,
	"retryPolicy": {
		"MaxAttempts": 20,
		"InitialBackoff": "0.500000s",
		"MaxBackoff": "5.000000s",
		"BackoffMultiplier": 0.500000,
		"RetryableStatusCodes": [ "CANCELLED","UNKNOWN","INVALID_ARGUMENT","DEADLINE_EXCEEDED","NOT_FOUND","ALREADY_EXISTS","PERMISSION_DENIED","RESOURCE_EXHAUSTED","FAILED_PRECONDITION","ABORTED","OUT_OF_RANGE","UNIMPLEMENTED","INTERNAL","UNAVAILABLE","DATA_LOSS","UNAUTHENTICATED" ]
	}
}]}`,
		},
		{
			name: "defaults",
			args: args{
				ext: &egv1a1.ExtensionManager{
					Service: &egv1a1.ExtensionService{
						BackendEndpoint: egv1a1.BackendEndpoint{
							FQDN: &egv1a1.FQDNEndpoint{
								Hostname: "foo.bar",
								Port:     44344,
							},
						},
					},
				},
			},
			want: `{
"methodConfig": [{
	"name": [{"service": "envoygateway.extension.EnvoyGatewayExtension"}],
	"waitForReady": true,
	"retryPolicy": {
		"MaxAttempts": 4,
		"InitialBackoff": "0.100000s",
		"MaxBackoff": "1.000000s",
		"BackoffMultiplier": 2.000000,
		"RetryableStatusCodes": [ "UNAVAILABLE" ]
	}
}]}`,
		},
		{
			name: "invalid-code",
			args: args{
				ext: &egv1a1.ExtensionManager{
					Service: &egv1a1.ExtensionService{
						BackendEndpoint: egv1a1.BackendEndpoint{
							FQDN: &egv1a1.FQDNEndpoint{
								Hostname: "foo.bar",
								Port:     44344,
							},
						},
						Retry: &egv1a1.ExtensionServiceRetry{
							RetryableStatusCodes: []egv1a1.RetryableGRPCStatusCode{
								"CANCELLED",
								"NOTACODE",
							},
						},
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildServiceConfig(tt.args.ext)
			if (err != nil) != tt.wantErr {
				t.Errorf("buildServiceConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("buildServiceConfig() got = %v, want %v", got, tt.want)
			}
		})
	}
}

type retryTestServer struct {
	extension.UnimplementedEnvoyGatewayExtensionServer
	attempts int
	mu       sync.Mutex
}

func (s *retryTestServer) PostRouteModify(ctx context.Context, req *extension.PostRouteModifyRequest) (*extension.PostRouteModifyResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.attempts++
	if s.attempts == 10 {
		return &extension.PostRouteModifyResponse{
			Route: req.Route,
		}, nil
	} else {
		return nil, status.Error(codes.Unavailable, "Service is currently unavailable")
	}
}

func Test_Integration_RetryPolicy_MaxAttempts(t *testing.T) {
	type args struct {
		retryPolicy *egv1a1.ExtensionServiceRetry
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "sufficient retries",
			args: args{
				retryPolicy: &egv1a1.ExtensionServiceRetry{
					MaxAttempts: ptr.To(10),
					RetryableStatusCodes: []egv1a1.RetryableGRPCStatusCode{
						"UNAVAILABLE",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "insufficient retries",
			args: args{
				retryPolicy: &egv1a1.ExtensionServiceRetry{
					MaxAttempts: ptr.To(5),
					RetryableStatusCodes: []egv1a1.RetryableGRPCStatusCode{
						"UNAVAILABLE",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "wrong retry code",
			args: args{
				retryPolicy: &egv1a1.ExtensionServiceRetry{
					MaxAttempts: ptr.To(5),
					RetryableStatusCodes: []egv1a1.RetryableGRPCStatusCode{
						"CANCELLED",
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			extManager := egv1a1.ExtensionManager{
				Hooks: &egv1a1.ExtensionHooks{
					XDSTranslator: &egv1a1.XDSTranslatorHooks{
						Post: []egv1a1.XDSTranslatorHook{
							egv1a1.XDSRoute,
						},
					},
				},
				Service: &egv1a1.ExtensionService{
					BackendEndpoint: egv1a1.BackendEndpoint{
						FQDN: &egv1a1.FQDNEndpoint{
							Hostname: "foo.bar",
							Port:     44344,
						},
					},
					Retry: tt.args.retryPolicy,
				},
			}

			mgr, _, err := NewInMemoryManager(extManager, &retryTestServer{})
			require.NoError(t, err)

			hook, err := mgr.GetPostXDSHookClient(egv1a1.XDSRoute)
			require.NoError(t, err)
			require.NotNil(t, hook)

			_, err = hook.PostRouteModifyHook(
				&v3.Route{
					Name: "test-route",
				}, nil, nil)

			if (err != nil) != tt.wantErr {
				t.Errorf("PostRouteModifyHook() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

type clusterUpdateTestServer struct {
	extension.UnimplementedEnvoyGatewayExtensionServer
}

func getTargetRefKind(obj *unstructured.Unstructured) (string, error) {
	targetRef, found, err := unstructured.NestedMap(obj.Object, "spec", "targetRef")
	if err != nil || !found {
		return "", errors.New("targetRef not found or error")
	}

	kind, ok := targetRef["kind"].(string)
	if !ok {
		return "", errors.New("kind is not a string or missing in targetRef")
	}

	return kind, nil
}

func (s *clusterUpdateTestServer) PostTranslateModify(ctx context.Context, req *extension.PostTranslateModifyRequest) (*extension.PostTranslateModifyResponse, error) {
	clusters := req.GetClusters()
	if clusters == nil {
		return &extension.PostTranslateModifyResponse{
			Clusters: clusters,
			Secrets:  req.GetSecrets(),
		}, errors.New("No clusters found")
	}

	if len(req.PostTranslateContext.ExtensionResources) == 0 {
		return &extension.PostTranslateModifyResponse{
			Clusters: clusters,
			Secrets:  req.GetSecrets(),
		}, errors.New("No policy found")
	}

	for _, extensionResourceBytes := range req.PostTranslateContext.ExtensionResources {
		extensionResource := unstructured.Unstructured{}
		if err := extensionResource.UnmarshalJSON(extensionResourceBytes.UnstructuredBytes); err != nil {
			return &extension.PostTranslateModifyResponse{
				Clusters: clusters,
				Secrets:  req.GetSecrets(),
			}, err
		}

		targetKind, err := getTargetRefKind(&extensionResource)
		if err != nil || extensionResource.GetObjectKind().GroupVersionKind().Kind != "ExampleExtPolicy" || targetKind != "Gateway" {
			return &extension.PostTranslateModifyResponse{
				Clusters: clusters,
				Secrets:  req.GetSecrets(),
			}, errors.New("No matching policy found")
		}
	}

	ret := &extension.PostTranslateModifyResponse{
		Clusters: clusters,
		Secrets:  req.GetSecrets(),
	}

	return ret, nil
}

func Test_Integration_ClusterUpdateExtensionServer(t *testing.T) {
	testCases := []struct {
		name              string
		extensionPolicies []*ir.UnstructuredRef
		errorExpected     bool
	}{
		{
			name: "valid extension policy with targetRef",
			extensionPolicies: []*ir.UnstructuredRef{
				{
					Object: &unstructured.Unstructured{
						Object: map[string]any{
							"apiVersion": "gateway.example.io/v1",
							"kind":       "ExampleExtPolicy",
							"metadata": map[string]any{
								"name":      "test",
								"namespace": "test",
							},
							"spec": map[string]any{
								"targetRef": map[string]any{
									"group": "gateway.networking.k8s.io",
									"kind":  "Gateway",
									"name":  "test",
								},
								"data": "some data",
							},
						},
					},
				},
			},
			errorExpected: false,
		},

		{
			name: "invalid extension policy - no target",
			extensionPolicies: []*ir.UnstructuredRef{
				{
					Object: &unstructured.Unstructured{
						Object: map[string]any{
							"apiVersion": "gateway.example.io/v1alpha1",
							"kind":       "ExampleExtPolicy",
							"metadata": map[string]any{
								"name":      "test",
								"namespace": "test",
							},
							"spec": map[string]any{
								"data": "some data",
							},
						},
					},
				},
			},
			errorExpected: true,
		},
		{
			name: "invalid extension policy - no spec",
			extensionPolicies: []*ir.UnstructuredRef{
				{
					Object: &unstructured.Unstructured{
						Object: map[string]any{
							"apiVersion": "gateway.example.io/v1alpha1",
							"kind":       "ExampleExtPolicy",
							"metadata": map[string]any{
								"name":      "test",
								"namespace": "test",
							},
						},
					},
				},
			},
			errorExpected: true,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			extManager := egv1a1.ExtensionManager{
				Hooks: &egv1a1.ExtensionHooks{
					XDSTranslator: &egv1a1.XDSTranslatorHooks{
						Post: []egv1a1.XDSTranslatorHook{
							egv1a1.XDSTranslation,
						},
					},
				},
				Service: &egv1a1.ExtensionService{
					BackendEndpoint: egv1a1.BackendEndpoint{
						FQDN: &egv1a1.FQDNEndpoint{
							Hostname: "example.foo",
							Port:     44344,
						},
					},
				},
			}

			mgr, _, err := NewInMemoryManager(extManager, &clusterUpdateTestServer{})
			require.NoError(t, err)

			hook, err := mgr.GetPostXDSHookClient(egv1a1.XDSTranslation)
			require.NoError(t, err)
			require.NotNil(t, hook)

			_, _, err = hook.PostTranslateModifyHook(
				[]*clusterv3.Cluster{
					{
						Name: "test-cluster",
					},
				}, nil, tt.extensionPolicies)

			if (err != nil) != tt.errorExpected {
				t.Errorf("PostRouteModifyHook() error = %v, errorExpected %v", err, tt.errorExpected)
				return
			}
		})
	}
}
