package secChecker

import (
	"path/filepath"
	"runtime"
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
	optionalInstallDir = map[string]string{
		OSArchWinAmd64: filepath.Join(`C:\Program Files\dataflux\` + ServiceName),
		OSArchWin386:   filepath.Join(`C:\Program Files (x86)\dataflux\` + ServiceName),

		OSArchLinuxArm:    filepath.Join(`/usr/local/cloudcare/dataflux/`, ServiceName),
		OSArchLinuxArm64:  filepath.Join(`/usr/local/cloudcare/dataflux/`, ServiceName),
		OSArchLinuxAmd64:  filepath.Join(`/usr/local/cloudcare/dataflux/`, ServiceName),
		OSArchLinux386:    filepath.Join(`/usr/local/cloudcare/dataflux/`, ServiceName),
		OSArchDarwinAmd64: filepath.Join(`/usr/local/cloudcare/dataflux/`, ServiceName),
	}

	InstallDir = optionalInstallDir[runtime.GOOS+"/"+runtime.GOARCH]

	MainConfPath = filepath.Join(InstallDir, "cfg")

	RulesDir = filepath.Join(InstallDir, "rules")
)
