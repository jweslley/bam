package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"regexp"
	"strings"
)

var (
	configPath = flag.String("config", os.Getenv("HOME")+"/.bam/apps.json", "Config file")
	httpAddr   = flag.String("http", ":42042", "HTTP service address")
	tld        = flag.String("tld", getValue("LOCALTLD", "app"),
		"Local top-level domain. Defaults to environment variable LOCALTLD")
)

var xipio = regexp.MustCompile("^(.*?)\\.?\\d+\\.\\d+\\.\\d+\\.\\d+\\.xip\\.io")

func getValue(key, defaultValue string) string {
	if value := os.Getenv(key); value == "" {
		return defaultValue
	} else {
		return value
	}
}

type App struct {
	Name string `json:"name"`
	Port int    `json:"port"`
}

func (a *App) Match(host, tld string) bool {
	apphost := a.Name + "." + tld
	return apphost == host || strings.HasSuffix(host, "."+apphost) || a.isXipDomain(host)
}

func (a *App) isXipDomain(host string) bool {
	name := xipio.ReplaceAllString(host, "$1")
	return a.Name == name || strings.HasSuffix(name, "."+a.Name)
}

func (a *App) Host() string {
	return fmt.Sprint("127.0.0.1:", a.Port)
}

type AppsProxy struct {
	httputil.ReverseProxy
	apps []App
	tld  string
}

func (p *AppsProxy) GetAppHost(host string) (string, bool) {
	for _, app := range p.apps {
		if app.Match(host, p.tld) {
			return app.Host(), true
		}
	}
	return "", false
}

func NewAppsProxy(tld string) *AppsProxy {
	p := &AppsProxy{tld: tld}
	p.Director = func(req *http.Request) {
		req.URL.Scheme = "http"
		if host, found := p.GetAppHost(req.Host); found {
			req.URL.Host = host
		} else {
			log.Println("[ERROR] bam: No host found: ", req.Host)
		}
	}
	return p
}

var proxy = NewAppsProxy(*tld)

func init() {
	flag.Parse()

	content, err := ioutil.ReadFile(*configPath)
	if err != nil {
		log.Fatalln("[ERROR]", err)
	}

	if err = json.Unmarshal(content, &proxy.apps); err != nil {
		log.Fatalln("[ERROR]", err)
	}
}

func main() {
	log.Println("Starting HTTP server at", *httpAddr)
	log.Fatal(http.ListenAndServe(*httpAddr, proxy))
}
