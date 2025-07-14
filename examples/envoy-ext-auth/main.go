// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	envoy_api_v3_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_service_auth_v3 "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/protobuf/types/known/structpb"
)

var (
	port      int
	certPath  string
	testUsers = map[string]string{
		"token1": "user1",
		"token2": "user2",
		"token3": "user3",
	}
)

func main() {
	flag.IntVar(&port, "port", 9002, "gRPC port")
	flag.StringVar(&certPath, "certPath", "", "path to server certificate and private key")
	flag.Parse()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("failed to listen to %d: %v", port, err)
	}

	users := testUsers

	// Load TLS credentials
	creds, err := loadTLSCredentials(certPath)
	if err != nil {
		log.Fatalf("Failed to load TLS credentials: %v", err)
	}
	gs := grpc.NewServer(grpc.Creds(creds))

	envoy_service_auth_v3.RegisterAuthorizationServer(gs, NewAuthServer(users))

	log.Printf("starting gRPC server on: %d\n", port)

	go func() {
		err = gs.Serve(lis)
		if err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	http.HandleFunc("/healthz", healthCheckHandler)
	http.HandleFunc("/", authCheckerHandler)
	http.HandleFunc("/auth", authCheckerHandler)
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

type authServer struct {
	users Users
}

var _ envoy_service_auth_v3.AuthorizationServer = &authServer{}

// NewAuthServer creates a new authorization server.
func NewAuthServer(users Users) envoy_service_auth_v3.AuthorizationServer {
	return &authServer{users}
}

// Check implements authorization's Check interface which performs authorization check based on the
// attributes associated with the incoming request.
func (s *authServer) Check(
	_ context.Context,
	req *envoy_service_auth_v3.CheckRequest,
) (*envoy_service_auth_v3.CheckResponse, error) {
	authorization := req.Attributes.Request.Http.Headers["authorization"]
	log.Println("GRPC check auth: ", authorization)

	extracted := strings.Fields(authorization)
	if len(extracted) == 2 && extracted[0] == "Bearer" {
		valid, user := s.users.Check(extracted[1])
		if valid {
			routeMetadata, err := extractE2ETestMetadata(req)
			if err != nil {
				log.Printf("Warning: Could not extract e2e-conformance-metadata: %v", err)
				routeMetadata = "unknown"
			}

			return &envoy_service_auth_v3.CheckResponse{
				HttpResponse: &envoy_service_auth_v3.CheckResponse_OkResponse{
					OkResponse: &envoy_service_auth_v3.OkHttpResponse{
						Headers: []*envoy_api_v3_core.HeaderValueOption{
							{
								AppendAction: envoy_api_v3_core.HeaderValueOption_OVERWRITE_IF_EXISTS_OR_ADD,
								Header: &envoy_api_v3_core.HeaderValue{
									// For a successful request, the authorization server sets the
									// x-current-user value.
									Key:   "x-current-user",
									Value: user,
								},
							},
							{
								AppendAction: envoy_api_v3_core.HeaderValueOption_OVERWRITE_IF_EXISTS_OR_ADD,
								Header: &envoy_api_v3_core.HeaderValue{
									Key:   "x-e2e-conformance-metadata",
									Value: routeMetadata,
								},
							},
						},
					},
				},
				Status: &status.Status{
					Code: int32(code.Code_OK),
				},
			}, nil
		}
	}

	return &envoy_service_auth_v3.CheckResponse{
		Status: &status.Status{
			Code: int32(code.Code_PERMISSION_DENIED),
		},
	}, nil
}

// extractE2ETestMetadata extracts the e2e-conformance-metadata static untyped route metadata.
// This metadata was set on the HTTPRoute via the https://gateway.envoyproxy.io/contributions/design/metadata/ feature,
// and sent to the authorization server in the CheckRequest via the SecurityPolicy accessibleMetadata field.
func extractE2ETestMetadata(req *envoy_service_auth_v3.CheckRequest) (string, error) {
	routeMetadata := req.GetAttributes().GetRouteMetadataContext().GetFilterMetadata()
	egMetadata, ok := routeMetadata["envoy-gateway"]
	if !ok {
		return "", fmt.Errorf("envoy-gateway namespace not found in untyped route metadata, found: %v", routeMetadata)
	}

	resources, ok := egMetadata.GetFields()["resources"]
	if !ok {
		return "", fmt.Errorf("resources not found in envoy-gateway metadata, found: %v", egMetadata.GetFields())
	}

	resourcesList, ok := resources.GetKind().(*structpb.Value_ListValue)
	if !ok {
		return "", fmt.Errorf("resources is not a list, found: %T", resources.GetKind())
	}
	if len(resourcesList.ListValue.GetValues()) == 0 {
		return "", fmt.Errorf("resources list is empty, found: %v", resourcesList.ListValue.GetValues())
	}

	resourcesStruct, ok := resourcesList.ListValue.GetValues()[0].GetKind().(*structpb.Value_StructValue)
	if !ok {
		return "", fmt.Errorf("resources is not a struct, found: %T", resources.GetKind())
	}

	annotations, ok := resourcesStruct.StructValue.GetFields()["annotations"]
	if !ok {
		return "", fmt.Errorf("annotations not found in resources, found: %v", resourcesStruct.StructValue.GetFields())
	}

	annotationsStruct, ok := annotations.GetKind().(*structpb.Value_StructValue)
	if !ok {
		return "", fmt.Errorf("annotations is not a struct, found: %T", annotations.GetKind())
	}

	testMetadata, ok := annotationsStruct.StructValue.GetFields()["e2e-conformance-metadata"]
	if !ok {
		return "", fmt.Errorf("e2e-conformance-metadata not found in annotations, found: %v", annotationsStruct.StructValue.GetFields())
	}

	testMetadataString, ok := testMetadata.GetKind().(*structpb.Value_StringValue)
	if !ok {
		return "", fmt.Errorf("e2e-conformance-metadata is not a string, found: %T", testMetadata.GetKind())
	}
	return testMetadataString.StringValue, nil
}

// Users holds a list of users.
type Users map[string]string

// Check checks if a key could retrieve a user from a list of users.
func (u Users) Check(key string) (bool, string) {
	value, ok := u[key]
	if !ok {
		return false, ""
	}
	return ok, value
}

func authCheckerHandler(w http.ResponseWriter, req *http.Request) {
	authorization := req.Header.Get("authorization")
	log.Println("HTTP check auth")
	if len(authorization) == 0 {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	extracted := strings.Split(authorization, " ")
	if len(extracted) == 2 && extracted[0] == "Bearer" {
		if user, ok := testUsers[extracted[1]]; ok {
			w.Header().Add("x-current-user", user) // this should be set before call WriteHeader
			w.WriteHeader(http.StatusOK)
			return
		}
	}

	w.WriteHeader(http.StatusForbidden)
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	certPool, err := loadCA(certPath)
	if err != nil {
		log.Fatalf("Could not load CA certificate: %v", err)
	}

	// Create TLS configuration
	tlsConfig := &tls.Config{
		RootCAs: certPool,
	}

	// Create gRPC dial options
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)),
	}

	conn, err := grpc.NewClient("localhost:9002", opts...)
	if err != nil {
		log.Fatalf("Could not connect: %v", err)
	}
	client := envoy_service_auth_v3.NewAuthorizationClient(conn)

	response, err := client.Check(context.Background(), &envoy_service_auth_v3.CheckRequest{
		Attributes: &envoy_service_auth_v3.AttributeContext{
			Request: &envoy_service_auth_v3.AttributeContext_Request{
				Http: &envoy_service_auth_v3.AttributeContext_HttpRequest{
					Headers: map[string]string{
						"authorization": "Bearer token1",
					},
				},
			},
		},
	})
	if err != nil {
		log.Fatalf("Could not check: %v", err)
	}
	if response != nil && response.Status.Code == int32(code.Code_OK) {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
}

