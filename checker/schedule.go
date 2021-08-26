package checker

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	cron "github.com/robfig/cron/v3"
)

var (
	specParser = cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month)
	pool       *statePool
)

//only exec cron timer cron
type TaskScheduler struct {
	rulesDir        string
	customRuleDir   string
	customRulesTime map[string]int64 // key:rules name .val:lastModify time int64
	tasks           map[string]*Rule //key:fileName
	manifests       map[string]*RuleManifest
	stop            chan struct{}
	lock            sync.Mutex
}

//return a Controller Scheduler
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
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		path := filepath.Join(ruleDir, file.Name())
		r := newRule(path)

		if strings.HasSuffix(file.Name(), ".lua") {
			if ruleDir == scheduler.customRuleDir {
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
			}
		}
	}
	if len(scheduler.tasks) == 0 {
		l.Warnf("There are no rules in the folder to load! . system exit at three second !!!")
		time.Sleep(time.Second * 3)
		os.Exit(0)
	}
}

//stop all
func (scheduler *TaskScheduler) Stop() {
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

//run task list
func (scheduler *TaskScheduler) run() {
	if len(scheduler.tasks) == 0 {
		l.Warnf("schedules is empty....")
	}
	for {
		time.Sleep(time.Second / 2)
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
func (scheduler *TaskScheduler) GetRuleByName(filename string) *Rule {
	for key, task := range scheduler.tasks {
		if key == filename {
			return task
		}
	}
	return nil
}

//return a task and key In task list
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
	scheduler.tasks[r.Name] = r
	scheduler.manifests[r.Name] = r.manifest
	scheduler.customRulesTime[r.Name] = fileModify(r.File)
}

func (scheduler *TaskScheduler) removeRule(r *Rule) {
	scheduler.lock.Lock()
	defer scheduler.lock.Unlock()
	delete(scheduler.tasks, r.Name)
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
	for range time.Tick(time.Second * 60) {
		scheduler.LoadFromFile(scheduler.customRuleDir)
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
