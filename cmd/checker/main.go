package main

import (
	"context"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/checker"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/luaext"
)

var (
	flagFuncs = flag.Bool("funcs", false, `show all supported lua-extend functions`)

	flagConfig = flag.String("config", "", "configuration file to load")

	flagTestLuaStr  = flag.String("c", ``, `test a lua string`)
	flagTestLuaFile = flag.String("f", ``, `test a lua file`)
)

func main() {

	flag.Parse()
	applyFlags()
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	if err := checker.LoadConfig(*flagConfig); err != nil {
		log.Fatalf("%s", err)
	}

	run()
}

func applyFlags() {

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

		c := checker.NewChecker(checker.Cfg.Output, checker.Cfg.RuleDir)
		c.Start(ctx)
	}()

	signals := make(chan os.Signal, 1)
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
	log.Printf("[info] quit")
}
