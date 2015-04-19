package main

import "testing"

func TestAddrPort(t *testing.T) {
	tests := []struct {
		addr string
		port int
	}{
		{"127.0.0.1:9000", 9000},
		{"127.0.0.1:80", 80},
		{"192.168.20.8:8080", 8080},
	}

	for _, test := range tests {
		port, _ := AddrPort(test.addr)
		if port != test.port {
			t.Errorf("got %d; expected %d", port, test.port)
		}
	}
}
