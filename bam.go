package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"text/template"

	"github.com/BurntSushi/toml"
)

const programVersion = "0.0.1-alpha"

var configTemplates = make(map[string]string)

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

func generate(name string, c *Config) {
	tpl, ok := configTemplates[name]
	if !ok {
		fmt.Fprintf(os.Stderr, configTemplates["help"])
		os.Exit(1)
	}

	fail(template.Must(template.New(name).Parse(tpl)).Execute(os.Stdout, c))
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [OPTION]...\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "A web server for developers.\n\n")
	flag.PrintDefaults()
}

func main() {
	versionFlag := flag.Bool("v", false, "print version information and exit")
	configFlag := flag.String("config", "", "use a configuration file")
	generateFlag := flag.String("generate", "", "generate configuration file(s)")

	flag.Usage = usage
	flag.Parse()

	if *versionFlag {
		fmt.Printf("bam %s\n", programVersion)
		return
	}

	cfg := parseConfig(*configFlag)

	if *generateFlag != "" {
		generate(*generateFlag, cfg)
		return
	}

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

const defaultConfig = `# BAM! configuration file

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
