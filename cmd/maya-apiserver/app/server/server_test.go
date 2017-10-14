package server

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"testing"

	"github.com/openebs/maya/cmd/maya-apiserver/app/config"
)

func getPort() int {
	addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		panic(err)
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port
}

func tmpDir(t testing.TB) string {
	dir, err := ioutil.TempDir("", "mapiserver")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	return dir
}

func makeMayaServer(t testing.TB, fnmc func(*config.MayaConfig)) (string, *MayaApiServer) {
	dir := tmpDir(t)

	// Customize the server configuration
	conf := config.DefaultMayaConfig()

	// Set the data_dir
	conf.DataDir = dir

	// Bind and set ports
	conf.BindAddr = "127.0.0.1"
	conf.Ports = &config.Ports{
		HTTP: getPort(),
	}
	conf.NodeName = fmt.Sprintf("Node %d", conf.Ports.HTTP)

	if fnmc != nil {
		fnmc(conf)
	}

	if err := conf.NormalizeAddrs(); err != nil {
		t.Fatalf("error normalizing config: %v", err)
	}

	maya, err := NewMayaApiServer(conf, os.Stderr)
	if err != nil {
		os.RemoveAll(dir)
		t.Fatalf("err: %v", err)
	}
	return dir, maya
}

func TestMayaServerConfig(t *testing.T) {
	conf := config.DefaultMayaConfig()

	conf.AdvertiseAddrs.HTTP = "10.10.11.1:4006"

	// Parses the advertise addrs correctly
	if err := conf.NormalizeAddrs(); err != nil {
		t.Fatalf("error normalizing config: %v", err)
	}

	// Assert addresses weren't changed
	if addr := conf.AdvertiseAddrs.HTTP; addr != "10.10.11.1:4006" {
		t.Fatalf("expect 10.11.11.1:4005, got: %v", addr)
	}

	// Sets up the ports properly
	conf.Addresses.HTTP = ""
	conf.Ports.HTTP = 4005

	if err := conf.NormalizeAddrs(); err != nil {
		t.Fatalf("error normalizing config: %v", err)
	}

	// Test if config prefers advertise over bind addr
	conf.BindAddr = "127.0.0.3"
	conf.Addresses.HTTP = "127.0.0.2"
	conf.AdvertiseAddrs.HTTP = "10.0.0.10"

	if err := conf.NormalizeAddrs(); err != nil {
		t.Fatalf("error normalizing config: %v", err)
	}

	if addr := conf.Addresses.HTTP; addr != "127.0.0.2" {
		t.Fatalf("expect HTTP addr 127.0.0.2, got: %s", addr)
	}

	if addr := conf.NormalizedAddrs.HTTP; addr != "127.0.0.2:4005" {
		t.Fatalf("expect 127.0.0.2:4005, got: %s", addr)
	}

	if addr := conf.AdvertiseAddrs.HTTP; addr != "10.0.0.10:4005" {
		t.Fatalf("expect 10.0.0.10:4005, got: %s", addr)
	}

	// Defaults to the global bind addr
	// when address & advertise address are blank
	conf.Addresses.HTTP = ""
	conf.AdvertiseAddrs.HTTP = ""
	conf.Ports.HTTP = 6666
	if err := conf.NormalizeAddrs(); err != nil {
		t.Fatalf("error normalizing config: %v", err)
	}
	if addr := conf.Addresses.HTTP; addr != "127.0.0.3" {
		t.Fatalf("expect 127.0.0.3, got: %s", addr)
	}
	if addr := conf.NormalizedAddrs.HTTP; addr != "127.0.0.3:6666" {
		t.Fatalf("expect 127.0.0.3:6666, got: %s", addr)
	}

}
