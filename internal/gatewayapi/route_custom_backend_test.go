// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"testing"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/gatewayapi/status"
	"github.com/envoyproxy/gateway/internal/ir"
)

// TestProcessDestinationCustomBackendWithBuiltinKind covers extension backend resources whose
// kind collides with a built-in one: they must be dispatched to the extension by group and kind,
// never to the built-in resolver for that kind.
func TestProcessDestinationCustomBackendWithBuiltinKind(t *testing.T) {
	const customGroup = "custom.example.io"

	customService := unstructured.Unstructured{}
	customService.SetAPIVersion(customGroup + "/v1")
	customService.SetKind(resource.KindService)
	customService.SetNamespace("default")
	customService.SetName("custom")

	testCases := []struct {
		name          string
		backendName   gwapiv1.ObjectName
		wantCustom    bool
		wantErrReason gwapiv1.RouteConditionReason
	}{
		{
			name:        "extension resource exists",
			backendName: "custom",
			wantCustom:  true,
		},
		{
			name:          "extension resource is missing",
			backendName:   "missing",
			wantErrReason: gwapiv1.RouteReasonBackendNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			translator := &Translator{
				ExtensionBackendGroupKinds: []schema.GroupKind{{Group: customGroup, Kind: resource.KindService}},
				TranslatorContext:          &TranslatorContext{},
			}
			route := &HTTPRouteContext{HTTPRoute: &gwapiv1.HTTPRoute{
				ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "route"},
			}}
			parentRef := &RouteParentContext{ParentReference: &gwapiv1.ParentReference{Name: "gateway"}}
			backendRef := BackendRefWithFilters{BackendRef: &gwapiv1.BackendRef{
				BackendObjectReference: gwapiv1.BackendObjectReference{
					Group: new(gwapiv1.Group(customGroup)),
					Kind:  new(gwapiv1.Kind(resource.KindService)),
					Name:  tc.backendName,
					Port:  new(gwapiv1.PortNumber(8080)),
				},
			}}
			resources := &resource.Resources{ExtensionRefFilters: []unstructured.Unstructured{customService}}

			var (
				ds              *ir.DestinationSetting
				unstructuredRef *ir.UnstructuredRef
				err             status.Error
			)
			require.NotPanics(t, func() {
				ds, unstructuredRef, err = translator.processDestination("dest", backendRef, parentRef, route,
					resources, nil, nil, nil)
			})
			require.NotNil(t, ds)
			if !tc.wantCustom {
				require.Error(t, err)
				require.Equal(t, tc.wantErrReason, err.Reason())
				require.Equal(t, gwapiv1.RouteConditionResolvedRefs, err.Type())
				require.Nil(t, unstructuredRef)
				return
			}
			require.NoError(t, err)
			require.True(t, ds.IsCustomBackend)
			require.NotNil(t, unstructuredRef)
			require.Equal(t, "custom", unstructuredRef.Object.GetName())
		})
	}
}

// TestBuiltinDestinationSettingMissingObject covers the built-in Service and ServiceImport
// resolvers being reached for an object that isn't in the translator's cache: they must report
// BackendNotFound instead of dereferencing a nil lookup result.
func TestBuiltinDestinationSettingMissingObject(t *testing.T) {
	translator := &Translator{TranslatorContext: &TranslatorContext{}}
	backendRef := gwapiv1.BackendObjectReference{Name: "missing", Port: new(gwapiv1.PortNumber(8080))}

	resolvers := map[string]func() (*ir.DestinationSetting, status.Error){
		resource.KindService: func() (*ir.DestinationSetting, status.Error) {
			return translator.processServiceDestinationSetting("dest", backendRef, "default", ir.HTTP, nil, nil)
		},
		resource.KindServiceImport: func() (*ir.DestinationSetting, status.Error) {
			return translator.processServiceImportDestinationSetting("dest", backendRef, "default", ir.HTTP, nil, nil)
		},
	}
	for kind, resolve := range resolvers {
		t.Run(kind, func(t *testing.T) {
			var (
				ds  *ir.DestinationSetting
				err status.Error
			)
			require.NotPanics(t, func() { ds, err = resolve() })
			require.Nil(t, ds)
			require.Error(t, err)
			require.Equal(t, gwapiv1.RouteReasonBackendNotFound, err.Reason())
			require.Equal(t, gwapiv1.RouteConditionResolvedRefs, err.Type())
		})
	}
}

