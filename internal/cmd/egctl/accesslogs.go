// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package egctl

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os/signal"
	"strconv"
	"syscall"

	accesslogv3 "github.com/envoyproxy/go-control-plane/envoy/service/accesslog/v3"
	"github.com/golang/protobuf/jsonpb"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

const (
	// accessLogServerAddress is the listening address of the access-log server.
	accessLogServerAddress = "0.0.0.0"
	// accessLogServerPort is the default listening port of the access-log server.
	accessLogServerPort = 9001
)

func NewAccessLogsCommand() *cobra.Command {
	var (
		port      int
		gateway   string
		listener  string
		namespace string
	)

	accessLogsCommand := &cobra.Command{
		Use:     "access-logs",
		Short:   "Streaming access logs from Envoy Proxy.",
		Example: ``, // TODO(sh2): add example
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("missing the name of one gateway")
			}

			ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
			defer stop()

			gateway = args[0]
			if err := accessLogs(ctx, port, gateway, listener, namespace); err != nil {
				return err
			}

			<-ctx.Done()
			return nil
		},
	}

	accessLogsCommand.PersistentFlags().IntVarP(&port, "port", "p", accessLogServerPort, "Listening port of access-log server.")
	accessLogsCommand.PersistentFlags().StringVarP(&listener, "listener", "l", "", "Listener name of the gateway.")
	accessLogsCommand.PersistentFlags().StringVarP(&namespace, "namespace", "n", "default", "Namespace of the gateway.")

	return accessLogsCommand
}

type accessLogServer struct {
	marshaler jsonpb.Marshaler
}

var _ accesslogv3.AccessLogServiceServer = &accessLogServer{}

func newAccessLogServer() accesslogv3.AccessLogServiceServer {
	return &accessLogServer{}
}

func (a *accessLogServer) StreamAccessLogs(server accesslogv3.AccessLogService_StreamAccessLogsServer) error {
	log.Println("Start streaming access logs from server")
	for {
		recv, err := server.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		str, _ := a.marshaler.MarshalToString(recv)
		log.Println(str) // TODO(sh2): prettify the json output
	}
}

func serveAccessLogServer(ctx context.Context, server *grpc.Server, addr string) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Printf("failed to listen on address %s: %v\n", addr, err)
		return
	}

	go func() {
		<-ctx.Done()
		log.Println("Shutting down access-log server")
		server.Stop()
	}()

	log.Printf("Serving access-log server on %s\n", addr)
	if err = server.Serve(listener); err != nil {
		log.Printf("failed to start access-log server: %v\n", err)
	}
}

func isValidPort(port int) bool {
	if port < 1024 || port > 65535 {
		return false
	}
	return true
}

func validateAccessLogsInputs(port int) error {
	if !isValidPort(port) {
		return fmt.Errorf("%d is not a valid port", port)
	}

	return nil
}

func accessLogs(ctx context.Context, port int, gateway, listener, namespace string) error {
	if err := validateAccessLogsInputs(port); err != nil {
		return err
	}

	// TODO(sh2): check whether the envoy patch policy is enabled, return err if not

	grpcServer := grpc.NewServer()
	accesslogv3.RegisterAccessLogServiceServer(grpcServer, newAccessLogServer())

	addr := net.JoinHostPort(accessLogServerAddress, strconv.Itoa(port))
	go serveAccessLogServer(ctx, grpcServer, addr)

	return nil
}
