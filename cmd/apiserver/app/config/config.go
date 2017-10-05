package config

import (
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"

	"sort"
	"strconv"
	"strings"
)

// MayaConfig is the configuration for Maya server.
type MayaConfig struct {
	// Region is the region this Maya server is supposed to deal in.
	// Defaults to global.
	Region string `mapstructure:"region"`

	// Datacenter is the datacenter this Maya server is supposed to deal in.
	// Defaults to dc1
	Datacenter string `mapstructure:"datacenter"`

	// NodeName is the name we register as. Defaults to hostname.
	NodeName string `mapstructure:"name"`

	// DataDir is the directory to store Maya server's state in
	DataDir string `mapstructure:"data_dir"`

	// LogLevel is the level of the logs to putout
	LogLevel string `mapstructure:"log_level"`

	// BindAddr is the address on which maya's services will
	// be bound. If not specified, this defaults to 127.0.0.1.
	BindAddr string `mapstructure:"bind_addr"`

	// EnableDebug is used to enable debugging HTTP endpoints
	EnableDebug bool `mapstructure:"enable_debug"`

	// Mayaserver can make use of various providers e.g. Nomad,
	// k8s etc
	ServiceProvider string `mapstructure:"service_provider"`

	// Ports is used to control the network ports we bind to.
	Ports *Ports `mapstructure:"ports"`

	// Addresses is used to override the network addresses we bind to.
	//
	// Use normalizedAddrs if you need the host+port to bind to.
	Addresses *Addresses `mapstructure:"addresses"`

	// NormalizedAddr is set to the Address+Port by normalizeAddrs()
	NormalizedAddrs *Addresses

	// AdvertiseAddrs is used to control the addresses we advertise.
	AdvertiseAddrs *AdvertiseAddrs `mapstructure:"advertise"`

	// LeaveOnInt is used to gracefully leave on the interrupt signal
	LeaveOnInt bool `mapstructure:"leave_on_interrupt"`

	// LeaveOnTerm is used to gracefully leave on the terminate signal
	LeaveOnTerm bool `mapstructure:"leave_on_terminate"`

	// EnableSyslog is used to enable sending logs to syslog
	EnableSyslog bool `mapstructure:"enable_syslog"`

	// SyslogFacility is used to control the syslog facility used.
	SyslogFacility string `mapstructure:"syslog_facility"`

	// NomadConfig is used to communicate with Nomad agent.
	//NomadConfig *nomad.Config `mapstructure:"nomad_config"`

	// Version information is set at compilation time
	Revision          string
	Version           string
	VersionPrerelease string

	// List of config files that have been loaded (in order)
	Files []string `mapstructure:"-"`

	// HTTPAPIResponseHeaders allows users to configure the Nomad http agent to
	// set arbritrary headers on API responses
	HTTPAPIResponseHeaders map[string]string `mapstructure:"http_api_response_headers"`
}

// Ports encapsulates the various ports we bind to for network services. If any
// are not specified then the defaults are used instead.
type Ports struct {
	HTTP int `mapstructure:"http"`
}

// Addresses encapsulates all of the addresses we bind to for various
// network services. Everything is optional and defaults to BindAddr.
type Addresses struct {
	HTTP string `mapstructure:"http"`
}

// AdvertiseAddrs is used to control the addresses we advertise out for
// different network services. All are optional and default to BindAddr and
// their default Port.
type AdvertiseAddrs struct {
	HTTP string `mapstructure:"http"`
}

// DefaultMayaConfig is a the baseline configuration for Maya server
func DefaultMayaConfig() *MayaConfig {
	return &MayaConfig{
		LogLevel:   "INFO",
		Region:     "global",
		Datacenter: "dc1",
		BindAddr:   "127.0.0.1",
		Ports: &Ports{
			HTTP: 5656,
		},
		Addresses:      &Addresses{},
		AdvertiseAddrs: &AdvertiseAddrs{},
		SyslogFacility: "LOCAL0",
	}
}

// Listener can be used to get a new listener using a custom bind address.
// If the bind provided address is empty, the BindAddr is used instead.
func (mc *MayaConfig) Listener(proto, addr string, port int) (net.Listener, error) {
	if addr == "" {
		addr = mc.BindAddr
	}

	// Do our own range check to avoid bugs in package net.
	//
	//   golang.org/issue/11715
	//   golang.org/issue/13447
	//
	// Both of the above bugs were fixed by golang.org/cl/12447 which will be
	// included in Go 1.6. The error returned below is the same as what Go 1.6
	// will return.
	if 0 > port || port > 65535 {
		return nil, &net.OpError{
			Op:  "listen",
			Net: proto,
			Err: &net.AddrError{Err: "invalid port", Addr: fmt.Sprint(port)},
		}
	}
	return net.Listen(proto, net.JoinHostPort(addr, strconv.Itoa(port)))
}

