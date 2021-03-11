package secChecker

import (
	"os"
	"path/filepath"
	"runtime"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
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

	CommonChanCap = 32
)

var (
	l = logger.DefaultSLogger("sec-checker")

	OptionalInstallDir = map[string]string{
		OSArchWinAmd64: filepath.Join(`C:\Program Files\dataflux\` + ServiceDirName),
		OSArchWin386:   filepath.Join(`C:\Program Files (x86)\dataflux\` + ServiceDirName),

		OSArchLinuxArm:    filepath.Join(`/usr/local/cloudcare/dataflux/`, ServiceDirName),
		OSArchLinuxArm64:  filepath.Join(`/usr/local/cloudcare/dataflux/`, ServiceDirName),
		OSArchLinuxAmd64:  filepath.Join(`/usr/local/cloudcare/dataflux/`, ServiceDirName),
		OSArchLinux386:    filepath.Join(`/usr/local/cloudcare/dataflux/`, ServiceDirName),
		OSArchDarwinAmd64: filepath.Join(`/usr/local/cloudcare/dataflux/`, ServiceDirName),
	}

	InstallDir = OptionalInstallDir[runtime.GOOS+"/"+runtime.GOARCH]

	MainConfPath = filepath.Join(InstallDir, "cfg")

	RulesDir = filepath.Join(InstallDir, "rules")
)

func InitDirs() {
	for _, dir := range []string{RulesDir} {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			l.Fatalf("create %s failed: %s", dir, err)
		}
	}
}
