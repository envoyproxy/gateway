// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/require"
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

const grpcEchoService = "gateway_api_conformance.echo_basic.grpcecho.GrpcEcho"

// grpcEchoDescriptorB64 is the descriptor set of the conformance grpcecho service,
// generated with --include_imports.
var grpcEchoDescriptorB64 = func() string {
	b, err := os.ReadFile(filepath.Join("testdata", "grpcecho-descriptor.b64"))
	if err != nil {
		panic(err)
	}
	return strings.TrimSpace(string(b))
}()

func grpcEchoDescriptorBin(t *testing.T) []byte {
	t.Helper()
	bin, err := base64.StdEncoding.DecodeString(grpcEchoDescriptorB64)
	require.NoError(t, err)
	return bin
}

func grpcEchoDescriptorSet(t *testing.T) *descriptorpb.FileDescriptorSet {
	t.Helper()
	fds := &descriptorpb.FileDescriptorSet{}
	require.NoError(t, proto.Unmarshal(grpcEchoDescriptorBin(t), fds))
	return fds
}

func configMap(name string, data map[string]string, binary map[string][]byte) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: name},
		Data:       data,
		BinaryData: binary,
	}
}

func protoDescriptorRef(name string) egv1a1.ProtoDescriptor {
	return egv1a1.ProtoDescriptor{
		ValueRef: gwapiv1.LocalObjectReference{Kind: "ConfigMap", Name: gwapiv1.ObjectName(name)},
	}
}

// loadDescriptorSet loads the descriptor set from a ConfigMap, like a route does.
func loadDescriptorSet(t *testing.T, fds *descriptorpb.FileDescriptorSet) (*parsedProtoDescriptor, error) {
	t.Helper()
	bin, err := proto.Marshal(fds)
	require.NoError(t, err)

	tr := &Translator{TranslatorContext: &TranslatorContext{}}
	tr.SetConfigMaps([]*corev1.ConfigMap{configMap("d", nil, map[string][]byte{"proto-descriptor": bin})})
	return tr.loadProtoDescriptor(protoDescriptorRef("d"), "default")
}

func TestLoadProtoDescriptor(t *testing.T) {
	bin := grpcEchoDescriptorBin(t)

	tr := &Translator{TranslatorContext: &TranslatorContext{}}
	tr.SetConfigMaps([]*corev1.ConfigMap{
		configMap("binary-data", nil, map[string][]byte{"proto-descriptor": bin}),
		configMap("data", map[string]string{"proto-descriptor": grpcEchoDescriptorB64}, nil),
		configMap("multi-line-data", map[string]string{
			"proto-descriptor": grpcEchoDescriptorB64[:100] + "\n  " + grpcEchoDescriptorB64[100:],
		}, nil),
		configMap("only-binary-data", nil, map[string][]byte{"echo.pb": bin}),
		configMap("only-data", map[string]string{"echo.pb": grpcEchoDescriptorB64}, nil),
		configMap("no-key", map[string]string{"a": grpcEchoDescriptorB64}, map[string][]byte{"b": bin}),
		configMap("not-base64", map[string]string{"proto-descriptor": "not base64!!"}, nil),
		configMap("raw-data", map[string]string{"proto-descriptor": string(utf8Descriptor(t))}, nil),
		configMap("not-descriptor", map[string]string{
			"proto-descriptor": base64.StdEncoding.EncodeToString([]byte("not a descriptor")),
		}, nil),
	})

	tests := []struct {
		name    string
		ref     egv1a1.ProtoDescriptor
		wantErr string
	}{
		{name: "binaryData", ref: protoDescriptorRef("binary-data")},
		{name: "base64 data", ref: protoDescriptorRef("data")},
		{name: "multi-line base64 data", ref: protoDescriptorRef("multi-line-data")},
		{name: "the only binaryData entry", ref: protoDescriptorRef("only-binary-data")},
		{
			name:    "the only data entry is not used",
			ref:     protoDescriptorRef("only-data"),
			wantErr: "can't find the key proto-descriptor",
		},
		{
			name:    "no proto-descriptor key",
			ref:     protoDescriptorRef("no-key"),
			wantErr: "can't find the key proto-descriptor",
		},
		{name: "neither base64 nor descriptor", ref: protoDescriptorRef("not-base64"), wantErr: "failed to parse the proto descriptor"},
		{name: "invalid descriptor", ref: protoDescriptorRef("not-descriptor"), wantErr: "failed to parse the proto descriptor"},
		{name: "configmap not found", ref: protoDescriptorRef("not-found"), wantErr: "can't find the referenced configmap"},
		{
			name: "not a configmap",
			ref: egv1a1.ProtoDescriptor{
				ValueRef: gwapiv1.LocalObjectReference{Kind: "Secret", Name: "data"},
			},
			wantErr: "only ConfigMap is supported",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			d, err := tr.loadProtoDescriptor(tc.ref, "default")
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, bin, d.bin)
		})
	}

	// kubectl create configmap --from-file stores a descriptor set that is valid UTF-8 in data.
	t.Run("raw data", func(t *testing.T) {
		d, err := tr.loadProtoDescriptor(protoDescriptorRef("raw-data"), "default")
		require.NoError(t, err)
		require.Equal(t, utf8Descriptor(t), d.bin)
	})
}