// Merge merges two configurations & returns a new one.
func (mc *MayaConfig) Merge(b *MayaConfig) *MayaConfig {
	result := *mc

	if b.Region != "" {
		result.Region = b.Region
	}
	if b.Datacenter != "" {
		result.Datacenter = b.Datacenter
	}
	if b.NodeName != "" {
		result.NodeName = b.NodeName
	}
	if b.DataDir != "" {
		result.DataDir = b.DataDir
	}
	if b.LogLevel != "" {
		result.LogLevel = b.LogLevel
	}
	if b.BindAddr != "" {
		result.BindAddr = b.BindAddr
	}
	if b.EnableDebug {
		result.EnableDebug = true
	}
	if b.LeaveOnInt {
		result.LeaveOnInt = true
	}
	if b.LeaveOnTerm {
		result.LeaveOnTerm = true
	}
	if b.EnableSyslog {
		result.EnableSyslog = true
	}
	if b.SyslogFacility != "" {
		result.SyslogFacility = b.SyslogFacility
	}

	// Apply the ports config
	if result.Ports == nil && b.Ports != nil {
		ports := *b.Ports
		result.Ports = &ports
	} else if b.Ports != nil {
		result.Ports = result.Ports.Merge(b.Ports)
	}

	// Apply the address config
	if result.Addresses == nil && b.Addresses != nil {
		addrs := *b.Addresses
		result.Addresses = &addrs
	} else if b.Addresses != nil {
		result.Addresses = result.Addresses.Merge(b.Addresses)
	}

	// Apply the advertise addrs config
	if result.AdvertiseAddrs == nil && b.AdvertiseAddrs != nil {
		advertise := *b.AdvertiseAddrs
		result.AdvertiseAddrs = &advertise
	} else if b.AdvertiseAddrs != nil {
		result.AdvertiseAddrs = result.AdvertiseAddrs.Merge(b.AdvertiseAddrs)
	}

	// Merge config files lists
	result.Files = append(result.Files, b.Files...)

	// Add the http API response header map values
	if result.HTTPAPIResponseHeaders == nil {
		result.HTTPAPIResponseHeaders = make(map[string]string)
	}
	for k, v := range b.HTTPAPIResponseHeaders {
		result.HTTPAPIResponseHeaders[k] = v
	}

	return &result
}

// NormalizeAddrs normalizes Addresses and AdvertiseAddrs to always be
// initialized and have sane defaults.
func (mc *MayaConfig) NormalizeAddrs() error {
	mc.Addresses.HTTP = normalizeBind(mc.Addresses.HTTP, mc.BindAddr)
	mc.NormalizedAddrs = &Addresses{
		HTTP: net.JoinHostPort(mc.Addresses.HTTP, strconv.Itoa(mc.Ports.HTTP)),
	}

	addr, err := normalizeAdvertise(mc.AdvertiseAddrs.HTTP, mc.Addresses.HTTP, mc.Ports.HTTP)
	if err != nil {
		return fmt.Errorf("Failed to parse HTTP advertise address: %v", err)
	}
	mc.AdvertiseAddrs.HTTP = addr

	return nil
}

// normalizeBind returns a normalized bind address.
//
// If addr is set it is used, if not the default bind address is used.
func normalizeBind(addr, bind string) string {
	if addr == "" {
		return bind
	}
	return addr
}

