package checker

import (
	"context"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	log "github.com/sirupsen/logrus"
	lua "github.com/yuin/gopher-lua"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/config"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/output"

	_ "gitlab.jiagouyun.com/cloudcare-tools/sec-checker/funcs/system"
	_ "gitlab.jiagouyun.com/cloudcare-tools/sec-checker/funcs/utils"
)

var (
	Chk *Checker
)

type (
	Checker struct {
		rulesDir  string
		rules     map[string]*Rule
		manifests map[string]*RuleManifest

		ruleGroups map[int64][]*Rule

		scheduler *Scheduler

		ruleMux     sync.Mutex
		manifestMux sync.Mutex

		loading int32
	}
)

// Start
func Start(ctx context.Context, rulesDir string, outputpath *config.ScOutput) {

	log.Debugf("output: %v", outputpath)
	log.Debugf("rule dir: %s", rulesDir)

	Chk = newChecker(rulesDir)

	output.Start(ctx, outputpath)
	Chk.start(ctx)
}

func newChecker(rulesDir string) *Checker {
	c := &Checker{
		rulesDir:  rulesDir,
		rules:     map[string]*Rule{},
		manifests: map[string]*RuleManifest{},
		scheduler: NewScheduler(),
	}

	lua.LuaPathDefault = filepath.Join(rulesDir, "/lib/?.lua")

	return c
}

func (c *Checker) findRule(rulefile string) *Rule {

	if rulefile == "" {
		return nil
	}
	c.ruleMux.Lock()
	defer c.ruleMux.Unlock()
	if r, ok := c.rules[rulefile]; ok {
		return r
	}
	return nil
}
func (c *Checker) addRule(r *Rule) {
	c.ruleMux.Lock()
	defer c.ruleMux.Unlock()

	scheduleID, err := c.scheduler.AddRule(r)
	if err != nil {
		log.Errorf("path=%s r.cron=%s err=%s", r.File, r.cron, err)
		return
	}
	r.scheduleID = scheduleID
	c.rules[r.File] = r
}

func (c *Checker) reSchedule(r *Rule) {
	scheduleID, err := c.scheduler.AddRule(r)
	if err != nil {
		log.Errorf("%s", err)
		return
	}
	r.scheduleID = scheduleID
}

func (c *Checker) delRules() {
	c.ruleMux.Lock()
	defer c.ruleMux.Unlock()

	for _, r := range c.rules {
		if atomic.LoadInt32(&r.markAsDelete) == 0 {
			continue
		}
		log.Debugf("removing %s", r.File)
		if atomic.LoadInt32(&r.running) > 0 {
			select {
			case <-r.stopch:
				c.doDelRule(r)
			case <-time.After(time.Second * 10):
				log.Warnf("remove rule failed, timeout")
			}
		} else {
			c.doDelRule(r)
		}
	}
}

// doDelRule 从checker中删除规则 也应当从文件列表中删除
func (c *Checker) doDelRule(r *Rule) {
	c.scheduler.DelRule(r)
	if r.rt != nil {
		r.rt.Close()
	}
	delete(c.rules, r.File)

	/*
		// delete lua file
		if err := os.Remove(r.File); err != nil {
			log.Warnf("删除lua文件错误 err=%v", err)
		}
		// delete manifest file
		index := strings.LastIndex(r.File, ".")
		manifestFile := r.File[:index] + ".manifest"
		if err := os.Remove(manifestFile); err != nil {
			log.Warnf("删除manifestFile文件错误 err=%v", err)
		}*/
}

func (c *Checker) addManifest(m *RuleManifest) {
	c.manifests[m.path] = m
}

func (c *Checker) getManifest(name string) (*RuleManifest, error) {
	c.manifestMux.Lock()
	defer c.manifestMux.Unlock()

	if !strings.HasSuffix(name, ".manifest") {
		name += ".manifest"
	}

	path := filepath.Join(c.rulesDir, name)
	m := c.manifests[path]
	if m == nil {
		m = newManifest(path)
		c.manifests[path] = m
	}

	err := m.load()
	if err != nil {
		return nil, fmt.Errorf("fail to load %s, %s", path, err)
	}

	return m, nil
}

func (c *Checker) start(ctx context.Context) {
	defer func() {
		if e := recover(); e != nil {
			buf := make([]byte, 2048)
			n := runtime.Stack(buf, false)
			log.Errorf("panic %s", e)
			log.Errorf("%s", string(buf[:n]))

		}
		output.Outputer.Close()
		log.Info("checker exit")
	}()

	if err := c.loadRules(ctx, c.rulesDir); err != nil {
		return
	}
	log.Warnf("------------调度启动------")
	c.scheduler.Start()

	select {
	case <-ctx.Done():
		return
	default:
	}

	if len(c.rules) == 0 {
		log.Warnf("no rule loaded")
	}

	go func() {
		for {
			sleepContext(ctx, time.Second*10)

			select {
			case <-ctx.Done():
				return
			default:
			}

			c.loadRules(ctx, c.rulesDir)
		}
	}()

	<-ctx.Done()
	c.scheduler.Stop()
}

func (c *Checker) loadRules(ctx context.Context, ruleDir string) error {

	if atomic.LoadInt32(&c.loading) > 0 {
		return nil
	}
	atomic.StoreInt32(&c.loading, 1)
	defer atomic.StoreInt32(&c.loading, 0)
	files, err := ioutil.ReadDir(ruleDir)
	if err != nil {
		log.Errorf("loadRules error ：filepath=%s err=%v", ruleDir, err)
		return err
	}

	for _, r := range c.rules {
		atomic.StoreInt32(&r.markAsDelete, 1)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		select {
		case <-ctx.Done():
			return nil
		default:
		}

		path := filepath.Join(ruleDir, file.Name())

		if !strings.HasSuffix(file.Name(), ".lua") {
			continue
		}

		time.Sleep(time.Millisecond * 20)

		if exist, ok := c.rules[path]; ok {
			atomic.StoreInt32(&exist.markAsDelete, 0)
			exist.load()
			continue
		}
		fmt.Printf("初始化 一个rule  参数path=%s \n", path)
		r := newRule(path)
		if err := r.load(); err == nil {
			c.addRule(r)
		}
	}

	c.delRules()

	cronNum, intervalNum := c.scheduler.countInfo()
	log.Debugf("cronNum=%d, intervalNum=%d", cronNum, intervalNum)

	return nil
}

func sleepContext(ctx context.Context, duration time.Duration) error {
	if duration == 0 {
		return nil
	}

	t := time.NewTimer(duration)
	select {
	case <-t.C:
		return nil
	case <-ctx.Done():
		t.Stop()
		return ctx.Err()
	}
}

func TestRule(rulepath string) {

	log.SetReportCaller(true)
	pwd, _ := os.Getwd()

	fmt.Println(filepath.Join(pwd, filepath.Dir(rulepath)))

	config.Cfg = &config.Config{
		System: &config.System{RuleDir: filepath.Join(pwd, filepath.Dir(rulepath))},
	}

	rulepath, _ = filepath.Abs(rulepath)
	rulepath = rulepath + ".lua"
	ruledir := filepath.Dir(rulepath)

	Chk = newChecker(ruledir)
	ctx, cancelfun := context.WithCancel(context.Background())
	output.Start(ctx, &config.ScOutput{})
	defer cancelfun()

	r := newRule(rulepath)

	err := r.load()
	if err != nil {
		return
	}
	Chk.rules[r.File] = r
	r.run()

	return
}
