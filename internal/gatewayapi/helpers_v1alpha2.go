package gatewayapi

import (
	"sigs.k8s.io/gateway-api/apis/v1alpha2"
	"sigs.k8s.io/gateway-api/apis/v1beta1"
)

// TODO: [v1alpha2-v1beta1]
// This file can be removed once TLSRoute graduates to v1beta1.

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

// UpgradeParentReference converts v1alpha2.ParentReference to v1beta1.ParentReference
func UpgradeParentReference(old v1alpha2.ParentReference) v1beta1.ParentReference {
	upgraded := v1beta1.ParentReference{}

	if old.Group != nil {
		upgraded.Group = GroupPtr(string(*old.Group))
	}

	if old.Kind != nil {
		upgraded.Kind = KindPtr(string(*old.Kind))
	}

	if old.Namespace != nil {
		upgraded.Namespace = NamespacePtr(string(*old.Namespace))
	}

	upgraded.Name = v1beta1.ObjectName(old.Name)

	if old.SectionName != nil {
		upgraded.SectionName = SectionNamePtr(string(*old.SectionName))
	}

	if old.Port != nil {
		upgraded.Port = PortNumPtr(int32(*old.Port))
	}

	return upgraded
}

func DowngradeParentReference(old v1beta1.ParentReference) v1alpha2.ParentReference {
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

	downgraded.Name = v1alpha2.ObjectName(old.Name)

	if old.SectionName != nil {
		downgraded.SectionName = SectionNamePtrV1Alpha2(string(*old.SectionName))
	}

	if old.Port != nil {
		downgraded.Port = PortNumPtrV1Alpha2(int(*old.Port))
	}

	return downgraded
}

func UpgradeRouteParentStatuses(routeParentStatuses []v1alpha2.RouteParentStatus) []v1beta1.RouteParentStatus {
	var res []v1beta1.RouteParentStatus

	for _, rps := range routeParentStatuses {
		res = append(res, v1beta1.RouteParentStatus{
			ParentRef:      UpgradeParentReference(rps.ParentRef),
			ControllerName: v1beta1.GatewayController(rps.ControllerName),
			Conditions:     rps.Conditions,
		})
	}

	return res
}

func DowngradeRouteParentStatuses(routeParentStatuses []v1beta1.RouteParentStatus) []v1alpha2.RouteParentStatus {
	var res []v1alpha2.RouteParentStatus

	for _, rps := range routeParentStatuses {
		res = append(res, v1alpha2.RouteParentStatus{
			ParentRef:      DowngradeParentReference(rps.ParentRef),
			ControllerName: v1alpha2.GatewayController(rps.ControllerName),
			Conditions:     rps.Conditions,
		})
	}

	return res
}

func NamespaceDerefOrAlpha(namespace *v1alpha2.Namespace, defaultNamespace string) string {
	if namespace != nil && *namespace != "" {
		return string(*namespace)
	}
	return defaultNamespace
}
