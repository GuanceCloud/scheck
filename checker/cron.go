package checker

import (
	"log"

	cron "github.com/robfig/cron/v3"
)

var (
	specParser = cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month)

	// 进程级别的cache
	//globalLuaCache = &module.LuaCache{}
)

type luaCron struct {
	*cron.Cron
}

func newLuaCron() *luaCron {
	return &luaCron{
		cron.New(cron.WithParser(specParser)),
	}
}

func (c *luaCron) addLuaScript(ls *luaState) (err error) {
	_, err = c.AddFunc(ls.rule.ruleCfg.Cron, func() {
		err := ls.rule.run(ls.lState)
		if err != nil {
			log.Printf("[error] run failed, %s", err)
		}
	})
	return
}

func (c *luaCron) start() {
	c.Start()
}

func (c *luaCron) stop() {
	c.Stop()
}
