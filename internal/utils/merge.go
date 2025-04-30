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
)

func Merge[T client.Object](original T, patch T, mergeType egv1a1.MergeType) (T, error) {
	var (
		patchedJSON  []byte
		originalJSON []byte
		patchJSON    []byte
		err          error
		empty        T
	)

	originalJSON, err = json.Marshal(original)
	if err != nil {
		return empty, fmt.Errorf("error marshaling original service: %w", err)
	}
	patchJSON, err = json.Marshal(patch)
	if err != nil {
		return empty, fmt.Errorf("error marshaling original service: %w", err)
	}
	switch mergeType {
	case egv1a1.StrategicMerge:
		patchedJSON, err = strategicpatch.StrategicMergePatch(originalJSON, patchJSON, egv1a1.BackendTrafficPolicy{})
		if err != nil {
			return empty, fmt.Errorf("error during strategic merge: %w", err)
		}
	case egv1a1.JSONMerge:
		patchedJSON, err = jsonpatch.MergePatch(originalJSON, patchJSON)
		if err != nil {
			return empty, fmt.Errorf("error during JSON merge: %w", err)
		}
	default:
		return empty, fmt.Errorf("unsupported merge type: %s", mergeType)
	}

	res := new(T)
	if err = json.Unmarshal(patchedJSON, res); err != nil {
		return empty, fmt.Errorf("error unmarshaling patched service: %w", err)
	}

	return *res, nil
}
