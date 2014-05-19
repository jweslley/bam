package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

var (
	configPath = flag.String("config", os.Getenv("HOME")+"/.bam/apps.json", "Config file")
	httpAddr   = flag.String("http", ":42042", "HTTP service address")
	tld        = flag.String("tld", getValue("LOCALTLD", "local"),
		"Local top-level domain. Defaults to environment variable LOCALTLD")
)

func getValue(key, defaultValue string) string {
	if value := os.Getenv(key); value == "" {
		return defaultValue
	} else {
		return value
	}
}

func parseConfig(file string) map[string]int {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatalln("[ERROR]", err)
	}

	s := make(map[string]int)
	if err = json.Unmarshal(content, &s); err != nil {
		log.Fatalln("[ERROR]", err)
	}

	return s
}

func main() {
	flag.Parse()

	s := parseConfig(*configPath)
	servers := []Server{}
	for name, port := range s {
		servers = append(servers, NewServer(name, port))
	}

	proxy := NewProxy(*tld, servers...)
	log.Println("Starting HTTP server at", *httpAddr)
	log.Fatal(http.ListenAndServe(*httpAddr, proxy))
}
