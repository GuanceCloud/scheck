package checker

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"text/template"

	log "github.com/sirupsen/logrus"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/config"
)

const (
	manifestTempStr = `
id='{{.ID}}'
category='test'
level='info'
title='test'
desc='{{.Content}}'
cron='{{.Cron}}'

disabled={{.Disabled}}
`

	luaTempStr = `
local files={
	'/etc/passwd',
	'/etc/group'
}

local function check(file)
	local cache_key=file
	local hashval = file_hash(file)

	local old = get_cache(cache_key)
	if not old then
		set_cache(cache_key, hashval)
		return
	end

	if old ~= hashval then
		trigger({File=file})
		set_cache(cache_key, hashval)
	end
end

for i,v in ipairs(files) do
	check(v)
end
`
)

var (
	manifestTemp *template.Template
)

type testRule struct {
	name     string
	manifest string
	script   string
}

func newTestRule(idx int, cronSrt, lua string) *testRule {
	if manifestTemp == nil {
		manifestTemp, _ = template.New("manifest").Parse(manifestTempStr)
	}
	t := &testRule{
		name: fmt.Sprintf("test-%d", idx),
	}
	m := map[string]string{
		"ID":       fmt.Sprintf("test-%d", idx),
		"Cron":     cronSrt,
		"Disabled": "false",
	}
	var buf bytes.Buffer
	if err := manifestTemp.Execute(&buf, m); err != nil {
		log.Fatal(err)
		return nil
	}
	t.manifest = buf.String()
	t.script = lua
	if t.script == "" {
		t.script = fmt.Sprintf("print('test-%d')", idx)
	}
	return t
}

func (r *testRule) updateManifest(ruledir string) error {
	file := filepath.Join(ruledir, r.name)
	err := ioutil.WriteFile(file+".manifest", []byte(r.manifest), 0664)
	return err
}

func (r *testRule) updateScript(ruledir string) error {
	file := filepath.Join(ruledir, r.name)
	err := ioutil.WriteFile(file+".lua", []byte(r.script), 0664)
	return err
}

func mockRules() (string, []*testRule) {
	var testRules []*testRule
	for i := 1; i <= 200; i++ {
		testRules = append(testRules, newTestRule(i, `0 */1 * * *`, luaTempStr))
	}

	var ruleDir string
	var err error
	if ruleDir, err = ioutil.TempDir("", "scheck-"); err != nil {
		log.Fatal(err)
		return "", nil
	}
	for _, r := range testRules {
		if err = r.updateManifest(ruleDir); err != nil {
			log.Fatal(err)
			return "", nil
		}

		if err = r.updateScript(ruleDir); err != nil {
			log.Fatal(err)
			return "", nil
		}
	}
	return ruleDir, testRules
}

func TestChecker(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	ruleDir, rules := mockRules()
	if ruleDir == "" || len(rules) == 0 {
		return
	}
	defer os.RemoveAll(ruleDir)

	ctx, cancel := context.WithCancel(context.Background())
	_ = cancel
	Start(ctx, nil, &config.ScOutput{})
}

func TestParse(t *testing.T) {
	cronStr := `* */1 * * *`

	it := checkRunTime(cronStr)
	log.Println(it)
}
