// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import (
	"encoding/json"
	"strings"

	"sigs.k8s.io/yaml"
)

// MarshalJSON overrides the default MarshalJSON logic
func (j *JSONPatchOperation) MarshalJSON() ([]byte, error) {
	value := j.Value
	if isYAML(j.Value) {
		jsonBytes, err := yaml.YAMLToJSON([]byte(j.Value))
		if err != nil {
			return nil, err
		}
		value = string(jsonBytes)
	}
	const placeHolder = "jsonValuePlaceHolder"

	// use an anonymous struct to avoid infinite recursive call to JSONPatchOperation.MarshalJSON
	tmp := struct {
		Op    JSONPatchOperationType `json:"op"`
		Path  string                 `json:"path"`
		Value string                 `json:"value"`
	}{
		Op:    j.Op,
		Path:  j.Path,
		Value: placeHolder,
	}

	jsonBytes, err := json.Marshal(tmp)
	if err != nil {
		return nil, err
	}

	jsonStr := strings.Replace(string(jsonBytes), "\""+placeHolder+"\"", value, 1)
	return []byte(jsonStr), nil
}

// UnmarshalJSON overrides the default UnmarshalJSON logic
func (j *JSONPatchOperation) UnmarshalJSON(jsonBytes []byte) error {
	var jsonData map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &jsonData); err != nil {
		return err
	}
	value := jsonData["value"]
	jsonValue, err := json.Marshal(value)
	if err != nil {
		return err
	}

	// convert json to yaml because yaml is more readable
	yamlValue, err := yaml.JSONToYAML(jsonValue)
	if err != nil {
		return err
	}

	jsonData["value"] = ""
	operationBytes, err := json.Marshal(jsonData)
	if err != nil {
		return err
	}

	// use an anonymous struct to avoid infinite recursive call to JSONPatchOperation.UnmarshalJSON
	tmp := struct {
		Op    JSONPatchOperationType `json:"op"`
		Path  string                 `json:"path"`
		Value string                 `json:"value"`
	}{
		Op:    j.Op,
		Path:  j.Path,
		Value: "",
	}
	if err := json.Unmarshal(operationBytes, &tmp); err != nil {
		return err
	}

	j.Op = tmp.Op
	j.Path = tmp.Path
	j.Value = string(yamlValue)
	return nil
}

func isYAML(data string) bool {
	var yamlData interface{}
	err := yaml.Unmarshal([]byte(data), &yamlData)
	return err == nil
}
