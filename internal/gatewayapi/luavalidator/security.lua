-- Security sandbox for Lua execution in Envoy Gateway
-- Blocks dangerous functions and validates paths to prevent access to sensitive system resources

-- ============================================================================
-- CRITICAL PATHS
-- ============================================================================

local critical_paths = {
    "/etc",
    "/proc",
    "/sys",
    "/certs",
    "/var/run/secrets",
}

-- ============================================================================
-- CRITICAL ENVIRONMENT VARIABLES
-- ============================================================================

local critical_env_vars = {
    ["PWD"] = true,
}

-- ============================================================================
-- HELPER FUNCTIONS
-- ============================================================================

local function to_absolute_normalized_path(path)
    if not path or type(path) ~= "string" then
        return path
    end
    
    local normalized_separators = path:gsub("\\", "/")
    
    local absolute_path
    if normalized_separators:match("^/") then
        absolute_path = normalized_separators
    else
        absolute_path = "/" .. normalized_separators
    end
    
    return absolute_path:match("^(.-)/*$")
end

local function contains_traversal(path)
    if not path or type(path) ~= "string" then
        return false
    end
    
    if path:match("/%.%./") or path:match("^%.%./") or path:match("/%.%.$") or path:match("^%.%.$") or
       path:match("\\%.%.\\") or path:match("^%.%.\\") or path:match("\\%.%.$") or path:match("^%.%.$") then
        return true
    end
    
    return false
end

local function is_critical_path(path)
    if not path or type(path) ~= "string" then
        return false
    end
    
    local normalized = to_absolute_normalized_path(path)
    
    for _, critical_path in ipairs(critical_paths) do
        local normalized_critical = to_absolute_normalized_path(critical_path)
        local escaped_critical = normalized_critical:gsub("%-", "%%-")
        
        if normalized == normalized_critical or normalized:match("^" .. escaped_critical .. "/") then
            return true
        end
    end
    
    return false
end

local function validate_path(path)
    if not path or type(path) ~= "string" then
        return
    end
    
    if contains_traversal(path) then
        error("path traversals are restricted for security")
    end
    
    if is_critical_path(path) then
        error("access to critical path " .. path .. " is restricted for security")
    end
end

local function is_critical_env_var(env_var)
    if not env_var or type(env_var) ~= "string" then
        return false
    end
    return critical_env_vars[env_var:upper()] == true
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
-- SANITIZED IO FUNCTIONS (path validation)
-- ============================================================================

do
    local _unsafe_io_open = io.open
    local _unsafe_io_input = io.input
    local _unsafe_io_output = io.output
    local _unsafe_io_lines = io.lines

    io.open = function(filename, mode)
        validate_path(filename)
        return _unsafe_io_open(filename, mode)
    end

    io.input = function(file)
        if file == nil then
            return _unsafe_io_input()
        end
        if type(file) == "string" then
            validate_path(file)
        end
        return _unsafe_io_input(file)
    end

    io.output = function(file)
        if file == nil then
            return _unsafe_io_output()
        end
        if type(file) == "string" then
            validate_path(file)
        end
        return _unsafe_io_output(file)
    end

    io.lines = function(filename)
        if filename then
            validate_path(filename)
        end
        return _unsafe_io_lines(filename)
    end
end

-- ============================================================================
-- SANITIZED OS FUNCTIONS (path/env var validation)
-- ============================================================================

do
    local _unsafe_os_remove = os.remove
    local _unsafe_os_rename = os.rename
    local _unsafe_os_getenv = os.getenv
    local _unsafe_os_setenv = os.setenv

    os.remove = function(pathname)
        validate_path(pathname)
        return _unsafe_os_remove(pathname)
    end

    os.rename = function(oldname, newname)
        validate_path(oldname)
        validate_path(newname)
        return _unsafe_os_rename(oldname, newname)
    end

    os.getenv = function(varname)
        if is_critical_env_var(varname) then
            error("access to critical environment variable " .. varname .. " is restricted for security")
        end
        return _unsafe_os_getenv(varname)
    end

    os.setenv = function(varname, value)
        if is_critical_env_var(varname) then
            error("setting critical environment variable " .. varname .. " is restricted for security")
        end
        return _unsafe_os_setenv(varname, value)
    end
end
