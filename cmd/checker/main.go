package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"

	"os"
	"os/signal"
	"sync"
	"syscall"

	log "github.com/sirupsen/logrus"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/checker"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/luaext"
)

var (
	flagFuncs   = flag.Bool("funcs", false, `show all supported lua-extend functions`)
	flagVersion = flag.Bool("version", false, `show version`)

	flagConfig = flag.String("config", "", "configuration file to load")

	flagTestLuaStr  = flag.String("c", ``, `test a lua string`)
	flagTestLuaFile = flag.String("f", ``, `test a lua file`)
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
	if checker.Cfg.Log != "" {
		lf, err := os.OpenFile(checker.Cfg.Log, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0664)
		if err != nil {
			log.Fatalf("%s", err)
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

func applyFlags() {

	if *flagVersion {
		fmt.Printf("security checker version %s\n", Version)
		os.Exit(0)
	}

	if *flagFuncs {
		luaext.DumpSupportLuaFuncs(os.Stdout)
		os.Exit(0)
	}

	if *flagTestLuaStr != "" {
		err := luaext.RunLuaScriptString(*flagTestLuaStr)
		if err != nil {
			log.Fatalf("[error] %s", err)
		}
		os.Exit(0)
	}

	if *flagTestLuaFile != "" {
		data, err := ioutil.ReadFile(*flagTestLuaFile)
		if err != nil {
			log.Fatalf("[error] %s", err)
		}
		err = luaext.RunLuaScriptString(string(data))
		if err != nil {
			log.Fatalf("%s", err)
		}
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
