// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package luavalidator

import (
	_ "embed"
	"fmt"
	"strings"

	lua "github.com/yuin/gopher-lua"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

const (
	envoyOnRequestFunctionName  = "envoy_on_request"
	envoyOnResponseFunctionName = "envoy_on_response"
)

// mockData contains mocks of Envoy supported APIs for Lua filters.
// Refer: https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/lua_filter#stream-handle-api
//
//go:embed mocks.lua
var mockData string

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
		if err := l.validate(mockData + "\n" + l.code + "\n" + envoyOnRequestFunctionName + "(StreamHandle)"); err != nil {
			return fmt.Errorf("failed to validate with %s: %w", envoyOnRequestFunctionName, err)
		}
	}
	if strings.Contains(l.code, envoyOnResponseFunctionName) {
		if err := l.validate(mockData + "\n" + l.code + "\n" + envoyOnResponseFunctionName + "(StreamHandle)"); err != nil {
			return fmt.Errorf("failed to validate with %s: %w", envoyOnResponseFunctionName, err)
		}
	}
	return nil
}

// validate runs the validation on given code
func (l *LuaValidator) validate(code string) error {
	switch l.getLuaValidation() {
	case egv1a1.LuaValidationSyntax:
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

// newLuaState creates a new Lua state with global settings applied
func (l *LuaValidator) newLuaState() *lua.LState {
	L := lua.NewState()
	// Suppress all print statements
	L.SetGlobal("print", L.NewFunction(func(L *lua.LState) int {
		return 0
	}))
	return L
}

// runLua interprets and runs the provided Lua code in runtime using gopher-lua
// Refer: https://github.com/yuin/gopher-lua?tab=readme-ov-file#differences-between-lua-and-gopherlua
func (l *LuaValidator) runLua(code string) error {
	L := l.newLuaState()
	defer L.Close()
	if err := L.DoString(code); err != nil {
		return err
	}
	return nil
}

// loadLua loads the Lua code into the Lua state, does not run it
// This is used to check for syntax errors in the Lua code
func (l *LuaValidator) loadLua(code string) error {
	L := l.newLuaState()
	defer L.Close()
	if _, err := L.LoadString(code); err != nil {
		return err
	}
	return nil
}
