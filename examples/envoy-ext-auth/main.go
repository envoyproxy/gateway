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
	"sort"
	"strings"
	"time"

	envoy_api_v3_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_service_auth_v3 "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	"github.com/golang/protobuf/ptypes/wrappers"
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

const (
	routeMetadataNamespace    = "envoy-gateway"
	routeMetadataResources    = "resources"
	routeMetadataKind         = "kind"
	routeMetadataName         = "name"
	routeMetadataNamespaceKey = "namespace"
	routeMetadataAnnotations  = "annotations"
	routeNameHeader           = "x-eg-route-name"
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
	delayIfNeed()

	authorization := req.Attributes.Request.Http.Headers["authorization"]
	log.Println("GRPC check auth: ", authorization)

	extracted := strings.Fields(authorization)
	if len(extracted) == 2 && extracted[0] == "Bearer" {
		valid, user := s.users.Check(extracted[1])
		if valid {
			headers := []*envoy_api_v3_core.HeaderValueOption{
				{
					Append: &wrappers.BoolValue{Value: false},
					Header: &envoy_api_v3_core.HeaderValue{
						// For a successful request, the authorization server sets the
						// x-current-user value.
						Key:   "x-current-user",
						Value: user,
					},
				},
			}
			if routeMetadata, ok := getRouteMetadata(req); ok {
				headers = append(headers, &envoy_api_v3_core.HeaderValueOption{
					Append: &wrappers.BoolValue{Value: false},
					Header: &envoy_api_v3_core.HeaderValue{
						Key:   routeNameHeader,
						Value: routeMetadata.Name,
					},
				})
				for _, key := range sortedAnnotationKeys(routeMetadata.Annotations) {
					headers = append(headers, &envoy_api_v3_core.HeaderValueOption{
						Append: &wrappers.BoolValue{Value: false},
						Header: &envoy_api_v3_core.HeaderValue{
							Key:   routeAnnotationHeaderName(key),
							Value: routeMetadata.Annotations[key],
						},
					})
				}
				log.Printf("GRPC route metadata: kind=%s namespace=%s name=%s", routeMetadata.Kind, routeMetadata.Namespace, routeMetadata.Name)
			}
			return &envoy_service_auth_v3.CheckResponse{
				HttpResponse: &envoy_service_auth_v3.CheckResponse_OkResponse{
					OkResponse: &envoy_service_auth_v3.OkHttpResponse{
						Headers: headers,
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

type routeMetadata struct {
	Kind        string
	Namespace   string
	Name        string
	Annotations map[string]string
}

func getRouteMetadata(req *envoy_service_auth_v3.CheckRequest) (*routeMetadata, bool) {
	if req == nil || req.Attributes == nil || req.Attributes.RouteMetadataContext == nil {
		return nil, false
	}

	ns := req.Attributes.RouteMetadataContext.FilterMetadata[routeMetadataNamespace]
	if ns == nil {
		return nil, false
	}

	resources := ns.Fields[routeMetadataResources]
	if resources == nil || resources.GetListValue() == nil || len(resources.GetListValue().Values) == 0 {
		return nil, false
	}

	resource := resources.GetListValue().Values[0].GetStructValue()
	if resource == nil {
		return nil, false
	}

	md := &routeMetadata{
		Kind:        structFieldString(resource, routeMetadataKind),
		Namespace:   structFieldString(resource, routeMetadataNamespaceKey),
		Name:        structFieldString(resource, routeMetadataName),
		Annotations: structFields(resource.Fields[routeMetadataAnnotations].GetStructValue()),
	}
	if md.Kind == "" || md.Namespace == "" || md.Name == "" {
		return nil, false
	}

	return md, true
}

func structFieldString(st *structpb.Struct, key string) string {
	if st == nil || st.Fields[key] == nil {
		return ""
	}
	return st.Fields[key].GetStringValue()
}

func structFields(st *structpb.Struct) map[string]string {
	if st == nil || len(st.Fields) == 0 {
		return nil
	}

	fields := make(map[string]string, len(st.Fields))
	for key, value := range st.Fields {
		if value == nil {
			continue
		}
		if stringValue := value.GetStringValue(); stringValue != "" {
			fields[key] = stringValue
		}
	}
	if len(fields) == 0 {
		return nil
	}
	return fields
}

func sortedAnnotationKeys(annotations map[string]string) []string {
	if len(annotations) == 0 {
		return nil
	}

	keys := make([]string, 0, len(annotations))
	for key := range annotations {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func routeAnnotationHeaderName(key string) string {
	return "x-eg-route-annotation-" + strings.ToLower(key)
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

func getDelayDuration() *time.Duration {
	delayDuration := os.Getenv("DELAY_DURATION")
	if delayDuration == "" {
		return nil
	}

	d, err := time.ParseDuration(delayDuration)
	if err != nil {
		// fallback to default value which means 10s
		d = time.Second * 10
	}

	return &d
}

func delayIfNeed() {
	d := getDelayDuration()
	if d != nil {
		time.Sleep(*d)
	}
}

func authCheckerHandler(w http.ResponseWriter, req *http.Request) {
	delayIfNeed()
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
			w.Header().Add("x-ext-auth-req-path", req.URL.Path)
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
