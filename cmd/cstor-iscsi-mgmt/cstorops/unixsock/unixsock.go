/*
Copyright 2018 The OpenEBS Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package unixsock

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

//SendCommand sends the command to istgt
func SendCommand(cmd string) error {
	c, err := net.Dial("unix", IstgtUctlUnxpath)
	if err != nil {
		log.Fatal("Dial error", err)
	}
	defer c.Close()
	Writer(c, cmd)
	Reader(c)
	return err
}
