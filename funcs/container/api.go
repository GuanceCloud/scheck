package container

import lua "github.com/yuin/gopher-lua"

var api = map[string]lua.LGFunction{
	// before other docker funcs ,sc_docker_exist  must at first !!!
	"sc_docker_exist":      DockerExist,
	"sc_docker_images":     DockerImagesList,
	"sc_docker_containers": DockerContainerList,
	"sc_docker_runlike":    DockerRunlike,

	// k8s
	"sc_tls_cipher_suites":    TLSCipherSuites,
	"sc_kubectl_checkVersion": CheckVersion,
}

func Loader(l *lua.LState) int {
	t := l.NewTable()
	mod := l.SetFuncs(t, api)
	l.Push(mod)
	return 1
}
