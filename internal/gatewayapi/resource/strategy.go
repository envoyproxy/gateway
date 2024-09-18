// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package resource

import (
	"context"
	"fmt"
	"sync"

	apiextensionshelpers "k8s.io/apiextensions-apiserver/pkg/apihelpers"
	apiextensionsinternal "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apiextensions-apiserver/pkg/apiserver"
	structuralschema "k8s.io/apiextensions-apiserver/pkg/apiserver/schema"
	structuraldefaulting "k8s.io/apiextensions-apiserver/pkg/apiserver/schema/defaulting"
	apiservervalidation "k8s.io/apiextensions-apiserver/pkg/apiserver/validation"
	"k8s.io/apiextensions-apiserver/pkg/crdserverscheme"
	"k8s.io/apiextensions-apiserver/pkg/registry/customresource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apiserver/pkg/registry/rest"
)

// versionedCreateStrategyMap only stores latest version of CreateStrategy.
type versionedCreateStrategyMap map[string]rest.RESTCreateStrategy

func (v versionedCreateStrategyMap) Validate(kind string, obj runtime.Object) error {
	createStrategy, ok := v[kind]
	if !ok {
		return nil
	}

	errList := createStrategy.Validate(context.TODO(), obj)
	if errList != nil {
		return errList.ToAggregate()
	}

	return nil
}

var (
	// crdCreateStrategyMap stores CreateStrategy for supported CRDs.
	// Since it's a lazy loading map, calling getCreateStrategyMapForCRDs
	// before using this map, directly using this map is not recommended.
	crdCreateStrategyMap versionedCreateStrategyMap
	once                 sync.Once
)

func getCreateStrategyMapForCRDs() versionedCreateStrategyMap {
	once.Do(func() {
		// This error should never happen. If it does, stop loading strategy map.
		crds, _ := loadCRDs()
		if len(crds) == 0 {
			return
		}
		crdCreateStrategyMap = make(versionedCreateStrategyMap, len(crds))
		for _, crd := range crds {
			vcs, _ := getVersionedCreateStrategyForCRD(crd)
			crdCreateStrategyMap[crd.Spec.Names.Kind] = vcs
		}
	})
	return crdCreateStrategyMap
}

