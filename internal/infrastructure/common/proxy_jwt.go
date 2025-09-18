// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package common

import "fmt"

const (
	SdsServiceAccountTokenFilename = "xds_service_account_token.json"
)

func GetSdsServiceAccountTokenConfigMapData(tokenPath string) string {
	return fmt.Sprintf(`{"resources":[{"@type":"type.googleapis.com/envoy.extensions.transport_sockets.tls.v3.Secret",`+
		`"name":"xds_service_account_token","service_account_token":{"path":"%s"}}]}`, tokenPath)
}
