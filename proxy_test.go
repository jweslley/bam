package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestMatchDomains(t *testing.T) {
	data := []struct {
		domain, subdomain string
	}{
		{"acme.local", "acme.local"},
		{"acme.local", "sub.acme.local"},
		{"acme.local", "sub.sub.acme.local"},
	}

	for _, d := range data {
		if !matchDomains(d.domain, d.subdomain) {
			t.Errorf("'%s' doesnt match '%s'", d.subdomain, d.domain)
		}
	}
}

func TestMatchXipDomains(t *testing.T) {
	data := []struct {
		name, host string
	}{
		{"acme", "acme.192.168.1.9.xip.io"},
		{"acme", "sub.acme.192.168.1.9.xip.io"},
		{"acme", "sub.sub.acme.192.168.1.9.xip.io"},
	}

	for _, d := range data {
		if !matchXipDomain(d.name, d.host) {
			t.Errorf("'%s' doesnt match '%s'", d.host, d.name)
		}
	}
}

func TestProxyResolve(t *testing.T) {
	servers := []Server{
		newServer("godoc", 6060),
		newServer("goapp", 8080),
		newServer("btsync", 8888),
	}
	p := NewProxy("local", newServers(servers))

	resolveCheck := func(name, host string) {
		s, ok := p.resolve(host)
		if !ok {
			t.Errorf("cant resolve %s", host)
		}

		if s.Name() != name {
			t.Errorf("wrong server. want: %s, got: %s", name, s.Name())
		}
	}

	// check equals match
	resolveCheck("godoc", "godoc.local")
	resolveCheck("goapp", "goapp.local")
	resolveCheck("btsync", "btsync.local")

	// check subdomain matches
	resolveCheck("godoc", "www.godoc.local")
	resolveCheck("goapp", "pt-br.goapp.local")
	resolveCheck("btsync", "p2p.btsync.local")

	// check xip domains
	resolveCheck("godoc", "godoc.192.168.1.11.xip.io")
	resolveCheck("goapp", "pt-br.goapp.172.20.1.200.xip.io")
	resolveCheck("btsync", "p.p.btsync.192.20.1.42.xip.io")

	unresolvedCheck := func(host string) {
		s, ok := p.resolve(host)
		if ok {
			t.Errorf("server '%s' was found for host '%s'", s.Name(), host)
		}
	}

	unresolvedCheck("p2p.local")
	unresolvedCheck("godoc.acme.local")
	unresolvedCheck("godoc.dev")
}

func TestProxy(t *testing.T) {
	servers := []Server{}
	createServer := func(name string, status int, content string) *httptest.Server {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(status)
			fmt.Fprint(w, content)
		}))
		port := getServerPort(t, s.URL)
		servers = append(servers, newServer(name, port))
		return s
	}

	const myappContent, myappStatus = "hello world", 201
	myapp := createServer("myapp", myappStatus, myappContent)
	defer myapp.Close()

	const fooContent, fooStatus = "foo bar", 404
	foo := createServer("foo", fooStatus, fooContent)
	defer foo.Close()

	proxy := httptest.NewServer(NewProxy("local", newServers(servers)))
	defer proxy.Close()

	requestCheck := func(host string, expectedStatus int, expectedContent string) {
		req, _ := http.NewRequest("GET", proxy.URL, nil)
		req.Host = host
		req.Close = true
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}

		if res.StatusCode != expectedStatus {
			t.Errorf("Status code: got %d; expected %d", res.StatusCode, expectedStatus)
		}
		bodyBytes, _ := ioutil.ReadAll(res.Body)
		gotContent := string(bodyBytes)
		if gotContent != expectedContent {
			t.Errorf("Body: got %s; expected %s", gotContent, expectedContent)
		}
	}

	tests := []struct {
		name    string
		status  int
		content string
	}{
		{"myapp.local", myappStatus, myappContent},
		{"subdomain.myapp.local", myappStatus, myappContent},
		{"pt.subdomain.myapp.local", myappStatus, myappContent},
		{"myapp.192.168.1.42.xip.io", myappStatus, myappContent},
		{"subdomain.myapp.192.168.1.42.xip.io", myappStatus, myappContent},
		{"en.subdomain.myapp.192.168.1.42.xip.io", myappStatus, myappContent},
		{"foo.local", fooStatus, fooContent},
		{"bar.foo.local", fooStatus, fooContent},
		{"foo.192.168.1.42.xip.io", fooStatus, fooContent},
		{"bar.foo.192.168.1.42.xip.io", fooStatus, fooContent},
	}

	for _, tt := range tests {
		requestCheck(tt.name, tt.status, tt.content)
	}
}

func getServerPort(t *testing.T, baseURL string) int {
	url, e := url.Parse(baseURL)
	if e != nil {
		t.Fatal(e)
	}

	port, err := AddrPort(url.Host)
	if err != nil {
		t.Fatal(err)
	}
	return port
}

type servers struct {
	server
	servers []Server
}

func (s *servers) List() []Server {
	return s.servers
}

func newServers(ss []Server) Servers {
	s := &servers{}
	s.servers = ss
	return s
}

func newServer(name string, port int) Server {
	return &server{name, port}
}
