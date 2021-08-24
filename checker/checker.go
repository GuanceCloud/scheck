package checker

import (
	"context"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	lua "github.com/yuin/gopher-lua"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/config"
	_ "gitlab.jiagouyun.com/cloudcare-tools/sec-checker/funcs/system"
	_ "gitlab.jiagouyun.com/cloudcare-tools/sec-checker/funcs/utils"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/output"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/service/cgroup"
)

var (
	Chk *Checker
	l   = logger.DefaultSLogger("check")
)

type (
	Checker struct {
		rulesDir      string
		customRuleDir string //用户自定义的路径
		taskScheduler *TaskScheduler
		ruleMux       sync.Mutex
		manifestMux   sync.Mutex
		loading       int32
	}
)

// Start

func Start(ctx context.Context, confSys *config.System, outputpath *config.ScOutput) {
	l = logger.SLogger("checker")

	Chk = newChecker(confSys.RuleDir, confSys.CustomRuleDir, confSys.LuaHotUpdate)

	output.Start(outputpath)
	Chk.start(ctx)
}

func newChecker(rulesDir, customRuleDir string, hotUpdate bool) *Checker {
	lua.LuaPathDefault = filepath.Join(rulesDir, "/lib/?.lua")

	c := &Checker{
		rulesDir:      rulesDir,
		customRuleDir: customRuleDir,
		taskScheduler: NewTaskScheduler(rulesDir, customRuleDir, hotUpdate),
	}

	InitStatePool(15, 20)

	return c
}

func (c *Checker) doDelRule(r *Rule) {
	c.taskScheduler.removeRule(r)
}

func GetManifestByName(fileName string) (*RuleManifest, error) {
	if Chk != nil && Chk.taskScheduler != nil {
		rule := Chk.taskScheduler.GetRuleByName(fileName)
		if rule != nil && rule.manifest != nil {
			l.Debugf("find by name from scheduler...")
			return rule.manifest, nil
		}
	}

	return GetManifest(filepath.Join("./rules.d/", fileName))

}

func GetManifest(filename string) (*RuleManifest, error) {

	if !strings.HasSuffix(filename, ".manifest") {
		filename += ".manifest"
	}
	m := &RuleManifest{path: filename}
	if err := m.parse(); err != nil {
		return nil, err
	}
	return m, nil
}
func (c *Checker) start(ctx context.Context) {
	defer func() {
		if e := recover(); e != nil {
			buf := make([]byte, 2048)
			n := runtime.Stack(buf, false)
			l.Errorf("panic %s", e)
			l.Errorf("%s", string(buf[:n]))

		}
		output.Close()
		l.Info("checker exit")
	}()

	l.Infof("scheduler start")

	go c.taskScheduler.run()

	select {
	case <-ctx.Done():
		return
	default:
	}

	go cgroup.Run() //所有规则加载完成后 启动cgroup

	// 发送一次消息到output上
	firstTrigger()
	<-ctx.Done()
	c.taskScheduler.Stop()
}
