package output

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/config"

	log "github.com/sirupsen/logrus"

	ifxcli "github.com/influxdata/influxdb1-client/v2"
)

type (
	sample struct {
		data []byte
	}

	DataOutputer struct {
		outputPath string
		scOutPut   *config.ScOutput
		httpCli    *http.Client
		outputFile *os.File
		Sls        *AliYunLog
		queue      chan *sample
	}
)

var (
	Outputer *DataOutputer
)

func NewOutputer(scOutPut *config.ScOutput) *DataOutputer {
	o := &DataOutputer{
		//outputPath: output,
		scOutPut: scOutPut,
		queue:    make(chan *sample, 1000),
	}
	if scOutPut.AliSls == nil && scOutPut.Log == nil && scOutPut.Http == nil {
		o.outputFile = os.Stdout
	} else if o.scOutPut.Http.Enable {
		if strings.HasPrefix(o.scOutPut.Http.Output, "http://") || strings.HasPrefix(o.scOutPut.Http.Output, "https://") {
			o.httpCli = &http.Client{
				Timeout: 30 * time.Second,
			}
		}
	} else if o.scOutPut.Log.Enable {
		if strings.HasPrefix(o.scOutPut.Log.Output, "file://") {
			path := strings.TrimPrefix(o.scOutPut.Log.Output, "file://")
			f, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
			if err != nil {
				os.MkdirAll(filepath.Dir(path), 0775)
				f, err = os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
				if err != nil {
					log.Errorf("%s", err)
				}
			} else {
				o.outputFile = f
			}
		}
	} else if o.scOutPut.AliSls.Enable {
		// Todo 阿里云sls 对接
		if o.scOutPut.AliSls.AccessKeyID == "" || o.scOutPut.AliSls.AccessKeySecret == "" || o.scOutPut.AliSls.EndPoint == "" {
			log.Errorf("%s", "access_key_id or access_key_secret or endpoint cannot be empty ")
		}

		var sls AliYunLog
		o.Sls = &sls
		o.Sls.conn(o.scOutPut.AliSls)
		//创建工程
		o.Sls.CreateProject()

	} else {
		o.outputFile = os.Stdout
	}

	Outputer = o

	return o
}

func Start(ctx context.Context, scOutPut *config.ScOutput) {

	Outputer = NewOutputer(scOutPut)

	go func() {
		defer Outputer.Close()
		ticker := time.NewTicker(time.Second * 10)
		defer ticker.Stop()

		var pending []*sample
		const maxPending int = 100
		for {
			select {
			case <-ctx.Done():
				Outputer.sendSamples(pending)
				return
			case sample, ok := <-Outputer.queue:
				if !ok {
					return
				}
				pending = append(pending, sample)
				if len(pending) >= maxPending {
					Outputer.sendSamples(pending)
					pending = pending[:0]
				}
			case <-ticker.C:
				Outputer.sendSamples(pending)
				pending = pending[:0]
			}
		}
	}()
}

func (o *DataOutputer) Close() {
	if o.outputFile != nil && o.outputFile != os.Stdout {
		if err := o.outputFile.Close(); err != nil {
			log.Errorf("%s", err)
		}
		o.outputFile = nil
	}
}

func SendMetric(measurement string, tags map[string]string, fields map[string]interface{}, t ...time.Time) error {
	var data []byte
	var err error
	// 阿里云日志处理
	if Outputer.scOutPut.AliSls.Enable {
		sls := make(map[string]interface{})
		sls["ruleid"] = measurement
		for k, v := range tags {
			sls[k] = v
		}
		for k, v := range fields {
			sls[k] = v
		}
		sls["timestamp"] = time.Now().UTC()
		data, err = json.Marshal(sls)
		if err != nil {
			return err
		}
	} else {
		data, err = makeMetric(measurement, tags, fields, t...)
		if err != nil {
			return err
		}
	}

	select {
	case Outputer.queue <- &sample{data}:
	case <-time.After(time.Second * 10):
		err = fmt.Errorf("timeout")
	}

	return err
}

func buildBody(data []byte) (body []byte, gzon bool, err error) {
	if len(data) > 1024 { // should not gzip on file output
		if body, err = gzipCompress(data); err != nil {
			log.Errorf("%s", err.Error())
			return
		}
		gzon = true
	} else {
		body = data
		gzon = false
	}

	return
}

