// Package impl export for system and container
package impl

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	hostutil "github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/process"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/internal/global"
)

var (
	tcpStates = []string{
		"UNKNOWN",
		"ESTABLISHED",
		"SYN_SENT",
		"SYN_RECV",
		"FIN_WAIT1",
		"FIN_WAIT2",
		"TIME_WAIT",
		"CLOSED",
		"CLOSE_WAIT",
		"LAST_ACK",
		"LISTEN",
		"CLOSING",
	}
	userNamespaceList = []string{
		"cgroup",
		"ipc",
		"mnt",
		"net",
		"pid",
		"user",
		"uts",
	}
)

var log = logger.DefaultSLogger("proc")

type (
	pidFdPair struct {
		pid int
		fd  int
	}

	ProcessInfo struct {
		Pid              int
		Parent           int
		Name             string
		Path             string
		PGroup           string
		State            string
		Nice             string
		Threads          int
		Cmdline          string
		Cwd              string
		Root             string
		UID              string
		EUID             string
		SUID             string
		GID              string
		EGID             string
		SGID             string
		OnDisk           int
		ResidentSize     int64
		TotalSize        int64
		UserTime         int64
		SystemTime       int64
		StartTime        int64
		DiskBytesRead    int64
		DiskBytesWritten int64
	}

	HistoryItem struct {
		Time    int64
		Command string
	}

	UserInfo struct {
		Name  string
		UID   string
		GID   string
		Home  string
		Shell string
	}

	SocketInfo struct {
		Socket         string
		Family         int
		Protocol       int
		Namespace      int64
		LocalAddress   string
		LocalPort      uint16
		RemoteAddress  string
		RemotePort     uint16
		UnixSocketPath string
		State          string
		PID            int
		Fd             int
	}

	SimpleProcStat struct {
		Name         string
		RealUID      string
		RealGID      string
		EffectiveUID string
		EffectiveGID string
		SavedUID     string
		SavedGID     string
		ResidentSize string
		TotalSize    string
		State        string
		Parent       string
		Group        string
		Nice         string
		Threads      string
		UserTime     string
		SystemTime   string
		StartTime    string
	}

	SimpleProcIo struct {
		ReadBytes          int64
		WriteBytes         int64
		CanceledWriteBytes int64
	}

	ProcessDescriptor struct {
		PID      int
		Fd       string
		LinkPath string
	}
)

func getProcAttrFilePath(pid int, attr string) string {
	return fmt.Sprintf("/proc/%d/%s", pid, attr)
}

// nolint
func getSimpleProcStat(pid int) (*SimpleProcStat, error) {
	path := getProcAttrFilePath(pid, "stat")
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	content := string(data)
	start := strings.LastIndex(content, ")")
	if start == -1 || len(content) <= start+2 {
		err = fmt.Errorf("invalid /proc/stat header")
		return nil, err
	}

	details := strings.Split(content[start+2:], " ")
	if len(details) <= 19 {
		err = fmt.Errorf("invalid /proc/stat content")
		return nil, err
	}

	stat := &SimpleProcStat{}
	stat.State = details[0]
	stat.Parent = details[1]
	stat.Group = details[2]
	stat.UserTime = details[11]
	stat.SystemTime = details[12]
	stat.Nice = details[16]
	stat.Threads = details[17]
	stat.StartTime = details[19]

	path = getProcAttrFilePath(pid, "status")
	data, err = ioutil.ReadFile(path)
	if err == nil {
		content = string(data)
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			detail := strings.Split(line, ":")
			if len(detail) != 2 {
				continue
			}
			k := detail[0]
			v := strings.TrimSpace(detail[1])
			switch k {
			case "Name":
				stat.Name = v
			case "VmRSS":
				if len(v) > 3 {
					v = v[0 : len(v)-3]
				}
				stat.ResidentSize = v + "000"
			case "VmSize":
				if len(v) > 3 {
					v = v[0 : len(v)-3]
				}
				stat.TotalSize = v + "000"
			case "Gid":
				gidDetail := strings.Split(v, "\t")
				if len(gidDetail) == 4 {
					stat.RealGID = gidDetail[0]
					stat.EffectiveGID = gidDetail[1]
					stat.SavedGID = gidDetail[2]
				}
			case "Uid":
				gidDetail := strings.Split(v, "\t")
				if len(gidDetail) == 4 {
					stat.RealUID = gidDetail[0]
					stat.EffectiveUID = gidDetail[1]
					stat.SavedUID = gidDetail[2]
				}
			}
		}
	} else {
		log.Warnf("Cannot read /proc/status, %s", err)
	}

	return stat, nil
}

