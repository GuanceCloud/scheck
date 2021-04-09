package impl

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
	"syscall"

	log "github.com/sirupsen/logrus"
)

var (
	LinuxProtocolNames = map[int]string{
		syscall.IPPROTO_ICMP:    "icmp",
		syscall.IPPROTO_TCP:     "tcp",
		syscall.IPPROTO_UDP:     "udp",
		syscall.IPPROTO_UDPLITE: "udplite",
		syscall.IPPROTO_RAW:     "raw",
	}

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

type (
	pidFdPair struct {
		pid int
		fd  string
	}

	processInfo struct {
		pid     int
		name    string
		cmdline string
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

	lastItemInfo struct {
		username string
		pid      int
		typ      int
		tty      string
		tm       int64
		host     string
	}

	SocketInfo struct {
		Socket    string
		Family    int
		Protocol  int
		Namespace int64

		LocalAddress  string
		LocalPort     uint16
		RemoteAddress string
		RemotePort    uint16

		UnixSocketPath string

		State string

		PID            int
		ProcessName    string
		ProcessCmdline string
		Fd             string
	}
)

func procGetProcessInfo(pi *processInfo) {

	path := fmt.Sprintf("/proc/%d/status", pi.pid)
	data, err := ioutil.ReadFile(path)
	if err == nil {
		content := string(data)
		lines := strings.Split(content, "\n")
		for _, line := range lines {
			if strings.Index(line, "Name:") != 0 {
				continue
			}
			fields := strings.Split(line, ":")
			if len(fields) > 1 {
				pi.name = strings.TrimSpace(fields[1])
			}
			break
		}
	}

	path = fmt.Sprintf("/proc/%d/cmdline", pi.pid)
	data, err = ioutil.ReadFile(path)
	if err == nil {
		pi.cmdline = strings.TrimSpace(string(data))
	}
}

func procEnumerateProcesses(detail bool) ([]*processInfo, error) {
	var ps []*processInfo
	dir, err := ioutil.ReadDir("/proc")
	if err != nil {
		return nil, err
	}
	for _, fi := range dir {
		if !fi.IsDir() {
			continue
		}

		if n, err := strconv.Atoi(fi.Name()); err == nil && n > 0 {
			pi := &processInfo{
				pid: n,
			}
			if detail {
				procGetProcessInfo(pi)
			}
			ps = append(ps, pi)
		}
	}
	return ps, nil
}

func procReadDescriptor(pid int, descriptor string) (string, error) {

	link := fmt.Sprintf("/proc/%d/fd/%s", pid, descriptor)
	return os.Readlink(link)
}

func procDecodeAddressFromHex(encodedAddress string, family int) string {
	var addr string
	if family == syscall.AF_INET {
		if len(encodedAddress) == 8 {
			var decoded uint32
			if _, err := fmt.Sscanf(encodedAddress, "%X", &decoded); err == nil {
				a := make([]byte, 4)
				binary.LittleEndian.PutUint32(a, decoded)
				ip4 := net.IPv4(a[0], a[1], a[2], a[3])
				if ip4 != nil {
					addr = ip4.String()
				}
			}
		}

	} else if family == syscall.AF_INET6 {
		if len(encodedAddress) == 32 {
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
				//log.Printf("[debug] link=%s for %s/%s, inode=%s", link, descriptorsPath, fd, inode)
				infoMap[inode] = pidFdPair{pid, fd}
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
		if n, err := strconv.ParseInt(link[:ed], 10, 64); err == nil {
			return n, nil
		}
	}
	return 0, nil
}

func procGetSocketListInet(family int, protocol int, ns int64, path string, content string) ([]*SocketInfo, error) {
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
			if r == ' ' {
				return true
			}
			return false
		})
		if len(fields) < 10 {
			log.Warnf("invalid socket descriptor found for %s, line: %s", path, line)
			continue
		}

		locals := strings.Split(fields[1], ":")
		remotes := strings.Split(fields[2], ":")

		if len(locals) != 2 || len(remotes) != 2 {
			log.Warnf("invalid socket address found for %s, local=%s, remote=%s", path, fields[1], fields[2])
			continue
		}

		// if path == "/proc/1/net/tcp" {
		// 	//log.Printf("### path: %s", path)
		// 	log.Printf("### line: %s", line)
		// 	log.Printf("### %d, fields: %v", len(fields), fields)
		// }

		var info SocketInfo
		info.Socket = fields[9]
		info.Namespace = ns
		info.Family = family
		info.Protocol = protocol
		info.LocalAddress = procDecodeAddressFromHex(locals[0], family)
		info.LocalPort = procDecodePortFromHex(locals[1])
		info.RemoteAddress = procDecodeAddressFromHex(remotes[0], family)
		info.RemotePort = procDecodePortFromHex(remotes[1])
		if protocol == syscall.IPPROTO_TCP {
			info.State = "UNKNOWN"
			if n, err := strconv.ParseUint(fields[3], 16, 64); err == nil {
				if n >= 0 && n < uint64(len(tcpStates)) {
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
			if r == ' ' {
				return true
			}
			return false
		})
		if len(fields) < 7 {
			log.Warnf("invalid socket descriptor found for %s, line: %s", path, line)
			continue
		}

		var info SocketInfo
		info.Socket = fields[6]
		info.Namespace = ns
		info.Family = syscall.AF_UNIX
		if p, err := strconv.Atoi(fields[2]); err == nil {
			info.Protocol = p
		}
		if len(fields) >= 8 {
			info.UnixSocketPath = fields[7]
		}
		infolist = append(infolist, &info)
	}

	return infolist, nil
}

func procGetSocketList(family int, protocol int, ns int64, pid int) ([]*SocketInfo, error) {
	path := fmt.Sprintf("/proc/%d/net/", pid)

	switch family {
	case syscall.AF_INET:
		if name, ok := LinuxProtocolNames[protocol]; ok {
			path += name
		} else {
			return nil, fmt.Errorf("Invalid protocol %d for AF_INET family", protocol)
		}

	case syscall.AF_INET6:
		if name, ok := LinuxProtocolNames[protocol]; ok {
			path += name + "6"
		} else {
			return nil, fmt.Errorf("Invalid protocol %d for AF_INET6 family", protocol)
		}

	case syscall.AF_UNIX:
		if protocol == syscall.IPPROTO_IP {
			path += "unix"
		} else {
			return nil, fmt.Errorf("Invalid protocol %d for AF_UNIX family", protocol)
		}

	default:
		return nil, fmt.Errorf("Invalid family %d", family)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, nil
	}

	var content string
	if data, err := ioutil.ReadFile(path); err != nil {
		return nil, err
	} else {
		content = string(data)
	}

	switch family {
	case syscall.AF_INET, syscall.AF_INET6:
		return procGetSocketListInet(family, protocol, ns, path, content)
	default:
		return procGetSocketListUnix(ns, path, content)
	}
}

func GetProcessOpenSockets() ([]*SocketInfo, error) {

	var pis []*processInfo
	var err error

	pis, err = procEnumerateProcesses(true)
	if err != nil {
		return nil, err
	}

	netnsList := map[int64]bool{}
	inodeProcMap := map[string]pidFdPair{}
	var socketList []*SocketInfo

	for _, pi := range pis {

		pid := pi.pid

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
				if list, err := procGetSocketList(syscall.AF_INET, k, ns, pid); err == nil && len(list) > 0 {
					socketList = append(socketList, list...)
				} else {
					if err != nil {
						log.Errorf("%s", err)
					}
				}

				if list, err := procGetSocketList(syscall.AF_INET6, k, ns, pid); err == nil && len(list) > 0 {
					socketList = append(socketList, list...)
				} else {
					if err != nil {
						log.Errorf("%s", err)
					}
				}
			}

			if list, err := procGetSocketList(syscall.AF_UNIX, syscall.IPPROTO_IP, ns, pid); err == nil && len(list) > 0 {
				socketList = append(socketList, list...)
			} else {
				if err != nil {
					log.Errorf("%s", err)
				}
			}
		}
	}

	for _, info := range socketList {
		if it, ok := inodeProcMap[info.Socket]; ok {
			info.PID = it.pid
			info.Fd = it.fd

			for _, pi := range pis {
				if pi.pid == info.PID {
					info.ProcessName = pi.name
					info.ProcessCmdline = pi.cmdline
					break
				}
			}

		} else {
			info.PID = -1
			info.Fd = "-1"
		}
	}

	return socketList, nil
}

