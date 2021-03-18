package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"

	"github.com/kardianos/service"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	secChecker "gitlab.jiagouyun.com/cloudcare-tools/sec-checker"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/cmd/installer/install"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/git"
)

var (
	flagVersion = flag.Bool("version", false, "show installer version info")
	flagUpgrade = flag.Bool("upgrade", false, "")
	fInstallDir = flag.String("installdir", "", "")

	OSSLocation = ""
	AppVersion  = ""

	packageDownloadURL = "https://" + path.Join(OSSLocation, secChecker.OSSPath,
		fmt.Sprintf("%s-%s-%s-%s.tar.gz", secChecker.BinName, runtime.GOOS, runtime.GOARCH, AppVersion))

	l *logger.Logger
)

func main() {

	flag.Parse()

	if *flagVersion {
		fmt.Printf(`
       Version: %s
      Build At: %s
`, git.Version, git.BuildAt)
		return
	}

	lopt := logger.OPT_DEFAULT | logger.OPT_STDOUT
	if runtime.GOOS != "windows" { // disable color on windows(some color not working under windows)
		lopt |= logger.OPT_COLOR
	}

	logger.SetGlobalRootLogger("", logger.DEBUG, lopt)
	l = logger.SLogger("installer")

	installDir := *fInstallDir
	if installDir != "" {
		if abspath, err := filepath.Abs(installDir); err != nil {
			l.Fatalf("%s", err)
		} else {
			installDir = abspath
		}
	}
	if installDir == "" {
		installDir = secChecker.GetServiceDefaltInstallDir(secChecker.ServiceDirName)
	}

	// create install dir if not exists
	if err := os.MkdirAll(installDir, 0775); err != nil {
		l.Fatal(err)
	}

	secChecker.ServiceExecutable = filepath.Join(installDir, secChecker.BinName)
	if runtime.GOOS == secChecker.OSWindows {
		secChecker.ServiceExecutable += ".exe"
	}

	svc, err := secChecker.NewService()
	if err != nil {
		l.Errorf("new %s service failed: %s", runtime.GOOS, err.Error())
		return
	}

	l.Info("stoping service...")
	if err := service.Control(svc, "stop"); err != nil {
		l.Warnf("stop service: %s, ignored", err.Error())
	}

	if err := install.Download(packageDownloadURL, installDir); err != nil {
		l.Fatalf("%s", err)
	}

	if *flagUpgrade { // upgrade new version
		l.Infof("Upgrading to version %s...", AppVersion)
	} else { // install

		if err := service.Control(svc, "uninstall"); err != nil {
			l.Warnf("uninstall service failed: %s, ignored", err.Error())
		}

		l.Infof("Installing version %s...", AppVersion)
		if err := secChecker.PreInstall(installDir); err != nil {
			l.Fatalf("%s", err)
		}
		if err := service.Control(svc, "install"); err != nil {
			l.Fatalf("install service: %s, ignored", err.Error())
		}
		secChecker.PostInstall(installDir)
	}

	l.Infof("starting service %s...", secChecker.ServiceName)
	if err = service.Control(svc, "start"); err != nil {
		l.Fatalf("star service: %s", err.Error())
	}

	if *flagUpgrade { // upgrade new version
		l.Info(":) Upgrade Success!")
	} else {
		l.Info(":) Install Success!")
	}
}
