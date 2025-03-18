// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package luavalidator

import (
	"strings"
	"testing"
)

func Test_Validate(t *testing.T) {
	type args struct {
		name                 string
		code                 string
		expectedErrSubstring string
	}
	tests := []args{
		{
			name:                 "empty body",
			code:                 "",
			expectedErrSubstring: "expected one of envoy_on_request() or envoy_on_response() to be defined",
		},
		{
			name: "logInfo: envoy_on_response",
			code: `function envoy_on_response(response_handle)
                     response_handle:logInfo("Goodbye.")
                   end`,
			expectedErrSubstring: "",
		},
		{
			name: "logInfo: envoy_on_request",
			code: `function envoy_on_request(request_handle)
                     request_handle:logInfo("Goodbye.")
                   end`,
			expectedErrSubstring: "",
		},
		{
			name: "stream:headers:Get",
			code: `function envoy_on_request(request_handle)
                     request_handle:headers():get("foo")
                   end`,
			expectedErrSubstring: "",
		},
		{
			name: "stream:connection:ssl:expirationPeerCertificate",
			code: `function envoy_on_request(request_handle)
                     request_handle:connection():ssl():expirationPeerCertificate()
                   end`,
			expectedErrSubstring: "",
		},
		{
			name: "stream:metadata:pairs",
			code: `function envoy_on_request(request_handle)
                     for key, value in pairs(request_handle:metadata()) do
                       print(key, value)
					 end
                   end`,
			expectedErrSubstring: "",
		},
		{
			name: "stream:httpCall",
			code: `function envoy_on_request(request_handle)
			  -- Make an HTTP call.
			  local headers, body = request_handle:httpCall(
			  "lua_cluster",
			  {
				[":method"] = "POST",
				[":path"] = "/",
				[":authority"] = "lua_cluster",
				["set-cookie"] = { "lang=lua; Path=/", "type=binding; Path=/" }
			  },
			  "hello world",
			  5000)
			
			  -- Response directly and set a header from the HTTP call. No further filter iteration
			  -- occurs.
			  request_handle:respond(
				{[":status"] = "403",
				 ["upstream_foo"] = headers["foo"]},
				"nope")
			end`,
			expectedErrSubstring: "",
		},
		{
			name: "stream:httpPostCall unsupported api",
			code: `function envoy_on_request(request_handle)
			  -- Make an HTTP call.
			  local headers, body = request_handle:httpPostCall(
			  "lua_cluster",
			  {
				[":method"] = "POST",
				[":path"] = "/",
				[":authority"] = "lua_cluster",
				["set-cookie"] = { "lang=lua; Path=/", "type=binding; Path=/" }
			  },
			  "hello world",
			  5000)
			
			  -- Response directly and set a header from the HTTP call. No further filter iteration
			  -- occurs.
			  request_handle:respond(
				{[":status"] = "403",
				 ["upstream_foo"] = headers["foo"]},
				"nope")
			end`,
			expectedErrSubstring: "attempt to call a non-function object",
		},
		{
			name: "stream:bodyChunks",
			code: `function envoy_on_response(response_handle)
			  -- Sets the content-type.
			  response_handle:headers():replace("content-type", "text/html")
			  local last
			  for chunk in response_handle:bodyChunks() do
				-- Clears each received chunk.
				chunk:setBytes("")
				last = chunk
			  end
			
			  last:setBytes("<html><b>Not Found<b></html>")
            end`,
			expectedErrSubstring: "",
		},
		{
			name: "unsupported api",
			code: `function envoy_on_request(request_handle)
                     request_handle:unknownApi()
                   end`,
			expectedErrSubstring: "attempt to call a non-function object",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewLuaValidator(tt.code)
			if err := l.Validate(); err != nil && tt.expectedErrSubstring == "" {
				t.Errorf("Unexpected error: %v", err)
			} else if err != nil && !strings.Contains(err.Error(), tt.expectedErrSubstring) {
				t.Errorf("Expected substring in error: %v, got error: %v", tt.expectedErrSubstring, err)
			} else if err == nil && tt.expectedErrSubstring != "" {
				t.Errorf("Expected error with substring: %v", tt.expectedErrSubstring)
			}
		})
	}
}
