//go:build !ignore_autogenerated
// +build !ignore_autogenerated

// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AuthenticationFilter) DeepCopyInto(out *AuthenticationFilter) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AuthenticationFilter.
func (in *AuthenticationFilter) DeepCopy() *AuthenticationFilter {
	if in == nil {
		return nil
	}
	out := new(AuthenticationFilter)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *AuthenticationFilter) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AuthenticationFilterList) DeepCopyInto(out *AuthenticationFilterList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]AuthenticationFilter, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AuthenticationFilterList.
func (in *AuthenticationFilterList) DeepCopy() *AuthenticationFilterList {
	if in == nil {
		return nil
	}
	out := new(AuthenticationFilterList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *AuthenticationFilterList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AuthenticationFilterSpec) DeepCopyInto(out *AuthenticationFilterSpec) {
	*out = *in
	if in.JwtProviders != nil {
		in, out := &in.JwtProviders, &out.JwtProviders
		*out = make([]JwtAuthenticationFilterProvider, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AuthenticationFilterSpec.
func (in *AuthenticationFilterSpec) DeepCopy() *AuthenticationFilterSpec {
	if in == nil {
		return nil
	}
	out := new(AuthenticationFilterSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CorsFilter) DeepCopyInto(out *CorsFilter) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CorsFilter.
func (in *CorsFilter) DeepCopy() *CorsFilter {
	if in == nil {
		return nil
	}
	out := new(CorsFilter)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CorsFilter) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CorsFilterList) DeepCopyInto(out *CorsFilterList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]CorsFilter, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CorsFilterList.
func (in *CorsFilterList) DeepCopy() *CorsFilterList {
	if in == nil {
		return nil
	}
	out := new(CorsFilterList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CorsFilterList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CorsFilterSpec) DeepCopyInto(out *CorsFilterSpec) {
	*out = *in
	in.CorsPolicy.DeepCopyInto(&out.CorsPolicy)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CorsFilterSpec.
func (in *CorsFilterSpec) DeepCopy() *CorsFilterSpec {
	if in == nil {
		return nil
	}
	out := new(CorsFilterSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CorsPolicy) DeepCopyInto(out *CorsPolicy) {
	*out = *in
	if in.AllowOrigins != nil {
		in, out := &in.AllowOrigins, &out.AllowOrigins
		*out = make([]*StringMatch, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(StringMatch)
				(*in).DeepCopyInto(*out)
			}
		}
	}
	if in.AllowMethods != nil {
		in, out := &in.AllowMethods, &out.AllowMethods
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.AllowHeaders != nil {
		in, out := &in.AllowHeaders, &out.AllowHeaders
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.ExposeHeaders != nil {
		in, out := &in.ExposeHeaders, &out.ExposeHeaders
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CorsPolicy.
func (in *CorsPolicy) DeepCopy() *CorsPolicy {
	if in == nil {
		return nil
	}
	out := new(CorsPolicy)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *EnvoyJSONPatchConfig) DeepCopyInto(out *EnvoyJSONPatchConfig) {
	*out = *in
	out.Operation = in.Operation
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new EnvoyJSONPatchConfig.
func (in *EnvoyJSONPatchConfig) DeepCopy() *EnvoyJSONPatchConfig {
	if in == nil {
		return nil
	}
	out := new(EnvoyJSONPatchConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *EnvoyPatchPolicy) DeepCopyInto(out *EnvoyPatchPolicy) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new EnvoyPatchPolicy.
func (in *EnvoyPatchPolicy) DeepCopy() *EnvoyPatchPolicy {
	if in == nil {
		return nil
	}
	out := new(EnvoyPatchPolicy)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *EnvoyPatchPolicy) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *EnvoyPatchPolicyList) DeepCopyInto(out *EnvoyPatchPolicyList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]EnvoyPatchPolicy, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new EnvoyPatchPolicyList.
func (in *EnvoyPatchPolicyList) DeepCopy() *EnvoyPatchPolicyList {
	if in == nil {
		return nil
	}
	out := new(EnvoyPatchPolicyList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *EnvoyPatchPolicyList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *EnvoyPatchPolicySpec) DeepCopyInto(out *EnvoyPatchPolicySpec) {
	*out = *in
	if in.JSONPatches != nil {
		in, out := &in.JSONPatches, &out.JSONPatches
		*out = make([]EnvoyJSONPatchConfig, len(*in))
		copy(*out, *in)
	}
	in.TargetRef.DeepCopyInto(&out.TargetRef)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new EnvoyPatchPolicySpec.
func (in *EnvoyPatchPolicySpec) DeepCopy() *EnvoyPatchPolicySpec {
	if in == nil {
		return nil
	}
	out := new(EnvoyPatchPolicySpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *EnvoyPatchPolicyStatus) DeepCopyInto(out *EnvoyPatchPolicyStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]v1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new EnvoyPatchPolicyStatus.
func (in *EnvoyPatchPolicyStatus) DeepCopy() *EnvoyPatchPolicyStatus {
	if in == nil {
		return nil
	}
	out := new(EnvoyPatchPolicyStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GlobalRateLimit) DeepCopyInto(out *GlobalRateLimit) {
	*out = *in
	if in.Rules != nil {
		in, out := &in.Rules, &out.Rules
		*out = make([]RateLimitRule, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GlobalRateLimit.
func (in *GlobalRateLimit) DeepCopy() *GlobalRateLimit {
	if in == nil {
		return nil
	}
	out := new(GlobalRateLimit)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GrpcJSONTranscoderFilter) DeepCopyInto(out *GrpcJSONTranscoderFilter) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GrpcJSONTranscoderFilter.
func (in *GrpcJSONTranscoderFilter) DeepCopy() *GrpcJSONTranscoderFilter {
	if in == nil {
		return nil
	}
	out := new(GrpcJSONTranscoderFilter)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *GrpcJSONTranscoderFilter) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GrpcJSONTranscoderFilterList) DeepCopyInto(out *GrpcJSONTranscoderFilterList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]GrpcJSONTranscoderFilter, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GrpcJSONTranscoderFilterList.
func (in *GrpcJSONTranscoderFilterList) DeepCopy() *GrpcJSONTranscoderFilterList {
	if in == nil {
		return nil
	}
	out := new(GrpcJSONTranscoderFilterList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *GrpcJSONTranscoderFilterList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GrpcJSONTranscoderFilterSpec) DeepCopyInto(out *GrpcJSONTranscoderFilterSpec) {
	*out = *in
	if in.Services != nil {
		in, out := &in.Services, &out.Services
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.PrintOptions != nil {
		in, out := &in.PrintOptions, &out.PrintOptions
		*out = new(PrintOptions)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GrpcJSONTranscoderFilterSpec.
func (in *GrpcJSONTranscoderFilterSpec) DeepCopy() *GrpcJSONTranscoderFilterSpec {
	if in == nil {
		return nil
	}
	out := new(GrpcJSONTranscoderFilterSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HeaderMatch) DeepCopyInto(out *HeaderMatch) {
	*out = *in
	if in.Type != nil {
		in, out := &in.Type, &out.Type
		*out = new(HeaderMatchType)
		**out = **in
	}
	if in.Value != nil {
		in, out := &in.Value, &out.Value
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HeaderMatch.
func (in *HeaderMatch) DeepCopy() *HeaderMatch {
	if in == nil {
		return nil
	}
	out := new(HeaderMatch)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *JSONPatchOperation) DeepCopyInto(out *JSONPatchOperation) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new JSONPatchOperation.
func (in *JSONPatchOperation) DeepCopy() *JSONPatchOperation {
	if in == nil {
		return nil
	}
	out := new(JSONPatchOperation)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *JwtAuthenticationFilterProvider) DeepCopyInto(out *JwtAuthenticationFilterProvider) {
	*out = *in
	if in.Audiences != nil {
		in, out := &in.Audiences, &out.Audiences
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	out.RemoteJWKS = in.RemoteJWKS
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new JwtAuthenticationFilterProvider.
func (in *JwtAuthenticationFilterProvider) DeepCopy() *JwtAuthenticationFilterProvider {
	if in == nil {
		return nil
	}
	out := new(JwtAuthenticationFilterProvider)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PrintOptions) DeepCopyInto(out *PrintOptions) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PrintOptions.
func (in *PrintOptions) DeepCopy() *PrintOptions {
	if in == nil {
		return nil
	}
	out := new(PrintOptions)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RateLimitFilter) DeepCopyInto(out *RateLimitFilter) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RateLimitFilter.
func (in *RateLimitFilter) DeepCopy() *RateLimitFilter {
	if in == nil {
		return nil
	}
	out := new(RateLimitFilter)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *RateLimitFilter) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RateLimitFilterList) DeepCopyInto(out *RateLimitFilterList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]RateLimitFilter, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RateLimitFilterList.
func (in *RateLimitFilterList) DeepCopy() *RateLimitFilterList {
	if in == nil {
		return nil
	}
	out := new(RateLimitFilterList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *RateLimitFilterList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RateLimitFilterSpec) DeepCopyInto(out *RateLimitFilterSpec) {
	*out = *in
	if in.Global != nil {
		in, out := &in.Global, &out.Global
		*out = new(GlobalRateLimit)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RateLimitFilterSpec.
func (in *RateLimitFilterSpec) DeepCopy() *RateLimitFilterSpec {
	if in == nil {
		return nil
	}
	out := new(RateLimitFilterSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RateLimitRule) DeepCopyInto(out *RateLimitRule) {
	*out = *in
	if in.ClientSelectors != nil {
		in, out := &in.ClientSelectors, &out.ClientSelectors
		*out = make([]RateLimitSelectCondition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	out.Limit = in.Limit
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RateLimitRule.
func (in *RateLimitRule) DeepCopy() *RateLimitRule {
	if in == nil {
		return nil
	}
	out := new(RateLimitRule)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RateLimitSelectCondition) DeepCopyInto(out *RateLimitSelectCondition) {
	*out = *in
	if in.Headers != nil {
		in, out := &in.Headers, &out.Headers
		*out = make([]HeaderMatch, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.SourceIP != nil {
		in, out := &in.SourceIP, &out.SourceIP
		*out = new(string)
		**out = **in
	}
	if in.SourceCIDR != nil {
		in, out := &in.SourceCIDR, &out.SourceCIDR
		*out = new(SourceMatch)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RateLimitSelectCondition.
func (in *RateLimitSelectCondition) DeepCopy() *RateLimitSelectCondition {
	if in == nil {
		return nil
	}
	out := new(RateLimitSelectCondition)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RateLimitValue) DeepCopyInto(out *RateLimitValue) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RateLimitValue.
func (in *RateLimitValue) DeepCopy() *RateLimitValue {
	if in == nil {
		return nil
	}
	out := new(RateLimitValue)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RemoteJWKS) DeepCopyInto(out *RemoteJWKS) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RemoteJWKS.
func (in *RemoteJWKS) DeepCopy() *RemoteJWKS {
	if in == nil {
		return nil
	}
	out := new(RemoteJWKS)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SourceMatch) DeepCopyInto(out *SourceMatch) {
	*out = *in
	if in.Type != nil {
		in, out := &in.Type, &out.Type
		*out = new(SourceMatchType)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SourceMatch.
func (in *SourceMatch) DeepCopy() *SourceMatch {
	if in == nil {
		return nil
	}
	out := new(SourceMatch)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StringMatch) DeepCopyInto(out *StringMatch) {
	*out = *in
	if in.Exact != nil {
		in, out := &in.Exact, &out.Exact
		*out = new(string)
		**out = **in
	}
	if in.Prefix != nil {
		in, out := &in.Prefix, &out.Prefix
		*out = new(string)
		**out = **in
	}
	if in.Regex != nil {
		in, out := &in.Regex, &out.Regex
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StringMatch.
func (in *StringMatch) DeepCopy() *StringMatch {
	if in == nil {
		return nil
	}
	out := new(StringMatch)
	in.DeepCopyInto(out)
	return out
}
