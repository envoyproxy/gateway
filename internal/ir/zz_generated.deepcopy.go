//go:build !ignore_autogenerated

// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

// Code generated by controller-gen. DO NOT EDIT.

package ir

import (
	"github.com/envoyproxy/gateway/api/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AccessLog) DeepCopyInto(out *AccessLog) {
	*out = *in
	if in.Text != nil {
		in, out := &in.Text, &out.Text
		*out = make([]*TextAccessLog, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(TextAccessLog)
				(*in).DeepCopyInto(*out)
			}
		}
	}
	if in.JSON != nil {
		in, out := &in.JSON, &out.JSON
		*out = make([]*JSONAccessLog, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(JSONAccessLog)
				(*in).DeepCopyInto(*out)
			}
		}
	}
	if in.OpenTelemetry != nil {
		in, out := &in.OpenTelemetry, &out.OpenTelemetry
		*out = make([]*OpenTelemetryAccessLog, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(OpenTelemetryAccessLog)
				(*in).DeepCopyInto(*out)
			}
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AccessLog.
func (in *AccessLog) DeepCopy() *AccessLog {
	if in == nil {
		return nil
	}
	out := new(AccessLog)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AddHeader) DeepCopyInto(out *AddHeader) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AddHeader.
func (in *AddHeader) DeepCopy() *AddHeader {
	if in == nil {
		return nil
	}
	out := new(AddHeader)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *BasicAuth) DeepCopyInto(out *BasicAuth) {
	*out = *in
	if in.Users != nil {
		in, out := &in.Users, &out.Users
		*out = make([]byte, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new BasicAuth.
func (in *BasicAuth) DeepCopy() *BasicAuth {
	if in == nil {
		return nil
	}
	out := new(BasicAuth)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CORS) DeepCopyInto(out *CORS) {
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
	if in.MaxAge != nil {
		in, out := &in.MaxAge, &out.MaxAge
		*out = new(v1.Duration)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CORS.
func (in *CORS) DeepCopy() *CORS {
	if in == nil {
		return nil
	}
	out := new(CORS)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ConsistentHash) DeepCopyInto(out *ConsistentHash) {
	*out = *in
	if in.SourceIP != nil {
		in, out := &in.SourceIP, &out.SourceIP
		*out = new(bool)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ConsistentHash.
func (in *ConsistentHash) DeepCopy() *ConsistentHash {
	if in == nil {
		return nil
	}
	out := new(ConsistentHash)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DestinationEndpoint) DeepCopyInto(out *DestinationEndpoint) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DestinationEndpoint.
func (in *DestinationEndpoint) DeepCopy() *DestinationEndpoint {
	if in == nil {
		return nil
	}
	out := new(DestinationEndpoint)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DestinationSetting) DeepCopyInto(out *DestinationSetting) {
	*out = *in
	if in.Weight != nil {
		in, out := &in.Weight, &out.Weight
		*out = new(uint32)
		**out = **in
	}
	if in.Endpoints != nil {
		in, out := &in.Endpoints, &out.Endpoints
		*out = make([]*DestinationEndpoint, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(DestinationEndpoint)
				**out = **in
			}
		}
	}
	if in.AddressType != nil {
		in, out := &in.AddressType, &out.AddressType
		*out = new(DestinationAddressType)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DestinationSetting.
func (in *DestinationSetting) DeepCopy() *DestinationSetting {
	if in == nil {
		return nil
	}
	out := new(DestinationSetting)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DirectResponse) DeepCopyInto(out *DirectResponse) {
	*out = *in
	if in.Body != nil {
		in, out := &in.Body, &out.Body
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DirectResponse.
func (in *DirectResponse) DeepCopy() *DirectResponse {
	if in == nil {
		return nil
	}
	out := new(DirectResponse)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *EnvoyPatchPolicy) DeepCopyInto(out *EnvoyPatchPolicy) {
	*out = *in
	in.EnvoyPatchPolicyStatus.DeepCopyInto(&out.EnvoyPatchPolicyStatus)
	if in.JSONPatches != nil {
		in, out := &in.JSONPatches, &out.JSONPatches
		*out = make([]*JSONPatchConfig, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(JSONPatchConfig)
				(*in).DeepCopyInto(*out)
			}
		}
	}
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

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *EnvoyPatchPolicyStatus) DeepCopyInto(out *EnvoyPatchPolicyStatus) {
	*out = *in
	if in.Status != nil {
		in, out := &in.Status, &out.Status
		*out = new(v1alpha1.EnvoyPatchPolicyStatus)
		(*in).DeepCopyInto(*out)
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
		*out = make([]*RateLimitRule, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(RateLimitRule)
				(*in).DeepCopyInto(*out)
			}
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
func (in *HTTPListener) DeepCopyInto(out *HTTPListener) {
	*out = *in
	if in.Hostnames != nil {
		in, out := &in.Hostnames, &out.Hostnames
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.TLS != nil {
		in, out := &in.TLS, &out.TLS
		*out = make([]*TLSListenerConfig, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(TLSListenerConfig)
				(*in).DeepCopyInto(*out)
			}
		}
	}
	if in.Routes != nil {
		in, out := &in.Routes, &out.Routes
		*out = make([]*HTTPRoute, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(HTTPRoute)
				(*in).DeepCopyInto(*out)
			}
		}
	}
	if in.TCPKeepalive != nil {
		in, out := &in.TCPKeepalive, &out.TCPKeepalive
		*out = new(TCPKeepalive)
		(*in).DeepCopyInto(*out)
	}
	if in.HTTP3 != nil {
		in, out := &in.HTTP3, &out.HTTP3
		*out = new(HTTP3Settings)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HTTPListener.
func (in *HTTPListener) DeepCopy() *HTTPListener {
	if in == nil {
		return nil
	}
	out := new(HTTPListener)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HTTPPathModifier) DeepCopyInto(out *HTTPPathModifier) {
	*out = *in
	if in.FullReplace != nil {
		in, out := &in.FullReplace, &out.FullReplace
		*out = new(string)
		**out = **in
	}
	if in.PrefixMatchReplace != nil {
		in, out := &in.PrefixMatchReplace, &out.PrefixMatchReplace
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HTTPPathModifier.
func (in *HTTPPathModifier) DeepCopy() *HTTPPathModifier {
	if in == nil {
		return nil
	}
	out := new(HTTPPathModifier)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HTTPRoute) DeepCopyInto(out *HTTPRoute) {
	*out = *in
	if in.PathMatch != nil {
		in, out := &in.PathMatch, &out.PathMatch
		*out = new(StringMatch)
		(*in).DeepCopyInto(*out)
	}
	if in.HeaderMatches != nil {
		in, out := &in.HeaderMatches, &out.HeaderMatches
		*out = make([]*StringMatch, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(StringMatch)
				(*in).DeepCopyInto(*out)
			}
		}
	}
	if in.QueryParamMatches != nil {
		in, out := &in.QueryParamMatches, &out.QueryParamMatches
		*out = make([]*StringMatch, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(StringMatch)
				(*in).DeepCopyInto(*out)
			}
		}
	}
	out.BackendWeights = in.BackendWeights
	if in.AddRequestHeaders != nil {
		in, out := &in.AddRequestHeaders, &out.AddRequestHeaders
		*out = make([]AddHeader, len(*in))
		copy(*out, *in)
	}
	if in.RemoveRequestHeaders != nil {
		in, out := &in.RemoveRequestHeaders, &out.RemoveRequestHeaders
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.AddResponseHeaders != nil {
		in, out := &in.AddResponseHeaders, &out.AddResponseHeaders
		*out = make([]AddHeader, len(*in))
		copy(*out, *in)
	}
	if in.RemoveResponseHeaders != nil {
		in, out := &in.RemoveResponseHeaders, &out.RemoveResponseHeaders
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.DirectResponse != nil {
		in, out := &in.DirectResponse, &out.DirectResponse
		*out = new(DirectResponse)
		(*in).DeepCopyInto(*out)
	}
	if in.Redirect != nil {
		in, out := &in.Redirect, &out.Redirect
		*out = new(Redirect)
		(*in).DeepCopyInto(*out)
	}
	if in.Mirrors != nil {
		in, out := &in.Mirrors, &out.Mirrors
		*out = make([]*RouteDestination, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(RouteDestination)
				(*in).DeepCopyInto(*out)
			}
		}
	}
	if in.Destination != nil {
		in, out := &in.Destination, &out.Destination
		*out = new(RouteDestination)
		(*in).DeepCopyInto(*out)
	}
	if in.URLRewrite != nil {
		in, out := &in.URLRewrite, &out.URLRewrite
		*out = new(URLRewrite)
		(*in).DeepCopyInto(*out)
	}
	if in.RateLimit != nil {
		in, out := &in.RateLimit, &out.RateLimit
		*out = new(RateLimit)
		(*in).DeepCopyInto(*out)
	}
	if in.Timeout != nil {
		in, out := &in.Timeout, &out.Timeout
		*out = new(v1.Duration)
		**out = **in
	}
	if in.LoadBalancer != nil {
		in, out := &in.LoadBalancer, &out.LoadBalancer
		*out = new(LoadBalancer)
		(*in).DeepCopyInto(*out)
	}
	if in.CORS != nil {
		in, out := &in.CORS, &out.CORS
		*out = new(CORS)
		(*in).DeepCopyInto(*out)
	}
	if in.JWT != nil {
		in, out := &in.JWT, &out.JWT
		*out = new(JWT)
		(*in).DeepCopyInto(*out)
	}
	if in.OIDC != nil {
		in, out := &in.OIDC, &out.OIDC
		*out = new(OIDC)
		(*in).DeepCopyInto(*out)
	}
	if in.ProxyProtocol != nil {
		in, out := &in.ProxyProtocol, &out.ProxyProtocol
		*out = new(ProxyProtocol)
		**out = **in
	}
	if in.BasicAuth != nil {
		in, out := &in.BasicAuth, &out.BasicAuth
		*out = new(BasicAuth)
		(*in).DeepCopyInto(*out)
	}
	if in.ExtensionRefs != nil {
		in, out := &in.ExtensionRefs, &out.ExtensionRefs
		*out = make([]*UnstructuredRef, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(UnstructuredRef)
				(*in).DeepCopyInto(*out)
			}
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HTTPRoute.
func (in *HTTPRoute) DeepCopy() *HTTPRoute {
	if in == nil {
		return nil
	}
	out := new(HTTPRoute)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Infra) DeepCopyInto(out *Infra) {
	*out = *in
	if in.Proxy != nil {
		in, out := &in.Proxy, &out.Proxy
		*out = new(ProxyInfra)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Infra.
func (in *Infra) DeepCopy() *Infra {
	if in == nil {
		return nil
	}
	out := new(Infra)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *InfraMetadata) DeepCopyInto(out *InfraMetadata) {
	*out = *in
	if in.Annotations != nil {
		in, out := &in.Annotations, &out.Annotations
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Labels != nil {
		in, out := &in.Labels, &out.Labels
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new InfraMetadata.
func (in *InfraMetadata) DeepCopy() *InfraMetadata {
	if in == nil {
		return nil
	}
	out := new(InfraMetadata)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *JSONAccessLog) DeepCopyInto(out *JSONAccessLog) {
	*out = *in
	if in.JSON != nil {
		in, out := &in.JSON, &out.JSON
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new JSONAccessLog.
func (in *JSONAccessLog) DeepCopy() *JSONAccessLog {
	if in == nil {
		return nil
	}
	out := new(JSONAccessLog)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *JSONPatchConfig) DeepCopyInto(out *JSONPatchConfig) {
	*out = *in
	in.Operation.DeepCopyInto(&out.Operation)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new JSONPatchConfig.
func (in *JSONPatchConfig) DeepCopy() *JSONPatchConfig {
	if in == nil {
		return nil
	}
	out := new(JSONPatchConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *JSONPatchOperation) DeepCopyInto(out *JSONPatchOperation) {
	*out = *in
	in.Value.DeepCopyInto(&out.Value)
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
func (in *JWT) DeepCopyInto(out *JWT) {
	*out = *in
	if in.Providers != nil {
		in, out := &in.Providers, &out.Providers
		*out = make([]v1alpha1.JWTProvider, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new JWT.
func (in *JWT) DeepCopy() *JWT {
	if in == nil {
		return nil
	}
	out := new(JWT)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LeastRequest) DeepCopyInto(out *LeastRequest) {
	*out = *in
	if in.SlowStart != nil {
		in, out := &in.SlowStart, &out.SlowStart
		*out = new(SlowStart)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LeastRequest.
func (in *LeastRequest) DeepCopy() *LeastRequest {
	if in == nil {
		return nil
	}
	out := new(LeastRequest)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ListenerPort) DeepCopyInto(out *ListenerPort) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ListenerPort.
func (in *ListenerPort) DeepCopy() *ListenerPort {
	if in == nil {
		return nil
	}
	out := new(ListenerPort)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LoadBalancer) DeepCopyInto(out *LoadBalancer) {
	*out = *in
	if in.RoundRobin != nil {
		in, out := &in.RoundRobin, &out.RoundRobin
		*out = new(RoundRobin)
		(*in).DeepCopyInto(*out)
	}
	if in.LeastRequest != nil {
		in, out := &in.LeastRequest, &out.LeastRequest
		*out = new(LeastRequest)
		(*in).DeepCopyInto(*out)
	}
	if in.Random != nil {
		in, out := &in.Random, &out.Random
		*out = new(Random)
		**out = **in
	}
	if in.ConsistentHash != nil {
		in, out := &in.ConsistentHash, &out.ConsistentHash
		*out = new(ConsistentHash)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LoadBalancer.
func (in *LoadBalancer) DeepCopy() *LoadBalancer {
	if in == nil {
		return nil
	}
	out := new(LoadBalancer)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Metrics) DeepCopyInto(out *Metrics) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Metrics.
func (in *Metrics) DeepCopy() *Metrics {
	if in == nil {
		return nil
	}
	out := new(Metrics)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OIDC) DeepCopyInto(out *OIDC) {
	*out = *in
	out.Provider = in.Provider
	if in.ClientSecret != nil {
		in, out := &in.ClientSecret, &out.ClientSecret
		*out = make([]byte, len(*in))
		copy(*out, *in)
	}
	if in.Scopes != nil {
		in, out := &in.Scopes, &out.Scopes
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OIDC.
func (in *OIDC) DeepCopy() *OIDC {
	if in == nil {
		return nil
	}
	out := new(OIDC)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OpenTelemetryAccessLog) DeepCopyInto(out *OpenTelemetryAccessLog) {
	*out = *in
	if in.Text != nil {
		in, out := &in.Text, &out.Text
		*out = new(string)
		**out = **in
	}
	if in.Attributes != nil {
		in, out := &in.Attributes, &out.Attributes
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Resources != nil {
		in, out := &in.Resources, &out.Resources
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OpenTelemetryAccessLog.
func (in *OpenTelemetryAccessLog) DeepCopy() *OpenTelemetryAccessLog {
	if in == nil {
		return nil
	}
	out := new(OpenTelemetryAccessLog)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ProxyInfra) DeepCopyInto(out *ProxyInfra) {
	*out = *in
	if in.Metadata != nil {
		in, out := &in.Metadata, &out.Metadata
		*out = new(InfraMetadata)
		(*in).DeepCopyInto(*out)
	}
	if in.Config != nil {
		in, out := &in.Config, &out.Config
		*out = new(v1alpha1.EnvoyProxy)
		(*in).DeepCopyInto(*out)
	}
	if in.Listeners != nil {
		in, out := &in.Listeners, &out.Listeners
		*out = make([]*ProxyListener, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(ProxyListener)
				(*in).DeepCopyInto(*out)
			}
		}
	}
	if in.Addresses != nil {
		in, out := &in.Addresses, &out.Addresses
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ProxyInfra.
func (in *ProxyInfra) DeepCopy() *ProxyInfra {
	if in == nil {
		return nil
	}
	out := new(ProxyInfra)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ProxyListener) DeepCopyInto(out *ProxyListener) {
	*out = *in
	if in.Address != nil {
		in, out := &in.Address, &out.Address
		*out = new(string)
		**out = **in
	}
	if in.Ports != nil {
		in, out := &in.Ports, &out.Ports
		*out = make([]ListenerPort, len(*in))
		copy(*out, *in)
	}
	if in.HTTP3 != nil {
		in, out := &in.HTTP3, &out.HTTP3
		*out = new(HTTP3Settings)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ProxyListener.
func (in *ProxyListener) DeepCopy() *ProxyListener {
	if in == nil {
		return nil
	}
	out := new(ProxyListener)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ProxyProtocol) DeepCopyInto(out *ProxyProtocol) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ProxyProtocol.
func (in *ProxyProtocol) DeepCopy() *ProxyProtocol {
	if in == nil {
		return nil
	}
	out := new(ProxyProtocol)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Random) DeepCopyInto(out *Random) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Random.
func (in *Random) DeepCopy() *Random {
	if in == nil {
		return nil
	}
	out := new(Random)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RateLimit) DeepCopyInto(out *RateLimit) {
	*out = *in
	if in.Global != nil {
		in, out := &in.Global, &out.Global
		*out = new(GlobalRateLimit)
		(*in).DeepCopyInto(*out)
	}
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
func (in *RateLimitRule) DeepCopyInto(out *RateLimitRule) {
	*out = *in
	if in.HeaderMatches != nil {
		in, out := &in.HeaderMatches, &out.HeaderMatches
		*out = make([]*StringMatch, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(StringMatch)
				(*in).DeepCopyInto(*out)
			}
		}
	}
	if in.CIDRMatch != nil {
		in, out := &in.CIDRMatch, &out.CIDRMatch
		*out = new(CIDRMatch)
		**out = **in
	}
	if in.Limit != nil {
		in, out := &in.Limit, &out.Limit
		*out = new(RateLimitValue)
		**out = **in
	}
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
func (in *Redirect) DeepCopyInto(out *Redirect) {
	*out = *in
	if in.Scheme != nil {
		in, out := &in.Scheme, &out.Scheme
		*out = new(string)
		**out = **in
	}
	if in.Hostname != nil {
		in, out := &in.Hostname, &out.Hostname
		*out = new(string)
		**out = **in
	}
	if in.Path != nil {
		in, out := &in.Path, &out.Path
		*out = new(HTTPPathModifier)
		(*in).DeepCopyInto(*out)
	}
	if in.Port != nil {
		in, out := &in.Port, &out.Port
		*out = new(uint32)
		**out = **in
	}
	if in.StatusCode != nil {
		in, out := &in.StatusCode, &out.StatusCode
		*out = new(int32)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Redirect.
func (in *Redirect) DeepCopy() *Redirect {
	if in == nil {
		return nil
	}
	out := new(Redirect)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RoundRobin) DeepCopyInto(out *RoundRobin) {
	*out = *in
	if in.SlowStart != nil {
		in, out := &in.SlowStart, &out.SlowStart
		*out = new(SlowStart)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RoundRobin.
func (in *RoundRobin) DeepCopy() *RoundRobin {
	if in == nil {
		return nil
	}
	out := new(RoundRobin)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RouteDestination) DeepCopyInto(out *RouteDestination) {
	*out = *in
	if in.Settings != nil {
		in, out := &in.Settings, &out.Settings
		*out = make([]*DestinationSetting, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(DestinationSetting)
				(*in).DeepCopyInto(*out)
			}
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RouteDestination.
func (in *RouteDestination) DeepCopy() *RouteDestination {
	if in == nil {
		return nil
	}
	out := new(RouteDestination)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SlowStart) DeepCopyInto(out *SlowStart) {
	*out = *in
	if in.Window != nil {
		in, out := &in.Window, &out.Window
		*out = new(v1.Duration)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SlowStart.
func (in *SlowStart) DeepCopy() *SlowStart {
	if in == nil {
		return nil
	}
	out := new(SlowStart)
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
	if in.Suffix != nil {
		in, out := &in.Suffix, &out.Suffix
		*out = new(string)
		**out = **in
	}
	if in.SafeRegex != nil {
		in, out := &in.SafeRegex, &out.SafeRegex
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

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TCPKeepalive) DeepCopyInto(out *TCPKeepalive) {
	*out = *in
	if in.Probes != nil {
		in, out := &in.Probes, &out.Probes
		*out = new(uint32)
		**out = **in
	}
	if in.IdleTime != nil {
		in, out := &in.IdleTime, &out.IdleTime
		*out = new(uint32)
		**out = **in
	}
	if in.Interval != nil {
		in, out := &in.Interval, &out.Interval
		*out = new(uint32)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TCPKeepalive.
func (in *TCPKeepalive) DeepCopy() *TCPKeepalive {
	if in == nil {
		return nil
	}
	out := new(TCPKeepalive)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TCPListener) DeepCopyInto(out *TCPListener) {
	*out = *in
	if in.TLS != nil {
		in, out := &in.TLS, &out.TLS
		*out = new(TLS)
		(*in).DeepCopyInto(*out)
	}
	if in.Destination != nil {
		in, out := &in.Destination, &out.Destination
		*out = new(RouteDestination)
		(*in).DeepCopyInto(*out)
	}
	if in.TCPKeepalive != nil {
		in, out := &in.TCPKeepalive, &out.TCPKeepalive
		*out = new(TCPKeepalive)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TCPListener.
func (in *TCPListener) DeepCopy() *TCPListener {
	if in == nil {
		return nil
	}
	out := new(TCPListener)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TLS) DeepCopyInto(out *TLS) {
	*out = *in
	if in.Passthrough != nil {
		in, out := &in.Passthrough, &out.Passthrough
		*out = new(TLSInspectorConfig)
		(*in).DeepCopyInto(*out)
	}
	if in.Terminate != nil {
		in, out := &in.Terminate, &out.Terminate
		*out = make([]*TLSListenerConfig, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(TLSListenerConfig)
				(*in).DeepCopyInto(*out)
			}
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TLS.
func (in *TLS) DeepCopy() *TLS {
	if in == nil {
		return nil
	}
	out := new(TLS)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TLSInspectorConfig) DeepCopyInto(out *TLSInspectorConfig) {
	*out = *in
	if in.SNIs != nil {
		in, out := &in.SNIs, &out.SNIs
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TLSInspectorConfig.
func (in *TLSInspectorConfig) DeepCopy() *TLSInspectorConfig {
	if in == nil {
		return nil
	}
	out := new(TLSInspectorConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TLSListenerConfig) DeepCopyInto(out *TLSListenerConfig) {
	*out = *in
	if in.ServerCertificate != nil {
		in, out := &in.ServerCertificate, &out.ServerCertificate
		*out = make([]byte, len(*in))
		copy(*out, *in)
	}
	if in.PrivateKey != nil {
		in, out := &in.PrivateKey, &out.PrivateKey
		*out = make([]byte, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TLSListenerConfig.
func (in *TLSListenerConfig) DeepCopy() *TLSListenerConfig {
	if in == nil {
		return nil
	}
	out := new(TLSListenerConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TextAccessLog) DeepCopyInto(out *TextAccessLog) {
	*out = *in
	if in.Format != nil {
		in, out := &in.Format, &out.Format
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TextAccessLog.
func (in *TextAccessLog) DeepCopy() *TextAccessLog {
	if in == nil {
		return nil
	}
	out := new(TextAccessLog)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Tracing) DeepCopyInto(out *Tracing) {
	*out = *in
	in.ProxyTracing.DeepCopyInto(&out.ProxyTracing)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Tracing.
func (in *Tracing) DeepCopy() *Tracing {
	if in == nil {
		return nil
	}
	out := new(Tracing)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *UDPListener) DeepCopyInto(out *UDPListener) {
	*out = *in
	if in.Destination != nil {
		in, out := &in.Destination, &out.Destination
		*out = new(RouteDestination)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new UDPListener.
func (in *UDPListener) DeepCopy() *UDPListener {
	if in == nil {
		return nil
	}
	out := new(UDPListener)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *URLRewrite) DeepCopyInto(out *URLRewrite) {
	*out = *in
	if in.Path != nil {
		in, out := &in.Path, &out.Path
		*out = new(HTTPPathModifier)
		(*in).DeepCopyInto(*out)
	}
	if in.Hostname != nil {
		in, out := &in.Hostname, &out.Hostname
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new URLRewrite.
func (in *URLRewrite) DeepCopy() *URLRewrite {
	if in == nil {
		return nil
	}
	out := new(URLRewrite)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *UnstructuredRef) DeepCopyInto(out *UnstructuredRef) {
	*out = *in
	if in.Object != nil {
		in, out := &in.Object, &out.Object
		*out = (*in).DeepCopy()
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new UnstructuredRef.
func (in *UnstructuredRef) DeepCopy() *UnstructuredRef {
	if in == nil {
		return nil
	}
	out := new(UnstructuredRef)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Xds) DeepCopyInto(out *Xds) {
	*out = *in
	if in.AccessLog != nil {
		in, out := &in.AccessLog, &out.AccessLog
		*out = new(AccessLog)
		(*in).DeepCopyInto(*out)
	}
	if in.Tracing != nil {
		in, out := &in.Tracing, &out.Tracing
		*out = new(Tracing)
		(*in).DeepCopyInto(*out)
	}
	if in.Metrics != nil {
		in, out := &in.Metrics, &out.Metrics
		*out = new(Metrics)
		**out = **in
	}
	if in.HTTP != nil {
		in, out := &in.HTTP, &out.HTTP
		*out = make([]*HTTPListener, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(HTTPListener)
				(*in).DeepCopyInto(*out)
			}
		}
	}
	if in.TCP != nil {
		in, out := &in.TCP, &out.TCP
		*out = make([]*TCPListener, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(TCPListener)
				(*in).DeepCopyInto(*out)
			}
		}
	}
	if in.UDP != nil {
		in, out := &in.UDP, &out.UDP
		*out = make([]*UDPListener, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(UDPListener)
				(*in).DeepCopyInto(*out)
			}
		}
	}
	if in.EnvoyPatchPolicies != nil {
		in, out := &in.EnvoyPatchPolicies, &out.EnvoyPatchPolicies
		*out = make([]*EnvoyPatchPolicy, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(EnvoyPatchPolicy)
				(*in).DeepCopyInto(*out)
			}
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Xds.
func (in *Xds) DeepCopy() *Xds {
	if in == nil {
		return nil
	}
	out := new(Xds)
	in.DeepCopyInto(out)
	return out
}