func getSimpleProcIo(pid int) (*SimpleProcIo, error) {
	path := getProcAttrFilePath(pid, "io")
	data, err := ioutil.ReadFile(filepath.Clean(path))
	if err != nil {
		log.Warnf("%s", err)
		return nil, err
	}
	info := &SimpleProcIo{}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		detail := strings.Split(line, ":")
		mustLen := 2
		if len(detail) != mustLen {
			continue
		}
		k := detail[0]
		v := strings.TrimSpace(detail[1])
		switch k {
		case "read_bytes":
			if n, e := strconv.ParseInt(v, global.ParseBase, global.ParseBitSize); e == nil {
				info.ReadBytes = n
			}
		case "write_bytes":
			if n, e := strconv.ParseInt(v, global.ParseBase, global.ParseBitSize); e == nil {
				info.WriteBytes = n
			}
		case "cancelled_write_bytes": // nolint:misspell
			if n, e := strconv.ParseInt(v, global.ParseBase, global.ParseBitSize); e == nil {
				info.CanceledWriteBytes = n
			}
		}
	}
	return info, nil
}

func readProcLink(pid int, attr string) string {
	path := getProcAttrFilePath(pid, attr)
	l, err := os.Readlink(path)
	if err != nil {
		return ""
	}
	return l
}

func readProcCMDLine(pid int) string {
	path := getProcAttrFilePath(pid, "cmdline")
	data, err := ioutil.ReadFile(filepath.Clean(path))
	if err != nil {
		log.Warnf("%s", err)
		return ""
	}
	for i, b := range data {
		if b == 0 {
			data[i] = ' '
		}
	}
	return strings.TrimSpace(string(data))
}

// nolint
func getOnDisk(pid int, path string) (string, int) {
	if path == "" {
		return "", -1
	}

	if !strings.HasSuffix(path, "deleted") {
		if _, err := os.Stat(path); err == nil {
			return path, 1
		}
		return path, 0
	}
	return strings.TrimSuffix(path, "deleted"), 0
}

func procEnumerateProcesses() ([]int, error) {
	var pids []int
	dir, err := ioutil.ReadDir("/proc")
	if err != nil {
		log.Errorf("%s", err)
		return nil, err
	}
	for _, fi := range dir {
		if !fi.IsDir() {
			continue
		}

		if n, err := strconv.Atoi(fi.Name()); err == nil && n > 0 {
			pids = append(pids, n)
		}
	}
	return pids, nil
}

func GetProcessSimpleInfo(pid int) (name, path, cmdline string, err error) {
	if pid < 0 {
		return
	}
	var procStat *SimpleProcStat
	procStat, err = getSimpleProcStat(pid)
	if err == nil {
		name = procStat.Name
		path = readProcLink(pid, "exe")
		cmdline = readProcCMDLine(pid)
	}
	return
}

