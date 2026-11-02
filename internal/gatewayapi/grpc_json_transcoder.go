// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"encoding/base64"
	"errors"
	"fmt"
	"slices"
	"strings"
	"unicode"

	"google.golang.org/genproto/googleapis/api/annotations"
	spb "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/known/anypb"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/utils/ptr"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/gatewayapi/status"
	"github.com/envoyproxy/gateway/internal/ir"
)

// ProtoDescriptorConfigMapKey is the ConfigMap key of the protobuf descriptor set.
const ProtoDescriptorConfigMapKey = "proto-descriptor"

// maxServicesInError is the max number of services listed in an error message.
const maxServicesInError = 10

// parsedProtoDescriptor is a protobuf descriptor set loaded from a ConfigMap.
type parsedProtoDescriptor struct {
	// err is the error of loading the descriptor set, cached to avoid parsing it again.
	err   error
	bin   []byte
	fds   *descriptorpb.FileDescriptorSet
	files *protoregistry.Files
	// services are all the services in the descriptor set.
	services sets.Set[string]
	// defaultServices are the services not defined in an imported file.
	defaultServices []string
}

// applyGRPCJSONTranscoder adds the gRPC-JSON transcoder of the HTTPRouteFilter to the filter context.
func (t *Translator) applyGRPCJSONTranscoder(
	hrf *egv1a1.HTTPRouteFilter, filterContext *HTTPFiltersContext,
) status.Error {
	kind := string(egv1a1.KindHTTPRouteFilter)

	if filterContext.Route.GetRouteType() == resource.KindGRPCRoute {
		return t.processInvalidHTTPFilter(kind, filterContext,
			errors.New("grpcJSONTranscoder is not supported on a GRPCRoute"))
	}

	if filterContext.GRPCJSONTranscoder != nil {
		return t.processInvalidHTTPFilter(kind, filterContext,
			errors.New("cannot configure multiple grpcJSONTranscoder filters for a single HTTPRouteRule"))
	}

	transcoder, err := t.buildGRPCJSONTranscoder(
		hrf.Spec.GRPCJSONTranscoder, irConfigName(hrf), filterContext.Route.GetNamespace())
	if err != nil {
		return t.processInvalidHTTPFilter(kind, filterContext, err)
	}

	filterContext.GRPCJSONTranscoder = transcoder
	return nil
}

// useGRPCUpstream makes a backend of a transcoded route use HTTP/2 like the backend of a GRPCRoute,
// because the transcoded request is a gRPC request.
func useGRPCUpstream(ds *ir.DestinationSetting) status.Error {
	if ds.ForceHTTP1Upstream {
		return status.NewRouteStatusError(
			errors.New("grpcJSONTranscoder can't be used with a WebSocket backend, it needs HTTP/2"),
			gwapiv1.RouteReasonUnsupportedProtocol)
	}
	if ds.Protocol == ir.HTTP {
		ds.Protocol = ir.GRPC
	}
	return nil
}

func (t *Translator) buildGRPCJSONTranscoder(
	cfg *egv1a1.GRPCJSONTranscoder, name, namespace string,
) (*ir.GRPCJSONTranscoder, error) {
	descriptor, err := t.loadProtoDescriptor(cfg.ProtoDescriptor, namespace)
	if err != nil {
		return nil, err
	}

	services, err := getTranscodedServices(descriptor, cfg.Services)
	if err != nil {
		return nil, err
	}

	if err := validateHTTPRules(descriptor.files, services); err != nil {
		return nil, err
	}

	if ptr.Deref(cfg.ConvertGRPCStatus, false) {
		if err := validateGRPCStatusProtos(descriptor); err != nil {
			return nil, err
		}
	}

	return &ir.GRPCJSONTranscoder{
		Name:                         name,
		ProtoDescriptorBin:           descriptor.bin,
		Services:                     services,
		PrintOptions:                 cfg.PrintOptions,
		IgnoreUnknownQueryParameters: cfg.IgnoreUnknownQueryParameters,
		ConvertGRPCStatus:            cfg.ConvertGRPCStatus,
	}, nil
}

// loadProtoDescriptor returns the descriptor set referenced by protoDesc.
// The result is cached, so a descriptor set used by many routes is only parsed once.
func (t *Translator) loadProtoDescriptor(
	protoDesc egv1a1.ProtoDescriptor, namespace string,
) (*parsedProtoDescriptor, error) {
	ref := protoDesc.ValueRef
	if string(ref.Group) != "" || string(ref.Kind) != resource.KindConfigMap {
		return nil, fmt.Errorf("unsupported valueRef %s/%s, only ConfigMap is supported", ref.Group, ref.Kind)
	}

	key := types.NamespacedName{Namespace: namespace, Name: string(ref.Name)}
	if t.protoDescriptors == nil {
		t.protoDescriptors = make(map[types.NamespacedName]*parsedProtoDescriptor)
	}
	d, ok := t.protoDescriptors[key]
	if !ok {
		d = parseProtoDescriptor(t.GetConfigMap(key.Namespace, key.Name), key)
		t.protoDescriptors[key] = d
	}
	if d.err != nil {
		return nil, d.err
	}
	return d, nil
}

