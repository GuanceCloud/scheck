package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/kardianos/service"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/cmd/installer/install"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/git"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/internal/global"
	dkservice "gitlab.jiagouyun.com/cloudcare-tools/sec-checker/internal/service"
)

var (
	DataKitBaseURL = ""
	DataKitVersion = ""

	oldInstallDir      = "/usr/local/cloudcare/dataflux/scheck"
	oldInstallDirWin   = `C:\Program Files\dataflux\scheck`
	oldInstallDirWin32 = `C:\Program Files (x86)\dataflux\scheck`

	datakitURL = "https://" + path.Join(DataKitBaseURL,
		fmt.Sprintf("scheck-%s-%s-%s.tar.gz", runtime.GOOS, runtime.GOARCH, DataKitVersion))

	l = logger.DefaultSLogger("installer")
)

var (
	flagUpgrade      = flag.Bool("upgrade", false, ``)
	flagInfo         = flag.Bool("info", false, "show installer info")
	flagDownloadOnly = flag.Bool("download-only", false, `download scheck only, not install`)
	flagInstallOnly  = flag.Bool("install-only", false, "install only, not start")
	flagInstallLog   = flag.String("install-log", "", "install log")
	flagOffline      = flag.Bool("offline", false, "offline install mode")
	flagSrcs         = flag.String("srcs",
		fmt.Sprintf("./scheck-%s-%s-%s.tar.gz,./data.tar.gz", runtime.GOOS, runtime.GOARCH, DataKitVersion),
		`local path of scheck and agent install files`)
)

func main() {
	flag.Parse()
	parseLog()

	dkservice.ServiceExecutable = filepath.Join(global.InstallDir, global.AppBin)
	if runtime.GOOS == global.OSWindows {
		dkservice.ServiceExecutable += ".exe"
	}

	svc, err := dkservice.NewService()
	if err != nil {
		l.Errorf("new %s service failed: %s", runtime.GOOS, err.Error())
		return
	}

	l.Info("stoping scheck...")
	if err = service.Control(svc, "stop"); err != nil {
		l.Warnf("stop service: %s, ignored", err.Error())
	}

	mvOldScheck(svc)
	applyFlags()

	// create install dir if not exists
	if err = os.MkdirAll(global.InstallDir, os.ModeDir|os.ModePerm); err != nil {
		l.Warnf("makeDirAll %v", err)
	}

	if *flagOffline && *flagSrcs != "" {
		for _, f := range strings.Split(*flagSrcs, ",") {
			_ = install.ExtractDatakit(f, global.InstallDir)
		}
	} else {
		install.CurDownloading = global.AppBin
		if err = install.Download(datakitURL, global.InstallDir, true, true, false); err != nil {
			l.Errorf("err = %v", err)
			return
		}
		// download version
		vURL := "https://" + path.Join(DataKitBaseURL, "version")
		if err = install.Download(vURL, filepath.Join(global.InstallDir, "version"), false, true, true); err != nil {
			l.Errorf("err = %v", err)
			return
		}
	}
	parseUpgrade(svc)

	global.CreateSymlinks()

	if *flagUpgrade { // upgrade new version
		l.Info(":) Upgrade Success!")
	} else {
		l.Info(":) Install Success!")
	}
}

func parseUpgrade(svc service.Service) {
	if *flagUpgrade {
		l.Infof("Upgrading to version %s...", DataKitVersion)
		if err := install.UpgradeDatakit(svc); err != nil {
			l.Fatalf("upgrade scheck: %s, ignored", err.Error())
		}
	} else {
		l.Infof("Installing version %s...", DataKitVersion)
		install.NewScheck(svc)
	}

	if !*flagInstallOnly {
		l.Infof("starting service %s...", dkservice.ServiceName)
		if err := service.Control(svc, "start"); err != nil {
			l.Warnf("star service: %s, ignored", err.Error())
		}
	}
}

func parseLog() {
	if *flagInstallLog == "" {
		lopt := logger.OPT_DEFAULT | logger.OPT_STDOUT
		if runtime.GOOS != "windows" { // disable color on windows(some color not working under windows)
			lopt |= logger.OPT_COLOR
		}
		opt := &logger.Option{Path: "", Level: "debug", Flags: lopt}
		if err := logger.InitRoot(opt); err != nil {
			l.Warnf("set root log failed: %s", err.Error())
		}
	} else {
		l.Infof("set log file to %s", *flagInstallLog)
		if err := logger.InitRoot(&logger.Option{Path: *flagInstallLog, Level: logger.DEBUG, Flags: logger.OPT_DEFAULT}); err != nil {
			l.Errorf("set root log failed: %s", err.Error())
		}
		install.Init()
	}
}

func applyFlags() {
	if *flagInfo {
		fmt.Printf(`
       Version: %s
      Build At: %s
Golang Version: %s
       BaseUrl: %s
       scheck: %s
`, global.Version, git.BuildAt, git.Golang, DataKitBaseURL, datakitURL)
		os.Exit(0)
	}

	if *flagDownloadOnly {
		install.DownloadOnly = true

		install.CurDownloading = global.AppBin

		if err := install.Download(datakitURL,
			fmt.Sprintf("scheck-%s-%s-%s.tar.gz",
				runtime.GOOS, runtime.GOARCH, DataKitVersion), true, true, true); err != nil {
			return
		}

		os.Exit(0)
	}
}

func mvOldScheck(svc service.Service) {
	olddir := oldInstallDir
	switch runtime.GOOS + "/" + runtime.GOARCH {
	case global.OSArchWinAmd64:
		olddir = oldInstallDirWin
	case global.OSArchWin386:
		olddir = oldInstallDirWin32
	}

	if _, err := os.Stat(olddir); err != nil {
		l.Debugf("path %s not exists, ingored", olddir)
		return
	}

	if err := service.Control(svc, "uninstall"); err != nil {
		l.Warnf("uninstall service scheck failed: %s, ignored", err.Error())
	}
}