// utf8Descriptor returns a small descriptor set that is valid UTF-8.
func utf8Descriptor(t *testing.T) []byte {
	t.Helper()
	bin, err := proto.Marshal(&descriptorpb.FileDescriptorSet{File: []*descriptorpb.FileDescriptorProto{{
		Name:        proto.String("a.proto"),
		Package:     proto.String("t"),
		Syntax:      proto.String("proto3"),
		MessageType: []*descriptorpb.DescriptorProto{{Name: proto.String("M")}},
		Service: []*descriptorpb.ServiceDescriptorProto{{
			Name: proto.String("S"),
			Method: []*descriptorpb.MethodDescriptorProto{{
				Name:       proto.String("P"),
				InputType:  proto.String(".t.M"),
				OutputType: proto.String(".t.M"),
			}},
		}},
	}}})
	require.NoError(t, err)
	require.True(t, utf8.Valid(bin))
	return bin
}

func TestLoadProtoDescriptorCache(t *testing.T) {
	tr := &Translator{TranslatorContext: &TranslatorContext{}}
	tr.SetConfigMaps([]*corev1.ConfigMap{
		configMap("valid", map[string]string{"proto-descriptor": grpcEchoDescriptorB64}, nil),
		configMap("invalid", map[string]string{"proto-descriptor": "bm90IGEgZGVzY3JpcHRvcg=="}, nil),
	})

	first, err := tr.loadProtoDescriptor(protoDescriptorRef("valid"), "default")
	require.NoError(t, err)
	second, err := tr.loadProtoDescriptor(protoDescriptorRef("valid"), "default")
	require.NoError(t, err)
	require.Same(t, first, second)

	// Errors are cached too.
	_, err1 := tr.loadProtoDescriptor(protoDescriptorRef("invalid"), "default")
	_, err2 := tr.loadProtoDescriptor(protoDescriptorRef("invalid"), "default")
	require.Error(t, err1)
	require.Equal(t, err1, err2)
	require.Len(t, tr.protoDescriptors, 2)
}

