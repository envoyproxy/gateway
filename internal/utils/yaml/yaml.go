// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package yaml

import (
	"reflect"

	"sigs.k8s.io/yaml"
)

// MergeYAML merges two yaml files. The second yaml file will override the first one if the same key exists.
// This method can add or override a value within a map, or add a new value to a list.
// Please note that this method can't override a value within a list.
func MergeYAML(base, override string) (string, error) {
	// declare two map to hold the yaml content
	map1 := map[string]interface{}{}
	map2 := map[string]interface{}{}

	if err := yaml.Unmarshal([]byte(base), &map1); err != nil {
		return "", err
	}

	if err := yaml.Unmarshal([]byte(override), &map2); err != nil {
		return "", err
	}

	// merge both yaml data recursively
	result := mergeMaps(map1, map2)

	out, err := yaml.Marshal(result)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func mergeMaps(map1, map2 map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(map1))
	for k, v := range map1 {
		out[k] = v
	}
	for k, v := range map2 {
		if v, ok := v.(map[string]interface{}); ok {
			if bv, ok := out[k]; ok {
				if bv, ok := bv.(map[string]interface{}); ok {
					out[k] = mergeMaps(bv, v)
					continue
				}
			}
		}
		value := reflect.ValueOf(v)
		if value.Kind() == reflect.Array || value.Kind() == reflect.Slice {
			out[k] = append(out[k].([]interface{}), v.([]interface{})...)
		} else {
			out[k] = v
		}
	}
	return out
}
