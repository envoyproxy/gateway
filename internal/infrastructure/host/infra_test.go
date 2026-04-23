// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package host

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/infrastructure/common"
	"github.com/envoyproxy/gateway/internal/utils/file"
)

func TestMaybeGenerateCertificates(t *testing.T) {
	cfg, err := config.New(io.Discard, io.Discard)
	require.NoError(t, err)

	certFiles := []string{"ca.crt", "tls.crt", "tls.key"}

	t.Run("all_files_exist", func(t *testing.T) {
		tmpDir := t.TempDir()
		certPath := filepath.Join(tmpDir, "envoy")

		// Create directory and dummy files
		require.NoError(t, os.MkdirAll(certPath, 0o750))
		for _, filename := range certFiles {
			fpath := filepath.Join(certPath, filename)
			require.NoError(t, os.WriteFile(fpath, []byte("dummy"), 0o600))
		}

		err := maybeGenerateCertificates(cfg, certPath)
		require.NoError(t, err)

		// Verify files still exist and unchanged size
		for _, filename := range certFiles {
			data, err := os.ReadFile(filepath.Join(certPath, filename))
			require.NoError(t, err)
			require.Len(t, data, 5) // "dummy"
		}
	})

	t.Run("missing_files", func(t *testing.T) {
		tmpDir := t.TempDir()
		certPath := filepath.Join(tmpDir, "envoy")

		err := maybeGenerateCertificates(cfg, certPath)
		require.NoError(t, err)

		// Verify directory created
		info, err := os.Stat(certPath)
		require.NoError(t, err)
		require.True(t, info.IsDir())

		// Verify all files created and non-empty
		for _, filename := range certFiles {
			data, err := os.ReadFile(filepath.Join(certPath, filename))
			require.NoError(t, err)
			require.NotEmpty(t, data, filename)
		}
	})

	t.Run("partial_files_missing", func(t *testing.T) {
		tmpDir := t.TempDir()
		certPath := filepath.Join(tmpDir, "envoy")

		require.NoError(t, os.MkdirAll(certPath, 0o750))

		// Create only one file
		require.NoError(t, os.WriteFile(filepath.Join(certPath, "ca.crt"), []byte("dummy"), 0o600))

		err := maybeGenerateCertificates(cfg, certPath)
		require.NoError(t, err)

		// Verify all files created and non-empty
		for _, filename := range certFiles {
			data, err := os.ReadFile(filepath.Join(certPath, filename))
			require.NoError(t, err)
			require.NotEmpty(t, data, filename)
		}
	})

	t.Run("cert_generation_fails", func(t *testing.T) {
		tmpDir := t.TempDir()
		// This tests mkdir fail by making parent unwritable
		unwritableDir := filepath.Join(tmpDir, "unwritable")
		require.NoError(t, os.Mkdir(unwritableDir, 0o555)) // Read-only

		badCertPath := filepath.Join(unwritableDir, "envoy")
		err := maybeGenerateCertificates(cfg, badCertPath)
		require.ErrorContains(t, err, "failed to create cert directory")
	})
}

func TestCreateSdsConfig(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dir := t.TempDir()
		// Create required cert files
		require.NoError(t, file.Write("test ca", filepath.Join(dir, XdsTLSCaFilename)))
		require.NoError(t, file.Write("test cert", filepath.Join(dir, XdsTLSCertFilename)))
		require.NoError(t, file.Write("test key", filepath.Join(dir, XdsTLSKeyFilename)))

		err := createSdsConfig(dir)
		require.NoError(t, err)

		// Verify CA config was created
		caConfigPath := filepath.Join(dir, common.SdsCAFilename)
		actualCAConfig, err := os.ReadFile(caConfigPath)
		require.NoError(t, err)
		require.NotEmpty(t, actualCAConfig)

		// Verify cert config was created
		certConfigPath := filepath.Join(dir, common.SdsCertFilename)
		actualCertConfig, err := os.ReadFile(certConfigPath)
		require.NoError(t, err)
		require.NotEmpty(t, actualCertConfig)
	})

	t.Run("error_writing_ca_config", func(t *testing.T) {
		// Use invalid path to force file.Write to fail
		invalidDir := filepath.Join("/", "nonexistent", "invalid", "path")
		err := createSdsConfig(invalidDir)
		require.Error(t, err)
	})
}
