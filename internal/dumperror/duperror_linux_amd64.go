package dumperror

import (
	"os"
	"path/filepath"
	"syscall"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/internal/global"
)

type dumperror struct {
	filename string
	needdump bool
}

func createDumpError(flag bool) *dumperror {
	tNow := time.Now()
	timeNow := tNow.Format("20060102-150405")
	logFolder := global.DumpFolder
	_, err := os.Stat(logFolder)
	if err != nil {
		if !os.IsNotExist(err) {
			l.Errorf("err=%v", err)
			return nil
		}
		err = os.Mkdir(logFolder, global.FileModeMkdir)
		if err != nil {
			l.Errorf("cannot mkdir err=%v ", err)
			return nil
		}
	}

	name := filepath.Join(logFolder, timeNow+"_dump.log")
	dump := &dumperror{
		filename: name,
		needdump: flag,
	}
	dump.DumpError()
	return dump
}

func (dump *dumperror) DumpError() {
	if dump.needdump {
		logFile, err := os.OpenFile(dump.filename, os.O_CREATE|os.O_APPEND|os.O_RDWR, global.FileModeRW)
		if err != nil {
			l.Errorf("open dump file err:%v", err)
			return
		}
		defer func() {
			_ = logFile.Close()
		}()
		_ = syscall.Dup2(int(logFile.Fd()), int(os.Stderr.Fd()))
	}
}

func Start() {
	createDumpError(true)
}
