//go:build !ignore_autogenerated
// +build !ignore_autogenerated

// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.



// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *EnvoyGateway) DeepCopyInto(out *EnvoyGateway) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.EnvoyGatewaySpec.DeepCopyInto(&out.EnvoyGatewaySpec)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new EnvoyGateway.
func (in *EnvoyGateway) DeepCopy() *EnvoyGateway {
	if in == nil {
		return nil
	}
	out := new(EnvoyGateway)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *EnvoyGateway) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *EnvoyGatewaySpec) DeepCopyInto(out *EnvoyGatewaySpec) {
	*out = *in
	if in.Gateway != nil {
		in, out := &in.Gateway, &out.Gateway
		*out = new(Gateway)
		**out = **in
	}
	if in.Provider != nil {
		in, out := &in.Provider, &out.Provider
		*out = new(Provider)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new EnvoyGatewaySpec.
func (in *EnvoyGatewaySpec) DeepCopy() *EnvoyGatewaySpec {
	if in == nil {
		return nil
	}
	out := new(EnvoyGatewaySpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *EnvoyProxy) DeepCopyInto(out *EnvoyProxy) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new EnvoyProxy.
func (in *EnvoyProxy) DeepCopy() *EnvoyProxy {
	if in == nil {
		return nil
	}
	out := new(EnvoyProxy)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *EnvoyProxy) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *EnvoyProxyList) DeepCopyInto(out *EnvoyProxyList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]EnvoyProxy, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new EnvoyProxyList.
func (in *EnvoyProxyList) DeepCopy() *EnvoyProxyList {
	if in == nil {
		return nil
	}
	out := new(EnvoyProxyList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *EnvoyProxyList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *EnvoyProxySpec) DeepCopyInto(out *EnvoyProxySpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new EnvoyProxySpec.
func (in *EnvoyProxySpec) DeepCopy() *EnvoyProxySpec {
	if in == nil {
		return nil
	}
	out := new(EnvoyProxySpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *EnvoyProxyStatus) DeepCopyInto(out *EnvoyProxyStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new EnvoyProxyStatus.
func (in *EnvoyProxyStatus) DeepCopy() *EnvoyProxyStatus {
	if in == nil {
		return nil
	}
	out := new(EnvoyProxyStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FileProvider) DeepCopyInto(out *FileProvider) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FileProvider.
func (in *FileProvider) DeepCopy() *FileProvider {
	if in == nil {
		return nil
	}
	out := new(FileProvider)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Gateway) DeepCopyInto(out *Gateway) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Gateway.
func (in *Gateway) DeepCopy() *Gateway {
	if in == nil {
		return nil
	}
	out := new(Gateway)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KubernetesProvider) DeepCopyInto(out *KubernetesProvider) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KubernetesProvider.
func (in *KubernetesProvider) DeepCopy() *KubernetesProvider {
	if in == nil {
		return nil
	}
	out := new(KubernetesProvider)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Provider) DeepCopyInto(out *Provider) {
	*out = *in
	if in.Kubernetes != nil {
		in, out := &in.Kubernetes, &out.Kubernetes
		*out = new(KubernetesProvider)
		**out = **in
	}
	if in.File != nil {
		in, out := &in.File, &out.File
		*out = new(FileProvider)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Provider.
func (in *Provider) DeepCopy() *Provider {
	if in == nil {
		return nil
	}
	out := new(Provider)
	in.DeepCopyInto(out)
	return out
}
