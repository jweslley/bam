package main

import (
	"fmt"
	"log"
	"net/http"
	"path"
	"path/filepath"
	"text/template"
)

type data map[string]interface{}

var templates = make(map[string]*template.Template)

func init() {
	for name, html := range pagesHTML {
		t := template.New(name)
		template.Must(t.Parse(html))
		template.Must(t.Parse(baseHTML))
		templates[name] = t
	}
}

func render(w http.ResponseWriter, name string, d data) {
	w.Header().Add("Content-Type", "text/html")

	t, ok := templates[name]
	if !ok {
		renderError(w, http.StatusNotFound, fmt.Errorf("Page not found: %s", name))
		return
	}

	err := t.ExecuteTemplate(w, "root", d)
	if err != nil {
		renderError(w, http.StatusInternalServerError, err)
	}
}

func renderError(w http.ResponseWriter, status int, e error) {
	w.WriteHeader(status)
	err := templates["error"].ExecuteTemplate(w, "root", data{
		"Title": fmt.Sprintf("Error %d", status),
		"Error": e,
	})

	if err != nil {
		log.Println(err)
	}
}

type CommandCenter struct {
	server
	tld       string
	autoStart bool
	apps      map[string]App
	servers   []Server
}

func NewCommandCenter(c *Config) *CommandCenter {
	cc := &CommandCenter{tld: c.Tld}
	cc.name = "bam"
	cc.apps = make(map[string]App)
	cc.servers = []Server{cc}
	cc.loadApps(c)
	cc.autoStart = c.AutoStart
	return cc
}

func (cc *CommandCenter) List() []Server {
	return cc.servers
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
	render(w, "index", data{
		"Title": "BAM!",
		"Tld":   cc.tld,
		"Apps":  cc.apps,
	})
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
	app, _ := cc.apps[name] // FIXME not found
	action(app)
	http.Redirect(w, r, "/", 302)
}

func (cc *CommandCenter) register(a App) {
	if _, ok := cc.apps[a.Name()]; ok {
		return
	}
	cc.apps[a.Name()] = a
	cc.servers = append(cc.servers, a)
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
{{ define "root" }}
<html>
  <head>
    <title>{{.Title}}</title>
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
			<h1>{{.Title}}</h1>
			{{ template "body" . }}
    </div>
  </body>
</html>
{{ end }}
`

var pagesHTML = map[string]string{
	"index": `
	{{ define "body" }}
		{{$tld := .Tld}}
		<ul>
			{{range .Apps}}
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
	{{ end }}`,
	"error": `
	{{ define "body" }}
		{{.Error}}
	{{ end }}`,
}
