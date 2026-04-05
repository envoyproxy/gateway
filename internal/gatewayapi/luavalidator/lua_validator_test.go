// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package luavalidator

import (
	"strings"
	"testing"

	"k8s.io/utils/ptr"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

func Test_BasicValidation(t *testing.T) {
	type testCase struct {
		name                 string
		code                 string
		proxy                *egv1a1.EnvoyProxy
		expectedErrSubstring string
	}
	tests := []testCase{
		{
			name:                 "empty body",
			code:                 "",
			expectedErrSubstring: "expected one of envoy_on_request() or envoy_on_response() to be defined",
		},
		{
			name: "logInfo: envoy_on_response",
			code: `function envoy_on_response(response_handle)
                     response_handle:logInfo("This log should not be printed.")
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
                     local headers, body = request_handle:httpPostCall(
                     "lua_cluster",
                     {
                         [":method"] = "POST",
                         [":path"] = "/",
                         [":authority"] = "lua_cluster"
                     },
                     "hello world",
                     5000)
                   end`,
			expectedErrSubstring: "attempt to call a non-function object",
		},
		{
			name: "stream:bodyChunks",
			code: `function envoy_on_response(response_handle)
                     response_handle:headers():replace("content-type", "text/html")
                     local last
                     for chunk in response_handle:bodyChunks() do
                         chunk:setBytes("")
                         last = chunk
                     end
                     last:setBytes("<html><b>Not Found<b></html>")
                   end`,
			expectedErrSubstring: "",
		},
		{
			name: "stream:body:getBytes",
			code: `function envoy_on_request(request_handle)
                     local body = request_handle:body(true):getBytes(0, request_handle:body():length())
                     request_handle:logErr("Request body: " .. body)
                   end`,
			expectedErrSubstring: "",
		},
		{
			name: "stream:bodyChunks:getBytes",
			code: `function envoy_on_request(request_handle)
                     for chunk in request_handle:bodyChunks() do
                         local bytes = chunk:getBytes(0, chunk:length())
                         request_handle:logErr("Chunk bytes: " .. bytes)
                     end
                   end`,
			expectedErrSubstring: "",
		},
		{
			name: "unsupported api - strict mode",
			code: `function envoy_on_request(request_handle)
                    request_handle:unknownApi()
                  end`,
			expectedErrSubstring: "attempt to call a non-function object",
		},
		{
			name: "unsupported api - insecure syntax mode allows",
			code: `function envoy_on_request(request_handle)
                    request_handle:unknownApi()
                  end`,
			proxy: &egv1a1.EnvoyProxy{
				Spec: egv1a1.EnvoyProxySpec{
					LuaValidation: ptr.To(egv1a1.LuaValidationInsecureSyntax),
				},
			},
			expectedErrSubstring: "",
		},
		{
			name: "invalid syntax - insecure syntax mode catches",
			code: `function envoy_on_response(response_handle)
                    response_handle:headers():replace("content-type", "text/html")
                    local last
                    for chunk in response_handle:bodyChunks() do
                        chunk:setBytes("")
                        last = chunk
                    last:setBytes("<html><b>Not Found<b></html>")
                  end`,
			proxy: &egv1a1.EnvoyProxy{
				Spec: egv1a1.EnvoyProxySpec{
					LuaValidation: ptr.To(egv1a1.LuaValidationInsecureSyntax),
				},
			},
			expectedErrSubstring: "<string> at EOF:   syntax error",
		},
		{
			name: "invalid syntax - disabled mode allows",
			code: `function envoy_on_response(response_handle)
                     response_handle:headers():replace("content-type", "text/html")
                     local last
                     for chunk in response_handle:bodyChunks() do
                         chunk:setBytes("")
                         last = chunk
                     last:setBytes("<html><b>Not Found<b></html>")
                   end`,
			proxy: &egv1a1.EnvoyProxy{
				Spec: egv1a1.EnvoyProxySpec{
					LuaValidation: ptr.To(egv1a1.LuaValidationDisabled),
				},
			},
			expectedErrSubstring: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewLuaValidator(tt.code, tt.proxy)
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

func Test_block_or_sanitize_io(t *testing.T) {
	type testCase struct {
		name                 string
		code                 string
		expectedErrSubstring string
	}
	tests := []testCase{
		// io.open tests
		{
			name: "io.open critical path /certs",
			code: `function envoy_on_response(response_handle)
                     local file = io.open("/certs/tls.crt", "w")
                     if file then file:close() end
                   end`,
			expectedErrSubstring: "critical path",
		},
		{
			name: "io.open non-critical path /tmp",
			code: `function envoy_on_response(response_handle)
                     local file = io.open("/tmp/tls.crt", "w")
                     if file then
                       file:write("test")
                       file:close()
                     end
                   end`,
			expectedErrSubstring: "",
		},
		{
			name: "io.open /etc/passwd",
			code: `function envoy_on_response(response_handle)
                     local file = io.open("/etc/passwd", "r")
                     if file then file:close() end
                   end`,
			expectedErrSubstring: "critical path",
		},
		{
			name: "io.open /proc/self/environ",
			code: `function envoy_on_response(response_handle)
                     local file = io.open("/proc/self/environ", "r")
                     if file then file:close() end
                   end`,
			expectedErrSubstring: "critical path",
		},
		{
			name: "io.open /sys/kernel",
			code: `function envoy_on_response(response_handle)
                     local file = io.open("/sys/kernel", "r")
                     if file then file:close() end
                   end`,
			expectedErrSubstring: "critical path",
		},
		{
			name: "io.open /var/run/secrets/token",
			code: `function envoy_on_response(response_handle)
                     local file = io.open("/var/run/secrets/token", "r")
                     if file then file:close() end
                   end`,
			expectedErrSubstring: "critical path",
		},
		{
			name: "io.open relative path etc/passwd",
			code: `function envoy_on_response(response_handle)
                     local file = io.open("etc/passwd", "r")
                     if file then file:close() end
                   end`,
			expectedErrSubstring: "critical path",
		},
		{
			name: "io.open path traversal /tmp/../etc/passwd",
			code: `function envoy_on_response(response_handle)
                     local file = io.open("/tmp/../etc/passwd", "r")
                     if file then file:close() end
                   end`,
			expectedErrSubstring: "path traversals",
		},
		{
			name: "io.open path traversal ../etc/passwd",
			code: `function envoy_on_response(response_handle)
                     local file = io.open("../etc/passwd", "r")
                     if file then file:close() end
                   end`,
			expectedErrSubstring: "path traversals",
		},
		{
			name: "io.open relative path certs/tls.crt",
			code: `function envoy_on_response(response_handle)
                     local file = io.open("certs/tls.crt", "r")
                     if file then file:close() end
                   end`,
			expectedErrSubstring: "critical path",
		},
		{
			name: "io.open relative path var/run/secrets/token",
			code: `function envoy_on_response(response_handle)
                     local file = io.open("var/run/secrets/token", "r")
                     if file then file:close() end
                   end`,
			expectedErrSubstring: "critical path",
		},
		{
			name: "io.open with backslash etc\\passwd",
			code: `function envoy_on_response(response_handle)
                     local file = io.open("etc\\passwd", "r")
                     if file then file:close() end
                   end`,
			expectedErrSubstring: "critical path",
		},
		{
			name: "io.open path traversal with backslash",
			code: `function envoy_on_response(response_handle)
                     local file = io.open("..\\etc\\passwd", "r")
                     if file then file:close() end
                   end`,
			expectedErrSubstring: "path traversals",
		},
		{
			name: "io.open with string concatenation",
			code: `function envoy_on_response(response_handle)
                     local path = "/" .. "etc" .. "/" .. "passwd"
                     local file = io.open(path, "r")
                     if file then file:close() end
                   end`,
			expectedErrSubstring: "critical path",
		},
		{
			name: "io.open with trailing slash /certs/",
			code: `function envoy_on_response(response_handle)
                     local file = io.open("/certs/", "r")
                     if file then file:close() end
                   end`,
			expectedErrSubstring: "critical path",
		},
		// io.input tests
		{
			name: "io.input critical path /certs",
			code: `function envoy_on_response(response_handle)
                     io.input("/certs/tls.crt")
                   end`,
			expectedErrSubstring: "critical path",
		},
		{
			name: "io.input non-critical path /tmp",
			code: `function envoy_on_response(response_handle)
                     local file = io.open("/tmp/tls.crt", "w")
                     if file then
                       file:write("test content")
                       file:close()
                     end
                     io.input("/tmp/tls.crt")
                   end`,
			expectedErrSubstring: "",
		},
		{
			name: "io.input /etc/passwd",
			code: `function envoy_on_response(response_handle)
                     io.input("/etc/passwd")
                   end`,
			expectedErrSubstring: "critical path",
		},
		// io.output tests
		{
			name: "io.output critical path /certs",
			code: `function envoy_on_response(response_handle)
                     io.output("/certs/tls.crt")
                   end`,
			expectedErrSubstring: "critical path",
		},
		{
			name: "io.output non-critical path /tmp",
			code: `function envoy_on_response(response_handle)
                     io.output("/tmp/tls.crt")
                   end`,
			expectedErrSubstring: "",
		},
		// io.lines tests
		{
			name: "io.lines critical path /certs",
			code: `function envoy_on_response(response_handle)
                     for line in io.lines("/certs/tls.crt") do
                       response_handle:logInfo(line)
                     end
                   end`,
			expectedErrSubstring: "critical path",
		},
		{
			name: "io.lines non-critical path /tmp",
			code: `function envoy_on_response(response_handle)
                     for line in io.lines("/tmp/tls.crt") do
                       response_handle:logInfo(line)
                     end
                   end`,
			expectedErrSubstring: "",
		},
		{
			name: "io.lines /etc/passwd",
			code: `function envoy_on_response(response_handle)
                     for line in io.lines("/etc/passwd") do
                       response_handle:logInfo(line)
                     end
                   end`,
			expectedErrSubstring: "critical path",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewLuaValidator(tt.code, nil)
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

func Test_block_or_sanitize_os(t *testing.T) {
	type testCase struct {
		name                 string
		code                 string
		expectedErrSubstring string
	}
	tests := []testCase{
		// os.remove tests
		{
			name: "os.remove critical path /certs",
			code: `function envoy_on_response(response_handle)
                     os.remove("/certs/tls.crt")
                   end`,
			expectedErrSubstring: "critical path",
		},
		{
			name: "os.remove non-critical path /tmp",
			code: `function envoy_on_response(response_handle)
                     os.remove("/tmp/tls.crt")
                   end`,
			expectedErrSubstring: "",
		},
		// os.rename tests
		{
			name: "os.rename critical path /certs",
			code: `function envoy_on_response(response_handle)
                     os.rename("/certs/tls.crt", "/certs/tls.crt.bak")
                   end`,
			expectedErrSubstring: "critical path",
		},
		{
			name: "os.rename non-critical path /tmp",
			code: `function envoy_on_response(response_handle)
                     os.rename("/tmp/tls.crt", "/tmp/tls.crt.bak")
                   end`,
			expectedErrSubstring: "",
		},
		{
			name: "os.rename critical source",
			code: `function envoy_on_response(response_handle)
                     os.rename("/certs/tls.crt", "/tmp/tls.crt.bak")
                   end`,
			expectedErrSubstring: "critical path",
		},
		{
			name: "os.rename critical destination",
			code: `function envoy_on_response(response_handle)
                     os.rename("/tmp/tls.crt", "/certs/tls.crt.bak")
                   end`,
			expectedErrSubstring: "critical path",
		},
		{
			name: "os.rename to critical path /etc",
			code: `function envoy_on_response(response_handle)
                     os.rename("/tmp/file", "/certs/file")
                   end`,
			expectedErrSubstring: "critical path",
		},
		{
			name: "os.rename from critical path /etc",
			code: `function envoy_on_response(response_handle)
                     os.rename("/certs/file", "/tmp/file")
                   end`,
			expectedErrSubstring: "critical path",
		},
		// os.getenv tests
		{
			name: "os.getenv critical env var PWD",
			code: `function envoy_on_response(response_handle)
                     local pwd = os.getenv("PWD")
                   end`,
			expectedErrSubstring: "critical environment variable",
		},
		{
			name: "os.getenv critical env var pwd (lowercase)",
			code: `function envoy_on_response(response_handle)
                     local pwd = os.getenv("pwd")
                   end`,
			expectedErrSubstring: "critical environment variable",
		},
		// os.setenv tests
		{
			name: "os.setenv non-critical env var allowed",
			code: `function envoy_on_response(response_handle)
                     os.setenv("TEST", "value")
                   end`,
			expectedErrSubstring: "",
		},
		{
			name: "os.setenv critical env var PWD",
			code: `function envoy_on_response(response_handle)
                     os.setenv("PWD", "/etc")
                   end`,
			expectedErrSubstring: "setting critical environment variable",
		},
		{
			name: "os.setenv critical env var pwd (lowercase)",
			code: `function envoy_on_response(response_handle)
                     os.setenv("pwd", "/etc")
                   end`,
			expectedErrSubstring: "setting critical environment variable",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewLuaValidator(tt.code, nil)
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

// Test_Blocked ensures all these functions are completely blocked
func Test_Blocked(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{
			name: "io.popen",
			code: `function envoy_on_response(response_handle)
                     io.popen("cat /etc/passwd")
                   end`,
		},
		{
			name: "os.execute",
			code: `function envoy_on_response(response_handle)
                     os.execute("cat /etc/passwd")
                   end`,
		},
		{
			name: "os.exit",
			code: `function envoy_on_response(response_handle)
                     os.exit(0)
                   end`,
		},
		{
			name: "require",
			code: `function envoy_on_response(response_handle)
                     local mod = require("mymodule")
                   end`,
		},
		{
			name: "loadfile",
			code: `function envoy_on_response(response_handle)
                     local func = loadfile("/tmp/module.lua")
                   end`,
		},
		{
			name: "dofile",
			code: `function envoy_on_response(response_handle)
                     dofile("/tmp/module.lua")
                   end`,
		},
		{
			name: "package.loadlib",
			code: `function envoy_on_response(response_handle)
                     package.loadlib("/tmp/lib.so", "init")
                   end`,
		},
		{
			name: "debug library",
			code: `function envoy_on_response(response_handle)
                     debug.getinfo(1)
                   end`,
		},
		{
			name: "getmetatable",
			code: `function envoy_on_response(response_handle)
                     local mt = getmetatable(io)
                   end`,
		},
		{
			name: "setmetatable",
			code: `function envoy_on_response(response_handle)
                     setmetatable({}, {})
                   end`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewLuaValidator(tt.code, nil)
			err := l.Validate()
			if err == nil {
				t.Errorf("Expected error for blocked function %s, got no error", tt.name)
			} else if !strings.Contains(err.Error(), "attempt to") && !strings.Contains(err.Error(), "blocked function") {
				t.Errorf("Expected 'attempt to' or 'blocked function' error for %s, got: %v", tt.name, err)
			}
		})
	}
}

// Test_UnsafePackageReference ensures attempts to reference unsafe packages are blocked
func Test_UnsafePackageReference(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{
			name: "cannot access _unsafe_io_open (closure protected)",
			code: `function envoy_on_response(response_handle)
                     local file = _unsafe_io_open("/etc/passwd", "r")
                   end`,
		},
		{
			name: "cannot access _unsafe_io_input (closure protected)",
			code: `function envoy_on_response(response_handle)
                     _unsafe_io_input("/etc/passwd")
                   end`,
		},
		{
			name: "cannot access _unsafe_io_output (closure protected)",
			code: `function envoy_on_response(response_handle)
                     _unsafe_io_output("/etc/passwd")
                   end`,
		},
		{
			name: "cannot access _unsafe_io_lines (closure protected)",
			code: `function envoy_on_response(response_handle)
                     for line in _unsafe_io_lines("/etc/passwd") do
                       response_handle:logInfo(line)
                     end
                   end`,
		},
		{
			name: "cannot access _unsafe_os_remove (closure protected)",
			code: `function envoy_on_response(response_handle)
                     _unsafe_os_remove("/tmp/file")
                   end`,
		},
		{
			name: "cannot access _unsafe_os_rename (closure protected)",
			code: `function envoy_on_response(response_handle)
                     _unsafe_os_rename("/tmp/old", "/tmp/new")
                   end`,
		},
		{
			name: "cannot access _unsafe_os_getenv (closure protected)",
			code: `function envoy_on_response(response_handle)
                     local pwd = _unsafe_os_getenv("PWD")
                   end`,
		},
		{
			name: "cannot access _unsafe_os_setenv (closure protected)",
			code: `function envoy_on_response(response_handle)
                     _unsafe_os_setenv("TEST", "value")
                   end`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewLuaValidator(tt.code, nil)
			err := l.Validate()
			if err == nil {
				t.Errorf("Expected error for closure-protected function %s, got no error", tt.name)
			}
			// The error should be about undefined global variable or attempt to call nil
			// since the closure-scoped variables are not accessible to user code
			if !strings.Contains(err.Error(), "not declared") &&
				!strings.Contains(err.Error(), "attempt to call") &&
				!strings.Contains(err.Error(), "attempt to index") {
				t.Logf("Got expected error for %s: %v", tt.name, err)
			}
		})
	}
}

// Test_ResourceLimitsAndPanicRecovery ensures that malicious code cannot crash the controller
func Test_ResourceLimitsAndPanicRecovery(t *testing.T) {
	tests := []struct {
		name                 string
		code                 string
		expectedErrSubstring string
	}{
		{
			name: "infinite loop should timeout",
			code: `function envoy_on_request(request_handle)
                     while true do
                       -- infinite loop
                     end
                   end`,
			expectedErrSubstring: "failed to validate with envoy_on_request: lua execution timeout: code took longer than 5s",
		},
		{
			name: "deep recursion should be limited by call stack",
			code: `function recurse(n)
                     recurse(n + 1)
                     recurse(n + 2)
                   end
                   function envoy_on_request(request_handle)
                     recurse(1)
                   end`,
			expectedErrSubstring: "stack overflow",
		},
		{
			name: "calling Go function that might panic should not crash",
			code: `function envoy_on_request(request_handle)
                     -- Trying to access fields on nil should error, not panic
                     local x = nil
                     local y = x.field
                   end`,
			expectedErrSubstring: "", // Should handle gracefully
		},
		{
			name: "string with null bytes should not crash",
			code: `function envoy_on_request(request_handle)
                     local s = "hello\000world"
                     request_handle:logInfo(s)
                   end`,
			expectedErrSubstring: "", // Should handle gracefully
		},
		{
			name: "coroutine errors should not crash",
			code: `function envoy_on_request(request_handle)
                     local co = coroutine.create(function()
                       error("coroutine error")
                     end)
                     coroutine.resume(co)
                   end`,
			expectedErrSubstring: "", // Should handle gracefully
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewLuaValidator(tt.code, nil)
			err := l.Validate()

			// The important thing is that it doesn't crash/panic the test
			if tt.expectedErrSubstring != "" {
				if err == nil {
					t.Errorf("Expected error containing '%s', got no error", tt.expectedErrSubstring)
				} else if !strings.Contains(err.Error(), tt.expectedErrSubstring) {
					t.Errorf("Expected substring '%s' in error, got: %v", tt.expectedErrSubstring, err)
				}
			}
			// If no error expected, it's fine if it succeeds or fails gracefully
		})
	}
}
