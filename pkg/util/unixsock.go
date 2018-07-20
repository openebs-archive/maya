package util

import (
	"io"
	"net"
	"strings"

	"github.com/golang/glog"
)

// IstgtUctlUnxpath is the storage path for the UNIX domain socket from istgt
const (
	IstgtUctlUnxpath = "/var/run/istgt_ctl_sock"
	EndOfLine        = "\r\n"
	IstgtHeader      = "iSCSI Target Controller version"
)

//Reader reads the response from unix domain socket
func Reader(r io.Reader, cmd string) []string {
	resp := []string{}
	//collect bytes into fulllines buffer till the end of line character is reached
	fulllines := []byte{}
	for {
		buf := make([]byte, 1024)
		n, err := r.Read(buf[:])
		if n > 0 {
			glog.Infof("Client got:", string(buf[0:n]))
			fulllines = append(fulllines, buf[0:n]...)
			if strings.HasSuffix(string(fulllines), EndOfLine) {
				lines := strings.Split(string(fulllines), EndOfLine)
				for _, line := range lines {
					if len(line) != 0 {
						glog.Infof("Appending line to resp : ", line)
						resp = append(resp, line+EndOfLine)
					}
				}
				//clear the fulllines buffer once the response lines are appended to the response
				fulllines = nil
			}

			if !strings.HasPrefix(resp[len(resp)-1], IstgtHeader) &&
				!strings.HasPrefix(resp[len(resp)-1], cmd) {
				glog.Infof("Breaking out of loop for line :", resp[len(resp)-1])
				break
			}
		}
		if err != nil {
			glog.Errorf("Read error : %v", err)
			break
		}
		buf = nil
	}
	glog.Infof("response : %v", resp)
	return resp
}

//Writer writes a command to unix domain socket
func Writer(w io.Writer, msg string) error {
	_, err := w.Write([]byte(msg))
	if err != nil {
		glog.Fatal("Write error: ", err)
	} else {
		glog.Infof("Client sent:", msg)
	}
	return err
}

//UnixSock operates on unix domain sockets
type UnixSock interface {
	SendCommand(cmd string) ([]string, error)
}

//RealUnixSock is used for sending data through real unix domain sockets
type RealUnixSock struct{}

//SendCommand for the real unix sock for the actual program,
func (r RealUnixSock) SendCommand(cmd string) ([]string, error) {
	c, err := net.Dial("unix", IstgtUctlUnxpath)
	if err != nil {
		glog.Fatal("Dial error", err)
	}
	err = Writer(c, cmd+EndOfLine)
	if err != nil {
		c.Close()
		return nil, err
	}
	resp := Reader(c, cmd)
	c.Close()
	return resp, err
}

//TestUnixSock is used as a dummy UnixSock
type TestUnixSock struct{}

//SendCommand for the real unix sock for the actual program,
func (r TestUnixSock) SendCommand(cmd string) ([]string, error) {
	return nil, nil
}
