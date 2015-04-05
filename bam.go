package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/BurntSushi/toml"
)

var (
	configPath     = flag.String("config", "", "Use a configuration file")
	dumpConfigFile = flag.Bool("dump-sample-config", false, "Print a sample configuration file")
)

type Config struct {
	AppsDir   string         `toml:"apps_dir"`
	Tld       string         `toml:"tld"`
	AutoStart bool           `toml:"auto_start"`
	ProxyPort int            `toml:"proxy_port"`
	Aliases   map[string]int `toml:"aliases"`
}

func parseConfig(file string) *Config {
	c := &Config{}

	_, err := toml.Decode(defaultConfig, &c)
	fail(err)

	if file != "" {
		_, err = toml.DecodeFile(file, &c)
		fail(err)
	}

	return c
}

func fail(e error) {
	if e != nil {
		log.Fatalln("ERROR ", e)
	}
}

func main() {
	flag.Parse()

	if *dumpConfigFile {
		fmt.Print(defaultConfig)
		return
	}

	cfg := parseConfig(*configPath)

	log.SetPrefix("[bam] ")
	cc := NewCommandCenter(cfg)
	go func() {
		log.Printf("Starting CommandCenter at http://bam.%s\n", cfg.Tld)
		fail(cc.Start())
	}()

	proxy := NewProxy(cfg.Tld, cc)
	proxyAddr := fmt.Sprintf(":%d", cfg.ProxyPort)
	log.Println("Starting Proxy at", proxyAddr)
	fail(http.ListenAndServe(proxyAddr, proxy))
}

const defaultConfig = `
# apps_dir is the path where Procfile-based applications will be searched.
apps_dir = "."

# tld is the top-level domain for local applications.
tld = "app"

# Automatically starts all applications found on startup if set as true.
auto_start = false

# proxy_port is the port where all :80 connections will be forwarded to before reaching any of the applications.
proxy_port = 42042

# aliases maps names for local ports used by applications not managed by bam.
#[aliases]
#btsync = 8080
#transmission = 9091
`
