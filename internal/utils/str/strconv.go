// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package str

// This file contains utility functions copied from github.com/prometheus/prometheus/util/strutil,
// which is conflicting with the current package.

import "regexp"

var invalidLabelCharRE = regexp.MustCompile(`[^a-zA-Z0-9_]`)

// SanitizeLabelName replaces anything that doesn't match
// client_label.LabelNameRE with an underscore.
// Note: this does not handle all Prometheus label name restrictions (such as
// not starting with a digit 0-9), and hence should only be used if the label
// name is prefixed with a known valid string.
func SanitizeLabelName(name string) string {
	return invalidLabelCharRE.ReplaceAllString(name, "_")
}
