// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package jsonpatch

import (
	"encoding/json"
	"errors"
	"fmt"

	jsonpatchv5 "github.com/evanphx/json-patch/v5"

	"github.com/envoyproxy/gateway/internal/ir"
)

// ApplyJSONPatches applies a series of JSONPatches to a provided JSON document.
// Patches are applied in order. Validation and JSONPath evaluation errors skip the
// affected patch and are aggregated. If applying a patch fails, processing stops
// and no modified document is returned. Callers must discard the result on any error.
// If a patch is applied to a JSONPath, then that JSONPath is first exploded to standard paths
// and the patch is applied to all matching paths.
func ApplyJSONPatches(document json.RawMessage, patches ...ir.JSONPatchOperation) (json.RawMessage, error) {
	opts := jsonpatchv5.NewApplyOptions()
	opts.EnsurePathExistsOnAdd = true

	var tErrs, err error
	for _, p := range patches {

		if err := p.Validate(); err != nil {
			tErrs = errors.Join(tErrs, err)
			continue
		}

		var jsonPointers []string
		if p.JSONPath != nil {
			path := ""
			if p.Path != nil {
				path = *p.Path
			}
			jsonPointers, err = ConvertPathToPointers(document, *p.JSONPath, path)
			if err != nil {
				tErr := fmt.Errorf("unable to convert jsonPath: '%s' into jsonPointers, err: %s", *p.JSONPath, err.Error())
				tErrs = errors.Join(tErrs, tErr)
				continue
			}
			if len(jsonPointers) == 0 {
				tErr := fmt.Errorf("no jsonPointers were found while evaluating the jsonPath: '%s'. "+
					"Ensure the elements you are trying to select with the jsonPath exist in the document. "+
					"If you need to add a non-existing property, use the 'path' attribute", *p.JSONPath)
				tErrs = errors.Join(tErrs, tErr)
				continue
			}
		} else {
			jsonPointers = []string{*p.Path}
		}

		patch := make(jsonpatchv5.Patch, 0, len(jsonPointers))
		for _, path := range jsonPointers {
			operation, err := toPatchOperation(path, p)
			if err != nil {
				tErrs = errors.Join(tErrs, err)
				continue
			}
			patch = append(patch, operation)
		}
		if len(patch) == 0 {
			continue
		}

		// Apply all matches with a single decode and encode of the document.
		document, err = patch.ApplyWithOptions(document, opts)
		if err != nil {
			tErr := fmt.Errorf("unable to apply patch: op=%s err: %w", string(p.Op), err)
			return nil, errors.Join(tErrs, tErr)
		}
	}
	return document, tErrs
}

func toPatchOperation(path string, original ir.JSONPatchOperation) (jsonpatchv5.Operation, error) {
	operation := make(jsonpatchv5.Operation, 4)

	rawOp, err := marshalJSONString(string(original.Op))
	if err != nil {
		return nil, fmt.Errorf("unable to marshal patch op %q: %w", string(original.Op), err)
	}
	operation["op"] = rawOp

	rawPath, err := marshalJSONString(path)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal patch path %q: %w", path, err)
	}
	operation["path"] = rawPath

	if original.From != nil {
		rawFrom, err := marshalJSONString(*original.From)
		if err != nil {
			return nil, fmt.Errorf("unable to marshal patch from %q: %w", *original.From, err)
		}
		operation["from"] = rawFrom
	}

	if original.Value != nil {
		operation["value"] = cloneRawJSON(original.Value.Raw)
	}

	return operation, nil
}

func marshalJSONString(value string) (*json.RawMessage, error) {
	b, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	raw := json.RawMessage(b)
	return &raw, nil
}

func cloneRawJSON(src []byte) *json.RawMessage {
	if src == nil {
		rawNull := json.RawMessage([]byte("null"))
		return &rawNull
	}
	cloned := make(json.RawMessage, len(src))
	copy(cloned, src)
	return &cloned
}
