// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package proxy

import (
	"fmt"

	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	providerutils "github.com/envoyproxy/gateway/internal/provider/utils"
)

const (
	SdsCAFilename   = "xds-trusted-ca.json"
	SdsCertFilename = "xds-certificate.json"
	// XdsTLSCertFilename is the fully qualified path of the file containing Envoy's
	// xDS server TLS certificate.
	XdsTLSCertFilename = "/certs/tls.crt"
	// XdsTLSKeyFilename is the fully qualified path of the file containing Envoy's
	// xDS server TLS key.
	XdsTLSKeyFilename = "/certs/tls.key"
	// XdsTLSCaFilename is the fully qualified path of the file containing Envoy's
	// trusted CA certificate.
	XdsTLSCaFilename = "/certs/ca.crt"
)

var (
	// xDS certificate rotation is supported by using SDS path-based resource files.
	SdsCAConfigMapData = fmt.Sprintf(`{"resources":[{"@type":"type.googleapis.com/envoy.extensions.transport_sockets.tls.v3.Secret",`+
		`"name":"xds_trusted_ca","validation_context":{"trusted_ca":{"filename":"%s"},`+
		`"match_typed_subject_alt_names":[{"san_type":"DNS","matcher":{"exact":"envoy-gateway"}}]}}]}`, XdsTLSCaFilename)
	SdsCertConfigMapData = fmt.Sprintf(`{"resources":[{"@type":"type.googleapis.com/envoy.extensions.transport_sockets.tls.v3.Secret",`+
		`"name":"xds_certificate","tls_certificate":{"certificate_chain":{"filename":"%s"},`+
		`"private_key":{"filename":"%s"}}}]}`, XdsTLSCertFilename, XdsTLSKeyFilename)
)

func ExpectedResourceHashedName(name string) string {
	hashedName := providerutils.GetHashedName(name)
	return fmt.Sprintf("%s-%s", config.EnvoyPrefix, hashedName)
}

// EnvoyAppLabel returns the labels used for all Envoy resources.
func EnvoyAppLabel() map[string]string {
	return map[string]string{
		"app.gateway.envoyproxy.io/name": "envoy",
	}
}

// EnvoyLabels returns the labels, including extraLbls, used for Envoy resources.
func EnvoyLabels(extraLbls map[string]string) map[string]string {
	lbls := EnvoyAppLabel()
	for k, v := range extraLbls {
		lbls[k] = v
	}

	return lbls
}
