package utils

import (
	"fmt"
	"net"
	"testing"
)

func Test_FindSelfIp(t *testing.T)  {
	targetIp := FindInetAddress(func(ipnet *net.IPNet) {
		fmt.Printf("current find ipnet v4 info %+v\n", *ipnet)
		fmt.Printf("current find ipnet v6 info %+v\n", *ipnet)
		fmt.Println("---------------------------------")
	})


	fmt.Println(targetIp)
}