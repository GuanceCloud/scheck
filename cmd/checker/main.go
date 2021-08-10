package main

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"sync"
	"syscall"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/service"

	log "github.com/sirupsen/logrus"
	securityChecker "gitlab.jiagouyun.com/cloudcare-tools/sec-checker"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/checker"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/config"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/man"
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

	if err := config.LoadConfig(*flagConfig); err != nil {
		//fmt.Printf("加载配置文件错误 err=%v \n", err)
		log.Fatalf("%s", err)
	}
	setupLogger()
	service.Entry = run
	if err := service.StartService(); err != nil {
		l.Errorf("start service failed: %s", err.Error())
		return
	}
	//	run()
}

func setupLogger() {
	if config.Cfg.System.DisableLog {
		log.SetLevel(log.PanicLevel)
	} else {
		log.SetReportCaller(true)
		if config.Cfg.Logging.Log != "" {
			lf, err := os.OpenFile(config.Cfg.Logging.Log, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0664)
			if err != nil {
				os.MkdirAll(filepath.Dir(config.Cfg.Logging.Log), 0775)
				lf, err = os.OpenFile(config.Cfg.Logging.Log, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0775)
				if err != nil {
					log.Fatalf("%s", err)
				}
			}
			log.SetOutput(lf)
			//log.SetOutput(os.Stdout) //20210721  暂时修改成终端输出 方便调试
			//log.AddHook() // todo 重写hook接口 就可以实现多端输出
		}
		switch config.Cfg.Logging.LogLevel {
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
		//fmt.Printf("scheck(%s): %s\n", ReleaseType, Version)
		if data, err := ioutil.ReadFile(`/usr/local/scheck/rules.d/version`); err == nil {
			type versionSt struct {
				Version string `json:"version"`
			}
			var verSt versionSt
			if json.Unmarshal(data, &verSt) == nil {
				log.Printf("rules: %s\n", verSt.Version)
			}
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
			log.Fatal(err)
		}
		defer resp.Body.Close()
		remoteVal, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}

		bin := `/usr/local/scheck/scheck`
		data, err := ioutil.ReadFile(bin)
		if err != nil {
			log.Fatalf("%s", err)
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
			log.Fatalf("%s", err)
		}
		os.Stdout.WriteString(string(res))

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
		confPath := "/usr/local/scheck"
		//*flagConfig = "scheck.conf"
		if runtime.GOOS == "windows" { // 设置路径
			confPath = "C:\\Program Files\\scheck"
		}
		*flagConfig = filepath.Join(confPath, "scheck.conf")
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
		log.Println("2 check run")
		// 测试packd2 20210723 测试通过
		man.ScheckCoreSyncDisk(config.Cfg.System.RuleDir)
		time.Sleep(time.Second * 5)
		checker.Start(ctx, config.Cfg.System.RuleDir, config.Cfg.System.CustomRuleDir, config.Cfg.ScOutput)
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
	service.Stop()

}
