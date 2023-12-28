// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package regex

import "testing"

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		regex   string
		wantErr bool
	}{
		{
			name:    "Valid regex",
			regex:   "^[a-z0-9-]+$",
			wantErr: false,
		},
		{
			name:    "Invalid regex",
			regex:   "^[a-z0-9-++$",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Validate(tt.regex); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