func GetProcessInfo(pid int, systemBootTime int64) (*ProcessInfo, error) {
	if pid < 0 {
		return nil, nil
	}

	procStat, err := getSimpleProcStat(pid)
	if err != nil {
		return nil, err
	}

	procIO, _ := getSimpleProcIo(pid)

	info := &ProcessInfo{}
	info.Pid = pid
	var pn int
	if pn, err = strconv.Atoi(procStat.Parent); err == nil {
		info.Parent = pn
	} else {
		info.Parent = -1
	}
	info.Name = procStat.Name
	info.Path = readProcLink(pid, "exe")
	info.PGroup = procStat.Group
	info.State = procStat.State
	info.Nice = procStat.Nice
	if pn, err = strconv.Atoi(procStat.Threads); err == nil {
		info.Threads = pn
	}
	info.Cmdline = readProcCMDLine(pid)
	info.Cwd = readProcLink(pid, "cwd")
	info.Root = readProcLink(pid, "root")
	info.UID = procStat.RealUID
	info.EUID = procStat.EffectiveUID
	info.SUID = procStat.SavedUID
	info.GID = procStat.RealGID
	info.EGID = procStat.EffectiveGID
	info.SGID = procStat.SavedGID

	info.Path, info.OnDisk = getOnDisk(pid, info.Path)
	var n int64
	if n, err = strconv.ParseInt(procStat.ResidentSize, global.ParseBase, global.ParseBitSize); err == nil {
		info.ResidentSize = n
	}

	if n, err = strconv.ParseInt(procStat.TotalSize, global.ParseBase, global.ParseBitSize); err == nil {
		info.TotalSize = n
	}

	var kCLKTCK int64 = 100

	kMSIn1CLKTCK := 1000 / kCLKTCK

	if n, err = strconv.ParseInt(procStat.UserTime, global.ParseBase, global.ParseBitSize); err == nil {
		info.UserTime = n * kMSIn1CLKTCK
	}

	if n, err = strconv.ParseInt(procStat.SystemTime, global.ParseBase, global.ParseBitSize); err == nil {
		info.SystemTime = n * kMSIn1CLKTCK
	}

	if n, err = strconv.ParseInt(procStat.StartTime, global.ParseBase, global.ParseBitSize); err == nil {
		if systemBootTime > 0 {
			info.StartTime = systemBootTime + n/kCLKTCK
		}
	} else {
		info.StartTime = -1
	}

	if procIO != nil {
		info.DiskBytesRead = procIO.ReadBytes
		info.DiskBytesWritten = procIO.WriteBytes - procIO.CanceledWriteBytes
	}

	return info, err
}

func GetProcesses() ([]*ProcessInfo, error) {
	var systemBoot uint64
	if info, err := hostutil.Info(); err == nil {
		systemBoot = info.Uptime
		if systemBoot > 0 {
			systemBoot = uint64(time.Now().Unix()) - systemBoot
		}
	}

	pids, err := procEnumerateProcesses()
	if err != nil {
		return nil, err
	}

	var processes []*ProcessInfo
	for _, pid := range pids {
		if p, err := GetProcessInfo(pid, int64(systemBoot)); err == nil && p != nil {
			processes = append(processes, p)
		}
	}
	return processes, nil
}

func EnumProcessesFds(pids []int) ([]*ProcessDescriptor, error) {
	var err error
	if len(pids) == 0 {
		pids, err = procEnumerateProcesses()
		if err != nil {
			return nil, err
		}
	}

	var fds []*ProcessDescriptor
	for _, pid := range pids {
		if vals, err := GetProcessFds(pid); err == nil {
			fds = append(fds, vals...)
		}
	}
	return fds, nil
}

func GetProcessFds(pid int) ([]*ProcessDescriptor, error) {
	dir := fmt.Sprintf("/proc/%d/fd", pid)
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var fds []*ProcessDescriptor
	for _, file := range files {
		link, err := os.Readlink(filepath.Join(dir, file.Name()))
		if err == nil {
			fds = append(fds, &ProcessDescriptor{
				PID:      pid,
				Fd:       file.Name(),
				LinkPath: link,
			})
		}
	}
	return fds, nil
}

