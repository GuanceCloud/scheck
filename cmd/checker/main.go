package main

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/internal/global"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/checker"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/config"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/funcs"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/git"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/internal/service"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/internal/tools"
)

var (
	flagFuncs   = flag.Bool("funcs", false, `show all supported lua-extend functions`)
	flagVersion = flag.Bool("version", false, `show version`)

	flagCheckMD5 = flag.Bool("check-md5", false, `md5 checksum`)

	flagCfgSample = flag.Bool("config-sample", false, `show config sample`)

	flagConfig = flag.String("config", "", "configuration file to load")

	flagTestRule = flag.String("test", ``, `the name of a rule, without file extension`)

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
	if config.TryLoadConfig(*flagConfig) {
		config.LoadConfig(*flagConfig)
	}
	if checkServiceExist() {
		l.Fatalf("service scheck is running!!!")
	}
	config.Cfg.Init()
	l = logger.SLogger("main")
	if config.Cfg.System.Pprof {
		go goPprof()
	}
	service.Entry = run
	if err := service.StartService(); err != nil {
		l.Errorf("start service failed: %s", err.Error())
		return
	}
	//	run()
}

func goPprof() {

	_ = http.ListenAndServe("0.0.0.0:6060", nil)
}

func applyFlags() {

	binDir := global.InstallDir
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

		urls := map[string]string{"release": global.ReleaseUrl, "test": global.TestUrl}

		url := fmt.Sprintf("https://%s/scheck-%s-%s-%s.md5", urls[ReleaseType], runtime.GOOS, runtime.GOARCH, Version)
		resp, err := http.Get(url)
		if err != nil {
			l.Fatal(err)
		}
		defer resp.Body.Close()
		remoteVal, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			l.Fatal(err)
		}

		data, err := ioutil.ReadFile(filepath.Join(binDir, "scheck"))
		if err != nil {
			l.Fatalf("%s", err)
		}
		newMd5 := md5.New()
		newMd5.Write(data)
		localVal := hex.EncodeToString(newMd5.Sum(nil))

		if localVal != "" && localVal == string(remoteVal) {
			l.Debug("MD5 verify ok")
		} else {
			l.Debug("[Error] MD5 checksum not match !!!")
		}

		os.Exit(0)
	}

	if *flagCfgSample {
		res, err := tools.TomlMarshal(config.DefaultConfig())
		if err != nil {
			l.Fatalf("%s", err)
		}
		_, _ = os.Stdout.WriteString(string(res))

		os.Exit(0)
	}

	if *flagFuncs {
		funcs.DumpSupportLuaFuncs(os.Stdout)
		os.Exit(0)
	}

	if *flagTestRule != "" {
		funcs.TestLua(*flagTestRule)

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
	if *flagConfig == "" {
		*flagConfig = filepath.Join(binDir, "scheck.conf")
		_, err := os.Stat(*flagConfig)
		if err != nil {
			res, err := tools.TomlMarshal(config.DefaultConfig())
			if err != nil {
				l.Fatalf("%s", err)
			}

			f, err := os.OpenFile(*flagConfig, os.O_CREATE|os.O_RDWR, 0644)
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

// checkServiceExist :The server cannot run two scheck at the same time
func checkServiceExist() bool {
	if runtime.GOOS == "windows" {
		//  tasklist /fi "SERVICES eq scheck"  or  IMAGENAME eq scheck
		cmd := exec.Command("tasklist", "/FI", "IMAGENAME eq scheck.exe")
		out, err := cmd.CombinedOutput()
		if err != nil {
			l.Error(err)
		}

		if n := strings.Count(string(out), "scheck"); n >= 2 {
			return true
		}
	} else {
		cmds := []*exec.Cmd{
			exec.Command("ps", "-ef"),
			exec.Command("grep", "scheck"),
			exec.Command("grep", "-v", "grep"),
			exec.Command("wc", "-l"),
		}
		result, _ := tools.ExecPipeLine(cmds...)

		if len(result) > 0 {
			n, err := strconv.Atoi(result)
			if err == nil {
				if n >= 2 {
					return true
				}
			}

		}
	}

	return false
}
