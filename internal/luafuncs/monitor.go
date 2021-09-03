package luafuncs

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/dustin/go-humanize"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/git"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/internal/global"
)

/*
lua 脚本在运行时候
	需要统计信息，包括：
		- 运行情况 ：正常 无法运行 运行时崩溃 配置错误
		- 单次运行时间 平均时间 最长时间 运行次数 上一次运行时间
		- tigger 次数
		- 错误次数
运行时统计 定时写入本地文件，可通过 scheck -luastatus 命令 ，使用md模版生成文档或直接打印出来。
*/

const (
	tatal = `## scheck lua脚本在运行情况
- scheck 主机名：%s 当前操作系统%s
- scheck 当前的版本为%s 发布时间%s
- 当前共有%d个lua脚本 其中自带的有%d个 属于用户自定义的有%d个
- 当前排序的方式为%s,排序方式可分为：运行次数(-count),用时(-time),名称(-name)。默认按照运行次数排序。

### 以下为各个lua脚本运行情况：
`

	temp = `
| lua名称 | 状态 | 平均用时   | 最大用时 | 上次运行时间 | 总运行次数  | 错误次数 | 上报次数 |
| ----   | :----:   | :----: | :----:       | :----: | :---: | :----:   | :---:    |
`

	format = "|`%s`|%s|%s|%s|%s|%d|%d|%d|"

	end = "> lua scripts运行情况放在文件 `%s` 中，文件的格式是markdown,可用过编译器或者浏览器等打开"
)

var (
	monitor *Monitor
	l       = logger.DefaultSLogger("internal.lua")
)

type Script struct {
	Name        string `json:"name"`
	Status      string `json:"status"`
	RuntimeAvg  int64  `json:"runtime_avg"`
	RuntimeMax  int64  `json:"runtime_max"`
	LastRuntime int64  `json:"last_runtime"`
	RunCount    int    `json:"run_count"`
	ErrCount    int    `json:"err_count"`
	TriggerNum  int    `json:"trigger_num"`
	Interval    int64  `json:"interval"`
	runTimes    []int64
	isOnce      bool
}

type Monitor struct {
	Scripts  map[string]*Script
	outFile  string
	jsonFile string
	Auto     bool // 保留
	lock     sync.Mutex
}

func newMonitor(outFile string) *Monitor {
	if outFile == "" {
		outFile = global.LuaStatusOutFile
	}
	m := &Monitor{
		Scripts:  make(map[string]*Script),
		outFile:  outFile,
		jsonFile: filepath.Join(global.InstallDir, global.LuaStatusFile),
	}
	_ = os.Remove(m.jsonFile)
	_, err := os.Create(m.jsonFile)
	if err != nil {
		l.Errorf("creat file err=%v", err)
	}
	return m
}

// Start:监控每一个lua的运行情况
func Start(outF string) {
	l = logger.SLogger("internal.lua")
	monitor = newMonitor(outF)
	go monitor.timeToSave(global.LuaStatusWriteFileInterval)
}

func newScriptStatus(name string, interval int64) *Script {
	return &Script{
		Name:     name,
		Status:   "ok",
		isOnce:   interval < 0,
		Interval: interval,
		runTimes: make([]int64, 0),
	}
}

// 定时 写入文件
func (m *Monitor) timeToSave(tickTime time.Duration) {
	ticker := time.NewTicker(tickTime)
	for range ticker.C {
		l.Debugf("进入 定时器")
		bts, err := ioutil.ReadFile(m.jsonFile)
		if err != nil {
			l.Errorf("err", err)
			continue
		}
		if len(bts) == 0 {
			wbts, err := m.MonitorMarshal()
			if err != nil {
				l.Errorf("marshal err =%v", err)
				continue
			}
			err = ioutil.WriteFile(m.jsonFile, wbts, global.FileModeRW)
			if err != nil {
				l.Errorf("write err =%v", err)
			}
			continue
		}

		l.Debug("merge script...")
		oldScript := make(map[string]*Script)
		err = json.Unmarshal(bts, &oldScript)
		if err != nil {
			l.Errorf("marshal err =%v", err)
			continue
		}
		m.lock.Lock()
		m.merge(oldScript)
		m.lock.Unlock()

		newBts, err := m.MonitorMarshal()
		if err != nil {
			l.Errorf("err=%v", err)
			continue
		}
		if err = ioutil.WriteFile(m.jsonFile, newBts, global.FileModeRW); err != nil {
			l.Errorf("err=%v", err)
		}
	}
}

func (m *Monitor) MonitorMarshal() ([]byte, error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	wbts, err := json.Marshal(m.Scripts)
	for name, script := range m.Scripts {
		m.Scripts[name] = newScriptStatus(script.Name, script.Interval)
	}
	if err != nil {
		return nil, err
	}
	return wbts, nil
}

func (m *Monitor) merge(oldScripts map[string]*Script) {
	for name, nsc := range m.Scripts {
		tot := int64(0)
		max := int64(0)
		osc, ok := oldScripts[name]
		if !ok {
			// 从文件中读取后 没有该脚本的运行数据，就以新的数据为准
			continue
		}
		for _, rt := range nsc.runTimes {
			if rt > max {
				max = rt
			}
			tot += rt
		}
		if oldScripts[name].RuntimeMax > max {
			max = oldScripts[name].RuntimeMax
		}
		tot += osc.RuntimeAvg * int64(osc.RunCount)

		nsc.RuntimeMax = max
		nsc.RunCount += osc.RunCount
		nsc.ErrCount += osc.ErrCount
		nsc.TriggerNum += osc.TriggerNum
		if nsc.LastRuntime == 0 && osc.LastRuntime != 0 {
			nsc.LastRuntime = osc.LastRuntime
		}
		if nsc.RunCount != 0 {
			// avg = ((old.avg * old.cont) + runTimes.tot)/(old count + runTimes.len)
			nsc.RuntimeAvg = tot / int64(nsc.RunCount)
		}
		m.Scripts[name] = nsc
	}
}

