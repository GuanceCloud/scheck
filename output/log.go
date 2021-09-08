package output

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"

	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/internal/global"
)

// localLog 本地存储
type localLog struct {
	filePath   string
	outputFile io.WriteCloser
}

func newLocalLog(filePath string) *localLog {
	filePath = strings.TrimPrefix(filePath, "file://")

	local := &localLog{filePath: filePath}
	if filePath == "stdout" {
		local.outputFile = os.Stdout
		l.Info("init stdout success")
		return local
	}
	_ = os.MkdirAll(filepath.Dir(filePath), global.FileModeMkdirAll)

	logf, err := rotatelogs.New(
		filePath+".%Y%m%d%H%M",            // 没有使用go语言规范的format格式
		rotatelogs.WithLinkName(filePath), // 快捷方式名称
		rotatelogs.WithMaxAge(global.LocalLogMaxAge),
		rotatelogs.WithRotationTime(global.LocalLogRotate),
	)
	if err != nil {
		l.Errorf("init rotatelogs err=%v", err)
		return nil
	}

	local.outputFile = logf
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
