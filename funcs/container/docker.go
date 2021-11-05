package container

import (
	"strings"

	lua "github.com/yuin/gopher-lua"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/funcs/container/impl"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/funcs/container/impl/utils"
)

// DockerExist : before other docker funcs ,sc_docker_exist  must at first !!!
func DockerExist(l *lua.LState) int {
	if _, err := impl.GetCli(); err != nil {
		l.Push(lua.LFalse)
		return 1
	}
	l.Push(lua.LTrue)
	return 1
}

func DockerImagesList(l *lua.LState) int {
	cli, err := impl.GetCli()
	if err != nil {
		l.RaiseError("%s", err)
		return lua.MultRet
	}
	images, err := impl.GetImageList(cli)
	if err != nil {
		l.RaiseError("%s", err)
		return lua.MultRet
	}
	tab := l.NewTable()
	for i := range images {
		img := l.NewTable()
		img.RawSetString("Containers", lua.LNumber(images[i].Containers))
		img.RawSetString("Created", lua.LNumber(images[i].Created))
		img.RawSetString("ID", lua.LString(images[i].ID))
		tab.Append(img)
	}
	l.Push(tab)
	return 1
}

// DockerContainerList :set container list to lua.table.
func DockerContainerList(l *lua.LState) int {
	cli, err := impl.GetCli()
	if err != nil {
		l.RaiseError("%s", err)
		return lua.MultRet
	}
	containers, err := impl.GetContainerList(cli)
	if err != nil {
		l.RaiseError("%s", err)
		return lua.MultRet
	}
	tab := l.NewTable()
	for i := range containers {
		img := l.NewTable()
		img.RawSetString("ID", lua.LString(containers[i].ID))
		tab.Append(img)
	}
	l.Push(tab)
	return 1
}

// DockerRunlike :获取docker启动时候的 run命令，参数等，返回是一个字符串。.
func DockerRunlike(l *lua.LState) int {
	lv := l.Get(1)
	if lv.Type() != lua.LTString {
		l.TypeError(1, lua.LTString)
		return lua.MultRet
	}

	name := string(lv.(lua.LString))
	name = strings.TrimSpace(name)
	result, err := utils.ContainerInspectMap(name)
	if err != nil {
		l.RaiseError("%s", err)
		return lua.MultRet
	}
	l.Push(lua.LString(impl.Inspect(result, false)))
	return 0
}
