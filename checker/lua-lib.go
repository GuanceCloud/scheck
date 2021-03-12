package checker

import lua "github.com/yuin/gopher-lua"

type luaLib struct {
	name            string
	fn              lua.LGFunction
	disabledMethods []string
}

var (
	supportLuaLibs = []luaLib{
		{lua.BaseLibName, lua.OpenBase, nil},
		{lua.TabLibName, lua.OpenTable, nil},
		{lua.StringLibName, lua.OpenString, nil},
		{lua.MathLibName, lua.OpenMath, nil},
		{lua.DebugLibName, lua.OpenDebug, nil},
	}
)

func unsupportFn(ls *lua.LState) int {
	lv := ls.Get(lua.UpvalueIndex(1))
	ls.RaiseError("'%s' diabled", lv.String())
	return 0
}

func loadLibs(ls *lua.LState) error {

	for _, lib := range supportLuaLibs {

		err := ls.CallByParam(lua.P{
			Fn:      ls.NewFunction(lib.fn),
			NRet:    1,
			Protect: true,
		}, lua.LString(lib.name))
		if err != nil {
			l.Fatalf("load lua lib failed, %s", err)
			return err
		}
		lvMod := ls.Get(-1)
		if lvMod.Type() == lua.LTTable {
			lt := lvMod.(*lua.LTable)
			for _, mth := range lib.disabledMethods {
				lt.RawSetString(mth, ls.NewClosure(unsupportFn, lua.LString(mth)))
			}
		}
		ls.Pop(1)
	}
	//l.Debugf("statck size: %d", ls.GetTop())
	return nil
}
