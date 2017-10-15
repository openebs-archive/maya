package config

import (
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

var (
	// trueValue/falseValue are used to get a pointer to a boolean
	trueValue  = true
	falseValue = false
)

func TestMayaConfig_Merge(t *testing.T) {
	c1 := &MayaConfig{
		Region:         "global",
		Datacenter:     "dc1",
		NodeName:       "node1",
		DataDir:        "/tmp/dir1",
		LogLevel:       "INFO",
		EnableDebug:    false,
		LeaveOnInt:     false,
		LeaveOnTerm:    false,
		EnableSyslog:   false,
		SyslogFacility: "local0.info",
		BindAddr:       "127.0.0.1",
		Ports: &Ports{
			HTTP: 4646,
		},
		Addresses: &Addresses{
			HTTP: "127.0.0.1",
		},
		AdvertiseAddrs: &AdvertiseAddrs{},
		HTTPAPIResponseHeaders: map[string]string{
			"Access-Control-Allow-Origin": "*",
		},
	}

	c2 := &MayaConfig{
		Region:         "region2",
		Datacenter:     "dc2",
		NodeName:       "node2",
		DataDir:        "/tmp/dir2",
		LogLevel:       "DEBUG",
		EnableDebug:    true,
		LeaveOnInt:     true,
		LeaveOnTerm:    true,
		EnableSyslog:   true,
		SyslogFacility: "local0.debug",
		BindAddr:       "127.0.0.2",
		Ports: &Ports{
			HTTP: 20000,
		},
		Addresses: &Addresses{
			HTTP: "127.0.0.2",
		},
		AdvertiseAddrs: &AdvertiseAddrs{},
		HTTPAPIResponseHeaders: map[string]string{
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Methods": "GET, POST, OPTIONS",
		},
	}

	result := c1.Merge(c2)
	if !reflect.DeepEqual(result, c2) {
		t.Fatalf("bad:\n%#v\n%#v", result, c2)
	}
}

func TestConfig_ParseMayaConfigFile(t *testing.T) {
	// Fails if the file doesn't exist
	if _, err := ParseMayaConfigFile("/unicorns/leprechauns"); err == nil {
		t.Fatalf("expected error, got nothing")
	}

	fh, err := ioutil.TempFile("", "nomad")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.RemoveAll(fh.Name())

	// Invalid content returns error
	if _, err := fh.WriteString("nope;!!!"); err != nil {
		t.Fatalf("err: %s", err)
	}
	if _, err := ParseMayaConfigFile(fh.Name()); err == nil {
		t.Fatalf("expected load error, got nothing")
	}

	// Valid content parses successfully
	if err := fh.Truncate(0); err != nil {
		t.Fatalf("err: %s", err)
	}
	if _, err := fh.Seek(0, 0); err != nil {
		t.Fatalf("err: %s", err)
	}
	if _, err := fh.WriteString(`{"region":"west"}`); err != nil {
		t.Fatalf("err: %s", err)
	}

	config, err := ParseMayaConfigFile(fh.Name())
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if config.Region != "west" {
		t.Fatalf("bad region: %q", config.Region)
	}
}

func TestConfig_LoadMayaConfigDir(t *testing.T) {
	// Fails if the dir doesn't exist.
	if _, err := LoadMayaConfigDir("/unicorns/leprechauns"); err == nil {
		t.Fatalf("expected error, got nothing")
	}

	dir, err := ioutil.TempDir("", "maya")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.RemoveAll(dir)

	// Returns empty config on empty dir
	config, err := LoadMayaConfig(dir)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if config == nil {
		t.Fatalf("should not be nil")
	}

	file1 := filepath.Join(dir, "conf1.hcl")
	err = ioutil.WriteFile(file1, []byte(`{"region":"west"}`), 0600)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	file2 := filepath.Join(dir, "conf2.hcl")
	err = ioutil.WriteFile(file2, []byte(`{"datacenter":"sfo"}`), 0600)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	file3 := filepath.Join(dir, "conf3.hcl")
	err = ioutil.WriteFile(file3, []byte(`nope;!!!`), 0600)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// Fails if we have a bad config file
	if _, err := LoadMayaConfigDir(dir); err == nil {
		t.Fatalf("expected load error, got nothing")
	}

	if err := os.Remove(file3); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Works if configs are valid
	config, err = LoadMayaConfigDir(dir)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if config.Region != "west" || config.Datacenter != "sfo" {
		t.Fatalf("bad: %#v", config)
	}
}

func TestConfig_LoadMayaConfig(t *testing.T) {
	// Fails if the target doesn't exist
	if _, err := LoadMayaConfig("/unicorns/leprechauns"); err == nil {
		t.Fatalf("expected error, got nothing")
	}

	fh, err := ioutil.TempFile("", "maya")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(fh.Name())

	if _, err := fh.WriteString(`{"region":"west"}`); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Works on a config file
	config, err := LoadMayaConfig(fh.Name())
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if config.Region != "west" {
		t.Fatalf("bad: %#v", config)
	}

	expectedConfigFiles := []string{fh.Name()}
	if !reflect.DeepEqual(config.Files, expectedConfigFiles) {
		t.Errorf("Loaded configs don't match\nExpected\n%+vGot\n%+v\n",
			expectedConfigFiles, config.Files)
	}

	dir, err := ioutil.TempDir("", "nomad")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.RemoveAll(dir)

	file1 := filepath.Join(dir, "config1.hcl")
	err = ioutil.WriteFile(file1, []byte(`{"datacenter":"sfo"}`), 0600)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// Works on config dir
	config, err = LoadMayaConfig(dir)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if config.Datacenter != "sfo" {
		t.Fatalf("bad: %#v", config)
	}

	expectedConfigFiles = []string{file1}
	if !reflect.DeepEqual(config.Files, expectedConfigFiles) {
		t.Errorf("Loaded configs don't match\nExpected\n%+vGot\n%+v\n",
			expectedConfigFiles, config.Files)
	}
}

func TestConfig_LoadMayaConfigsFileOrder(t *testing.T) {
	config1, err := LoadMayaConfigDir("../mockit/etcmayaserver")
	if err != nil {
		t.Fatalf("Failed to load config: %s", err)
	}

	config2, err := LoadMayaConfig("../mockit/partial_mayaserver_config")
	if err != nil {
		t.Fatalf("Failed to load config: %s", err)
	}

	expected := []string{
		// filepath.FromSlash changes these to backslash \ on Windows
		filepath.FromSlash("../mockit/etcmayaserver/common.hcl"),
		filepath.FromSlash("../mockit/etcmayaserver/server.json"),
		filepath.FromSlash("../mockit/partial_mayaserver_config"),
	}

	config := config1.Merge(config2)

	if !reflect.DeepEqual(config.Files, expected) {
		t.Errorf("Loaded configs don't match\nwant: %+v\n got: %+v\n",
			expected, config.Files)
	}
}

func TestConfig_Listener(t *testing.T) {
	config := DefaultMayaConfig()

	// Fails on invalid input
	if ln, err := config.Listener("tcp", "nope", 8080); err == nil {
		ln.Close()
		t.Fatalf("expected addr error")
	}
	if ln, err := config.Listener("nope", "127.0.0.1", 8080); err == nil {
		ln.Close()
		t.Fatalf("expected protocol err")
	}
	if ln, err := config.Listener("tcp", "127.0.0.1", -1); err == nil {
		ln.Close()
		t.Fatalf("expected port error")
	}

	// Works with valid inputs
	ln, err := config.Listener("tcp", "127.0.0.1", 24000)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	ln.Close()

	if net := ln.Addr().Network(); net != "tcp" {
		t.Fatalf("expected tcp, got: %q", net)
	}
	if addr := ln.Addr().String(); addr != "127.0.0.1:24000" {
		t.Fatalf("expected 127.0.0.1:4646, got: %q", addr)
	}

	// Falls back to default bind address if non provided
	config.BindAddr = "0.0.0.0"
	ln, err = config.Listener("tcp4", "", 24000)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	ln.Close()

	if addr := ln.Addr().String(); addr != "0.0.0.0:24000" {
		t.Fatalf("expected 0.0.0.0:24000, got: %q", addr)
	}
}

func TestIsMissingPort(t *testing.T) {
	_, _, err := net.SplitHostPort("localhost")

	// The port should be missing
	if missing := isMissingPort(err); !missing {
		t.Errorf("expected missing port error, but got %v", err)
	}

	// The port should not be missing
	_, _, err = net.SplitHostPort("localhost:9000")
	if missing := isMissingPort(err); missing {
		t.Errorf("expected no error, but got %v", err)
	}
}
