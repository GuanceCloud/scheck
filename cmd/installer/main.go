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
	securityChecker "gitlab.jiagouyun.com/cloudcare-tools/sec-checker"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/cmd/installer/install"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/config"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/git"
	dkservice "gitlab.jiagouyun.com/cloudcare-tools/sec-checker/service"
)

var (
	DataKitBaseURL = "" // 作为下载时使用 应为 <your-oss-bucket>.oss-cn-hangzhou.aliyuncs.com/xxx/scheck
	DataKitVersion = "" // 版本号

	oldInstallDir      = "/usr/local/cloudcare/dataflux/scheck"
	oldInstallDirWin   = `C:\Program Files\dataflux\scheck`
	oldInstallDirWin32 = `C:\Program Files (x86)\dataflux\scheck`

	datakitUrl = "https://" + path.Join(DataKitBaseURL,
		fmt.Sprintf("scheck-%s-%s-%s.tar.gz", runtime.GOOS, runtime.GOARCH, DataKitVersion))

	// dataUrl = "https://" + path.Join(DataKitBaseURL, "data.tar.gz")

	l = logger.DefaultSLogger("installer")
)

var (
	flagUpgrade     = flag.Bool("upgrade", false, ``)
	flagDatawayHTTP = flag.String("dataway", "", `address of dataway(http://IP:Port?token=xxx), port default 9528`)

	flagInfo         = flag.Bool("info", false, "show installer info")
	flagDownloadOnly = flag.Bool("download-only", false, `download scheck only, not install`)
	flagInstallOnly  = flag.Bool("install-only", false, "install only, not start")

	flagEnableInputs = flag.String("enable-inputs", "", `default enable inputs(comma splited, example: cpu,mem,disk)`)
	flagDatakitName  = flag.String("name", "", `specify scheck name, example: prod-env-scheck`)
	flagGlobalTags   = flag.String("global-tags", "", `enable global tags, example: host=__datakit_hostname,ip=__datakit_ip`)

	flagPort      = flag.Int("port", 9529, "scheck HTTP port")
	flagListen    = flag.String("listen", "localhost", "scheck HTTP listen")
	flagNamespace = flag.String("namespace", "", "scheck namespace")

	flagInstallLog    = flag.String("install-log", "", "install log")
	flagOTA           = flag.Bool("ota", false, "auto update")
	flagCloudProvider = flag.String("cloud-provider", "", "specify cloud provider(accept aliyun/tencent/aws)")

	flagOffline = flag.Bool("offline", false, "offline install mode")
	flagSrcs    = flag.String("srcs", fmt.Sprintf("./scheck-%s-%s-%s.tar.gz,./data.tar.gz", runtime.GOOS, runtime.GOARCH, DataKitVersion), `local path of scheck and agent install files`)
)

const (
	datakitBin = "scheck"
	dlDatakit  = "scheck"
	dlData     = "data"
)

