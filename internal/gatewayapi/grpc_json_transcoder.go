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

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/gatewayapi/status"
	"github.com/envoyproxy/gateway/internal/ir"
)

// ProtoDescriptorConfigMapKey is the preferred ConfigMap key holding the FileDescriptorSet.
// It must stay in the provider's cachedConfigMapKeys, or the informer transform drops it
// from any ConfigMap carrying more than one data entry.
const ProtoDescriptorConfigMapKey = "proto-descriptor"

// buildGRPCJSONTranscoder resolves the descriptor referenced by an HTTPRouteFilter into
// the IR. name identifies the owning HTTPRouteFilter and becomes the HCM filter instance
// name; namespace is the route's, since extensionRef never crosses namespaces.
func (t *Translator) buildGRPCJSONTranscoder(
	cfg *egv1a1.GRPCJSONTranscoder, name, namespace string,
) (*ir.GRPCJSONTranscoder, error) {
	descriptor, err := t.loadProtoDescriptor(cfg.ProtoDescriptor, namespace)
	if err != nil {
		return nil, err
	}

	services, err := resolveTranscodedServices(descriptor, cfg.Services)
	if err != nil {
		return nil, err
	}
	if err := validateHTTPBindings(descriptor.files, services); err != nil {
		return nil, err
	}
	if ptr.Deref(cfg.ConvertGRPCStatus, false) {
		if err := validateGRPCStatusBuiltins(descriptor); err != nil {
			return nil, err
		}
	}

	return &ir.GRPCJSONTranscoder{
		Name:                         name,
		ProtoDescriptorBin:           descriptor.bin,
		Services:                     services,
		PrintOptions:                 cfg.PrintOptions,
		IgnoredQueryParameters:       cfg.IgnoredQueryParameters,
		AutoMapping:                  cfg.AutoMapping,
		IgnoreUnknownQueryParameters: cfg.IgnoreUnknownQueryParameters,
		ConvertGRPCStatus:            cfg.ConvertGRPCStatus,
	}, nil
}

// parsedProtoDescriptor is a descriptor that has been decoded, validated, and had its
// service names extracted.
type parsedProtoDescriptor struct {
	// err records a descriptor that failed to load, so a broken ConfigMap is not re-read
	// and re-unmarshalled once per referencing rule on every translation.
	err   error
	bin   []byte
	fds   *descriptorpb.FileDescriptorSet
	files *protoregistry.Files
	// all is every service declared in the set; roots omits those declared by files that
	// another file imports.
	all   sets.Set[string]
	roots []string
}

// loadProtoDescriptor returns the descriptor referenced by protoDesc, decoding and
// validating it once per translation.
func (t *Translator) loadProtoDescriptor(
	protoDesc egv1a1.ProtoDescriptor, namespace string,
) (*parsedProtoDescriptor, error) {
	ref := protoDesc.ValueRef
	if g, k := string(ref.Group), string(ref.Kind); g != "" || k != resource.KindConfigMap {
		return nil, fmt.Errorf("unsupported valueRef %s/%s, only ConfigMap is supported", g, k)
	}

	key := types.NamespacedName{Namespace: namespace, Name: string(ref.Name)}
	if d, ok := t.protoDescriptors[key]; ok {
		if d.err != nil {
			return nil, d.err
		}
		return d, nil
	}

	d, err := parseProtoDescriptor(t.GetConfigMap(key.Namespace, key.Name), key)
	if t.protoDescriptors == nil {
		t.protoDescriptors = map[types.NamespacedName]*parsedProtoDescriptor{}
	}
	if err != nil {
		t.protoDescriptors[key] = &parsedProtoDescriptor{err: err}
		return nil, err
	}
	t.protoDescriptors[key] = d
	return d, nil
}

