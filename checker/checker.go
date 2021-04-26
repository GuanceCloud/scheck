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

	cron "github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
	lua "github.com/yuin/gopher-lua"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/output"

	_ "gitlab.jiagouyun.com/cloudcare-tools/sec-checker/funcs/system"
	_ "gitlab.jiagouyun.com/cloudcare-tools/sec-checker/funcs/utils"
)

var (
	checker *Checker
)

type (
	Checker struct {
		rulesDir  string
		rules     map[string]*Rule
		manifests map[string]*RuleManifest

		cron *luaCron

		ruleMux     sync.Mutex
		manifestMux sync.Mutex

		loading int32
	}
)

// Start
func Start(ctx context.Context, rulesDir, outputpath string) {

	log.Debugf("output: %s", outputpath)
	log.Debugf("rule dir: %s", rulesDir)

	checker = newChecker(rulesDir)
	output.Start(ctx, outputpath)
	checker.start(ctx)
}

func newChecker(rulesDir string) *Checker {
	c := &Checker{
		rulesDir:  rulesDir,
		rules:     map[string]*Rule{},
		manifests: map[string]*RuleManifest{},
		cron:      newLuaCron(),
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

	cronID, err := c.cron.addRule(r)
	if err != nil {
		log.Errorf("%s", err)
		return
	}
	r.cronID = cronID
	c.rules[r.File] = r
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

func (c *Checker) doDelRule(r *Rule) {
	c.cron.Cron.Remove(cron.EntryID(r.cronID))
	if r.rt != nil {
		r.rt.Close()
	}
	delete(c.rules, r.File)
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

	c.cron.start()

	if err := c.loadRules(ctx, c.rulesDir); err != nil {
		return
	}

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
			log.Debugf("reload rules")

			select {
			case <-ctx.Done():
				return
			default:
			}

			c.loadRules(ctx, c.rulesDir)
		}
	}()

	<-ctx.Done()
	c.cron.stop()
}

func (c *Checker) loadRules(ctx context.Context, ruleDir string) error {

	if atomic.LoadInt32(&c.loading) > 0 {
		return nil
	}
	atomic.AddInt32(&c.loading, 1)
	defer atomic.AddInt32(&c.loading, -1)
	ls, err := ioutil.ReadDir(ruleDir)
	if err != nil {
		log.Errorf("%s", err)
		return err
	}

	for _, r := range c.rules {
		atomic.AddInt32(&r.markAsDelete, 1)
	}

	for _, f := range ls {
		if f.IsDir() {
			continue
		}

		select {
		case <-ctx.Done():
			return nil
		default:
		}

		path := filepath.Join(ruleDir, f.Name())

		if !strings.HasSuffix(f.Name(), ".lua") {
			continue
		}

		if exist, ok := c.rules[path]; ok {
			atomic.AddInt32(&exist.markAsDelete, -1)
			exist.load()
			continue
		}

		r := newRule(path)
		if err := r.load(); err == nil {
			c.addRule(r)
		}
	}

	c.delRules()

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

	rulepath, _ = filepath.Abs(rulepath)
	rulepath = rulepath + ".lua"
	ruledir := filepath.Dir(rulepath)

	checker = newChecker(ruledir)
	ctx, cancelfun := context.WithCancel(context.Background())
	output.Start(ctx, "")
	defer cancelfun()

	r := newRule(rulepath)

	err := r.load()
	if err != nil {
		return
	}
	checker.rules[r.File] = r
	r.run()

	return
}
