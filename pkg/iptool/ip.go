package iptool

import (
	"net"
	"strings"
	externalip "github.com/glendc/go-external-ip"
)

// Ranges of addresses allocated by IANA for private internets, as per RFC1918 and RFC3927
// for Link-local addresses.
var PrivateNetworks = []string{
	"10.0.0.0/8",
	"172.16.0.0/12",
	"169.254.0.0/16",
	"192.168.0.0/16",
	"fe80::/10",
	"fd00::/8",
}

var privateNets []*net.IPNet

func init() {
	for _, network := range PrivateNetworks {
		_, ipnet, err := net.ParseCIDR(network)
		if err != nil {
			panic(err)
		}
		privateNets = append(privateNets, ipnet)
	}
}

// GetHostIPs returns a list of IP addresses of all host's interfaces.
func GetHostIPs() ([]net.IP, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	var ips []net.IP
	for _, iface := range ifaces {
		if strings.HasPrefix(iface.Name, "docker") {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok {
				ips = append(ips, ipnet.IP)
			}
		}
	}

	return ips, nil
}

// GetPrivateHostIPs returns a list of host's private IP addresses.
func GetPrivateHostIPs() ([]net.IP, error) {
	ips, err := GetHostIPs()
	if err != nil {
		return nil, err
	}

	var privateIPs []net.IP
	for _, ip := range ips {
		// skip loopback, non-IPv4 and non-private addresses
		if ip.IsLoopback() || ip.To4() == nil || !IsPrivate(ip) {
			continue
		}
		privateIPs = append(privateIPs, ip)
	}

	return privateIPs, nil
}

// IsPrivate determines whether a passed IP address is from one of private blocks or not.
func IsPrivate(ip net.IP) bool {
	for _, ipnet := range privateNets {
		if ipnet.Contains(ip) {
			return true
		}
	}
	return false
}


func GetExternalIP() (eip string, err error) {
	// Create the default consensus,
	// using the default configuration and no logger.
	consensus := externalip.DefaultConsensus(nil, nil)
	consensus.AddVoter(externalip.NewHTTPSource("http://ip.cip.cc"), 1)
	// Get your IP,
	// which is never <nil> when err is <nil>.
	ip, err := consensus.ExternalIP()
	if err != nil {
		return "", err
	}
	return ip.String(), nil
}