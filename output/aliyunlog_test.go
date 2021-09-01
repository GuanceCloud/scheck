package output

import (
	"os"
	"testing"

	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/config"
)

func TestAliYunLog_CreateProject(t *testing.T) {
	ret := AliYunLog{}
	ali := &config.AliSls{
		Enable:          true,
		EndPoint:        "https://cn-hangzhou.log.aliyuncs.com",
		AccessKeyID:     os.Getenv("LOG_AccessKeyID"),
		AccessKeySecret: os.Getenv("LOG_AccessKeySecret"),
		ProjectName:     "zhuyun-scheck2",
		LogStoreName:    "scheck",
	}
	ret.conn(ali)
	ret.CreateProject()
}
