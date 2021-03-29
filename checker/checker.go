package checker

import (
	"context"
	"io/ioutil"
	"log"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	lua "github.com/yuin/gopher-lua"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/luaext"
)

type (
	Checker struct {
		rulesDir string
		rules    []*Rule
		lStates  []*luaState

		cron *luaCron

		outputer *outputer

		luaExtends *luaext.LuaExt
	}

	luaState struct {
		lState *lua.LState
		rule   *Rule
	}
)

// NewChecker
func NewChecker(output, rulesDir string) *Checker {
	c := &Checker{
		outputer:   newOutputer(output),
		rulesDir:   rulesDir,
		luaExtends: luaext.NewLuaExt(),
		cron:       newLuaCron(),
	}
	return c
}

// Start
func (c *Checker) Start(ctx context.Context) {

	defer func() {
		if e := recover(); e != nil {
			buf := make([]byte, 2048)
			n := runtime.Stack(buf, false)
			log.Printf("[panic] %s", e)
			log.Printf("[panic] %s", string(buf[:n]))

		}
		c.outputer.close()
		log.Printf("[info] checker exit")
	}()

	log.Printf("[debug] rule dir: %s", c.rulesDir)

	if err := c.loadFiles(); err != nil {
		return
	}

	if len(c.rules) == 0 {
		log.Printf("[warn] no rule found")
		return
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

func (c *Checker) loadFiles() error {

	ls, err := ioutil.ReadDir(c.rulesDir)
	if err != nil {
		log.Printf("[error] %s", err)
		return err
	}

	for _, f := range ls {
		if f.IsDir() {
			continue
		}

		path := filepath.Join(c.rulesDir, f.Name())

		if !strings.HasSuffix(f.Name(), ".lua") {
			//log.Printf("[debug] ignore non-lua %s", path)
			return nil
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
