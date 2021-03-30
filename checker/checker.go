package checker

import (
	"context"
	"io/ioutil"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	lua "github.com/yuin/gopher-lua"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/luaext"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/output"
)

var (
	checker *Checker
)

type (
	Checker struct {
		rulesDir string
		rules    []*Rule
		lStates  []*luaState

		cron *luaCron

		luaExtends *luaext.LuaExt
	}

	luaState struct {
		lState *lua.LState
		rule   *Rule
	}
)

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

	log.Debugf("rule dir: %s", c.rulesDir)

	if err := c.loadFiles(); err != nil {
		return
	}

	if len(c.rules) == 0 {
		log.Warnf("no rule found")
	}

	c.cron.start()

	for _, r := range c.rules {
		ls := c.newLuaState(r)
		if ls != nil {
			c.lStates = append(c.lStates, ls)
			c.cron.addLuaScript(ls)
		}
	}

	<-ctx.Done()
	c.cron.stop()
}

// Start
func Start(ctx context.Context, rulesDir, outputpath string) {

	checker = &Checker{
		rulesDir:   rulesDir,
		luaExtends: luaext.NewLuaExt(),
		cron:       newLuaCron(),
	}

	log.Debugf("output: %s", outputpath)

	output.NewOutputer(outputpath)
	checker.start(ctx)
}

func (c *Checker) loadFiles() error {

	ls, err := ioutil.ReadDir(c.rulesDir)
	if err != nil {
		log.Errorf("%s", err)
		return err
	}

	for _, f := range ls {
		if f.IsDir() {
			continue
		}

		path := filepath.Join(c.rulesDir, f.Name())

		if !strings.HasSuffix(f.Name(), ".lua") {
			continue
		}

		if r, err := c.newRuleFromFile(path); err == nil {
			c.rules = append(c.rules, r)
		}
	}

	return nil
}

func (c *Checker) newLuaState(r *Rule) *luaState {
	ls := lua.NewState(lua.Options{SkipOpenLibs: true})
	if err := c.luaExtends.Register(ls); err != nil {
		log.Printf("[error] %s", err)
		return nil
	}
	return &luaState{
		lState: ls,
		rule:   r,
	}
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
