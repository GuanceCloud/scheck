package build

import (
	"crypto/md5" // #nosec
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/git"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/internal/global"
)

var (
	/* Use:
		go tool dist list
	to get current os/arch list */

	OSArches = []string{ // supported os/arch list
		`linux/386`,
		`linux/amd64`,
		`linux/arm`,
		`linux/arm64`,
		`windows/386`,
		`windows/amd64`,
	}
	ArchLen        = 2
	ReleaseVersion = git.Version
	CommandName    = "go"
)

var (
	l = logger.DefaultSLogger("build")

	BuildDir        = "build"
	BuildInstallDir = "build/install"
	PubDir          = "pub"
	AppBin          = "scheck"
	OSSPath         = "scheck"
	AllArch         = "all"
	LocalArch       = "local"
	Archs           string
	Release         string
	MainEntry       string
	DownloadAddr    string
	ReleaseType     string
)

func prepare() *versionDesc {
	_ = os.RemoveAll(BuildDir)
	if err := os.MkdirAll(BuildDir, os.ModePerm); err != nil {
		l.Fatalf("MkdirAll %s error, err: %s", BuildDir, err)
	}
	l.Info("PubDir: %s", filepath.Join(PubDir, Release))
	if err := os.MkdirAll(filepath.Join(PubDir, Release), os.ModePerm); err != nil {
		l.Fatalf("MkdirAll %s error, err: %s", PubDir, err)
	}

	// create version info
	vd := &versionDesc{
		Version:  strings.TrimSpace(git.Version),
		Date:     git.BuildAt,
		Uploader: git.Uploader,
		Branch:   git.Branch,
		Commit:   git.Commit,
		Go:       git.Golang,
	}

	versionInfo, err := json.MarshalIndent(vd, "", "    ")
	if err != nil {
		l.Fatal(err)
	}

	if err := ioutil.WriteFile(filepath.Join(PubDir, Release, "version"), versionInfo, os.ModeAppend|os.ModePerm); err != nil {
		l.Fatal(err)
	}

	return vd
}

func Compile() {
	start := time.Now()

	vd := prepare()

	var archs []string

	switch Archs {
	case AllArch:
		archs = OSArches

		// read cmd-line env
		if x := os.Getenv("ALL_ARCHS"); x != "" {
			archs = strings.Split(x, "|")
		}
	case LocalArch:
		archs = []string{runtime.GOOS + "/" + runtime.GOARCH}
		if x := os.Getenv("LOCAL"); x != "" {
			archs = strings.Split(x, "|")
		}
	default:
		archs = strings.Split(Archs, "|")
	}

	for idx := range archs {
		parts := strings.Split(archs[idx], "/")
		if len(parts) != ArchLen {
			l.Fatalf("invalid arch %q", parts)
		}

		goos, goarch := parts[0], parts[1]
		if goos == global.OSDarwin && runtime.GOOS != global.OSDarwin {
			l.Warnf("skip build datakit under %s", archs[idx])
			continue
		}
		dir := fmt.Sprintf("%s/%s-%s-%s", BuildDir, AppBin, goos, goarch)

		err := os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			l.Fatalf("failed to mkdir: %v", err)
		}

		dir, err = filepath.Abs(dir)
		if err != nil {
			l.Fatal(err)
		}

		compileArch(AppBin, goos, goarch, dir, vd.Version)

		if goos == global.OSWindows {
			installerExe = fmt.Sprintf("installer-%s-%s-%s.exe", goos, goarch, ReleaseVersion)
			noVerInstallerExe = fmt.Sprintf("installer-%s-%s.exe", goos, goarch)
		} else {
			installerExe = fmt.Sprintf("installer-%s-%s-%s", goos, goarch, ReleaseVersion)
			noVerInstallerExe = fmt.Sprintf("installer-%s-%s", goos, goarch)
		}

		// build installer 将install/main.go 编译成exe文件
		buidAllInstaller(BuildInstallDir, goos, goarch)
	}
	l.Infof("Done!(elapsed %v)", time.Since(start))
}

func calcMD5(path string) string {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return ""
	}
	// #nosec
	return fmt.Sprintf("%x", md5.Sum(data))
}

func compileArch(bin, goos, goarch, dir, version string) {
	output := filepath.Join(dir, bin)

	if goos == global.OSWindows {
		output += global.WindowsExt
	}

	md5File := fmt.Sprintf("%s-%s-%s-%s.md5", bin, goos, goarch, version)

	cgoEnabled := "0"
	if goos == global.OSDarwin {
		cgoEnabled = "1"
	}

	args := []string{
		"build",
		"-o", output,
		"-ldflags",
		fmt.Sprintf("-w -s -X main.ReleaseType=%s -X main.Version=%s -X DownloadURL=%s", ReleaseType, version, DownloadAddr),
		MainEntry,
	}

	env := []string{
		"GOOS=" + goos,
		"GOARCH=" + goarch,
		`GO111MODULE=off`,
		"CGO_ENABLED=" + cgoEnabled,
	}

	l.Debugf("building %s", fmt.Sprintf("%s-%s/%s", goos, goarch, bin))
	msg, err := runEnv(args, env)
	if err != nil {
		l.Fatalf("failed to run %v, envs: %v: %v, msg: %s", args, env, err, string(msg))
	}

	fileMd5 := calcMD5(output)
	if fileMd5 == "" {
		l.Fatalf("fail to compute md5: %s", output)
	}
	if err := ioutil.WriteFile(filepath.Join(PubDir, Release, md5File), []byte(fileMd5), os.ModeAppend|os.ModePerm); err != nil {
		l.Fatalf("fail to write md5 checksum, %s", err)
	}
}

func buidAllInstaller(outdir, goos, goarch string) {
	buildInstaller(outdir, goos, goarch, installerExe)
	buildInstaller(outdir, goos, goarch, noVerInstallerExe)
}

func buildInstaller(outdir, goos, goarch, installerName string) {
	args := []string{
		"build",
		"-o", filepath.Join(outdir, installerName),
		"-ldflags",
		fmt.Sprintf("-w -s -X main.DataKitBaseURL=%s -X main.DataKitVersion=%s", DownloadAddr, ReleaseVersion),
		"cmd/installer/main.go",
	}

	env := []string{
		"GOOS=" + goos,
		"GOARCH=" + goarch,
	}

	msg, err := runEnv(args, env)
	if err != nil {
		l.Fatalf("failed to run %v, envs: %v: %v, msg: %s", args, env, err, string(msg))
	}
	l.Infof("build installer to %s", filepath.Join(outdir, installerName))
}

func runEnv(args, env []string) ([]byte, error) {
	cmd := exec.Command(CommandName, args...)
	if env != nil {
		cmd.Env = append(os.Environ(), env...)
	}
	return cmd.CombinedOutput()
}
