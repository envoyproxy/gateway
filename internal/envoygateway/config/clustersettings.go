// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package config

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"

	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	structuralschema "k8s.io/apiextensions-apiserver/pkg/apiserver/schema"
	"k8s.io/apiextensions-apiserver/pkg/apiserver/schema/cel"
	"k8s.io/apiextensions-apiserver/pkg/apiserver/schema/defaulting"
	"k8s.io/apiextensions-apiserver/pkg/apiserver/schema/listtype"
	apiservervalidation "k8s.io/apiextensions-apiserver/pkg/apiserver/validation"
	utiljson "k8s.io/apimachinery/pkg/util/json"
	"k8s.io/apimachinery/pkg/util/validation/field"
	utilyaml "k8s.io/apimachinery/pkg/util/yaml"
	celconfig "k8s.io/apiserver/pkg/apis/cel"

	envoygateway "github.com/envoyproxy/gateway"
	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

// EnvoyGateway is loaded as static configuration rather than admitted as a CRD, so the
// kubebuilder defaults, OpenAPI constraints and CEL rules declared on ClusterSettings are
// never applied by the API server. ClusterSettings is not a CRD on its own either, so its
// schema is borrowed from a CRD that embeds it through BackendSettings.
const (
	clusterSettingsSchemaCRD = "securitypolicies.gateway.envoyproxy.io"
)

// clusterSettingsSchemaPath is the path of a BackendSettings node in clusterSettingsSchemaCRD.
var clusterSettingsSchemaPath = []string{"spec", "extAuth", "grpc", "backendSettings"}

type clusterSettingsSchema struct {
	structural    *structuralschema.Structural
	schemaChecker apiservervalidation.SchemaValidator
	celValidator  *cel.Validator
}

var loadClusterSettingsSchema = sync.OnceValues(func() (*clusterSettingsSchema, error) {
	props, err := findClusterSettingsSchema()
	if err != nil {
		return nil, err
	}

	internal := &apiextensions.JSONSchemaProps{}
	if err := apiextensionsv1.Convert_v1_JSONSchemaProps_To_apiextensions_JSONSchemaProps(props, internal, nil); err != nil {
		return nil, err
	}
	structural, err := structuralschema.NewStructural(internal)
	if err != nil {
		return nil, err
	}
	schemaChecker, _, err := apiservervalidation.NewSchemaValidator(internal)
	if err != nil {
		return nil, err
	}

	return &clusterSettingsSchema{
		structural:    structural,
		schemaChecker: schemaChecker,
		celValidator:  cel.NewValidator(structural, false, celconfig.PerCallLimit),
	}, nil
})

func findClusterSettingsSchema() (*apiextensionsv1.JSONSchemaProps, error) {
	decoder := utilyaml.NewYAMLOrJSONDecoder(bytes.NewReader(envoygateway.GatewayCRDs), 4096)
	for {
		crd := &apiextensionsv1.CustomResourceDefinition{}
		if err := decoder.Decode(crd); err != nil {
			if errors.Is(err, io.EOF) {
				return nil, fmt.Errorf("CRD %s not found", clusterSettingsSchemaCRD)
			}
			return nil, err
		}
		if crd.Name != clusterSettingsSchemaCRD {
			continue
		}

		for _, v := range crd.Spec.Versions {
			if v.Name != egv1a1.GroupVersion.Version || v.Schema == nil || v.Schema.OpenAPIV3Schema == nil {
				continue
			}
			props := v.Schema.OpenAPIV3Schema
			for _, p := range clusterSettingsSchemaPath {
				next, ok := props.Properties[p]
				if !ok {
					return nil, fmt.Errorf("schema of %v not found in CRD %s", clusterSettingsSchemaPath, clusterSettingsSchemaCRD)
				}
				props = &next
			}
			return props, nil
		}
		return nil, fmt.Errorf("version %s not found in CRD %s", egv1a1.GroupVersion.Version, clusterSettingsSchemaCRD)
	}
}

// DefaultClusterSettings returns a copy of cs with the defaults declared in the
// ClusterSettings CRD schema applied, as the API server would on admission.
func DefaultClusterSettings(cs *egv1a1.ClusterSettings) (*egv1a1.ClusterSettings, error) {
	if cs == nil {
		return nil, nil
	}
	s, err := loadClusterSettingsSchema()
	if err != nil {
		return nil, err
	}
	obj, err := toUnstructured(cs)
	if err != nil {
		return nil, err
	}
	defaulting.Default(obj, s.structural)

	data, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}
	out := &egv1a1.ClusterSettings{}
	if err := json.Unmarshal(data, out); err != nil {
		return nil, err
	}
	return out, nil
}

// ValidateClusterSettings validates cs against the OpenAPI constraints, list types and CEL
// rules declared in the ClusterSettings CRD schema, after applying its defaults, as the API
// server would on admission.
//
// It doesn't check whether a field is supported where cs is used.
func ValidateClusterSettings(fldPath *field.Path, cs *egv1a1.ClusterSettings) error {
	if cs == nil {
		return nil
	}
	s, err := loadClusterSettingsSchema()
	if err != nil {
		return err
	}
	obj, err := toUnstructured(cs)
	if err != nil {
		return err
	}
	defaulting.Default(obj, s.structural)

	errs := apiservervalidation.ValidateCustomResource(fldPath, obj, s.schemaChecker)
	errs = append(errs, listtype.ValidateListSetsAndMaps(fldPath, s.structural, obj)...)
	celErrs, _ := s.celValidator.Validate(context.Background(), fldPath, s.structural, obj, nil, celconfig.RuntimeCELCostBudget)
	errs = append(errs, celErrs...)
	return errs.ToAggregate()
}

func toUnstructured(cs *egv1a1.ClusterSettings) (map[string]any, error) {
	data, err := json.Marshal(cs)
	if err != nil {
		return nil, err
	}
	// Unlike encoding/json, utiljson decodes whole numbers as int64, which the
	// integer-typed schema nodes and CEL rules expect.
	obj := map[string]any{}
	if err := utiljson.Unmarshal(data, &obj); err != nil {
		return nil, err
	}
	return obj, nil
}
