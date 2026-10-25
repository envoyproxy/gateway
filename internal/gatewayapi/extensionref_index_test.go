// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
)

func extRes(apiVersion, kind, ns, name, marker string) unstructured.Unstructured {
	u := unstructured.Unstructured{}
	u.SetAPIVersion(apiVersion)
	u.SetKind(kind)
	u.SetNamespace(ns)
	u.SetName(name)
	u.SetLabels(map[string]string{"marker": marker})
	return u
}

func TestGetExtensionRef(t *testing.T) {
	tests := []struct {
		name          string
		resources     []unstructured.Unstructured
		wantMarker    string // marker of the expected resource, empty for none
		wantMalformed bool
	}{
		{
			name:       "match",
			resources:  []unstructured.Unstructured{extRes("foo.example.io/v1", "Foo", "ns", "a", "match")},
			wantMarker: "match",
		},
		{
			name: "first of several matches wins",
			resources: []unstructured.Unstructured{
				extRes("foo.example.io/v1", "Foo", "ns", "a", "first"),
				extRes("foo.example.io/v2", "Foo", "ns", "a", "second"),
			},
			wantMarker: "first",
		},
		{
			name: "other group, namespace or name does not match",
			resources: []unstructured.Unstructured{
				extRes("bar.example.io/v1", "Foo", "ns", "a", "group"),
				extRes("foo.example.io/v1", "Foo", "other", "a", "namespace"),
				extRes("foo.example.io/v1", "Foo", "ns", "b", "name"),
				extRes("foo.example.io/v1", "Bar", "ns", "a", "kind"),
			},
		},
		{
			name: "match after another group",
			resources: []unstructured.Unstructured{
				extRes("bar.example.io/v1", "Foo", "ns", "a", "other-group"),
				extRes("foo.example.io/v1", "Foo", "ns", "a", "match"),
			},
			wantMarker: "match",
		},
		{
			name:          "malformed apiVersion only",
			resources:     []unstructured.Unstructured{extRes("v1", "Foo", "ns", "a", "malformed")},
			wantMalformed: true,
		},
		{
			name: "malformed apiVersion before the match",
			resources: []unstructured.Unstructured{
				extRes("v1", "Foo", "ns", "a", "malformed"),
				extRes("foo.example.io/v1", "Foo", "ns", "a", "match"),
			},
			wantMarker:    "match",
			wantMalformed: true,
		},
		{
			name: "malformed apiVersion after the match",
			resources: []unstructured.Unstructured{
				extRes("foo.example.io/v1", "Foo", "ns", "a", "match"),
				extRes("v1", "Foo", "ns", "a", "malformed"),
			},
			wantMarker: "match",
		},
		{
			name: "malformed apiVersion for another name",
			resources: []unstructured.Unstructured{
				extRes("v1", "Foo", "ns", "b", "malformed"),
				extRes("foo.example.io/v1", "Foo", "ns", "a", "match"),
			},
			wantMarker: "match",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resources := &resource.Resources{ExtensionRefFilters: tc.resources}
			tr := &Translator{TranslatorContext: &TranslatorContext{
				ExtensionRefIndex: BuildExtensionRefIndex(resources.ExtensionRefFilters),
			}}

			got, malformed := tr.getExtensionRef(resources, "foo.example.io", "Foo", "ns", "a")
			require.Equal(t, tc.wantMalformed, malformed)
			if tc.wantMarker == "" {
				require.Nil(t, got)
				return
			}
			require.NotNil(t, got)
			require.Equal(t, tc.wantMarker, got.GetLabels()["marker"])
		})
	}
}

func TestGetExtensionRefReturnsCopy(t *testing.T) {
	resources := &resource.Resources{ExtensionRefFilters: []unstructured.Unstructured{
		extRes("foo.example.io/v1", "Foo", "ns", "a", "match"),
	}}
	tr := &Translator{TranslatorContext: &TranslatorContext{
		ExtensionRefIndex: BuildExtensionRefIndex(resources.ExtensionRefFilters),
	}}

	first, _ := tr.getExtensionRef(resources, "foo.example.io", "Foo", "ns", "a")
	second, _ := tr.getExtensionRef(resources, "foo.example.io", "Foo", "ns", "a")
	require.NotSame(t, first, second)
	require.NotSame(t, &resources.ExtensionRefFilters[0], first)
}
