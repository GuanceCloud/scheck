package global

import (
	"path/filepath"
	"runtime"
	"time"

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

	LuaRet1                    = 1
	LuaRet2                    = 2
	LuaArgIdx1                 = 1
	LuaArgIdx2                 = 2
	LuaArgIdx3                 = 3
	LuaConfiguration           = "__this_configuration"
	LuaConfigurationKey        = "ruleFile"
	LuaStatusFile              = ".status.json"
	LuaStatusOutFileMD         = "./%s.lua_status.md"
	LuaStatusOutFileHTML       = "./%s.lua_status.html"
	LuaStatusWriteFileInterval = time.Minute * 5
	LuaCronDisable             = "disable"
	LuaScriptTimeout           = time.Second * 10

	FileModeRW       = 0644
	FileModeMkdir    = 0666
	FileModeMkdirAll = 0755

	KB = 1024
	MB = KB * 1024
	GB = MB * 1024

	ParseBase    = 10
	ParseBase16  = 16
	ParseBitSize = 64
)

var (
	Version            = git.Version
	LocalGOOS          = runtime.GOOS
	LocalGOARCH        = runtime.GOARCH
	InstallDir         = optionalInstallDir[LocalGOOS+"/"+LocalGOARCH]
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
	DefLogPath      = "/var/log/scheck"
	DefRulesDir     = "rules.d"
	LuaLocalLibPath = "lib"
	PublicLuaLib    = "?.lua"
	DumpFolder      = filepath.Join(InstallDir, "dump")
	CgroupPeriod    = 1000000

	DefPprofPort     = ":6060"
	DefOutputPending = 100
	DefLuaPoolCap    = 15
	DefLuaPoolMaxCap = 20

	LocalLogMaxAge = time.Hour * 24 * 7
	LocalLogRotate = time.Hour * 24
)