func parseProtoDescriptor(cm *corev1.ConfigMap, key types.NamespacedName) *parsedProtoDescriptor {
	bin, err := readProtoDescriptor(cm, key)
	if err != nil {
		return &parsedProtoDescriptor{err: err}
	}

	fds := &descriptorpb.FileDescriptorSet{}
	if err := proto.Unmarshal(bin, fds); err != nil {
		return &parsedProtoDescriptor{err: fmt.Errorf("failed to parse the proto descriptor: %s", protoErrorMessage(err))}
	}
	if err := validateImports(fds); err != nil {
		return &parsedProtoDescriptor{err: err}
	}
	files, err := protodesc.NewFiles(fds)
	if err != nil {
		return &parsedProtoDescriptor{err: fmt.Errorf("invalid proto descriptor: %s", protoErrorMessage(err))}
	}

	imported := sets.New[string]()
	for _, file := range fds.GetFile() {
		imported.Insert(file.GetDependency()...)
	}

	d := &parsedProtoDescriptor{bin: bin, fds: fds, files: files, services: sets.New[string]()}
	for _, file := range fds.GetFile() {
		for _, svc := range file.GetService() {
			name := svc.GetName()
			if file.GetPackage() != "" {
				name = file.GetPackage() + "." + name
			}
			if !imported.Has(file.GetName()) && !d.services.Has(name) {
				d.defaultServices = append(d.defaultServices, name)
			}
			d.services.Insert(name)
		}
	}
	if d.services.Len() == 0 {
		return &parsedProtoDescriptor{err: errors.New("no gRPC service found in the proto descriptor")}
	}
	return d
}

func readProtoDescriptor(cm *corev1.ConfigMap, key types.NamespacedName) ([]byte, error) {
	if cm == nil {
		return nil, fmt.Errorf("can't find the referenced configmap %s", key)
	}

	if v, ok := cm.BinaryData[ProtoDescriptorConfigMapKey]; ok {
		return v, nil
	}
	if v, ok := cm.Data[ProtoDescriptorConfigMapKey]; ok {
		return decodeProtoDescriptor(v)
	}
	// Fallback to the only binaryData entry, which is what `kubectl create configmap --from-file`
	// creates. This is not done for data, because the provider only caches some keys of data.
	if len(cm.Data) == 0 && len(cm.BinaryData) == 1 {
		for _, v := range cm.BinaryData {
			return v, nil
		}
	}

	return nil, fmt.Errorf("can't find the key %s in the referenced configmap %s", ProtoDescriptorConfigMapKey, key)
}

// decodeProtoDescriptor returns the descriptor set in a data value. The value is usually
// base64-encoded, maybe in multiple lines. It is not encoded if `kubectl create configmap
// --from-file` created it, because kubectl stores a file that is valid UTF-8 in data.
func decodeProtoDescriptor(s string) ([]byte, error) {
	stripped := strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, s)

	bin, err := base64.StdEncoding.DecodeString(stripped)
	if err != nil {
		return []byte(s), nil
	}
	if len(bin) == 0 {
		return nil, errors.New("the proto descriptor is empty")
	}
	return bin, nil
}

// validateImports checks that every imported file is in the descriptor set and comes before
// the file that imports it, because Envoy loads the files in order.
func validateImports(fds *descriptorpb.FileDescriptorSet) error {
	all := sets.New[string]()
	for _, f := range fds.GetFile() {
		all.Insert(f.GetName())
	}

	loaded := sets.New[string]()
	for _, f := range fds.GetFile() {
		for _, dep := range f.GetDependency() {
			if loaded.Has(dep) {
				continue
			}
			if all.Has(dep) {
				return fmt.Errorf("%s is listed before its import %s in the proto descriptor", f.GetName(), dep)
			}
			return fmt.Errorf("%s imported by %s is missing in the proto descriptor, "+
				"generate it with protoc --include_imports", dep, f.GetName())
		}
		loaded.Insert(f.GetName())
	}
	return nil
}

