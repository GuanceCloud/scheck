package checker

import (
	"sync/atomic"

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

func (c *luaCron) addLuaScript(r *Rule) (int, error) {
	var id cron.EntryID
	var err error
	id, err = c.AddFunc(r.cron, func() {

		if atomic.LoadInt32(&r.running) > 0 {
			return
		}
		defer func() {
			if e := recover(); e != nil {
				log.Errorf("panic, %v", e)
			}
			atomic.AddInt32(&r.running, -1)
			close(r.stopch)
		}()
		atomic.AddInt32(&r.running, 1)
		r.stopch = make(chan bool)
		if r.disabled {
			return
		}
		log.Debugf("start run %s", r.File)
		err := r.run()
		if err != nil {
			log.Errorf("run %s failed, %s", r.File, err)
		} else {
			log.Debugf("run %s ok", r.File)
		}
	})
	return int(id), err
}

func (c *luaCron) start() {
	c.Start()
}

func (c *luaCron) stop() {
	c.Stop()
}