func parseProtoDescriptor(cm *corev1.ConfigMap, key types.NamespacedName) (*parsedProtoDescriptor, error) {
	bin, err := readProtoDescriptor(cm, key)
	if err != nil {
		return nil, err
	}

	fds := &descriptorpb.FileDescriptorSet{}
	if err := proto.Unmarshal(bin, fds); err != nil {
		return nil, fmt.Errorf("failed to parse proto descriptor as a FileDescriptorSet: %s",
			trimProtoPrefix(err))
	}
	fds.File = dropRepeatedFiles(fds.File)
	if err := validateDescriptorClosure(fds); err != nil {
		return nil, err
	}
	if err := validateDescriptorStructure(fds); err != nil {
		return nil, err
	}
	files, err := validateDescriptorPool(fds)
	if err != nil {
		return nil, err
	}

	// Imported files can declare services of their own (google.longrunning.Operations, for
	// one). Only files nothing else imports were compiled by the user.
	imported := sets.New[string]()
	for _, file := range fds.GetFile() {
		imported.Insert(file.GetDependency()...)
	}

	d := &parsedProtoDescriptor{bin: bin, fds: fds, files: files, all: sets.New[string]()}
	for _, file := range fds.GetFile() {
		for _, svc := range file.GetService() {
			name := svc.GetName()
			if pkg := file.GetPackage(); pkg != "" {
				name = pkg + "." + name
			}
			if !imported.Has(file.GetName()) && !d.all.Has(name) {
				d.roots = append(d.roots, name)
			}
			d.all.Insert(name)
		}
	}
	if d.all.Len() == 0 {
		return nil, errors.New("proto descriptor contains no gRPC services")
	}
	return d, nil
}

func readProtoDescriptor(cm *corev1.ConfigMap, key types.NamespacedName) ([]byte, error) {
	if cm == nil {
		return nil, fmt.Errorf("proto descriptor ConfigMap %s not found", key)
	}

	// `kubectl create configmap --from-file` puts a descriptor in BinaryData, already decoded.
	if v, ok := cm.BinaryData[ProtoDescriptorConfigMapKey]; ok {
		return v, nil
	}
	if v, ok := cm.Data[ProtoDescriptorConfigMapKey]; ok {
		return decodeProtoDescriptor(v)
	}
	// Only BinaryData can be matched by being the sole entry: the informer cache trims
	// Data to its first key, so "exactly one entry" is not the same fact in-cluster as it
	// is offline, and the two providers would disagree. BinaryData is never trimmed.
	if len(cm.Data) == 0 && len(cm.BinaryData) == 1 {
		for _, v := range cm.BinaryData {
			return v, nil
		}
	}

	return nil, fmt.Errorf(
		"proto descriptor not found in ConfigMap %s: expected key %q, or a sole binaryData entry",
		key, ProtoDescriptorConfigMapKey)
}

// decodeProtoDescriptor strips whitespace before decoding because YAML block scalars
// fold in newlines.
func decodeProtoDescriptor(s string) ([]byte, error) {
	stripped := strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, s)

	bin, err := base64.StdEncoding.DecodeString(stripped)
	if err != nil {
		return nil, fmt.Errorf("proto descriptor is not valid base64: %w", err)
	}
	if len(bin) == 0 {
		return nil, errors.New("proto descriptor is empty")
	}
	return bin, nil
}

// validateDescriptorClosure ensures every imported file is present in the set. Envoy
// reports only "Unable to build proto descriptor pool" when one is missing.
func validateDescriptorClosure(fds *descriptorpb.FileDescriptorSet) error {
	present := sets.New[string]()
	for _, f := range fds.GetFile() {
		present.Insert(f.GetName())
	}

	missing := sets.New[string]()
	for _, f := range fds.GetFile() {
		for _, dep := range f.GetDependency() {
			if !present.Has(dep) {
				missing.Insert(fmt.Sprintf("%s (imported by %s)", dep, f.GetName()))
			}
		}
	}
	if missing.Len() > 0 {
		return fmt.Errorf(
			"proto descriptor is missing imported files: %s; regenerate it with "+
				"`protoc --include_imports --descriptor_set_out=...`",
			strings.Join(sets.List(missing), ", "))
	}
	return nil
}

// validateDescriptorStructure checks what protodesc.NewFiles does not but Envoy's pool
// does. Envoy builds the files one at a time in set order with no fallback, so each import
// must precede the file importing it, dependency indices must be in range, and options must
// already be interpreted. protoc output meets all three; a concatenation of sets may not.
func validateDescriptorStructure(fds *descriptorpb.FileDescriptorSet) error {
	built := sets.New[string]()
	for _, f := range fds.GetFile() {
		for _, dep := range f.GetDependency() {
			if !built.Has(dep) {
				return fmt.Errorf("proto descriptor lists %s before its import %s, and Envoy loads files "+
					"in order; generate it with `protoc --include_imports` instead of concatenating sets",
					f.GetName(), dep)
			}
		}
		for _, i := range slices.Concat(f.GetPublicDependency(), f.GetWeakDependency()) {
			if i < 0 || int(i) >= len(f.GetDependency()) {
				return fmt.Errorf("proto descriptor file %s has dependency index %d out of range", f.GetName(), i)
			}
		}
		if hasUninterpretedOption(f.ProtoReflect()) {
			return fmt.Errorf("proto descriptor file %s has uninterpreted options, which protoc never emits", f.GetName())
		}
		if err := validateJSONNames(f); err != nil {
			return fmt.Errorf("proto descriptor file %s: %w", f.GetName(), err)
		}
		built.Insert(f.GetName())
	}
	return nil
}