// validateHTTPRules checks the google.api.http options of the services, because Envoy
// rejects the whole listener if one of them is invalid. protoc doesn't check them.
func validateHTTPRules(files *protoregistry.Files, services []string) error {
	for _, svc := range services {
		desc, err := files.FindDescriptorByName(protoreflect.FullName(svc))
		if err != nil {
			return err
		}
		methods := desc.(protoreflect.ServiceDescriptor).Methods()
		for i := range methods.Len() {
			method := methods.Get(i)
			opts, ok := method.Options().(*descriptorpb.MethodOptions)
			if !ok || !proto.HasExtension(opts, annotations.E_Http) {
				continue
			}
			rule := proto.GetExtension(opts, annotations.E_Http).(*annotations.HttpRule)

			if err := validatePaths(rule); err != nil {
				return fmt.Errorf("%w of method %s", err, method.FullName())
			}
			// Envoy only checks the body and response_body of the top-level rule.
			if _, err := findField(method.Input(), rule.GetBody()); err != nil {
				return fmt.Errorf("invalid body %q of method %s: %w", rule.GetBody(), method.FullName(), err)
			}
			field, err := findField(method.Output(), rule.GetResponseBody())
			if err != nil {
				return fmt.Errorf("invalid response_body %q of method %s: %w", rule.GetResponseBody(), method.FullName(), err)
			}
			// Envoy only supports google.api.HttpBody for a response_body field.
			if field != nil && (field.Message() == nil || field.Message().FullName() != "google.api.HttpBody") {
				return fmt.Errorf("response_body %q of method %s must be a google.api.HttpBody field",
					rule.GetResponseBody(), method.FullName())
			}
		}
	}
	return nil
}

// validatePaths checks the path templates of the rule and its additional bindings.
func validatePaths(rule *annotations.HttpRule) error {
	var path string
	switch p := rule.GetPattern().(type) {
	case *annotations.HttpRule_Get:
		path = p.Get
	case *annotations.HttpRule_Put:
		path = p.Put
	case *annotations.HttpRule_Post:
		path = p.Post
	case *annotations.HttpRule_Delete:
		path = p.Delete
	case *annotations.HttpRule_Patch:
		path = p.Patch
	case *annotations.HttpRule_Custom:
		path = p.Custom.GetPath()
	}
	if rule.GetPattern() != nil && !isValidPathTemplate(path) {
		return fmt.Errorf("invalid path template %q", path)
	}

	for _, binding := range rule.GetAdditionalBindings() {
		if err := validatePaths(binding); err != nil {
			return err
		}
	}
	return nil
}

// isValidPathTemplate checks the syntax of a path template:
//
//	Template  = "/" | "/" Segments [ ":" Verb ] ;
//	Segments  = Segment { "/" Segment } ;
//	Segment   = "*" | "**" | LITERAL | Variable ;
//	Variable  = "{" FieldPath [ "=" Segments ] "}" ;
//	FieldPath = IDENT { "." IDENT } ;
//	Verb      = LITERAL ;
//
// Like Envoy, only literals can follow "**", and a variable can't contain another variable.
// A literal ends at "/", ":" or "}", and an IDENT ends at ".", "}" or "=".
func isValidPathTemplate(path string) bool {
	if path == "/" {
		return true
	}
	p := &templateParser{s: path}
	if !p.consume('/') || !p.segments(false) {
		return false
	}
	if p.consume(':') && p.literal() == "" {
		return false
	}
	return p.i == len(p.s)
}

type templateParser struct {
	s string
	i int
	// afterWildcard is true after a "**" segment.
	afterWildcard bool
}

func (p *templateParser) consume(c byte) bool {
	if p.i < len(p.s) && p.s[p.i] == c {
		p.i++
		return true
	}
	return false
}

func (p *templateParser) segments(inVariable bool) bool {
	for {
		if !p.segment(inVariable) {
			return false
		}
		if !p.consume('/') {
			return true
		}
	}
}

func (p *templateParser) segment(inVariable bool) bool {
	switch {
	case strings.HasPrefix(p.s[p.i:], "**"):
		p.i += 2
		if p.afterWildcard {
			return false
		}
		p.afterWildcard = true
		return true
	case p.consume('*'):
		return !p.afterWildcard
	case p.consume('{'):
		if inVariable || p.afterWildcard || !p.fieldPath() {
			return false
		}
		if p.consume('=') && !p.segments(true) {
			return false
		}
		return p.consume('}')
	default:
		return p.literal() != ""
	}
}

func (p *templateParser) fieldPath() bool {
	if !p.ident() {
		return false
	}
	for p.consume('.') {
		if !p.ident() {
			return false
		}
	}
	return true
}

func (p *templateParser) literal() string {
	start := p.i
	for p.i < len(p.s) && !strings.ContainsRune("/:}", rune(p.s[p.i])) {
		p.i++
	}
	return p.s[start:p.i]
}

