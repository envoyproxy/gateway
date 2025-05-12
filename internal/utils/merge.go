// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package utils

import (
	"encoding/json"
	"fmt"

	jsonpatch "github.com/evanphx/json-patch"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"sigs.k8s.io/controller-runtime/pkg/client"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
)

func MergeWithPatch[T client.Object](original T, patch *egv1a1.KubernetesPatchSpec) (T, error) {
	if patch == nil {
		return original, nil
	}

	mergeType := egv1a1.StrategicMerge
	if patch.Type != nil {
		mergeType = *patch.Type
	}

	return mergeInternal(original, patch.Value.Raw, mergeType)
}

func mergeInternal[T client.Object](original T, patchJSON []byte, mergeType egv1a1.MergeType) (T, error) {
	var (
		patchedJSON  []byte
		originalJSON []byte
		err          error
		empty        T
	)

	originalJSON, err = json.Marshal(original)
	if err != nil {
		return empty, fmt.Errorf("error marshaling original service: %w", err)
	}
	switch mergeType {
	case egv1a1.StrategicMerge:
		patchedJSON, err = strategicpatch.StrategicMergePatch(originalJSON, patchJSON, empty)
		if err != nil {
			return empty, fmt.Errorf("error during strategic merge: %w", err)
		}
	case egv1a1.JSONMerge:
		patchedJSON, err = jsonpatch.MergePatch(originalJSON, patchJSON)
		if err != nil {
			return empty, fmt.Errorf("error during JSON merge: %w", err)
		}
	default:
		return empty, fmt.Errorf("unsupported merge type: %v", mergeType)
	}

	res := new(T)
	if err = json.Unmarshal(patchedJSON, res); err != nil {
		return empty, fmt.Errorf("error unmarshaling patched service: %w", err)
	}

	return *res, nil
}

func Merge[T client.Object](original, patch T, mergeType egv1a1.MergeType) (T, error) {
	var (
		patchJSON []byte
		err       error
		empty     T
	)

	patchJSON, err = json.Marshal(patch)
	if err != nil {
		return empty, fmt.Errorf("error marshaling original service: %w", err)
	}
	return mergeInternal(original, patchJSON, mergeType)
}

func MergeGlobalRLInternal[T *ir.GlobalRateLimit](original T, patchJSON []byte, mergeType egv1a1.MergeType) (T, error) {
	var (
		patchedJSON  []byte
		originalJSON []byte
		err          error
		empty        T
	)

	originalJSON, err = json.Marshal(original)
	if err != nil {
		return empty, fmt.Errorf("error marshaling original service: %w", err)
	}
	switch mergeType {
	case egv1a1.StrategicMerge:
		patchedJSON, err = strategicpatch.StrategicMergePatch(originalJSON, patchJSON, empty)
		if err != nil {
			return empty, fmt.Errorf("error during strategic merge: %w", err)
		}
	case egv1a1.JSONMerge:
		patchedJSON, err = jsonpatch.MergePatch(originalJSON, patchJSON)
		if err != nil {
			return empty, fmt.Errorf("error during JSON merge: %w", err)
		}
	default:
		return empty, fmt.Errorf("unsupported merge type: %v", mergeType)
	}

	res := new(T)
	if err = json.Unmarshal(patchedJSON, res); err != nil {
		return empty, fmt.Errorf("error unmarshaling patched service: %w", err)
	}

	return *res, nil
}

func MergeGlobalRL[T *ir.GlobalRateLimit](original, patch T, mergeType egv1a1.MergeType) (T, error) {
	var (
		patchJSON []byte
		err       error
		empty     T
	)

	patchJSON, err = json.Marshal(patch)
	if err != nil {
		return empty, fmt.Errorf("error marshaling original service: %w", err)
	}
	return MergeGlobalRLInternal(original, patchJSON, mergeType)
}
