// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"testing"
)

func Test_validateRegex(t *testing.T) {
	tests := []struct {
		name    string
		regex   string
		wantErr bool
	}{
		{
			name:    "valid regex",
			regex:   "^[a-zA-Z0-9-]+$",
			wantErr: false,
		},
		{
			name:    "invalid regex",
			regex:   "*.foo.com",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRegex(tt.regex)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateRegex() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