func (o *DataOutputer) sendSamples(samples []*sample) error {
	if len(samples) == 0 {
		return nil
	}
	var datas [][]byte
	for _, s := range samples {
		datas = append(datas, s.data)
	}
	if o.scOutPut.AliSls.Enable {
		return o.sendSls(datas)
	} else {
		return o.sendData(bytes.Join(datas, []byte{'\n'}))
	}
}

// sls 发送方法
func (o *DataOutputer) sendSls(datas [][]byte) error {
	for _, i := range datas {
		fields := make(map[string]interface{})
		if err := json.Unmarshal(i, &fields); err != nil {
			log.Fatalf("data 序列号失败 %s", err)
			return err
		}
		var err error
		o.Sls.conn(o.scOutPut.AliSls)
		err = o.Sls.CreateIndex(fields)
		err = o.Sls.PutLogs(fields)
		return err
	}
	return nil
}

func (o *DataOutputer) sendData(data []byte) error {
	if len(data) == 0 {
		return nil
	}

	if o.outputFile != nil {
		if _, err := o.outputFile.Write(append(data, '\n')); err != nil {
			log.Errorf("%s", err)
			return err
		}
		return nil
	}

	if o.httpCli == nil {
		return nil
	}

	body, gz, err := buildBody(data)
	if err != nil {
		return err
	}

	log.Debugf("body:  %s", string(body))

	req, err := http.NewRequest("POST", o.outputPath, bytes.NewBuffer(body))
	if err != nil {
		log.Printf("%s", err)
		return err
	}

	if gz {
		req.Header.Set("Content-Encoding", "gzip")
	}

	postbeg := time.Now()

	resp, err := o.httpCli.Do(req)
	if err != nil {
		log.Printf("%s", err)
		return err
	}

	defer resp.Body.Close()
	respbody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("%s", err)
		return err
	}

	switch resp.StatusCode / 100 {
	case 2:
		log.Debugf("post %d to %s ok(gz: %v), cost %v, response: %s",
			len(body), o.outputPath, gz, time.Since(postbeg), string(respbody))
		return nil

	case 4:
		log.Errorf("post %d to %s failed(HTTP: %d): %s, cost %v, data dropped", len(body), o.outputPath, resp.StatusCode, string(respbody), time.Since(postbeg))
		return fmt.Errorf("4xx error")

	case 5:
		log.Errorf("post %d to %s failed(HTTP: %s): %s, cost %v",
			len(body), o.outputPath, resp.Status, string(respbody), time.Since(postbeg))
		return fmt.Errorf("5xx error")
	}

	return nil
}

func makeMetric(name string, tags map[string]string, fields map[string]interface{}, t ...time.Time) ([]byte, error) {
	var tm time.Time
	if len(t) > 0 {
		tm = t[0]
	}

	if tm.IsZero() {
		tm = time.Now().UTC()
	}

	for k, v := range tags { // remove any suffix `\` in all tag values
		tags[k] = trimSuffixAll(v, `\`)
	}

	for k, v := range fields { // convert uint to int
		switch v.(type) {
		case uint64:
			if v.(uint64) > uint64(math.MaxInt64) {
				log.Printf("on input `%s', filed %s, get uint64 %d > MaxInt64(%d), dropped", name, k, v.(uint64), uint64(math.MaxInt64))
				delete(fields, k)
			} else { // convert uint64 -> int64
				fields[k] = int64(v.(uint64))
			}
		case int, uint32, uint16, uint8, int64, int32, int16, int8, bool, string, float32, float64:
		default:
			log.Printf("invalid filed type `%s', from `%s', on filed `%s', got value `%+#v'", reflect.TypeOf(v).String(), name, k, fields[k])
			return nil, fmt.Errorf("invalid field type")
		}
	}

	pt, err := ifxcli.NewPoint(name, tags, fields, tm)
	if err != nil {
		return nil, err
	}
	return []byte(pt.String()), nil
}

func trimSuffixAll(s, sfx string) string {
	var x string
	for {
		x = strings.TrimSuffix(s, sfx)
		if x == s {
			break
		}
		s = x
	}

	return x
}

func gzipCompress(data []byte) ([]byte, error) {
	var z bytes.Buffer
	zw := gzip.NewWriter(&z)
	if _, err := zw.Write(data); err != nil {
		return nil, err
	}

	zw.Flush()
	zw.Close()
	return z.Bytes(), nil
}
