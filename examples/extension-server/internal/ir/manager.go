package ir

// The manager manages additions/deletions/updates to the ir. For example, the controller will get updates for CRUDs to the lua scripts
// the manager is used in that instance to make sure that everywhere else that is using the ir this manager controls does not have that ir
// updated in the middle of it being used
type IRManager struct {
	ir              *IR
	syncdLuaScripts *syncdLuaScripts
}

func NewIRManager(ir *IR) *IRManager {
	return &IRManager{
		ir: ir,
		syncdLuaScripts: &syncdLuaScripts{
			luaScripts: make(map[string]*GlobalLuaScript),
		},
	}
}

func (m *IRManager) StoreLuaScript(luaScript *GlobalLuaScript) {
	m.syncdLuaScripts.Store(luaScript)
	m.updateLuaScripts()
}

func (m *IRManager) DeleteLuaScript(id string) {
	m.syncdLuaScripts.Delete(id)
	m.updateLuaScripts()
}

func (m *IRManager) updateLuaScripts() {
	luaScripts := make(map[string]*GlobalLuaScript)

	m.syncdLuaScripts.ForEach(func(id string, luaScript *GlobalLuaScript) error {
		luaScripts[id] = luaScript
		return nil
	})

	m.ir.StoreLuaScripts(luaScripts)
}
