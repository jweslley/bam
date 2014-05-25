package main

import (
	"net"
	"strconv"
	"strings"
)

// AddrPort returns the port from a network end point address.
func AddrPort(addr string) (int, error) {
	s := strings.SplitN(addr, ":", 2)
	port, err := strconv.Atoi(s[1])
	if err != nil {
		return -1, err
	}
	return port, nil
}

// FreePort returns an unused port.
func FreePort() (int, error) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return -1, err
	}

	port, err := AddrPort(l.Addr().String())
	if err != nil {
		return -1, err
	}

	l.Close()
	return port, nil
}
