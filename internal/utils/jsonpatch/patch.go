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
	"sigs.k8s.io/yaml"

	"github.com/envoyproxy/gateway/internal/ir"
)

// ApplyJSONPatches applies a series of JSONPatches to a provided JSON document.
// Patches are applied in order, and any errors are aggregated into the return value.
// An error with a specific patch just means that this specific patch is skipped, the document
// will still be modified with any other provided patch operation.
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

		for _, path := range jsonPointers {
			op := ir.JSONPatchOperation{
				Path:  &path,
				Op:    p.Op,
				Value: p.Value,
				From:  p.From,
			}

			// Convert patch to JSON
			// The patch library expects an array so convert it into one
			y, err := yaml.Marshal([]ir.JSONPatchOperation{op})
			if err != nil {
				tErr := fmt.Errorf("unable to marshal patch %+v, err: %s", op, err.Error())
				tErrs = errors.Join(tErrs, tErr)
				continue
			}
			jsonBytes, err := yaml.YAMLToJSON(y)
			if err != nil {
				tErr := fmt.Errorf("unable to convert patch to json %s, err: %s", string(y), err.Error())
				tErrs = errors.Join(tErrs, tErr)
				continue
			}
			patchObj, err := jsonpatchv5.DecodePatch(jsonBytes)
			if err != nil {
				tErr := fmt.Errorf("unable to decode patch %s, err: %s", string(jsonBytes), err.Error())
				tErrs = errors.Join(tErrs, tErr)
				continue
			}

			// Apply patch
			document, err = patchObj.ApplyWithOptions(document, opts)
			if err != nil {
				tErr := fmt.Errorf("unable to apply patch:\n%s on resource:\n%s, err: %s", string(jsonBytes), string(document), err.Error())
				tErrs = errors.Join(tErrs, tErr)
				continue
			}
		}
	}
	return document, tErrs
}
