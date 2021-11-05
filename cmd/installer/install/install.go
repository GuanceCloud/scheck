package install

import (
	"github.com/kardianos/service"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	dkservice "gitlab.jiagouyun.com/cloudcare-tools/sec-checker/internal/service"
)

var (
	DownloadOnly   bool
	CurDownloading = ""
	l              = logger.DefaultSLogger("install")
)

func NewScheck(svc service.Service) {
	if err := service.Control(svc, "uninstall"); err != nil {
		l.Warnf("uninstall service: %s, ignored", err.Error())
	}
	l.Infof("installing service %s...", dkservice.ServiceName)
	if err := service.Control(svc, "install"); err != nil {
		l.Warnf("install service: %s, ignored", err.Error())
	}
}

func UpgradeDatakit(svc service.Service) error {
	if err := service.Control(svc, "stop"); err != nil {
		l.Warnf("stop service: %s, ignored", err.Error())
	}
	if err := service.Control(svc, "install"); err != nil {
		l.Warnf("install datakit service: %s, ignored", err.Error())
	}

	return nil
}

func Init() {
	l = logger.SLogger("install")
}
