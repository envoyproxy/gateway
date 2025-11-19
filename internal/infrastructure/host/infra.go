// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package host

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	func_e "github.com/tetratelabs/func-e"
	func_e_api "github.com/tetratelabs/func-e/api"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/crypto"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/infrastructure/common"
	"github.com/envoyproxy/gateway/internal/logging"
	"github.com/envoyproxy/gateway/internal/utils/file"
)

const (
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
	// Paths contains the XDG-compliant directory paths.
	Paths  *Paths
	Logger logging.Logger

	// EnvoyGateway is the configuration used to startup Envoy Gateway.
	EnvoyGateway *egv1a1.EnvoyGateway

	// proxyContextMap store the context of each running proxy by its name for lifecycle management.
	proxyContextMap sync.Map

	// sdsConfigPath is the path to SDS configuration files.
	sdsConfigPath string

	// defaultEnvoyImage is the default Envoy image to use if no Envoy version is set.
	defaultEnvoyImage string

	// Stdout is the writer for standard output (for func-e and Envoy stdout).
	Stdout io.Writer
	// Stderr is the writer for error output (for Envoy stderr).
	Stderr io.Writer

	// envoyRunner runs Envoy (can be overridden in tests).
	envoyRunner func_e_api.RunFunc
}

func NewInfra(runnerCtx context.Context, cfg *config.Server, logger logging.Logger) (*Infra, error) {
	// Get configuration from provider
	var hostCfg *egv1a1.EnvoyGatewayHostInfrastructureProvider
	if p := cfg.EnvoyGateway.Provider; p != nil && p.Custom != nil &&
		p.Custom.Infrastructure != nil && p.Custom.Infrastructure.Host != nil {
		hostCfg = p.Custom.Infrastructure.Host
	}

	// Get paths using helper
	paths, err := GetPaths(hostCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to determine paths: %w", err)
	}

	// Ensure the data directory exists
	if err := os.MkdirAll(paths.DataHome, 0o750); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	// Check if certificates exist, generate them if not
	certPath := paths.CertDir("envoy")
	if err := maybeGenerateCertificates(cfg, certPath); err != nil {
		return nil, err
	}

	// Ensure the sds config exist
	if err := createSdsConfig(certPath); err != nil {
		return nil, fmt.Errorf("failed to create sds config: %w", err)
	}

	infra := &Infra{
		Paths:             paths,
		Logger:            logger,
		EnvoyGateway:      cfg.EnvoyGateway,
		sdsConfigPath:     certPath,
		defaultEnvoyImage: egv1a1.DefaultEnvoyProxyImage,
		Stdout:            cfg.Stdout,
		Stderr:            cfg.Stderr,
		envoyRunner:       func_e.Run,
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

// maybeGenerateCertificates checks if all required certificate files exist and generates them if any is missing.
func maybeGenerateCertificates(cfg *config.Server, certPath string) error {
	certFiles := []string{"ca.crt", "tls.crt", "tls.key"}

	// Check if any cert file is missing
	var missing bool
	for _, filename := range certFiles {
		filePath := filepath.Join(certPath, filename)
		_, err := os.Lstat(filePath)
		if os.IsNotExist(err) {
			missing = true
			break
		}
		if err != nil {
			return fmt.Errorf("failed to stat %s: %w", filename, err)
		}
	}

	if !missing {
		// All files exist, nothing to do
		return nil
	}

	// Generate certificates automatically
	certs, err := crypto.GenerateCerts(cfg)
	if err != nil {
		return fmt.Errorf("failed to generate certificates: %w", err)
	}

	// Create the cert directory
	if err := os.MkdirAll(certPath, 0o750); err != nil {
		return fmt.Errorf("failed to create cert directory: %w", err)
	}

	// Write cert files
	certMap := map[string][]byte{
		"ca.crt":  certs.CACertificate,
		"tls.crt": certs.EnvoyCertificate,
		"tls.key": certs.EnvoyPrivateKey,
	}

	for filename, content := range certMap {
		if err := file.Write(string(content), filepath.Join(certPath, filename)); err != nil {
			return fmt.Errorf("failed to write %s: %w", filename, err)
		}
	}
	return nil
}