// dropRepeatedFiles removes later copies of a file that are identical to an earlier one, as
// concatenating two sets that share an import produces. Envoy's pool accepts such a copy
// only when it serializes the same as the loaded file, which drops source_code_info, so a
// copy carrying that is kept and rejected by linking.
func dropRepeatedFiles(files []*descriptorpb.FileDescriptorProto) []*descriptorpb.FileDescriptorProto {
	first := map[string]*descriptorpb.FileDescriptorProto{}
	out := files[:0:0]
	for _, f := range files {
		if prev, ok := first[f.GetName()]; ok && f.SourceCodeInfo == nil && proto.Equal(prev, f) {
			continue
		}
		if _, ok := first[f.GetName()]; !ok {
			first[f.GetName()] = f
		}
		out = append(out, f)
	}
	return out
}

// validateJSONNames mirrors protobuf's CheckFieldJsonNameUniqueness and
// CheckEnumValueUniqueness, which Envoy's pool enforces and protodesc.NewFiles does not.
// Field JSON names must be unique both with and without custom json_name values; in proto2
// a clash involving a default name is only a warning there. Enum values must stay distinct
// once the enum-name prefix is stripped and they are PascalCased. Editions files are
// treated as proto3, which is stricter than protobuf when they opt into LEGACY_BEST_EFFORT.
func validateJSONNames(f *descriptorpb.FileDescriptorProto) error {
	proto2 := f.GetSyntax() == "" || f.GetSyntax() == "proto2"

	checkEnum := func(scope string, e *descriptorpb.EnumDescriptorProto) error {
		if proto2 && e.GetOptions().GetDeprecatedLegacyJsonFieldConflicts() {
			return nil
		}
		prefix := strings.ToLower(strings.ReplaceAll(e.GetName(), "_", ""))
		seen := map[string]*descriptorpb.EnumValueDescriptorProto{}
		for _, v := range e.GetValue() {
			key := enumValueToPascalCase(removeEnumPrefix(prefix, v.GetName()))
			prev, ok := seen[key]
			if !ok {
				seen[key] = v
				continue
			}
			if prev.GetName() != v.GetName() && prev.GetNumber() != v.GetNumber() {
				return fmt.Errorf("enum %s: values %s and %s collide once the enum-name prefix is "+
					"stripped", qualify(scope, e.GetName()), prev.GetName(), v.GetName())
			}
		}
		return nil
	}

	var checkMessage func(scope string, m *descriptorpb.DescriptorProto) error
	checkMessage = func(scope string, m *descriptorpb.DescriptorProto) error {
		name := qualify(scope, m.GetName())
		if !m.GetOptions().GetDeprecatedLegacyJsonFieldConflicts() {
			for _, useCustom := range []bool{false, true} {
				type seenField struct {
					field  *descriptorpb.FieldDescriptorProto
					custom bool
				}
				seen := map[string]seenField{}
				for _, fd := range m.GetField() {
					jsonName, custom := jsonCamelCase(fd.GetName()), false
					if useCustom && fd.JsonName != nil && fd.GetJsonName() != jsonName {
						jsonName, custom = fd.GetJsonName(), true
					}
					if custom && strings.HasPrefix(jsonName, "[") && strings.HasSuffix(jsonName, "]") {
						return fmt.Errorf("message %s: custom JSON name %q of field %s may not start with "+
							"'[' and end with ']'", name, jsonName, fd.GetName())
					}
					prev, ok := seen[jsonName]
					if !ok {
						seen[jsonName] = seenField{fd, custom}
						continue
					}
					if (useCustom && !custom && !prev.custom) || (proto2 && (!custom || !prev.custom)) {
						continue
					}
					return fmt.Errorf("message %s: fields %s and %s have the same JSON name %q",
						name, prev.field.GetName(), fd.GetName(), jsonName)
				}
			}
		}
		for _, e := range m.GetEnumType() {
			if err := checkEnum(name, e); err != nil {
				return err
			}
		}
		for _, nested := range m.GetNestedType() {
			if err := checkMessage(name, nested); err != nil {
				return err
			}
		}
		return nil
	}

	for _, e := range f.GetEnumType() {
		if err := checkEnum(f.GetPackage(), e); err != nil {
			return err
		}
	}
	for _, m := range f.GetMessageType() {
		if err := checkMessage(f.GetPackage(), m); err != nil {
			return err
		}
	}
	return nil
}

