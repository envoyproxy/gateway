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

-- is_allowed_path returns true when the path equals an allowed entry or falls within its subtree.
-- Both sides are normalized so relative, backslash, and double-slash forms match consistently.
-- The subtree check uses plain (non-pattern) string matching so allowed prefixes containing Lua
-- magic characters (e.g. "." in "/var/lib/app.v1") are treated literally and define an exact boundary.
local function is_allowed_path(path)
    if not path or type(path) ~= "string" then
        return false
    end

    local normalized = to_absolute_normalized_path(path)

    for _, allowed in ipairs(allowed_paths) do
        local normalized_allowed = to_absolute_normalized_path(allowed)

        -- Skip blank entries: "" would match every absolute path and disable the sandbox.
        if normalized_allowed ~= "" and
            (normalized == normalized_allowed
                or normalized:find(normalized_allowed .. "/", 1, true) == 1) then
            return true
        end
    end

    return false
end

-- validate_path rejects traversal segments unconditionally, then enforces the path allowlist.
local function validate_path(fn_name, path)
    if not path or type(path) ~= "string" then
        return
    end

    if contains_traversal(path) then
        error("path traversals are restricted for security")
    end

    if not is_allowed_path(path) then
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
