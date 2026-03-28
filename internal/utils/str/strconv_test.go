// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package str

import "testing"

func TestSanitizeLabelName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "already valid",
			input: "valid_label_123",
			want:  "valid_label_123",
		},
		{
			name:  "hyphens replaced",
			input: "my-label-name",
			want:  "my_label_name",
		},
		{
			name:  "dots replaced",
			input: "my.label.name",
			want:  "my_label_name",
		},
		{
			name:  "mixed special characters",
			input: "my-label.name/path:value",
			want:  "my_label_name_path_value",
		},
		{
			name:  "spaces replaced",
			input: "my label name",
			want:  "my_label_name",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SanitizeLabelName(tt.input); got != tt.want {
				t.Errorf("SanitizeLabelName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
