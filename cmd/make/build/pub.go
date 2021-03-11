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

type versionDesc struct {
	Version  string `json:"version"`
	Date     string `json:"date_utc"`
	Uploader string `json:"uploader"`
	Branch   string `json:"branch"`
	Commit   string `json:"commit"`
	Go       string `json:"go"`
}

func Publish() {

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

	// upload all build archs
	archs := []string{}

	switch Archs {
	case "all":
		archs = OSArches
	case "local":
		archs = []string{runtime.GOOS + "/" + runtime.GOARCH}
	default:
		archs = strings.Split(Archs, "|")
	}

	ossfiles := map[string]string{
		path.Join(PubDir, Release, "version"): path.Join(OSSPath, "version"),
	}

	renameOssFiles := map[string]string{}

	// tar files and collect OSS upload/backup info
	for _, arch := range archs {
		parts := strings.Split(arch, "/")
		if len(parts) != 2 {
			l.Fatalf("invalid arch %q", parts)
		}
		goos, goarch := parts[0], parts[1]

		tarFiles(parts[0], parts[1])

		gzName := fmt.Sprintf("%s-%s-%s.tar.gz", BinName, goos+"-"+goarch, git.Version)

		ossfiles[path.Join(PubDir, Release, gzName)] = path.Join(OSSPath, gzName)

		installerExeName := ""
		if goos == "windows" {
			installerExeName = fmt.Sprintf("installer-%s-%s.exe", goos, goarch)

			if curVd != nil && curVd.Version != git.Version {
				renameOssFiles[path.Join(OSSPath, installerExeName)] =
					path.Join(OSSPath, fmt.Sprintf("installer-%s-%s-%s.exe", goos, goarch, curVd.Version))
			}

		} else {
			installerExeName = fmt.Sprintf("installer-%s-%s", goos, goarch)

			if curVd != nil && curVd.Version != git.Version {
				renameOssFiles[path.Join(OSSPath, installerExeName)] =
					path.Join(OSSPath, fmt.Sprintf("installer-%s-%s-%s", goos, goarch, curVd.Version))
			}
		}

		ossfiles[path.Join(PubDir, Release, installerExeName)] = path.Join(OSSPath, installerExeName)
	}

	// backup old installer script online, make it possible to install old version if required
	for k, v := range renameOssFiles {
		if err := oc.Move(k, v); err != nil {
			l.Debugf("backup %s -> %s failed: %s, ignored", k, v, err.Error())
			continue
		}

		l.Debugf("backup %s -> %s ok", k, v)
	}

	// test if all file ok before uploading
	for k := range ossfiles {
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

func tarFiles(goos, goarch string) {

	gz := filepath.Join(PubDir, Release, fmt.Sprintf("%s-%s-%s-%s.tar.gz",
		BinName, goos, goarch, git.Version))
	args := []string{
		`czf`,
		gz,
		`-C`,
		// the whole buildDir/datakit-<goos>-<goarch> dir
		filepath.Join(BuildDir, fmt.Sprintf("%s-%s-%s", BinName, goos, goarch)), `.`,
	}

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