func loadTLSCredentials(certPath string) (credentials.TransportCredentials, error) {
	// Load server's certificate and private key
	crt := "server.crt"
	key := "server.key"

	if certPath != "" {
		if !strings.HasSuffix(certPath, "/") {
			certPath = fmt.Sprintf("%s/", certPath)
		}
		crt = fmt.Sprintf("%s%s", certPath, crt)
		key = fmt.Sprintf("%s%s", certPath, key)
	}
	certificate, err := tls.LoadX509KeyPair(crt, key)
	if err != nil {
		return nil, fmt.Errorf("could not load server key pair: %s", err)
	}

	// Create a new credentials object
	creds := credentials.NewTLS(&tls.Config{Certificates: []tls.Certificate{certificate}})

	return creds, nil
}

func loadCA(caPath string) (*x509.CertPool, error) {
	ca := x509.NewCertPool()
	caCertPath := "server.crt"
	if caPath != "" {
		if !strings.HasSuffix(caPath, "/") {
			caPath = fmt.Sprintf("%s/", caPath)
		}
		caCertPath = fmt.Sprintf("%s%s", caPath, caCertPath)
	}
	caCert, err := os.ReadFile(caCertPath)
	if err != nil {
		return nil, fmt.Errorf("could not read ca certificate: %s", err)
	}
	ca.AppendCertsFromPEM(caCert)
	return ca, nil
}
