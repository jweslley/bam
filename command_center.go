package main

import (
	"fmt"
	"net/http"
)

type CommandCenter struct {
	server
	tld     string
	servers []Server
}

func NewCommandCenter(tld string, s ...Server) *CommandCenter {
	cc := &CommandCenter{tld: tld}
	cc.name = "bam"
	cc.servers = s
	return cc
}

func (cc *CommandCenter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	fmt.Fprintf(w, `<http><head><title>Bam's CommandCenter</title></head><body><ul>`)
	for _, s := range cc.servers {
		fmt.Fprintf(w, "<li><a href=\"http://%s.%s\">%s</a></li>", s.Name(), cc.tld, s.Name())
	}
	fmt.Fprintf(w, `</ul></body></html>`)
}

func (cc *CommandCenter) Start() error {
	port, _ := FreePort()
	cc.port = port
	return http.ListenAndServe(fmt.Sprintf(":%d", cc.port), cc)
}
