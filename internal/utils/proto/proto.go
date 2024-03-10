// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

// Copied from https://github.com/kumahq/kuma/tree/9ea78e31147a855ac54a7a2c92c724ee9a75de46/pkg/util/proto
// to avoid importing the entire kuma codebase breaking our go.mod file

package proto

import (
	"bytes"

	"github.com/golang/protobuf/jsonpb"
	protov1 "github.com/golang/protobuf/proto"
	"google.golang.org/protobuf/proto"
	"sigs.k8s.io/yaml"
)

func FromYAML(content []byte, pb proto.Message) error {
	json, err := yaml.YAMLToJSON(content)
	if err != nil {
		return err
	}
	return FromJSON(json, pb)
}

func ToYAML(pb proto.Message) ([]byte, error) {
	marshaler := &jsonpb.Marshaler{}
	json, err := marshaler.MarshalToString(protov1.MessageV1(pb))
	if err != nil {
		return nil, err
	}
	return yaml.JSONToYAML([]byte(json))
}

func FromJSON(content []byte, out proto.Message) error {
	unmarshaler := &jsonpb.Unmarshaler{AllowUnknownFields: true}
	return unmarshaler.Unmarshal(bytes.NewReader(content), protov1.MessageV1(out))
}