func qualify(scope, name string) string {
	if scope == "" {
		return name
	}
	return scope + "." + name
}

// jsonCamelCase is protobuf's ToJsonName: drop each '_' and upper-case the byte after it.
func jsonCamelCase(s string) string {
	var b strings.Builder
	upper := false
	for i := 0; i < len(s); i++ {
		switch c := s[i]; {
		case c == '_':
			upper = true
		case upper:
			b.WriteByte(asciiUpper(c))
			upper = false
		default:
			b.WriteByte(c)
		}
	}
	return b.String()
}

// enumValueToPascalCase is protobuf's EnumValueToPascalCase.
func enumValueToPascalCase(s string) string {
	var b strings.Builder
	upper := true
	for i := 0; i < len(s); i++ {
		switch c := s[i]; {
		case c == '_':
			upper = true
		case upper:
			b.WriteByte(asciiUpper(c))
			upper = false
		default:
			b.WriteByte(asciiLower(c))
		}
	}
	return b.String()
}

// removeEnumPrefix is protobuf's PrefixRemover::MaybeRemove; prefix is the enum name
// lower-cased with underscores removed.
func removeEnumPrefix(prefix, s string) string {
	i, j := 0, 0
	for ; i < len(s) && j < len(prefix); i++ {
		if s[i] == '_' {
			continue
		}
		if asciiLower(s[i]) != prefix[j] {
			return s
		}
		j++
	}
	if j < len(prefix) {
		return s
	}
	for i < len(s) && s[i] == '_' {
		i++
	}
	if i == len(s) {
		return s
	}
	return s[i:]
}

func asciiUpper(c byte) byte {
	if 'a' <= c && c <= 'z' {
		return c - 'a' + 'A'
	}
	return c
}

func asciiLower(c byte) byte {
	if 'A' <= c && c <= 'Z' {
		return c - 'A' + 'a'
	}
	return c
}

// validateGRPCStatusBuiltins mirrors what Envoy does for convert_grpc_status once the set is
// loaded: for google.protobuf.Any and then google.rpc.Status, if the pool lacks the symbol it
// adds the defining file from its own compiled-in protos, into the same pool. That load fails
// when the set declares Any in a file with another path, since status.proto imports it by
// its canonical one, or when the set already has a file at one of those paths.
func validateGRPCStatusBuiltins(d *parsedProtoDescriptor) error {
	fds := &descriptorpb.FileDescriptorSet{File: slices.Clone(d.fds.GetFile())}
	paths := sets.New[string]()
	for _, f := range fds.File {
		paths.Insert(f.GetName())
	}
	for _, builtin := range []protoreflect.FileDescriptor{
		anypb.File_google_protobuf_any_proto,
		spb.File_google_rpc_status_proto,
	} {
		if _, err := d.files.FindDescriptorByName(builtin.Messages().Get(0).FullName()); err == nil {
			continue
		}
		for i := range builtin.Imports().Len() {
			if dep := builtin.Imports().Get(i).Path(); !paths.Has(dep) {
				return fmt.Errorf("convertGRPCStatus: Envoy adds its own %s, which imports %s, and the "+
					"proto descriptor has no file at that path", builtin.Path(), dep)
			}
		}
		fds.File = append(fds.File, protodesc.ToFileDescriptorProto(builtin))
		paths.Insert(builtin.Path())
	}
	if _, err := protodesc.NewFiles(fds); err != nil {
		return fmt.Errorf("convertGRPCStatus: Envoy adds its own google/protobuf/any.proto and "+
			"google/rpc/status.proto where missing, and they conflict with the proto descriptor: %s",
			trimProtoPrefix(err))
	}
	return nil
}

// hasUninterpretedOption reports whether m or any message nested in it sets
// uninterpreted_option.
func hasUninterpretedOption(m protoreflect.Message) bool {
	found := false
	m.Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		switch {
		case fd.Name() == "uninterpreted_option":
			found = true
		case fd.Message() == nil || fd.IsMap():
		case fd.IsList():
			for i := 0; i < v.List().Len() && !found; i++ {
				found = hasUninterpretedOption(v.List().Get(i).Message())
			}
		default:
			found = hasUninterpretedOption(v.Message())
		}
		return !found
	})
	return found
}

