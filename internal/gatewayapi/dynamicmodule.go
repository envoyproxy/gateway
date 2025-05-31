// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"fmt"
	"strconv"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/ir"
)

func (t *Translator) buildDynamicModules(
	policy *egv1a1.EnvoyExtensionPolicy,
	_ *resource.Resources,
) ([]ir.DynamicModule, error) {
	var dynamicModuleIRList []ir.DynamicModule

	if policy == nil {
		return nil, nil
	}

	for idx, dm := range policy.Spec.DynamicModules {
		name := irConfigNameForDynamicModule(policy, idx)
		dynamicModuleIR, err := t.buildDynamicModule(name, dm)
		if err != nil {
			return nil, err
		}
		dynamicModuleIRList = append(dynamicModuleIRList, *dynamicModuleIR)
	}
	return dynamicModuleIRList, nil
}

func (t *Translator) buildDynamicModule(
	name string,
	config egv1a1.DynamicModule,
) (*ir.DynamicModule, error) {
	// Validate required fields
	if config.Module == "" {
		return nil, fmt.Errorf("module is required")
	}

	dynamicModuleName := name
	if config.ExtensionName != nil {
		dynamicModuleName = *config.ExtensionName
	}

	// Set DoNotClose if specified
	doNotClose := false
	if config.DoNotClose != nil {
		doNotClose = *config.DoNotClose
	}

	dynamicModuleIR := &ir.DynamicModule{
		ExtensionName:   dynamicModuleName,
		Module:          config.Module,
		ExtensionConfig: config.ExtensionConfig,
		DoNotClose:      doNotClose,
	}

	return dynamicModuleIR, nil
}

func irConfigNameForDynamicModule(policy *egv1a1.EnvoyExtensionPolicy, index int) string {
	return fmt.Sprintf(
		"%s/dynamicmodule/%s",
		irConfigName(policy),
		strconv.Itoa(index))
}
