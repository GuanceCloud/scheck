package utils

import (
	"bufio"
	"database/sql"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	lua "github.com/yuin/gopher-lua"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/config"
)

type MySqlServiceInfo struct {
	Host     string
	Port     string
	UserName string
	PassWord string
}

type MySqlChecker struct {
	running int32

	dictPath string
	scanner  *bufio.Scanner

	lastErr error
}

func (c *MySqlChecker) Check(host string, port string, user string) (hit bool, pwd string, err error) {

	var fd *os.File
	fd, err = os.Open(c.dictPath)
	if err != nil {
		return
	}
	defer fd.Close()
	c.scanner = bufio.NewScanner(fd)

	var info MySqlServiceInfo
	info.Host = host
	info.Port = port
	info.UserName = user

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
		time.Sleep(time.Millisecond * 50)
	}
}

func (c *MySqlChecker) hit(info *MySqlServiceInfo) (bool, error) {

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

	var mysqlChecker MySqlChecker
	mysqlChecker.dictPath = filepath.Join(config.Cfg.System.RuleDir, `passwd_dict`, `dict.txt`)

	lv := l.Get(1)
	if lv.Type() != lua.LTString {
		l.TypeError(1, lua.LTString)
		return lua.MultRet
	}
	host := lv.(lua.LString).String()

	lv = l.Get(2)
	if lv.Type() != lua.LTString {
		l.TypeError(2, lua.LTString)
		return lua.MultRet
	}
	port := lv.(lua.LString).String()

	user := "root"
	lv = l.Get(3)
	if lv != lua.LNil {
		if lv.Type() != lua.LTString {
			l.TypeError(3, lua.LTString)
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
		return 2
	}

	return 1
}