func main() {

	flag.Parse()

	if *flagInstallLog == "" {
		lopt := logger.OPT_DEFAULT | logger.OPT_STDOUT
		if runtime.GOOS != "windows" { // disable color on windows(some color not working under windows)
			lopt |= logger.OPT_COLOR
		}

		if err := logger.SetGlobalRootLogger("", logger.DEBUG, lopt); err != nil {
			l.Warnf("set root log failed: %s", err.Error())
		}
	} else {
		l.Infof("set log file to %s", *flagInstallLog)
		if err := logger.SetGlobalRootLogger(*flagInstallLog, logger.DEBUG, logger.OPT_DEFAULT); err != nil {
			l.Errorf("set root log failed: %s", err.Error())
		}
		install.Init()
	}

	l = logger.SLogger("installer")

	dkservice.ServiceExecutable = filepath.Join(securityChecker.InstallDir, datakitBin)
	if runtime.GOOS == securityChecker.OSWindows {
		dkservice.ServiceExecutable += ".exe"
	}

	svc, err := dkservice.NewService()
	if err != nil {
		l.Errorf("new %s service failed: %s", runtime.GOOS, err.Error())
		return
	}

	l.Info("stoping scheck...")
	if err := service.Control(svc, "stop"); err != nil {
		l.Warnf("stop service: %s, ignored", err.Error())
	}

	// 迁移老版本 scheck 数据目录
	mvOldDatakit(svc)

	applyFlags()
	l.Infof("打印系统参数 DataKitBaseURL=%s  Version:=%s", DataKitBaseURL, DataKitVersion)
	// create install dir if not exists
	if err := os.MkdirAll(securityChecker.InstallDir, 0775); err != nil {
		l.Warnf("makeDirAll %v", err)
	}

	if *flagOffline && *flagSrcs != "" {
		for _, f := range strings.Split(*flagSrcs, ",") {
			_ = install.ExtractDatakit(f, securityChecker.InstallDir)
		}
	} else {
		l.Debugf("进入下载")
		install.CurDownloading = dlDatakit
		if err := install.Download(datakitUrl, securityChecker.InstallDir, true, true, false); err != nil {
			l.Errorf("err = %v", err)
			return
		}
		// download version
		l.Debugf("download version...")
		vUrl := "https://" + path.Join(DataKitBaseURL, "Version")
		if err := install.Download(vUrl, filepath.Join(securityChecker.InstallDir, "version"), false, true, true); err != nil {
			l.Errorf("err = %v", err)
			return
		}

	}

	if *flagUpgrade { // upgrade new version
		l.Infof("Upgrading to version %s...", DataKitVersion)
		if err := install.UpgradeDatakit(svc); err != nil {
			l.Fatalf("upgrade scheck: %s, ignored", err.Error())
		}
	} else { // install new scheck
		l.Infof("Installing version %s...", DataKitVersion)
		install.InstallNewDatakit(svc)
	}

	// 启动 服务
	if !*flagInstallOnly {
		l.Infof("starting service %s...", dkservice.ServiceName)
		if err = service.Control(svc, "start"); err != nil {
			l.Warnf("star service: %s, ignored", err.Error())
		}
	}

	config.CreateSymlinks()

	if *flagUpgrade { // upgrade new version
		l.Info(":) Upgrade Success!")
	} else {
		l.Info(":) Install Success!")
	}

	// promptReferences()
}

func promptReferences() {
	fmt.Printf("\n\tVisit http://localhost:%d/man/changelog to see scheck change logs.\n", *flagPort)
	fmt.Printf("\tVisit http://localhost:%d/monitor to see scheck running status.\n", *flagPort)
	fmt.Printf("\tVisit http://localhost:%d/man to see scheck manuals.\n\n", *flagPort)
}

func applyFlags() {

	if *flagInfo {
		fmt.Printf(`
       Version: %s
      Build At: %s
Golang Version: %s
       BaseUrl: %s
       scheck: %s
`, securityChecker.Version, git.BuildAt, git.Golang, DataKitBaseURL, datakitUrl)
		os.Exit(0)
	}

	if *flagDownloadOnly {
		install.DownloadOnly = true

		install.CurDownloading = dlDatakit

		if err := install.Download(datakitUrl,
			fmt.Sprintf("scheck-%s-%s-%s.tar.gz",
				runtime.GOOS, runtime.GOARCH, DataKitVersion), true, true, true); err != nil {
			return
		}
		//fmt.Printf("\n")
		/*
			install.CurDownloading = dlData
			if err := install.Download(dataUrl, "data.tar.gz", true, true); err != nil {
				return
			}

			fmt.Printf("\n")
		*/
		os.Exit(0)
	}

}

func mvOldDatakit(svc service.Service) {
	olddir := oldInstallDir
	switch runtime.GOOS + "/" + runtime.GOARCH {
	case securityChecker.OSArchWinAmd64:
		olddir = oldInstallDirWin
	case securityChecker.OSArchWin386:
		olddir = oldInstallDirWin32
	}

	if _, err := os.Stat(olddir); err != nil {
		l.Debugf("path %s not exists, ingored", olddir)
	}

	if err := service.Control(svc, "uninstall"); err != nil {
		l.Warnf("uninstall service scheck failed: %s, ignored", err.Error())
	}

	// 如果测试可以 屏蔽重命名操作 避免引起权限问题
	/*if err := os.Rename(olddir, securityChecker.InstallDir); err != nil {
		l.Fatalf("move %s -> %s failed: %s", olddir, securityChecker.InstallDir, err.Error())
	}*/
	//move C:\Program Files\dataflux\scheck -> C:\Program Files\scheck failed: rename C:\Program Files\dataflux\scheck C:\Program Files\scheck: Access is denied

}
