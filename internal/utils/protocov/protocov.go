// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package protocov

import (
	"errors"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

const (
	APIPrefix = "type.googleapis.com/"
)

var (
	marshalOpts = proto.MarshalOptions{}
)

func ToAnyWithError(msg proto.Message) (*anypb.Any, error) {
	if msg == nil {
		return nil, errors.New("empty message received")
	}
	b, err := marshalOpts.Marshal(msg)
	if err != nil {
		return nil, err
	}
	return &anypb.Any{
		TypeUrl: APIPrefix + string(msg.ProtoReflect().Descriptor().FullName()),
		Value:   b,
	}, nil
}

func ToAny(msg proto.Message) *anypb.Any {
	res, err := ToAnyWithError(msg)
	if err != nil {
		return nil
	}
	return res
}
