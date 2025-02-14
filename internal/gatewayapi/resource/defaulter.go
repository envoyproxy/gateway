// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package resource

import (
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"sort"
	"strings"

	gospec "github.com/go-openapi/spec"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/validate"
	"github.com/go-openapi/validate/post"
	"golang.org/x/exp/maps"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/openapi"
	"k8s.io/kube-openapi/pkg/spec3"
	kubespec "k8s.io/kube-openapi/pkg/validation/spec"
	"sigs.k8s.io/kubectl-validate/pkg/openapiclient"
	"sigs.k8s.io/kubectl-validate/pkg/utils"
	"sigs.k8s.io/kubectl-validate/pkg/validator"
)

// This file contains code derived from kubectl-validate,
// https://github.com/kubernetes-sigs/kubectl-validate
// from the source file
// https://github.com/kubernetes-sigs/kubectl-validate/blob/main/pkg/validator/validator.go
// and is provided here subject to the following:
// Copyright Project kubectl-validate Authors
// SPDX-License-Identifier: Apache-2.0
//
// The Defaulter in this file is derived from Validator in kubectl-validate,
// since the Validator field `validatorCache` is not exposed and we would like
// to use the parsed schema for our CRD from it, we build this Defaulter that
// meets our needs.
// TODO: remove this file once can directly get schema from the Validator in kubectl-validate.

var gatewaySchemaDefaulter, _ = newDefaulter(openapiclient.NewLocalCRDFiles(gatewayCRDsFS))

// Defaulter can set default values for crd object according to their schema.
type Defaulter struct {
	gvs         map[string]openapi.GroupVersion
	schemaCache map[schema.GroupVersionKind]*kubespec.Schema
}

func newDefaulter(client openapi.Client) (*Defaulter, error) {
	gvs, err := client.Paths()
	if err != nil {
		return nil, err
	}

	return &Defaulter{
		gvs:         gvs,
		schemaCache: map[schema.GroupVersionKind]*kubespec.Schema{},
	}, nil
}

// ApplyDefault applies default values for input object, and return the object with default values.
func (d *Defaulter) ApplyDefault(obj *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	if obj == nil || obj.Object == nil {
		return nil, fmt.Errorf("passed object cannot be nil")
	}

	// shallow copy input object, this method can modify apiVersion, kind, or metadata.
	obj = &unstructured.Unstructured{Object: maps.Clone(obj.UnstructuredContent())}
	// deep copy metadata object.
	obj.Object["metadata"] = runtime.DeepCopyJSONValue(obj.Object["metadata"])
	gvk := obj.GroupVersionKind()
	schema, err := d.parseSchema(gvk)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve validator: %w", err)
	}

	// convert kube-openapi spec to go-openapi spec via JSON format.
	schemaBytes, err := schema.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal schema: %w", err)
	}
	var goSchema gospec.Schema
	err = goSchema.UnmarshalJSON(schemaBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal schema: %w", err)
	}

	v := validate.NewSchemaValidator(&goSchema, nil, "", strfmt.Default)
	rs := v.Validate(obj.Object)
	post.ApplyDefaults(rs)
	// convert output object into unstructured one.
	output, ok := rs.Data().(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("failed to convert output object")
	}

	return &unstructured.Unstructured{Object: output}, nil
}

