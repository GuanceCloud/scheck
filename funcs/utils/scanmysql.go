/*
	A substantial portion of this code is based on:
	https://medium.com/@alexanderravikovich/\
	writing-mysql-proxy-in-go-for-learning-purposes-part-2-decoding-connection-phase-server-response-7091d87e877e
	https://dev.mysql.com/doc/internals/en/capability-flags.html
*/

package utils

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	lua "github.com/yuin/gopher-lua"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/funcs/impl"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/internal/global"
)

/*
PacketHeader represents packet header.
*/
type PacketHeader struct {
	Length     uint32
	SequenceID uint8
}

/*
InitialHandshakePacket represents initial handshake packet sent by MySQL Server.
*/
type InitialHandshakePacket struct {
	ProtocolVersion   uint8
	ServerVersion     []byte
	ConnectionID      uint32
	AuthPluginData    []byte
	Filler            byte
	CapabilitiesFlags CapabilityFlag
	CharacterSet      uint8
	StatusFlags       uint16
	AuthPluginDataLen uint8
	AuthPluginName    []byte
	header            *PacketHeader
}

/*
Decode decodes the first packet received from the MySQl Server
It's assumed to be a handshake packet
*/
// nolint
func (r *InitialHandshakePacket) Decode(conn net.Conn) error {
	timeoutDuration := 100 * time.Millisecond
	// net 读超时设置
	_ = conn.SetReadDeadline(time.Now().Add(timeoutDuration))

	data := make([]byte, global.KB)
	_, err := conn.Read(data)
	if err != nil {
		return err
	}

	header := &PacketHeader{}
	ln := []byte{data[0], data[1], data[2], 0x00}
	header.Length = binary.LittleEndian.Uint32(ln)
	// Single byte integer is the same in BigEndian and LittleEndian
	header.SequenceID = data[3]

	// Header Sanity check
	if header.Length >= global.KB {
		return errors.New("Header sanity check failed!")
	}

	r.header = header
	/**
	Assign payload only data to new var just for convenience
	*/
	payload := data[4 : header.Length+4]
	position := 0
	/**
	As defined in the documentation, this value is alway 10 (0x00 in hex)
	1	[0a] protocol version
	*/
	r.ProtocolVersion = payload[0]
	if r.ProtocolVersion != 0x0a {

		// This is not the best way but appears to work for this POC.
		// Hopefully this message has not been localized
		// Find index right of first terminal character (0x00)
		termIndex := bytes.IndexByte(data, byte(0x00)) + 1
		s := string(payload[termIndex:header.Length])
		if strings.Contains(s, "is not allowed to connect to this MySQL server") {
			return errors.New(s)
		}

		if r.ProtocolVersion == 0x09 {
			return errors.New("version 9 is not yet supported")
		}

		return errors.New("only version 10 is supported. Unknown procotcol version")
	}

	position++

	/*
		Extract server version, by finding the terminal character (0x00) index,
		and extracting the data in between
		string[NUL]    server version
	*/
	index := bytes.IndexByte(payload, byte(0x00))
	r.ServerVersion = payload[position:index]
	position = index + 1

	connectionID := payload[position : position+4]
	id := binary.LittleEndian.Uint32(connectionID)
	r.ConnectionID = id
	position += 4

	/*
		The auth-plugin-data is the concatenation of strings
		auth-plugin-data-part-1 and auth-plugin-data-part-2.
	*/
	r.AuthPluginData = make([]byte, 8)
	copy(r.AuthPluginData, payload[position:position+8])

	position += 8

	r.Filler = payload[position]
	if r.Filler != 0x00 {
		return errors.New("Unable to decode filler value")
	}

	position++

	capabilitiesFlags1 := payload[position : position+2]
	position += 2

	r.CharacterSet = payload[position]
	position++

	r.StatusFlags = binary.LittleEndian.Uint16(payload[position : position+2])
	position += 2

	capabilityFlags2 := payload[position : position+2]
	position += 2

	/*
		Reconstruct 32 bit integer from two 16 bit integers.
		Take low 2 bytes and high 2 bytes, ans sum it.
	*/
	capLow := binary.LittleEndian.Uint16(capabilitiesFlags1)
	capHi := binary.LittleEndian.Uint16(capabilityFlags2)
	capL := uint32(capLow) | uint32(capHi)<<16

	r.CapabilitiesFlags = CapabilityFlag(capL)

	if r.CapabilitiesFlags&clientPluginAuth != 0 {
		r.AuthPluginDataLen = payload[position]
		if r.AuthPluginDataLen == 0 {
			return errors.New("Wrong auth plugin data len")
		}
	}

	/*
		Skip reserved bytes
		string[10]     reserved (all [00])
	*/
	position += 1 + 10

	/**
	This flag tell us that the client should hash the password using algorithm described here:
	https://dev.mysql.com/doc/internals/en/secure-password-authentication.html#packet-Authentication::Native41
	*/
	if r.CapabilitiesFlags&clientSecureConn != 0 {
		/*
			The auth-plugin-data is the concatenation of strings auth-plugin-data-part-1 and auth-plugin-data-part-2.
		*/
		end := position + Max(13, int(r.AuthPluginDataLen)-8)
		r.AuthPluginData = append(r.AuthPluginData, payload[position:end]...)
		position = end
	}

	index = bytes.IndexByte(payload[position:], byte(0x00))

	/*
		Due to Bug#59453 the auth-plugin-name is missing the terminating NUL-char in versions prior to 5.5.10 and 5.6.2.
		We know the length of the payload, so if there is no NUL-char, just read all the data until the end
	*/
	if index != -1 {
		r.AuthPluginName = payload[position : position+index]
	} else {
		r.AuthPluginName = payload[position:]
	}
	return nil
}

