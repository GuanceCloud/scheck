package luaext

import (
	"bytes"
	"log"
	"testing"
	"text/template"
)

func TestDemo(t *testing.T) {

	// bashTimestampRx, _ := regexp.Compile("^#([0-9]+)$")

	// res := bashTimestampRx.FindAllString(`#1018`, 1)
	// log.Printf("%v", res)

	//getLasts()

	ss := `dfd fdf {{.name}} dfd {{.job}}`

	tmpl, err := template.New("test").Parse(ss)
	if err != nil {
		log.Fatal(err)
	}
	_ = tmpl

	buf1 := bytes.NewBufferString("")
	if err = tmpl.Execute(buf1, map[string]string{
		"name": "wqc",
	}); err != nil {
		log.Fatalf("%s", err)
	}
	log.Printf("%s", buf1.String())

	buf2 := bytes.NewBufferString("")
	tmpl.Execute(buf2, map[string]string{
		"name": "wyc",
		"job":  "dev",
	})
	log.Printf("%s", buf2.String())

	//ulimitInfo(nil)
}
