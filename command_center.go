package main

import (
	"fmt"
	"log"
	"net/http"
	"text/template"
)

type CommandCenter struct {
	server
	tld       string
	apps      []*App
	servers   []Server
	autoStart bool
}

var tmpl *template.Template

func init() {
	t, err := template.New("-").Parse(baseHTML)
	fail(err)
	tmpl = t
}

func NewCommandCenter(c *Config) *CommandCenter {
	cc := &CommandCenter{tld: c.Tld}
	cc.name = "bam"
	cc.servers = createAliasedServers(c.Aliases)
	cc.apps = LoadApps(c.AppsDir)
	cc.autoStart = c.AutoStart
	return cc
}

func (cc *CommandCenter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	data := map[string]interface{}{
		"Tld":            cc.tld,
		"aliasedServers": cc.servers,
		"apps":           cc.apps,
	}
	tmpl.Execute(w, data)
}

func (cc *CommandCenter) Start() error {
	port, _ := FreePort()
	cc.port = port
	if cc.autoStart {
		go func() {
			cc.startApps()
		}()
	}
	return http.ListenAndServe(fmt.Sprintf(":%d", cc.port), cc)
}

func (cc *CommandCenter) List() []Server {
	s := cc.servers
	s = append(s, cc)
	for _, app := range cc.apps {
		s = append(s, app)
	}
	return s
}

func (cc *CommandCenter) startApps() {
	for _, app := range cc.apps {
		log.Printf("Starting app %s\n", app.Name())
		go app.Start()
	}
}

func createAliasedServers(aliases map[string]int) []Server {
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
		{{$tld := .Tld}}
    <div id="container">
      <h1>BAM!!!</h1>
      <h2>Aliased servers</h2>
      <ul>
				{{range .aliasedServers}}
					<li><a href="http://{{.Name}}.{{$tld}}">{{.Name}}</a></li>
				{{end}}
      </ul>
      <h2>Applications</h2>
      <ul>
				{{range .apps}}
					{{ if .Started}}
						<li class="green">
					{{ else }}
						<li class="red">
					{{ end }}
						<a href="http://{{.Name}}.{{$tld}}">{{.Name}}</a>
					</li>
				{{end}}
      </ul>
    </div>
  </body>
</html>
`
