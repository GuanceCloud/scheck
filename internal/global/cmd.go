package global

import (
	"crypto/md5" // #nosec
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"

	"github.com/shirou/gopsutil/process"
)

func SavePid() error {
	pidFile := filepath.Join(InstallDir, PidExt)
	if isRuning(pidFile) {
		return fmt.Errorf("Scheck still running, PID: " + pidFile)
	}

	pid := os.Getpid()
	return ioutil.WriteFile(pidFile, []byte(fmt.Sprintf("%d", pid)), FileModeRW)
}

func isRuning(pidFile string) bool {
	var oidPid int64
	var name string
	var p *process.Process

	cont, err := ioutil.ReadFile(pidFile)

	// pid文件不存在
	if err != nil {
		return false
	}

	oidPid, err = strconv.ParseInt(string(cont), ParseBase, ParseBitSize)
	if err != nil {
		return false
	}

	p, _ = process.NewProcess(int32(oidPid))
	name, _ = p.Name()

	return name == getBinName()
}

func getBinName() string {
	bin := AppBin
	if runtime.GOOS == OSWindows {
		bin += WindowsExt
	}
	return bin
}

func CheckMd5(releaseType string) {
	urls := map[string]string{"release": ReleaseURL, "test": TestURL}

	httpURL := fmt.Sprintf("https://%s/scheck-%s-%s-%s.md5", urls[releaseType], runtime.GOOS, runtime.GOARCH, Version)
	resp, err := http.Get(httpURL) // #nosec
	if err != nil {
		fmt.Printf("http get err=%v \n", err)
		return
	}
	defer resp.Body.Close()
	remoteVal, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("read response body err=%v \n", err)
		return
	}
	checkLocalPath := filepath.Join(InstallDir, AppBin)
	data, err := ioutil.ReadFile(checkLocalPath)
	if err != nil {
		fmt.Printf("readFile path=%s, err=%v \n", checkLocalPath, err)
		return
	}
	newMd5 := md5.New() // #nosec
	newMd5.Write(data)
	localVal := hex.EncodeToString(newMd5.Sum(nil))

	if localVal != "" && localVal == string(remoteVal) {
		fmt.Println("MD5 verify ok")
		return
	}
	fmt.Printf("[Error] MD5 checksum not match !!!")
}

func CreateSymlinks() {
	var x [][2]string
	//nolint:gofmt
	if runtime.GOOS == OSWindows {
		x = [][2]string{
			[2]string{
				filepath.Join(InstallDir, AppBin+WindowsExt),
				`C:\WINDOWS\system32\scheck.exe`,
			},
		}
	} else {
		x = [][2]string{
			[2]string{
				filepath.Join(InstallDir, AppBin),
				"/usr/local/bin/scheck",
			},

			[2]string{
				filepath.Join(InstallDir, AppBin),
				"/usr/local/sbin/scheck",
			},

			[2]string{
				filepath.Join(InstallDir, AppBin),
				"/sbin/scheck",
			},

			[2]string{
				filepath.Join(InstallDir, AppBin),
				"/usr/sbin/scheck",
			},

			[2]string{
				filepath.Join(InstallDir, AppBin),
				"/usr/bin/scheck",
			},
		}
	}

	for _, item := range x {
		if err := symlink(item[0], item[1]); err != nil {
			fmt.Printf("create scheck symlink: %s -> %s: %s, ignored", item[1], item[0], err.Error())
		}
	}
}

func symlink(src, dst string) error {
	fmt.Printf("remove link %s... \n", dst)
	if err := os.Remove(dst); err != nil {
		fmt.Printf("%s, ignored \n", err)
	}
	return os.Symlink(src, dst)
}
