// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

// Copied from https://github.com/kumahq/kuma/tree/9ea78e31147a855ac54a7a2c92c724ee9a75de46/pkg/util/proto
// to avoid importing the entire kuma codebase breaking our go.mod file

package proto

import (
	"bytes"
	"errors"

	"github.com/golang/protobuf/jsonpb"
	protov1 "github.com/golang/protobuf/proto"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"sigs.k8s.io/yaml"

	_ "github.com/envoyproxy/gateway/internal/xds/extensions" // DON'T REMOVE: import of all extensions
)

var (
	// Setting the OrigName flag to true will preserve the expected snake case field names in the JSON output.
	// Otherwise, camel case is produced, and it causes issues with the func-e library used to unmarshal the
	// bootstrap configuration.
	marshaler   = &jsonpb.Marshaler{OrigName: true}
	unmarshaler = &jsonpb.Unmarshaler{AllowUnknownFields: true}
)

func FromYAML(content []byte, pb proto.Message) error {
	json, err := yaml.YAMLToJSON(content)
	if err != nil {
		return err
	}
	return FromJSON(json, pb)
}

func ToYAML(pb proto.Message) ([]byte, error) {
	json, err := marshaler.MarshalToString(protov1.MessageV1(pb))
	if err != nil {
		return nil, err
	}
	return yaml.JSONToYAML([]byte(json))
}

func FromJSON(content []byte, out proto.Message) error {
	return unmarshaler.Unmarshal(bytes.NewReader(content), protov1.MessageV1(out))
}

func ToAnyWithValidation(msg proto.Message) (*anypb.Any, error) {
	if msg == nil {
		return nil, errors.New("empty message received")
	}

	// If the message has a ValidateAll method, call it before marshaling.
	if err := Validate(msg); err != nil {
		return nil, err
	}

	any, err := anypb.New(msg)
	if err != nil {
		return nil, err
	}
	return any, nil
}

// Validate validates the given message by calling its ValidateAll or Validate methods.
func Validate(msg proto.Message) error {
	// If the message has a ValidateAll method, call it
	if validator, ok := msg.(interface{ ValidateAll() error }); ok {
		return validator.ValidateAll()
	}

	// If the message has a Validate method, call it
	if validator, ok := msg.(interface{ Validate() error }); ok {
		return validator.Validate()
	}
	return nil
}
