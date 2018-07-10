package util

import (
	"io"
	"log"
	"net"
	"strings"
)

// IstgtUctlUnxpath is the storage path for the UNIX domain socket from istgt
const (
	IstgtUctlUnxpath = "/var/run/istgt_ctl_sock"
	EndOfStream      = "\r\n"
	Prefix           = "iSCSI Target Controller version"
)

//Reader reads the response from unix domain socket
func Reader(r io.Reader) []byte {
	resp := make([]byte, 4*1024)
	buf := make([]byte, 1024)
	for {
		n, err := r.Read(buf[:])
		if err != nil {
			println("return resp : ", string(resp))
			return resp
		}
		println("Client got:", string(buf[0:n]))
		if strings.HasPrefix(string(buf[0:n]), Prefix) {
			continue
		}
		resp = append(resp, buf[0:n]...)
		if strings.HasSuffix(string(buf[0:n]), EndOfStream) {
			println("End of stream. Return resp : ", string(resp))
			return resp
		}
	}
}

//Writer writes a command to unix domain socket
func Writer(w io.Writer, msg string) {
	_, err := w.Write([]byte(msg))
	if err != nil {
		log.Fatal("Write error:", err)
	}
	println("Client sent:", msg)
}

//UnixSock operates on unix domain sockets
type UnixSock interface {
	SendCommand(cmd string) ([]byte, error)
}

//RealUnixSock is used for sending data through real unix domain sockets
type RealUnixSock struct{}

//SendCommand for the real unix sock for the actual program,
func (r RealUnixSock) SendCommand(cmd string) ([]byte, error) {
	c, err := net.Dial("unix", IstgtUctlUnxpath)
	if err != nil {
		log.Fatal("Dial error", err)
	}
	defer c.Close()
	Writer(c, cmd)
	resp := Reader(c)
	return resp, err
}
