// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

// This file contains code derived from Contour,
// https://github.com/projectcontour/contour
// and is provided here subject to the following:
// Copyright Project Contour Authors
// SPDX-License-Identifier: Apache-2.0

package gatewayapi

import (
	"sigs.k8s.io/gateway-api/apis/v1alpha2"
	"sigs.k8s.io/gateway-api/apis/v1"
)

// TODO: [v1alpha2-v1]
// This file can be removed once TLSRoute graduates to v1.

func GroupPtrV1Alpha2(group string) *v1alpha2.Group {
	gwGroup := v1alpha2.Group(group)
	return &gwGroup
}

func KindPtrV1Alpha2(kind string) *v1alpha2.Kind {
	gwKind := v1alpha2.Kind(kind)
	return &gwKind
}

func NamespacePtrV1Alpha2(namespace string) *v1alpha2.Namespace {
	gwNamespace := v1alpha2.Namespace(namespace)
	return &gwNamespace
}

func SectionNamePtrV1Alpha2(sectionName string) *v1alpha2.SectionName {
	gwSectionName := v1alpha2.SectionName(sectionName)
	return &gwSectionName
}

func PortNumPtrV1Alpha2(port int) *v1alpha2.PortNumber {
	pn := v1alpha2.PortNumber(port)
	return &pn
}

func UpgradeParentReferences(old []v1alpha2.ParentReference) []v1.ParentReference {
	newParentReferences := make([]v1.ParentReference, len(old))
	for i, o := range old {
		newParentReferences[i] = UpgradeParentReference(o)
	}
	return newParentReferences
}

// UpgradeParentReference converts v1alpha2.ParentReference to v1.ParentReference
func UpgradeParentReference(old v1alpha2.ParentReference) v1.ParentReference {
	upgraded := v1.ParentReference{}

	if old.Group != nil {
		upgraded.Group = GroupPtr(string(*old.Group))
	}

	if old.Kind != nil {
		upgraded.Kind = KindPtr(string(*old.Kind))
	}

	if old.Namespace != nil {
		upgraded.Namespace = NamespacePtr(string(*old.Namespace))
	}

	upgraded.Name = old.Name

	if old.SectionName != nil {
		upgraded.SectionName = SectionNamePtr(string(*old.SectionName))
	}

	if old.Port != nil {
		upgraded.Port = PortNumPtr(int32(*old.Port))
	}

	return upgraded
}

func DowngradeParentReference(old v1.ParentReference) v1alpha2.ParentReference {
	downgraded := v1alpha2.ParentReference{}

	if old.Group != nil {
		downgraded.Group = GroupPtrV1Alpha2(string(*old.Group))
	}

	if old.Kind != nil {
		downgraded.Kind = KindPtrV1Alpha2(string(*old.Kind))
	}

	if old.Namespace != nil {
		downgraded.Namespace = NamespacePtrV1Alpha2(string(*old.Namespace))
	}

	downgraded.Name = old.Name

	if old.SectionName != nil {
		downgraded.SectionName = SectionNamePtrV1Alpha2(string(*old.SectionName))
	}

	if old.Port != nil {
		downgraded.Port = PortNumPtrV1Alpha2(int(*old.Port))
	}

	return downgraded
}

func UpgradeRouteParentStatuses(routeParentStatuses []v1alpha2.RouteParentStatus) []v1.RouteParentStatus {
	var res []v1.RouteParentStatus

	for _, rps := range routeParentStatuses {
		res = append(res, v1.RouteParentStatus{
			ParentRef:      UpgradeParentReference(rps.ParentRef),
			ControllerName: rps.ControllerName,
			Conditions:     rps.Conditions,
		})
	}

	return res
}

func DowngradeRouteParentStatuses(routeParentStatuses []v1.RouteParentStatus) []v1alpha2.RouteParentStatus {
	var res []v1alpha2.RouteParentStatus

	for _, rps := range routeParentStatuses {
		res = append(res, v1alpha2.RouteParentStatus{
			ParentRef:      DowngradeParentReference(rps.ParentRef),
			ControllerName: rps.ControllerName,
			Conditions:     rps.Conditions,
		})
	}

	return res
}

// UpgradeBackendRef converts v1alpha2.BackendRef to v1.BackendRef
func UpgradeBackendRef(old v1alpha2.BackendRef) v1.BackendRef {
	upgraded := v1.BackendRef{}

	if old.Group != nil {
		upgraded.Group = GroupPtr(string(*old.Group))
	}

	if old.Kind != nil {
		upgraded.Kind = KindPtr(string(*old.Kind))
	}

	if old.Namespace != nil {
		upgraded.Namespace = NamespacePtr(string(*old.Namespace))
	}

	upgraded.Name = old.Name

	if old.Port != nil {
		upgraded.Port = PortNumPtr(int32(*old.Port))
	}

	return upgraded
}

func DowngradeBackendRef(old v1.BackendRef) v1alpha2.BackendRef {
	downgraded := v1alpha2.BackendRef{}

	if old.Group != nil {
		downgraded.Group = GroupPtrV1Alpha2(string(*old.Group))
	}

	if old.Kind != nil {
		downgraded.Kind = KindPtrV1Alpha2(string(*old.Kind))
	}

	if old.Namespace != nil {
		downgraded.Namespace = NamespacePtrV1Alpha2(string(*old.Namespace))
	}

	downgraded.Name = old.Name

	if old.Port != nil {
		downgraded.Port = PortNumPtrV1Alpha2(int(*old.Port))
	}

	return downgraded
}

func NamespaceDerefOrAlpha(namespace *v1alpha2.Namespace, defaultNamespace string) string {
	if namespace != nil && *namespace != "" {
		return string(*namespace)
	}
	return defaultNamespace
}
