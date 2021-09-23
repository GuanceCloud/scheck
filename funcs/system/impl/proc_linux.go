//+build linux

package impl

import (
	"syscall"
)

var (
	LinuxProtocolNames = map[int]string{
		syscall.IPPROTO_ICMP:    "icmp",
		syscall.IPPROTO_TCP:     "tcp",
		syscall.IPPROTO_UDP:     "udp",
		syscall.IPPROTO_UDPLITE: "udplite",
		syscall.IPPROTO_RAW:     "raw",
	}
)
