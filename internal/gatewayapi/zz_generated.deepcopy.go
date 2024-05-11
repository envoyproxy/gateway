//go:build !ignore_autogenerated

// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

// Code generated by controller-gen. DO NOT EDIT.

package gatewayapi

import (
	apiv1alpha1 "github.com/envoyproxy/gateway/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"
	"sigs.k8s.io/gateway-api/apis/v1alpha3"
	"sigs.k8s.io/gateway-api/apis/v1beta1"
	"sigs.k8s.io/mcs-api/pkg/apis/v1alpha1"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Resources) DeepCopyInto(out *Resources) {
	*out = *in
	if in.GatewayClass != nil {
		in, out := &in.GatewayClass, &out.GatewayClass
		*out = new(v1.GatewayClass)
		(*in).DeepCopyInto(*out)
	}
	if in.Gateways != nil {
		in, out := &in.Gateways, &out.Gateways
		*out = make([]*v1.Gateway, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(v1.Gateway)
				(*in).DeepCopyInto(*out)
			}
		}
	}
	if in.HTTPRoutes != nil {
		in, out := &in.HTTPRoutes, &out.HTTPRoutes
		*out = make([]*v1.HTTPRoute, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(v1.HTTPRoute)
				(*in).DeepCopyInto(*out)
			}
		}
	}
	if in.GRPCRoutes != nil {
		in, out := &in.GRPCRoutes, &out.GRPCRoutes
		*out = make([]*v1.GRPCRoute, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(v1.GRPCRoute)
				(*in).DeepCopyInto(*out)
			}
		}
	}
	if in.TLSRoutes != nil {
		in, out := &in.TLSRoutes, &out.TLSRoutes
		*out = make([]*v1alpha2.TLSRoute, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(v1alpha2.TLSRoute)
				(*in).DeepCopyInto(*out)
			}
		}
	}
	if in.TCPRoutes != nil {
		in, out := &in.TCPRoutes, &out.TCPRoutes
		*out = make([]*v1alpha2.TCPRoute, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(v1alpha2.TCPRoute)
				(*in).DeepCopyInto(*out)
			}
		}
	}
	if in.UDPRoutes != nil {
		in, out := &in.UDPRoutes, &out.UDPRoutes
		*out = make([]*v1alpha2.UDPRoute, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(v1alpha2.UDPRoute)
				(*in).DeepCopyInto(*out)
			}
		}
	}
	if in.ReferenceGrants != nil {
		in, out := &in.ReferenceGrants, &out.ReferenceGrants
		*out = make([]*v1beta1.ReferenceGrant, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(v1beta1.ReferenceGrant)
				(*in).DeepCopyInto(*out)
			}
		}
	}
	if in.Namespaces != nil {
		in, out := &in.Namespaces, &out.Namespaces
		*out = make([]*corev1.Namespace, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(corev1.Namespace)
				(*in).DeepCopyInto(*out)
			}
		}
	}
	if in.Services != nil {
		in, out := &in.Services, &out.Services
		*out = make([]*corev1.Service, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(corev1.Service)
				(*in).DeepCopyInto(*out)
			}
		}
	}
	if in.ServiceImports != nil {
		in, out := &in.ServiceImports, &out.ServiceImports
		*out = make([]*v1alpha1.ServiceImport, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(v1alpha1.ServiceImport)
				(*in).DeepCopyInto(*out)
			}
		}
	}
	if in.EndpointSlices != nil {
		in, out := &in.EndpointSlices, &out.EndpointSlices
		*out = make([]*discoveryv1.EndpointSlice, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(discoveryv1.EndpointSlice)
				(*in).DeepCopyInto(*out)
			}
		}
	}
	if in.Secrets != nil {
		in, out := &in.Secrets, &out.Secrets
		*out = make([]*corev1.Secret, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(corev1.Secret)
				(*in).DeepCopyInto(*out)
			}
		}
	}
	if in.ConfigMaps != nil {
		in, out := &in.ConfigMaps, &out.ConfigMaps
		*out = make([]*corev1.ConfigMap, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(corev1.ConfigMap)
				(*in).DeepCopyInto(*out)
			}
		}
	}
	if in.EnvoyProxy != nil {
		in, out := &in.EnvoyProxy, &out.EnvoyProxy
		*out = new(apiv1alpha1.EnvoyProxy)
		(*in).DeepCopyInto(*out)
	}
	if in.ExtensionRefFilters != nil {
		in, out := &in.ExtensionRefFilters, &out.ExtensionRefFilters
		*out = make([]unstructured.Unstructured, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.EnvoyPatchPolicies != nil {
		in, out := &in.EnvoyPatchPolicies, &out.EnvoyPatchPolicies
		*out = make([]*apiv1alpha1.EnvoyPatchPolicy, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(apiv1alpha1.EnvoyPatchPolicy)
				(*in).DeepCopyInto(*out)
			}
		}
	}
	if in.ClientTrafficPolicies != nil {
		in, out := &in.ClientTrafficPolicies, &out.ClientTrafficPolicies
		*out = make([]*apiv1alpha1.ClientTrafficPolicy, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(apiv1alpha1.ClientTrafficPolicy)
				(*in).DeepCopyInto(*out)
			}
		}
	}
	if in.BackendTrafficPolicies != nil {
		in, out := &in.BackendTrafficPolicies, &out.BackendTrafficPolicies
		*out = make([]*apiv1alpha1.BackendTrafficPolicy, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(apiv1alpha1.BackendTrafficPolicy)
				(*in).DeepCopyInto(*out)
			}
		}
	}
	if in.SecurityPolicies != nil {
		in, out := &in.SecurityPolicies, &out.SecurityPolicies
		*out = make([]*apiv1alpha1.SecurityPolicy, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(apiv1alpha1.SecurityPolicy)
				(*in).DeepCopyInto(*out)
			}
		}
	}
	if in.BackendTLSPolicies != nil {
		in, out := &in.BackendTLSPolicies, &out.BackendTLSPolicies
		*out = make([]*v1alpha3.BackendTLSPolicy, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(v1alpha3.BackendTLSPolicy)
				(*in).DeepCopyInto(*out)
			}
		}
	}
	if in.EnvoyExtensionPolicies != nil {
		in, out := &in.EnvoyExtensionPolicies, &out.EnvoyExtensionPolicies
		*out = make([]*apiv1alpha1.EnvoyExtensionPolicy, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(apiv1alpha1.EnvoyExtensionPolicy)
				(*in).DeepCopyInto(*out)
			}
		}
	}
	if in.ExtServerPolicies != nil {
		in, out := &in.ExtServerPolicies, &out.ExtServerPolicies
		*out = make([]unstructured.Unstructured, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Resources.
func (in *Resources) DeepCopy() *Resources {
	if in == nil {
		return nil
	}
	out := new(Resources)
	in.DeepCopyInto(out)
	return out
}
