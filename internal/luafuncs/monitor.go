package luafuncs

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	term_markdown "github.com/MichaelMure/go-term-markdown"
	"github.com/dustin/go-humanize"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/git"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/internal/global"
	"golang.org/x/term"
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
	title = `## scheck lua脚本在运行情况
- scheck 主机名：%s 当前操作系统%s
- scheck 当前的版本为%s 发布时间%s
- scheck 已经运行了%s
- 当前共有%d个lua脚本 其中自带的有%d个 属于用户自定义的有%d个
- 当前排序的方式为%s,排序方式可分为三种：运行次数(-count),用时(-time),名称(-name)。默认按照运行次数排序。

### 以下为各个lua脚本运行情况：
`

	temp = `
| lua名称 | 类型 | 状态 | 平均用时   | 最大用时 | 最小用时 | 上次运行时间 | 总运行次数  | 错误次数 | 上报次数 |
| ----   | :----:   | :----:   | :----: | :----:       | :----:       | :----: | :---: | :----:   | :---:    |
`

	format = "|`%s`|%s|%s|%s|%s|%s|%s|%d|%d|%d|"

	end = "\n > lua scripts运行情况放在文件: `%s` 文件的格式是markdown, `%s`文件格式为html 可用过编译器或者浏览器等打开"
)

var (
	monitor *Monitor
	l       = logger.DefaultSLogger("internal.lua")
)

type Script struct {
	Name        string `json:"name"`
	Category    string `json:"category"`
	Status      string `json:"status"`
	RuntimeAvg  int64  `json:"runtime_avg"`
	RuntimeMax  int64  `json:"runtime_max"`
	RuntimeMin  int64  `json:"runtime_min"`
	LastRuntime int64  `json:"last_runtime"`
	RunCount    int    `json:"run_count"`
	ErrCount    int    `json:"err_count"`
	TriggerNum  int    `json:"trigger_num"`
	Interval    int64  `json:"interval"`
	runTimes    []int64
	isOnce      bool
}

type Monitor struct {
	Scripts   map[string]*Script `json:"scripts"`
	StartTime time.Time          `json:"start_time"`
	RuleNum   int                `json:"rule_num"`
	CustomNum int                `json:"custom_num"`
	outFile   string
	jsonFile  string
	lock      sync.Mutex
}

func newMonitor() *Monitor {
	m := &Monitor{
		Scripts:   make(map[string]*Script),
		outFile:   global.LuaStatusOutFileMD,
		jsonFile:  global.LuaStatusFile,
		StartTime: time.Now(),
	}
	_ = os.Remove(m.jsonFile)
	_, err := os.Create(m.jsonFile)
	if err != nil {
		l.Errorf("creat file err=%v", err)
	}
	return m
}

// Start:监控每一个lua的运行情况
func Start() {
	l = logger.SLogger("internal.lua")
	monitor = newMonitor()
	go monitor.timeToSave(global.LuaStatusWriteFileInterval)
}

