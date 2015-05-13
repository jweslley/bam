package main

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/miekg/dns"
)

func TestLocalDNS(t *testing.T) {
	port, _ := FreePort()
	addrstr := fmt.Sprintf(":%d", port)
	ldns := NewLocalDNS(port, "app")
	err := ldns.Start()
	if err != nil {
		t.Fatalf("Failed to start LocalDNS at %d: %v", port, err)
	}

	if !ldns.Running() {
		t.Fatalf("LocalDNS should be started now")
	}

	<-time.After(1 * time.Second) // wait for start

	c := new(dns.Client)
	lookupCheckIPv4 := func(question string, answer net.IP) {
		m := new(dns.Msg)
		m.SetQuestion(question, dns.TypeA)
		r, _, err := c.Exchange(m, addrstr)
		if err != nil || len(r.Answer) == 0 {
			t.Fatalf("failed to exchange %s: %v", question, err)
		}
		a := r.Answer[0].(*dns.A)
		if a.A.String() != answer.String() {
			t.Errorf("Wrong answer: got %d; expected %d", a.A, answer)
		}
	}

	lookupCheckIPv6 := func(question string, answer net.IP) {
		m := new(dns.Msg)
		m.SetQuestion(question, dns.TypeAAAA)
		r, _, err := c.Exchange(m, addrstr)
		if err != nil || len(r.Answer) == 0 {
			t.Fatalf("failed to exchange %s: %v", question, err)
		}
		a := r.Answer[0].(*dns.AAAA)
		if a.AAAA.String() != answer.String() {
			t.Errorf("Wrong answer: got %d; expected %d", a.AAAA, answer)
		}
	}

	questions := []string{"myapp.app.", "subdomain.myapp.app.", "pt.subdomain.myapp.app."}

	for _, q := range questions {
		lookupCheckIPv4(q, ip4loopback)
		lookupCheckIPv6(q, net.IPv6loopback)
	}

	ldns.Stop()
	if ldns.Running() {
		t.Fatalf("LocalDNS should be stopped now")
	}
}