func GetUserDetail(username string) ([]*UserInfo, error) {

	file, err := os.Open("/etc/passwd")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var result []*UserInfo

	reader := bufio.NewReader(file)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}

		fields := strings.Split(line, ":")
		if len(fields) < 7 {
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

	return result, nil
}

func GenShellHistoryFromFile(file string) ([]*HistoryItem, error) {
	var err error

	if _, err = os.Stat(file); err != nil && os.IsNotExist(err) {
		return nil, nil
	}

	content, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	bashTimestampRx, _ := regexp.Compile("^#([0-9]+)$")
	zshTimestampRx, _ := regexp.Compile("^: {0,10}([0-9]{1,11}):[0-9]+;(.*)$")
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
			if t, e := strconv.ParseInt(prevBashTimestamp[1:], 10, 64); e == nil {
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

func getLasts() ([]*lastItemInfo, error) {

	// cstr := C.CString("")
	// if nok, err := C.getlast_start(cstr); nok == 0 {
	// 	log.Printf("[error] %s", err)
	// 	return nil, err
	// }

	// var lasts []*lastItemInfo

	// const USER_PROCESS = 7
	// const DEAD_PROCESS = 8

	// for {
	// 	st := C.getlast()
	// 	if st == nil {
	// 		break
	// 	}
	// 	if st.ut_type == USER_PROCESS || st.ut_type == DEAD_PROCESS {
	// 		//log.Printf("%s", C.GoString(&st.ut_user[0]))

	// 		item := &lastItemInfo{
	// 			username: C.GoString(&st.ut_user[0]),
	// 			tty:      C.GoString(&st.ut_line[0]),
	// 			pid:      int(st.ut_pid),
	// 			typ:      int(st.ut_type),
	// 			tm:       int64(&st.ut_time),
	// 			host:     C.GoString(&st.ut_host[0]),
	// 		}
	// 		lasts = append(lasts, item)
	// 	}
	// }

	// C.getlast_end()

	return nil, nil
}