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
	"sync"
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
		port int
	)

	accessLogsCommand := &cobra.Command{
		Use:     "access-logs",
		Short:   "Get access logs from Envoy.",
		Example: ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			return accessLogs(port)
		},
	}

	accessLogsCommand.PersistentFlags().IntVarP(&port, "port", "p", accessLogServerPort, "Listening port of access-log server.")

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
		log.Println(str)
	}
}

func serveAccessLogServer(ctx context.Context, wg *sync.WaitGroup, server *grpc.Server, addr string) {
	defer wg.Done()

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

func isValidatePort(port int) bool {
	if port < 1024 || port > 65535 {
		return false
	}
	return true
}

func validateAccessLogsInputs(port int) error {
	if !isValidatePort(port) {
		return fmt.Errorf("%d is not a valid port", port)
	}

	return nil
}

func accessLogs(port int) error {
	if err := validateAccessLogsInputs(port); err != nil {
		return err
	}

	wg := &sync.WaitGroup{}
	ctx, stop := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	defer stop()

	grpcServer := grpc.NewServer()
	accesslogv3.RegisterAccessLogServiceServer(grpcServer, newAccessLogServer())
	addr := net.JoinHostPort(accessLogServerAddress, strconv.Itoa(port))

	wg.Add(1)
	go serveAccessLogServer(ctx, wg, grpcServer, addr)

	wg.Wait()

	return nil
}
