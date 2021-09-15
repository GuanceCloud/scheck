//+build windows

package impl

import (
	"syscall"
)

var (
	LinuxProtocolNames = map[int]string{
		syscall.IPPROTO_TCP: "tcp",
		syscall.IPPROTO_UDP: "udp",
	}
)
