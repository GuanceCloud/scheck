package checker

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	lua "github.com/yuin/gopher-lua"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"

	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/funcs"
)

const (
	MaxLuaStates = 20
)

var (
	l *logger.Logger
)

type (
	Checker struct {
		rulesDir string
		rules    []*Rule
		lStates  []*luaState
		ch       chan *Rule
	}

	luaState struct {
		lState *lua.LState
	}
)

func NewChecker(rulesDir string) *Checker {
	c := &Checker{
		rulesDir: rulesDir,
		ch:       make(chan *Rule, 10),
	}
	return c
}

func (c *Checker) loadFiles() error {

	if err := filepath.Walk(c.rulesDir, func(fp string, f os.FileInfo, err error) error {
		if err != nil {
			l.Error(err)
		}

		if f.Name() == "." || f.Name() == ".." {
			return nil
		}

		if f.IsDir() {
			return nil
		}

		if !strings.HasSuffix(f.Name(), ".lua") {
			l.Debugf("ignore non-lua %s", fp)
			return nil
		}

		if r, err := newRuleFromFile(fp); err == nil {
			c.rules = append(c.rules, r)
			l.Debugf("load %s ok", fp)
		}

		return nil
	}); err != nil {
		l.Error(err)
		return err
	}
	return nil
}

func (c *Checker) newLuaState() *luaState {
	ls := lua.NewState(lua.Options{SkipOpenLibs: true})
	if err := funcs.LoadLuaLibs(ls); err != nil {
		l.Errorf("%s", err)
		return nil
	}
	for _, fn := range funcs.SupportFuncs {
		ls.Register(fn.Name, fn.Fn)
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
				err := r.run(ls.lState)
				r.LastRun = time.Now()
				if err != nil {
					l.Errorf("run failed, %s", err)
				}
			} else {
				SleepContext(ctx, time.Second*3)

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

func (c *Checker) Start(ctx context.Context) {

	l = logger.SLogger("checker")

	defer func() {
		if e := recover(); e != nil {
			l.Errorf("panic: %s", e)
		}
		l.Infof("checker exit")
	}()

	if err := c.loadFiles(); err != nil {
		return
	}

	if len(c.rules) == 0 {
		l.Warnf("no rule found")
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

func TestLuaScriptString(script string) error {
	ls := lua.NewState()
	if err := funcs.LoadLuaLibs(ls); err != nil {
		return err
	}
	defer ls.Close()

	return ls.DoString(script)
}

// SleepContext sleeps until the context is closed or the duration is reached.
func SleepContext(ctx context.Context, duration time.Duration) error {
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
