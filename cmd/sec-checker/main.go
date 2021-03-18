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

	flagConfig       = flag.String("config", "", "configuration file to load")
	flagSampleConfig = flag.Bool("sample-config", false, "print out full sample configuration")

	flagTestLuaStr  = flag.String("c", ``, `test a lua string`)
	flagTestLuaFile = flag.String("f", ``, `test a lua file`)

	l = logger.DefaultSLogger("main")
)

var (
	AppName = "sec-checker"
)

func main() {

	flag.Parse()
	applyFlags()

	if err := secChecker.LoadConfig(*flagConfig); err != nil {
		log.Fatalf("fail to laod config file %s, %s", *flagConfig, err)
		return
	}

	setupLogger()

	l.Infof("%s(%s-%s-%s)", AppName, git.Branch, git.Version, git.BuildAt)

	run()
}

func setupLogger() {
	mainCfg := secChecker.Cfg
	if mainCfg.Log != "" {
		if mainCfg.LogRotate > 0 {
			logger.MaxSize = mainCfg.LogRotate
		}
		logger.SetGlobalRootLogger(mainCfg.Log, mainCfg.LogLevel, logger.OPT_DEFAULT)
	}
	l = logger.SLogger("main")
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

	if *flagSampleConfig {
		fmt.Printf(secChecker.SampleMainConfig)
		os.Exit(0)
	}

}

func run() {

	ctx, cancel := context.WithCancel(context.Background())
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

		c := checker.NewChecker(secChecker.Cfg.RuleDir)
		c.Start(ctx)
	}()

	signals := make(chan os.Signal, secChecker.CommonChanCap)
	signal.Notify(signals, os.Interrupt, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		select {
		case sig := <-signals:
			if sig == syscall.SIGHUP {
				// reaload config
			}
			cancel()
		}
	}()

	wg.Wait()

	l.Info("quit")
}
