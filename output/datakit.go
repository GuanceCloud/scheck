package output

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/git"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/internal/global"
)

// DatakitWriter :send Msg to datakit.
type DatakitWriter struct {
	httpURL      string
	pending      []*sample
	maxPending   int
	lastSendTime int64
	samSig       chan *sample
}

func newDatakitWriter(httpURL string, maxPending int) *DatakitWriter {
	if !strings.HasPrefix(httpURL, "http://") && !strings.HasPrefix(httpURL, "https://") {
		httpURL = "http://" + httpURL
	}
	chanL := 10
	version := spiltVersion(git.Version)
	if version != "" {
		version = "?version=" + version // 请求参数添加版本信息
	}
	dk := &DatakitWriter{
		httpURL:      httpURL + version,
		maxPending:   maxPending,
		lastSendTime: time.Now().Unix(),
		samSig:       make(chan *sample, chanL),
	}
	go dk.start()
	l.Infof("init output for datakit ok,path=%s", httpURL)
	return dk
}

func (dk *DatakitWriter) Stop() {
}

func (dk *DatakitWriter) ReadMsg(measurement string,
	tags map[string]string, fields map[string]interface{}, t ...time.Time) {
	var data []byte
	var err error
	data, err = makeMetric(measurement, tags, fields, t...)
	if err != nil {
		return
	}
	dk.samSig <- &sample{data: data}
}

func (dk *DatakitWriter) start() {
	timeSleep := 10
	for {
		select {
		case sam := <-dk.samSig:
			dk.pending = append(dk.pending, sam)
			if len(dk.pending) > dk.maxPending {
				dk.ToUpstream(dk.pending...)
				dk.pending = make([]*sample, 0)
			}
		case <-time.After(time.Duration(timeSleep) * time.Second):
			if len(dk.pending) != 0 {
				dk.ToUpstream(dk.pending...)
				dk.pending = make([]*sample, 0)
			}
		}
	}
}

// ToUpstream :send msg to datakit. The buffer will be emptied whether the message is sent successfully or not.
func (dk *DatakitWriter) ToUpstream(sams ...*sample) {
	var datas [][]byte
	for _, s := range sams {
		datas = append(datas, s.data)
	}
	data := bytes.Join(datas, []byte{'\n'})

	body, gz, err := buildBody(data)
	if err != nil {
		l.Errorf("build data err %v", err)
		return
	}

	req, err := http.NewRequest("POST", dk.httpURL, bytes.NewBuffer(body))
	if err != nil {
		l.Errorf("%s", err)
		return
	}

	if gz {
		req.Header.Set("Content-Encoding", "gzip")
	}
	http.DefaultClient.Timeout = global.OutputDefTimeout
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		l.Errorf("%s", err)
		return
	}

	defer func() {
		_ = resp.Body.Close()
	}()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		l.Errorf("%s", err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		l.Errorf("post %d to %s failed(HTTP: %d): %s, cost %v, data dropped",
			len(body), dk.httpURL, resp.StatusCode, string(respBody), time.Since(time.Now()))
	}
}

func spiltVersion(version string) string {
	if version == "" {
		return ""
	}
	if !strings.Contains(version, "-") {
		return version
	}
	return strings.Split(version, "-")[0]
}
