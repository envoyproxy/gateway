-- Security sandbox for Lua execution in Envoy Gateway: blocks dangerous functions and enforces a
-- fail-closed allowlist of filesystem paths and environment variables during validation.
--
-- The allowed sets are injected by the Go validator before this script runs, as the globals
-- `__lua_allowed_paths` (array of path prefixes) and `__lua_allowed_env_vars` (map name -> true).
-- An absent or empty table denies that entire category.

-- ============================================================================
-- ALLOWLISTS (injected by the Go validator; default to empty = deny all)
-- ============================================================================

local allowed_paths = __lua_allowed_paths or {}
local allowed_env_vars = __lua_allowed_env_vars or {}

-- Remove the injected globals so user code cannot read or mutate the allowlists.
__lua_allowed_paths = nil
__lua_allowed_env_vars = nil

-- Capture the Go-provided symlink resolver, then clear the global so user code cannot reach it.
local resolve_path = __lua_resolve_path
__lua_resolve_path = nil

-- ============================================================================
-- DENYLIST (hardcoded; always denied, even when the allowlist would permit them)
-- ============================================================================

-- Secret and kernel/process paths that are always denied, even if the allowlist permits them.
local denied_paths = {
    "/etc",
    "/proc",
    "/sys",
    "/certs",
    "/var/run/secrets",
    -- "/var/run" is a symlink to "/run" on Debian-derived (distroless) images.
    "/run/secrets",
}

-- ============================================================================
-- HELPER FUNCTIONS
-- ============================================================================

local function to_absolute_normalized_path(path)
    if not path or type(path) ~= "string" then
        return path
    end

    local normalized_separators = path:gsub("\\", "/")

    local collapsed_separators = normalized_separators:gsub("/+", "/")

    local absolute_path
    if collapsed_separators:match("^/") then
        absolute_path = collapsed_separators
    else
        absolute_path = "/" .. collapsed_separators
    end

    return absolute_path:match("^(.-)/*$")
end

-- normalized_real returns the normalized, symlink-resolved absolute form of path. Resolving through
-- the Go helper means a symlink under an allowed directory cannot alias a path outside it (or a
-- denied path): io.open resolves symlinks at open time, so lexical checks alone can be bypassed.
-- Falls back to the lexical normalized form when no resolver is available (e.g. in unit tests).
local function normalized_real(path)
    local normalized = to_absolute_normalized_path(path)
    if resolve_path and type(normalized) == "string" then
        return to_absolute_normalized_path(resolve_path(normalized))
    end
    return normalized
end

local function contains_traversal(path)
    if not path or type(path) ~= "string" then
        return false
    end

    -- Reject any "." or ".." segment regardless of position or separator style.
    -- Trailing "/" ensures the last segment is matched by the "([^/]*)/" pattern.
    local normalized = path:gsub("\\", "/")
    for segment in (normalized .. "/"):gmatch("([^/]*)/") do
        if segment == "." or segment == ".." then
            return true
        end
    end

    return false
end

-- is_allowed_path returns true when resolved (an already symlink-resolved, normalized path) equals
-- an allowed entry or falls within its subtree. Entries are symlink-resolved too so both sides match
-- consistently. The subtree check uses plain (non-pattern) string matching so allowed prefixes
-- containing Lua magic characters (e.g. "." in "/var/lib/app.v1") are treated literally and define
-- an exact boundary.
local function is_allowed_path(resolved)
    for _, allowed in ipairs(allowed_paths) do
        -- Skip blank entries (checked on the lexical form): "" would match every path.
        if to_absolute_normalized_path(allowed) ~= "" then
            local normalized_allowed = normalized_real(allowed)
            if resolved == normalized_allowed
                or resolved:find(normalized_allowed .. "/", 1, true) == 1 then
                return true
            end
        end
    end

    return false
end

-- is_denied_path returns true when resolved (an already symlink-resolved, normalized path) equals a
-- denied entry or falls within its subtree. Entries are symlink-resolved too so both sides match
-- consistently. Uses plain (non-pattern) matching so magic characters in entries are treated literally.
local function is_denied_path(resolved)
    for _, denied in ipairs(denied_paths) do
        local normalized_denied = normalized_real(denied)
        if resolved == normalized_denied
            or resolved:find(normalized_denied .. "/", 1, true) == 1 then
            return true
        end
    end

    return false
end

-- validate_path rejects traversal segments and denied paths unconditionally, then enforces the allowlist.
local function validate_path(fn_name, path)
    if not path or type(path) ~= "string" then
        return
    end

    if contains_traversal(path) then
        error("path traversals are restricted for security")
    end

    -- Resolve symlinks so the denylist/allowlist apply to the real target, not a lexical alias.
    local resolved = normalized_real(path)

    if is_denied_path(resolved) then
        error(fn_name .. " restricted for param " .. path .. " (protected system path)")
    end

    if not is_allowed_path(resolved) then
        error(fn_name .. " restricted for param " .. path)
    end
end

-- validate_env_var enforces the env var allowlist (exact, case-sensitive match).
local function validate_env_var(fn_name, env_var)
    if not env_var or type(env_var) ~= "string" or allowed_env_vars[env_var] ~= true then
        error(fn_name .. " restricted for param " .. tostring(env_var))
    end
end

-- ============================================================================
-- COMPLETELY BLOCKED FUNCTIONS
-- ============================================================================

io.popen = nil
os.execute = nil
os.exit = nil
require = nil
loadfile = nil
dofile = nil
package = nil
debug = nil
load = nil
loadstring = nil
rawget = nil
rawset = nil
getmetatable = nil
setmetatable = nil
-- Block access to global table to prevent _G["_unsafe_*"] bypasses
_G = nil

-- ============================================================================
-- SANITIZED IO FUNCTIONS (path allowlist)
-- ============================================================================

do
    local _unsafe_io_open = io.open
    local _unsafe_io_input = io.input
    local _unsafe_io_output = io.output
    local _unsafe_io_lines = io.lines

    io.open = function(filename, mode)
        validate_path("io.open", filename)
        return _unsafe_io_open(filename, mode)
    end

    io.input = function(file)
        if file == nil then
            return _unsafe_io_input()
        end
        if type(file) == "string" then
            validate_path("io.input", file)
        end
        return _unsafe_io_input(file)
    end

    io.output = function(file)
        if file == nil then
            return _unsafe_io_output()
        end
        if type(file) == "string" then
            validate_path("io.output", file)
        end
        return _unsafe_io_output(file)
    end

    io.lines = function(filename)
        if filename then
            validate_path("io.lines", filename)
        end
        return _unsafe_io_lines(filename)
    end
end

-- ============================================================================
-- SANITIZED OS FUNCTIONS (path / env var allowlist)
-- ============================================================================

do
    local _unsafe_os_remove = os.remove
    local _unsafe_os_rename = os.rename
    local _unsafe_os_getenv = os.getenv
    local _unsafe_os_setenv = os.setenv

    os.remove = function(pathname)
        validate_path("os.remove", pathname)
        return _unsafe_os_remove(pathname)
    end

    os.rename = function(oldname, newname)
        validate_path("os.rename", oldname)
        validate_path("os.rename", newname)
        return _unsafe_os_rename(oldname, newname)
    end

    os.getenv = function(varname)
        validate_env_var("os.getenv", varname)
        return _unsafe_os_getenv(varname)
    end

    os.setenv = function(varname, value)
        validate_env_var("os.setenv", varname)
        return _unsafe_os_setenv(varname, value)
    end
end
