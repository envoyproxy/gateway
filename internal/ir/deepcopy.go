// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package ir

// The certificate payloads below are hand-written rather than generated so that a
// deep copy shares their byte slices instead of duplicating them. The IR is deep
// copied once per store and once per subscriber at the watchable boundary, so a CA
// bundle referenced by many destinations was being held once per destination per
// copy - on a gateway whose backends share one CA that was the bulk of the heap.
//
// This is safe because the bytes are immutable once translation has produced them:
// the translator aliases them straight out of the Secret, and the xDS translator
// only reads them into protobuf InlineBytes. Code that needs to change a
// certificate must replace the slice; appending to one or writing into it would
// now be visible to every other copy, including through spare capacity that the
// copy's own length hides.

// DeepCopyInto copies the receiver into out, sharing the immutable certificate bytes.
func (in *TLSCACertificate) DeepCopyInto(out *TLSCACertificate) {
	*out = *in
	// Delegate rather than copying SDSConfig field by field, so a field added to it
	// later is deep-copied by the regenerated code instead of being silently shared.
	out.SDS = in.SDS.DeepCopy()
}

// DeepCopyInto copies the receiver into out, sharing the immutable certificate,
// private key and OCSP staple bytes.
func (t *TLSCertificate) DeepCopyInto(out *TLSCertificate) {
	*out = *t
	out.SDS = t.SDS.DeepCopy()
}

// DeepCopyInto copies the receiver into out, sharing the immutable CRL bytes.
func (in *TLSCrl) DeepCopyInto(out *TLSCrl) {
	*out = *in
}
