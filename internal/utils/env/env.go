// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package env

import (
	"os"
	"strconv"
	"time"
)

type Var interface {
	string | int | time.Duration
}

// Lookup get specific value by env key, default value will be used when not found and invalid convert.
func Lookup[T Var](key string, defaultValue T) T {
	value, ok := os.LookupEnv(key)
	if !ok {
		return defaultValue
	}

	var ret any
	switch any(defaultValue).(type) {
	case time.Duration:
		d, err := time.ParseDuration(value)
		if err != nil {
			return defaultValue
		}
		ret = d
	case string:
		ret = value
	case int:
		i, err := strconv.ParseInt(value, 10, 0)
		if err != nil {
			return defaultValue
		}
		ret = int(i)
	}
	return ret.(T)
}
