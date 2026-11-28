// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import "github.com/envoyproxy/gateway/internal/ir"

// clientCertificateIndex resolves a name-only TLSCertificate against ir.Xds.ClientCertificates.
type clientCertificateIndex map[string]*ir.TLSCertificate

func newClientCertificateIndex(xdsIR *ir.Xds) clientCertificateIndex {
	idx := make(clientCertificateIndex, len(xdsIR.ClientCertificates))
	for _, cert := range xdsIR.ClientCertificates {
		idx[cert.Name] = cert
	}
	return idx
}

// resolveClientCertificate returns cert itself when it carries its own bytes or is served by
// SDS or an extension, otherwise the shared entry it names. A name with no entry also falls
// back to cert.
func resolveClientCertificate(cert *ir.TLSCertificate, idx clientCertificateIndex) *ir.TLSCertificate {
	if cert.SDS != nil || cert.ExtensionRef != nil || len(cert.Certificate) > 0 {
		return cert
	}
	if shared, ok := idx[cert.Name]; ok {
		return shared
	}
	return cert
}
