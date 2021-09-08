package checker

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/internal/luafuncs"

	cron "github.com/robfig/cron/v3"
)

var (
	specParser = cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month)
	pool       *statePool
)

// only exec cron timer cron
type TaskScheduler struct {
	rulesDir        string
	customRuleDir   string
	customRulesTime map[string]int64 // key:rules name .val:lastModify time int64
	tasks           map[string]*Rule // key:fileName
	onceTasks       map[string]*Rule
	manifests       map[string]*RuleManifest
	stop            chan struct{}
	lock            sync.Mutex
}

// NewTaskScheduler: return a Controller Scheduler
func NewTaskScheduler(rulesDir, customRuleDir string, hotUpdate bool) *TaskScheduler {
	schedule := &TaskScheduler{
		rulesDir:        rulesDir,
		customRuleDir:   customRuleDir,
		customRulesTime: make(map[string]int64),
		tasks:           make(map[string]*Rule),
		manifests:       make(map[string]*RuleManifest),
		stop:            make(chan struct{}),
	}
	schedule.LoadFromFile(rulesDir)
	schedule.LoadFromFile(customRuleDir)
	if hotUpdate {
		go schedule.hotUpdate()
	}
	return schedule
}

func (scheduler *TaskScheduler) LoadFromFile(ruleDir string) {
	files, err := ioutil.ReadDir(ruleDir)
	if err != nil {
		l.Errorf("loadRules error ：filepath=%s err=%v", ruleDir, err)
		return
	}
	isCustom := ruleDir == scheduler.customRuleDir
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		path := filepath.Join(ruleDir, file.Name())
		r := newRule(path)

		if strings.HasSuffix(file.Name(), ".lua") {
			if isCustom {
				rulename := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
				if t, ok := scheduler.customRulesTime[rulename]; ok {
					if t == fileModify(path) {
						continue
					}
				}
			}
			if err := r.load(); err != nil {
				l.Errorf("load from dir err=%v", err)
				continue
			}
			if !r.disabled {
				scheduler.addRule(r)
				luafuncs.Add(r.Name, r.manifest.Category, r.interval, isCustom)
			}
		}
	}
	if len(scheduler.tasks) == 0 {
		l.Warnf("There are no rules in the folder to load! . system exit at three second !!!")
		os.Exit(-1)
	}
}

// stop all
func (scheduler *TaskScheduler) Stop() {
	if len(scheduler.onceTasks) != 0 {
		scheduler.stop <- struct{}{}
	}
	scheduler.stop <- struct{}{}
}

// doAndReset
func (scheduler *TaskScheduler) doAndReset(key string) {
	scheduler.lock.Lock()
	defer scheduler.lock.Unlock()
	if task, ok := scheduler.tasks[key]; ok {
		task.RunTime = time.Now().Unix() + task.interval
		scheduler.tasks[key] = task
	}
}

// run task list
func (scheduler *TaskScheduler) run() {
	if len(scheduler.tasks) == 0 {
		l.Warnf("schedules is empty....")
		return
	}
	for {
		now := time.Now()
		task, key := scheduler.GetTask()
		runTime := task.RunTime
		i64 := runTime - now.Unix()
		if i64 <= 0 {
			if task != nil {
				go task.RunJob()
			}
			scheduler.doAndReset(key)
			continue
		}
		timer := time.NewTimer(time.Second * time.Duration(i64))
		l.Debugf("scheduler new time.timer, at %d Seconds start...", i64)
		for {
			select {
			case <-timer.C:
				scheduler.doAndReset(key)
				if task != nil {
					go task.RunJob()
					timer.Stop()
				}
			case <-scheduler.stop:
				l.Info("scheduler Stop ...")
				timer.Stop()
				return
			}
			break
		}
	}
}

// runOther: if rule.cron=disable then rule run once.
func (scheduler *TaskScheduler) runOnce() {
	if len(scheduler.onceTasks) == 0 {
		l.Warnf("schedules  is empty....")
		return
	}
	cxtMap := sync.Map{}                                   // 主动停止通知信号
	errChan := make(chan string, len(scheduler.onceTasks)) // 被动停止通知信号
	count := 0                                             // 运行的数量
	for _, rule := range scheduler.onceTasks {
		cxt := context.Background()
		go rule.RunOnce(cxt, errChan)
		cxtMap.Store(rule.Name, cxt)
		count++
	}
	for {
		select {
		case <-scheduler.stop:
			// all context stop
			cxtMap.Range(func(key, value interface{}) bool {
				if cxt, ok := value.(context.Context); ok {
					cxt.Done()
				}
				return false
			})
		case <-time.After(time.Minute):
			// 检查运行数量
			if len(scheduler.onceTasks) != count {
				// do something...
				l.Warnf("Unexpected reduction in number of runs !!! tot=%d run=%d", len(scheduler.onceTasks), count)
			}
			if count == 0 {
				close(errChan)
			}
		case name := <-errChan:
			// to call monitor 。。。
			count--
			l.Errorf("rule name = %s is stop!!!", name)
		}
	}
}

func (scheduler *TaskScheduler) GetRuleByName(filename string) *Rule {
	for key, task := range scheduler.tasks {
		if key == filename {
			return task
		}
	}
	return nil
}

// return a task and key In task list
func (scheduler *TaskScheduler) GetTask() (task *Rule, tempKey string) {
	min := int64(0)
	tempKey = ""
	i := 0
	for key, task := range scheduler.tasks {
		if i == 0 {
			i++
			min = task.RunTime
			tempKey = key
			continue
		}
		tTime := task.RunTime
		if min <= tTime {
			continue
		}
		if min > tTime {
			tempKey = key
			min = tTime
			continue
		}
	}
	task = scheduler.tasks[tempKey]
	return task, tempKey
}

func (scheduler *TaskScheduler) addRule(r *Rule) {
	scheduler.lock.Lock()
	defer scheduler.lock.Unlock()
	if r.cron == "" || r.cron == "disable" {
		scheduler.onceTasks[r.Name] = r
	} else {
		scheduler.tasks[r.Name] = r
	}
	scheduler.manifests[r.Name] = r.manifest
	scheduler.customRulesTime[r.Name] = fileModify(r.File)
}

// hotUpdate to hotUpdate users rules dir
func (scheduler *TaskScheduler) hotUpdate() {
	files, err := ioutil.ReadDir(scheduler.customRuleDir)
	if err != nil {
		l.Errorf("hotUpdate :loadRules filepath=%s err=%v", scheduler.customRuleDir, err)
		return
	}
	if files != nil && len(files) == 0 {
		return
	}
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-scheduler.stop:
			l.Info("Done!")
			return
		case <-ticker.C:
			scheduler.LoadFromFile(scheduler.customRuleDir)
		}
	}
}

func GetRuleNum() int {
	if Chk != nil && Chk.taskScheduler != nil {
		return len(Chk.taskScheduler.tasks)
	}
	return 0
}

func fileModify(filePath string) int64 {
	fileInfo, _ := os.Stat(filePath)
	ruleModify := fileInfo.ModTime().Unix()
	ruledir := filepath.Dir(filePath)
	rulename := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))

	manifestPath := filepath.Join(ruledir, rulename+".manifest")
	MfileInfo, err := os.Stat(manifestPath)
	if err != nil {
		l.Error("cannot find manifest file ,lua and manifest  must exist !!!")
		return 0
	}
	manifestModify := MfileInfo.ModTime().Unix()
	return ruleModify + manifestModify
}
