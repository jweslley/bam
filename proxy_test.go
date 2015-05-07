package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestProxyResolve(t *testing.T) {
	apps := []App{
		newApp("godoc", 6060),
		newApp("goapp", 8080),
		newApp("btsync", 8888),
	}
	p := NewProxy(newAppCenter(apps), "local")

	resolveCheck := func(name, host string) {
		a, ok := p.resolve(host)
		if !ok {
			t.Errorf("cant resolve %s", host)
		}

		if a.Name() != name {
			t.Errorf("wrong app. want: %s, got: %s", name, a.Name())
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
		a, ok := p.resolve(host)
		if ok {
			t.Errorf("app '%s' was found for host '%s'", a.Name(), host)
		}
	}

	unresolvedCheck("p2p.local")
	unresolvedCheck("godoc.acme.local")
	unresolvedCheck("godoc.dev")
}

func TestProxy(t *testing.T) {
	apps := []App{}
	createServer := func(name string, status int, content string) *httptest.Server {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(status)
			fmt.Fprint(w, content)
		}))
		port := getServerPort(t, s.URL)
		apps = append(apps, newApp(name, port))
		return s
	}

	const myappContent, myappStatus = "hello world", 201
	myapp := createServer("myapp", myappStatus, myappContent)
	defer myapp.Close()

	const fooContent, fooStatus = "foo bar", 404
	foo := createServer("foo", fooStatus, fooContent)
	defer foo.Close()

	proxy := httptest.NewServer(NewProxy(newAppCenter(apps), "local"))
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

	missedRequestCheck := func(host string, expectedStatus int) {
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
	}

	t.SkipNow()
	missedHosts := []string{"foo", "bar", "baz"}
	for _, host := range missedHosts {
		missedRequestCheck(host, http.StatusNotFound)
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

type fakeApp struct {
	app
}

func (a *fakeApp) Start() error  { return nil }
func (a *fakeApp) Stop() error   { return nil }
func (a *fakeApp) Running() bool { return true }

type fakeAppCenter struct {
	fakeApp
	apps map[string]App
}

func (ac *fakeAppCenter) Get(name string) (App, bool) {
	a, ok := ac.apps[name]
	return a, ok
}

func newAppCenter(apps []App) AppCenter {
	ac := &fakeAppCenter{}
	ac.apps = make(map[string]App)
	for _, a := range apps {
		ac.apps[a.Name()] = a
	}
	return ac
}

func newApp(name string, port int) App {
	a := &fakeApp{}
	a.name = name
	a.port = port
	return a
}
