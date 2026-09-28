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
	"strings"
	"testing"

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

// The conformance grpcecho FileDescriptorSet, generated with --include_imports.
var grpcEchoDescriptorB64 = func() string {
	b, err := os.ReadFile(filepath.Join("testdata", "grpcecho-descriptor.b64"))
	if err != nil {
		panic(err)
	}
	return strings.TrimSpace(string(b))
}()

const grpcEchoService = "gateway_api_conformance.echo_basic.grpcecho.GrpcEcho"

func grpcEchoDescriptorBin(t *testing.T) []byte {
	t.Helper()
	bin, err := base64.StdEncoding.DecodeString(grpcEchoDescriptorB64)
	require.NoError(t, err)
	return bin
}

func configMap(name string, data map[string]string, binary map[string][]byte) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: name},
		Data:       data,
		BinaryData: binary,
	}
}

func TestLoadProtoDescriptor(t *testing.T) {
	bin := grpcEchoDescriptorBin(t)

	tr := &Translator{TranslatorContext: &TranslatorContext{}}
	tr.SetConfigMaps([]*corev1.ConfigMap{
		configMap("keyed-binary", nil, map[string][]byte{"proto-descriptor": bin}),
		configMap("keyed-data", map[string]string{"proto-descriptor": grpcEchoDescriptorB64}, nil),
		configMap("sole-binary", nil, map[string][]byte{"anything.pb": bin}),
		configMap("sole-data", map[string]string{"anything.pb": grpcEchoDescriptorB64}, nil),
		configMap("ambiguous", map[string]string{"a": grpcEchoDescriptorB64, "b": grpcEchoDescriptorB64}, nil),
		configMap("split", map[string]string{"a": grpcEchoDescriptorB64}, map[string][]byte{"b": bin}),
		configMap("garbage", map[string]string{"proto-descriptor": "not base64!!"}, nil),
		// YAML block scalars fold in newlines, so the decoder must tolerate whitespace.
		configMap("wrapped", map[string]string{
			"proto-descriptor": grpcEchoDescriptorB64[:100] + "\n  " + grpcEchoDescriptorB64[100:],
		}, nil),
	})

	valueRef := func(name string) egv1a1.ProtoDescriptor {
		return egv1a1.ProtoDescriptor{
			ValueRef: gwapiv1.LocalObjectReference{Kind: "ConfigMap", Name: gwapiv1.ObjectName(name)},
		}
	}

	tests := []struct {
		name    string
		desc    egv1a1.ProtoDescriptor
		want    []byte
		wantErr string
	}{
		{name: "configmap binaryData used as-is", desc: valueRef("keyed-binary"), want: bin},
		{name: "configmap data is base64 decoded", desc: valueRef("keyed-data"), want: bin},
		{name: "configmap sole binaryData entry", desc: valueRef("sole-binary"), want: bin},
		{
			// The informer cache trims Data to its first key, so a sole Data entry means
			// something different in-cluster than it does offline. Only the named key is
			// honoured there, or the providers disagree on the same ConfigMap.
			name:    "configmap sole data entry is rejected",
			desc:    valueRef("sole-data"),
			wantErr: "expected key",
		},
		{name: "configmap data tolerates folded whitespace", desc: valueRef("wrapped"), want: bin},
		{
			name:    "configmap with several entries and no known key is rejected",
			desc:    valueRef("ambiguous"),
			wantErr: "expected key",
		},
		{
			name:    "one entry in each of data and binaryData is still ambiguous",
			desc:    valueRef("split"),
			wantErr: "expected key",
		},
		{name: "non base64 is rejected", desc: valueRef("garbage"), wantErr: "not valid base64"},
		{name: "missing configmap", desc: valueRef("nope"), wantErr: "not found"},
		{name: "unsupported kind is rejected", desc: egv1a1.ProtoDescriptor{
			ValueRef: gwapiv1.LocalObjectReference{Kind: "Secret", Name: "keyed-data"},
		}, wantErr: "only ConfigMap is supported"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			d, err := tr.loadProtoDescriptor(tc.desc, "default")
			var got []byte
			if d != nil {
				got = d.bin
			}
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}

func TestResolveTranscodedServices(t *testing.T) {
	bin := grpcEchoDescriptorBin(t)

	t.Run("empty list is expanded, not passed through", func(t *testing.T) {
		// An empty list would leave the filter disabled in Envoy.
		got, err := resolveTranscodedServices(mustParse(t, bin), nil)
		require.NoError(t, err)
		require.Equal(t, []string{grpcEchoService}, got)
	})

	t.Run("explicit list is preserved", func(t *testing.T) {
		got, err := resolveTranscodedServices(mustParse(t, bin), []string{grpcEchoService})
		require.NoError(t, err)
		require.Equal(t, []string{grpcEchoService}, got)
	})

	t.Run("unknown service is rejected", func(t *testing.T) {
		_, err := resolveTranscodedServices(mustParse(t, bin), []string{"does.not.Exist"})
		require.ErrorContains(t, err, `service "does.not.Exist" not found`)
	})

	t.Run("available service list is bounded for status conditions", func(t *testing.T) {
		all := sets.New[string]()
		for i := range 25 {
			all.Insert(fmt.Sprintf("pkg.v1.Service%02d", i))
		}
		_, err := resolveTranscodedServices(&parsedProtoDescriptor{all: all}, []string{"nope"})
		require.ErrorContains(t, err, "and 15 more")
		require.NotContains(t, err.Error(), "Service24")
	})

	t.Run("garbage descriptor is rejected", func(t *testing.T) {
		tr := &Translator{TranslatorContext: &TranslatorContext{}}
		tr.SetConfigMaps([]*corev1.ConfigMap{
			configMap("garbage", map[string]string{
				"proto-descriptor": base64.StdEncoding.EncodeToString([]byte("not a descriptor set")),
			}, nil),
		})
		_, err := tr.loadProtoDescriptor(egv1a1.ProtoDescriptor{
			ValueRef: gwapiv1.LocalObjectReference{Kind: "ConfigMap", Name: "garbage"},
		}, "default")
		require.ErrorContains(t, err, "FileDescriptorSet")
	})
}

func TestBuildGRPCJSONTranscoder(t *testing.T) {
	hrf := &egv1a1.HTTPRouteFilter{
		ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "transcode"},
		Spec: egv1a1.HTTPRouteFilterSpec{
			GRPCJSONTranscoder: &egv1a1.GRPCJSONTranscoder{
				ProtoDescriptor: egv1a1.ProtoDescriptor{
					ValueRef: gwapiv1.LocalObjectReference{Kind: "ConfigMap", Name: "descriptor"},
				},
			},
		},
	}
	hrf.SetGroupVersionKind(egv1a1.GroupVersion.WithKind(egv1a1.KindHTTPRouteFilter))

	tr := &Translator{TranslatorContext: &TranslatorContext{}}
	tr.SetConfigMaps([]*corev1.ConfigMap{
		configMap("descriptor", map[string]string{"proto-descriptor": grpcEchoDescriptorB64}, nil),
	})

	got, err := tr.buildGRPCJSONTranscoder(hrf.Spec.GRPCJSONTranscoder, irConfigName(hrf), hrf.Namespace)
	require.NoError(t, err)
	require.Equal(t, "httproutefilter/default/transcode", got.Name)
	require.Equal(t, grpcEchoDescriptorBin(t), got.ProtoDescriptorBin)
	require.Equal(t, []string{grpcEchoService}, got.Services)
}

// Guards the failure Envoy reports only as "Unable to build proto descriptor pool".
func TestValidateDescriptorClosure(t *testing.T) {
	t.Run("complete descriptor passes", func(t *testing.T) {
		fds := &descriptorpb.FileDescriptorSet{}
		require.NoError(t, proto.Unmarshal(grpcEchoDescriptorBin(t), fds))
		require.NoError(t, validateDescriptorClosure(fds))
		require.Greater(t, len(fds.GetFile()), 1, "must carry its imports, not just grpcecho.proto")
	})

	t.Run("dangling import is rejected with actionable advice", func(t *testing.T) {
		fds := &descriptorpb.FileDescriptorSet{File: []*descriptorpb.FileDescriptorProto{{
			Name:       proto.String("grpcecho.proto"),
			Package:    proto.String("example"),
			Dependency: []string{"google/api/annotations.proto"},
		}}}
		err := validateDescriptorClosure(fds)
		require.ErrorContains(t, err, "google/api/annotations.proto (imported by grpcecho.proto)")
		require.ErrorContains(t, err, "--include_imports")
	})
}

// mustParse builds a parsedProtoDescriptor straight from bytes, bypassing the ConfigMap.
func mustParse(t *testing.T, bin []byte) *parsedProtoDescriptor {
	t.Helper()
	tr := &Translator{TranslatorContext: &TranslatorContext{}}
	tr.SetConfigMaps([]*corev1.ConfigMap{
		configMap("d", map[string]string{"proto-descriptor": base64.StdEncoding.EncodeToString(bin)}, nil),
	})
	d, err := tr.loadProtoDescriptor(egv1a1.ProtoDescriptor{
		ValueRef: gwapiv1.LocalObjectReference{Kind: "ConfigMap", Name: "d"},
	}, "default")
	require.NoError(t, err)
	return d
}

// The merge path builds traffic features several times per route; the descriptor must only
// be decoded and validated once.
func TestLoadProtoDescriptorIsMemoized(t *testing.T) {
	tr := &Translator{TranslatorContext: &TranslatorContext{}}
	tr.SetConfigMaps([]*corev1.ConfigMap{
		configMap("descriptor", map[string]string{"proto-descriptor": grpcEchoDescriptorB64}, nil),
	})
	ref := egv1a1.ProtoDescriptor{
		ValueRef: gwapiv1.LocalObjectReference{Kind: "ConfigMap", Name: "descriptor"},
	}

	first, err := tr.loadProtoDescriptor(ref, "default")
	require.NoError(t, err)
	second, err := tr.loadProtoDescriptor(ref, "default")
	require.NoError(t, err)
	require.Same(t, first, second, "the second load must come from the cache")
	require.Len(t, tr.protoDescriptors, 1)
}

// A repeated service leaves Envoy unable to transcode, with nothing in its logs and the
// route still Accepted, so duplicates must never reach the IR.
func TestResolveTranscodedServicesDedupes(t *testing.T) {
	d := mustParse(t, grpcEchoDescriptorBin(t))

	got, err := resolveTranscodedServices(d, []string{grpcEchoService, grpcEchoService})
	require.NoError(t, err)
	require.Equal(t, []string{grpcEchoService}, got)

	// The default expansion must be free of duplicates too.
	roots, err := resolveTranscodedServices(d, nil)
	require.NoError(t, err)
	require.Len(t, roots, sets.New(roots...).Len(), "roots must not repeat a service")
}

// A broken descriptor is referenced once per rule, so re-reading and re-unmarshalling it
// on every translation is wasted work exactly when the config is already wrong.
func TestLoadProtoDescriptorCachesFailures(t *testing.T) {
	tr := &Translator{TranslatorContext: &TranslatorContext{}}
	tr.SetConfigMaps([]*corev1.ConfigMap{
		configMap("descriptor", map[string]string{"proto-descriptor": "bm90IGEgZGVzY3JpcHRvcg=="}, nil),
	})
	ref := egv1a1.ProtoDescriptor{
		ValueRef: gwapiv1.LocalObjectReference{Kind: "ConfigMap", Name: "descriptor"},
	}

	_, first := tr.loadProtoDescriptor(ref, "default")
	require.Error(t, first)
	_, second := tr.loadProtoDescriptor(ref, "default")
	require.Error(t, second)
	require.Equal(t, first.Error(), second.Error())
	require.Len(t, tr.protoDescriptors, 1, "the failure must be cached, not re-parsed")
}

func TestLoadProtoDescriptorRejectsNonConfigMap(t *testing.T) {
	tr := &Translator{TranslatorContext: &TranslatorContext{}}
	for _, ref := range []gwapiv1.LocalObjectReference{
		{Group: "apps", Kind: "ConfigMap", Name: "d"},
		{Group: "", Kind: "Secret", Name: "d"},
	} {
		_, err := tr.loadProtoDescriptor(egv1a1.ProtoDescriptor{ValueRef: ref}, "default")
		require.ErrorContains(t, err, "only ConfigMap is supported")
	}
}

// loadDescriptorSet runs fds through the same path a route takes, so these tests prove the
// failure reaches route status rather than only that a helper returns an error.
func loadDescriptorSet(t *testing.T, fds *descriptorpb.FileDescriptorSet) error {
	t.Helper()
	bin, err := proto.Marshal(fds)
	require.NoError(t, err)

	tr := &Translator{TranslatorContext: &TranslatorContext{}}
	tr.SetConfigMaps([]*corev1.ConfigMap{
		configMap("d", nil, map[string][]byte{"proto-descriptor": bin}),
	})
	_, err = tr.loadProtoDescriptor(egv1a1.ProtoDescriptor{
		ValueRef: gwapiv1.LocalObjectReference{Kind: "ConfigMap", Name: "d"},
	}, "default")
	return err
}

// Guards the descriptors that unmarshal cleanly but leave Envoy unable to build its pool,
// which NACKs the listener instead of failing this route.
func TestValidateDescriptorPool(t *testing.T) {
	t.Run("linkable descriptor passes", func(t *testing.T) {
		fds := &descriptorpb.FileDescriptorSet{}
		require.NoError(t, proto.Unmarshal(grpcEchoDescriptorBin(t), fds))
		_, err := validateDescriptorPool(fds)
		require.NoError(t, err, "must not be stricter than protoc's output")
	})

	t.Run("unresolved method type is rejected", func(t *testing.T) {
		fds := &descriptorpb.FileDescriptorSet{File: []*descriptorpb.FileDescriptorProto{{
			Name:    proto.String("echo.proto"),
			Package: proto.String("test.v1"),
			Syntax:  proto.String("proto3"),
			MessageType: []*descriptorpb.DescriptorProto{
				{Name: proto.String("PingResponse")},
			},
			Service: []*descriptorpb.ServiceDescriptorProto{{
				Name: proto.String("Echo"),
				Method: []*descriptorpb.MethodDescriptorProto{{
					Name:       proto.String("Ping"),
					InputType:  proto.String(".test.v1.PingRequest"),
					OutputType: proto.String(".test.v1.PingResponse"),
				}},
			}},
		}}}
		err := loadDescriptorSet(t, fds)
		require.ErrorContains(t, err, "failed to build a proto descriptor pool")
		require.ErrorContains(t, err, "test.v1.Echo.Ping")
		require.NotRegexp(t, `proto:[ \x{00a0}]`, err.Error(),
			"protobuf-go's prefix picks its space per binary; it must not reach a condition")
	})

	t.Run("duplicate symbol is rejected", func(t *testing.T) {
		dup := func(file string) *descriptorpb.FileDescriptorProto {
			return &descriptorpb.FileDescriptorProto{
				Name:        proto.String(file),
				Package:     proto.String("test.v1"),
				Syntax:      proto.String("proto3"),
				MessageType: []*descriptorpb.DescriptorProto{{Name: proto.String("Ping")}},
			}
		}
		err := loadDescriptorSet(t, &descriptorpb.FileDescriptorSet{
			File: []*descriptorpb.FileDescriptorProto{dup("a.proto"), dup("b.proto")},
		})
		require.ErrorContains(t, err, "failed to build a proto descriptor pool")
		require.ErrorContains(t, err, "test.v1.Ping")
	})
}

// Each case is a binding that protoc accepts but that fails Envoy's JsonTranscoderConfig
// constructor, rejecting the listener.
func TestValidateHTTPBindings(t *testing.T) {
	msg := func(name string, fields ...*descriptorpb.FieldDescriptorProto) *descriptorpb.DescriptorProto {
		return &descriptorpb.DescriptorProto{Name: proto.String(name), Field: fields}
	}
	field := func(name string, num int32, typeName string) *descriptorpb.FieldDescriptorProto {
		f := &descriptorpb.FieldDescriptorProto{
			Name:   proto.String(name),
			Number: proto.Int32(num),
			Label:  descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
			Type:   descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
		}
		if typeName != "" {
			f.Type = descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum()
			f.TypeName = proto.String(typeName)
		}
		return f
	}
	shadow := field("shadow", 3, "")
	shadow.JsonName = proto.String("deep")
	build := func(t *testing.T, rule *annotations.HttpRule) error {
		t.Helper()
		opts := &descriptorpb.MethodOptions{}
		proto.SetExtension(opts, annotations.E_Http, rule)
		bin, err := proto.Marshal(&descriptorpb.FileDescriptorSet{File: []*descriptorpb.FileDescriptorProto{
			{
				Name:        proto.String("google/api/httpbody.proto"),
				Package:     proto.String("google.api"),
				Syntax:      proto.String("proto3"),
				MessageType: []*descriptorpb.DescriptorProto{msg("HttpBody")},
			},
			{
				Name:    proto.String("echo.proto"),
				Package: proto.String("test.v1"),
				// proto2, where protoc only warns when a json_name shadows another field's
				// default JSON name.
				Syntax:     proto.String("proto2"),
				Dependency: []string{"google/api/httpbody.proto"},
				MessageType: []*descriptorpb.DescriptorProto{
					msg("Inner", field("name", 1, "")),
					msg("Req", field("user_name", 1, ""), field("inner", 2, ".test.v1.Inner"),
						shadow, field("deep", 4, ".test.v1.Inner")),
					msg("Resp", field("text", 1, ""), field("raw", 2, ".google.api.HttpBody")),
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
		}})
		require.NoError(t, err)

		tr := &Translator{TranslatorContext: &TranslatorContext{}}
		tr.SetConfigMaps([]*corev1.ConfigMap{configMap("d", nil, map[string][]byte{"proto-descriptor": bin})})
		_, err = tr.buildGRPCJSONTranscoder(&egv1a1.GRPCJSONTranscoder{
			ProtoDescriptor: egv1a1.ProtoDescriptor{
				ValueRef: gwapiv1.LocalObjectReference{Kind: "ConfigMap", Name: "d"},
			},
		}, "n", "default")
		return err
	}
	post := func(body, responseBody string) *annotations.HttpRule {
		return &annotations.HttpRule{
			Pattern:      &annotations.HttpRule_Post{Post: "/v1/ping"},
			Body:         body,
			ResponseBody: responseBody,
		}
	}

	for _, tc := range []struct {
		name    string
		rule    *annotations.HttpRule
		wantErr string
	}{
		{name: "whole message", rule: post("*", "")},
		{name: "nested field by proto name", rule: post("inner.name", "")},
		{name: "field by JSON name", rule: post("userName", "")},
		{name: "HttpBody response_body", rule: post("", "raw")},
		{
			name: "additional_bindings body is not checked, as in Envoy",
			rule: &annotations.HttpRule{
				Pattern: &annotations.HttpRule_Get{Get: "/v1/ping"},
				AdditionalBindings: []*annotations.HttpRule{{
					Pattern: &annotations.HttpRule_Post{Post: "/v1/ping"},
					Body:    "missing",
				}},
			},
		},
		{name: "missing body field", rule: post("missing", ""), wantErr: `body "missing": no field "missing" in test.v1.Req`},
		{name: "path through a scalar", rule: post("user_name.x", ""), wantErr: "user_name is not a message"},
		{name: "missing response_body field", rule: post("", "missing"), wantErr: `response_body "missing": no field "missing" in test.v1.Resp`},
		{name: "non-HttpBody response_body", rule: post("", "text"), wantErr: "must be a google.api.HttpBody field"},
		{
			name:    "JSON name wins over proto name, as in Envoy",
			rule:    post("deep.name", ""),
			wantErr: "shadow is not a message",
		},
		{
			name: "templates of every binding are parsed",
			rule: &annotations.HttpRule{
				Pattern: &annotations.HttpRule_Get{Get: "/v1/{inner.name=users/*}"},
				AdditionalBindings: []*annotations.HttpRule{{
					Pattern: &annotations.HttpRule_Custom{Custom: &annotations.CustomHttpPattern{Kind: "HEAD", Path: "/v1/ping:peek"}},
				}},
			},
		},
		{
			name:    "malformed template",
			rule:    &annotations.HttpRule{Pattern: &annotations.HttpRule_Get{Get: "v1/ping"}},
			wantErr: `invalid path template "v1/ping"`,
		},
		{
			name: "malformed template in an additional binding",
			rule: &annotations.HttpRule{
				Pattern:            &annotations.HttpRule_Get{Get: "/v1/ping"},
				AdditionalBindings: []*annotations.HttpRule{{Pattern: &annotations.HttpRule_Post{Post: "/v1/{name"}}},
			},
			wantErr: `invalid path template "/v1/{name"`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := build(t, tc.rule)
			if tc.wantErr == "" {
				require.NoError(t, err)
				return
			}
			require.ErrorContains(t, err, "method test.v1.Echo.Ping")
			require.ErrorContains(t, err, tc.wantErr)
		})
	}
}

// Envoy's HttpTemplate::Parse must accept everything valid here. Invalid cases are ones
// Envoy rejects too, except where noted.
func TestValidHTTPTemplate(t *testing.T) {
	for _, tmpl := range []string{
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
	} {
		require.True(t, validHTTPTemplate(tmpl), tmpl)
	}
	for _, tmpl := range []string{
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
		"/v1/{9name}", // Envoy accepts any identifier; field names cannot start with a digit
		"/v1/a*b",     // Envoy accepts "*" inside a literal
	} {
		require.False(t, validHTTPTemplate(tmpl), tmpl)
	}
}

// protobuf-go picks the space in its "proto:" prefix from a hash of the running binary, so
// both forms have to be trimmed or the status condition text depends on the compile.
func TestTrimProtoPrefix(t *testing.T) {
	for _, space := range []string{" ", "\u00a0"} {
		got := trimProtoPrefix(fmt.Errorf("proto:%scannot parse invalid wire-format data", space))
		require.Equal(t, "cannot parse invalid wire-format data", got)
	}
	// Anything else is passed through untouched.
	require.Equal(t, "plain error", trimProtoPrefix(errors.New("plain error")))
}
