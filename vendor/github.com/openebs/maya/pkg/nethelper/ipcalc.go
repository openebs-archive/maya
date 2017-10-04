// This is an adaptation of below gist:
//
//  https://gist.github.com/kotakanbe/d3059af990252ba89a82
package nethelper

import (
	"fmt"
	"net"
	"os/exec"
	"strings"
)

// IsCIDR validates if passed argument i.e. cidr is actually
// in CIDR format
func IsCIDR(cidr string) bool {
	// This handles validation aspects
	_, _, err := net.ParseCIDR(cidr)
	if err != nil {
		return false
	}

	return true
}

// CIDRSubnet will return the IPMask in decimal format
// e.g. 192.0.2.1/24 will return ("24", nil)
func CIDRSubnet(cidr string) (string, error) {

	// This handles validation aspects
	_, _, err := net.ParseCIDR(cidr)
	if err != nil {
		return "", err
	}

	return strings.Split(cidr, "/")[1], nil
}

// IPs accepts a network address in CIDR notation and returns a
// list of IP address.
//
// NOTE:
//    The returned list removes the network & broadcast addresses
func IPs(cidr string) ([]string, error) {

	// Get the IP address & the Network CIDR address
	// e.g. ParseCIDR("192.0.2.1/24") returns the
	// IP address 198.0.2.1 and the network 198.0.2.0/24
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}

	var ips []string
	// Start from network address
	// Keep selecting the IP till the network has the IP
	// http://play.golang.org/p/m8TNTtygK0
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
		ips = append(ips, ip.String())
	}

	// remove network address and broadcast address
	return ips[1 : len(ips)-1], nil
}

//  http://play.golang.org/p/m8TNTtygK0
func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

type Pong struct {
	Ip    string
	Alive bool
}

// ping does a network ping of the ip that it recives in its channel
// i.e. pingChannel. The result of this ping is built as a Pong structure &
// pushed into another channel i.e. pongChannel
//
// NOTE:
//  This is expected to be spawned via concurrent green threads to fasten
// the process of this ping logic execution.
func ping(pingChan <-chan string, pongChan chan<- Pong) {

	for ip := range pingChan {
		_, err := exec.Command("ping", "-c1", "-t1", ip).Output()

		var alive bool
		if err != nil {
			// Alternatively this IP addr is available
			alive = false
		} else {
			alive = true
		}

		pongChan <- Pong{Ip: ip, Alive: alive}
	}
}

// filterAlives iterates through the pong channel, filters the IP addresses
// that are alive/active & pushes them into done channel.
//
//    The parameter `maxAttempts` indicates the max number of attempts to
// receive value from pong channel.
func filterAlives(maxAttempts int, pongChan <-chan Pong, doneChan chan<- []string) {
	var alives []string

	for i := 0; i < maxAttempts; i++ {
		pong := <-pongChan
		if pong.Alive {
			alives = append(alives, pong.Ip)
		}
	}

	doneChan <- alives
}

// filterAvails iterates through the pong channel, filters the IP addresses
// that are available & pushes them into done channel.
//
//    The parameter `maxAttempts` indicates the max number of attempts to
// receive value from pong channel.
//
//    The parameter `reqdAttempts` indicates the required number of available
// IP addresses.
func filterAvails(maxAttempts int, reqdAttempts int, pongChan <-chan Pong, doneChan chan<- []string) {
	var avails []string

	for i := 0; i < maxAttempts; i++ {
		pong := <-pongChan
		if !pong.Alive {
			avails = append(avails, pong.Ip)
		}

		if len(avails) >= reqdAttempts {
			break
		}
	}

	doneChan <- avails
}

// GetAvailableIPs provides IPs that are available in the passed cidr
// network address range.
//
// The parameter `reqdCount` specifies the number of IP addresses to
// be returned.
//
// TODO IMPORTANT
// Check if these channels leak ? Do we need to close them ?
func GetAvailableIPs(cidr string, reqdCount int) ([]string, error) {

	if reqdCount == 0 {
		return nil, nil
	}

	// Get all possible IPs from the given ip address in cidr notation
	allIPs, err := IPs(cidr)
	if err != nil {
		return nil, err
	}

	concurrentMax := 100

	// Create a ping channel
	pingChan := make(chan string, concurrentMax)

	// Create a pong channel that will contain the ping result
	// along with the corresponding ip address
	pongChan := make(chan Pong, len(allIPs))

	// Create a done channel that will contain the list of available
	// ip addresses
	doneChan := make(chan []string)

	// Make available the ping logic in `concurrentMax` no of
	// threads. i.e. start the execution of ping function as an
	// independent concurrent thread of control, or goroutine,
	// within the same address space.
	for i := 0; i < concurrentMax; i++ {
		go ping(pingChan, pongChan)
	}

	// goroutine to filter the available IPs
	go filterAvails(len(allIPs), reqdCount, pongChan, doneChan)

	for _, ip := range allIPs {
		// Push the ip to ping channel, that will inturn make this ip
		// available to ping processing logic & will subsequently be
		// available in the drain i.e. doneChan
		pingChan <- ip
	}

	avails := <-doneChan

	if len(avails) < reqdCount {
		return nil, fmt.Errorf("Failed to fetch IPs. Required: '%d', Got: '%d'", reqdCount, len(avails))
	}

	return avails, nil
}
