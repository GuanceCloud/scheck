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
	"os/signal"
	"runtime"
	"sync"
	"syscall"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	securityChecker "gitlab.jiagouyun.com/cloudcare-tools/sec-checker"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/checker"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/config"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/man"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/service"
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
	config.Cfg.Init()
	l = logger.DefaultSLogger("main")
	//go pprof()
	service.Entry = run
	if err := service.StartService(); err != nil {
		l.Errorf("start service failed: %s", err.Error())
		return
	}
	//	run()
}

func pprof() {
	_ = http.ListenAndServe("0.0.0.0:6060", nil)
}

func applyFlags() {

	if *flagVersion {
		//fmt.Printf("scheck(%s): %s\n", ReleaseType, Version)
		if data, err := ioutil.ReadFile(`/usr/local/scheck/version`); err == nil {
			/*type versionSt struct {
				Version  string `json:"version"`
				BuildAt  string `json:"date_utc"`
				Uploader string `json:"uploader"`
				Branch   string `json:"branch"`
				Commit   string `json:"commit"`
				Golang   string `json:"go"`
			}
			var verSt versionSt
			if json.Unmarshal(data, &verSt) == nil {
				l.Errorf("rules: %s\n", verSt.Version)
			}*/
			fmt.Println(string(data))
		}
		os.Exit(0)
	}

	if *flagCheckMD5 {

		urls := map[string]string{
			"release": `zhuyun-static-files-production.oss-cn-hangzhou.aliyuncs.com/security-checker`,
			"test":    `zhuyun-static-files-testing.oss-cn-hangzhou.aliyuncs.com/security-checker`,
		}

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

		bin := `/usr/local/scheck/scheck`
		data, err := ioutil.ReadFile(bin)
		if err != nil {
			l.Fatalf("%s", err)
		}
		md5 := md5.New()
		md5.Write(data)
		localVal := hex.EncodeToString(md5.Sum(nil))

		if localVal != "" && localVal == string(remoteVal) {
			fmt.Printf("MD5 verify ok\n")
		} else {
			fmt.Printf("[Error] MD5 checksum not match !!!\n")
		}

		os.Exit(0)
	}

	if *flagCfgSample {
		res, err := securityChecker.TomlMarshal(config.DefaultConfig())
		if err != nil {
			l.Fatalf("%s", err)
		}
		os.Stdout.WriteString(string(res))

		os.Exit(0)
	}

	if *flagFuncs {
		securityChecker.DumpSupportLuaFuncs(os.Stdout)
		os.Exit(0)
	}

	if *flagTestRule != "" {
		checker.TestRule(*flagTestRule)
		os.Exit(0)
	}

	if *flagRulesToDoc {
		if *flagOutDir == "" {

			man.ToMakeMdFile(man.GetAllName(), "doc")
		} else {
			man.ToMakeMdFile(man.GetAllName(), *flagOutDir)
		}
		os.Exit(0)
	}

	if *flagRulesToTemplate {
		if *flagOutDir == "" {
			man.DfTemplate(man.GetAllName(), "C://Users/gitee")
		} else {
			man.DfTemplate(man.GetAllName(), *flagOutDir)
		}
		os.Exit(0)
	}
	if *flagConfig == "" {
		*flagConfig = "/usr/local/scheck/scheck.conf"
		if runtime.GOOS == "windows" {
			*flagConfig = "C:\\Program Files\\scheck\\scheck.conf"
		}
		// 查看本地是否有配置文件
		_, err := os.Stat(*flagConfig)
		if err != nil {
			res, err := securityChecker.TomlMarshal(config.DefaultConfig())
			if err != nil {
				l.Fatalf("%s", err)
			}
			fmt.Println(string(res))
			f, err := os.OpenFile(*flagConfig, os.O_CREATE|os.O_RDWR, 0644)
			if err != nil {
				l.Fatalf("%s", err)
			}
			f.WriteString(string(res))
			/*if err = ioutil.WriteFile(*flagConfig, res, 0644); err != nil {
				l.Fatalf("%s", err)
			}*/
		}
	}
}

func run() {

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)
	//内置系统
	go func() {
		defer func() {
			wg.Done()
		}()
		man.SetLog()
		man.ScheckCoreSyncDisk(config.Cfg.System.RuleDir)
		checker.Start(ctx, config.Cfg.System.RuleDir, config.Cfg.System.CustomRuleDir, config.Cfg.ScOutput)
	}()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		select {
		case sig := <-signals:
			if sig == syscall.SIGHUP {
				l.Debugf("reload config")
			}
			cancel()
		}
	}()
	wg.Wait()
	service.Stop()

}