func procReadDescriptor(pid int, descriptor string) (string, error) {
	link := fmt.Sprintf("/proc/%d/fd/%s", pid, descriptor)
	return os.Readlink(link)
}

func procDecodeAddressFromHex(encodedAddress string, family int) string {
	var addr string
	afLen := 8
	af6Len := 32
	if family == syscall.AF_INET {
		if len(encodedAddress) == afLen {
			var decoded uint32
			if _, err := fmt.Sscanf(encodedAddress, "%X", &decoded); err == nil {
				addL := 4
				a := make([]byte, addL)
				binary.LittleEndian.PutUint32(a, decoded)
				ip4 := net.IPv4(a[0], a[1], a[2], a[3])
				if ip4 != nil {
					addr = ip4.String()
				}
			}
		}
	} else if family == syscall.AF_INET6 {
		if len(encodedAddress) == af6Len {
			var n1, n2, n3, n4 uint32
			if cnt, err := fmt.Sscanf(encodedAddress, "%8x%8x%8x%8x", &n1, &n2, &n3, &n4); err == nil && cnt == 4 {
				ip6 := net.IPv6zero
				binary.LittleEndian.PutUint32(ip6[:4], n1)
				binary.LittleEndian.PutUint32(ip6[4:8], n2)
				binary.LittleEndian.PutUint32(ip6[8:12], n3)
				binary.LittleEndian.PutUint32(ip6[12:16], n4)
				addr = ip6.String()
			}
		}
	}
	return addr
}

func procDecodePortFromHex(encodedPort string) uint16 {
	var port uint16
	if len(encodedPort) == 4 {
		if _, err := fmt.Sscanf(encodedPort, "%X", &port); err != nil {
			log.Errorf("fail to convert port %s, error: %s", encodedPort, err)
		}
	}
	return port
}

func procGetSocketInodeToProcessInfoMap(pid int, infoMap map[string]pidFdPair) error {
	descriptorsPath := fmt.Sprintf("/proc/%d/fd", pid)

	dir, err := ioutil.ReadDir(descriptorsPath)
	if err != nil {
		return err
	}

	for _, fi := range dir {
		if fi.IsDir() {
			continue
		}
		fd := fi.Name()
		if link, err := procReadDescriptor(pid, fd); err == nil {
			if strings.Index(link, "socket:[") == 0 {
				inode := link[8 : len(link)-1]
				if nfd, err := strconv.Atoi(fd); err == nil {
					infoMap[inode] = pidFdPair{pid, nfd}
				}
			}
		}
	}
	return nil
}

func procGetProcessNamespaces(pid int, namespaces []string) map[string]int64 {
	if len(namespaces) == 0 {
		namespaces = userNamespaceList
	}

	namespaceInfo := map[string]int64{}

	processNamespaceRoot := fmt.Sprintf("/proc/%d/ns", pid)

	for _, name := range namespaces {
		inode, err := procGetNamespaceInode(name, processNamespaceRoot)
		if err == nil && inode > 0 {
			namespaceInfo[name] = inode
		}
	}
	return namespaceInfo
}

func procGetNamespaceInode(namespaceName, processNamespaceRoot string) (int64, error) {
	path := processNamespaceRoot + "/" + namespaceName
	link, err := os.Readlink(path)
	if err != nil {
		return 0, err
	}
	if len(link) < len(namespaceName)+2 {
		return 0, nil
	}

	if !strings.HasPrefix(link, namespaceName+":[") {
		return 0, nil
	}

	st := len(namespaceName + ":[")
	link = link[st:]
	ed := strings.Index(link, "]")
	if ed != -1 {
		if n, err := strconv.ParseInt(link[:ed], global.ParseBase, global.ParseBitSize); err == nil {
			return n, nil
		}
	}
	return 0, nil
}

