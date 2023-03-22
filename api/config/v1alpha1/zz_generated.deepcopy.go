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
	if in.RateLimit != nil {
		in, out := &in.RateLimit, &out.RateLimit
		*out = new(RateLimit)
		(*in).DeepCopyInto(*out)
	}
	if in.Extension != nil {
		in, out := &in.Extension, &out.Extension
		*out = new(Extension)
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
	in.Spec.DeepCopyInto(&out.Spec)
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
	if in.Provider != nil {
		in, out := &in.Provider, &out.Provider
		*out = new(ResourceProvider)
		(*in).DeepCopyInto(*out)
	}
	in.Logging.DeepCopyInto(&out.Logging)
	if in.Bootstrap != nil {
		in, out := &in.Bootstrap, &out.Bootstrap
		*out = new(string)
		**out = **in
	}
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
func (in *Extension) DeepCopyInto(out *Extension) {
	*out = *in
	if in.Resources != nil {
		in, out := &in.Resources, &out.Resources
		*out = make([]GroupVersionKind, len(*in))
		copy(*out, *in)
	}
	if in.Hooks != nil {
		in, out := &in.Hooks, &out.Hooks
		*out = new(ExtensionHooks)
		(*in).DeepCopyInto(*out)
	}
	if in.Service != nil {
		in, out := &in.Service, &out.Service
		*out = new(ExtensionService)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Extension.
func (in *Extension) DeepCopy() *Extension {
	if in == nil {
		return nil
	}
	out := new(Extension)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ExtensionHooks) DeepCopyInto(out *ExtensionHooks) {
	*out = *in
	if in.XDSTranslation != nil {
		in, out := &in.XDSTranslation, &out.XDSTranslation
		*out = new(XDSTranslationHooks)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ExtensionHooks.
func (in *ExtensionHooks) DeepCopy() *ExtensionHooks {
	if in == nil {
		return nil
	}
	out := new(ExtensionHooks)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ExtensionService) DeepCopyInto(out *ExtensionService) {
	*out = *in
	if in.TLS != nil {
		in, out := &in.TLS, &out.TLS
		*out = new(ExtensionTLS)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ExtensionService.
func (in *ExtensionService) DeepCopy() *ExtensionService {
	if in == nil {
		return nil
	}
	out := new(ExtensionService)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ExtensionTLS) DeepCopyInto(out *ExtensionTLS) {
	*out = *in
	in.CertificateRef.DeepCopyInto(&out.CertificateRef)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ExtensionTLS.
func (in *ExtensionTLS) DeepCopy() *ExtensionTLS {
	if in == nil {
		return nil
	}
	out := new(ExtensionTLS)
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
func (in *GroupVersionKind) DeepCopyInto(out *GroupVersionKind) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GroupVersionKind.
func (in *GroupVersionKind) DeepCopy() *GroupVersionKind {
	if in == nil {
		return nil
	}
	out := new(GroupVersionKind)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KubernetesDeploymentSpec) DeepCopyInto(out *KubernetesDeploymentSpec) {
	*out = *in
	if in.Replicas != nil {
		in, out := &in.Replicas, &out.Replicas
		*out = new(int32)
		**out = **in
	}
	if in.PodAnnotations != nil {
		in, out := &in.PodAnnotations, &out.PodAnnotations
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KubernetesDeploymentSpec.
func (in *KubernetesDeploymentSpec) DeepCopy() *KubernetesDeploymentSpec {
	if in == nil {
		return nil
	}
	out := new(KubernetesDeploymentSpec)
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
func (in *KubernetesResourceProvider) DeepCopyInto(out *KubernetesResourceProvider) {
	*out = *in
	if in.EnvoyDeployment != nil {
		in, out := &in.EnvoyDeployment, &out.EnvoyDeployment
		*out = new(KubernetesDeploymentSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.EnvoyService != nil {
		in, out := &in.EnvoyService, &out.EnvoyService
		*out = new(KubernetesServiceSpec)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KubernetesResourceProvider.
func (in *KubernetesResourceProvider) DeepCopy() *KubernetesResourceProvider {
	if in == nil {
		return nil
	}
	out := new(KubernetesResourceProvider)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KubernetesServiceSpec) DeepCopyInto(out *KubernetesServiceSpec) {
	*out = *in
	if in.Annotations != nil {
		in, out := &in.Annotations, &out.Annotations
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KubernetesServiceSpec.
func (in *KubernetesServiceSpec) DeepCopy() *KubernetesServiceSpec {
	if in == nil {
		return nil
	}
	out := new(KubernetesServiceSpec)
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

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ProxyLogging) DeepCopyInto(out *ProxyLogging) {
	*out = *in
	if in.Level != nil {
		in, out := &in.Level, &out.Level
		*out = make(map[LogComponent]LogLevel, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ProxyLogging.
func (in *ProxyLogging) DeepCopy() *ProxyLogging {
	if in == nil {
		return nil
	}
	out := new(ProxyLogging)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RateLimit) DeepCopyInto(out *RateLimit) {
	*out = *in
	in.Backend.DeepCopyInto(&out.Backend)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RateLimit.
func (in *RateLimit) DeepCopy() *RateLimit {
	if in == nil {
		return nil
	}
	out := new(RateLimit)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RateLimitDatabaseBackend) DeepCopyInto(out *RateLimitDatabaseBackend) {
	*out = *in
	if in.Redis != nil {
		in, out := &in.Redis, &out.Redis
		*out = new(RateLimitRedisSettings)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RateLimitDatabaseBackend.
func (in *RateLimitDatabaseBackend) DeepCopy() *RateLimitDatabaseBackend {
	if in == nil {
		return nil
	}
	out := new(RateLimitDatabaseBackend)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RateLimitRedisSettings) DeepCopyInto(out *RateLimitRedisSettings) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RateLimitRedisSettings.
func (in *RateLimitRedisSettings) DeepCopy() *RateLimitRedisSettings {
	if in == nil {
		return nil
	}
	out := new(RateLimitRedisSettings)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ResourceProvider) DeepCopyInto(out *ResourceProvider) {
	*out = *in
	if in.Kubernetes != nil {
		in, out := &in.Kubernetes, &out.Kubernetes
		*out = new(KubernetesResourceProvider)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResourceProvider.
func (in *ResourceProvider) DeepCopy() *ResourceProvider {
	if in == nil {
		return nil
	}
	out := new(ResourceProvider)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *XDSTranslationHooks) DeepCopyInto(out *XDSTranslationHooks) {
	*out = *in
	if in.Pre != nil {
		in, out := &in.Pre, &out.Pre
		*out = make([]XDSTranslationHook, len(*in))
		copy(*out, *in)
	}
	if in.Post != nil {
		in, out := &in.Post, &out.Post
		*out = make([]XDSTranslationHook, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new XDSTranslationHooks.
func (in *XDSTranslationHooks) DeepCopy() *XDSTranslationHooks {
	if in == nil {
		return nil
	}
	out := new(XDSTranslationHooks)
	in.DeepCopyInto(out)
	return out
}
