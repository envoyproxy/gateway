// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package extensionserver

import (
	"log/slog"

	pb "github.com/envoyproxy/gateway/proto/extension"
)

type Server struct {
	pb.UnimplementedEnvoyGatewayExtensionServer

	log *slog.Logger
}

func New(logger *slog.Logger) *Server {
	return &Server{
		log: logger,
	}
}
