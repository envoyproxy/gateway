// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package proto

import (
	"bytes"
	"testing"

	"google.golang.org/protobuf/types/known/structpb"
)

// TestToAnyWithValidationDeterministic ensures ToAnyWithValidation marshals map fields in a stable
// key order across calls. The golden xDS tests cannot guard this: they compare protojson output,
// which re-sorts map keys, so a non-deterministic marshaler would still pass them.
func TestToAnyWithValidationDeterministic(t *testing.T) {
	// structpb.Struct is a map<string, Value>; enough keys that a non-deterministic marshaler is
	// overwhelmingly likely to vary the byte order between calls.
	fields := make(map[string]*structpb.Value, 24)
	for i := 0; i < 24; i++ {
		fields[string(rune('a'+i))] = structpb.NewStringValue(string(rune('A' + i)))
	}
	msg := &structpb.Struct{Fields: fields}

	first, err := ToAnyWithValidation(msg)
	if err != nil {
		t.Fatalf("ToAnyWithValidation: %v", err)
	}
	for i := 2; i <= 100; i++ {
		got, err := ToAnyWithValidation(msg)
		if err != nil {
			t.Fatalf("ToAnyWithValidation: %v", err)
		}
		if !bytes.Equal(got.GetValue(), first.GetValue()) {
			t.Fatalf("ToAnyWithValidation produced non-deterministic bytes on call %d", i)
		}
	}
}
