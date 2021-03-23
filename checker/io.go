package checker

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"reflect"
	"strings"
	"time"

	ifxcli "github.com/influxdata/influxdb1-client/v2"
)

type (
	outputer struct {
		outputPath string

		httpCli    *http.Client
		outputFile *os.File
	}
)

func newOutputer(output string) *outputer {
	o := &outputer{
		outputPath: output,
	}

	if o.outputPath == "" {
		o.outputFile = os.Stdout
	} else if strings.HasPrefix(o.outputPath, "file://") {
		path := strings.TrimPrefix(o.outputPath, "file://")
		f, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
		if err != nil {
			log.Printf("[error] %s", err)
		} else {
			o.outputFile = f
		}
	} else if strings.HasPrefix(o.outputPath, "http://") || strings.HasPrefix(o.outputPath, "https://") {
		o.httpCli = &http.Client{
			Timeout: 30 * time.Second,
		}
	} else {
		log.Printf("[warn] invalid output: %s", output)
	}

	return o
}

func (o *outputer) close() {
	if o.outputFile != nil && o.outputFile != os.Stdout {
		if err := o.outputFile.Close(); err != nil {
			log.Printf("[error] %s", err)
		}
	}
}

func (c *outputer) sendMetric(ctx context.Context, measurement string, tags map[string]string, fields map[string]interface{}, t ...time.Time) error {

	data, err := makeMetric(measurement, tags, fields, t...)
	if err != nil {
		return err
	}

	return c.sendData(data)
}

func buildBody(data []byte) (body []byte, gzon bool, err error) {
	if len(data) > 1024 { // should not gzip on file output
		if body, err = gzipCompress(data); err != nil {
			log.Printf("[error] %s", err.Error())
			return
		}
		gzon = true
	}

	return
}

func (o *outputer) sendData(data []byte) error {

	if len(data) == 0 {
		return nil
	}

	body, gz, err := buildBody(data)
	if err != nil {
		return err
	}

	if o.outputFile != nil {
		if _, err := o.outputFile.Write(append(body, '\n')); err != nil {
			log.Printf("%s", err)
			return err
		}
		return nil
	}

	if o.httpCli == nil {
		return nil
	}

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
		log.Printf("[debug] post %d to %s ok(gz: %v), cost %v, response: %s",
			len(body), o.outputPath, gz, time.Since(postbeg), string(respbody))
		return nil

	case 4:
		log.Printf("[error] post %d to %s failed(HTTP: %d): %s, cost %v, data dropped", len(body), o.outputPath, resp.StatusCode, string(respbody), time.Since(postbeg))
		return fmt.Errorf("4xx error")

	case 5:
		log.Printf("[error] post %d to %s failed(HTTP: %s): %s, cost %v",
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
