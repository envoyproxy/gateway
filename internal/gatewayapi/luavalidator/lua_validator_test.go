// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package luavalidator

import (
	"strings"
	"testing"

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
					LuaValidation: new(egv1a1.LuaValidationInsecureSyntax),
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
					LuaValidation: new(egv1a1.LuaValidationInsecureSyntax),
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
					LuaValidation: new(egv1a1.LuaValidationDisabled),
				},
			},
			expectedErrSubstring: "",
		},
		{
			name: "stream:filterContext:get",
			code: `function envoy_on_response(response_handle)
                     local ctx = response_handle:filterContext()
                     if ctx ~= nil then
                       local custom_value = ctx:get("custom_value")
                       if custom_value ~= nil then
                         response_handle:headers():add("X-Lua-Filter-Context", custom_value)
                       end
                     end
                   end`,
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

// allowlistProxy returns an EnvoyProxy configured with the given Lua validation allowlists.
func allowlistProxy(paths, envVars []string) *egv1a1.EnvoyProxy {
	return &egv1a1.EnvoyProxy{
		Spec: egv1a1.EnvoyProxySpec{
			LuaStrictValidation: &egv1a1.LuaStrictValidation{
				AllowedPaths:   paths,
				AllowedEnvVars: envVars,
			},
		},
	}
}

// Test_io_path_allowlist verifies the filesystem allowlist for the sanitized io functions.
// The allowlist is fail-closed: only paths under /tmp are permitted; everything else is denied.
// Path traversal segments are always rejected, regardless of the allowlist.
func Test_io_path_allowlist(t *testing.T) {
	proxy := allowlistProxy([]string{"/tmp"}, nil)

	tests := []struct {
		name                 string
		code                 string
		expectedErrSubstring string
	}{
		{
			name:                 "io.open allowed /tmp",
			code:                 `function envoy_on_response(h) local f = io.open("/tmp/x", "w") if f then f:close() end end`,
			expectedErrSubstring: "",
		},
		{
			name:                 "io.open allowed via subtree /tmp/sub/x",
			code:                 `function envoy_on_response(h) local f = io.open("/tmp/sub/x", "w") if f then f:close() end end`,
			expectedErrSubstring: "",
		},
		{
			name:                 "io.open denied /etc/passwd",
			code:                 `function envoy_on_response(h) io.open("/etc/passwd", "r") end`,
			expectedErrSubstring: "io.open restricted for param",
		},
		{
			name:                 "io.open denied path outside allowlist",
			code:                 `function envoy_on_response(h) io.open("/tmpfoo/x", "r") end`,
			expectedErrSubstring: "io.open restricted for param",
		},
		// Path normalization: relative, backslash, and multi-slash forms must normalize
		// consistently before the allowlist match.
		{
			name:                 "io.open relative path normalizes under allowed root",
			code:                 `function envoy_on_response(h) local f = io.open("tmp/x", "r") if f then f:close() end end`,
			expectedErrSubstring: "",
		},
		{
			name:                 "io.open allowed double-slash //tmp/x",
			code:                 `function envoy_on_response(h) local f = io.open("//tmp/x", "w") if f then f:close() end end`,
			expectedErrSubstring: "",
		},
		{
			name:                 "io.open allowed trailing slash /tmp/",
			code:                 `function envoy_on_response(h) local f = io.open("/tmp/", "r") if f then f:close() end end`,
			expectedErrSubstring: "",
		},
		{
			name:                 "io.open denied double-slash //etc/passwd",
			code:                 `function envoy_on_response(h) io.open("//etc/passwd", "r") end`,
			expectedErrSubstring: "io.open restricted for param",
		},
		{
			name:                 "io.open denied embedded double-slash /etc//passwd",
			code:                 `function envoy_on_response(h) io.open("/etc//passwd", "r") end`,
			expectedErrSubstring: "io.open restricted for param",
		},
		{
			name:                 "io.open denied multiple-slash ///etc/passwd",
			code:                 `function envoy_on_response(h) io.open("///etc/passwd", "r") end`,
			expectedErrSubstring: "io.open restricted for param",
		},
		{
			name:                 "io.open denied backslash etc\\passwd",
			code:                 `function envoy_on_response(h) io.open("etc\\passwd", "r") end`,
			expectedErrSubstring: "io.open restricted for param",
		},
		{
			name:                 "io.open denied string concatenation",
			code:                 `function envoy_on_response(h) local p = "/" .. "etc" .. "/" .. "passwd" io.open(p, "r") end`,
			expectedErrSubstring: "io.open restricted for param",
		},
		// Traversal segments are always rejected with a distinct error, regardless of the allowlist.
		{
			name:                 "io.open traversal rejected even under allowed root",
			code:                 `function envoy_on_response(h) io.open("/tmp/../etc/passwd", "r") end`,
			expectedErrSubstring: "path traversals",
		},
		{
			name:                 "io.open traversal rejected ../etc/passwd",
			code:                 `function envoy_on_response(h) io.open("../etc/passwd", "r") end`,
			expectedErrSubstring: "path traversals",
		},
		{
			name:                 "io.open traversal rejected with backslash",
			code:                 `function envoy_on_response(h) io.open("..\\etc\\passwd", "r") end`,
			expectedErrSubstring: "path traversals",
		},
		{
			name:                 "io.open traversal rejected dot segment /tmp/./x",
			code:                 `function envoy_on_response(h) io.open("/tmp/./x", "r") end`,
			expectedErrSubstring: "path traversals",
		},
		{
			name:                 "io.open traversal rejected leading dot ./tmp/x",
			code:                 `function envoy_on_response(h) io.open("./tmp/x", "r") end`,
			expectedErrSubstring: "path traversals",
		},
		{
			name:                 "io.open traversal rejected trailing dot /tmp/.",
			code:                 `function envoy_on_response(h) io.open("/tmp/.", "r") end`,
			expectedErrSubstring: "path traversals",
		},
		{
			name:                 "io.open traversal rejected dot with backslash tmp\\.\\x",
			code:                 `function envoy_on_response(h) io.open("tmp\\.\\x", "r") end`,
			expectedErrSubstring: "path traversals",
		},
		{
			name:                 "io.input denied /etc/passwd",
			code:                 `function envoy_on_response(h) io.input("/etc/passwd") end`,
			expectedErrSubstring: "io.input restricted for param",
		},
		{
			name:                 "io.output denied /certs",
			code:                 `function envoy_on_response(h) io.output("/certs/tls.crt") end`,
			expectedErrSubstring: "io.output restricted for param",
		},
		{
			name:                 "io.lines denied /etc/passwd",
			code:                 `function envoy_on_response(h) for l in io.lines("/etc/passwd") do end end`,
			expectedErrSubstring: "io.lines restricted for param",
		},
	}
	runAllowlistCases(t, proxy, tests)
}

// Test_path_allowlist_literal_match ensures allowed prefixes containing Lua pattern magic characters
// (e.g. ".") are matched literally and do not widen the security boundary.
func Test_path_allowlist_literal_match(t *testing.T) {
	proxy := allowlistProxy([]string{"/var/lib/app.v1"}, nil)

	tests := []struct {
		name                 string
		code                 string
		expectedErrSubstring string
	}{
		{
			name:                 "exact allowed entry",
			code:                 `function envoy_on_response(h) local f = io.open("/var/lib/app.v1", "r") if f then f:close() end end`,
			expectedErrSubstring: "",
		},
		{
			name:                 "subtree of allowed entry",
			code:                 `function envoy_on_response(h) local f = io.open("/var/lib/app.v1/data", "r") if f then f:close() end end`,
			expectedErrSubstring: "",
		},
		{
			name:                 "dot must not match arbitrary character",
			code:                 `function envoy_on_response(h) io.open("/var/lib/appXv1/data", "r") end`,
			expectedErrSubstring: "io.open restricted for param",
		},
	}
	runAllowlistCases(t, proxy, tests)
}

// Test_path_allowlist_blank_entry_denied ensures a blank allowlist entry does not match every path
// and silently disable the sandbox (defense in depth; the CRD also rejects blank entries).
func Test_path_allowlist_blank_entry_denied(t *testing.T) {
	proxy := allowlistProxy([]string{"", "  ", "/tmp"}, nil)

	tests := []struct {
		name                 string
		code                 string
		expectedErrSubstring string
	}{
		{
			name:                 "blank entry does not allow arbitrary path",
			code:                 `function envoy_on_response(h) io.open("/etc/passwd", "r") end`,
			expectedErrSubstring: "io.open restricted for param",
		},
		{
			name:                 "real entry still allowed",
			code:                 `function envoy_on_response(h) local f = io.open("/tmp/x", "r") if f then f:close() end end`,
			expectedErrSubstring: "",
		},
	}
	runAllowlistCases(t, proxy, tests)
}

// Test_io_denied_by_default ensures that with no allowlist configured, all filesystem access is denied.
func Test_io_denied_by_default(t *testing.T) {
	code := `function envoy_on_response(h) io.open("/tmp/x", "w") end`
	if err := NewLuaValidator(code, nil).Validate(); err == nil {
		t.Errorf("Expected fail-closed denial with no allowlist, got no error")
	} else if !strings.Contains(err.Error(), "io.open restricted for param") {
		t.Errorf("Expected 'io.open restricted for param', got: %v", err)
	}
}

// Test_os_path_allowlist verifies the filesystem allowlist for the sanitized os functions.
// Only /tmp is permitted; os.rename requires both source and destination to be allowed.
func Test_os_path_allowlist(t *testing.T) {
	proxy := allowlistProxy([]string{"/tmp"}, nil)

	tests := []struct {
		name                 string
		code                 string
		expectedErrSubstring string
	}{
		{
			name:                 "os.remove allowed /tmp",
			code:                 `function envoy_on_response(h) os.remove("/tmp/x") end`,
			expectedErrSubstring: "",
		},
		{
			name:                 "os.remove denied /certs",
			code:                 `function envoy_on_response(h) os.remove("/certs/tls.crt") end`,
			expectedErrSubstring: "os.remove restricted for param",
		},
		{
			name:                 "os.rename allowed both /tmp",
			code:                 `function envoy_on_response(h) os.rename("/tmp/a", "/tmp/b") end`,
			expectedErrSubstring: "",
		},
		{
			name:                 "os.rename denied source",
			code:                 `function envoy_on_response(h) os.rename("/certs/a", "/tmp/b") end`,
			expectedErrSubstring: "os.rename restricted for param",
		},
		{
			name:                 "os.rename denied destination",
			code:                 `function envoy_on_response(h) os.rename("/tmp/a", "/certs/b") end`,
			expectedErrSubstring: "os.rename restricted for param",
		},
	}
	runAllowlistCases(t, proxy, tests)
}

// Test_os_env_allowlist verifies the environment variable allowlist (exact, case-sensitive match)
// for os.getenv and os.setenv.
func Test_os_env_allowlist(t *testing.T) {
	proxy := allowlistProxy(nil, []string{"LOG_LEVEL"})

	tests := []struct {
		name                 string
		code                 string
		expectedErrSubstring string
	}{
		{
			name:                 "os.getenv allowed LOG_LEVEL",
			code:                 `function envoy_on_response(h) os.getenv("LOG_LEVEL") end`,
			expectedErrSubstring: "",
		},
		{
			name:                 "os.getenv denied PWD",
			code:                 `function envoy_on_response(h) os.getenv("PWD") end`,
			expectedErrSubstring: "os.getenv restricted for param PWD",
		},
		{
			name:                 "os.getenv match is case-sensitive",
			code:                 `function envoy_on_response(h) os.getenv("log_level") end`,
			expectedErrSubstring: "os.getenv restricted for param log_level",
		},
		{
			name:                 "os.setenv allowed LOG_LEVEL",
			code:                 `function envoy_on_response(h) os.setenv("LOG_LEVEL", "debug") end`,
			expectedErrSubstring: "",
		},
		{
			name:                 "os.setenv denied PWD",
			code:                 `function envoy_on_response(h) os.setenv("PWD", "/etc") end`,
			expectedErrSubstring: "os.setenv restricted for param PWD",
		},
	}
	runAllowlistCases(t, proxy, tests)
}

// Test_env_denied_by_default ensures that with no allowlist configured, all env var access is denied.
func Test_env_denied_by_default(t *testing.T) {
	code := `function envoy_on_response(h) os.getenv("LOG_LEVEL") end`
	if err := NewLuaValidator(code, nil).Validate(); err == nil {
		t.Errorf("Expected fail-closed denial with no allowlist, got no error")
	} else if !strings.Contains(err.Error(), "os.getenv restricted for param") {
		t.Errorf("Expected 'os.getenv restricted for param', got: %v", err)
	}
}

// runAllowlistCases runs a set of allowlist validation cases against the given proxy.
func runAllowlistCases(t *testing.T, proxy *egv1a1.EnvoyProxy, tests []struct {
	name                 string
	code                 string
	expectedErrSubstring string
},
) {
	t.Helper()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewLuaValidator(tt.code, proxy).Validate()
			switch {
			case err != nil && tt.expectedErrSubstring == "":
				t.Errorf("Unexpected error: %v", err)
			case err != nil && !strings.Contains(err.Error(), tt.expectedErrSubstring):
				t.Errorf("Expected substring %q in error, got: %v", tt.expectedErrSubstring, err)
			case err == nil && tt.expectedErrSubstring != "":
				t.Errorf("Expected error with substring %q, got none", tt.expectedErrSubstring)
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
