package network

import (
	"net"
	"strconv"
)

var myAddress = ""

// Ask the kernel for a free open port that is ready to use
func GetPort() int {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
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

func GetHostIp() string {
	myAddress := "127.0.0.1"
	addrs, _ := net.InterfaceAddrs()
	for _, address := range addrs {

		// check the address type and if it is not a loopback then display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				myAddress = ipnet.IP.String()
			}

		}
	}
	return myAddress
}

func GenerateAddress(ip string, port string) string {
	myIp := ip
	myPort := port

	if myIp == "" {
		myIp = GetHostIp()
	}
	if myPort == "" {
		p := GetPort()
		myPort = ":" + strconv.Itoa(p)
	}

	myAddress = "http://" + myIp + myPort

	return myAddress
}

func GetMyAddress() string {
	if myAddress != "" {
		return myAddress
	}

	return GenerateAddress("", "")
}
