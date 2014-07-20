package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

var (
	configPath = flag.String("config", "config.json", "Config file")
)

const defaultConfig = `{
  "apps_dir": ".",
  "tld": "app",
  "auto_start": true,
  "proxy_port": 42042,
  "aliases": { }
}`

type Config struct {
	AppsDir   string         `json:"apps_dir"`
	Aliases   map[string]int `json:"aliases"`
	ProxyPort int            `json:"proxy_port"`
	AutoStart bool           `json:"auto_start"`
	Tld       string         `json:"tld"`
}

func parseConfig(file string) *Config {
	content, err := ioutil.ReadFile(file)
	fail(err)

	c := &Config{}
	fail(json.Unmarshal([]byte(defaultConfig), &c))
	fail(json.Unmarshal(content, &c))
	return c
}

func fail(e error) {
	if e != nil {
		log.Fatalln("ERROR ", e)
	}
}

func main() {
	flag.Parse()

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
