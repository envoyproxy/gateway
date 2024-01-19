// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package regex

import (
	"fmt"
	"regexp"
)

// Validate validates a regex string.
func Validate(regex string) error {
	if _, err := regexp.Compile(regex); err != nil {
		return fmt.Errorf("regex %q is invalid: %w", regex, err)
	}
	return nil
}
