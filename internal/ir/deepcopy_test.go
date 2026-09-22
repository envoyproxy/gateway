// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package ir

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The certificate payloads are deliberately shared by a deep copy rather than
// duplicated; see internal/ir/deepcopy.go. These tests fail loudly if a
// regeneration or refactor reintroduces the copy, which on a gateway whose
// backends share one CA costs gigabytes.

func TestTLSCACertificateDeepCopySharesCertificateBytes(t *testing.T) {
	in := &TLSCACertificate{
		Name:        "ca",
		Certificate: []byte("ca-cert"),
		SDS:         &SDSConfig{SecretName: "sds"},
	}

	out := in.DeepCopy()

	require.Equal(t, in.Certificate, out.Certificate)
	assert.Same(t, &in.Certificate[0], &out.Certificate[0])
	assert.NotSame(t, in.SDS, out.SDS)
}

func TestTLSCertificateDeepCopySharesKeyMaterial(t *testing.T) {
	in := &TLSCertificate{
		Name:        "cert",
		Certificate: []byte("cert"),
		PrivateKey:  PrivateBytes("key"),
		OCSPStaple:  []byte("staple"),
		SDS:         &SDSConfig{SecretName: "sds"},
	}

	out := in.DeepCopy()

	require.Equal(t, in.Certificate, out.Certificate)
	assert.Same(t, &in.Certificate[0], &out.Certificate[0])
	assert.Same(t, &in.PrivateKey[0], &out.PrivateKey[0])
	assert.Same(t, &in.OCSPStaple[0], &out.OCSPStaple[0])
	assert.NotSame(t, in.SDS, out.SDS)
}

// SDS is copied by delegating to its generated DeepCopy, which is nil-safe.
func TestTLSDeepCopyHandlesNilSDS(t *testing.T) {
	ca := (&TLSCACertificate{Name: "ca", Certificate: []byte("ca-cert")}).DeepCopy()
	assert.Nil(t, ca.SDS)

	cert := (&TLSCertificate{Name: "cert", Certificate: []byte("cert")}).DeepCopy()
	assert.Nil(t, cert.SDS)
}

func TestTLSCrlDeepCopySharesData(t *testing.T) {
	in := &TLSCrl{Name: "crl", Data: []byte("crl")}

	out := in.DeepCopy()

	require.Equal(t, in.Data, out.Data)
	assert.Same(t, &in.Data[0], &out.Data[0])
}

// The sharing must survive a copy of the whole IR, which is what the watchable
// map does once per store and once per subscriber.
func TestXdsDeepCopySharesUpstreamCertificateBytes(t *testing.T) {
	ca := []byte("shared-ca")
	in := &Xds{
		HTTP: []*HTTPListener{{
			CoreListenerDetails: CoreListenerDetails{Name: "listener"},
			Routes: []*HTTPRoute{{
				Name: "route",
				Destination: &RouteDestination{
					Name: "dest",
					Settings: []*DestinationSetting{{
						Name: "setting",
						TLS:  &TLSUpstreamConfig{CACertificate: &TLSCACertificate{Name: "ca", Certificate: ca}},
					}},
				},
			}},
		}},
	}

	out := in.DeepCopy()

	got := out.HTTP[0].Routes[0].Destination.Settings[0].TLS.CACertificate.Certificate
	require.Equal(t, ca, got)
	assert.Same(t, &ca[0], &got[0])
}

// The hand-written DeepCopyInto methods copy these structs with *out = *in and then
// deep-copy only SDS. That is correct only while every remaining field is a scalar or
// one of the byte slices that is shared on purpose. A reference-typed field added later
// would be aliased silently, where the generated code would have deep-copied it, so pin
// the shape: when this fails, decide what the new field needs and update DeepCopyInto.
func TestTLSDeepCopyTypesHaveNoUnexpectedFields(t *testing.T) {
	cases := []struct {
		value  any
		fields []string
	}{
		{TLSCACertificate{}, []string{
			"Name string",
			"Certificate []uint8",
			"SDS *ir.SDSConfig",
		}},
		{TLSCertificate{}, []string{
			"Name string",
			"SDS *ir.SDSConfig",
			"Certificate []uint8",
			"PrivateKey ir.PrivateBytes",
			"OCSPStaple []uint8",
		}},
		{TLSCrl{}, []string{
			"Name string",
			"Data []uint8",
			"OnlyVerifyLeafCertificate bool",
		}},
	}

	for _, tc := range cases {
		typ := reflect.TypeOf(tc.value)
		t.Run(typ.Name(), func(t *testing.T) {
			got := make([]string, 0, typ.NumField())
			for i := range typ.NumField() {
				f := typ.Field(i)
				got = append(got, f.Name+" "+f.Type.String())
			}
			assert.Equal(t, tc.fields, got)
		})
	}
}
