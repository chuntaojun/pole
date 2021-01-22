// Copyright (c) 2020, pole-group. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package utils

import "net"

func FindSelfIP() string {
	return FindInetAddress(nil)
}

func FindInetAddress(f func(ipnet *net.IPNet)) string {
	if ip := GetStringFromEnv("conf.core.inet.address"); ip != "" {
		return ip
	}
	
	openIPV6 := GetBoolFromEnvOptional("conf.core.inet.open-ipv6", false)
	
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		panic(err)
	}
	
	var localhost net.IP
	var expectIP net.IP
	
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok {
			
			if f != nil {
				f(ipnet)
			}
			
			// IsUnspecified reports whether ip is an unspecified address, either
			// the IPv4 address "0.0.0.0" or the IPv6 address "::".
			if ipnet.IP.IsUnspecified() {
				continue
			}
			
			// IsLoopback reports whether ip is a loopback address.
			if ipnet.IP.IsLoopback() {
				localhost = ipnet.IP
				continue
			}
			
			if expectIP == nil {
				if openIPV6 {
					expectIP = ipnet.IP.To16()
				} else {
					expectIP = ipnet.IP.To4()
				}
			}
		}
	}
	
	if expectIP == nil {
		expectIP = localhost
	}
	
	return expectIP.String()
	
}
