package output

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"math"
	"reflect"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"

	ifxcli "github.com/influxdata/influxdb1-client/v2"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/config"
)

var (
	l       = logger.DefaultSLogger("output")
	uploads map[string]outPuterInterface
)

type outPuterInterface interface {
	ReadMsg(measurement string, tags map[string]string, fields map[string]interface{}, t ...time.Time)
	ToUpstream(samples ...*sample)
	Stop()
}

type sample struct {
	data []byte
}

func newOutputer(scOutPut *config.ScOutput) {

	uploads = make(map[string]outPuterInterface)

	flag := false
	if scOutPut.Log != nil && scOutPut.Log.Enable {
		uploads["local"] = newLocalLog(scOutPut.Log.Output)
		flag = true
	}
	if scOutPut.Http != nil && scOutPut.Http.Enable {
		uploads["http"] = newDatakitWriter(scOutPut.Http.Output, 100)
		flag = true
	}
	if scOutPut.AliSls != nil && scOutPut.AliSls.Enable {
		if scOutPut.AliSls.AccessKeyID == "" || scOutPut.AliSls.AccessKeySecret == "" || scOutPut.AliSls.EndPoint == "" {
			l.Errorf("%s", "access_key_id or access_key_secret or endpoint cannot be empty ")
		} else {
			uploads["sls"] = newSls(scOutPut.AliSls, 100)
			flag = true
		}

	}
	if !flag {
		uploads["stdout"] = newLocalLog(scOutPut.Log.Output)
	}

}

func Start(scOutPut *config.ScOutput) {
	l = logger.SLogger("output")
	newOutputer(scOutPut)

}
func Close() {
	for _, upload := range uploads {
		upload.Stop()
	}
}

// SendMetric lua.trigger callback
func SendMetric(measurement string, tags map[string]string, fields map[string]interface{}, t ...time.Time) error {
	if uploads == nil {
		uploads = make(map[string]outPuterInterface)
	}
	if len(uploads) == 0 {
		// when:test/lua.CompileLua/  the uploads.len == 0 , use os.stdout
		uploads["stdout"] = newLocalLog("stdout")
	}
	for _, writer := range uploads {
		writer.ReadMsg(measurement, tags, fields, t...)
	}
	return nil
}

func buildBody(data []byte) (body []byte, gzon bool, err error) {
	if len(data) > 1024 { // should not gzip on file output
		if body, err = gzipCompress(data); err != nil {
			l.Errorf("%s", err.Error())
			return
		}
		gzon = true
	} else {
		body = data
		gzon = false
	}

	return
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
				l.Warnf("on input `%s', filed %s, get uint64 %d > MaxInt64(%d), dropped", name, k, v.(uint64), uint64(math.MaxInt64))
				delete(fields, k)
			} else { // convert uint64 -> int64
				fields[k] = int64(v.(uint64))
			}
		case int, uint32, uint16, uint8, int64, int32, int16, int8, bool, string, float32, float64:
		default:
			l.Warnf("invalid filed type `%s', from `%s', on filed `%s', got value `%+#v'", reflect.TypeOf(v).String(), name, k, fields[k])
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

	_ = zw.Flush()
	_ = zw.Close()
	return z.Bytes(), nil
}
