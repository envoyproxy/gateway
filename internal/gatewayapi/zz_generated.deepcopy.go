//go:build !ignore_autogenerated
// +build !ignore_autogenerated

// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

// Code generated by controller-gen. DO NOT EDIT.

package gatewayapi

import (
	configv1alpha1 "github.com/envoyproxy/gateway/api/config/v1alpha1"
	"github.com/envoyproxy/gateway/api/v1alpha1"
	"k8s.io/api/core/v1"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"
	"sigs.k8s.io/gateway-api/apis/v1beta1"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Resources) DeepCopyInto(out *Resources) {
	*out = *in
	if in.Gateways != nil {
		in, out := &in.Gateways, &out.Gateways
		*out = make([]*v1beta1.Gateway, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(v1beta1.Gateway)
				(*in).DeepCopyInto(*out)
			}
		}
	}
	if in.HTTPRoutes != nil {
		in, out := &in.HTTPRoutes, &out.HTTPRoutes
		*out = make([]*v1beta1.HTTPRoute, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(v1beta1.HTTPRoute)
				(*in).DeepCopyInto(*out)
			}
		}
	}
	if in.GRPCRoutes != nil {
		in, out := &in.GRPCRoutes, &out.GRPCRoutes
		*out = make([]*v1alpha2.GRPCRoute, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(v1alpha2.GRPCRoute)
				(*in).DeepCopyInto(*out)
			}
		}
	}
	if in.CustomGRPCRoutes != nil {
		in, out := &in.CustomGRPCRoutes, &out.CustomGRPCRoutes
		*out = make([]*v1alpha2.CustomGRPCRoute, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(v1alpha2.CustomGRPCRoute)
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
		*out = make([]*v1alpha2.ReferenceGrant, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(v1alpha2.ReferenceGrant)
				(*in).DeepCopyInto(*out)
			}
		}
	}
	if in.Namespaces != nil {
		in, out := &in.Namespaces, &out.Namespaces
		*out = make([]*v1.Namespace, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(v1.Namespace)
				(*in).DeepCopyInto(*out)
			}
		}
	}
	if in.Services != nil {
		in, out := &in.Services, &out.Services
		*out = make([]*v1.Service, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(v1.Service)
				(*in).DeepCopyInto(*out)
			}
		}
	}
	if in.Secrets != nil {
		in, out := &in.Secrets, &out.Secrets
		*out = make([]*v1.Secret, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(v1.Secret)
				(*in).DeepCopyInto(*out)
			}
		}
	}
	if in.AuthenticationFilters != nil {
		in, out := &in.AuthenticationFilters, &out.AuthenticationFilters
		*out = make([]*v1alpha1.AuthenticationFilter, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(v1alpha1.AuthenticationFilter)
				(*in).DeepCopyInto(*out)
			}
		}
	}
	if in.RateLimitFilters != nil {
		in, out := &in.RateLimitFilters, &out.RateLimitFilters
		*out = make([]*v1alpha1.RateLimitFilter, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(v1alpha1.RateLimitFilter)
				(*in).DeepCopyInto(*out)
			}
		}
	}
	if in.EnvoyProxy != nil {
		in, out := &in.EnvoyProxy, &out.EnvoyProxy
		*out = new(configv1alpha1.EnvoyProxy)
		(*in).DeepCopyInto(*out)
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
