package luafuncs

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/dustin/go-humanize"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/internal/global"
	"golang.org/x/tools/go/ssa/interp/testdata/src/fmt"
)

/*
lua 脚本在运行时候
	需要统计信息，包括：
		- 运行情况 ：正常 无法运行 运行时崩溃 配置错误
		- 单次运行时间 平均时间 最长时间 运行次数 上一次运行时间
		- tigger 次数
		-

运行时统计 定时写入本地文件，可通过 scheck -runstatus 命令 ，使用md模版生成文档或直接打印出来。
*/

const (
	temp = `
| lua | 状态 | 平均用时   | 最大用时 | 上次运行时间 | 总运行次数  | 错误次数 | 上报次数 |
| ----   | :----:   | :----: | :----:       | :----: | :---: | :----:   | :---:    |
`
	tatal = `### scheck lua脚本在运行情况
`
	format = "|`%s`|%s|%s|%d|%d|%d|%d|%d|\n"
)

var (
	monitor *RunStatusMonitor
	l       = logger.DefaultSLogger("internal.lua")
)

type scripts map[string]*Script

type Script struct {
	Name        string          `json:"name"`
	Status      string          `json:"status"`
	runTimes    []time.Duration `json:"-"`
	RuntimeAvg  time.Duration   `json:"runtime_avg"`
	RuntimeMax  time.Duration   `json:"runtime_max"`
	LastRuntime time.Time       `json:"last_runtime"`
	RunCount    int             `json:"run_count"`
	ErrCount    int             `json:"err_count"`
	TriggerNum  int             `json:"trigger_num"`
	Interval    int64           `json:"-"`
}

// RunStatusMonitor: 可导出的状态图结构体
type RunStatusMonitor struct {
	HostName  string
	OsArch    string
	SCVersion string
	RunTime   time.Time
	Scripts   scripts
	lock      sync.Mutex
}

func NewRunStatusMonitor() *RunStatusMonitor {
	hostName, _ := os.Hostname()
	return &RunStatusMonitor{
		HostName:  hostName,
		OsArch:    global.LocalGOOS + "/" + global.LocalGOARCH,
		SCVersion: global.Version,
		Scripts:   make(map[string]*Script),
	}
}

func Start() {
	l = logger.SLogger("internal.lua")
	initFile()
	monitor = NewRunStatusMonitor()
	// time.ticker:to sum avg....
	go monitor.timeToSave()
}

func initFile() {
	_ = os.Remove(global.LuaStatusFile)
	_ = os.Remove(global.LuaStatusOutFile)
}

func newSctiptStatus(name string, isOnce bool, interval int64) *Script {
	if isOnce {
		interval = 0
	}
	return &Script{
		Name:     name,
		Interval: interval,
		runTimes: make([]time.Duration, 0),
	}
}

/*
排序 ：按照平均用时
func (a RunStatusMonitor) Len() int {    // 重写 Len() 方法
	return len(a)
}
func (a RunStatusMonitor) Swap(i, j int){     // 重写 Swap() 方法
	a[i], a[j] = a[j], a[i]
}
func (a RunStatusMonitor) Less(i, j int) bool {    // 重写 Less() 方法， 从大到小排序
	return a[j].Age < a[i].Age
}
*/

// 定时 写入文件
func (m *RunStatusMonitor) timeToSave() {
	// 先读 合并数据集 再写
	ticker := time.NewTicker(global.LuaStatusWriteFileInterval)
	for range ticker.C {
		f, err := os.OpenFile(global.LuaStatusFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, global.FileModeRW)
		if err != nil {
			continue
		}
		bts, err := ioutil.ReadAll(f)
		if err != nil {
			l.Errorf("err", err)
			_ = f.Close()
			continue
		}
		if len(bts) == 0 {
			wbts, err := json.Marshal(monitor.Scripts)
			if err != nil {
				l.Errorf("marshal err =%v", err)
				continue
			}
			_, err = f.Write(wbts)
			if err != nil {
				l.Errorf("write err =%v", err)
			}
			_ = f.Close()
			continue
		}
		// 合并数据
		oldScript := make(map[string]*Script)
		err = json.Unmarshal(bts, &oldScript)
		if err != nil {
			l.Errorf("marshal err =%v", err)
			continue
		}

	}
}

// todo
func merge(oldScript *Script) *Script {

	return nil
}

// 初始化一个。。。
func Add(name string, isOnce bool, interval int64) {
	ss := newSctiptStatus(name, isOnce, interval)
	setToMonitor(ss)
}

func setToMonitor(ss *Script) {
	if monitor == nil {
		return
	}
	monitor.lock.Lock()
	defer monitor.lock.Unlock()
	monitor.Scripts[ss.Name] = ss
}

func updateStatus(name string, runTime time.Duration, err error) {
	if monitor != nil {
		monitor.lock.Lock()
		defer monitor.lock.Unlock()
		ss, ok := monitor.Scripts[name]
		if !ok {
			monitor.Scripts[name] = newSctiptStatus(name, false, 0)
			return
		}
		if err != nil {
			ss.ErrCount++
		}
		ss.RunCount++
		ss.runTimes = append(ss.runTimes, runTime)
		monitor.Scripts[name] = ss
	}
}

func UpdateTriggerCount(name string) {
	if monitor != nil {
		monitor.lock.Lock()
		defer monitor.lock.Unlock()
		if _, ok := monitor.Scripts[name]; ok {
			monitor.Scripts[name].TriggerNum++
		}
	}

}

// ExportAsMD :从文件中读取数据，整理后输出到文件并且打印出来
func ExportAsMD() string {
	l.Debug("to export of all lua run status...")
	bts, err := ioutil.ReadFile(global.LuaStatusFile)
	if err != nil {
		l.Errorf("readFile err=%v", err)
		return ""
	}
	scripts := make(map[string]*Script)
	err = json.Unmarshal(bts, &scripts)
	if err != nil {
		l.Errorf("marshal err=%v", err)
		return ""
	}
	if len(scripts) == 0 {
		l.Errorf("lua scripts lens is 0")
		return ""
	}

	now := time.Now()
	rows := make([]string, 0)
	for _, script := range scripts {
		// lua | 状态 | 平均用时   | 最大用时 | 上次运行时间 | 总运行次数  | 错误次数 | 上报次数
		timeAvg := int64(0)
		totLen := int64(len(script.runTimes))
		totNum := time.Duration(0)
		for _, rt := range script.runTimes {
			totNum += rt
		}
		last := humanize.RelTime(script.LastRuntime, now, "ago", "")
		timeAvg = int64(totNum) / totLen
		rows = append(rows, fmt.Sprintf(format,
			script.Name, "ok", time.Duration(timeAvg).String(), script.RuntimeMax, last, script.RunCount, script.ErrCount, script.TriggerNum))
	}
	sort.Strings(rows)
	// to .... file
	tot := tatal + temp + strings.Join(rows, "\n")
	outFile := fmt.Sprintf(global.LuaStatusOutFile, time.Now().Format("20060102-150405"))
	err = ioutil.WriteFile(outFile, []byte(tot), global.FileModeRW)
	if err != nil {
		l.Errorf("write to file err=%v", err)
	}
	return tot
}
