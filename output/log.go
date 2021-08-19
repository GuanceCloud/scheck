package output

import (
	"os"
	"path/filepath"
	"strings"
	"time"
)

// localLog 本地存储
type localLog struct {
	filePath   string
	outputFile *os.File
}

func newLocalLog(filePath string) *localLog {
	if strings.HasPrefix(filePath, "file://") {
		filePath = strings.TrimPrefix(filePath, "file://")
	}
	local := &localLog{filePath: filePath}

	if filePath == "stdout" {
		local.outputFile = os.Stdout
		l.Info("init stdout success")
		return local
	}

	f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		os.MkdirAll(filepath.Dir(filePath), 0775)
		f, err = os.OpenFile(filePath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
		if err != nil {
			l.Errorf("%s", err)
		}
	}

	local.outputFile = f
	l.Infof("init log ok! path=%s", filePath)
	return local
}
func (log *localLog) Stop() {
	_ = log.outputFile.Close()
}

func (log *localLog) ReadMsg(measurement string, tags map[string]string, fields map[string]interface{}, t ...time.Time) {
	var data []byte
	var err error
	data, err = makeMetric(measurement, tags, fields, t...)
	if err != nil {
		l.Errorf("err %v", err)
		return
	}
	sa := &sample{data: data}
	log.ToUpstream(sa)
}

func (log *localLog) ToUpstream(sam ...*sample) {

	if _, err := log.outputFile.Write(append(sam[0].data, byte('\n'))); err != nil {
		l.Errorf("%s", err)
	}

}
