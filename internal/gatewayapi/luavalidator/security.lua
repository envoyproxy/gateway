-- Security sandbox for Lua execution in Envoy Gateway. Go performs the security checks; this file
-- only blocks dangerous functions and wraps functions that require path or environment validation.

-- ============================================================================
-- VALIDATORS (injected by the Go validator)
-- ============================================================================

local validate_path = __lua_validate_path
local validate_env_var = __lua_validate_env_var

-- Remove the injected globals so user code cannot access the validators directly.
__lua_validate_path = nil
__lua_validate_env_var = nil

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
