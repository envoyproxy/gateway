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
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	"google.golang.org/grpc/credentials"

	envoy_api_v3_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_service_proc_v3 "github.com/envoyproxy/go-control-plane/envoy/service/ext_proc/v3"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type extProcServer struct{}

var (
	port     int
	certPath string
)

func main() {
	flag.IntVar(&port, "port", 9002, "gRPC port")
	flag.StringVar(&certPath, "certPath", "", "path to extProcServer certificate and private key")
	flag.Parse()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	creds, err := loadTLSCredentials(certPath)
	if err != nil {
		log.Fatalf("Failed to load TLS credentials: %v", err)
	}
	gs := grpc.NewServer(grpc.Creds(creds))
	envoy_service_proc_v3.RegisterExternalProcessorServer(gs, &extProcServer{})

	go func() {
		err = gs.Serve(lis)
		if err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	// Create Unix listener
	gus := grpc.NewServer(grpc.Creds(creds))
	envoy_service_proc_v3.RegisterExternalProcessorServer(gus, &extProcServer{})

	udsAddr := "/var/run/ext-proc/extproc.sock"
	if _, err := os.Stat(udsAddr); err == nil {
		if err := os.RemoveAll(udsAddr); err != nil {
			log.Fatalf("failed to remove: %v", err)
		}
	}

	ul, err := net.Listen("unix", udsAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	err = os.Chmod(udsAddr, 0700)
	if err != nil {
		log.Fatalf("failed to set permissions: %v", err)
	}

	// envoy distroless uid
	err = os.Chown(udsAddr, 65532, 0)
	if err != nil {
		log.Fatalf("failed to set permissions: %v", err)
	}

	go func() {
		err = gus.Serve(ul)
		if err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	http.HandleFunc("/healthz", healthCheckHandler)
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

// used by k8s readiness probes
// makes a processing request to check if the processor service is healthy
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	certPool, err := loadCA(certPath)
	if err != nil {
		log.Fatalf("Could not load CA certificate: %v", err)
	}

	// Create TLS configuration
	tlsConfig := &tls.Config{
		RootCAs:    certPool,
		ServerName: "grpc-ext-proc.envoygateway",
	}

	// Create gRPC dial options
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)),
	}

	conn, err := grpc.Dial("localhost:9002", opts...)
	if err != nil {
		log.Fatalf("Could not connect: %v", err)
	}
	client := envoy_service_proc_v3.NewExternalProcessorClient(conn)

	processor, err := client.Process(context.Background())
	if err != nil {
		log.Fatalf("Could not check: %v", err)
	}

	err = processor.Send(&envoy_service_proc_v3.ProcessingRequest{
		Request: &envoy_service_proc_v3.ProcessingRequest_RequestHeaders{
			RequestHeaders: &envoy_service_proc_v3.HttpHeaders{},
		},
	})
	if err != nil {
		log.Fatalf("Could not check: %v", err)
	}

	response, err := processor.Recv()
	if err != nil {
		log.Fatalf("Could not check: %v", err)
	}

	if response != nil && response.GetRequestHeaders().Response.Status == envoy_service_proc_v3.CommonResponse_CONTINUE {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
}

func loadTLSCredentials(certPath string) (credentials.TransportCredentials, error) {
	// Load extProcServer's certificate and private key
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
		return nil, fmt.Errorf("could not load extProcServer key pair: %s", err)
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

func (s *extProcServer) Process(srv envoy_service_proc_v3.ExternalProcessor_ProcessServer) error {
	ctx := srv.Context()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		req, err := srv.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return status.Errorf(codes.Unknown, "cannot receive stream request: %v", err)
		}

		resp := &envoy_service_proc_v3.ProcessingResponse{}
		switch v := req.Request.(type) {
		case *envoy_service_proc_v3.ProcessingRequest_RequestHeaders:
			xrch := ""
			if v.RequestHeaders != nil {
				hdrs := v.RequestHeaders.Headers.GetHeaders()
				for _, hdr := range hdrs {
					if hdr.Key == "x-request-client-header" {
						xrch = string(hdr.RawValue)
					}
				}
			}

			rhq := &envoy_service_proc_v3.HeadersResponse{
				Response: &envoy_service_proc_v3.CommonResponse{
					HeaderMutation: &envoy_service_proc_v3.HeaderMutation{
						SetHeaders: []*envoy_api_v3_core.HeaderValueOption{
							{
								Header: &envoy_api_v3_core.HeaderValue{
									Key:      "x-request-ext-processed",
									RawValue: []byte("true"),
								},
							},
						},
					},
				},
			}

			if xrch != "" {
				rhq.Response.HeaderMutation.SetHeaders = append(rhq.Response.HeaderMutation.SetHeaders,
					&envoy_api_v3_core.HeaderValueOption{
						Header: &envoy_api_v3_core.HeaderValue{
							Key:      "x-request-client-header",
							RawValue: []byte("mutated"),
						},
					})
				rhq.Response.HeaderMutation.SetHeaders = append(rhq.Response.HeaderMutation.SetHeaders,
					&envoy_api_v3_core.HeaderValueOption{
						Header: &envoy_api_v3_core.HeaderValue{
							Key:      "x-request-client-header-received",
							RawValue: []byte(xrch),
						},
					})
			}

			resp = &envoy_service_proc_v3.ProcessingResponse{
				Response: &envoy_service_proc_v3.ProcessingResponse_RequestHeaders{
					RequestHeaders: rhq,
				},
			}
			break
		case *envoy_service_proc_v3.ProcessingRequest_ResponseHeaders:
			rhq := &envoy_service_proc_v3.HeadersResponse{
				Response: &envoy_service_proc_v3.CommonResponse{
					HeaderMutation: &envoy_service_proc_v3.HeaderMutation{
						SetHeaders: []*envoy_api_v3_core.HeaderValueOption{
							{
								Header: &envoy_api_v3_core.HeaderValue{
									Key:      "x-response-ext-processed",
									RawValue: []byte("true"),
								},
							},
						},
					},
				},
			}
			resp = &envoy_service_proc_v3.ProcessingResponse{
				Response: &envoy_service_proc_v3.ProcessingResponse_ResponseHeaders{
					ResponseHeaders: rhq,
				},
			}
			break
		default:
			log.Printf("Unknown Request type %v\n", v)
		}
		if err := srv.Send(resp); err != nil {
			log.Printf("send error %v", err)
		}
	}
}