// TestProcessDestinationBuiltinKindAsExtensionResource covers a built-in backend kind registered
// as an extension resource (ExtensionManager.Resources, e.g. for use as an extensionRef filter)
// rather than a backend resource: backendRefs to it must still use the built-in resolver.
func TestProcessDestinationBuiltinKindAsExtensionResource(t *testing.T) {
	translator := &Translator{
		ExtensionGroupKinds: []schema.GroupKind{{Group: "", Kind: resource.KindService}},
		TranslatorContext:   &TranslatorContext{},
	}
	route := &HTTPRouteContext{HTTPRoute: &gwapiv1.HTTPRoute{
		ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "route"},
	}}
	parentRef := &RouteParentContext{ParentReference: &gwapiv1.ParentReference{Name: "gateway"}}
	backendRef := BackendRefWithFilters{BackendRef: &gwapiv1.BackendRef{
		BackendObjectReference: gwapiv1.BackendObjectReference{
			Name: "missing",
			Port: new(gwapiv1.PortNumber(8080)),
		},
	}}

	ds, unstructuredRef, err := translator.processDestination("dest", backendRef, parentRef, route,
		&resource.Resources{}, nil, nil, nil)
	require.NotNil(t, ds)
	require.False(t, ds.IsCustomBackend)
	require.Nil(t, unstructuredRef)
	require.Error(t, err)
	require.Equal(t, gwapiv1.RouteReasonBackendNotFound, err.Reason())
	// Reported by the built-in backendRef validation, not the custom backend lookup.
	require.EqualError(t, err, "service default/missing not found")
}

// TestProcessDestinationCustomBackendKind covers a custom backend resource of kind Backend in its
// own group: it must not be looked up as an Envoy Gateway Backend for Backend TLS settings.
func TestProcessDestinationCustomBackendKind(t *testing.T) {
	const customGroup = "custom.example.io"

	customBackend := unstructured.Unstructured{}
	customBackend.SetAPIVersion(customGroup + "/v1")
	customBackend.SetKind(egv1a1.KindBackend)
	customBackend.SetNamespace("default")
	customBackend.SetName("custom")

	translator := &Translator{
		ExtensionBackendGroupKinds: []schema.GroupKind{{Group: customGroup, Kind: egv1a1.KindBackend}},
		TranslatorContext:          &TranslatorContext{},
	}
	route := &HTTPRouteContext{HTTPRoute: &gwapiv1.HTTPRoute{
		ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "route"},
	}}
	parentRef := &RouteParentContext{ParentReference: &gwapiv1.ParentReference{Name: "gateway"}}
	backendRef := BackendRefWithFilters{BackendRef: &gwapiv1.BackendRef{
		BackendObjectReference: gwapiv1.BackendObjectReference{
			Group: new(gwapiv1.Group(customGroup)),
			Kind:  new(gwapiv1.Kind(egv1a1.KindBackend)),
			Name:  "custom",
		},
	}}

	ds, unstructuredRef, err := translator.processDestination("dest", backendRef, parentRef, route,
		&resource.Resources{ExtensionRefFilters: []unstructured.Unstructured{customBackend}}, nil, nil, nil)
	require.NoError(t, err)
	require.True(t, ds.IsCustomBackend)
	require.NotNil(t, unstructuredRef)
	require.Equal(t, "custom", unstructuredRef.Object.GetName())
}