func newScriptStatus(name, category string, interval int64) *Script {
	return &Script{
		Name:     name,
		Category: category,
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
		bts, err := ioutil.ReadFile(m.jsonFile)
		if err != nil {
			l.Errorf("err=%v", err)
			continue
		}
		if len(bts) == 0 {
			bts, err = m.MonitorMarshal()
			if err != nil {
				l.Errorf("marshal err =%v", err)
				continue
			}
			err = ioutil.WriteFile(m.jsonFile, bts, global.FileModeRW)
			if err != nil {
				l.Errorf("write err =%v", err)
			}
			continue
		}

		l.Debug("merge script...")
		oldMonitor := new(Monitor)
		err = json.Unmarshal(bts, oldMonitor)
		if err != nil {
			l.Errorf("marshal err =%v", err)
			continue
		}
		m.lock.Lock()
		m.mergeOld(oldMonitor.Scripts)
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

func (m *Monitor) mergeRuntime() {
	for name, nsc := range m.Scripts {
		tot := int64(0)
		max := int64(0)
		min := int64(0)
		for i, rt := range nsc.runTimes {
			if i == 0 && min == 0 {
				min = rt
			}
			if rt < min {
				min = rt
			}
			if rt > max {
				max = rt
			}
			tot += rt
		}
		if tot != 0 {
			nsc.RuntimeAvg = tot / int64(len(nsc.runTimes))
		}
		nsc.RuntimeMax = max
		nsc.RuntimeMin = min
		m.Scripts[name] = nsc
	}
}

func (m *Monitor) MonitorMarshal() ([]byte, error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	bts, err := json.MarshalIndent(m, "", "	")
	for name, script := range m.Scripts {
		m.Scripts[name] = newScriptStatus(script.Name, script.Category, script.Interval)
	}
	if err != nil {
		return nil, err
	}
	return bts, nil
}

func (m *Monitor) mergeOld(oldScripts map[string]*Script) {
	m.mergeRuntime()
	for name, nsc := range m.Scripts {
		tot := nsc.RuntimeAvg * int64(nsc.RunCount)
		max := nsc.RuntimeMax
		min := nsc.RuntimeMin
		osc, ok := oldScripts[name]
		if !ok {
			// 从文件中读取后 没有该脚本的运行数据，就以新的数据为准，在这里直接返回等待下一次
			continue
		}
		if oldScripts[name].RuntimeMax > max {
			max = oldScripts[name].RuntimeMax
		}
		if oldScripts[name].RuntimeMin != 0 && oldScripts[name].RuntimeMin < min {
			min = oldScripts[name].RuntimeMin
		}
		tot += osc.RuntimeAvg * int64(osc.RunCount)

		nsc.RuntimeMax = max
		nsc.RuntimeMin = min
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
func Add(name, category string, interval int64, isCustom bool) {
	ss := newScriptStatus(name, category, interval)
	if monitor == nil {
		return
	}
	monitor.lock.Lock()
	defer monitor.lock.Unlock()
	monitor.Scripts[ss.Name] = ss
	if isCustom {
		monitor.CustomNum++
	} else {
		monitor.RuleNum++
	}
	l.Debugf("name = %s :add to monitor", name)
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
		l.Debugf(" %s update to monitor", name)
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
	Category    string
	Status      string
	RuntimeAvg  string
	RuntimeMax  string
	RuntimeMin  string
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
	Scripts       []*OutType
	ScriptsSortBy string
	LuaStatF      string
}

func (rsm *RunStatusMonitor) Len() int {
	return len(rsm.Scripts)
}

func (rsm *RunStatusMonitor) Swap(i, j int) {
	rsm.Scripts[i], rsm.Scripts[j] = rsm.Scripts[j], rsm.Scripts[i]
}

func (rsm *RunStatusMonitor) Less(i, j int) bool {
	switch rsm.ScriptsSortBy {
	case "", "count":
		return rsm.Scripts[j].RunCount < rsm.Scripts[i].RunCount
	case "time":
		return rsm.Scripts[j].RuntimeMax < rsm.Scripts[i].RuntimeMax
	default:
	}
	return rsm.Scripts[j].RunCount < rsm.Scripts[i].RunCount
}

func (rsm *RunStatusMonitor) getStatus() (out string) {
	l.Debug("to export of all lua run status...")
	bts, err := ioutil.ReadFile(rsm.LuaStatF)
	if err != nil {
		l.Errorf("readFile err=%v", err)
		return ""
	}
	monitor := new(Monitor)
	err = json.Unmarshal(bts, monitor)
	if err != nil {
		l.Errorf("marshal err=%v", err)
		return ""
	}
	if len(monitor.Scripts) == 0 {
		l.Errorf("lua scripts lens is 0")
		return ""
	}
	// 将scripts放到RunStatusMonitor 中
	now := time.Now()

	rows := make([]string, 0)
	for _, script := range monitor.Scripts {
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
		timeMin := time.Duration(script.RuntimeMin).String()

		out := &OutType{Name: script.Name,
			Status:      script.Status,
			Category:    script.Category,
			RuntimeAvg:  timeAvg,
			RuntimeMax:  timeMax,
			RuntimeMin:  timeMin,
			RunCount:    script.RunCount,
			ErrCount:    script.ErrCount,
			TriggerNum:  script.TriggerNum,
			LastRuntime: last,
			Interval:    script.Interval}
		rsm.Scripts = append(rsm.Scripts, out)
	}
	systemRunTime := humanize.RelTime(monitor.StartTime, now, "ago", "")
	fmtTatal := fmt.Sprintf(title,
		rsm.HostName, rsm.OsArch, rsm.SCVersion, git.BuildAt, systemRunTime,
		monitor.RuleNum+monitor.CustomNum, monitor.RuleNum, monitor.CustomNum, rsm.ScriptsSortBy)

	sort.Sort(rsm)
	for i := 0; i < len(rsm.Scripts); i++ {
		sc := rsm.Scripts[i]
		rows = append(rows,
			fmt.Sprintf(format,
				sc.Name, sc.Category, sc.Status, sc.RuntimeAvg, sc.RuntimeMax, sc.RuntimeMin,
				sc.LastRuntime, sc.RunCount, sc.ErrCount, sc.TriggerNum))
	}
	if rsm.ScriptsSortBy == "name" {
		sort.Strings(rows)
	}
	out = fmtTatal + temp + strings.Join(rows, "\n")
	return out
}

// ExportAsMD :从文件中读取数据，整理后输出到文件并且打印出来
func ExportAsMD(sortBy string) {
	mdFile := fmt.Sprintf(global.LuaStatusOutFileMD, time.Now().Format("20060102-150405"))
	htmlFile := fmt.Sprintf(global.LuaStatusOutFileHTML, time.Now().Format("20060102-150405"))
	if sortBy == "" {
		sortBy = global.LuaSortByCount
	}
	hostName, _ := os.Hostname()
	rsm := RunStatusMonitor{
		HostName:      hostName,
		OsArch:        global.LocalGOOS + "/" + global.LocalGOARCH,
		SCVersion:     global.Version,
		Scripts:       make([]*OutType, 0),
		ScriptsSortBy: strings.ToLower(sortBy),
		LuaStatF:      global.LuaStatusFile,
	}

	tot := rsm.getStatus()
	if tot == "" {
		l.Errorf("lua status is null ,wait 5 minter")
		return
	}
	tot += fmt.Sprintf(end, mdFile, htmlFile)

	// write to (md/html) file
	err := ioutil.WriteFile(htmlFile, getHTML(tot), global.FileModeRW)
	if err != nil {
		l.Errorf("write to file err=%v", err)
	}
	_ = ioutil.WriteFile(mdFile, getHTML(tot), global.FileModeRW)

	width := 100
	if term.IsTerminal(0) {
		if width, _, err = term.GetSize(0); err != nil {
			width = 100
		}
	}

	leftPad := 2
	fmt.Println(string(term_markdown.Render(tot, width, leftPad)))
}

func getHTML(tot string) []byte {
	psr := parser.NewWithExtensions(parser.CommonExtensions)
	htmlFlags := html.CommonFlags | html.HrefTargetBlank | html.CompletePage
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)
	return markdown.ToHTML([]byte(tot), psr, renderer)
}
