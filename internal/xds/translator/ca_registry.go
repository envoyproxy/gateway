// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import "github.com/envoyproxy/gateway/internal/ir"

// caCertificateIndex resolves a TLSCACertificate's Digest against ir.Xds.CACertificates.
type caCertificateIndex map[string][]byte

func newCACertificateIndex(xdsIR *ir.Xds) caCertificateIndex {
	idx := make(caCertificateIndex, len(xdsIR.CACertificates))
	for _, ca := range xdsIR.CACertificates {
		idx[ca.Digest] = ca.Certificate
	}
	return idx
}

// resolveCACertificate returns ca's bytes: ca.Certificate directly when carried inline (no
// gateway scope was available to register against), otherwise the shared entry it names.
func resolveCACertificate(ca *ir.TLSCACertificate, idx caCertificateIndex) []byte {
	if ca == nil {
		return nil
	}
	if len(ca.Certificate) > 0 {
		return ca.Certificate
	}
	return idx[ca.Digest]
}