// normalizeAdvertise returns a normalized advertise address.
//
// If addr is set, it is used and the default port is appended if no port is
// set.
//
// If addr is not set and bind is a valid address, the returned string is the
// bind+port.
//
// If addr is not set and bind is not a valid advertise address, the hostname
// is resolved and returned with the port.
//
// Note - Loopback is considered a valid advertise address
func normalizeAdvertise(addr string, bind string, defport int) (string, error) {
	if addr != "" {
		// Default to using manually configured address
		_, _, err := net.SplitHostPort(addr)
		if err != nil {
			if !isMissingPort(err) {
				return "", fmt.Errorf("Error parsing advertise address %q: %v", addr, err)
			}

			// missing port, append the default
			return net.JoinHostPort(addr, strconv.Itoa(defport)), nil
		}
		return addr, nil
	}

	// Fallback to bind address first, and then try resolving the local hostname
	ips, err := net.LookupIP(bind)
	if err != nil {
		return "", fmt.Errorf("Error resolving bind address %q: %v", bind, err)
	}

	// Return the first unicast address
	for _, ip := range ips {
		if ip.IsLinkLocalUnicast() || ip.IsGlobalUnicast() {
			return net.JoinHostPort(ip.String(), strconv.Itoa(defport)), nil
		}
		if ip.IsLoopback() {
			// loopback is fine
			return net.JoinHostPort(ip.String(), strconv.Itoa(defport)), nil
		}
	}

	// As a last resort resolve the hostname and use it if it's not
	// localhost (as localhost is never a sensible default)
	host, err := os.Hostname()
	if err != nil {
		return "", fmt.Errorf("Unable to get hostname to set advertise address: %v", err)
	}

	ips, err = net.LookupIP(host)
	if err != nil {
		return "", fmt.Errorf("Error resolving hostname %q for advertise address: %v", host, err)
	}

	// Return the first unicast address
	for _, ip := range ips {
		if ip.IsLinkLocalUnicast() || ip.IsGlobalUnicast() {
			return net.JoinHostPort(ip.String(), strconv.Itoa(defport)), nil
		}
		if ip.IsLoopback() {
			// loopback is fine
			return net.JoinHostPort(ip.String(), strconv.Itoa(defport)), nil
		}
	}
	return "", fmt.Errorf("No valid advertise addresses, please set `advertise` manually")
}

// isMissingPort returns true if an error is a "missing port" error from
// net.SplitHostPort.
func isMissingPort(err error) bool {
	// matches error const in net/ipsock.go
	const missingPort = "missing port in address"
	return err != nil && strings.Contains(err.Error(), missingPort)
}

// Merge is used to merge two port configurations.
func (a *Ports) Merge(b *Ports) *Ports {
	result := *a

	if b.HTTP != 0 {
		result.HTTP = b.HTTP
	}
	return &result
}

// Merge is used to merge two address configs together.
func (a *Addresses) Merge(b *Addresses) *Addresses {
	result := *a

	if b.HTTP != "" {
		result.HTTP = b.HTTP
	}
	return &result
}

// Merge merges two advertise addrs configs together.
func (a *AdvertiseAddrs) Merge(b *AdvertiseAddrs) *AdvertiseAddrs {
	result := *a

	if b.HTTP != "" {
		result.HTTP = b.HTTP
	}
	return &result
}

// LoadMayaConfig loads the configuration at the given path, regardless if
// its a file or directory.
func LoadMayaConfig(path string) (*MayaConfig, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	if fi.IsDir() {
		return LoadMayaConfigDir(path)
	}

	cleaned := filepath.Clean(path)
	mconfig, err := ParseMayaConfigFile(cleaned)
	if err != nil {
		return nil, fmt.Errorf("Error loading %s: %s", cleaned, err)
	}

	mconfig.Files = append(mconfig.Files, cleaned)
	return mconfig, nil
}

// LoadMayaConfigDir loads all the configurations in the given directory
// in alphabetical order.
func LoadMayaConfigDir(dir string) (*MayaConfig, error) {
	f, err := os.Open(dir)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if !fi.IsDir() {
		return nil, fmt.Errorf(
			"configuration path must be a directory: %s", dir)
	}

	var files []string
	err = nil
	for err != io.EOF {
		var fis []os.FileInfo
		fis, err = f.Readdir(128)
		if err != nil && err != io.EOF {
			return nil, err
		}

		for _, fi := range fis {
			// Ignore directories
			if fi.IsDir() {
				continue
			}

			// Only care about files that are valid to load.
			name := fi.Name()
			skip := true
			if strings.HasSuffix(name, ".hcl") {
				skip = false
			} else if strings.HasSuffix(name, ".json") {
				skip = false
			}
			if skip || isTemporaryFile(name) {
				continue
			}

			path := filepath.Join(dir, name)
			files = append(files, path)
		}
	}

	// Fast-path if we have no files
	if len(files) == 0 {
		return &MayaConfig{}, nil
	}

	sort.Strings(files)

	var result *MayaConfig
	for _, f := range files {
		mconfig, err := ParseMayaConfigFile(f)
		if err != nil {
			return nil, fmt.Errorf("Error loading %s: %s", f, err)
		}
		mconfig.Files = append(mconfig.Files, f)

		if result == nil {
			result = mconfig
		} else {
			result = result.Merge(mconfig)
		}
	}

	return result, nil
}

// isTemporaryFile returns true or false depending on whether the
// provided file name is a temporary file for the following editors:
// emacs or vim.
func isTemporaryFile(name string) bool {
	return strings.HasSuffix(name, "~") || // vim
		strings.HasPrefix(name, ".#") || // emacs
		(strings.HasPrefix(name, "#") && strings.HasSuffix(name, "#")) // emacs
}
