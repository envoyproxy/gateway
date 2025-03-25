// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package host

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/infrastructure/common"
	"github.com/envoyproxy/gateway/internal/logging"
	"github.com/envoyproxy/gateway/internal/utils/file"
)

const (
	// TODO: Make these path configurable.
	defaultHomeDir          = "/tmp/envoy-gateway"
	defaultLocalCertPathDir = "/tmp/envoy-gateway/certs/envoy"

	// XdsTLSCertFilename is the fully qualified name of the file containing Envoy's
	// xDS server TLS certificate.
	XdsTLSCertFilename = "tls.crt"
	// XdsTLSKeyFilename is the fully qualified name of the file containing Envoy's
	// xDS server TLS key.
	XdsTLSKeyFilename = "tls.key"
	// XdsTLSCaFilename is the fully qualified name of the file containing Envoy's
	// trusted CA certificate.
	XdsTLSCaFilename = "ca.crt"
)

// Infra manages the creation and deletion of host process
// based on Infra IR resources.
type Infra struct {
	HomeDir string
	Logger  logging.Logger

	// EnvoyGateway is the configuration used to startup Envoy Gateway.
	EnvoyGateway *egv1a1.EnvoyGateway

	// proxyContextMap store the context of each running proxy by its name for lifecycle management.
	proxyContextMap map[string]*proxyContext

	// TODO: remove this field once it supports the configurable homeDir
	sdsConfigPath string
}

func NewInfra(runnerCtx context.Context, cfg *config.Server, logger logging.Logger) (*Infra, error) {
	// Ensure the home directory exist.
	if err := os.MkdirAll(defaultHomeDir, 0o750); err != nil {
		return nil, fmt.Errorf("failed to create dir: %w", err)
	}

	// Check local certificates dir exist.
	if _, err := os.Lstat(defaultLocalCertPathDir); err != nil {
		return nil, fmt.Errorf("failed to stat dir: %w", err)
	}

	// Ensure the sds config exist.
	if err := createSdsConfig(defaultLocalCertPathDir); err != nil {
		return nil, fmt.Errorf("failed to create sds config: %w", err)
	}

	infra := &Infra{
		HomeDir:         defaultHomeDir,
		Logger:          logger,
		EnvoyGateway:    cfg.EnvoyGateway,
		proxyContextMap: make(map[string]*proxyContext),
		sdsConfigPath:   defaultLocalCertPathDir,
	}
	return infra, nil
}

// createSdsConfig creates the needing SDS config under certain directory.
func createSdsConfig(dir string) error {
	if err := file.Write(common.GetSdsCAConfigMapData(
		filepath.Join(dir, XdsTLSCaFilename)),
		filepath.Join(dir, common.SdsCAFilename)); err != nil {
		return err
	}

	if err := file.Write(common.GetSdsCertConfigMapData(
		filepath.Join(dir, XdsTLSCertFilename),
		filepath.Join(dir, XdsTLSKeyFilename)),
		filepath.Join(dir, common.SdsCertFilename)); err != nil {
		return err
	}

	return nil
}