func (p *templateParser) ident() bool {
	start := p.i
	for p.i < len(p.s) && !strings.ContainsRune(".}=", rune(p.s[p.i])) {
		p.i++
	}
	return p.i > start
}

// validateGRPCStatusProtos checks that the descriptor set works with convertGRPCStatus.
// Envoy adds its own google/protobuf/any.proto and google/rpc/status.proto to the descriptor
// set if google.protobuf.Any or google.rpc.Status is not found. This fails if the descriptor
// set already has one of these files, or has google.protobuf.Any in another file, because
// status.proto imports google/protobuf/any.proto.
func validateGRPCStatusProtos(d *parsedProtoDescriptor) error {
	fds := &descriptorpb.FileDescriptorSet{File: slices.Clone(d.fds.GetFile())}
	names := sets.New[string]()
	for _, f := range fds.File {
		names.Insert(f.GetName())
	}

	for _, file := range []protoreflect.FileDescriptor{
		anypb.File_google_protobuf_any_proto,
		spb.File_google_rpc_status_proto,
	} {
		if _, err := d.files.FindDescriptorByName(file.Messages().Get(0).FullName()); err == nil {
			continue
		}
		for i := range file.Imports().Len() {
			if dep := file.Imports().Get(i).Path(); !names.Has(dep) {
				return fmt.Errorf("convertGRPCStatus: Envoy adds %s, but %s imported by it is missing in the proto descriptor",
					file.Path(), dep)
			}
		}
		fds.File = append(fds.File, protodesc.ToFileDescriptorProto(file))
		names.Insert(file.Path())
	}

	if _, err := protodesc.NewFiles(fds); err != nil {
		return fmt.Errorf("convertGRPCStatus: Envoy adds google/protobuf/any.proto and google/rpc/status.proto "+
			"to the proto descriptor, and they conflict with it: %s", protoErrorMessage(err))
	}
	return nil
}

// findField finds the field of a field path like "a.b.c" in the message.
// It returns nil for "" and "*", which mean the whole message.
// Like Envoy, a field is matched by its JSON name first, and then by its proto name.
func findField(msg protoreflect.MessageDescriptor, path string) (protoreflect.FieldDescriptor, error) {
	if path == "" || path == "*" {
		return nil, nil
	}

	var field protoreflect.FieldDescriptor
	for name := range strings.SplitSeq(path, ".") {
		if field != nil {
			if field.Message() == nil {
				return nil, fmt.Errorf("%s is not a message", field.Name())
			}
			msg = field.Message()
		}
		field = msg.Fields().ByJSONName(name)
		if field == nil {
			field = msg.Fields().ByName(protoreflect.Name(name))
		}
		if field == nil {
			return nil, fmt.Errorf("can't find field %s in %s", name, msg.FullName())
		}
	}
	return field, nil
}

// getTranscodedServices returns the services to transcode.
// If no service is specified, the default services of the descriptor set are used,
// because Envoy doesn't transcode anything with an empty service list.
func getTranscodedServices(d *parsedProtoDescriptor, services []string) ([]string, error) {
	if len(services) == 0 {
		if len(d.defaultServices) == 0 {
			return nil, errors.New("all the services in the proto descriptor are from imported files, " +
				"please specify the services to transcode")
		}
		return d.defaultServices, nil
	}

	// Duplicated services are removed, because Envoy doesn't transcode anything if a service
	// is listed twice. The API server already rejects them, but the file provider doesn't.
	result := make([]string, 0, len(services))
	seen := sets.New[string]()
	for _, svc := range services {
		if !d.services.Has(svc) {
			return nil, fmt.Errorf("can't find service %s in the proto descriptor, available services: %s",
				svc, listServices(d.services))
		}
		if !seen.Has(svc) {
			seen.Insert(svc)
			result = append(result, svc)
		}
	}
	return result, nil
}

// listServices lists the services for an error message.
// Only the first few services are listed to keep the status message short.
func listServices(services sets.Set[string]) string {
	list := sets.List(services)
	if len(list) <= maxServicesInError {
		return strings.Join(list, ", ")
	}
	return fmt.Sprintf("%s and %d more", strings.Join(list[:maxServicesInError], ", "), len(list)-maxServicesInError)
}

// protoErrorMessage returns the message of a protobuf error without the "proto:" prefix.
// protobuf-go randomly adds a normal or a non-breaking space after the prefix, which would
// change the route status every time Envoy Gateway restarts.
func protoErrorMessage(err error) string {
	msg := err.Error()
	for _, prefix := range []string{"proto: ", "proto: "} {
		if s, ok := strings.CutPrefix(msg, prefix); ok {
			return s
		}
	}
	return msg
}
