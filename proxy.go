package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"regexp"
	"strings"
)

var xipio = regexp.MustCompile("^(.*?)\\.?\\d+\\.\\d+\\.\\d+\\.\\d+\\.xip\\.io")

// Server is the interface the wraps the Name and Port methods.
type Server interface {
	Name() string
	Port() int
}

// Servers provides a list of available servers.
type Servers interface {
	Server
	List() []Server
}

// Proxy is a ReverseProxy that takes an incoming request and
// sends it to one of the known servers based on server's name,
// after proxying the response back to the client.
type Proxy struct {
	httputil.ReverseProxy
	servers Servers
	tld     string
}

// NewProxy returns a new Proxy.
func NewProxy(tld string, s Servers) *Proxy {
	p := &Proxy{tld: tld, servers: s}
	p.Director = func(req *http.Request) {
		req.URL.Scheme = "http"
		server, found := p.resolve(req.Host)
		if found {
			req.URL.Host = fmt.Sprint("localhost:", server.Port())
		} else {
			log.Printf("WARN No server found for host %s\n", req.Host)
			req.URL.Host = fmt.Sprint("localhost:", s.Port())
			req.URL.Path = "/not-found"
			req.URL.RawQuery = fmt.Sprintf("app=%s", strings.TrimSuffix(req.Host, tld))
		}
	}
	return p
}

// Resolve finds a Server matching the given host.
// Return false, if any Server matches host.
func (p *Proxy) resolve(host string) (Server, bool) {
	for _, s := range p.servers.List() {
		if p.match(s, host) {
			return s, true
		}
	}
	return nil, false
}

// match checks is host matches server's name.
func (p *Proxy) match(s Server, host string) bool {
	return matchDomains(s.Name()+"."+p.tld, host) || matchXipDomain(s.Name(), host)
}

// matchDomains checks whether 'a' domain is equals to 'b' domain, or,
// 'b' domain is a subdomain of 'a' domain.
func matchDomains(a, b string) bool {
	return a == b || strings.HasSuffix(b, "."+a)
}

// matchXipDomain checks whether host is a xip domain of name.
func matchXipDomain(name, host string) bool {
	subdomain := xipio.ReplaceAllString(host, "$1")
	return matchDomains(name, subdomain)
}

type server struct {
	name string
	port int
}

func (s *server) Name() string {
	return s.name
}

func (s *server) Port() int {
	return s.port
}

func (s *server) String() string {
	return fmt.Sprintf("%s:%d", s.name, s.port)
}