func (r *InitialHandshakePacket) String() string {
	fields := make([]string, 0)
	fields = append(fields, fmt.Sprintf("ProtocolVersion: %d", r.ProtocolVersion),
		fmt.Sprintf("ServerVersion: %s", r.ServerVersion),
		fmt.Sprintf("ConnectionId: %d", r.ConnectionID),
		fmt.Sprintf("AuthPluginName: %s", r.AuthPluginName),
		fmt.Sprintf("StatusFlags: %d\n", r.StatusFlags))
	return strings.Join(fields, "\n")
}

type CapabilityFlag uint32

/**
Each flag is just an number, that can be represented just by having a single bit ON.
It allows us to use fast bitwise operations. Each flag is just a number with applied << operator,
that is equivalent of multiply by 2

1 = 00000001
2 = 00000010
4 = 00000100
...

To check if the flag is set, we use & operator

00000111 & 00000001 = 1 => true
00000111 & 01000000 = 0 => false
*/

func (r CapabilityFlag) Has(flag CapabilityFlag) bool {
	return r&flag != 0
}

// String :Debug Helper.
func (r CapabilityFlag) String() string {
	var names []string

	for i := uint64(1); i <= uint64(1)<<31; i <<= 1 {
		name, ok := flags[CapabilityFlag(i)]
		if ok {
			names = append(names, fmt.Sprintf("0x%08x - %032b - %s", i, i, name))
		}
	}

	return strings.Join(names, "\n")
}

const (
	clientLongPassword CapabilityFlag = 1 << iota
	clientFoundRows
	clientLongFlag
	clientConnectWithDB
	clientNoSchema
	clientCompress
	clientODBC
	clientLocalFiles
	clientIgnoreSpace
	clientProtocol41
	clientInteractive
	clientSSL
	clientIgnoreSIGPIPE
	clientTransactions
	clientReserved
	clientSecureConn
	clientMultiStatements
	clientMultiResults
	clientPSMultiResults
	clientPluginAuth
	clientConnectAttrs
	clientPluginAuthLenEncClientData
	clientCanHandleExpiredPasswords
	clientSessionTrack
	clientDeprecateEOF
)

var flags = map[CapabilityFlag]string{
	clientLongPassword:               "clientLongPassword",
	clientFoundRows:                  "clientFoundRows",
	clientLongFlag:                   "clientLongFlag",
	clientConnectWithDB:              "clientConnectWithDB",
	clientNoSchema:                   "clientNoSchema",
	clientCompress:                   "clientCompress",
	clientODBC:                       "clientODBC",
	clientLocalFiles:                 "clientLocalFiles",
	clientIgnoreSpace:                "clientIgnoreSpace",
	clientProtocol41:                 "clientProtocol41",
	clientInteractive:                "clientInteractive",
	clientSSL:                        "clientSSL",
	clientIgnoreSIGPIPE:              "clientIgnoreSIGPIPE",
	clientTransactions:               "clientTransactions",
	clientReserved:                   "clientReserved",
	clientSecureConn:                 "clientSecureConn",
	clientMultiStatements:            "clientMultiStatements",
	clientMultiResults:               "clientMultiResults",
	clientPSMultiResults:             "clientPSMultiResults",
	clientPluginAuth:                 "clientPluginAuth",
	clientConnectAttrs:               "clientConnectAttrs",
	clientPluginAuthLenEncClientData: "clientPluginAuthLenEncClientData",
	clientCanHandleExpiredPasswords:  "clientCanHandleExpiredPasswords",
	clientSessionTrack:               "clientSessionTrack",
	clientDeprecateEOF:               "clientDeprecateEOF",
}

func Max(x, y int) int {
	if x > y {
		return x
	}
	return y
}

// nolint
func MysqlPortsList(l *lua.LState) int {
	var result lua.LTable
	listenPorts := impl.GetListeningPorts()
	for i := range listenPorts {
		if listenPorts[i]["state"] == "LISTEN" {
			port := listenPorts[i]["port"]
			target := fmt.Sprintf("%s:%d", "localhost", port)

			conn, err := net.Dial("tcp", target)
			if err != nil {
				log.Debugf("Failed to connect %s\n", err.Error())
				continue
			}
			handshakePacket := &InitialHandshakePacket{}
			err = handshakePacket.Decode(conn)
			if err != nil {
				log.Debugf("Failed to decode packet: %s\n", err.Error())
				// return
				continue
			}
			var item lua.LTable

			if handshakePacket.ProtocolVersion != 0 {
				item.RawSetString("port", lua.LNumber(port.(uint16)))
				item.RawSetString("protocolversion", lua.LString(fmt.Sprintf("%d", handshakePacket.ProtocolVersion)))
				item.RawSetString("statusflags", lua.LString(fmt.Sprintf("%d", handshakePacket.StatusFlags)))
				item.RawSetString("authpluginname", lua.LString(handshakePacket.AuthPluginName))
				item.RawSetString("s️erverversion", lua.LString(handshakePacket.ServerVersion))
				item.RawSetString("state", lua.LString(fmt.Sprintf("%s", listenPorts[i]["state"])))
				item.RawSetString("cmdline", lua.LString(fmt.Sprintf("%s", listenPorts[i]["cmdline"])))
				item.RawSetString("pid", lua.LNumber(listenPorts[i]["pid"].(int)))

				result.Append(&item)
			}
		}
	}
	l.Push(&result)
	return 1
}
