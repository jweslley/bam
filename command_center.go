package main

import (
	"fmt"
	"log"
	"net/http"
	"text/template"
)

type CommandCenter struct {
	server
	tld     string
	apps    []*App
	servers []Server
}

func NewCommandCenter(tld, appsDir string, aliases map[string]int) *CommandCenter {
	cc := &CommandCenter{tld: tld}
	cc.name = "bam"
	cc.servers = createServers(aliases)
	cc.apps = LoadApps(appsDir)
	return cc
}

func (cc *CommandCenter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	data := map[string]interface{}{
		"Tld":     cc.tld,
		"Servers": cc.servers,
	}
	tmpl, err := template.New("-").Parse(baseHTML)
	log.Print(err)
	tmpl.Execute(w, data)
}

func (cc *CommandCenter) Start() error {
	port, _ := FreePort()
	cc.port = port
	go func() {
		for _, app := range cc.apps {
			log.Printf("starting app at cc %s", app.Name())
			go app.Start()
		}
	}()
	return http.ListenAndServe(fmt.Sprintf(":%d", cc.port), cc)
}

func createServers(aliases map[string]int) []Server {
	servers := []Server{}
	for name, port := range aliases {
		servers = append(servers, NewServer(name, port))
	}
	return servers
}

const baseHTML = `
<html>
  <head>
    <title>Bam's Command Center</title>
    <style type="text/css">
      body {
        background-color: #ecf0f1;
        font-family: Helvetica,Arial,sans-serif;
      }
      #container {
        width: 90%;
        max-width: 750px;
        padding-right: 15px;
        padding-left: 15px;
        margin-right: auto;
        margin-left: auto;
      }
      ul {
        list-style: none;
        padding: 0;
        margin: 10px 0;
      }
      li {
        padding: 10px;
        margin: 5px;
        background-color: #3498db;
      }
      li.green { background-color: #1abc9c; }
      li.red { background-color: #e74c3c; }
      a {
        color: white;
        text-decoration: none;
      }
    </style>
  </head>
  <body>
    <div id="container">
      <h1>Applications</h1>
      <ul>
				{{$tld := .Tld}}
				{{range .Servers}}
					<li><a href="http://{{.Name}}.{{$tld}}">{{.Name}}</a></li>
				{{end}}
      </ul>
    </div>
  </body>
</html>
`