// validateDescriptorPool links the descriptor graph, which unmarshalling does not: a set
// whose method references an undeclared message decodes cleanly, then loses the whole
// listener when Envoy fails to build its own pool. Linking here makes it this route's
// problem instead.
func validateDescriptorPool(fds *descriptorpb.FileDescriptorSet) (*protoregistry.Files, error) {
	files, err := protodesc.NewFiles(fds)
	if err != nil {
		return nil, fmt.Errorf("failed to build a proto descriptor pool: %s", trimProtoPrefix(err))
	}
	return files, nil
}

// validateHTTPBindings rejects the google.api.http bindings that fail Envoy's
// JsonTranscoderConfig constructor, which would reject the listener. protoc checks none
// of them. Only the top-level rule's body and response_body are resolved there; every
// binding's path template is parsed.
func validateHTTPBindings(files *protoregistry.Files, services []string) error {
	for _, svc := range services {
		d, err := files.FindDescriptorByName(protoreflect.FullName(svc))
		if err != nil {
			return err
		}
		methods := d.(protoreflect.ServiceDescriptor).Methods()
		for i := range methods.Len() {
			m := methods.Get(i)
			opts, ok := m.Options().(*descriptorpb.MethodOptions)
			if !ok || !proto.HasExtension(opts, annotations.E_Http) {
				continue
			}
			rule := proto.GetExtension(opts, annotations.E_Http).(*annotations.HttpRule)

			if _, err := resolveFieldPath(m.Input(), rule.GetBody()); err != nil {
				return fmt.Errorf("method %s: body %q: %w", m.FullName(), rule.GetBody(), err)
			}
			field, err := resolveFieldPath(m.Output(), rule.GetResponseBody())
			if err != nil {
				return fmt.Errorf("method %s: response_body %q: %w", m.FullName(), rule.GetResponseBody(), err)
			}
			if field != nil && (field.Message() == nil || field.Message().FullName() != "google.api.HttpBody") {
				return fmt.Errorf("method %s: response_body %q must be a google.api.HttpBody field, "+
					"Envoy does not support other types", m.FullName(), rule.GetResponseBody())
			}
			if err := validateHTTPPatterns(rule); err != nil {
				return fmt.Errorf("method %s: %w", m.FullName(), err)
			}
		}
	}
	return nil
}

// validateHTTPPatterns parses the path template of rule and of every additional binding,
// as PathMatcherUtility::RegisterByHttpRule registers them. A rule with no pattern
// registers nothing.
func validateHTTPPatterns(rule *annotations.HttpRule) error {
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
	if rule.GetPattern() != nil && !validHTTPTemplate(path) {
		return fmt.Errorf("invalid path template %q", path)
	}
	for _, b := range rule.GetAdditionalBindings() {
		if err := validateHTTPPatterns(b); err != nil {
			return err
		}
	}
	return nil
}

// validHTTPTemplate accepts a subset of grpc-httpjson-transcoding's HttpTemplate::Parse, so
// anything it accepts Envoy accepts too:
//
//	Template = "/" | "/" Segments [ ":" Literal ] ;
//	Segments = Segment { "/" Segment } ;
//	Segment  = "*" | "**" | Literal | "{" Ident { "." Ident } [ "=" Segments ] "}" ;
//
// It is stricter where Envoy is loose: literals exclude "/:{}*", identifiers are protobuf
// field names, and nothing but literals may follow "**" (Envoy's ValidateParts), which also
// rules out a variable there since "{x}" means "{x=*}".
func validHTTPTemplate(t string) bool {
	if t == "/" {
		return true
	}
	p := &templateParser{s: t}
	return p.consume('/') && p.segments(false) && (!p.consume(':') || p.literal()) && p.i == len(p.s)
}

type templateParser struct {
	s        string
	i        int
	wildcard bool // a "**" segment has been parsed
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
		if p.wildcard {
			return false
		}
		p.wildcard = true
		return true
	case p.consume('*'):
		return !p.wildcard
	case p.consume('{'):
		if inVariable || p.wildcard || !p.ident() {
			return false
		}
		for p.consume('.') {
			if !p.ident() {
				return false
			}
		}
		if p.consume('=') && !p.segments(true) {
			return false
		}
		return p.consume('}')
	default:
		return p.literal()
	}
}

