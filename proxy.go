package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"regexp"
	"strings"
)

var xipio = regexp.MustCompile("^(.*?)\\.?\\d+\\.\\d+\\.\\d+\\.\\d+\\.xip\\.io")

// AppCenter provides a registry of available apps.
type AppCenter interface {
	// Port which the AppCenter is binded to.
	Port() int

	// Get searchs for an application by name.
	Get(string) (App, bool)
}

// Proxy is a ReverseProxy that takes an incoming request and
// sends it to one of the known servers based on app's name,
// after proxying the response back to the client.
type Proxy struct {
	httputil.ReverseProxy
	ac  AppCenter
	tld string
}

func NewProxy(ac AppCenter, tld string) *Proxy {
	p := &Proxy{ac: ac, tld: tld}
	p.Director = func(req *http.Request) {
		req.URL.Scheme = "http"
		app, found := p.resolve(req.Host)
		if found && app.Running() {
			req.URL.Host = fmt.Sprint("localhost:", app.Port())
		} else {
			req.URL.Host = fmt.Sprint("localhost:", ac.Port())
			req.URL.Path = fmt.Sprintf("/apps/%s", p.appNameFromHost(req.Host))
		}
	}
	return p
}

func (p *Proxy) resolve(host string) (App, bool) {
	name := p.appNameFromHost(host)
	return p.ac.Get(name)
}

func (p *Proxy) appNameFromHost(host string) string {
	var prefix string
	if xipio.MatchString(host) {
		prefix = xipio.ReplaceAllString(host, "$1")
	} else {
		prefix = strings.TrimSuffix(host, "."+p.tld)
	}
	t := strings.Split(prefix, ".")
	return t[len(t)-1]
}
