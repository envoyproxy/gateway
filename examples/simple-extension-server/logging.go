// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package main

import (
	"encoding/json"
	"log/slog"

	tlsv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	pb "github.com/envoyproxy/gateway/proto/extension"
)

// logRequest logs that hook was called, with the name of the resource it was called for and the
// identity (apiVersion, kind, namespace and name) of each extension resource it received, e.g. the
// InferencePools passed to each cluster.
//
// The request and the full extension resources are logged only with --log-requests, because they
// can hold credentials: the Translation hook receives every xDS Secret, with inline TLS private
// keys, listeners and clusters can carry inline authentication material, and an extension resource
// can hold anything. Even then, Secrets are logged by name only. Extension resources are bytes
// fields, so they are logged decoded. The request is logged as nested JSON. Envoy extension types in typed_config fields are
// registered in extensions.gen.go; if one is still unknown, the request is logged in text format
// instead, which leaves such fields undecoded rather than failing.
func (s *Server) logRequest(hook string, req proto.Message) {
	attrs := []any{slog.String("hook", hook)}
	if name := resourceName(req); name != "" {
		attrs = append(attrs, slog.String("name", name))
	}

	if s.logRequests {
		req = redactSecrets(req)
		if body, err := protojson.Marshal(req); err == nil {
			attrs = append(attrs, slog.Any("request", json.RawMessage(body)))
		} else {
			attrs = append(attrs,
				slog.String("request", prototext.Format(req)),
				slog.String("requestJSONError", err.Error()))
		}
	}

	var resources []any
	for _, res := range extensionResources(req.ProtoReflect()) {
		b := res.GetUnstructuredBytes()
		switch {
		case !s.logRequests:
			resources = append(resources, resourceIdentity(b))
		case json.Valid(b):
			resources = append(resources, json.RawMessage(b))
		default:
			resources = append(resources, string(b))
		}
	}
	if len(resources) > 0 {
		attrs = append(attrs, slog.Any("extensionResources", resources))
	}

	s.log.Info(hook+" callback was invoked", attrs...)
}

// resourceIdentity returns "<apiVersion>/<kind>/<namespace>/<name>" for an extension resource,
// without any of its other fields.
func resourceIdentity(b []byte) string {
	var obj struct {
		APIVersion string `json:"apiVersion"`
		Kind       string `json:"kind"`
		Metadata   struct {
			Name      string `json:"name"`
			Namespace string `json:"namespace"`
		} `json:"metadata"`
	}
	if err := json.Unmarshal(b, &obj); err != nil {
		return "<invalid extension resource>"
	}
	return obj.APIVersion + "/" + obj.Kind + "/" + obj.Metadata.Namespace + "/" + obj.Metadata.Name
}

// resourceName returns the name of the xDS resource a hook was called for, if it has one.
func resourceName(req proto.Message) string {
	switch r := req.(type) {
	case *pb.PostRouteModifyRequest:
		return r.GetRoute().GetName()
	case *pb.PostVirtualHostModifyRequest:
		return r.GetVirtualHost().GetName()
	case *pb.PostHTTPListenerModifyRequest:
		return r.GetListener().GetName()
	case *pb.PostClusterModifyRequest:
		return r.GetCluster().GetName()
	}
	return ""
}

// redactSecrets returns req with every xDS Secret reduced to its name, so private keys and other
// secret material are never logged. req itself is not modified.
func redactSecrets(req proto.Message) proto.Message {
	r, ok := req.(*pb.PostTranslateModifyRequest)
	if !ok || len(r.GetSecrets()) == 0 {
		return req
	}
	redacted := proto.Clone(r).(*pb.PostTranslateModifyRequest)
	for i, secret := range redacted.Secrets {
		redacted.Secrets[i] = &tlsv3.Secret{Name: secret.GetName()}
	}
	return redacted
}

// extensionResources returns every ExtensionResource found anywhere in msg.
func extensionResources(msg protoreflect.Message) []*pb.ExtensionResource {
	if res, ok := msg.Interface().(*pb.ExtensionResource); ok {
		return []*pb.ExtensionResource{res}
	}
	var out []*pb.ExtensionResource
	msg.Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		if fd.Kind() != protoreflect.MessageKind || fd.IsMap() {
			return true
		}
		if fd.IsList() {
			for i := range v.List().Len() {
				out = append(out, extensionResources(v.List().Get(i).Message())...)
			}
			return true
		}
		out = append(out, extensionResources(v.Message())...)
		return true
	})
	return out
}
