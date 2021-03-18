package secChecker

import (
	"fmt"
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
)

func GetServiceDefaltInstallDir(serviceDirName string) string {
	d := OptionalInstallDir[runtime.GOOS+"/"+runtime.GOARCH]
	if d != "" {
		return filepath.Join(d, serviceDirName)
	}

	return ""
}

func MainConfPath(installDir string) string {
	return filepath.Join(installDir, "cfg")
}

func RulesDir(installDir string) string {
	return filepath.Join(installDir, "rules")
}

func createDefaultDirs(installDir string) error {
	for _, dir := range []string{RulesDir(installDir)} {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return fmt.Errorf("create %s failed: %s", dir, err)
		}
	}
	return nil
}

func PreInstall(installDir string) error {
	err := createDefaultConfigFile(installDir)
	if err != nil {
		return err
	}
	return createDefaultDirs(installDir)
}

func PostInstall(installDir string) error {
	return nil
}
