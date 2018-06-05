package util

import (
	"io"
	"log"
	"net"
)

// IstgtUctlUnxpath is the storage path for the UNIX domain socket from istgt
const (
	IstgtUctlUnxpath = "/var/run/istgt_ctl_sock"
)

//Reader reads the response from unix domain socket
func Reader(r io.Reader) {
	buf := make([]byte, 1024)
	for {
		n, err := r.Read(buf[:])
		if err != nil {
			return
		}
		println("Client got:", string(buf[0:n]))
		// OK
	}
}

//Writer writes a command to unix domain socket
func Writer(w io.Writer, msg string) {
	// msg := "REFRESH\n"
	_, err := w.Write([]byte(msg))
	if err != nil {
		log.Fatal("Write error:", err)
	}
	println("Client sent:", msg)
}

//UnixSock operates on unix domain sockets
type UnixSock interface {
	SendCommand(cmd string) error
}

//RealUnixSock is used for sending data through real unix domain sockets
type RealUnixSock struct{}

//SendCommand for the real unix sock for the actual program,
func (r RealUnixSock) SendCommand(cmd string) error {
	c, err := net.Dial("unix", IstgtUctlUnxpath)
	if err != nil {
		log.Fatal("Dial error", err)
	}
	defer c.Close()
	Writer(c, cmd)
	Reader(c)
	return err
}