func procGetSocketListInet(family, protocol int, ns int64, path, content string) ([]*SocketInfo, error) {
	lines := strings.Split(content, "\n")
	header := true
	var infolist []*SocketInfo
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if header {
			if strings.Index(line, "sl") != 0 && strings.Index(line, "sk") != 0 {
				return nil, fmt.Errorf("invalid file header for %s, header: %s", path, line)
			}
			header = false
			continue
		}

		if line == "" {
			continue
		}

		fields := strings.FieldsFunc(line, func(r rune) bool {
			return r == ' '
		})
		fieldsL := 10
		if len(fields) < fieldsL {
			log.Warnf("invalid socket descriptor found for %s, line: %s", path, line)
			continue
		}

		locals := strings.Split(fields[1], ":")
		remotes := strings.Split(fields[2], ":")

		if len(locals) != 2 || len(remotes) != 2 {
			log.Warnf("invalid socket address found for %s, local=%s, remote=%s", path, fields[1], fields[2])
			continue
		}

		var info SocketInfo
		info.Socket = fields[fieldsL-1]
		info.Namespace = ns
		info.Family = family
		info.Protocol = protocol
		info.LocalAddress = procDecodeAddressFromHex(locals[0], family)
		info.LocalPort = procDecodePortFromHex(locals[1])
		info.RemoteAddress = procDecodeAddressFromHex(remotes[0], family)
		info.RemotePort = procDecodePortFromHex(remotes[1])
		if protocol == syscall.IPPROTO_TCP {
			info.State = "UNKNOWN"
			if n, err := strconv.ParseUint(fields[3], global.ParseBase16, global.ParseBitSize); err == nil {
				if n < uint64(len(tcpStates)) {
					info.State = tcpStates[int(n)]
				}
			}
		}
		infolist = append(infolist, &info)
	}
	return infolist, nil
}

func procGetSocketListUnix(ns int64, path, content string) ([]*SocketInfo, error) {
	var infolist []*SocketInfo
	lines := strings.Split(content, "\n")
	header := true
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if header {
			if strings.Index(line, "Num") != 0 {
				return nil, fmt.Errorf("invalid file header for %s, header: %s", path, line)
			}
			header = false
			continue
		}

		if line == "" {
			continue
		}

		fields := strings.FieldsFunc(line, func(r rune) bool {
			return r == ' '
		})
		lisL := 7
		if len(fields) < lisL {
			log.Warnf("invalid socket descriptor found for %s, line: %s", path, line)
			continue
		}

		var info SocketInfo
		info.Socket = fields[lisL-1]
		info.Namespace = ns
		info.Family = syscall.AF_UNIX
		if p, err := strconv.Atoi(fields[2]); err == nil {
			info.Protocol = p
		}
		if len(fields) >= lisL+1 {
			info.UnixSocketPath = fields[lisL]
		}
		infolist = append(infolist, &info)
	}
	return infolist, nil
}

func procGetSocketList(family, protocol, pid int, ns int64) ([]*SocketInfo, error) {
	path := fmt.Sprintf("/proc/%d/net/", pid)

	switch family {
	case syscall.AF_INET:
		if name, ok := LinuxProtocolNames[protocol]; ok {
			path += name
		} else {
			return nil, fmt.Errorf("invalid protocol %d for AF_INET family", protocol)
		}
	case syscall.AF_INET6:
		if name, ok := LinuxProtocolNames[protocol]; ok {
			path += name + "6"
		} else {
			return nil, fmt.Errorf("invalid protocol %d for AF_INET6 family", protocol)
		}
	case syscall.AF_UNIX:
		if protocol == syscall.IPPROTO_IP {
			path += "unix"
		} else {
			return nil, fmt.Errorf("invalid protocol %d for AF_UNIX family", protocol)
		}
	default:
		return nil, fmt.Errorf("invalid family %d", family)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, nil
	}

	var content string
	data, err := ioutil.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, err
	}
	content = string(data)

	switch family {
	case syscall.AF_INET, syscall.AF_INET6:
		return procGetSocketListInet(family, protocol, ns, path, content)
	default:
		return procGetSocketListUnix(ns, path, content)
	}
}

