// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             (unknown)
// source: proto/extension/service.proto

package extension

import (
	context "context"

	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

const (
	EnvoyGatewayExtension_PostRouteModify_FullMethodName        = "/envoygateway.extension.EnvoyGatewayExtension/PostRouteModify"
	EnvoyGatewayExtension_PostVirtualHostModify_FullMethodName  = "/envoygateway.extension.EnvoyGatewayExtension/PostVirtualHostModify"
	EnvoyGatewayExtension_PostHTTPListenerModify_FullMethodName = "/envoygateway.extension.EnvoyGatewayExtension/PostHTTPListenerModify"
	EnvoyGatewayExtension_PostTranslateModify_FullMethodName    = "/envoygateway.extension.EnvoyGatewayExtension/PostTranslateModify"
)

// EnvoyGatewayExtensionClient is the client API for EnvoyGatewayExtension service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type EnvoyGatewayExtensionClient interface {
	// PostRouteModify provides a way for extensions to modify a route generated by Envoy Gateway before it is finalized.
	// Doing so allows extensions to configure/modify route fields configured by Envoy Gateway and also to configure the
	// Route's TypedPerFilterConfig which may be desirable to do things such as pass settings and information to
	// ext_authz filters.
	// PostRouteModify also passes a list of Unstructured data for the externalRefs owned by the extension on the HTTPRoute that
	// created this xDS route
	// PostRouteModify will only be executed if an extension is loaded and only on Routes which were generated from an HTTPRoute
	// that uses extension resources as externalRef filters.
	PostRouteModify(ctx context.Context, in *PostRouteModifyRequest, opts ...grpc.CallOption) (*PostRouteModifyResponse, error)
	// PostVirtualHostModify provides a way for extensions to modify a VirtualHost generated by Envoy Gateway before it is finalized.
	// An extension can also make use of this hook to generate and insert entirely new Routes not generated by Envoy Gateway.
	// PostVirtualHostModify is always executed when an extension is loaded. An extension may return nil to not make any changes
	// to it.
	PostVirtualHostModify(ctx context.Context, in *PostVirtualHostModifyRequest, opts ...grpc.CallOption) (*PostVirtualHostModifyResponse, error)
	// PostHTTPListenerModify allows an extension to make changes to a Listener generated by Envoy Gateway before it is finalized.
	// PostHTTPListenerModify is always executed when an extension is loaded. An extension may return nil
	// in order to not make any changes to it.
	PostHTTPListenerModify(ctx context.Context, in *PostHTTPListenerModifyRequest, opts ...grpc.CallOption) (*PostHTTPListenerModifyResponse, error)
	// PostTranslateModify allows an extension to modify the clusters and secrets in the xDS config.
	// This allows for inserting clusters that may change along with extension specific configuration to be dynamically created rather than
	// using custom bootstrap config which would be sufficient for clusters that are static and not prone to have their configurations changed.
	// An example of how this may be used is to inject a cluster that will be used by an ext_authz http filter created by the extension.
	// The list of clusters and secrets returned by the extension are used as the final list of all clusters and secrets
	// PostTranslateModify is always executed when an extension is loaded
	PostTranslateModify(ctx context.Context, in *PostTranslateModifyRequest, opts ...grpc.CallOption) (*PostTranslateModifyResponse, error)
}

type envoyGatewayExtensionClient struct {
	cc grpc.ClientConnInterface
}

func NewEnvoyGatewayExtensionClient(cc grpc.ClientConnInterface) EnvoyGatewayExtensionClient {
	return &envoyGatewayExtensionClient{cc}
}

