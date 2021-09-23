package utils

import (
	"bufio"
	"database/sql"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	lua "github.com/yuin/gopher-lua"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/config"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/internal/global"
)

type MySQLServiceInfo struct {
	Host     string
	Port     string
	UserName string
	PassWord string
}

type MySQLChecker struct {
	dictPath string
	scanner  *bufio.Scanner
}

func (c *MySQLChecker) Check(host, port, user string) (hit bool, pwd string, err error) {
	var fd *os.File
	fd, err = os.Open(c.dictPath)
	if err != nil {
		return
	}
	defer fd.Close()
	c.scanner = bufio.NewScanner(fd)

	var info MySQLServiceInfo
	info.Host = host
	info.Port = port
	info.UserName = user

	var timeSleep = 50
	for {
		if !c.scanner.Scan() {
			err = c.scanner.Err()
			if err == nil {
				err = io.EOF
			}
			return
		}
		line := c.scanner.Text()
		line = strings.TrimSpace(line)
		info.PassWord = line
		hit, err = c.hit(&info)
		if err != nil {
			return
		}
		if hit {
			pwd = line
			return
		}
		time.Sleep(time.Millisecond * time.Duration(timeSleep))
	}
}

func (c *MySQLChecker) hit(info *MySQLServiceInfo) (bool, error) {
	source := info.UserName + ":" + info.PassWord + "@tcp(" + info.Host + ":" + info.Port + ")/mysql?charset=utf8"
	db, err := sql.Open("mysql", source)
	if err != nil {
		return false, err
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		return false, nil
	}
	return true, nil
}

func (p *provider) checkMysqlWeakPassword(l *lua.LState) int {
	var mysqlChecker MySQLChecker
	mysqlChecker.dictPath = filepath.Join(config.Cfg.System.RuleDir, `passwd_dict`, `dict.txt`)

	lv := l.Get(1)
	if lv.Type() != lua.LTString {
		l.TypeError(1, lua.LTString)
		return lua.MultRet
	}
	host := lv.(lua.LString).String()

	lv = l.Get(global.LuaArgIdx2)
	if lv.Type() != lua.LTString {
		l.TypeError(global.LuaArgIdx2, lua.LTString)
		return lua.MultRet
	}
	port := lv.(lua.LString).String()

	user := "root"
	lv = l.Get(global.LuaArgIdx3)
	if lv != lua.LNil {
		if lv.Type() != lua.LTString {
			l.TypeError(global.LuaArgIdx3, lua.LTString)
			return lua.MultRet
		}
		user = lv.(lua.LString).String()
	}

	hit, pwd, err := mysqlChecker.Check(host, port, user)

	if err != nil && err != io.EOF {
		l.RaiseError("%s", err)
		return lua.MultRet
	}

	l.Push(lua.LBool(hit))
	if hit {
		l.Push(lua.LString(pwd))
		return global.LuaRet2
	}
	return 1
}
