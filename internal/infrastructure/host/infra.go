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
	if err := createSdsConfig(); err != nil {
		return nil, fmt.Errorf("failed to create sds config: %w", err)
	}

	infra := &Infra{
		HomeDir:         defaultHomeDir,
		Logger:          logger,
		EnvoyGateway:    cfg.EnvoyGateway,
		proxyContextMap: make(map[string]*proxyContext),
	}
	go infra.cleanProxy(runnerCtx)

	return infra, nil
}

// cleanProxy stops all the running proxies when infra provider is closing.
func (i *Infra) cleanProxy(ctx context.Context) {
	<-ctx.Done()
	if len(i.proxyContextMap) < 1 {
		return
	}

	i.Logger.Info("start cleaning up proxies")
	for name, proxyCtx := range i.proxyContextMap {
		proxyCtx.cancel()
		i.Logger.Info("proxy closed", "name", name)
	}
	i.Logger.Info("all proxies has been cleaned up")
}

// createSdsConfig creates the needing Sds config under defaultLocalCertPathDir.
func createSdsConfig() error {
	writeFile := func(fn, text string) error {
		f, wErr := os.Create(fn)
		if wErr != nil {
			return wErr
		}
		defer f.Close()

		_, wErr = f.WriteString(text)
		return wErr
	}

	if err := writeFile(filepath.Join(defaultLocalCertPathDir, common.SdsCAFilename),
		common.GetSdsCAConfigMapData(
			filepath.Join(defaultLocalCertPathDir, XdsTLSCaFilename))); err != nil {
		return err
	}

	if err := writeFile(filepath.Join(defaultLocalCertPathDir, common.SdsCertFilename),
		common.GetSdsCertConfigMapData(
			filepath.Join(defaultLocalCertPathDir, XdsTLSCertFilename),
			filepath.Join(defaultLocalCertPathDir, XdsTLSKeyFilename))); err != nil {
		return err
	}

	return nil
}
