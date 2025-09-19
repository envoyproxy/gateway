// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package common

import "fmt"

const (
	// #nosec G101 - This is a file path, not a credential
	SdsServiceAccountTokenFilename = "xds-service-account-token.json"
)

// GetSdsServiceAccountTokenConfigMapData returns the SDS config map data for the service account token used in GatewayNamespaceMode.
func GetSdsServiceAccountTokenConfigMapData(tokenPath string) string {
	return fmt.Sprintf(`{"resources":[{"@type":"type.googleapis.com/envoy.extensions.transport_sockets.tls.v3.Secret",`+
		`"name":"xds-service-account-token","service_account_token":{"path":"%s"}}]}`, tokenPath)
}
