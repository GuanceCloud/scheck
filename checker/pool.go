package checker

import (
	"sync"

	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/internal/luafuncs"
)

type sig struct{}

type statePool struct {
	states     []*luafuncs.ScriptRunTime
	poolStatus map[int]bool
	cap        int
	initCap    int
	running    int
	freeSignal chan sig
	lock       sync.Mutex
}

func InitStatePool(initCap, totCap int) {
	p := &statePool{
		states:     make([]*luafuncs.ScriptRunTime, 0),
		poolStatus: make(map[int]bool),
		cap:        totCap,
		initCap:    initCap,
		running:    0,
		freeSignal: make(chan sig, initCap),
	}
	for i := 0; i < p.initCap; i++ {
		state := luafuncs.NewScriptRunTime()
		state.ID = i
		p.states = append(p.states, state)
		p.poolStatus[i] = false
		p.freeSignal <- sig{}
	}
	l.Debugf("init pool ok")
	pool = p
}

// 从池子中获取一个lua state
func (p *statePool) getState() *luafuncs.ScriptRunTime {
	var w *luafuncs.ScriptRunTime
	waiting := false

	p.lock.Lock()
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
	}
	p.lock.Unlock()

	if waiting {
		l.Debug("wait channel ...waiting last state run over")
		<-p.freeSignal
		p.lock.Lock()
		workers = p.states
		n = p.getFreeIndex()
		if n >= 0 {
			w = workers[n]
			p.poolStatus[n] = true
		}
		p.lock.Unlock()
	} else if w == nil {
		w = luafuncs.NewScriptRunTime()
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

func (p *statePool) putPool(srt *luafuncs.ScriptRunTime) {
	p.lock.Lock()
	defer p.lock.Unlock()

	p.running--
	if srt.ID == -1 {
		srt.Ls.Close()
		return
	}
	p.poolStatus[srt.ID] = false
	p.freeSignal <- sig{}
}
