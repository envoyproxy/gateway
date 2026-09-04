// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
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
		configMap("sole-entry", map[string]string{"anything.pb": grpcEchoDescriptorB64}, nil),
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
		{name: "configmap sole entry", desc: valueRef("sole-entry"), want: bin},
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
