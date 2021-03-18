package io

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"reflect"
	"strings"
	"time"

	ifxcli "github.com/influxdata/influxdb1-client/v2"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	secChecker "gitlab.jiagouyun.com/cloudcare-tools/sec-checker"
)

var (
	l          = logger.DefaultSLogger("io")
	httpCli    *http.Client
	outputURL  string
	outputFile *os.File

	inputCh = make(chan *iodata, 32)

	cache        = map[string][][]byte{}
	curCacheCnt  = 0
	curCacheSize = 0
)

const (
	maxCacheCnt      = 1000
	MaxPostFailCache = 1024

	minGZSize = 1024
)

type iodata struct {
	name string
	data []byte // line-protocol or json or others
}

func Start(ctx context.Context) {

	l = logger.SLogger("io")

	defer func() {
		if e := recover(); e != nil {
			l.Errorf("panic, %v", e)
		}

		if outputFile != nil {
			if err := outputFile.Close(); err != nil {
				l.Error(err)
			}
		}
	}()

	outpath := secChecker.Cfg.Output

	if outpath == "stdout" {
		outputFile = os.Stdout
	} else if strings.HasPrefix(outpath, "file://") {
		path := strings.TrimPrefix(outpath, "file://")
		f, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
		if err != nil {
			l.Error(err)
		} else {
			outputFile = f
		}
	} else if strings.HasPrefix(outpath, "http://") || strings.HasPrefix(outpath, "https://") {
		outputURL = outpath
	}

	httpCli = &http.Client{
		Timeout: 30 * time.Second,
	}

	tick := time.NewTicker(time.Second * 10)
	defer tick.Stop()

	for {
		select {
		case d := <-inputCh:
			cacheData(d)

		case <-tick.C:
			flushAll(ctx)

		case <-ctx.Done():
			l.Info("exit")
			return
		}
	}
}

func cacheData(d *iodata) {
	if d == nil {
		l.Warn("get empty data, ignored")
		return
	}

	cache[d.name] = append(cache[d.name], d.data)
	curCacheCnt++
}

func flushAll(ctx context.Context) {
	flush(ctx)

	if curCacheCnt > 0 {
		l.Warnf("post failed cache count: %d", curCacheCnt)
	}

	if curCacheCnt > MaxPostFailCache {
		l.Warnf("failed cache count reach max limit(%d), cleanning cache...", MaxPostFailCache)
		for k := range cache {
			cache[k] = nil
		}
		curCacheCnt = 0
	}
}

func flush(ctx context.Context) {

	if httpCli != nil {
		defer httpCli.CloseIdleConnections()
	}

	for k, v := range cache {

		select {
		case <-ctx.Done():
			return
		default:
		}

		if err := doFlush(v, k); err != nil {
			l.Errorf("post %d to %s failed", len(v), k)
		} else {
			curCacheCnt -= len(v)
			l.Debugf("clean %d/%d cache on %s", len(v), curCacheCnt, k)
			cache[k] = nil
		}
	}
}

func buildBody(bodies [][]byte) (body []byte, gzon bool, err error) {
	body = bytes.Join(bodies, []byte("\n"))
	if len(body) > minGZSize { // should not gzip on file output
		if body, err = GZip(body); err != nil {
			l.Errorf("gz: %s", err.Error())
			return
		}
		gzon = true
	}

	return
}

func doFlush(bodies [][]byte, name string) error {

	if bodies == nil {
		return nil
	}

	body, gz, err := buildBody(bodies)
	if err != nil {
		return err
	}

	if outputFile != nil {
		if _, err := outputFile.Write(append(body, '\n')); err != nil {
			l.Error(err)
			return err
		}
		return nil
	}

	if outputURL == "" {
		l.Errorf("empty url")
		return nil
	}

	req, err := http.NewRequest("POST", outputURL, bytes.NewBuffer(body))
	if err != nil {
		l.Error(err)
		return err
	}

	if gz {
		req.Header.Set("Content-Encoding", "gzip")
	}

	postbeg := time.Now()

	resp, err := httpCli.Do(req)
	if err != nil {
		l.Error(err)
		return err
	}

	defer resp.Body.Close()
	respbody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		l.Error(err)
		return err
	}

	switch resp.StatusCode / 100 {
	case 2:
		l.Debugf("post %d to %s ok(gz: %v), cost %v, response: %s",
			len(body), outputURL, gz, time.Since(postbeg), string(respbody))
		return nil

	case 4:
		l.Debugf("post %d to %s failed(HTTP: %s): %s, cost %v, data dropped", len(body), outputURL, resp.StatusCode, string(respbody), time.Since(postbeg))
		return fmt.Errorf("4xx error")

	case 5:
		l.Errorf("post %d to %s failed(HTTP: %s): %s, cost %v",
			len(body), outputURL, resp.Status, string(respbody), time.Since(postbeg))
		return fmt.Errorf("5xx error")
	}

	return nil
}

func NamedFeedEx(ctx context.Context, name, metric string,
	tags map[string]string,
	fields map[string]interface{},
	t ...time.Time) error {
	return doFeedEx(ctx, name, metric, tags, fields, t...)
}

func doFeedEx(ctx context.Context, name, metric string,
	tags map[string]string,
	fields map[string]interface{},
	t ...time.Time) error {

	data, err := MakeMetric(metric, tags, fields, t...)
	if err != nil {
		return err
	}
	return doFeed(ctx, data, name)
}

func doFeed(ctx context.Context, data []byte, name string) error {

	select {
	case inputCh <- &iodata{
		data: data,
		name: name,
	}: // XXX: blocking

	case <-ctx.Done():
		l.Warnf("%s feed skipped on global exit", name)
	}

	return nil
}

func MakeMetric(name string, tags map[string]string, fields map[string]interface{}, t ...time.Time) ([]byte, error) {
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
			l.Errorf("invalid filed type `%s', from `%s', on filed `%s', got value `%+#v'", reflect.TypeOf(v).String(), name, k, fields[k])
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

func GZip(data []byte) ([]byte, error) {
	var z bytes.Buffer
	zw := gzip.NewWriter(&z)
	if _, err := zw.Write(data); err != nil {
		return nil, err
	}

	zw.Flush()
	zw.Close()
	return z.Bytes(), nil
}
