package main

import (
	"net"
	"strconv"
)

// AddrPort returns the port from a network end point address.
func AddrPort(addr string) (int, error) {
	_, portStr, e := net.SplitHostPort(addr)
	if e != nil {
		return 0, e
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return 0, err
	}
	return port, nil
}

// FreePort returns an unused port.
func FreePort() (int, error) {
	l, err := NewLocalListener()
	if err != nil {
		return -1, err
	}
	defer l.Close()

	port, err := AddrPort(l.Addr().String())
	if err != nil {
		return -1, err
	}

	return port, nil
}

func NewLocalListener() (net.Listener, error) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		if l, err = net.Listen("tcp6", "[::1]:0"); err != nil {
			return nil, err
		}
	}
	return l, nil
}
