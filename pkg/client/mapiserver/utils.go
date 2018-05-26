package mapiserver

import (
	"net"
	"os"
)

var apiserverFlagPresent bool = false
var apiserverFlagValue string = ""

func Initialize() {
	mapiaddr := os.Getenv("MAPI_ADDR")
	if mapiaddr == "" {
		mapiaddr = getDefaultAddr()
		os.Setenv("MAPI_ADDR", mapiaddr)
	}
}

// Function to set value for variable apiserverFlagValue if apiserver flag is present
func SetFlag(value string) {
	apiserverFlagPresent = true
	apiserverFlagValue = value
}

func GetURL() string {
	//If apiserver flag is present get maya server ip from apiserver flag
	if apiserverFlagPresent {
		return apiserverFlagValue
	}
	// If flag is not presetn get maya server ip from environment variable
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
