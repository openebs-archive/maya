package mapiserver

import (
	"net"
	"os"
)

func Initialize() {
	mapiaddr := os.Getenv("MAPI_ADDR")
	if mapiaddr == "" {
		mapiaddr = getDefaultAddr()
		os.Setenv("MAPI_ADDR", mapiaddr)
	}
}

func GetURL() string {
	return os.Getenv("MAPI_ADDR")
}

func GetConnectionStatus() string {
	_, err := GetStatus()
	if err != nil {
		return "not reachable"
	}
	return "running"
}

func getDefaultAddr() string {
	env := "127.0.0.1"
	host, _ := os.Hostname()
	addrs, _ := net.LookupIP(host)
	for _, addr := range addrs {
		if ipv4 := addr.To4(); ipv4 != nil {
			if ipv4.String() != "127.0.0.1" {
				env = ipv4.String()
				break
			}
		}
	}
	return "http://" + env + ":5656"
}
