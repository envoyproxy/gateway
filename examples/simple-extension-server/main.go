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

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
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
					&cli.StringFlag{
						Name:        "suffix",
						Usage:       "suffix appended to VirtualHost domains and used in response headers",
						DefaultText: "extserver",
						Value:       "extserver",
					},
				},
			},
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		panic(err)
	}
}

var grpcServer *grpc.Server

func handleSignals(_ *cli.Context) error {
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
	pb.RegisterEnvoyGatewayExtensionServer(grpcServer, New(logger, sig, cCtx.String("suffix")))
	return grpcServer.Serve(lis)
}

type Server struct {
	pb.UnimplementedEnvoyGatewayExtensionServer
	sig    chan int
	log    *slog.Logger
	suffix string
}

func New(logger *slog.Logger, sig chan int, suffix string) *Server {
	return &Server{
		log:    logger,
		sig:    sig,
		suffix: suffix,
	}
}

func (s *Server) PostRouteModify(_ context.Context, req *pb.PostRouteModifyRequest) (*pb.PostRouteModifyResponse, error) {
	s.log.Info("PostRouteModify callback was invoked")

	if req.PostRouteContext != nil && len(req.PostRouteContext.ExtensionResources) > 0 {
		s.log.Info("PostRouteModify received extension resources, adding response header",
			slog.Int("count", len(req.PostRouteContext.ExtensionResources)))
		req.Route.ResponseHeadersToAdd = append(req.Route.ResponseHeadersToAdd, &corev3.HeaderValueOption{
			Header: &corev3.HeaderValue{
				Key:   "x-ext-server",
				Value: s.suffix,
			},
		})
	}

	return &pb.PostRouteModifyResponse{
		Route: req.Route,
	}, nil
}

func (s *Server) PostVirtualHostModify(_ context.Context, req *pb.PostVirtualHostModifyRequest) (*pb.PostVirtualHostModifyResponse, error) {
	s.log.Info("PostVirtualHostModify callback was invoked")

	if strings.Contains(req.VirtualHost.Name, "fail") {
		s.log.Info("PostVirtualHostModify returning unavailable error")
		return nil, status.Error(codes.Unavailable, "Service is currently unavailable")
	}

	s.log.Info("PostVirtualHostModify sending response")
	if len(req.VirtualHost.Domains) > 0 {
		lastDomain := len(req.VirtualHost.Domains) - 1
		newDomain := fmt.Sprintf("%s.%s", req.VirtualHost.Domains[lastDomain], s.suffix)
		s.log.Info("PostVirtualHostModify appending suffix to last domain",
			slog.String("originalDomain", req.VirtualHost.Domains[lastDomain]),
			slog.String("newDomain", newDomain))

		req.VirtualHost.Domains = append(req.VirtualHost.Domains, newDomain)
	}
	return &pb.PostVirtualHostModifyResponse{
		VirtualHost: req.VirtualHost,
	}, nil
}

func (s *Server) PostHTTPListenerModify(_ context.Context, req *pb.PostHTTPListenerModifyRequest) (*pb.PostHTTPListenerModifyResponse, error) {
	s.log.Info("postHTTPListenerModify callback was invoked")

	return &pb.PostHTTPListenerModifyResponse{
		Listener: req.Listener,
	}, nil
}

func (s *Server) PostTranslateModify(_ context.Context, req *pb.PostTranslateModifyRequest) (*pb.PostTranslateModifyResponse, error) {
	s.log.Info("PostTranslateModify callback was invoked")

	return &pb.PostTranslateModifyResponse{
		Secrets:  req.Secrets,
		Clusters: req.Clusters,
	}, nil
}
