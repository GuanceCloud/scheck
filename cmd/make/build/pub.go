package build

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/git"

	"github.com/dustin/go-humanize"
)

var (
	installerExe      string
	noVerInstallerExe string
)

type versionDesc struct {
	Version  string `json:"version"`
	Date     string `json:"date_utc"`
	Uploader string `json:"uploader"`
	Branch   string `json:"branch"`
	Commit   string `json:"commit"`
	Go       string `json:"go"`
}

func (vd *versionDesc) withoutGitCommit() string {
	parts := strings.Split(vd.Version, "-")
	if len(parts) != 3 {
		l.Fatalf("version info not in v<x.x>-<n>-g<commit-id> format: %s", vd.Version)
	}

	return strings.Join(parts[:2], "-")
}

func tarFiles(goos, goarch string) {
	bin := AppBin
	if goos == "windows" {
		bin += ".exe"
	}
	gz := filepath.Join(PubDir, Release, fmt.Sprintf("%s-%s-%s-%s.tar.gz",
		AppBin, goos, goarch, git.Version))
	args := []string{
		`czf`,
		gz,
		`autostart`,
		`-C`,
		filepath.Join(BuildDir, fmt.Sprintf("%s-%s-%s", AppBin, goos, goarch)),
		bin,
	}
	l.Debug(args)
	cmd := exec.Command("tar", args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		l.Fatal(err)
	}
}

func getCurrentVersionInfo(url string) *versionDesc {

	l.Infof("get current online version: %s", url)

	resp, err := http.Get(url)
	if err != nil {
		l.Fatal(err)
	}

	if resp.StatusCode != 200 {
		l.Warn("get current online version failed, ignored")
		return nil
	}

	defer resp.Body.Close()
	info, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		l.Fatal(err)
	}

	l.Infof("current online version: %s", string(info))
	var vd versionDesc
	if err := json.Unmarshal(info, &vd); err != nil {
		l.Fatal(err)
	}
	return &vd
}

func PubDatakit() {
	start := time.Now()
	var ak, sk, bucket, ossHost string

	// 在你本地设置好这些 oss-key 环境变量
	switch Release {
	case `test`, `local`, `release`, `preprod`:
		tag := strings.ToUpper(Release)
		ak = os.Getenv(tag + "_OSS_ACCESS_KEY")
		sk = os.Getenv(tag + "_OSS_SECRET_KEY")
		bucket = os.Getenv(tag + "_OSS_BUCKET")
		ossHost = os.Getenv(tag + "_OSS_HOST")
	default:
		l.Fatalf("unknown release type: %s", Release)
	}

	if ak == "" || sk == "" {
		l.Fatalf("oss access key or secret key missing, tag=%s", strings.ToUpper(Release))
	}

	oc := &cliutils.OssCli{
		Host:       ossHost,
		PartSize:   512 * 1024 * 1024,
		AccessKey:  ak,
		SecretKey:  sk,
		BucketName: bucket,
		WorkDir:    OSSPath,
	}

	if err := oc.Init(); err != nil {
		l.Fatal(err)
	}

	// 请求线上版本信息
	url := fmt.Sprintf("http://%s.%s/%s/%s", bucket, ossHost, OSSPath, "version")
	curVd := getCurrentVersionInfo(url)
	_ = curVd

	// upload all build archs
	archs := []string{}

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

	ossfiles := map[string]string{
		path.Join(PubDir, Release, "version"): path.Join(OSSPath, "version"),
	}

	var ver versionDesc
	verdata, err := ioutil.ReadFile(filepath.Join(PubDir, Release, "version"))
	if err != nil {
		l.Fatalf("%s", err)
	}
	if err = json.Unmarshal(verdata, &ver); err != nil {
		l.Fatalf("%s", err)
	}
	for _, v := range archs {
		fields := strings.Split(v, "/")
		f := fmt.Sprintf("%s-%s-%s-%s.md5", AppBin, fields[0], fields[1], ver.Version)
		ossfiles[path.Join(PubDir, Release, f)] = path.Join(OSSPath, f)
	}

	if Archs == "darwin/amd64" {
		delete(ossfiles, path.Join(PubDir, Release, "version"))
	}

	renameOssFiles := map[string]string{}

	// tar files and collect OSS upload/backup info
	for _, arch := range archs {
		if arch == "darwin/amd64" && runtime.GOOS != "darwin" {
			l.Warn("Not a darwin system, skip the upload of related files.")
			continue
		}
		parts := strings.Split(arch, "/")
		if len(parts) != 2 {
			l.Fatalf("invalid arch %q", parts)
		}
		goos, goarch := parts[0], parts[1]

		tarFiles(parts[0], parts[1])

		gzName := fmt.Sprintf("%s-%s-%s.tar.gz", AppBin, goos+"-"+goarch, git.Version)

		gzNameNoVer := fmt.Sprintf("%s-%s.tar.gz", AppBin, goos+"-"+goarch)

		ossfiles[path.Join(PubDir, Release, gzName)] = path.Join(OSSPath, gzName)

		renameOssFiles[path.Join(OSSPath, gzName)] = path.Join(OSSPath, gzNameNoVer)

		if goos == "windows" {
			installerExe = fmt.Sprintf("installer-%s-%s-%s.exe", goos, goarch, git.Version)
			noVerInstallerExe = fmt.Sprintf("installer-%s-%s.exe", goos, goarch)

		} else {
			installerExe = fmt.Sprintf("installer-%s-%s-%s", goos, goarch, git.Version)
			noVerInstallerExe = fmt.Sprintf("installer-%s-%s", goos, goarch)
		}

		ossfiles[path.Join(BuildInstallDir, installerExe)] = path.Join(OSSPath, installerExe)
		ossfiles[path.Join(BuildInstallDir, noVerInstallerExe)] = path.Join(OSSPath, noVerInstallerExe)
	}

	// test if all file ok before uploading
	for k, _ := range ossfiles {
		if _, err := os.Stat(k); err != nil {
			l.Fatal(err)
		}
	}

	for k, v := range ossfiles {

		fi, _ := os.Stat(k)
		l.Debugf("upload %s(%s)...", k, humanize.Bytes(uint64(fi.Size())))

		if err := oc.Upload(k, v); err != nil {
			l.Fatal(err)
		}
	}

	l.Infof("Done!(elapsed: %v)", time.Since(start))
}
