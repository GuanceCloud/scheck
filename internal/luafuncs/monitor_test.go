package luafuncs

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	termmarkdown "github.com/MichaelMure/go-term-markdown"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/internal/global"
)

// nolint
func TestExportAsMD(t *testing.T) {
	hostName, _ := os.Hostname()
	rsm := RunStatusMonitor{
		HostName:      hostName,
		OsArch:        global.LocalGOOS + "/" + global.LocalGOARCH,
		SCVersion:     global.Version,
		Scripts:       make([]*OutType, 0),
		ScriptsSortBy: "count",
		LuaStatF:      global.LuaStatusFile,
	}
	fmtTatal := fmt.Sprintf(title,
		rsm.HostName, rsm.OsArch, rsm.SCVersion, "", "systemRunTime",
		0, 0, 0, rsm.ScriptsSortBy)
	rows := make([]string, 0)
	sc := OutType{
		Name:        "0001",
		Category:    "system",
		Status:      "ok",
		RuntimeAvg:  "100",
		RuntimeMax:  "100",
		RuntimeMin:  "100",
		LastRuntime: "100",
		RunCount:    100,
		ErrCount:    0,
		TriggerNum:  0,
	}
	for i := 0; i < 10; i++ {
		rows = append(rows,
			fmt.Sprintf(format,
				sc.Name, sc.Category, sc.Status, sc.RuntimeAvg, sc.RuntimeMax, sc.RuntimeMin,
				sc.LastRuntime, sc.RunCount, sc.ErrCount, sc.TriggerNum))
	}
	tot := fmtTatal + temp + strings.Join(rows, "\n")
	if tot == "" {
		l.Errorf("lua status is null ,wait 5 minter")
		return
	}
	// tot += fmt.Sprintf(end, filepath.Join(global.InstallDir, "htmlFile"))

	t.Log(string(termmarkdown.Render(tot, 80, 2)))
}

func TestMonitor_saveData(t *testing.T) {
	testdata := "testdata/monitor.json"
	monitor := &Monitor{
		jsonFile: "testdata/monitor-out.test",
	}
	_, _ = os.Create("testdata/monitor-out.test")
	// 合并数据 需要monitor进行初始化map[string]script
	bts, err := ioutil.ReadFile(testdata)
	if err != nil {
		t.Logf("readFile err = %v", err)
		return
	}
	err = json.Unmarshal(bts, monitor)
	if err != nil {
		t.Logf("Unmarshal err = %v", err)
		return
	}
	monitor.saveData()
}

/*
func TestExportAsMD1(t *testing.T) {
	input := []byte("### 5lmh.com是个不错的go文档网站")
	unsafe := blackfriday.MarkdownCommon(input)
	fmt.Println(string(unsafe))
	// html := bluemonday.UGCPolicy().SanitizeBytes(unsafe)
	// fmt.Println(string(html))
}
*/
