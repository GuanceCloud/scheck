package checker

import (
	cron "github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
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
		log.Debugf("start run %s", ls.rule.File)
		err := ls.rule.run(ls.lState)
		if err != nil {
			log.Errorf("run %s failed, %s", ls.rule.File, err)
		} else {
			log.Debugf("run %s ok", ls.rule.File)
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
