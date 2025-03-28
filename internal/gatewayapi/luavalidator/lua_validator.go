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
)

// mockData contains mocks of Envoy supported APIs for Lua filters.
// Refer: https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/lua_filter#stream-handle-api
//
//go:embed mocks.lua
var mockData []byte

// LuaValidator validates user provided Lua for compatibility with Envoy supported Lua HTTP filter
type LuaValidator struct {
	code string
}

// NewLuaValidator returns a LuaValidator for user provided Lua code
func NewLuaValidator(code string) *LuaValidator {
	return &LuaValidator{
		code: code,
	}
}

// Validate runs all validations for the LuaValidator
func (l *LuaValidator) Validate() error {
	if !strings.Contains(l.code, "envoy_on_request") && !strings.Contains(l.code, "envoy_on_response") {
		return fmt.Errorf("expected one of envoy_on_request() or envoy_on_response() to be defined")
	}
	if strings.Contains(l.code, "envoy_on_request") {
		if err := l.runLua(string(mockData) + "\n" + l.code + "\nenvoy_on_request(StreamHandle)"); err != nil {
			return fmt.Errorf("failed to mock run envoy_on_request: %w", err)
		}
	}
	if strings.Contains(l.code, "envoy_on_response") {
		if err := l.runLua(string(mockData) + "\n" + l.code + "\nenvoy_on_response(StreamHandle)"); err != nil {
			return fmt.Errorf("failed to mock run envoy_on_response: %w", err)
		}
	}
	return nil
}

// runLua interprets and runs the provided Lua code in runtime using gopher-lua
// Refer: https://github.com/yuin/gopher-lua?tab=readme-ov-file#differences-between-lua-and-gopherlua
func (l *LuaValidator) runLua(code string) error {
	L := lua.NewState()
	defer L.Close()
	if err := L.DoString(code); err != nil {
		return err
	}
	return nil
}
