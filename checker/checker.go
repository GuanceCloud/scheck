package checker

import (
	"context"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
	"sync"
	"time"

	lua "github.com/yuin/gopher-lua"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/luaext"
)

type (
	Checker struct {
		rulesDir string
		rules    []*Rule
		lStates  []*luaState

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
	}
	return c
}

// Start
func (c *Checker) Start(ctx context.Context) {

	defer func() {
		if e := recover(); e != nil {
			log.Printf("[panic] %s", e)
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

	var wg sync.WaitGroup
	for _, r := range c.rules {
		ls := c.newLuaState(r)
		if ls != nil {
			wg.Add(1)
			c.lStates = append(c.lStates, ls)
			go func() {
				defer wg.Done()
				c.startState(ctx, ls)
			}()
		}
	}

	wg.Wait()
}

func (c *Checker) loadFiles() error {

	ls, err := ioutil.ReadDir(c.rulesDir)
	if err != nil {
		return err
	}

	for _, f := range ls {
		if f.IsDir() {
			continue
		}

		path := filepath.Join(c.rulesDir, f.Name())

		if !strings.HasSuffix(f.Name(), ".lua") {
			log.Printf("[debug] ignore non-lua %s", path)
			return nil
		}

		if r, err := NewRuleFromFile(filepath.Join(c.rulesDir, path)); err == nil {
			c.rules = append(c.rules, r)
		} else {
			log.Printf("[error] load %s failed, %s", path, err)
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

func (c *Checker) startState(ctx context.Context, ls *luaState) {
	for {
		select {
		case <-ctx.Done():
			break
		}
		if ls.rule.LastRun.IsZero() || time.Now().Sub(ls.rule.LastRun) >= ls.rule.Interval {
			err := ls.rule.Run(ls.lState)
			ls.rule.LastRun = time.Now()
			if err != nil {
				log.Printf("[error] run failed, %s", err)
			}
		}
		sleepContext(ctx, ls.rule.Interval)
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