// getVersionedCreateStrategyForCRD will only get latest version of the input CRD and
// generate its CreateStrategy correspondingly.
//
// This function contains code copied and modified from kubernetes/apiextensions-apiserver,
// https://github.com/kubernetes/apiextensions-apiserver
// from the source file
// https://github.com/kubernetes/apiextensions-apiserver/blob/master/pkg/apiserver/customresource_handler.go#L605
// and is provided here subject to the following:
// Copyright Project Kubernetes Authors
// SPDX-License-Identifier: Apache-2.0
//
// Modifications:
// - Remove everything except how to derive NewStrategy, and separate them into some new functions.
func getVersionedCreateStrategyForCRD(crd *apiextensionsv1.CustomResourceDefinition) (rest.RESTCreateStrategy, error) {
	structuralSchema, err := getStructuralSchema(crd)
	if err != nil {
		return nil, err
	}

	v := crd.Spec.Versions[0]
	// In addition to Unstructured objects (Custom Resources), we also may sometimes need to
	// decode unversioned Options objects, so we delegate to parameterScheme for such types.
	parameterScheme := runtime.NewScheme()
	parameterScheme.AddUnversionedTypes(schema.GroupVersion{Group: crd.Spec.Group, Version: v.Name},
		&metav1.ListOptions{},
		&metav1.GetOptions{},
		&metav1.DeleteOptions{},
	)

	kind := schema.GroupVersionKind{Group: crd.Spec.Group, Version: v.Name, Kind: crd.Status.AcceptedNames.Kind}
	typer := apiserver.UnstructuredObjectTyper{
		Delegate:          parameterScheme,
		UnstructuredTyper: crdserverscheme.NewUnstructuredObjectTyper(),
	}

	validationSchema, err := apiextensionshelpers.GetSchemaForVersion(crd, v.Name)
	if err != nil {
		return nil, fmt.Errorf("the server could not properly serve the CR schema")
	}

	var internalSchemaProps *apiextensionsinternal.JSONSchemaProps
	var internalValidationSchema *apiextensionsinternal.CustomResourceValidation
	if validationSchema != nil {
		internalValidationSchema = &apiextensionsinternal.CustomResourceValidation{}
		if err = apiextensionsv1.Convert_v1_CustomResourceValidation_To_apiextensions_CustomResourceValidation(validationSchema, internalValidationSchema, nil); err != nil {
			return nil, fmt.Errorf("failed to convert CRD validation to internal version: %w", err)
		}
		internalSchemaProps = internalValidationSchema.OpenAPIV3Schema
	}

	validator, _, err := apiservervalidation.NewSchemaValidator(internalSchemaProps)
	if err != nil {
		return nil, err
	}

	var statusSpec *apiextensionsinternal.CustomResourceSubresourceStatus
	var statusValidator apiservervalidation.SchemaValidator
	subResources, err := apiextensionshelpers.GetSubresourcesForVersion(crd, v.Name)
	if err != nil {
		return nil, fmt.Errorf("the server could not properly serve the CR subresources")
	}
	if subResources != nil && subResources.Status != nil {
		statusSpec = &apiextensionsinternal.CustomResourceSubresourceStatus{}
		if err = apiextensionsv1.Convert_v1_CustomResourceSubresourceStatus_To_apiextensions_CustomResourceSubresourceStatus(subResources.Status, statusSpec, nil); err != nil {
			return nil, fmt.Errorf("failed converting CRD status subresource to internal version: %w", err)
		}
		// for the status subresource, validate only against the status schema
		if internalValidationSchema != nil && internalValidationSchema.OpenAPIV3Schema != nil && internalValidationSchema.OpenAPIV3Schema.Properties != nil {
			if statusSchema, ok := internalValidationSchema.OpenAPIV3Schema.Properties["status"]; ok {
				statusValidator, _, err = apiservervalidation.NewSchemaValidator(&statusSchema)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	var scaleSpec *apiextensionsinternal.CustomResourceSubresourceScale
	if subResources != nil && subResources.Scale != nil {
		scaleSpec = &apiextensionsinternal.CustomResourceSubresourceScale{}
		if err = apiextensionsv1.Convert_v1_CustomResourceSubresourceScale_To_apiextensions_CustomResourceSubresourceScale(subResources.Scale, scaleSpec, nil); err != nil {
			return nil, fmt.Errorf("failed converting CRD status subresource to internal version: %w", err)
		}
	}

	return customresource.NewStrategy(
		typer,
		crd.Spec.Scope == apiextensionsv1.NamespaceScoped,
		kind,
		validator,
		statusValidator,
		structuralSchema,
		statusSpec,
		scaleSpec,
		v.SelectableFields,
	), nil
}

func getStructuralSchema(crd *apiextensionsv1.CustomResourceDefinition) (*structuralschema.Structural, error) {
	v := crd.Spec.Versions[0]

	val, err := apiextensionshelpers.GetSchemaForVersion(crd, v.Name)
	if err != nil {
		return nil, fmt.Errorf("the server could not properly serve the CR schema")
	}

	internalValidation := &apiextensionsinternal.CustomResourceValidation{}
	if err = apiextensionsv1.Convert_v1_CustomResourceValidation_To_apiextensions_CustomResourceValidation(val, internalValidation, nil); err != nil {
		return nil, fmt.Errorf("failed converting CRD validation to internal version: %w", err)
	}

	s, err := structuralschema.NewStructural(internalValidation.OpenAPIV3Schema)
	if !crd.Spec.PreserveUnknownFields && err != nil {
		// This should never happen. If it does, it is a programming error.
		return nil, fmt.Errorf("the server could not properly serve the CR schema") // validation should avoid this
	}

	if !crd.Spec.PreserveUnknownFields {
		// we don't own s completely, e.g. defaults are not deep-copied. So better make a copy here.
		s = s.DeepCopy()

		if err = structuraldefaulting.PruneDefaults(s); err != nil {
			// This should never happen. If it does, it is a programming error.
			return nil, fmt.Errorf("the server could not properly serve the CR schema") // validation should avoid this
		}
	}

	return s, nil
}
