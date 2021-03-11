package install

import (
	"github.com/kardianos/service"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	secChecker "gitlab.jiagouyun.com/cloudcare-tools/sec-checker"
)

var (
	l = logger.DefaultSLogger("install")

	//OSArch = runtime.GOOS + "/" + runtime.GOARCH

	InstallDir = ""
)

func Install(svc service.Service) {

	if err := service.Control(svc, "uninstall"); err != nil {
		l.Warnf("uninstall service: %s, ignored", err.Error())
	}

	// build datakit main config
	cfg := secChecker.DefaultConfig()
	if err := cfg.InitCfg(secChecker.MainConfPath); err != nil {
		l.Fatalf("failed to init datakit main config: %s", err.Error())
	}

	l.Infof("installing service %s...", secChecker.ServiceName)
	if err := service.Control(svc, "install"); err != nil {
		l.Warnf("install service: %s, ignored", err.Error())
	}
}

func Upgrade(svc service.Service) {
}
