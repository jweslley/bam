package main

import (
	"fmt"
	"net"
	"strings"

	"github.com/miekg/dns"
)

const dnsTtl = 30

var ip4loopback = net.IPv4(127, 0, 0, 1)

type LocalDNS struct {
	port int
	// qualified tld. must be in format '.tld.'
	qtld string
	s    *dns.Server
}

func (l *LocalDNS) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	if r.Opcode != dns.OpcodeQuery || len(r.Question) == 0 {
		dns.HandleFailed(w, r)
		return
	}

	q := r.Question[0]
	if !(q.Qclass == dns.ClassINET && (q.Qtype == dns.TypeA || q.Qtype == dns.TypeAAAA)) {
		dns.HandleFailed(w, r)
		return
	}

	domain := dns.Fqdn(q.Name)
	if !strings.HasSuffix(domain, l.qtld) {
		m := new(dns.Msg)
		m.SetRcode(r, dns.RcodeNameError)
		w.WriteMsg(m)
		return
	}

	m := new(dns.Msg)
	m.SetReply(r)

	rr := dns.RR_Header{Name: q.Name, Class: dns.ClassINET, Rrtype: q.Qtype, Ttl: dnsTtl}
	if q.Qtype == dns.TypeA {
		m.Answer = append(m.Answer, &dns.A{rr, ip4loopback})
	} else {
		m.Answer = append(m.Answer, &dns.AAAA{rr, net.IPv6loopback})
	}

	w.WriteMsg(m)
}

func (l *LocalDNS) Start() error {
	if l.Running() {
		return errAlreadyStarted
	}

	l.s = &dns.Server{Addr: fmt.Sprintf(":%d", l.port), Net: "udp", Handler: l}
	go func() {
		l.s.ListenAndServe()
		l.s = nil
	}()
	return nil
}

func (l *LocalDNS) Stop() error {
	if !l.Running() {
		return errNotStarted
	}

	err := l.s.Shutdown()
	l.s = nil
	return err
}

func (l *LocalDNS) Running() bool {
	return l.s != nil
}

func NewLocalDNS(port int, tld string) *LocalDNS {
	return &LocalDNS{port: port, qtld: fmt.Sprintf(".%s.", tld)}
}
