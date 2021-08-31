package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/checker"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/config"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/funcs"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/git"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/internal/global"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/internal/luafuncs"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/internal/service"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/internal/tools"
)

var (
	flagFuncs           = flag.Bool("funcs", false, `show all supported lua-extend functions`)
	flagVersion         = flag.Bool("version", false, `show version`)
	flagCheckMD5        = flag.Bool("check-md5", false, `md5 checksum`)
	flagCfgSample       = flag.Bool("config-sample", false, `show config sample`)
	flagConfig          = flag.String("config", "", "configuration file to load")
	flagTestRule        = flag.String("test", ``, `the name of a rule, without file extension`)
	flagRulesToDoc      = flag.Bool("doc", false, `Generate doc document from manifest file`)
	flagRulesToTemplate = flag.Bool("tpl", false, `Generate doc document from template file`)
	flagOutDir          = flag.String("dir", "", `document Exported directory`)
)

var (
	Version     = ""
	ReleaseType = ""

	l = logger.DefaultSLogger("main")
)

func main() {
	flag.Parse()
	applyFlags()
	parseConfig()

	if config.TryLoadConfig(*flagConfig) {
		config.LoadConfig(*flagConfig)
	}

	if err := global.SavePid(); err != nil {
		fmt.Println(err.Error())
		os.Exit(-1)
	}
	config.Cfg.Init()
	checker.InitLuaGlobalFunc()
	l = logger.SLogger("main")
	if config.Cfg.System.Pprof {
		go goPprof()
	}
	service.Entry = run
	if err := service.StartService(); err != nil {
		l.Errorf("start service failed: %s", err.Error())
		return
	}
}

func goPprof() {
	_ = http.ListenAndServe(global.DefPprofPort, nil)
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
ReleasedInputs: %s
`, Version, git.Commit, git.Branch, git.BuildAt, git.Golang, git.Uploader, ReleaseType)

		os.Exit(0)
	}
	if *flagCheckMD5 {
		global.CheckMd5(ReleaseType)
		os.Exit(0)
	}

	if *flagCfgSample {
		res, err := tools.TomlMarshal(config.DefaultConfig())
		if err != nil {
			l.Fatalf("%s", err)
		}
		fmt.Println(string(res))
		os.Exit(0)
	}

	if *flagFuncs {
		funcs.DumpSupportLuaFuncs(os.Stdout)
		os.Exit(0)
	}

	if *flagTestRule != "" {
		checker.InitLuaGlobalFunc()
		luafuncs.TestLua(*flagTestRule)
		os.Exit(0)
	}

	if *flagRulesToDoc {
		if *flagOutDir == "" {
			tools.ToMakeMdFile(tools.GetAllName(), "doc")
		} else {
			tools.ToMakeMdFile(tools.GetAllName(), *flagOutDir)
		}
		os.Exit(0)
	}

	if *flagRulesToTemplate {
		if *flagOutDir == "" {
			tools.DfTemplate(tools.GetAllName(), "C://Users/gitee")
		} else {
			tools.DfTemplate(tools.GetAllName(), *flagOutDir)
		}
		os.Exit(0)
	}
}

func parseConfig() {
	binDir := global.InstallDir
	if *flagConfig == "" {
		*flagConfig = filepath.Join(binDir, "scheck.conf")
		_, err := os.Stat(*flagConfig)
		if err != nil {
			res, err := tools.TomlMarshal(config.DefaultConfig())
			if err != nil {
				l.Fatalf("%s", err)
			}
			f, err := os.OpenFile(*flagConfig, os.O_CREATE|os.O_RDWR, os.ModeAppend|os.ModePerm)
			if err != nil {
				l.Fatalf("%s", err)
			}
			_, err = f.WriteString(string(res))
			if err != nil {
				l.Fatalf("write to configFile err =%v", err)
			}
		}
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

		tools.ScheckCoreSyncDisk(config.Cfg.System.RuleDir)
		checker.Start(ctx, config.Cfg.System, config.Cfg.ScOutput)
	}()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		sig := <-signals
		if sig == syscall.SIGHUP {
			l.Debugf("reload config")
		}
		cancel()
	}()
	wg.Wait()
	service.Stop()
}
