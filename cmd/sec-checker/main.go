package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	secChecker "gitlab.jiagouyun.com/cloudcare-tools/sec-checker"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/checker"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/funcs"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/git"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/io"
)

var (
	flagVersion = flag.Bool("version", false, `show version info`)
	flagFuncs   = flag.Bool("funcs", false, `show all support lua functions`)
	flagCmd     = flag.Bool("cmd", false, "run as command")

	flagRuleDir = flag.String("rule", ``, `directory which contains the lua script files`)

	flagTestLuaStr  = flag.String("c", ``, `test a lua string`)
	flagTestLuaFile = flag.String("f", ``, `test a lua file`)

	l = logger.DefaultSLogger("main")
)

func main() {

	flag.Parse()

	applyFlags()

	if err := secChecker.LoadConfig(secChecker.MainConfPath(secChecker.InstallDir)); err != nil {
		log.Fatalf("fail to laod config file, %s", err)
		return
	}

	mainCfg := secChecker.Cfg
	if mainCfg.Log != "" {
		logger.MaxSize = mainCfg.LogRotate
		logger.SetGlobalRootLogger(mainCfg.Log, mainCfg.LogLevel, logger.OPT_DEFAULT)
		l = logger.SLogger("main")
	}

	l.Infof("%s(%s-%s-%s)", secChecker.ServiceName, git.Branch, git.Version, git.BuildAt)

	funcs.Init()

	if *flagCmd {
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

	if *flagFuncs {
		names := []string{``}
		for _, f := range funcs.SupportFuncs {
			names = append(names, f.Name)
		}
		log.Printf("%s", strings.Join(names, "\n"))
		os.Exit(0)
	}

	if *flagTestLuaStr != "" {
		err := checker.TestLuaScriptString(*flagTestLuaStr)
		if err != nil {
			log.Fatalf("%s", err)
		}
		os.Exit(0)
	}

	if *flagTestLuaFile != "" {
		data, err := ioutil.ReadFile(*flagTestLuaFile)
		if err != nil {
			log.Fatalf("%s", err)
		}
		err = checker.TestLuaScriptString(string(data))
		if err != nil {
			log.Fatalf("%s", err)
		}
		os.Exit(0)
	}

}

func run() {

	ctx, cancelFun := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer func() {
			wg.Done()
		}()
		io.Start(ctx)
	}()

	wg.Add(1)
	go func() {
		defer func() {
			wg.Done()
		}()
		ruleDir := *flagRuleDir
		if ruleDir == "" {
			ruleDir = secChecker.RulesDir(secChecker.InstallDir)
		}
		c := checker.NewChecker(ruleDir)
		c.Start(ctx)
	}()

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
			cancelFun()
			wg.Wait()
			secChecker.Quit()
		}

	case <-secChecker.StopCh:
		l.Infof("service stopping")
		cancelFun()
		wg.Wait()
		secChecker.Quit()
	}

	l.Info("exit")
}
