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
	"time"
)

// IstgtUctlUnxpath is the storage path for the UNIX domain socket from istgt
const (
	IstgtUctlUnxpath = "/var/run/istgt_ctl_sock"
)

func reader(r io.Reader) {
	buf := make([]byte, 1024)
	for {
		n, err := r.Read(buf[:])
		if err != nil {
			return
		}
		println("Client got:", string(buf[0:n]))
	}
}

func statusreader(r io.Reader) {
	buf := make([]byte, 1024)
	for {
		n, err := r.Read(buf[:])
		if err != nil {
			return
		}
		println("Client got:", string(buf[0:n]))
	}
}

func statuswriter(w io.Writer) {
	msg := "STATUS\n"
	_, err := w.Write([]byte(msg))
	if err != nil {
		log.Fatal("Write error:", err)
	}
	println("Client sent:", msg)
}

// Status gives the status
func Status() error {
	c, err := net.Dial("unix", IstgtUctlUnxpath)
	if err != nil {
		log.Fatal("Dial error", err)
	}
	defer c.Close()
	statuswriter(c)
	statusreader(c)
	return err
}

func refreshreader(r io.Reader) {
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

func refreshwriter(w io.Writer) {
	msg := "REFRESH\n"
	_, err := w.Write([]byte(msg))
	if err != nil {
		log.Fatal("Write error:", err)
	}
	println("Client sent:", msg)
}

// SendRefresh sends refresh command to istgt
func SendRefresh() error {
	c, err := net.Dial("unix", IstgtUctlUnxpath)
	if err != nil {
		log.Fatal("Dial error", err)
	}
	defer c.Close()
	refreshwriter(c)
	refreshreader(c)
	return err
}

func communicate() {
	c, err := net.DialTimeout("unix", IstgtUctlUnxpath, 10*1000000000)
	if err != nil {
		log.Fatal("Dial error", err)
	}
	defer c.Close()

	go reader(c)
	for {
		msg := "hi"
		_, err := c.Write([]byte(msg))
		if err != nil {
			log.Fatal("Write error:", err)
			break
		}
		println("Client sent:", msg)
		time.Sleep(1e9)
	}
}

func mmrmain() {
	c, err := net.Dial("unix", "/var/run/istgt_ctl_sock")
	if err != nil {
		log.Fatal("Dial error", err)
	}
	defer c.Close()
	writer(c)
	reader(c)
	//time.Sleep(1 * time.Second)
	//writer(c)
}

func writer(w io.Writer) {
	msg := "IOSTATS\n"
	_, err := w.Write([]byte(msg))
	if err != nil {
		log.Fatal("Write error:", err)
	}
	println("Client sent:", msg)
}
