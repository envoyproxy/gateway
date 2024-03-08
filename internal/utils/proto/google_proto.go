// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

// Copied from https://github.com/kumahq/kuma/tree/9ea78e31147a855ac54a7a2c92c724ee9a75de46/pkg/util/proto
// to avoid importing the entire kuma codebase breaking our go.mod file

package proto

import (
	"fmt"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/durationpb"
)

type (
	MergeFunction func(dst, src protoreflect.Message)
	mergeOptions  struct {
		customMergeFn map[protoreflect.FullName]MergeFunction
	}
)
type OptionFn func(options mergeOptions) mergeOptions

func MergeFunctionOptionFn(name protoreflect.FullName, function MergeFunction) OptionFn {
	return func(options mergeOptions) mergeOptions {
		options.customMergeFn[name] = function
		return options
	}
}

// ReplaceMergeFn instead of merging all subfields one by one, takes src and set it to dest
var ReplaceMergeFn MergeFunction = func(dst, src protoreflect.Message) {
	dst.Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		dst.Clear(fd)
		return true
	})
	src.Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		dst.Set(fd, v)
		return true
	})
}

func Merge(dst, src proto.Message) {
	duration := &durationpb.Duration{}
	merge(dst, src, MergeFunctionOptionFn(duration.ProtoReflect().Descriptor().FullName(), ReplaceMergeFn))
}

// Merge Code of proto.Merge with modifications to support custom types
func merge(dst, src proto.Message, opts ...OptionFn) {
	mo := mergeOptions{customMergeFn: map[protoreflect.FullName]MergeFunction{}}
	for _, opt := range opts {
		mo = opt(mo)
	}
	mo.mergeMessage(dst.ProtoReflect(), src.ProtoReflect())
}

func (o mergeOptions) mergeMessage(dst, src protoreflect.Message) {
	// The regular proto.mergeMessage would have a fast path method option here.
	// As we want to have exceptions we always use the slow path.
	if !dst.IsValid() {
		panic(fmt.Sprintf("cannot merge into invalid %v message", dst.Descriptor().FullName()))
	}

	src.Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		switch {
		case fd.IsList():
			o.mergeList(dst.Mutable(fd).List(), v.List(), fd)
		case fd.IsMap():
			o.mergeMap(dst.Mutable(fd).Map(), v.Map(), fd.MapValue())
		case fd.Message() != nil:
			mergeFn, exists := o.customMergeFn[fd.Message().FullName()]
			if exists {
				mergeFn(dst.Mutable(fd).Message(), v.Message())
			} else {
				o.mergeMessage(dst.Mutable(fd).Message(), v.Message())
			}
		case fd.Kind() == protoreflect.BytesKind:
			dst.Set(fd, o.cloneBytes(v))
		default:
			dst.Set(fd, v)
		}
		return true
	})

	if len(src.GetUnknown()) > 0 {
		dst.SetUnknown(append(dst.GetUnknown(), src.GetUnknown()...))
	}
}

func (o mergeOptions) mergeList(dst, src protoreflect.List, fd protoreflect.FieldDescriptor) {
	// Merge semantics appends to the end of the existing list.
	for i, n := 0, src.Len(); i < n; i++ {
		switch v := src.Get(i); {
		case fd.Message() != nil:
			dstv := dst.NewElement()
			o.mergeMessage(dstv.Message(), v.Message())
			dst.Append(dstv)
		case fd.Kind() == protoreflect.BytesKind:
			dst.Append(o.cloneBytes(v))
		default:
			dst.Append(v)
		}
	}
}

func (o mergeOptions) mergeMap(dst, src protoreflect.Map, fd protoreflect.FieldDescriptor) {
	// Merge semantics replaces, rather than merges into existing entries.
	src.Range(func(k protoreflect.MapKey, v protoreflect.Value) bool {
		switch {
		case fd.Message() != nil:
			dstv := dst.NewValue()
			o.mergeMessage(dstv.Message(), v.Message())
			dst.Set(k, dstv)
		case fd.Kind() == protoreflect.BytesKind:
			dst.Set(k, o.cloneBytes(v))
		default:
			dst.Set(k, v)
		}
		return true
	})
}

func (o mergeOptions) cloneBytes(v protoreflect.Value) protoreflect.Value {
	return protoreflect.ValueOfBytes(append([]byte{}, v.Bytes()...))
}
