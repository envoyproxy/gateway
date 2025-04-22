// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package utils

import (
	"bytes"

	xdstype "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"k8s.io/utils/ptr"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/yaml"
)

func FromFraction(fraction *gwapiv1.Fraction) *xdstype.Percent {
	if fraction == nil {
		return nil
	}

	v := float64(fraction.Numerator) / float64(ptr.Deref(fraction.Denominator, 100))
	return &xdstype.Percent{
		Value: v,
	}
}

func FromUInt32(val *uint32) *wrapperspb.UInt32Value {
	if val == nil {
		return nil
	}

	return wrapperspb.UInt32(*val)
}

func MarshalResourcesToJSON(resources []types.Resource) ([]byte, error) {
	msgs := make([]proto.Message, 0)
	for _, resource := range resources {
		msgs = append(msgs, resource.(proto.Message))
	}
	var buffer bytes.Buffer
	buffer.WriteByte('[')
	for idx, msg := range msgs {
		if idx != 0 {
			buffer.WriteByte(',')
		}
		b, err := protojson.Marshal(msg)
		if err != nil {
			return nil, err
		}
		buffer.Write(b)
	}
	buffer.WriteByte(']')
	return buffer.Bytes(), nil
}

// ResourcesToYAMLString converts xDS Resource types into YAML string
func ResourcesToYAMLString(resources []types.Resource) (string, error) {
	jsonBytes, err := MarshalResourcesToJSON(resources)
	if err != nil {
		return "", err
	}
	data, err := yaml.JSONToYAML(jsonBytes)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
