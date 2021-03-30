package luaext

import (
	"fmt"
	"time"

	lua "github.com/yuin/gopher-lua"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/output"
)

func sendMetric(l *lua.LState) int {

	var err error

	var measurement string
	var fields map[string]interface{}
	var tags map[string]string
	var tm time.Time

	lv := l.Get(1)
	if v, ok := lv.(lua.LString); !ok {
		err = fmt.Errorf("bad argument 1 (%v expected, got %v)", lua.LTString.String(), lv.Type().String())
	} else {
		measurement = string(v)
		if measurement == "" {
			err = fmt.Errorf("measurement cannot be empty")
		}
	}

	if err != nil {
		goto End
	}

	lv = l.Get(2)
	if v, ok := lv.(*lua.LTable); !ok {
		err = fmt.Errorf("bad argument 2 (%v expected, got %v)", lua.LTTable.String(), lv.Type().String())
	} else {
		fields, err = table2fields(v)
		if err == nil {
			if len(fields) == 0 {
				err = fmt.Errorf("fields requires at least one key-value")
			}
		}
	}

	if err != nil {
		goto End
	}

	lv = l.Get(3)
	switch lv.Type() {
	case lua.LTNil:
	case lua.LTTable:
		tags, err = table2tags(lv.(*lua.LTable))
	case lua.LTNumber:
		num := lv.(lua.LNumber)
		if int64(num) > 0 {
			tm = time.Unix(int64(num), 0)
		}
	default:
		err = fmt.Errorf("bad argument 3 ( %v(tags) or %v(timestamp) expected, got %v )", lua.LTTable.String(), lua.LTNumber.String(), lv.Type().String())
	}

	if err != nil {
		goto End
	}

	if l.GetTop() >= 4 {
		lv = l.Get(4)
		if v, ok := lv.(lua.LNumber); !ok {
			err = fmt.Errorf("bad argument 4 (%v expected, got %v)", lua.LTNumber.String(), lv.Type().String())
		} else {
			if int64(v) > 0 {
				tm = time.Unix(int64(v), 0)
			}
		}
	}

	if err != nil {
		goto End
	}

	err = output.Outputer.SendMetric(measurement, tags, fields, tm)

End:
	if err != nil {
		l.Push(lua.LString(err.Error()))
	} else {
		l.Push(lua.LString(""))
	}
	return 1
}

func table2tags(tbl *lua.LTable) (map[string]string, error) {
	tags := map[string]string{}
	var err error

	cb := func(k lua.LValue, v lua.LValue) {
		if err != nil {
			return
		}

		switch v.Type() {
		case lua.LTBool:
		case lua.LTNumber:
		case lua.LTString:
		default:
			err = fmt.Errorf("invalid value of tags key: '%s'. tags value only support 'boolean', 'string' and 'number'", k.String())
		}
		if err == nil {
			tags[k.String()] = v.String()
		}
	}

	tbl.ForEach(cb)

	if err != nil {
		return nil, err
	}
	return tags, nil
}

func table2fields(tbl *lua.LTable) (map[string]interface{}, error) {
	fields := map[string]interface{}{}

	var err error

	cb := func(k lua.LValue, v lua.LValue) {
		if err != nil {
			return
		}

		switch v.Type() {
		case lua.LTBool:
		case lua.LTNumber:
		case lua.LTString:
		default:
			err = fmt.Errorf("invalid value of fields key: '%s'. fields value only support 'boolean', 'string' and 'number'", k.String())
		}
		if err == nil {
			var val interface{}
			if v.Type() == lua.LTNumber {
				num := v.(lua.LNumber)
				if float64(num) == float64(int64(num)) {
					val = int64(num)
				} else {
					val = float64(num)
				}
			} else {
				val = v.String()
			}
			fields[k.String()] = val
		}
	}

	tbl.ForEach(cb)

	if err != nil {
		return nil, err
	}
	return fields, nil
}

type cache struct {
	store map[string]string
}

func (c *cache) set(k string, v string) {

}
