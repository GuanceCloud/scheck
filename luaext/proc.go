package luaext

import (
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"syscall"
)

var (
	linuxProtocolNames = map[int]string{
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
			log.Printf("[error] fail to convert port %s, error: %s", encodedPort, err)
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

func procGetSocketListInet(family int, protocol int, ns int64, path string, content string) ([]*socketInfo, error) {
	lines := strings.Split(content, "\n")
	header := true
	var infolist []*socketInfo
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
			log.Printf("[warn] invalid socket descriptor found for %s, line: %s", path, line)
			continue
		}

		locals := strings.Split(fields[1], ":")
		remotes := strings.Split(fields[2], ":")

		if len(locals) != 2 || len(remotes) != 2 {
			log.Printf("[warn] invalid socket address found for %s, local=%s, remote=%s", path, fields[1], fields[2])
			continue
		}

		// if path == "/proc/1/net/tcp" {
		// 	//log.Printf("### path: %s", path)
		// 	log.Printf("### line: %s", line)
		// 	log.Printf("### %d, fields: %v", len(fields), fields)
		// }

		var info socketInfo
		info.socket = fields[9]
		info.ns = ns
		info.family = family
		info.protocol = protocol
		info.localAddress = procDecodeAddressFromHex(locals[0], family)
		info.localPort = procDecodePortFromHex(locals[1])
		info.remoteAddress = procDecodeAddressFromHex(remotes[0], family)
		info.remotePort = procDecodePortFromHex(remotes[1])
		if protocol == syscall.IPPROTO_TCP {
			info.state = "UNKNOWN"
			if n, err := strconv.ParseUint(fields[3], 16, 64); err == nil {
				if n >= 0 && n < uint64(len(tcpStates)) {
					info.state = tcpStates[int(n)]
				}
			}
		}

		infolist = append(infolist, &info)
	}

	return infolist, nil
}

func procGetSocketListUnix(ns int64, path, content string) ([]*socketInfo, error) {
	var infolist []*socketInfo
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
			log.Printf("[warn] invalid socket descriptor found for %s, line: %s", path, line)
			continue
		}

		var info socketInfo
		info.socket = fields[6]
		info.ns = ns
		info.family = syscall.AF_UNIX
		if p, err := strconv.Atoi(fields[2]); err == nil {
			info.protocol = p
		}
		if len(fields) >= 8 {
			info.unixSocketPath = fields[7]
		}
		infolist = append(infolist, &info)
	}

	return infolist, nil
}

func procGetSocketList(family int, protocol int, ns int64, pid int) ([]*socketInfo, error) {
	path := fmt.Sprintf("/proc/%d/net/", pid)

	switch family {
	case syscall.AF_INET:
		if name, ok := linuxProtocolNames[protocol]; ok {
			path += name
		} else {
			return nil, fmt.Errorf("Invalid protocol %d for AF_INET family", protocol)
		}

	case syscall.AF_INET6:
		if name, ok := linuxProtocolNames[protocol]; ok {
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

func getProcessOpenSockets() ([]*socketInfo, error) {

	var pis []*processInfo
	var err error

	pis, err = procEnumerateProcesses(true)
	if err != nil {
		return nil, err
	}

	netnsList := map[int64]bool{}
	inodeProcMap := map[string]pidFdPair{}
	var socketList []*socketInfo

	for _, pi := range pis {

		pid := pi.pid

		err = procGetSocketInodeToProcessInfoMap(pid, inodeProcMap)
		if err != nil {
			log.Printf("[error] %s", err)
		}

		var ns int64
		namespaces := procGetProcessNamespaces(pid, []string{"net"})
		if len(namespaces) > 0 {
			ns = namespaces["net"]
		}

		if _, ok := netnsList[ns]; !ok {
			netnsList[ns] = true

			log.Printf("xxxx - %d", pid)

			for k := range linuxProtocolNames {
				if list, err := procGetSocketList(syscall.AF_INET, k, ns, pid); err != nil {
					log.Printf("[error] %s", err)
				} else {
					socketList = append(socketList, list...)
				}

				if list, err := procGetSocketList(syscall.AF_INET6, k, ns, pid); err != nil {
					log.Printf("[error] %s", err)
				} else {
					socketList = append(socketList, list...)
				}
			}

			if list, err := procGetSocketList(syscall.AF_UNIX, syscall.IPPROTO_IP, ns, pid); err != nil {
				log.Printf("[error] %s", err)
			} else {
				socketList = append(socketList, list...)
			}
		}
	}

	for _, info := range socketList {
		if it, ok := inodeProcMap[info.socket]; ok {
			info.pid = it.pid
			info.fd = it.fd

			for _, pi := range pis {
				if pi.pid == info.pid {
					info.pname = pi.name
					info.pcmdline = pi.cmdline
					break
				}
			}

		} else {
			info.pid = -1
			info.fd = "-1"
		}
	}

	return socketList, nil
}