func TestParseProtoDescriptor(t *testing.T) {
	t.Run("valid descriptor", func(t *testing.T) {
		d, err := loadDescriptorSet(t, grpcEchoDescriptorSet(t))
		require.NoError(t, err)
		require.Equal(t, []string{grpcEchoService}, d.defaultServices)
	})

	t.Run("missing import", func(t *testing.T) {
		fds := grpcEchoDescriptorSet(t)
		fds.File = fds.File[len(fds.File)-1:]
		_, err := loadDescriptorSet(t, fds)
		require.ErrorContains(t, err, "imported by grpcecho.proto is missing in the proto descriptor")
	})

	t.Run("import listed after the file importing it", func(t *testing.T) {
		fds := grpcEchoDescriptorSet(t)
		slices.Reverse(fds.File)
		_, err := loadDescriptorSet(t, fds)
		require.ErrorContains(t, err, "grpcecho.proto is listed before its import")
	})

	t.Run("unknown type", func(t *testing.T) {
		fds := &descriptorpb.FileDescriptorSet{File: []*descriptorpb.FileDescriptorProto{{
			Name:        proto.String("echo.proto"),
			Package:     proto.String("test.v1"),
			Syntax:      proto.String("proto3"),
			MessageType: []*descriptorpb.DescriptorProto{{Name: proto.String("Resp")}},
			Service: []*descriptorpb.ServiceDescriptorProto{{
				Name: proto.String("Echo"),
				Method: []*descriptorpb.MethodDescriptorProto{{
					Name:       proto.String("Ping"),
					InputType:  proto.String(".test.v1.Req"),
					OutputType: proto.String(".test.v1.Resp"),
				}},
			}},
		}}}
		_, err := loadDescriptorSet(t, fds)
		require.ErrorContains(t, err, "invalid proto descriptor")
		require.ErrorContains(t, err, "test.v1.Req")
	})

	t.Run("no service", func(t *testing.T) {
		fds := &descriptorpb.FileDescriptorSet{File: []*descriptorpb.FileDescriptorProto{{
			Name:   proto.String("empty.proto"),
			Syntax: proto.String("proto3"),
		}}}
		_, err := loadDescriptorSet(t, fds)
		require.ErrorContains(t, err, "no gRPC service found")
	})
}

func TestGetTranscodedServices(t *testing.T) {
	d, err := loadDescriptorSet(t, grpcEchoDescriptorSet(t))
	require.NoError(t, err)

	t.Run("default services", func(t *testing.T) {
		got, err := getTranscodedServices(d, nil)
		require.NoError(t, err)
		require.Equal(t, []string{grpcEchoService}, got)
	})

	t.Run("duplicated services are removed", func(t *testing.T) {
		got, err := getTranscodedServices(d, []string{grpcEchoService, grpcEchoService})
		require.NoError(t, err)
		require.Equal(t, []string{grpcEchoService}, got)
	})

	t.Run("unknown service", func(t *testing.T) {
		_, err := getTranscodedServices(d, []string{"test.v1.Unknown"})
		require.ErrorContains(t, err, "can't find service test.v1.Unknown")
	})

	t.Run("only the first services are listed in the error", func(t *testing.T) {
		services := sets.New[string]()
		for i := range 25 {
			services.Insert(fmt.Sprintf("test.v1.Service%02d", i))
		}
		_, err := getTranscodedServices(&parsedProtoDescriptor{services: services}, []string{"test.v1.Unknown"})
		require.ErrorContains(t, err, "and 15 more")
		require.NotContains(t, err.Error(), "Service24")
	})
}

func TestBuildGRPCJSONTranscoder(t *testing.T) {
	tr := &Translator{TranslatorContext: &TranslatorContext{}}
	tr.SetConfigMaps([]*corev1.ConfigMap{
		configMap("descriptor", map[string]string{"proto-descriptor": grpcEchoDescriptorB64}, nil),
	})

	got, err := tr.buildGRPCJSONTranscoder(&egv1a1.GRPCJSONTranscoder{
		ProtoDescriptor: protoDescriptorRef("descriptor"),
	}, "httproutefilter/default/transcode", "default")
	require.NoError(t, err)
	require.Equal(t, "httproutefilter/default/transcode", got.Name)
	require.Equal(t, grpcEchoDescriptorBin(t), got.ProtoDescriptorBin)
	require.Equal(t, []string{grpcEchoService}, got.Services)
}

