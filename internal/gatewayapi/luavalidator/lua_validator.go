// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package luavalidator

import (
	"context"
	_ "embed"
	"fmt"
	"strings"
	"time"

	lua "github.com/yuin/gopher-lua"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

const (
	envoyOnRequestFunctionName  = "envoy_on_request"
	envoyOnResponseFunctionName = "envoy_on_response"
	luaExecutionTimeout         = 5 * time.Second
)

// mockData contains mocks of Envoy supported APIs for Lua filters.
// Refer: https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/lua_filter#stream-handle-api
//
//go:embed mocks.lua
var mockData string

// securityData contains Lua security wrappers that restrict access to sensitive filesystem paths and
// critical environment variables during Lua code validation in the gateway controller.
//
// TODO: Create a configurable set of filesystem paths and environment variables to check for in the proxy apart from default configured here.
//
//go:embed security.lua
var securityData string

// LuaValidator validates user provided Lua for compatibility with Envoy supported Lua HTTP filter
// Validation strictness is controlled by the validation field
type LuaValidator struct {
	code       string
	envoyProxy *egv1a1.EnvoyProxy
}

// NewLuaValidator returns a LuaValidator for user provided Lua code
func NewLuaValidator(code string, envoyProxy *egv1a1.EnvoyProxy) *LuaValidator {
	return &LuaValidator{
		code:       code,
		envoyProxy: envoyProxy,
	}
}

// Validate runs all validations for the LuaValidator
func (l *LuaValidator) Validate() error {
	if !strings.Contains(l.code, envoyOnRequestFunctionName) && !strings.Contains(l.code, envoyOnResponseFunctionName) {
		return fmt.Errorf("expected one of %s() or %s() to be defined", envoyOnRequestFunctionName, envoyOnResponseFunctionName)
	}
	if strings.Contains(l.code, envoyOnRequestFunctionName) {
		if err := l.validate(l.code + "\n" + envoyOnRequestFunctionName + "(StreamHandle)"); err != nil {
			return fmt.Errorf("failed to validate with %s: %w", envoyOnRequestFunctionName, err)
		}
	}
	if strings.Contains(l.code, envoyOnResponseFunctionName) {
		if err := l.validate(l.code + "\n" + envoyOnResponseFunctionName + "(StreamHandle)"); err != nil {
			return fmt.Errorf("failed to validate with %s: %w", envoyOnResponseFunctionName, err)
		}
	}
	return nil
}

// validate runs the validation on given code
func (l *LuaValidator) validate(code string) error {
	switch l.getLuaValidation() {
	case egv1a1.LuaValidationInsecureSyntax:
		return l.loadLua(code)
	case egv1a1.LuaValidationDisabled:
		return nil
	default:
		return l.runLua(code)
	}
}

// getLuaValidation returns the Lua validation level, defaulting to strict if not configured
func (l *LuaValidator) getLuaValidation() egv1a1.LuaValidation {
	if l.envoyProxy != nil && l.envoyProxy.Spec.LuaValidation != nil {
		return *l.envoyProxy.Spec.LuaValidation
	}
	return egv1a1.LuaValidationStrict
}

// newLuaState creates a new Lua state with global settings and resource limits applied
// Returns the Lua state and a cancel function that must be called when done
func (l *LuaValidator) newLuaState() (*lua.LState, context.CancelFunc) {
	// Configure Lua VM with resource limits to prevent DoS
	L := lua.NewState(lua.Options{
		CallStackSize:       256,   // Default call stack depth
		RegistrySize:        5120,  // Default registry size (256 * 20)
		RegistryMaxSize:     5120,  // Disable registry growth for security
		RegistryGrowStep:    32,    // Default registry growth step (not used since max = initial)
		SkipOpenLibs:        false, // We need standard libraries (filtered by security.lua)
		IncludeGoStackTrace: false, // Don't leak Go stack traces to user errors
	})

	// Create context with timeout to prevent infinite loops or long-running code
	ctx, cancel := context.WithTimeout(context.Background(), luaExecutionTimeout)
	L.SetContext(ctx)

	// Suppress all print statements
	L.SetGlobal("print", L.NewFunction(func(_ *lua.LState) int {
		return 0
	}))
	return L, cancel
}

// runLua interprets and runs the provided Lua code in runtime using gopher-lua
// Refer: https://github.com/yuin/gopher-lua?tab=readme-ov-file#differences-between-lua-and-gopherlua
func (l *LuaValidator) runLua(code string) error {
	L, cancel := l.newLuaState()
	defer cancel()
	defer L.Close()

	// Execute mocks first (trusted code, needs setmetatable, defines StreamHandle, etc.)
	_ = L.DoString(mockData)
	// Execute Lua security wrappers (trusted code) to protect the gateway controller
	// See security advisory: https://github.com/envoyproxy/gateway/security/advisories/GHSA-xrwg-mqj6-6m22
	_ = L.DoString(securityData)

	// Execute user-provided code with panic recovery to prevent controller crashes
	// Although gopher-lua returns errors if it internally panics,
	// this is a defensive measure to prevent the controller from crashing.
	var err error
	func() {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("lua execution panic: %v", r)
			}
		}()
		err = L.DoString(code)
	}()

	if err != nil {
		// Check if timeout occurred
		if L.Context().Err() == context.DeadlineExceeded {
			return fmt.Errorf("lua execution timeout: code took longer than %v", luaExecutionTimeout)
		}
		return err
	}
	return nil
}

// loadLua loads the Lua code into the Lua state, does not run it
// This is used to check for syntax errors in the Lua code
func (l *LuaValidator) loadLua(code string) error {
	L, cancel := l.newLuaState()
	defer cancel()
	defer L.Close()
	if _, err := L.LoadString(code); err != nil {
		return err
	}
	return nil
}
