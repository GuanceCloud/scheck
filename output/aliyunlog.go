package output

import (
	"encoding/json"
	"time"

	sls "github.com/aliyun/aliyun-log-go-sdk"
	"github.com/gogo/protobuf/proto"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/config"
)

type AliYunLog struct {
	AliSls       *config.AliSls
	Client       sls.ClientInterface
	pending      []*sample
	maxPending   int
	lastSendTime int64
}

func (a *AliYunLog) conn(aliSls *config.AliSls) {
	a.AliSls = aliSls
	client := sls.CreateNormalInterface(aliSls.EndPoint, aliSls.AccessKeyID, aliSls.AccessKeySecret, aliSls.SecurityToken)
	a.Client = client
}

func (a *AliYunLog) CreateProject() {

	if a.AliSls.ProjectName == "" {
		a.AliSls.ProjectName = "zhuyun-scheck"
	}
	if a.AliSls.Description == "" {
		a.AliSls.Description = "This is the Zhuyun security component Scheck project"
	}

	_, err := a.Client.GetProject(a.AliSls.ProjectName)
	if err != nil {
		_, err := a.Client.CreateProject(a.AliSls.ProjectName, a.AliSls.Description)
		if err != nil {
			l.Errorf("Create project : %s failed %v\n, ", a.AliSls.Description, err)
		}
	}
	count := 0
	for {
		_, err := a.Client.GetProject(a.AliSls.ProjectName)
		if err == nil {
			l.Info("%s exist", a.AliSls.ProjectName)
			break
		} else if count > 10 {
			l.Errorf("Create project timeout ")
			break
		}
		time.Sleep(2 * time.Second)
		count++
	}

	if a.AliSls.LogStoreName == "" {
		a.AliSls.LogStoreName = "scheck"
	}

	if _, err := a.Client.GetLogStore(a.AliSls.ProjectName, a.AliSls.LogStoreName); err != nil {
		if err := a.Client.CreateLogStore(a.AliSls.ProjectName, a.AliSls.LogStoreName, 3, 2, true, 6); err != nil {
			l.Errorf("Create LogStore : %s failed %v\n", a.AliSls.LogStoreName, err)
		}
	}

	logStoreCount := 0
	for {
		_, err := a.Client.GetLogStore(a.AliSls.ProjectName, a.AliSls.LogStoreName)
		if err == nil {
			l.Info("%s exist", a.AliSls.LogStoreName)
			break
		} else if logStoreCount > 10 {
			l.Errorf("Create LogStore timeout ")
			break
		}
		time.Sleep(2 * time.Second)
		logStoreCount++
	}

}

func (a *AliYunLog) CreateIndex(fields map[string]interface{}) error {

	indexKey := map[string]sls.IndexKey{}
	for i := range fields {
		indexKey[i] = sls.IndexKey{
			Token:         []string{""},
			CaseSensitive: false,
			Type:          "text",
			DocValue:      true,
		}
	}

	index := sls.Index{
		Keys: indexKey,
		Line: &sls.IndexLine{
			Token:         []string{",", ":", " "},
			CaseSensitive: false,
			IncludeKeys:   []string{},
			ExcludeKeys:   []string{},
		},
	}

	err := a.Client.CreateIndex(a.AliSls.ProjectName, a.AliSls.LogStoreName, index)
	if err != nil {
		if err = a.Client.UpdateIndex(a.AliSls.ProjectName, a.AliSls.LogStoreName, index); err != nil {
			l.Errorf("Index error %v\n", err)
		}
		return err
	}
	time.Sleep(4 * time.Second)

	return nil
}

func (a *AliYunLog) PutLogs(fields map[string]interface{}) error {
	logs := []*sls.Log{}
	content := []*sls.LogContent{}
	for i := range fields {
		content = append(content, &sls.LogContent{
			Key:   proto.String(i),
			Value: proto.String(fields[i].(string)),
		})
	}
	slslog := &sls.Log{
		Time:     proto.Uint32(uint32(time.Now().Unix())),
		Contents: content,
	}
	logs = append(logs, slslog)
	loggroup := &sls.LogGroup{
		Logs: logs,
	}

	if err := a.Client.PutLogs(a.AliSls.ProjectName, a.AliSls.LogStoreName, loggroup); err == nil {
		l.Debug("PutLogs success")
		return nil
	} else {
		l.Errorf("PutLogs failed %v\n", err)
		return err
	}
}
func newSls(aliSls *config.AliSls, maxpending int) *AliYunLog {
	ali := &AliYunLog{
		AliSls:     aliSls,
		maxPending: maxpending,
	}
	ali.conn(aliSls)
	ali.CreateProject()
	return ali
}

func (a *AliYunLog) Stop() {

}

//
func (a *AliYunLog) ReadMsg(measurement string, tags map[string]string, fields map[string]interface{}, t ...time.Time) {
	var data []byte
	var err error

	slsBody := make(map[string]interface{})
	slsBody["ruleid"] = measurement
	for k, v := range tags {
		slsBody[k] = v
	}
	for k, v := range fields {
		slsBody[k] = v
	}
	slsBody["timestamp"] = time.Now().UTC()
	data, err = json.Marshal(slsBody)
	if err != nil {
		return
	}

	a.pending = append(a.pending, &sample{data: data})
	timenow := time.Now().Unix()
	if len(a.pending) >= a.maxPending || (timenow-a.lastSendTime) > 10 {
		a.ToUpstream(a.pending...)
		a.pending = make([]*sample, 0)
		a.lastSendTime = timenow
		return
	}
}

func (a *AliYunLog) ToUpstream(sams ...*sample) {
	for _, s := range sams {
		fields := make(map[string]interface{})
		if err := json.Unmarshal(s.data, &fields); err != nil {
			l.Fatalf("data 序列号失败 %s", err)

		}
		a.conn(a.AliSls)
		if err := a.CreateIndex(fields); err != nil {
			l.Errorf("CreateIndex err %v", err)
		}
		if err := a.PutLogs(fields); err != nil {
			l.Errorf("CreateIndex err %v", err)
		}

	}
}