func TestValidateHTTPRules(t *testing.T) {
	field := func(name string, number int32, typeName string) *descriptorpb.FieldDescriptorProto {
		f := &descriptorpb.FieldDescriptorProto{
			Name:   proto.String(name),
			Number: proto.Int32(number),
			Label:  descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
			Type:   descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
		}
		if typeName != "" {
			f.Type = descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum()
			f.TypeName = proto.String(typeName)
		}
		return f
	}
	message := func(name string, fields ...*descriptorpb.FieldDescriptorProto) *descriptorpb.DescriptorProto {
		return &descriptorpb.DescriptorProto{Name: proto.String(name), Field: fields}
	}
	// descriptorSet returns a descriptor set with one method using the rule.
	descriptorSet := func(rule *annotations.HttpRule) *descriptorpb.FileDescriptorSet {
		opts := &descriptorpb.MethodOptions{}
		proto.SetExtension(opts, annotations.E_Http, rule)
		return &descriptorpb.FileDescriptorSet{File: []*descriptorpb.FileDescriptorProto{
			{
				Name:        proto.String("google/api/httpbody.proto"),
				Package:     proto.String("google.api"),
				Syntax:      proto.String("proto3"),
				MessageType: []*descriptorpb.DescriptorProto{message("HttpBody")},
			},
			{
				Name:       proto.String("echo.proto"),
				Package:    proto.String("test.v1"),
				Syntax:     proto.String("proto3"),
				Dependency: []string{"google/api/httpbody.proto"},
				MessageType: []*descriptorpb.DescriptorProto{
					message("Inner", field("name", 1, "")),
					message("Req", field("user_name", 1, ""), field("inner", 2, ".test.v1.Inner")),
					message("Resp", field("text", 1, ""), field("raw", 2, ".google.api.HttpBody")),
				},
				Service: []*descriptorpb.ServiceDescriptorProto{{
					Name: proto.String("Echo"),
					Method: []*descriptorpb.MethodDescriptorProto{{
						Name:       proto.String("Ping"),
						InputType:  proto.String(".test.v1.Req"),
						OutputType: proto.String(".test.v1.Resp"),
						Options:    opts,
					}},
				}},
			},
		}}
	}
	post := func(body, responseBody string) *annotations.HttpRule {
		return &annotations.HttpRule{
			Pattern:      &annotations.HttpRule_Post{Post: "/v1/ping"},
			Body:         body,
			ResponseBody: responseBody,
		}
	}

	tests := []struct {
		name    string
		rule    *annotations.HttpRule
		wantErr string
	}{
		{name: "whole message as body", rule: post("*", "")},
		{name: "body by JSON name", rule: post("userName", "")},
		{name: "body by proto name", rule: post("user_name", "")},
		{name: "nested body", rule: post("inner.name", "")},
		{name: "HttpBody response_body", rule: post("", "raw")},
		{name: "unknown body", rule: post("missing", ""), wantErr: `invalid body "missing" of method test.v1.Echo.Ping`},
		{name: "body under a string field", rule: post("user_name.x", ""), wantErr: "user_name is not a message"},
		{name: "unknown response_body", rule: post("", "missing"), wantErr: `invalid response_body "missing"`},
		{name: "response_body not HttpBody", rule: post("", "text"), wantErr: "must be a google.api.HttpBody field"},
		{
			name:    "invalid path template",
			rule:    &annotations.HttpRule{Pattern: &annotations.HttpRule_Get{Get: "v1/ping"}},
			wantErr: `invalid path template "v1/ping" of method test.v1.Echo.Ping`,
		},
		{
			name: "invalid path template in an additional binding",
			rule: &annotations.HttpRule{
				Pattern: &annotations.HttpRule_Get{Get: "/v1/ping"},
				AdditionalBindings: []*annotations.HttpRule{
					{Pattern: &annotations.HttpRule_Custom{Custom: &annotations.CustomHttpPattern{Kind: "HEAD", Path: "/v1/{name"}}},
				},
			},
			wantErr: `invalid path template "/v1/{name"`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			d, err := loadDescriptorSet(t, descriptorSet(tc.rule))
			require.NoError(t, err)
			err = validateHTTPRules(d.files, []string{"test.v1.Echo"})
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestIsValidPathTemplate(t *testing.T) {
	for _, path := range []string{
		"/",
		"/v1",
		"/v1/users/{name}",
		"/v1/{name=users/*}/books/{book.id}",
		"/v1/{name=shelves/*/books/*}:publish",
		"/v1/*/ping",
		"/v1/**",
		"/v1/{path=**}",
		"/v1/{path=**}/raw",
		"/v1/**/raw:get",
		"/v1/a-b.c~d%20e=f",
		"/v1/a*b",
		"/v1/a{b",
		"/v1/{9name}",
	} {
		require.True(t, isValidPathTemplate(path), path)
	}

	for _, path := range []string{
		"",
		"v1/ping",
		"/v1/",
		"/v1//ping",
		"/v1/{name",
		"/v1/{}",
		"/v1/{name=}",
		"/v1/{a={b}}",
		"/v1/**/*",
		"/v1/**/**",
		"/v1/**/{name}",
		"/v1/ping:",
		"/v1/ping:a:b",
		"/v1/ping:a/b",
		"/v1/***",
	} {
		require.False(t, isValidPathTemplate(path), path)
	}
}

func TestValidateGRPCStatusProtos(t *testing.T) {
	message := func(name string) *descriptorpb.DescriptorProto {
		return &descriptorpb.DescriptorProto{Name: proto.String(name)}
	}
	file := func(name, pkg string, deps []string, messages ...*descriptorpb.DescriptorProto) *descriptorpb.FileDescriptorProto {
		return &descriptorpb.FileDescriptorProto{
			Name:        proto.String(name),
			Package:     proto.String(pkg),
			Syntax:      proto.String("proto3"),
			Dependency:  deps,
			MessageType: messages,
		}
	}
	service := func(deps ...string) *descriptorpb.FileDescriptorProto {
		f := file("echo.proto", "test.v1", deps, message("Req"), message("Resp"))
		f.Service = []*descriptorpb.ServiceDescriptorProto{{
			Name: proto.String("Echo"),
			Method: []*descriptorpb.MethodDescriptorProto{{
				Name:       proto.String("Ping"),
				InputType:  proto.String(".test.v1.Req"),
				OutputType: proto.String(".test.v1.Resp"),
			}},
		}}
		return f
	}
	build := func(t *testing.T, convertGRPCStatus bool, files ...*descriptorpb.FileDescriptorProto) error {
		t.Helper()
		bin, err := proto.Marshal(&descriptorpb.FileDescriptorSet{File: files})
		require.NoError(t, err)

		tr := &Translator{TranslatorContext: &TranslatorContext{}}
		tr.SetConfigMaps([]*corev1.ConfigMap{configMap("d", nil, map[string][]byte{"proto-descriptor": bin})})
		_, err = tr.buildGRPCJSONTranscoder(&egv1a1.GRPCJSONTranscoder{
			ProtoDescriptor:   protoDescriptorRef("d"),
			ConvertGRPCStatus: &convertGRPCStatus,
		}, "n", "default")
		return err
	}
	// google.protobuf.Any in another file, as some repos vendor it.
	vendoredAny := file("third_party/google/protobuf/any.proto", "google.protobuf", nil, message("Any"))

	t.Run("descriptor without any.proto and status.proto", func(t *testing.T) {
		require.NoError(t, build(t, true, service()))
	})
	t.Run("google.protobuf.Any in another file", func(t *testing.T) {
		require.ErrorContains(t, build(t, true, vendoredAny, service(vendoredAny.GetName())),
			"Envoy adds google/rpc/status.proto, but google/protobuf/any.proto imported by it is missing")
	})
	t.Run("google.protobuf.Any in another file without convertGRPCStatus", func(t *testing.T) {
		require.NoError(t, build(t, false, vendoredAny, service(vendoredAny.GetName())))
	})
	t.Run("another file named google/rpc/status.proto", func(t *testing.T) {
		require.ErrorContains(t, build(t, true, file("google/rpc/status.proto", "other.v1", nil, message("Other")), service()),
			"they conflict with it")
	})
}

func TestProtoErrorMessage(t *testing.T) {
	for _, prefix := range []string{"proto: ", "proto: "} {
		require.Equal(t, "bad input", protoErrorMessage(errors.New(prefix+"bad input")))
	}
	require.Equal(t, "bad input", protoErrorMessage(errors.New("bad input")))
}
