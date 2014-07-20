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
  "proxy_port": 42042,
  "aliases": { }
}`

type Config struct {
	AppsDir   string         `json:"apps_dir"`
	Aliases   map[string]int `json:"aliases"`
	ProxyPort int            `json:"proxy_port"`
	Tld       string         `json:"tld"`
}

func parseConfig(file string) *Config {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatalln("[ERROR]", err)
	}

	c := &Config{}
	if err = json.Unmarshal([]byte(defaultConfig), &c); err != nil {
		log.Fatalln("[ERROR]", err)
	}
	if err = json.Unmarshal(content, &c); err != nil {
		log.Fatalln("[ERROR]", err)
	}

	return c
}

func main() {
	flag.Parse()

	c := parseConfig(*configPath)

	cc := NewCommandCenter(c.Tld, c.AppsDir, c.Aliases)
	go func() {
		log.Printf("Starting CommandCenter at http://bam.%s\n", c.Tld)
		log.Fatal(cc.Start())
	}()

	proxy := NewProxy(c.Tld, cc)
	proxyAddr := fmt.Sprintf(":%d", c.ProxyPort)
	log.Println("Starting Proxy at", proxyAddr)
	log.Fatal(http.ListenAndServe(proxyAddr, proxy))
}
