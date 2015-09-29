package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"os/user"
	"sync"

	"github.com/BurntSushi/toml"
)

var (
	programName    = "bam"
	programVersion = "dev"
)

type Config struct {
	AppsDir   string         `toml:"apps_dir"`
	User      string         `toml:"user"`
	Tld       string         `toml:"tld"`
	AutoStart bool           `toml:"auto_start"`
	ProxyPort int            `toml:"proxy_port"`
	Aliases   map[string]int `toml:"aliases"`
	user      *user.User
}

func (c *Config) GetUser() *user.User {
	return c.user
}

func parseConfig(file string) *Config {
	c := &Config{}

	_, err := toml.Decode(defaultConfig, &c)
	fail(err)

	if file != "" {
		_, err = toml.DecodeFile(file, &c)
		fail(err)
	}

	user, err := lookupUser(c.User)
	fail(err)
	c.user = user
	return c
}

func lookupUser(username string) (*user.User, error) {
	if username == "" {
		return nil, nil
	}

	return user.Lookup(username)
}

func fail(e error) {
	if e != nil {
		log.Fatalln("ERROR ", e)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [OPTION]...\n", programName)
	fmt.Fprintf(os.Stderr, "A web server for developers.\n\n")
	flag.PrintDefaults()
}

func main() {
	versionFlag := flag.Bool("v", false, "print version information and exit")
	sampleConfigFlag := flag.Bool("dump-sample-config", false, "print a sample configuration file")
	configFlag := flag.String("config", "", "use a configuration file")

	flag.Usage = usage
	flag.Parse()

	if *versionFlag {
		fmt.Printf("%s %s\n", programName, programVersion)
		return
	}

	if *sampleConfigFlag {
		fmt.Print(defaultConfig)
		return
	}

	cfg := parseConfig(*configFlag)
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

# define the user credentials used by application processes. If empty, application processes run under the same user as bam.
user = ""

# tld is the top-level domain for local applications.
tld = "dev"

# Automatically starts all applications found on startup if set as true.
auto_start = false

# proxy_port is the port where all connections will be forwarded to before reaching any of the applications.
proxy_port = 80

# aliases maps names for local ports used by applications not managed by bam.
#[aliases]
#btsync = 8080
#transmission = 9091
`
