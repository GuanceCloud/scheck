package checker

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	log "github.com/sirupsen/logrus"

	cron "github.com/robfig/cron/v3"
)

var (
	specParser = cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month)
)

type Scheduler struct {
	cron *cron.Cron

	intervalGroups map[time.Duration]*intervalGroup
	mux            sync.RWMutex

	ctx       context.Context
	cancelFun context.CancelFunc
}

func NewScheduler() *Scheduler {
	s := &Scheduler{
		cron:           cron.New(cron.WithParser(specParser)),
		intervalGroups: map[time.Duration]*intervalGroup{},
	}
	s.ctx, s.cancelFun = context.WithCancel(context.Background())
	s.cron.Start()
	return s
}

func (s *Scheduler) getGroup(interval time.Duration) (*intervalGroup, time.Duration) {
	for {
		g := s.intervalGroups[interval]
		if g == nil {
			g = newIntervalGroup(interval)
			s.intervalGroups[interval] = g
		}
		if len(g.runs) >= 100 {
			interval += time.Millisecond * 50
		} else {
			return g, interval
		}
	}
}

func (s *Scheduler) countInfo() (cron int, interval int) {
	cron = len(s.cron.Entries())
	for _, g := range s.intervalGroups {
		interval += len(g.runs)
	}
	return
}

func (s *Scheduler) AddRule(r *Rule) (int, error) {
	fmt.Printf("添加一个rule  打印结构体 %+v \n", r)
	if r.interval > 0 { // 设置时间大于0 按照设置的时间间隔进行执行 否则使用默认的cron
		s.mux.Lock()
		defer s.mux.Unlock()
		g, interval := s.getGroup(r.interval)
		r.interval = interval
		return g.add(r), nil
	} else {
		var id cron.EntryID
		var err error
		id, err = s.cron.AddFunc(r.cron, func() { // 添加到调度中的定时器中
			r.run()
		})
		return int(id), err
	}

}

func (s *Scheduler) DelRule(r *Rule) {
	if r.interval > 0 {
		g := s.intervalGroups[r.interval]
		if g == nil {
			return
		}
		g.del(r)
	} else {
		s.cron.Remove(cron.EntryID(r.scheduleID))
	}
}

func (s *Scheduler) Start() {
	s.cron.Start()
	for _, g := range s.intervalGroups {
		log.Infof("调度启动for循环启动  当前的runId=%d ", g.runid)
		g.start(s.ctx)
	}
}

func (s *Scheduler) Stop() {
	s.cancelFun()
	s.cron.Stop()
}

type intervalGroup struct {
	interval time.Duration

	runs  map[int]func()
	runid int

	mux sync.RWMutex

	running int32
}

func newIntervalGroup(interval time.Duration) *intervalGroup {
	return &intervalGroup{
		interval: interval,
		runs:     map[int]func(){},
	}
}

func (g *intervalGroup) add(r *Rule) int {
	g.mux.Lock()
	defer g.mux.Unlock()
	g.runid++
	g.runs[g.runid] = func() {
		r.run()
	}
	return g.runid
}

func (g *intervalGroup) del(r *Rule) {
	g.mux.Lock()
	defer g.mux.Unlock()
	delete(g.runs, r.scheduleID)
}

func (g *intervalGroup) start(ctx context.Context) {

	if atomic.LoadInt32(&g.running) == 1 {
		return
	}

	go func() {
		atomic.StoreInt32(&g.running, 1)
		defer func() {
			atomic.StoreInt32(&g.running, 0)
		}()

		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			start := time.Now()
			fmt.Printf("run group[%s](%d)\n", g.interval, len(g.runs))

			var keys []int
			for k := range g.runs {
				keys = append(keys, k)
			}
			sort.Ints(keys)
			for _, id := range keys {
				select {
				case <-ctx.Done():
					return
				default:
				}
				run := g.runs[id]
				if run != nil {
					log.Infof("当前执行的runId=%d funcID=%d", g.runid, id)
					run()
				}
				time.Sleep(time.Millisecond * 100)
			}
			used := time.Now().Sub(start)
			if used < g.interval { // 线程执行时间小于设置间隔时间 休眠两者时间差 总体时间还是设置的时间 超过就休眠一秒 继续执行
				sleepContext(ctx, g.interval-used)
			} else {
				sleepContext(ctx, time.Second)
			}
		}
	}()

}
