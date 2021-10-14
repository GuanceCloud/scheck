package checker

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/internal/global"

	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/internal/luafuncs"
)

var (
	pool *statePool
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
		onceTasks:       make(map[string]*Rule),
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
	checkfiles := make(map[string]bool)
	for _, file := range files {
		/*
			文件夹直接忽略
			因为 每次当lua脚本运行需要加载lib库时 都是实时加载lib中的文件，可以做到实时更新
			lua规则文件需要检测和热更
		*/
		if file.IsDir() {
			continue
		}
		path := filepath.Join(ruleDir, file.Name())
		r := newRule(path)

		if strings.HasSuffix(file.Name(), global.LuaExt) {
			if isCustom {
				rulename := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
				checkfiles[rulename] = true
				modifyTime := fileModify(path)
				if t, ok := scheduler.customRulesTime[rulename]; ok {
					if t == modifyTime {
						continue
					}
				}
				l.Debugf("users rules is  change name = %s", rulename)
				scheduler.customRulesTime[rulename] = modifyTime
			}
			if err := r.load(); err != nil {
				l.Errorf("load from dir err=%v", err)
				continue
			}
			if !r.disabled {
				if isCustom {
					if !scheduler.checkName(r.Name) {
						l.Warnf("scheck: is forbidden to have the same name,and the file in %s", filepath.Join(global.InstallDir, global.DefRulesDir))
						continue
					}
				}
				scheduler.addRule(r)
				luafuncs.Add(r.Name, r.manifest.Category, r.interval, isCustom)
			}
		}
	}
	if isCustom && len(checkfiles) < len(scheduler.customRulesTime) {
		scheduler.deleteUserRule(checkfiles)
	}
	if len(scheduler.tasks) == 0 {
		l.Warnf("There are no rules in the folder to load! . system exit !!!")
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

// doAndReset: reset the rule next runtime
func (scheduler *TaskScheduler) doAndReset(key string) {
	scheduler.lock.Lock()
	defer scheduler.lock.Unlock()
	if rule, ok := scheduler.tasks[key]; ok {
		// 20211009 slq ：当前时间+间隔时间+运行使用时间，这样的优点是 使规则运行趋于平滑，减少同一时间同时执行的数量
		rule.RunTime = time.Now().UnixNano()/1e6 + rule.interval + rule.useTime
		scheduler.tasks[key] = rule
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
		rule, key := scheduler.GetTask()
		runTime := rule.RunTime
		next := runTime - now.UnixNano()/1e6
		if next <= 0 {
			if rule != nil {
				state := pool.getState()
				go rule.RunJob(state)
			}
			scheduler.doAndReset(key)
			continue
		}
		timer := time.NewTimer(time.Millisecond * time.Duration(next))
		l.Debugf("scheduler new time.timer, at %d millisecond start...", next)
		for {
			select {
			case <-timer.C:
				state := pool.getState()
				scheduler.doAndReset(key)
				if rule != nil {
					go rule.RunJob(state)
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
	var onces = len(scheduler.onceTasks)
	if onces == 0 {
		l.Warnf("schedules:the long term ruler task is empty....")
		return
	}
	cxtMap := sync.Map{}                // 主动停止通知信号
	errChan := make(chan string, onces) // 被动停止通知信号
	count := 0                          // 运行的数量
	for _, rule := range scheduler.onceTasks {
		cxt, cancel := context.WithCancel(context.Background())
		go rule.RunOnce(cxt, errChan)
		cxtMap.Store(rule.Name, cancel)
		count++
	}
	l.Infof("Long term type lua was running,len is %d", onces)
	for {
		select {
		case <-scheduler.stop:
			l.Error("receive exit signal ,stop all lua script")
			// all context stop
			cxtMap.Range(func(key, value interface{}) bool {
				if cancel, ok := value.(context.CancelFunc); ok {
					cancel()
				}
				return false
			})
		case <-time.After(time.Minute):
			// 检查运行数量
			if onces != count {
				// do something...
				l.Warnf("Unexpected reduction in number of runs !!! tot=%d run=%d", onces, count)
				onces = count // The warn message will always remind if onces != count
			}
			if count == 0 {
				close(errChan)
				return
			}
		case name := <-errChan:
			// to call monitor or trigger
			count--
			if count < 0 {
				l.Errorf("no rule is running!")
				return
			}
			l.Errorf("rule name = %s is stop!!!", name)
		}
	}
}

func (scheduler *TaskScheduler) checkName(name string) bool {
	if scheduler.manifests != nil {
		for tName := range scheduler.manifests {
			if tName == name {
				return false
			}
		}
	}
	if !checkID(name) {
		l.Warnf("ID=%s check is illegal. custom rule must be greater than 10000 !", name)
	}
	return true
}

func (scheduler *TaskScheduler) GetRuleByName(filename string) *Rule {
	for key, task := range scheduler.tasks {
		if key == filename {
			return task
		}
	}
	for key, task := range scheduler.onceTasks {
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
	if r.cron == "" || r.cron == global.LuaCronDisable {
		scheduler.onceTasks[r.Name] = r
	} else {
		scheduler.tasks[r.Name] = r
	}
	scheduler.manifests[r.Name] = r.manifest
}

// hotUpdate :timing to Update users rules if file has change
func (scheduler *TaskScheduler) hotUpdate() {
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

// deleteUserRule: delete rule and delete monitor
func (scheduler *TaskScheduler) deleteUserRule(checkFiles map[string]bool) {
	scheduler.lock.Lock()
	defer scheduler.lock.Unlock()
	for name := range scheduler.customRulesTime {
		if _, ok := checkFiles[name]; !ok {
			if _, ok := scheduler.tasks[name]; ok {
				l.Infof("delete rule from task,name=%s", name)
				delete(scheduler.tasks, name)
			}
			if _, ok := scheduler.onceTasks[name]; ok {
				l.Infof("delete rule from onceTask,name=%s", name)
				delete(scheduler.tasks, name)
			}
			luafuncs.Delete(name)
		}
	}
}

func GetRuleNum() int {
	if Chk != nil && Chk.taskScheduler != nil {
		return len(Chk.taskScheduler.tasks) + len(Chk.taskScheduler.onceTasks)
	}
	return 0
}

func checkID(luaName string) bool {
	names := strings.Split(luaName, "-")
	ID := names[0]
	id, err := strconv.Atoi(ID)
	if err != nil {
		return false
	}
	if id < global.LuaCustomIDStart {
		return false
	}
	return true
}

// fileModify:return modify times sum of lua file and manifest file.
func fileModify(filePath string) int64 {
	fileInfo, _ := os.Stat(filePath)
	ruleModify := fileInfo.ModTime().Unix()
	ruledir := filepath.Dir(filePath)
	rulename := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))

	manifestPath := filepath.Join(ruledir, rulename+global.LuaManifestExt)
	MfileInfo, err := os.Stat(manifestPath)
	if err != nil {
		l.Error("cannot find manifest file ,lua and manifest  must exist !!!")
		return 0
	}
	manifestModify := MfileInfo.ModTime().Unix()
	return ruleModify + manifestModify
}
