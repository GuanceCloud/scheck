package build

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/git"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/logger"
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
	ReleaseVersion = git.Version
)

func runEnv(args, env []string) ([]byte, error) {
	cmd := exec.Command(args[0], args[1:]...)
	if env != nil {
		cmd.Env = append(os.Environ(), env...)
	}

	return cmd.CombinedOutput()
}

var (
	l = logger.DefaultSLogger("build")

	BuildDir        = "build"
	BuildInstallDir = "build/install"
	PubDir          = "pub"
	//AppName      = "security-checker"
	AppBin       = "scheck"
	OSSPath      = "security-checker"
	Archs        string
	Release      string
	MainEntry    string
	DownloadAddr string
	ReleaseType  string
)

func prepare() *versionDesc {

	os.RemoveAll(BuildDir)
	_ = os.MkdirAll(BuildDir, os.ModePerm)
	_ = os.MkdirAll(filepath.Join(PubDir, Release), os.ModePerm)

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

	if err := ioutil.WriteFile(filepath.Join(PubDir, Release, "version"), versionInfo, 0666); err != nil {
		l.Fatal(err)
	}

	return vd
}

func Compile() {
	start := time.Now()

	vd := prepare()

	var archs []string

	switch Archs {
	case "all":
		archs = OSArches

		// read cmd-line env
		if x := os.Getenv("ALL_ARCHS"); x != "" {
			archs = strings.Split(x, "|")
		}
	case "local":
		archs = []string{runtime.GOOS + "/" + runtime.GOARCH}
		if x := os.Getenv("LOCAL"); x != "" {
			archs = strings.Split(x, "|")
		}
	default:
		archs = strings.Split(Archs, "|")
	}

	for idx := range archs {

		parts := strings.Split(archs[idx], "/")
		if len(parts) != 2 {
			l.Fatalf("invalid arch %q", parts)
		}

		goos, goarch := parts[0], parts[1]
		if goos == "darwin" && runtime.GOOS != "darwin" {
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

		if goos == "windows" {
			installerExe = fmt.Sprintf("installer-%s-%s-%s.exe", goos, goarch, ReleaseVersion)
			noVerInstallerExe = fmt.Sprintf("installer-%s-%s.exe", goos, goarch)
		} else {
			installerExe = fmt.Sprintf("installer-%s-%s-%s", goos, goarch, ReleaseVersion)
			noVerInstallerExe = fmt.Sprintf("installer-%s-%s", goos, goarch)
		}

		// build installer 将install/main.go 编译成exe文件 (slq:outdir随意填的 后面改 20210805P)
		buidAllInstaller(BuildInstallDir, goos, goarch)
	}
	l.Infof("Done!(elapsed %v)", time.Since(start))
}

func calcMD5(path string) string {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return ""
	}
	m := md5.New()
	m.Write(data)
	return hex.EncodeToString(m.Sum(nil))
}

func compileArch(bin, goos, goarch, dir, version string) {

	output := filepath.Join(dir, bin)

	if goos == "windows" {
		output += ".exe"
	}

	md5File := fmt.Sprintf("%s-%s-%s-%s.md5", bin, goos, goarch, version)

	cgo_enabled := "0"
	if goos == "darwin" {
		cgo_enabled = "1"
	}

	args := []string{
		"go", "build",
		"-o", output,
		"-ldflags",
		fmt.Sprintf("-w -s -X main.ReleaseType=%s -X main.Version=%s", ReleaseType, version),
		MainEntry,
	}

	env := []string{
		"GOOS=" + goos,
		"GOARCH=" + goarch,
		`GO111MODULE=off`,
		"CGO_ENABLED=" + cgo_enabled,
	}

	l.Debugf("building %s", fmt.Sprintf("%s-%s/%s", goos, goarch, bin))
	msg, err := runEnv(args, env)
	if err != nil {
		l.Fatalf("failed to run %v, envs: %v: %v, msg: %s", args, env, err, string(msg))
	}

	md5 := calcMD5(output)
	if md5 == "" {
		l.Fatalf("fail to compute md5: %s", output)
	}
	if err := ioutil.WriteFile(filepath.Join(PubDir, Release, md5File), []byte(md5), 0664); err != nil {
		l.Fatalf("fail to write md5 checksum, %s", err)
	}
}

type installInfo struct {
	Name         string
	DownloadAddr string
	Version      string
}

func buidAllInstaller(outdir, goos, goarch string) {
	buildInstaller(outdir, goos, goarch, installerExe)
	buildInstaller(outdir, goos, goarch, noVerInstallerExe)
}
func buildInstaller(outdir, goos, goarch, installerName string) {

	// ------------ 生成系统的install文件------------
	args := []string{
		"go", "build",
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