func (d *Defaulter) parseSchema(gvk schema.GroupVersionKind) (*kubespec.Schema, error) {
	if existing, ok := d.schemaCache[gvk]; ok {
		return existing, nil
	}

	// Otherwise, fetch the open API schema for this GV and do the above
	// Lookup gvk in client
	// Guess the rest mapping since we don't have a rest mapper for the target
	// cluster
	gvPath := "apis/" + gvk.Group + "/" + gvk.Version
	if len(gvk.Group) == 0 {
		gvPath = "api/" + gvk.Version
	}
	gvFetcher, exists := d.gvs[gvPath]
	if !exists {
		return nil, fmt.Errorf("failed to locate OpenAPI spec for GV: %v", gvk.GroupVersion())
	}

	documentBytes, err := gvFetcher.Schema("application/json")
	if err != nil {
		return nil, fmt.Errorf("error fetching openapi at path %s: %w", gvPath, err)
	}

	openapiSpec := spec3.OpenAPI{}
	if err := json.Unmarshal(documentBytes, &openapiSpec); err != nil {
		return nil, fmt.Errorf("error parsing openapi spec: %w", err)
	}

	// Apply our transformations to workaround known k8s schema deficiencies
	for name, def := range openapiSpec.Components.Schemas {
		//!TODO: would be useful to know which version of k8s each schema is believed
		// to come from.
		openapiSpec.Components.Schemas[name] = validator.ApplySchemaPatches(0, gvk.GroupVersion(), name, def)
	}

	// Remove all references/indirection.
	// This is kinda hacky because we still do allow recursive schemas via
	// pointer trickery.
	// No need for stack/queue approach since we mutate same dictionary/slice instances
	// destructively.
	// Replaces subschemas that contain refs with copy of the thing they refer to
	// !TODO validate that no cyces are created by this process. If so, do not
	// allow structural schema creation via JSON
	// !TODO: track unresolved references?
	// !TODO: Once Declarative Validation for native types lands we will be
	//	able to validate against the spec.Schema directly rather than
	//	StructuralSchema, so this will be able to be removed
	var referenceErrors []error
	for name, def := range openapiSpec.Components.Schemas {
		// This hack only works because top level schemas never have references
		// so we can reliably copy them knowing they won't change and pointer-share
		// their subfields. The only schemas being modified here should be sub-fields.
		openapiSpec.Components.Schemas[name] = utils.VisitSchema(name, def, utils.PreorderVisitor(func(ctx utils.VisitingContext, sch *kubespec.Schema) (*kubespec.Schema, bool) {
			defName := sch.Ref.String()

			if len(sch.AllOf) == 1 && len(sch.AllOf[0].Ref.String()) > 0 {
				// SPECIAL CASE
				// OpenAPIV3 does not support having Refs in schemas with fields like
				// Description, Default filled in. So k8s stuffs the Ref into a standalone
				// AllOf in these cases.
				// But structural schema doesn't like schemas that specify fields inside AllOf
				// SO in the case of
				// Properties
				//	-> AllOf
				//		-> Ref
				defName = sch.AllOf[0].Ref.String()
			}

			if len(defName) == 0 {
				// Nothing to do for no references
				return sch, true
			}

			defName = path.Base(defName)
			resolved, ok := openapiSpec.Components.Schemas[defName]
			if !ok {
				// Can't resolve schema. This is an error.
				var path []string
				for cursor := &ctx; cursor != nil; cursor = cursor.Parent {
					if len(cursor.Key) == 0 {
						path = append(path, fmt.Sprint(cursor.Index))
					} else {
						path = append(path, cursor.Key)
					}
				}
				sort.Stable(sort.Reverse(sort.StringSlice(path)))
				referenceErrors = append(referenceErrors, fmt.Errorf("cannot resolve reference %v in %v.%v", defName, name, strings.Join(path, ".")))
				return sch, true
			}

			resolvedCopy := *resolved

			if sch.Default != nil {
				resolvedCopy.Default = sch.Default
			}

			// NOTE: No way to tell if field overrides nullable
			// or if it is unset. Right now if the referred schema is
			// nullable we will resolve to a nullable schema.
			// There are no upstream schemas where nullable is used as a field
			// level override, so we will assume `false` means `unset`.
			// But this should be fixed in kube-openapi.
			resolvedCopy.Nullable = resolvedCopy.Nullable || sch.Nullable

			if len(sch.Type) > 0 {
				resolvedCopy.Type = sch.Type
			}

			if len(sch.Description) > 0 {
				resolvedCopy.Description = sch.Description
			}

			newExtensions := kubespec.Extensions{}
			for k, v := range resolvedCopy.Extensions {
				newExtensions.Add(k, v)
			}
			for k, v := range sch.Extensions {
				newExtensions.Add(k, v)
			}
			if len(newExtensions) > 0 {
				resolvedCopy.Extensions = newExtensions
			}

			// Don't explore children. This was a reference node and shares
			// pointers with its schema which will be traversed in this loop.
			return &resolvedCopy, false
		}))
	}

	if len(referenceErrors) > 0 {
		return nil, errors.Join(referenceErrors...)
	}

	namespaced := sets.New[schema.GroupVersionKind]()
	if openapiSpec.Paths != nil {
		for path, pathInfo := range openapiSpec.Paths.Paths {
			for _, gvk := range utils.ExtractPathGVKs(pathInfo) {
				if !namespaced.Has(gvk) {
					if strings.Contains(path, "namespaces/{namespace}") {
						namespaced.Insert(gvk)
					}
				}
			}
		}
	}

	for _, def := range openapiSpec.Components.Schemas {
		gvks := utils.ExtractExtensionGVKs(def.Extensions)
		if len(gvks) == 0 {
			continue
		}

		for _, specGVK := range gvks {
			d.schemaCache[specGVK] = def
		}
	}

	// Check again to see if the desired GVK was added to the spec cache.
	// If so, create validator for it
	if existing, ok := d.schemaCache[gvk]; ok {
		return existing, nil
	}

	return nil, fmt.Errorf("kind %v not found in %v groupversion", gvk.Kind, gvk.GroupVersion())
}
