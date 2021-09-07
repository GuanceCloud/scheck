package dumperror

import (
	"os"
	"syscall"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/internal/global"
	// "gmshield/common/logs"
)

type dumperror struct {
	filename string
	needdump bool
}

func CreateDumpError(flag bool) *dumperror {
	tNow := time.Now()
	timeNow := tNow.Format("20060102-150405")
	logFolder := global.DumpFolder
	_, err := os.Stat(logFolder)
	if err != nil {
		if !os.IsNotExist(err) {
			l.Errorf("err=%v", err)
			return nil
		}
		err = os.Mkdir(logFolder, os.ModePerm)
		if err != nil {
			l.Errorf("cannot mkdir err=%v ", err)
			return nil
		}
	}

	name := string(logFolder + "/" + timeNow + "_dump.log")
	dump := &dumperror{
		filename: name,
		needdump: flag}
	dump.DumpError()
	return dump
}

func (Dump *dumperror) DumpError() {
	if Dump.needdump {
		logFile, err := os.OpenFile(Dump.filename, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0660)
		if err != nil {
			l.Errorf("open dump file err:%v", err)
			return
		}
		defer logFile.Close()
		syscall.Dup2(int(logFile.Fd()), int(os.Stderr.Fd()))
	}
}

func Start() {
	CreateDumpError(true)
}
