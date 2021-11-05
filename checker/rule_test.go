package checker

import (
	"io/ioutil"
	"testing"

	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"
)

func TestGetManifest(t *testing.T) {
	/*
		4117-k8s-edct-conf-priv
		4118-k8s-etcd-ownership
		4121-k8s-edct-dir-priv
		4122-k8s-etcd-dir-ownership
	*/
	bts, err := ioutil.ReadFile("../man/libs/4122-k8s-etcd-dir-ownership.manifest")
	if err != nil {
		t.Log(err)
		return
	}
	// t.Log(bts)
	// contents := bytes.TrimPrefix(bts, []byte("\xef\xbb\xbf"))
	tbl, err := toml.Parse(bts)
	if err != nil {
		t.Logf("toml.Parse err=%v", err)
		return
	}
	for k, v := range tbl.Fields {
		if kv, ok := v.(*ast.KeyValue); ok {
			switch kt := kv.Value.(type) {
			case *ast.String:
				t.Logf("k=%s  v=%v", k, kt.Value)
			case *ast.Integer:
				i64, err := kt.Int()
				if err != nil {
					t.Error(err)
					continue
				}
				t.Logf("k=%s  v=%d", k, i64)
			case *ast.Array:
				t.Logf("k=%s  v=%s", k, kt.Source())
			}
		}
	}
}

func Test_checkRunTime(t *testing.T) {
	type args struct {
		cronStr string
		want    int64
	}
	tests := []args{
		{cronStr: "* */1 * * *", want: 60 * 1000},
		{cronStr: "* */5 * * *", want: 300 * 1000},
		{cronStr: "1 * * * *", want: 60 * 1000},
		{cronStr: "* * */1 * *", want: 3600 * 1000},
	}
	for _, arg := range tests {
		// args 重新赋值 会消除go lint错误
		// https://stackoverflow.com/questions/68559574/using-the-variable-on-range-scope-x-in-function-literal-scopelint
		arg := arg
		t.Run(arg.cronStr, func(t *testing.T) {
			if got := checkRunTime(arg.cronStr); got != arg.want {
				t.Errorf("checkRunTime() = %v, want %v", got, arg.want)
			}
		})
	}
}
