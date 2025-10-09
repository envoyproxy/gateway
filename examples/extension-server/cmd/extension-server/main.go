// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package main

import (
	extensionv1alpha1 "github.com/exampleorg/envoygateway-extension/api/v1alpha1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/exampleorg/envoygateway-extension/internal/extensionserver"
	"github.com/urfave/cli/v2"
	"google.golang.org/grpc"

	pb "github.com/envoyproxy/gateway/proto/extension"
)

func init() {
	runtime.Must(extensionv1alpha1.AddToScheme(scheme.Scheme))
}

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
	pb.RegisterEnvoyGatewayExtensionServer(grpcServer, extensionserver.New(logger))
	return grpcServer.Serve(lis)
}
