package utils

import (
	"encoding/json"
	"errors"
	"strings"

	lua "github.com/yuin/gopher-lua"
)

func JSONEncode(l *lua.LState) int {
	lv := l.Get(1)
	data, err := Encode(lv)
	if err != nil {
		l.RaiseError("%s", err)
		return lua.MultRet
	}

	l.Push(lua.LString(string(data)))
	return 1
}

func JSONDecode(l *lua.LState) int {
	lv := l.Get(1)
	v, ok := lv.(lua.LString)
	if !ok {
		l.TypeError(1, lua.LTString)
		return lua.MultRet
	}
	delv, err := Decode(l, []byte(v))
	if err != nil {
		l.RaiseError("%s", err)
		return lua.MultRet
	}
	l.Push(delv)

	return 1
}

var (
	errNested      = errors.New("cannot encode recursively nested tables to JSON")
	errSparseArray = errors.New("cannot encode sparse array")
	errInvalidKeys = errors.New("cannot encode mixed or invalid key types")
)

type invalidTypeError lua.LValueType

func (i invalidTypeError) Error() string {
	return `cannot encode ` + lua.LValueType(i).String() + ` to JSON`
}

// Encode returns the JSON encoding of value.
func Encode(value lua.LValue) ([]byte, error) {
	return json.Marshal(jsonValue{
		LValue:  value,
		visited: make(map[*lua.LTable]bool),
	})
}

type jsonValue struct {
	lua.LValue
	visited map[*lua.LTable]bool
}

func (j jsonValue) MarshalJSON() (data []byte, err error) {
	switch converted := j.LValue.(type) {
	case lua.LBool:
		data, err = json.Marshal(bool(converted))
	case lua.LNumber:
		data, err = json.Marshal(float64(converted))
	case *lua.LNilType:
		data = []byte(`null`)
	case lua.LString:
		data, err = json.Marshal(string(converted))
	case *lua.LTable:
		if j.visited[converted] {
			return nil, errNested
		}
		j.visited[converted] = true

		key, value := converted.Next(lua.LNil)

		switch key.Type() { // nolint
		case lua.LTNil: // empty table
			data = []byte(`[]`)
		case lua.LTNumber:
			arr := make([]jsonValue, 0, converted.Len())
			expectedKey := lua.LNumber(1)
			for key != lua.LNil {
				if key.Type() != lua.LTNumber {
					err = errInvalidKeys
					return
				}
				if expectedKey != key {
					err = errSparseArray
					return
				}
				arr = append(arr, jsonValue{value, j.visited})
				expectedKey++
				key, value = converted.Next(key)
			}
			data, err = json.Marshal(arr)
		case lua.LTString:
			obj := make(map[string]jsonValue)
			for key != lua.LNil {
				if key.Type() != lua.LTString {
					err = errInvalidKeys
					return
				}
				obj[key.String()] = jsonValue{value, j.visited}
				key, value = converted.Next(key)
			}
			data, err = json.Marshal(obj)
		default:
			err = errInvalidKeys
		}
	default:
		err = invalidTypeError(j.LValue.Type())
	}
	return data, err
}

// Decode converts the JSON encoded data to Lua values.
func Decode(l *lua.LState, data []byte) (lua.LValue, error) {
	var value interface{}
	err := json.Unmarshal(data, &value)
	if err != nil {
		return nil, err
	}
	return DecodeValue(l, value), nil
}

// DecodeValue converts the value to a Lua value.
//
// This function only converts values that the encoding/json package decodes to.
// All other values will return lua.LNil.
func DecodeValue(l *lua.LState, value interface{}) lua.LValue {
	switch converted := value.(type) {
	case bool:
		return lua.LBool(converted)
	case float64:
		return lua.LNumber(converted)
	case string:
		return lua.LString(converted)
	case json.Number:
		return lua.LString(converted)
	case []interface{}:
		arr := l.CreateTable(len(converted), 0)
		for _, item := range converted {
			arr.Append(DecodeValue(l, item))
		}
		return arr
	case map[string]interface{}:
		tbl := l.CreateTable(0, len(converted))
		for key, item := range converted {
			tbl.RawSetH(lua.LString(key), DecodeValue(l, item))
		}
		return tbl
	case nil:
		return lua.LNil
	}

	return lua.LNil
}

func CommandValue(l *lua.LState) int { // key=val
	lv := l.Get(1)
	if lv.Type() != lua.LTString {
		l.TypeError(1, lua.LTString)
		return lua.MultRet
	}
	str := string(lv.(lua.LString))

	lv2 := l.Get(2)
	if lv.Type() != lua.LTString {
		l.TypeError(1, lua.LTString)
		return lua.MultRet
	}
	command := string(lv2.(lua.LString))
	// str : key=val
	cmds := strings.Split(str, ",")
	argL := 2
	for _, cmd := range cmds {
		if strings.Contains(cmd, command) {
			args := strings.Split(cmd, command)
			if len(args) == argL {
				l.Push(lua.LString(args[1]))
				return 1
			}
		}
	}
	l.Push(lua.LNil)
	return 1
}
