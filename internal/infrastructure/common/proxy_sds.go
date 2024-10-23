// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package common

import "fmt"

// xDS certificate rotation is supported by using SDS path-based resource files.

const (
	SdsCAFilename   = "xds-trusted-ca.json"
	SdsCertFilename = "xds-certificate.json"
)

func GetSdsCAConfigMapData(ca string) string {
	return fmt.Sprintf(`{"resources":[{"@type":"type.googleapis.com/envoy.extensions.transport_sockets.tls.v3.Secret",`+
		`"name":"xds_trusted_ca","validation_context":{"trusted_ca":{"filename":"%s"},`+
		`"match_typed_subject_alt_names":[{"san_type":"DNS","matcher":{"exact":"envoy-gateway"}}]}}]}`, ca)
}

func GetSdsCertConfigMapData(tlsCert, tlsKey string) string {
	return fmt.Sprintf(`{"resources":[{"@type":"type.googleapis.com/envoy.extensions.transport_sockets.tls.v3.Secret",`+
		`"name":"xds_certificate","tls_certificate":{"certificate_chain":{"filename":"%s"},`+
		`"private_key":{"filename":"%s"}}}]}`, tlsCert, tlsKey)
}
