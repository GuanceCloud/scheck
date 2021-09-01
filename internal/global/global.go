package global

import (
	"runtime"

	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/git"
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

	AppBin = "scheck"

	WindowsExt     = ".exe"
	LuaManifestExt = ".manifest"
	LuaExt         = ".lua"
	PidExt         = ".pid"

	ReleaseURL = "zhuyun-static-files-production.oss-cn-hangzhou.aliyuncs.com/security-checker"
	TestURL    = "zhuyun-static-files-testing.oss-cn-hangzhou.aliyuncs.com/security-checker"
)

// 公用参数：go和lua的相互调用函数使用的常量、文件操作使用的mode参数、等
const (
	LuaRet1    = 1
	LuaRet2    = 2
	LuaArgIdx1 = 1
	LuaArgIdx2 = 2
	LuaArgIdx3 = 3

	FileModeRW       = 0644
	FileModeMkdir    = 0666
	FileModeMkdirAll = 0755

	KB = 1024
	MB = KB * 1024
	GB = MB * 1024

	ParseBase    = 10
	ParseBitSize = 64
)

var (
	Version            = git.Version
	InstallDir         = optionalInstallDir[runtime.GOOS+"/"+runtime.GOARCH]
	optionalInstallDir = map[string]string{
		OSArchWinAmd64: `C:\\Program Files\\scheck`,
		OSArchWin386:   `C:\\Program Files (x86)\\scheck`,

		OSArchLinuxArm:    `/usr/local/scheck`,
		OSArchLinuxArm64:  `/usr/local/scheck`,
		OSArchLinuxAmd64:  `/usr/local/scheck`,
		OSArchLinux386:    `/usr/local/scheck`,
		OSArchDarwinAmd64: `/usr/local/scheck`,
	}

	// DefLogPath is default config
	DefLogPath  = "/var/log/scheck"
	DefRulesDir = "rules.d"

	DefPprofPort     = ":6060"
	DefOutputPending = 100
)