func EnumProOpenSockets(pids []int) ([]*SocketInfo, error) { // nolint
	var err error
	if len(pids) == 0 {
		pids, err = procEnumerateProcesses()
		if err != nil {
			return nil, err
		}
	}

	netnsList := map[int64]bool{}
	inodeProcMap := map[string]pidFdPair{}
	var socketList []*SocketInfo

	for _, pid := range pids {
		err = procGetSocketInodeToProcessInfoMap(pid, inodeProcMap)
		if err != nil {
			log.Errorf("%s", err)
		}
		var ns int64
		namespaces := procGetProcessNamespaces(pid, []string{"net"})
		if len(namespaces) > 0 {
			ns = namespaces["net"]
		}

		if _, ok := netnsList[ns]; !ok {
			netnsList[ns] = true
			for k := range LinuxProtocolNames {
				if list, err := procGetSocketList(syscall.AF_INET, k, pid, ns); err == nil && len(list) > 0 {
					socketList = append(socketList, list...)
				} else if err != nil {
					log.Errorf("%s", err)
				}

				if list, err := procGetSocketList(syscall.AF_INET6, k, pid, ns); err == nil && len(list) > 0 {
					socketList = append(socketList, list...)
				} else if err != nil {
					log.Errorf("%s", err)
				}
			}
			if list, err := procGetSocketList(syscall.AF_UNIX, syscall.IPPROTO_IP, pid, ns); err == nil && len(list) > 0 {
				socketList = append(socketList, list...)
			} else if err != nil {
				log.Errorf("%s", err)
			}
		}
	}

	for _, info := range socketList {
		if it, ok := inodeProcMap[info.Socket]; ok {
			info.PID = it.pid
			info.Fd = it.fd
		} else {
			info.PID = -1
			info.Fd = -1
		}
	}
	return socketList, nil
}

func GetUserDetail(username string) ([]*UserInfo, error) {
	file, err := os.Open("/etc/passwd")
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = file.Close()
	}()

	var result []*UserInfo
	reader := bufio.NewReader(file)
	passedL := 7
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}

		fields := strings.Split(line, ":")
		if len(fields) < passedL {
			continue
		}

		if username != "" {
			if fields[0] != username {
				continue
			}
		}

		u := &UserInfo{
			Name:  fields[0],
			UID:   fields[2],
			GID:   fields[3],
			Home:  fields[5],
			Shell: strings.TrimSpace(fields[6]),
		}
		result = append(result, u)
	}

	return result, err
}

func GenShellHistoryFromFile(file string) ([]*HistoryItem, error) {
	var err error
	if _, err = os.Stat(file); err != nil && os.IsNotExist(err) {
		return nil, nil
	}

	content, err := ioutil.ReadFile(filepath.Clean(file))
	if err != nil {
		return nil, err
	}

	bashTimestampRx := regexp.MustCompile("^#([0-9]+)$")                        // nolint
	zshTimestampRx := regexp.MustCompile("^: {0,10}([0-9]{1,11}):[0-9]+;(.*)$") // nolint
	_ = bashTimestampRx
	_ = zshTimestampRx

	prevBashTimestamp := ""

	var cmds []*HistoryItem
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		var tm int64
		if prevBashTimestamp == "" {
			prevBashTimestamp = bashTimestampRx.FindString(line)
		}
		if prevBashTimestamp != "" {
			if t, e := strconv.ParseInt(prevBashTimestamp[1:], global.ParseBase, global.ParseBitSize); e == nil {
				tm = t
			}
			prevBashTimestamp = ""
		}
		cmds = append(cmds, &HistoryItem{
			Command: line,
			Time:    tm,
		})
	}
	return cmds, err
}

