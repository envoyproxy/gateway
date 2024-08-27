// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package bootstrap

import (
	"fmt"

	bootstrapv3 "github.com/envoyproxy/go-control-plane/envoy/config/bootstrap/v3"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/utils/proto"
	_ "github.com/envoyproxy/gateway/internal/xds/extensions" // DON'T REMOVE: import of all extensions
)

// ApplyBootstrapConfig applies the bootstrap config to the default bootstrap config and return the result config.
func ApplyBootstrapConfig(boostrapConfig *egv1a1.ProxyBootstrap, defaultBootstrap string) (string, error) {
	bootstrapType := boostrapConfig.Type
	if bootstrapType == nil {
		// The documentation defines that a nil bootstrapType defaults to the "Replace" operation
		return *boostrapConfig.Value, nil
	}
	switch *bootstrapType {
	case egv1a1.BootstrapTypeMerge:
		mergedBootstrap, err := mergeBootstrap(defaultBootstrap, boostrapConfig.Value)
		if err != nil {
			return "", err
		}
		return mergedBootstrap, nil
	case egv1a1.BootstrapTypeReplace:
		return *boostrapConfig.Value, nil
	case egv1a1.BootstrapTypeJSONPatch:
		// TODO: Implement
		return defaultBootstrap, nil
	default:
		// This is unreachable code due to the CEL validation on egv1a1.ProxyBootstrap
		return defaultBootstrap, fmt.Errorf("unsupported bootstrap patch type %s", *bootstrapType)
	}
}

func mergeBootstrap(base string, override *string) (string, error) {
	if override == nil {
		return base, nil
	}
	dst := &bootstrapv3.Bootstrap{}
	if err := proto.FromYAML([]byte(base), dst); err != nil {
		return "", fmt.Errorf("failed to parse default bootstrap config: %w", err)
	}

	src := &bootstrapv3.Bootstrap{}
	if err := proto.FromYAML([]byte(*override), src); err != nil {
		return "", fmt.Errorf("failed to parse override bootstrap config: %w", err)
	}

	proto.Merge(dst, src)

	if err := dst.Validate(); err != nil {
		return "", fmt.Errorf("failed to validate merged bootstrap config: %w", err)
	}

	data, err := proto.ToYAML(dst)
	if err != nil {
		return "", fmt.Errorf("failed to convert proto message to YAML: %w", err)
	}

	return string(data), nil
}
