package main

import (
	"fmt"
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
	Get(string) (Server, bool)
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
			req.URL.Host = fmt.Sprint("localhost:", s.Port())
			req.URL.Path = fmt.Sprintf("/apps/%s", p.serverNameFromHost(req.Host))
		}
	}
	return p
}

func (p *Proxy) resolve(host string) (Server, bool) {
	name := p.serverNameFromHost(host)
	return p.servers.Get(name)
}

func (p *Proxy) serverNameFromHost(host string) string {
	var prefix string
	if xipio.MatchString(host) {
		prefix = xipio.ReplaceAllString(host, "$1")
	} else {
		prefix = strings.TrimSuffix(host, "."+p.tld)
	}
	t := strings.Split(prefix, ".")
	return t[len(t)-1]
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
