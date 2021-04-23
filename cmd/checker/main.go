package main

import (
	"context"
	"flag"
	"fmt"
	"path/filepath"

	"os"
	"os/signal"
	"sync"
	"syscall"

	log "github.com/sirupsen/logrus"
	securityChecker "gitlab.jiagouyun.com/cloudcare-tools/sec-checker"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/checker"
)

var (
	flagFuncs   = flag.Bool("funcs", false, `show all supported lua-extend functions`)
	flagVersion = flag.Bool("version", false, `show version`)

	flagCfgSample = flag.Bool("config-sample", false, `show config sample`)

	flagConfig = flag.String("config", "", "configuration file to load")

	flagTestRule = flag.String("test", ``, `the name of a rule, without file extension`)
)

var (
	Version = ""
)

func main() {

	flag.Parse()
	applyFlags()

	if err := checker.LoadConfig(*flagConfig); err != nil {
		log.Fatalf("%s", err)
	}

	setupLogger()

	run()
}

func setupLogger() {
	if checker.Cfg.DisableLog {
		log.SetLevel(log.PanicLevel)
	} else {
		log.SetReportCaller(true)
		if checker.Cfg.Log != "" {
			lf, err := os.OpenFile(checker.Cfg.Log, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0664)
			if err != nil {
				os.MkdirAll(filepath.Dir(checker.Cfg.Log), 0775)
				lf, err = os.OpenFile(checker.Cfg.Log, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0664)
				if err != nil {
					log.Fatalf("%s", err)
				}
			}
			log.SetOutput(lf)
		}
		switch checker.Cfg.LogLevel {
		case "debug":
			log.SetLevel(log.DebugLevel)
		case "warn":
			log.SetLevel(log.WarnLevel)
		case "error":
			log.SetLevel(log.ErrorLevel)
		default:
			log.SetLevel(log.InfoLevel)
		}
	}
}

func applyFlags() {

	if *flagVersion {
		fmt.Printf("security checker version %s\n", Version)
		os.Exit(0)
	}

	if *flagCfgSample {
		os.Stdout.WriteString(checker.MainConfigSample)
		os.Exit(0)
	}

	if *flagFuncs {
		securityChecker.DumpSupportLuaFuncs(os.Stdout)
		os.Exit(0)
	}

	if *flagTestRule != "" {
		log.SetLevel(log.DebugLevel)
		checker.TestRule(*flagTestRule)
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

		checker.Start(ctx, checker.Cfg.RuleDir, checker.Cfg.Output)
	}()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		select {
		case sig := <-signals:
			if sig == syscall.SIGHUP {
				log.Debugf("reload config")
			}
			cancel()
		}
	}()

	wg.Wait()
	log.Printf("[info] quit")
}
