package main

import (
	"fmt"
	"log"
	"net/http"
	"path"
	"path/filepath"
	"text/template"
)

type CommandCenter struct {
	server
	tld       string
	apps      map[string]App
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
	cc.apps = make(map[string]App)
	cc.loadApps(c)
	cc.autoStart = c.AutoStart
	return cc
}

func (cc *CommandCenter) List() []Server {
	s := []Server{cc}
	for _, app := range cc.apps {
		s = append(s, app)
	}
	return s
}

func (cc *CommandCenter) Start() error {
	port, _ := FreePort()
	cc.port = port
	if cc.autoStart {
		go func() {
			cc.startApps()
		}()
	}
	return http.ListenAndServe(fmt.Sprintf(":%d", cc.port), cc.createHandler())
}

func (cc *CommandCenter) startApps() {
	for _, app := range cc.apps {
		log.Printf("Starting app %s\n", app.Name())
		go app.Start()
	}
}

func (cc *CommandCenter) createHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", cc.index)
	mux.HandleFunc("/start", cc.start)
	mux.HandleFunc("/stop", cc.stop)
	return mux
}

func (cc *CommandCenter) index(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	data := map[string]interface{}{
		"Tld":  cc.tld,
		"apps": cc.apps,
	}
	tmpl.Execute(w, data)
}

func (cc *CommandCenter) start(w http.ResponseWriter, r *http.Request) {
	cc.action(w, r, func(a App) error {
		log.Printf("Starting app %s\n", a.Name())
		return a.Start()
	})
}

func (cc *CommandCenter) stop(w http.ResponseWriter, r *http.Request) {
	cc.action(w, r, func(a App) error {
		log.Printf("Stopping app %s\n", a.Name())
		return a.Stop()
	})
}

func (cc *CommandCenter) action(w http.ResponseWriter, r *http.Request, action func(a App) error) {
	name := r.URL.Query().Get("app")
	for _, app := range cc.apps {
		if app.Name() == name {
			action(app)
			break
		}
	}
	http.Redirect(w, r, "/", 302)
}

func (cc *CommandCenter) register(a App) {
	if _, ok := cc.apps[a.Name()]; ok {
		return
	}
	cc.apps[a.Name()] = a
}

func (cc *CommandCenter) loadApps(c *Config) {
	cc.loadAliasApps(c.Aliases)
	cc.loadProcessApps(c.AppsDir)
	cc.loadWebServerApps(c.AppsDir)
}

func (cc *CommandCenter) loadAliasApps(aliases map[string]int) {
	for name, port := range aliases {
		cc.register(NewAliasApp(name, port))
	}
}

func (cc *CommandCenter) loadProcessApps(dir string) {
	procfiles, _ := filepath.Glob(fmt.Sprintf("%s/*/Procfile", dir))
	for _, p := range procfiles {
		app, err := NewProcessApp(p)
		if err != nil {
			log.Printf("Unable to load application %s. Error: %s\n", p, err.Error())
		} else {
			cc.register(app)
		}
	}
}

func (cc *CommandCenter) loadWebServerApps(dir string) {
	pages, _ := filepath.Glob(fmt.Sprintf("%s/*/index.html", dir))
	for _, p := range pages {
		cc.register(NewWebServerApp(path.Dir(p)))
	}
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
      <ul>
				{{range .apps}}
					{{ if .Running}}
						<li class="green">
					{{ else }}
						<li class="red">
					{{ end }}
						<div style="text-align: left">
						<a href="http://{{.Name}}.{{$tld}}">{{.Name}}</a>
						</div>
						<div style="text-align: right">
						{{ if .Running}}
							<a href="http://bam.{{$tld}}/stop?app={{.Name}}">Stop</a>
						{{ else }}
							<a href="http://bam.{{$tld}}/start?app={{.Name}}">Start</a>
						{{ end }}
						</div>
					</li>
				{{end}}
      </ul>
    </div>
  </body>
</html>
`
