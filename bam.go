package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"text/template"

	"github.com/BurntSushi/toml"
)

const programName = "bam"
const programVersion = "0.1.0"

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
	fmt.Fprintf(os.Stderr, "Usage: %s [OPTION]...\n", programName)
	fmt.Fprintf(os.Stderr, "A web server for developers.\n\n")
	flag.PrintDefaults()
}

func main() {
	versionFlag := flag.Bool("v", false, "print version information and exit")
	configFlag := flag.String("config", "", "use a configuration file")
	generateFlag := flag.String("generate", "", "generate configuration file(s). Use 'bam -generate help' to show generate options.")

	flag.Usage = usage
	flag.Parse()

	if *versionFlag {
		fmt.Printf("%s %s\n", programName, programVersion)
		return
	}

	cfg := parseConfig(*configFlag)

	if *generateFlag != "" {
		generate(*generateFlag, cfg)
		return
	}

	log.SetPrefix("[bam] ")
	cc := NewCommandCenter(programName, cfg)
	go func() {
		log.Printf("Starting CommandCenter at %s\n", cc.rootURL())
		fail(cc.Start())
	}()

	proxyAddr := fmt.Sprintf(":%d", cfg.ProxyPort)
	l, err := net.Listen("tcp", proxyAddr)
	fail(err)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	forceShutdown := false
	var gracefulShutdown sync.Once
	go func() {
		for sig := range c {
			if forceShutdown {
				log.Println("Exiting")
				os.Exit(1)
			}

			go gracefulShutdown.Do(func() {
				log.Printf("%v signal received, stopping applications and exiting.", sig)
				forceShutdown = true

				log.Printf("stopping CommandCenter")
				cc.Stop()

				log.Printf("stopping proxy")
				l.Close()
			})
		}
	}()

	proxy := NewProxy(cc, cfg.Tld)
	log.Println("Starting Proxy at", proxyAddr)
	s := http.Server{Handler: proxy}
	s.Serve(l)
}

const defaultConfig = `# BAM! configuration file

# apps_dir is the path where Procfile-based applications will be searched.
apps_dir = "."

# tld is the top-level domain for local applications.
tld = "dev"

# Automatically starts all applications found on startup if set as true.
auto_start = false

# proxy_port is the port where all :80 connections will be forwarded to before reaching any of the applications.
proxy_port = 42042

# aliases maps names for local ports used by applications not managed by bam.
#[aliases]
#btsync = 8080
#transmission = 9091
`
