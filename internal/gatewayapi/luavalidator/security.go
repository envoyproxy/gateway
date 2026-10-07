// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package luavalidator

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"

	lua "github.com/yuin/gopher-lua"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

var defaultDeniedPaths = []string{
	"/etc",
	"/proc",
	"/sys",
	"/certs",
	"/var/run/secrets",
	// /var/run is a symlink to /run on Debian-derived (distroless) images.
	"/run/secrets",
}

type securityValidator struct {
	allowedPaths   []string
	deniedPaths    []string
	allowedEnvVars map[string]struct{}
}

func newSecurityValidator(envoyProxy *egv1a1.EnvoyProxy) *securityValidator {
	validator := &securityValidator{
		deniedPaths: resolvePaths(defaultDeniedPaths),
	}

	if envoyProxy == nil || envoyProxy.Spec.Lua == nil || envoyProxy.Spec.Lua.StrictValidation == nil {
		return validator
	}

	strictValidation := envoyProxy.Spec.Lua.StrictValidation
	validator.allowedPaths = resolvePaths(strictValidation.AllowedPaths)
	validator.allowedEnvVars = make(map[string]struct{}, len(strictValidation.AllowedEnvVars))
	for _, envVar := range strictValidation.AllowedEnvVars {
		validator.allowedEnvVars[envVar] = struct{}{}
	}

	return validator
}

// install exposes the Go validators to the trusted Lua wrappers. security.lua captures the
// functions as locals and clears the globals before user-provided code executes.
func (v *securityValidator) install(state *lua.LState) {
	state.SetGlobal("__lua_validate_path", state.NewFunction(v.luaValidatePath))
	state.SetGlobal("__lua_validate_env_var", state.NewFunction(v.luaValidateEnvVar))
}

func (v *securityValidator) luaValidatePath(state *lua.LState) int {
	fnName := state.CheckString(1)
	pathValue := state.Get(2)
	if pathValue.Type() != lua.LTString {
		return 0
	}

	if err := v.validatePath(fnName, pathValue.String()); err != nil {
		state.RaiseError("%s", err)
	}
	return 0
}

func (v *securityValidator) luaValidateEnvVar(state *lua.LState) int {
	fnName := state.CheckString(1)
	envVar := state.Get(2)
	if _, allowed := v.allowedEnvVars[envVar.String()]; envVar.Type() != lua.LTString || !allowed {
		state.RaiseError("%s restricted for param %s", fnName, envVar.String())
	}
	return 0
}

func (v *securityValidator) validatePath(fnName, candidate string) error {
	if containsTraversal(candidate) {
		return fmt.Errorf("path traversals are restricted for security")
	}

	resolved := resolvePath(candidate)
	if matchesPath(resolved, v.deniedPaths) {
		return fmt.Errorf("%s restricted for param %s (protected system path)", fnName, candidate)
	}
	if !matchesPath(resolved, v.allowedPaths) {
		return fmt.Errorf("%s restricted for param %s", fnName, candidate)
	}

	return nil
}

// resolvePaths normalizes and resolves each path once when the validator is created. Blank entries
// are excluded so an empty allowlist value cannot match every absolute path.
func resolvePaths(values []string) []string {
	resolved := make([]string, 0, len(values))
	for _, value := range values {
		resolvedPath := resolvePath(value)
		if resolvedPath == "" {
			continue
		}
		resolved = append(resolved, resolvedPath)
	}
	return resolved
}

// resolvePath returns a normalized, symlink-resolved form of value so the sandbox enforces its
// allowlist/denylist against the real target rather than a lexical alias. Because the target may not
// exist yet, it resolves the longest existing ancestor and re-appends the remaining components.
func resolvePath(value string) string {
	abs := normalizedPath(value)
	if abs == "" {
		return ""
	}

	cur, tail := abs, ""
	for {
		if resolved, err := filepath.EvalSymlinks(cur); err == nil {
			return normalizedPath(filepath.Join(resolved, tail))
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			return abs
		}
		tail = filepath.Join(filepath.Base(cur), tail)
		cur = parent
	}
}

func normalizedPath(value string) string {
	value = strings.ReplaceAll(value, `\`, "/")
	if !strings.HasPrefix(value, "/") {
		value = "/" + value
	}
	return strings.TrimRight(path.Clean(value), "/")
}

func containsTraversal(value string) bool {
	for _, segment := range strings.Split(strings.ReplaceAll(value, `\`, "/"), "/") {
		if segment == "." || segment == ".." {
			return true
		}
	}
	return false
}

func matchesPath(candidate string, paths []string) bool {
	for _, prefix := range paths {
		if candidate == prefix || strings.HasPrefix(candidate, prefix+"/") {
			return true
		}
	}
	return false
}
