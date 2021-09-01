package securitychecker

import (
	"fmt"
	"github.com/shirou/gopsutil/process"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/git"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/internal/global"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
)

const (
	OSWindows = `windows`
	OSLinux   = `linux`
	OSDarwin  = `darwin`

	OSArchWinAmd64    = "windows/amd64"
	OSArchWin386      = "windows/386"
	OSArchLinuxArm    = "linux/arm"
	OSArchLinuxArm64  = "linux/arm64"
	OSArchLinux386    = "linux/386"
	OSArchLinuxAmd64  = "linux/amd64"
	OSArchDarwinAmd64 = "darwin/amd64"
)

var (
	Version            = git.Version
	InstallDir         = optionalInstallDir[runtime.GOOS+"/"+runtime.GOARCH]
	optionalInstallDir = map[string]string{
		OSArchWinAmd64: `C:\Program Files\scheck`,
		OSArchWin386:   `C:\Program Files (x86)\scheck`,

		OSArchLinuxArm:    `/usr/local/scheck`,
		OSArchLinuxArm64:  `/usr/local/scheck`,
		OSArchLinuxAmd64:  `/usr/local/scheck`,
		OSArchLinux386:    `/usr/local/scheck`,
		OSArchDarwinAmd64: `/usr/local/scheck`,
	}

	AppBin  = "scheck"
	pidFile = filepath.Join(InstallDir, ".pid")
)

func SavePid() error {
	if isRuning() {
		return fmt.Errorf("scheck still running, PID: %s", pidFile)
	}

	pid := os.Getpid()
	return ioutil.WriteFile(pidFile, []byte(fmt.Sprintf("%d", pid)), os.ModePerm)
}

func isRuning() bool {
	var oidPid int64
	var name string
	var p *process.Process

	cont, err := ioutil.ReadFile(pidFile)

	// pid文件不存在
	if err != nil {
		return false
	}

	oidPid, err = strconv.ParseInt(string(cont), global.ParseBase, 32)
	if err != nil {
		return false
	}

	p, _ = process.NewProcess(int32(oidPid))
	name, _ = p.Name()
	return name == getBinName()
}

func getBinName() string {
	bin := AppBin

	if runtime.GOOS == "windows" {
		bin += ".exe"
	}

	return bin
}
