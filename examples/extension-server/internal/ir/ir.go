package ir

import (
	"sync/atomic"
)

// IR contains the current snapshot of our custom resources
type IR struct {
	// We store the resources as an atomic.Value that is a map of all our LuaScripts
	// key:   namespace.name
	// value: the resource
	luaScripts atomic.Value
}

func NewIR() *IR {
	var luaScripts atomic.Value

	luaScripts.Store(make(map[string]*GlobalLuaScript))

	return &IR{
		luaScripts: luaScripts,
	}
}

// StoreLuaScripts overwrites the current global lua scripts
func (ir *IR) StoreLuaScripts(luaScripts map[string]*GlobalLuaScript) {
	ir.luaScripts.Store(luaScripts)
}

// LoadLuaScript returns a lua script if it exists in the ir or nil if it doesn't
func (ir *IR) LoadLuaScript(luaScriptID string) *GlobalLuaScript {
	luaScripts := ir.LoadLuaScripts()

	if luaScript, ok := luaScripts[luaScriptID]; ok {
		return luaScript
	}

	return nil
}

func (ir *IR) LoadLuaScripts() map[string]*GlobalLuaScript {
	aVal := ir.luaScripts.Load()
	if aVal == nil {
		return make(map[string]*GlobalLuaScript)
	}

	if luaScripts, ok := aVal.(map[string]*GlobalLuaScript); ok {
		return luaScripts
	}

	return make(map[string]*GlobalLuaScript)
}
