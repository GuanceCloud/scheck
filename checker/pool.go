package checker

import (
	"sync"

	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/internal/lua"
)

type sig struct{}

type statePool struct {
	states     []*lua.ScriptRunTime
	poolStatus map[int]bool
	cap        int
	initCap    int
	running    int
	freeSignal chan sig
	lock       *sync.Mutex
}

func InitStatePool(initCap, totCap int) {
	p := &statePool{
		states:     make([]*lua.ScriptRunTime, 0),
		poolStatus: make(map[int]bool),
		cap:        totCap,
		initCap:    initCap,
		running:    0,
		freeSignal: make(chan sig, initCap),
		lock:       new(sync.Mutex),
	}
	for i := 0; i < p.initCap; i++ {
		state := lua.NewScriptRunTime()
		state.ID = i
		p.states = append(p.states, state)
		p.poolStatus[i] = false
		p.freeSignal <- sig{}
	}
	l.Infof("init lua state pool ok")
	pool = p
}

// 从池子中获取一个lua state.
func (p *statePool) getState() *lua.ScriptRunTime {
	p.lock.Lock()
	var w *lua.ScriptRunTime
	waiting := false
	workers := p.states
	n := p.getFreeIndex()
	if n < 0 {
		if p.running >= p.cap {
			waiting = true
		} else {
			p.running++
		}
	} else {
		<-p.freeSignal
		w = workers[n]
		p.poolStatus[n] = true
		p.running++
	}
	p.lock.Unlock()

	if waiting {
		<-p.freeSignal
		l.Debug("wait channel ok")
		p.lock.Lock()
		n = p.getFreeIndex()
		if n >= 0 {
			w = p.states[n]
			p.poolStatus[n] = true
			p.running++
		}
		p.lock.Unlock()
	} else if w == nil {
		w = lua.NewScriptRunTime()
		w.ID = -1
	}
	return w
}

func (p *statePool) getFreeIndex() int {
	n := -1
	for index, b := range p.poolStatus {
		if !b {
			return index
		}
	}
	return n
}

func (p *statePool) putPool(srt *lua.ScriptRunTime) {
	p.lock.Lock()
	p.running--
	if srt.ID != -1 {
		p.poolStatus[srt.ID] = false
		p.freeSignal <- sig{}
	} else {
		srt.Ls.Close()
	}
	p.lock.Unlock()
}