func (c *envoyGatewayExtensionClient) PostRouteModify(ctx context.Context, in *PostRouteModifyRequest, opts ...grpc.CallOption) (*PostRouteModifyResponse, error) {
	out := new(PostRouteModifyResponse)
	err := c.cc.Invoke(ctx, EnvoyGatewayExtension_PostRouteModify_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *envoyGatewayExtensionClient) PostVirtualHostModify(ctx context.Context, in *PostVirtualHostModifyRequest, opts ...grpc.CallOption) (*PostVirtualHostModifyResponse, error) {
	out := new(PostVirtualHostModifyResponse)
	err := c.cc.Invoke(ctx, EnvoyGatewayExtension_PostVirtualHostModify_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *envoyGatewayExtensionClient) PostHTTPListenerModify(ctx context.Context, in *PostHTTPListenerModifyRequest, opts ...grpc.CallOption) (*PostHTTPListenerModifyResponse, error) {
	out := new(PostHTTPListenerModifyResponse)
	err := c.cc.Invoke(ctx, EnvoyGatewayExtension_PostHTTPListenerModify_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *envoyGatewayExtensionClient) PostTranslateModify(ctx context.Context, in *PostTranslateModifyRequest, opts ...grpc.CallOption) (*PostTranslateModifyResponse, error) {
	out := new(PostTranslateModifyResponse)
	err := c.cc.Invoke(ctx, EnvoyGatewayExtension_PostTranslateModify_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// EnvoyGatewayExtensionServer is the server API for EnvoyGatewayExtension service.
// All implementations must embed UnimplementedEnvoyGatewayExtensionServer
// for forward compatibility
type EnvoyGatewayExtensionServer interface {
	// PostRouteModify provides a way for extensions to modify a route generated by Envoy Gateway before it is finalized.
	// Doing so allows extensions to configure/modify route fields configured by Envoy Gateway and also to configure the
	// Route's TypedPerFilterConfig which may be desirable to do things such as pass settings and information to
	// ext_authz filters.
	// PostRouteModify also passes a list of Unstructured data for the externalRefs owned by the extension on the HTTPRoute that
	// created this xDS route
	// PostRouteModify will only be executed if an extension is loaded and only on Routes which were generated from an HTTPRoute
	// that uses extension resources as externalRef filters.
	PostRouteModify(context.Context, *PostRouteModifyRequest) (*PostRouteModifyResponse, error)
	// PostVirtualHostModify provides a way for extensions to modify a VirtualHost generated by Envoy Gateway before it is finalized.
	// An extension can also make use of this hook to generate and insert entirely new Routes not generated by Envoy Gateway.
	// PostVirtualHostModify is always executed when an extension is loaded. An extension may return nil to not make any changes
	// to it.
	PostVirtualHostModify(context.Context, *PostVirtualHostModifyRequest) (*PostVirtualHostModifyResponse, error)
	// PostHTTPListenerModify allows an extension to make changes to a Listener generated by Envoy Gateway before it is finalized.
	// PostHTTPListenerModify is always executed when an extension is loaded. An extension may return nil
	// in order to not make any changes to it.
	PostHTTPListenerModify(context.Context, *PostHTTPListenerModifyRequest) (*PostHTTPListenerModifyResponse, error)
	// PostTranslateModify allows an extension to modify the clusters and secrets in the xDS config.
	// This allows for inserting clusters that may change along with extension specific configuration to be dynamically created rather than
	// using custom bootstrap config which would be sufficient for clusters that are static and not prone to have their configurations changed.
	// An example of how this may be used is to inject a cluster that will be used by an ext_authz http filter created by the extension.
	// The list of clusters and secrets returned by the extension are used as the final list of all clusters and secrets
	// PostTranslateModify is always executed when an extension is loaded
	PostTranslateModify(context.Context, *PostTranslateModifyRequest) (*PostTranslateModifyResponse, error)
	mustEmbedUnimplementedEnvoyGatewayExtensionServer()
}

// UnimplementedEnvoyGatewayExtensionServer must be embedded to have forward compatible implementations.
type UnimplementedEnvoyGatewayExtensionServer struct {
}

func (UnimplementedEnvoyGatewayExtensionServer) PostRouteModify(context.Context, *PostRouteModifyRequest) (*PostRouteModifyResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PostRouteModify not implemented")
}
func (UnimplementedEnvoyGatewayExtensionServer) PostVirtualHostModify(context.Context, *PostVirtualHostModifyRequest) (*PostVirtualHostModifyResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PostVirtualHostModify not implemented")
}
func (UnimplementedEnvoyGatewayExtensionServer) PostHTTPListenerModify(context.Context, *PostHTTPListenerModifyRequest) (*PostHTTPListenerModifyResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PostHTTPListenerModify not implemented")
}
func (UnimplementedEnvoyGatewayExtensionServer) PostTranslateModify(context.Context, *PostTranslateModifyRequest) (*PostTranslateModifyResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PostTranslateModify not implemented")
}
func (UnimplementedEnvoyGatewayExtensionServer) mustEmbedUnimplementedEnvoyGatewayExtensionServer() {}

// UnsafeEnvoyGatewayExtensionServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to EnvoyGatewayExtensionServer will
// result in compilation errors.
type UnsafeEnvoyGatewayExtensionServer interface {
	mustEmbedUnimplementedEnvoyGatewayExtensionServer()
}

func RegisterEnvoyGatewayExtensionServer(s grpc.ServiceRegistrar, srv EnvoyGatewayExtensionServer) {
	s.RegisterService(&EnvoyGatewayExtension_ServiceDesc, srv)
}

func _EnvoyGatewayExtension_PostRouteModify_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PostRouteModifyRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(EnvoyGatewayExtensionServer).PostRouteModify(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: EnvoyGatewayExtension_PostRouteModify_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(EnvoyGatewayExtensionServer).PostRouteModify(ctx, req.(*PostRouteModifyRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _EnvoyGatewayExtension_PostVirtualHostModify_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PostVirtualHostModifyRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(EnvoyGatewayExtensionServer).PostVirtualHostModify(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: EnvoyGatewayExtension_PostVirtualHostModify_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(EnvoyGatewayExtensionServer).PostVirtualHostModify(ctx, req.(*PostVirtualHostModifyRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _EnvoyGatewayExtension_PostHTTPListenerModify_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PostHTTPListenerModifyRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(EnvoyGatewayExtensionServer).PostHTTPListenerModify(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: EnvoyGatewayExtension_PostHTTPListenerModify_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(EnvoyGatewayExtensionServer).PostHTTPListenerModify(ctx, req.(*PostHTTPListenerModifyRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _EnvoyGatewayExtension_PostTranslateModify_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PostTranslateModifyRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(EnvoyGatewayExtensionServer).PostTranslateModify(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: EnvoyGatewayExtension_PostTranslateModify_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(EnvoyGatewayExtensionServer).PostTranslateModify(ctx, req.(*PostTranslateModifyRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// EnvoyGatewayExtension_ServiceDesc is the grpc.ServiceDesc for EnvoyGatewayExtension service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var EnvoyGatewayExtension_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "envoygateway.extension.EnvoyGatewayExtension",
	HandlerType: (*EnvoyGatewayExtensionServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "PostRouteModify",
			Handler:    _EnvoyGatewayExtension_PostRouteModify_Handler,
		},
		{
			MethodName: "PostVirtualHostModify",
			Handler:    _EnvoyGatewayExtension_PostVirtualHostModify_Handler,
		},
		{
			MethodName: "PostHTTPListenerModify",
			Handler:    _EnvoyGatewayExtension_PostHTTPListenerModify_Handler,
		},
		{
			MethodName: "PostTranslateModify",
			Handler:    _EnvoyGatewayExtension_PostTranslateModify_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/extension/service.proto",
}
