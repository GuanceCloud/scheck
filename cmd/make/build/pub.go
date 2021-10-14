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

	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/internal/global"

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

func tarFiles(goos, goarch string) {
	bin := AppBin
	if goos == global.OSWindows {
		bin += global.WindowsExt
	}
	gz := filepath.Join(PubDir, Release, fmt.Sprintf("%s-%s-%s-%s.tar.gz",
		AppBin, goos, goarch, git.Version))
	args := []string{
		`czf`,
		gz,
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

func tarData() {
	gz := filepath.Join(PubDir, Release, dataTar)
	args := []string{
		`czf`,
		gz,
		"data/dict.txt",
	}
	cmd := exec.Command("tar", args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		l.Fatal(err)
	}
}

func getCurrentVersionInfo(urlSrt string) *versionDesc {
	l.Infof("get current online version: %s", urlSrt)

	// #nosec
	resp, err := http.Get(urlSrt)
	if err != nil {
		l.Fatal(err)
	}

	if resp.StatusCode != http.StatusOK {
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
	oc := InitOC()
	var archs []string
	// 请求线上版本信息
	urlSrt := fmt.Sprintf("http://%s.%s/%s/%s", oc.BucketName, oc.Host, OSSPath, "version")
	curVd := getCurrentVersionInfo(urlSrt)
	_ = curVd

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
	ossfiles := installAllArchs(archs)
	uploadToOSS(ossfiles, oc)

	l.Infof("Done!(elapsed: %v)", time.Since(start))
}

// installAllArchs : tar files and collect OSS upload/backup info
func installAllArchs(archs []string) map[string]string {
	ossfiles := map[string]string{
		path.Join(OSSPath, "version"):                                     path.Join(PubDir, Release, "version"),
		path.Join(OSSPath, dataTar):                                       path.Join(PubDir, Release, dataTar), // 20211008 slq 添加data目录打包
		path.Join(OSSPath, "install.sh"):                                  "install.sh",
		path.Join(OSSPath, "install.ps1"):                                 "install.ps1",
		path.Join(OSSPath, fmt.Sprintf("install-%s.sh", ReleaseVersion)):  "install.sh",
		path.Join(OSSPath, fmt.Sprintf("install-%s.ps1", ReleaseVersion)): "install.ps1",
	}

	for _, arch := range archs {
		if arch == "darwin/amd64" && runtime.GOOS != "darwin" {
			l.Warn("Not a darwin system, skip the upload of related files.")
			continue
		}
		parts := strings.Split(arch, "/")
		if len(parts) != ArchLen {
			l.Fatalf("invalid arch %q", parts)
		}
		goos, goarch := parts[0], parts[1]
		md5File := fmt.Sprintf("%s-%s-%s-%s.md5", AppBin, goos, goarch, git.Version)
		ossfiles[path.Join(OSSPath, md5File)] = path.Join(PubDir, Release, md5File)

		tarFiles(parts[0], parts[1])

		gzName := fmt.Sprintf("%s-%s-%s.tar.gz", AppBin, goos+"-"+goarch, git.Version)

		ossfiles[path.Join(OSSPath, gzName)] = path.Join(PubDir, Release, gzName)

		if goos == global.OSWindows {
			installerExe = fmt.Sprintf("installer-%s-%s-%s.exe", goos, goarch, git.Version)
			noVerInstallerExe = fmt.Sprintf("installer-%s-%s.exe", goos, goarch)
		} else {
			installerExe = fmt.Sprintf("installer-%s-%s-%s", goos, goarch, git.Version)
			noVerInstallerExe = fmt.Sprintf("installer-%s-%s", goos, goarch)
		}

		ossfiles[path.Join(OSSPath, installerExe)] = path.Join(BuildInstallDir, installerExe)
		ossfiles[path.Join(OSSPath, noVerInstallerExe)] = path.Join(BuildInstallDir, noVerInstallerExe)
	}
	return ossfiles
}

func InitOC() *cliutils.OssCli {
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
	var split = 2 // at most 2 substrings
	ossSlice := strings.SplitN(DownloadAddr, "/", split)
	if len(ossSlice) != split {
		l.Fatalf("downloadAddr:%s err", DownloadAddr)
	}
	OSSPath = ossSlice[1]

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
	return oc
}

func uploadToOSS(ossFiles map[string]string, oc *cliutils.OssCli) {
	// test if all file ok before uploading
	for _, v := range ossFiles {
		if _, err := os.Stat(v); err != nil {
			l.Fatal(err)
		}
	}

	for k, v := range ossFiles {
		fi, _ := os.Stat(v)
		l.Debugf("upload %s(%s)...", v, humanize.Bytes(uint64(fi.Size())))
		l.Debugf("key: %s, value: %s", k, v)
		if err := oc.Upload(v, k); err != nil {
			l.Fatal(err)
		}
	}
}