func GetListeningPorts() []map[string]interface{} {
	var socketList []*SocketInfo

	pids := make([]int, 0)
	socketList, _ = EnumProOpenSockets(pids)

	listenPortList := make([]map[string]interface{}, 0)
	for _, info := range socketList {
		if info.Family == syscall.AF_UNIX && info.UnixSocketPath == "" {
			continue
		}

		if (info.Family == syscall.AF_INET || info.Family == syscall.AF_INET6) && info.RemotePort != 0 {
			continue
		}

		item := make(map[string]interface{})
		item["pid"] = info.PID
		item["state"] = info.State

		if pname, _, pcmdline, err := GetProcessSimpleInfo(info.PID); err == nil {
			item["process_name"] = pname
			item["cmdline"] = pcmdline
		}

		if info.Family == syscall.AF_UNIX {
			item["port"] = 0
			item["path"] = info.UnixSocketPath
			item["socket"] = 0
			item["family"] = "AF_UNIX"
			item["protocol"] = "ip"
		} else {
			item["port"] = info.LocalPort
			item["address"] = info.LocalAddress
			item["socket"] = info.Socket
			if info.Family == syscall.AF_INET {
				item["family"] = "AF_INET"
			} else if info.Family == syscall.AF_INET6 {
				item["family"] = "AF_INET6"
			}
			item["protocol"] = LinuxProtocolNames[info.Protocol]
		}
		listenPortList = append(listenPortList, item)
	}
	return listenPortList
}

// GetCommandLine :获取一个进程的命令行参数.
func GetCommandLine(processName string) string {
	var systemBoot uint64
	if info, err := hostutil.Info(); err == nil {
		systemBoot = info.Uptime
		if systemBoot > 0 {
			systemBoot = uint64(time.Now().Unix()) - systemBoot
		}
	}

	pids, err := process.Pids()
	if err != nil {
		return ""
	}

	for _, pid := range pids {
		if p, err := GetProcessInfo(int(pid), int64(systemBoot)); err == nil && p != nil {
			if p.Name == processName {
				return p.Cmdline
			}
		}
	}
	return ""
}

//  GetCmdline 将命令行参数 解析到map中
func GetCmdline(cmdline string) map[string]string {
	cmds := strings.Split(cmdline, " ")
	if len(cmds) == 0 {
		return nil
	}
	arguments := make(map[string]string)
	cmds = cmds[1:] // first one command name
	argL := 2
	for len(cmds) > 0 {
		// --开始 判断有没有 = 没有 ，判断下一个是不是-或者==开头 不是的话 赋值前一个，是的话 控制。
		cmd := cmds[0]
		isLong := len(cmd) > 2 && cmd[0:2] == "--"
		isShort := len(cmd) > 2 && cmd[0] == '-' && cmd[1] != '-'
		if isLong || isShort {
			var key string
			if isShort {
				key = cmd[1:]
			} else {
				key = cmd[2:]
			}
			if strings.Contains(key, "=") {
				subs := strings.Split(key, "=")
				if len(subs) == argL {
					if str, ok := arguments[subs[0]]; ok { // 有可能出现两个相同的key 值需要用逗号分隔 并进行拼接。
						arguments[subs[0]] = str + "," + subs[1]
					} else {
						arguments[subs[0]] = subs[1]
					}
				} else if len(subs) > argL {
					index := strings.Index(key, "=") // key : abc=d=true
					arguments[subs[0]] = key[index+1:]
				}
				cmds = cmds[1:]
			} else { // nolint
				// 没有等号 值有可能在后一个: -v 1.3 。也可能没有： -i -t
				if len(cmds) > 1 {
					if strings.HasPrefix(cmds[1], "-") {
						arguments[key] = ""
						cmds = cmds[1:]
					} else {
						arguments[key] = cmds[1]
						cmds = cmds[2:]
					}
				}
			}
		} else {
			// 没有-开头  进入下一个
			if len(cmds) > 1 {
				cmds = cmds[1:]
			} else {
				break
			}
		}
	}
	return arguments
}
