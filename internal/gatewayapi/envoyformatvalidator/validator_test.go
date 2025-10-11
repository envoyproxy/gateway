// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package envoyformatvalidator

import (
	"strings"
	"testing"
)

func TestValidateEnvoyFormatString(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		wantError bool
		errMsg    string
	}{
		// Valid cases
		{
			name:      "empty string",
			value:     "",
			wantError: false,
		},
		{
			name:      "plain text no operators",
			value:     "simple-value",
			wantError: false,
		},
		{
			name:      "valid REQ with single header",
			value:     "%REQ(:METHOD)%",
			wantError: false,
		},
		{
			name:      "valid REQ with one alternative",
			value:     "%REQ(X-Real-IP?x-client-ip)%",
			wantError: false,
		},
		{
			name:      "valid REQ with pseudo header",
			value:     "%REQ(:AUTHORITY)%",
			wantError: false,
		},
		{
			name:      "valid RESP with alternative",
			value:     "%RESP(content-type?Content-Type)%",
			wantError: false,
		},
		{
			name:      "valid TRAILER with alternative",
			value:     "%TRAILER(grpc-status?Grpc-Status)%",
			wantError: false,
		},
		{
			name:      "valid START_TIME with format",
			value:     "%START_TIME(%s.%6f)%",
			wantError: false,
		},
		{
			name:      "valid START_TIME simple format",
			value:     "%START_TIME(%Y-%m-%d)%",
			wantError: false,
		},
		{
			name:      "valid DURATION",
			value:     "%DURATION%",
			wantError: false,
		},
		{
			name:      "valid RESPONSE_CODE",
			value:     "%RESPONSE_CODE%",
			wantError: false,
		},
		{
			name:      "valid BYTES_RECEIVED",
			value:     "%BYTES_RECEIVED%",
			wantError: false,
		},
		{
			name:      "valid BYTES_SENT",
			value:     "%BYTES_SENT%",
			wantError: false,
		},
		{
			name:      "valid UPSTREAM_HOST",
			value:     "%UPSTREAM_HOST%",
			wantError: false,
		},
		{
			name:      "valid DOWNSTREAM_REMOTE_ADDRESS",
			value:     "%DOWNSTREAM_REMOTE_ADDRESS%",
			wantError: false,
		},
		{
			name:      "valid with truncation",
			value:     "%REQ(X-Request-Id):10%",
			wantError: false,
		},
		{
			name:      "valid mixed text and operator",
			value:     "t=%START_TIME(%s.%6f)%",
			wantError: false,
		},
		{
			name:      "valid multiple operators",
			value:     "%REQ(:METHOD)% %REQ(:PATH)% %PROTOCOL%",
			wantError: false,
		},
		{
			name:      "valid PROTOCOL",
			value:     "%PROTOCOL%",
			wantError: false,
		},
		{
			name:      "valid RESPONSE_FLAGS",
			value:     "%RESPONSE_FLAGS%",
			wantError: false,
		},
		{
			name:      "valid ENVIRONMENT",
			value:     "%ENVIRONMENT(HOME)%",
			wantError: false,
		},
		{
			name:      "valid QUERY_PARAM",
			value:     "%QUERY_PARAM(foo)%",
			wantError: false,
		},
		{
			name:      "valid PATH",
			value:     "%PATH(WQ)%",
			wantError: false,
		},
		{
			name:      "valid DYNAMIC_METADATA",
			value:     "%DYNAMIC_METADATA(com.test:key)%",
			wantError: false,
		},
		{
			name:      "valid FILTER_STATE",
			value:     "%FILTER_STATE(key)%",
			wantError: false,
		},
		{
			name:      "valid complex header value",
			value:     "Bearer %REQ(Authorization?authorization)%",
			wantError: false,
		},
		// Invalid cases - multiple alternatives
		{
			name:      "invalid REQ with multiple alternatives",
			value:     "%REQ(X-Real-IP?x-client-ip?CF-Connecting-IP)%",
			wantError: true,
			errMsg:    "more than 1 alternative header specified in token",
		},
		{
			name:      "invalid REQ with three alternatives",
			value:     "%REQ(A?B?C?D)%",
			wantError: true,
			errMsg:    "more than 1 alternative header specified in token",
		},
		{
			name:      "invalid RESP with multiple alternatives",
			value:     "%RESP(content-type?Content-Type?CONTENT-TYPE)%",
			wantError: true,
			errMsg:    "more than 1 alternative header specified in token",
		},
		{
			name:      "invalid TRAILER with multiple alternatives",
			value:     "%TRAILER(A?B?C)%",
			wantError: true,
			errMsg:    "more than 1 alternative header specified in token",
		},
		// Invalid cases - unknown operators
		{
			name:      "invalid unknown operator",
			value:     "%INVALID_OPERATOR%",
			wantError: true,
			errMsg:    "unknown command operator",
		},
		{
			name:      "invalid custom operator",
			value:     "%MY_CUSTOM_THING%",
			wantError: true,
			errMsg:    "unknown command operator",
		},
		// Invalid cases - missing required arguments
		{
			name:      "invalid REQ without args",
			value:     "%REQ%",
			wantError: true,
			errMsg:    "command operator REQ requires arguments",
		},
		{
			name:      "invalid RESP without args",
			value:     "%RESP%",
			wantError: true,
			errMsg:    "command operator RESP requires arguments",
		},
		{
			name:      "invalid ENVIRONMENT without args",
			value:     "%ENVIRONMENT%",
			wantError: true,
			errMsg:    "command operator ENVIRONMENT requires arguments",
		},
		{
			name:      "invalid QUERY_PARAM without args",
			value:     "%QUERY_PARAM%",
			wantError: true,
			errMsg:    "command operator QUERY_PARAM requires arguments",
		},
		// Invalid cases - malformed syntax
		{
			name:      "invalid trailing percent",
			value:     "%REQ(:METHOD)% trailing %",
			wantError: true,
			errMsg:    "malformed command operator",
		},
		{
			name:      "invalid leading percent",
			value:     "% leading %REQ(:METHOD)%",
			wantError: true,
			errMsg:    "malformed command operator",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEnvoyFormatString(tt.value)

			if tt.wantError {
				if err == nil {
					t.Errorf("ValidateEnvoyFormatString() expected error but got nil")
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidateEnvoyFormatString() error = %v, want error containing %q", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateEnvoyFormatString() unexpected error = %v", err)
				}
			}
		})
	}
}