func (p *templateParser) literal() bool {
	start := p.i
	for p.i < len(p.s) && !strings.ContainsRune("/:{}*", rune(p.s[p.i])) {
		p.i++
	}
	return p.i > start
}

func (p *templateParser) ident() bool {
	start := p.i
	for ; p.i < len(p.s); p.i++ {
		c := p.s[p.i]
		letter := c == '_' || ('a' <= c && c <= 'z') || ('A' <= c && c <= 'Z')
		if !letter && (p.i == start || c < '0' || c > '9') {
			break
		}
	}
	return p.i > start
}

// resolveFieldPath mirrors Envoy's TypeHelper::ResolveFieldPath: "*" or "" selects the whole
// message (nil field), a segment matches a field's JSON name and then its proto name (the
// order of proto-converter's TypeInfo::FindField), and every segment but the last must be a
// message.
func resolveFieldPath(msg protoreflect.MessageDescriptor, path string) (protoreflect.FieldDescriptor, error) {
	if path == "*" {
		return nil, nil
	}
	var field protoreflect.FieldDescriptor
	for seg := range strings.SplitSeq(path, ".") {
		if seg == "" {
			continue
		}
		if field != nil {
			if field.Kind() != protoreflect.MessageKind {
				return nil, fmt.Errorf("%s is not a message", field.Name())
			}
			msg = field.Message()
		}
		if field = msg.Fields().ByJSONName(seg); field == nil {
			field = msg.Fields().ByName(protoreflect.Name(seg))
		}
		if field == nil {
			return nil, fmt.Errorf("no field %q in %s", seg, msg.FullName())
		}
	}
	return field, nil
}

// resolveTranscodedServices returns the services to transcode. Envoy treats an empty list
// as "filter disabled", so an omitted list is expanded rather than passed through.
func resolveTranscodedServices(d *parsedProtoDescriptor, want []string) ([]string, error) {
	if len(want) == 0 {
		if len(d.roots) == 0 {
			return nil, errors.New(
				"every gRPC service in the proto descriptor comes from an imported file; " +
					"set services explicitly to choose which ones to transcode")
		}
		return d.roots, nil
	}

	// Duplicates are dropped rather than rejected: Envoy registers every method of each
	// listed service into one path matcher and a repeated service leaves the filter unable
	// to transcode, with nothing in its logs and the route still Accepted. `+listType=set`
	// catches this at admission, but the file provider has no API server to enforce it.
	seen := sets.New[string]()
	out := make([]string, 0, len(want))
	for _, svc := range want {
		if !d.all.Has(svc) {
			return nil, fmt.Errorf("service %q not found in the proto descriptor, available services: %s",
				svc, availableServices(d.all))
		}
		if seen.Has(svc) {
			continue
		}
		seen.Insert(svc)
		out = append(out, svc)
	}
	return out, nil
}

// availableServices renders a descriptor's services for an error message. The list is bounded
// because the message ends up in a status condition, which the API server caps at 32Ki.
func availableServices(all sets.Set[string]) string {
	const shown = 10

	list := sets.List(all)
	if len(list) <= shown {
		return strings.Join(list, ", ")
	}
	return fmt.Sprintf("%s, and %d more", strings.Join(list[:shown], ", "), len(list)-shown)
}

// applyGRPCJSONTranscoder resolves an HTTPRouteFilter's transcoder config onto the filter
// context, rejecting the positions where the transcoder cannot work.
func (t *Translator) applyGRPCJSONTranscoder(
	hrf *egv1a1.HTTPRouteFilter, filterContext *HTTPFiltersContext,
) status.Error {
	kind := string(egv1a1.KindHTTPRouteFilter)

	if filterContext.Route.GetRouteType() == resource.KindGRPCRoute {
		return t.processInvalidHTTPFilter(kind, filterContext,
			errors.New("grpcJSONTranscoder is not supported on a GRPCRoute, attach it to the HTTPRoute carrying the JSON traffic"))
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

// trimProtoPrefix drops protobuf-go's "proto:" prefix, whose trailing space is U+0020 or
// U+00A0 depending on a hash of the running binary -- left in, it rewrites the status
// condition on an upgrade for no semantic reason. protobuf-go strips the prefix when it
// chains errors, so it appears at most once, at the front.
func trimProtoPrefix(err error) string {
	s := err.Error()
	for _, p := range [...]string{"proto: ", "proto:\u00a0"} {
		if t, ok := strings.CutPrefix(s, p); ok {
			return t
		}
	}
	return s
}
