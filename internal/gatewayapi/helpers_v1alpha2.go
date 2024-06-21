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
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
)

// TODO: [gwapiv1a2-gwapiv1]
// This file can be removed once all routes graduates to gwapiv1.

func GroupPtrV1Alpha2(group string) *gwapiv1a2.Group {
	gwGroup := gwapiv1a2.Group(group)
	return &gwGroup
}

func KindPtrV1Alpha2(kind string) *gwapiv1a2.Kind {
	gwKind := gwapiv1a2.Kind(kind)
	return &gwKind
}

func NamespacePtrV1Alpha2(namespace string) *gwapiv1a2.Namespace {
	gwNamespace := gwapiv1a2.Namespace(namespace)
	return &gwNamespace
}

func SectionNamePtrV1Alpha2(sectionName string) *gwapiv1a2.SectionName {
	gwSectionName := gwapiv1a2.SectionName(sectionName)
	return &gwSectionName
}

func PortNumPtrV1Alpha2(port int) *gwapiv1a2.PortNumber {
	pn := gwapiv1a2.PortNumber(port)
	return &pn
}

func DowngradeParentReference(old gwapiv1.ParentReference) gwapiv1a2.ParentReference {
	downgraded := gwapiv1a2.ParentReference{}

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

// UpgradeBackendRef converts gwapiv1a2.BackendRef to gwapiv1.BackendRef
func UpgradeBackendRef(old gwapiv1a2.BackendRef) gwapiv1.BackendRef {
	upgraded := gwapiv1.BackendRef{}

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
