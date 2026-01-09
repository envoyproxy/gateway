// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package test

import (
	"strings"

	"github.com/envoyproxy/gateway/internal/utils/cert"
)

// NormalizeCertPath replaces platform-specific cert path with canonical path for consistent golden files.
func NormalizeCertPath(content string) string {
	return strings.ReplaceAll(content, cert.SystemCertPath, cert.CanonicalCertPath)
}

// DenormalizeCertPath replaces canonical cert path with actual system cert path for cross-platform compatibility.
func DenormalizeCertPath(content string) string {
	return strings.ReplaceAll(content, cert.CanonicalCertPath, cert.SystemCertPath)
}
