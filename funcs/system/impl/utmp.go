package impl

import (
	"bytes"
	"encoding/binary"
	"io"
	"net"
	"time"
)

const (
	lineSize = 32
	nameSize = 32
	hostSize = 256
	// Display time format
	TimeFormat = time.RFC1123
)

type (
	// utmp structures
	// see man utmp
	ExitStatus struct {
		Termination int16
		Exit        int16
	}

	timeVal struct {
		Sec  int32
		Usec int32
	}

	utmpImpl struct {
		Type int16
		// alignment
		_       [2]byte
		Pid     int32
		Device  [lineSize]byte
		Id      [4]byte
		User    [nameSize]byte
		Host    [hostSize]byte
		Exit    ExitStatus
		Session int32
		Time    timeVal
		AddrV6  [16]byte
		// Reserved member
		Reserved [20]byte
	}

	Utmp struct {
		Type    int
		Pid     int
		Device  string
		Id      string
		User    string
		Host    string
		Exit    ExitStatus
		Session int
		Time    int64
		Addr    string
	}
)

func ParseUtmp(file io.Reader) ([]*Utmp, error) {
	var us []*Utmp

	for {
		u, readErr := readLine(file)
		if readErr != nil {
			if readErr == io.EOF {
				break
			}
			return nil, readErr
		}
		us = append(us, newUtmp(u))
	}

	return us, nil
}

// read utmp
func readLine(file io.Reader) (*utmpImpl, error) {
	u := new(utmpImpl)

	err := binary.Read(file, binary.LittleEndian, u)
	if err != nil {
		return nil, err
	}

	return u, nil
}

func (r *utmpImpl) addr() net.IP {
	ip := make(net.IP, 16)
	// no error checking: reading from r.AddrV6 cannot fail
	binary.Read(bytes.NewReader(r.AddrV6[:]), binary.BigEndian, ip)
	if bytes.Equal(ip[4:], net.IPv6zero[4:]) {
		// IPv4 address, shorten the slice so that net.IP behaves correctly:
		ip = ip[:4]
	}
	return ip
}

func newUtmp(u *utmpImpl) *Utmp {
	return &Utmp{
		Type:    int(u.Type),
		Pid:     int(u.Pid),
		Device:  string(u.Device[:getByteLen(u.Device[:])]),
		Id:      string(u.Id[:getByteLen(u.Id[:])]),
		User:    string(u.User[:getByteLen(u.User[:])]),
		Host:    string(u.Host[:getByteLen(u.Host[:])]),
		Exit:    u.Exit,
		Session: int(u.Session),
		Time:    int64(u.Time.Sec),
		Addr:    u.addr().String(),
	}
}

// get byte \0 index
func getByteLen(byteArray []byte) int {
	n := bytes.IndexByte(byteArray[:], 0)
	if n == -1 {
		return 0
	}

	return n
}
