// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"unicode"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/types/descriptorpb"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"

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

	return &ir.GRPCJSONTranscoder{
		Name:                         name,
		ProtoDescriptorBin:           descriptor.bin,
		Services:                     services,
		PrintOptions:                 cfg.PrintOptions,
		MatchIncomingRequestRoute:    cfg.MatchIncomingRequestRoute,
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
	err error
	bin []byte
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
	if err := validateDescriptorClosure(fds); err != nil {
		return nil, err
	}
	if err := validateDescriptorPool(fds); err != nil {
		return nil, err
	}

	// Imported files can declare services of their own (google.longrunning.Operations, for
	// one). Only files nothing else imports were compiled by the user.
	imported := sets.New[string]()
	for _, file := range fds.GetFile() {
		imported.Insert(file.GetDependency()...)
	}

	d := &parsedProtoDescriptor{bin: bin, all: sets.New[string]()}
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

// validateDescriptorPool links the descriptor graph, which unmarshalling does not: a set
// whose method references an undeclared message decodes cleanly, then loses the whole
// listener when Envoy fails to build its own pool. Linking here makes it this route's
// problem instead.
func validateDescriptorPool(fds *descriptorpb.FileDescriptorSet) error {
	if _, err := protodesc.NewFiles(fds); err != nil {
		return fmt.Errorf("failed to build a proto descriptor pool: %s", trimProtoPrefix(err))
	}
	return nil
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
