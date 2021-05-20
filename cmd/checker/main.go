package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"runtime"
	"text/template"

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

	flagCheckMD5 = flag.Bool("check-md5", false, `md5 checksum`)

	flagCfgSample = flag.Bool("config-sample", false, `show config sample`)

	flagConfig = flag.String("config", "", "configuration file to load")

	flagTestRule = flag.String("test", ``, `the name of a rule, without file extension`)

	flagUpdateRules = flag.Bool("update-rules", false, `update rules`)
)

var (
	Version     = ""
	ReleaseType = ""
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
		fmt.Printf("scheck(%s): %s\n", ReleaseType, Version)
		if data, err := ioutil.ReadFile(`/usr/local/scheck/rules.d/version`); err == nil {
			type versionSt struct {
				Version string `json:"version"`
			}
			var verSt versionSt
			if json.Unmarshal(data, &verSt) == nil {
				fmt.Printf("rules: %s\n", verSt.Version)
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

		defaultOutput := `file:///var/log/scheck/event.log`
		output := os.Getenv("SCHECK_OUTPUT")
		if output == "" {
			output = defaultOutput
		}
		tmp, err := template.New("cfg").Parse(checker.MainConfigSample)
		if err != nil {
			log.Fatalf("%s", err)
		}
		var buf bytes.Buffer
		err = tmp.Execute(&buf, map[string]string{
			"Output": output,
		})
		if err != nil {
			log.Fatalf("%s", err)
		}
		os.Stdout.WriteString(buf.String())
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

	if *flagUpdateRules {
		if !updateRules() {
			os.Exit(1)
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

func updateRules() bool {
	ruleDir := "/usr/local/scheck/rules.d"
	if err := os.MkdirAll(ruleDir, 0775); err != nil {
		fmt.Printf("[error] %s\n", err)
		return false
	}
	ruleVer := `https://security-checker-prod.oss-cn-hangzhou.aliyuncs.com/prod/version`
	resp, err := http.Get(ruleVer)
	if err != nil {
		fmt.Printf("[error] %s\n", err)
		return false
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		resp.Body.Close()
		fmt.Printf("[error] %s\n", err)
		return false
	}
	resp.Body.Close()

	type versionSt struct {
		Version string `json:"version"`
	}
	var verSt versionSt
	if err = json.Unmarshal(data, &verSt); err != nil {
		fmt.Printf("[error] %s\n", err)
		return false
	}

	ruleUrl := fmt.Sprintf("https://security-checker-prod.oss-cn-hangzhou.aliyuncs.com/prod/rule-%s.tar.gz", verSt.Version)
	fmt.Println("Downloading rules...")
	resp, err = http.Get(ruleUrl)
	if err != nil {
		fmt.Printf("[error] %s\n", err)
		return false
	}
	defer resp.Body.Close()
	data, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("[error] %s\n", err)
		return false
	}
	fmt.Println("Intsalling rules...")
	buf := bytes.NewBuffer(data)
	gr, err := gzip.NewReader(buf)
	if err != nil {
		fmt.Printf("[error] %s\n", err)
		return false
	}
	defer gr.Close()
	tr := tar.NewReader(gr)
	for {
		hdr, err := tr.Next()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				fmt.Printf("[error] %s\n", err)
				return false
			}
		}
		fpath := filepath.Join(ruleDir, hdr.Name)
		if !hdr.FileInfo().IsDir() {
			fdir := filepath.Dir(fpath)
			if err := os.MkdirAll(fdir, 0775); err != nil {
				fmt.Printf("[error] %s\n", err)
				return false
			}
			file, err := os.Create(fpath)
			if err != nil {
				fmt.Printf("[error] %s\n", err)
				return false
			}
			if _, err := io.Copy(file, tr); err != nil {
				fmt.Printf("[error] %s\n", err)
				return false
			}
		}
	}
	fmt.Println("Install rules successfully")
	return true
}
