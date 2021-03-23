package checker

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	lua "github.com/yuin/gopher-lua"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/luaext"
)

const (
	MaxLuaStates = 20
)

type (
	Checker struct {
		rulesDir string
		rules    []*Rule
		lStates  []*luaState
		ch       chan *Rule

		outputer *outputer

		luaExtends *luaext.LuaExt
	}

	luaState struct {
		lState *lua.LState
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

	c.ch = make(chan *Rule, len(c.rules))
	for _, rule := range c.rules {
		c.ch <- rule
	}

	var wg sync.WaitGroup
	for i := 0; i < MaxLuaStates; i++ {
		ls := c.newLuaState()
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

	if err := filepath.Walk(c.rulesDir, func(fp string, f os.FileInfo, err error) error {

		if f == nil {
			return nil
		}

		if err != nil {
			log.Printf("%s", err)
		}

		if f.Name() == "." || f.Name() == ".." {
			return nil
		}

		if f.IsDir() {
			return nil
		}

		if !strings.HasSuffix(f.Name(), ".lua") {
			log.Printf("[debug] ignore non-lua %s", fp)
			return nil
		}

		if r, err := NewRuleFromFile(fp); err == nil {
			c.rules = append(c.rules, r)
		} else {
			log.Printf("[error] load %s failed, %s", fp, err)
		}

		return nil
	}); err != nil {
		log.Printf("[error] %s", err)
		return err
	}
	return nil
}

func (c *Checker) newLuaState() *luaState {
	ls := lua.NewState(lua.Options{SkipOpenLibs: true})
	if err := c.luaExtends.Register(ls); err != nil {
		log.Printf("[error] %s", err)
		return nil
	}
	return &luaState{
		lState: ls,
	}
}

func (c *Checker) startState(ctx context.Context, ls *luaState) {
	for {
		select {
		case r := <-c.ch:
			if r.LastRun.IsZero() || time.Now().Sub(r.LastRun) >= r.Interval {
				err := r.Run(ls.lState)
				r.LastRun = time.Now()
				if err != nil {
					log.Printf("[error] run failed, %s", err)
				}
			} else {
				sleepContext(ctx, time.Second*3)

				select {
				case <-ctx.Done():
					return
				default:
				}
			}
			c.ch <- r
		case <-ctx.Done():
			return
		}
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
