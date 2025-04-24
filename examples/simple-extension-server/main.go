// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/urfave/cli/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/envoyproxy/gateway/proto/extension"
)

func main() {
	app := cli.App{
		Name:           "extension-server",
		Version:        "0.0.1",
		Description:    "Example Envoy Gateway Extension Server",
		DefaultCommand: "server",
		Commands: []*cli.Command{
			{
				Name:   "server",
				Usage:  "runs the Extension Server",
				Before: handleSignals,
				Action: startExtensionServer,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "host",
						Usage:       "the host on which to listen",
						DefaultText: "0.0.0.0",
						Value:       "0.0.0.0",
					},
					&cli.IntFlag{
						Name:        "port",
						Usage:       "the port on which to listen",
						DefaultText: "5005",
						Value:       5005,
					},
					&cli.StringFlag{
						Name:        "log-level",
						Usage:       "the log level, should be one of Debug/Info/Warn/Error",
						DefaultText: "Info",
						Value:       "Info",
					},
				},
			},
		},
	}
	app.Run(os.Args)
}

var grpcServer *grpc.Server

func handleSignals(cCtx *cli.Context) error {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGQUIT)
	go func() {
		for range c {
			if grpcServer != nil {
				grpcServer.Stop()
				os.Exit(0)
			}
		}
	}()
	return nil
}

func startExtensionServer(cCtx *cli.Context) error {
	var level slog.Level
	if err := level.UnmarshalText([]byte(cCtx.String("log-level"))); err != nil {
		level = slog.LevelInfo
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: level,
	}))
	address := net.JoinHostPort(cCtx.String("host"), cCtx.String("port"))
	logger.Info("Starting the extension server", slog.String("host", address))
	lis, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	var opts []grpc.ServerOption
	grpcServer = grpc.NewServer(opts...)
	sig := make(chan int, 1)
	pb.RegisterEnvoyGatewayExtensionServer(grpcServer, New(logger, sig))
	return grpcServer.Serve(lis)
}

type Server struct {
	pb.UnimplementedEnvoyGatewayExtensionServer
	sig chan int
	log *slog.Logger
}

func New(logger *slog.Logger, sig chan int) *Server {
	return &Server{
		log: logger,
		sig: sig,
	}
}

func (s *Server) PostRouteModify(ctx context.Context, req *pb.PostRouteModifyRequest) (*pb.PostRouteModifyResponse, error) {
	s.log.Info("PostRouteModify callback was invoked")

	return &pb.PostRouteModifyResponse{
		Route: req.Route,
	}, nil
}

func (s *Server) PostVirtualHostModify(ctx context.Context, req *pb.PostVirtualHostModifyRequest) (*pb.PostVirtualHostModifyResponse, error) {
	s.log.Info("PostVirtualHostModify callback was invoked")

	if strings.Contains(req.VirtualHost.Name, "fail") {
		s.log.Info("PostVirtualHostModify returning unavailable error")
		return nil, status.Error(codes.Unavailable, "Service is currently unavailable")
	} else {
		s.log.Info("PostVirtualHostModify sending response")
		if len(req.VirtualHost.Domains) > 0 {
			req.VirtualHost.Domains = append(req.VirtualHost.Domains, fmt.Sprintf("%s.extserver", req.VirtualHost.Domains[0]))
		}
		return &pb.PostVirtualHostModifyResponse{
			VirtualHost: req.VirtualHost,
		}, nil
	}
}

func (s *Server) PostHTTPListenerModify(ctx context.Context, req *pb.PostHTTPListenerModifyRequest) (*pb.PostHTTPListenerModifyResponse, error) {
	s.log.Info("postHTTPListenerModify callback was invoked")

	return &pb.PostHTTPListenerModifyResponse{
		Listener: req.Listener,
	}, nil
}

func (s *Server) PostTranslateModify(ctx context.Context, req *pb.PostTranslateModifyRequest) (*pb.PostTranslateModifyResponse, error) {
	s.log.Info("PostTranslateModify callback was invoked")

	return &pb.PostTranslateModifyResponse{
		Secrets:  req.Secrets,
		Clusters: req.Clusters,
	}, nil
}
