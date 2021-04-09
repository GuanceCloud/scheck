package checker

import (
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

func (c *luaCron) addRule(r *Rule) (int, error) {
	var id cron.EntryID
	var err error
	id, err = c.AddFunc(r.cron, func() {
		r.run()
	})
	return int(id), err
}

func (c *luaCron) start() {
	c.Start()
}

func (c *luaCron) stop() {
	c.Stop()
}