// Add:all rule add to monitor.
func Add(name string, interval int64) {
	ss := newScriptStatus(name, interval)
	if monitor == nil {
		return
	}
	monitor.lock.Lock()
	defer monitor.lock.Unlock()
	monitor.Scripts[ss.Name] = ss
	l.Debug("add to monitor")
}

func UpdateStatus(name string, runTime time.Duration, isErr bool) {
	if monitor != nil {
		monitor.lock.Lock()
		defer monitor.lock.Unlock()
		ss, ok := monitor.Scripts[name]
		if !ok {
			l.Errorf("lua name=%s not in monitor!", name)
			return
		}
		if isErr {
			ss.ErrCount++
		}
		ss.RunCount++
		ss.runTimes = append(ss.runTimes, int64(runTime))
		ss.LastRuntime = time.Now().Unix()
		monitor.Scripts[name] = ss
		l.Debug("update to monitor")
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

type OutType struct {
	Name        string
	Status      string
	RuntimeAvg  string
	RuntimeMax  string
	RunCount    int
	ErrCount    int
	TriggerNum  int
	LastRuntime string // 只有在最后一次才会添加
	Interval    int64  // 同上
}

// RunStatusMonitor: 可导出的状态图结构体
type RunStatusMonitor struct {
	HostName      string
	OsArch        string
	SCVersion     string
	BuildTime     string
	RunTime       time.Time
	hostNum       int
	customNum     int
	Scripts       []*OutType
	ScriptsSortBy string
}

func (rsm RunStatusMonitor) Len() int {
	return len(rsm.Scripts)
}

func (rsm RunStatusMonitor) Swap(i, j int) {
	rsm.Scripts[i], rsm.Scripts[j] = rsm.Scripts[j], rsm.Scripts[i]
}

func (rsm RunStatusMonitor) Less(i, j int) bool {
	switch rsm.ScriptsSortBy {
	case "", "count":
		return rsm.Scripts[j].RunCount < rsm.Scripts[i].RunCount
	case "time":
		return rsm.Scripts[j].RuntimeMax < rsm.Scripts[i].RuntimeMax
	default:
	}
	return rsm.Scripts[j].RunCount < rsm.Scripts[i].RunCount
}

// ExportAsMD :从文件中读取数据，整理后输出到文件并且打印出来
func ExportAsMD(sortBy string) string {
	l.Debug("to export of all lua run status...")
	sortBy = strings.ToLower(sortBy)
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
	// 将scripts放到RunStatusMonitor 中
	now := time.Now()
	hostName, _ := os.Hostname()
	runstatus := &RunStatusMonitor{
		HostName:      hostName,
		OsArch:        global.LocalGOOS + "/" + global.LocalGOARCH,
		SCVersion:     global.Version,
		Scripts:       make([]*OutType, 0),
		ScriptsSortBy: sortBy,
	}

	rows := make([]string, 0)
	for _, script := range scripts {
		if script.Interval < 0 {
			continue
		}
		// lua | 状态 | 平均用时   | 最大用时 | 上次运行时间 | 总运行次数  | 错误次数 | 上报次数
		last := "-"
		if script.LastRuntime != 0 {
			last = humanize.RelTime(time.Unix(script.LastRuntime, 0), now, "ago", "")
		}

		timeAvg := time.Duration(script.RuntimeAvg).String()
		timeMax := time.Duration(script.RuntimeMax).String()
		out := &OutType{Name: script.Name,
			Status:      script.Status,
			RuntimeAvg:  timeAvg,
			RuntimeMax:  timeMax,
			RunCount:    script.RunCount,
			ErrCount:    script.ErrCount,
			TriggerNum:  script.TriggerNum,
			LastRuntime: last,
			Interval:    script.Interval}
		runstatus.Scripts = append(runstatus.Scripts, out)
	}
	fmtTatal := fmt.Sprintf(tatal,
		runstatus.HostName, runstatus.OsArch, runstatus.SCVersion, git.BuildAt,
		runstatus.hostNum+runstatus.customNum, runstatus.hostNum, runstatus.customNum, runstatus.ScriptsSortBy)

	sort.Sort(runstatus)
	for i := 0; i < len(runstatus.Scripts); i++ {
		sc := runstatus.Scripts[i]
		rows = append(rows, fmt.Sprintf(format, sc.Name, sc.Status, sc.RuntimeAvg, sc.RuntimeMax, sc.LastRuntime, sc.RunCount, sc.ErrCount, sc.TriggerNum))
	}
	if runstatus.ScriptsSortBy == "name" {
		sort.Strings(rows)
	}
	tot := fmtTatal + temp + strings.Join(rows, "\n")

	outFile := fmt.Sprintf(global.LuaStatusOutFile, time.Now().Format("20060102-150405"))
	tot += fmt.Sprintf(end, filepath.Join(global.InstallDir, outFile))
	// to .... file
	err = ioutil.WriteFile(outFile, []byte(tot), global.FileModeRW)
	if err != nil {
		l.Errorf("write to file err=%v", err)
	}
	return tot
}
