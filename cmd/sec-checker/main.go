package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	secChecker "gitlab.jiagouyun.com/cloudcare-tools/sec-checker"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/git"
)

var (
	flagVersion = flag.Bool("version", false, `show version info`)
	flagDocker  = flag.Bool("docker", false, "run within docker")

	l = logger.DefaultSLogger("main")
)

func main() {

	flag.Parse()
	applyFlags()

	secChecker.LoadConfig(secChecker.MainConfPath)

	mainCfg := secChecker.Cfg
	if mainCfg.Log != "" {
		// set global log root
		l.Infof("set log to %s", mainCfg.Log)
		logger.MaxSize = mainCfg.LogRotate
		logger.SetGlobalRootLogger(mainCfg.Log, mainCfg.LogLevel, logger.OPT_DEFAULT)
		l = logger.SLogger("main")
	}

	if *flagDocker {
		run()
	} else {
		secChecker.Entry = run
		if err := secChecker.StartService(); err != nil {
			l.Errorf("start service failed: %s", err.Error())
			return
		}
	}

}

func applyFlags() {

	if *flagVersion {
		fmt.Printf(`
       Version: %s
        Commit: %s
        Branch: %s
 Build At(UTC): %s
Golang Version: %s
      Uploader: %s
`, git.Version, git.Commit, git.Branch, git.BuildAt, git.Golang, git.Uploader)

		os.Exit(0)
	}

}

func run() {

	// NOTE:
	// Actually, the datakit process been managed by system service, no matter on
	// windows/UNIX, datakit should exit via `service-stop' operation, so the signal
	// branch should not reached, but for daily debugging(ctrl-c), we kept the signal
	// exit option.
	signals := make(chan os.Signal, secChecker.CommonChanCap)
	signal.Notify(signals, os.Interrupt, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGINT)
	select {
	case sig := <-signals:
		if sig == syscall.SIGHUP {
			// TODO: reload configures
		} else {
			l.Infof("get signal %v, wait & exit", sig)
			secChecker.Quit()
		}

	case <-secChecker.StopCh:
		l.Infof("service stopping")
		secChecker.Quit()
	}

	l.Info("exit")
}
