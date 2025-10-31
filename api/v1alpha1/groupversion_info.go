// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

const GroupName = "gateway.envoyproxy.io"

var (
	// GroupVersion is group version used to register these objects
	GroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1alpha1"}

	// SchemeGroupVersion is an alias for GroupVersion for code-generator compatibility
	SchemeGroupVersion = GroupVersion

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)

	// localSchemeBuilder is used for controller-runtime compatibility
	localSchemeBuilder = &scheme.Builder{GroupVersion: GroupVersion}

	// AddToScheme adds the types in this group-version to the given scheme.
	AddToScheme = SchemeBuilder.AddToScheme
)

// Resource takes an unqualified resource and returns a Group qualified GroupResource
func Resource(resource string) schema.GroupResource {
	return GroupVersion.WithResource(resource).GroupResource()
}

// addKnownTypes adds the set of types defined in this package to the supplied scheme.
func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(GroupVersion,
		&Backend{},
		&BackendList{},
		&BackendTrafficPolicy{},
		&BackendTrafficPolicyList{},
		&ClientTrafficPolicy{},
		&ClientTrafficPolicyList{},
		&EnvoyExtensionPolicy{},
		&EnvoyExtensionPolicyList{},
		&EnvoyGateway{},
		&EnvoyPatchPolicy{},
		&EnvoyPatchPolicyList{},
		&EnvoyProxy{},
		&EnvoyProxyList{},
		&HTTPRouteFilter{},
		&HTTPRouteFilterList{},
		&SecurityPolicy{},
		&SecurityPolicyList{},
	)
	metav1.AddToGroupVersion(scheme, GroupVersion)
	return nil
}
