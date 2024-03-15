package ir

import "sync"

type GlobalLuaScript struct {
	*Metadata
	Lua string
}

// If you want to add more custom resources, then make a similar struct
// and copy it into the ir manager above and create set/delete funcs for the interface
type syncdLuaScripts struct {
	luaScripts map[string]*GlobalLuaScript
	mu         sync.RWMutex
}

func (sf *syncdLuaScripts) Store(luaScript *GlobalLuaScript) {
	sf.mu.Lock()
	defer sf.mu.Unlock()
	sf.luaScripts[luaScript.ID()] = luaScript
}

func (sf *syncdLuaScripts) Load(id string) *GlobalLuaScript {
	sf.mu.RLock()
	defer sf.mu.RUnlock()
	return sf.luaScripts[id]
}

func (sf *syncdLuaScripts) Delete(id string) {
	sf.mu.Lock()
	defer sf.mu.Unlock()
	delete(sf.luaScripts, id)
}

func (sf *syncdLuaScripts) ForEach(f func(id string, luaScript *GlobalLuaScript) error) error {
	sf.mu.RLock()
	defer sf.mu.RUnlock()

	for id, luaScript := range sf.luaScripts {
		if err := f(id, luaScript); err != nil {
			return err
		}
	}
	return nil
}
