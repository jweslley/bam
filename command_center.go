package main

//go:generate esc -o command_center_assets.go -prefix=public public

import (
	"fmt"
	"log"
	"net/http"
	"path"
	"path/filepath"
	"text/template"
)

type data map[string]interface{}

type CommandCenter struct {
	server
	tld       string
	autoStart bool
	apps      map[string]App
	servers   []Server
	templates map[string]*template.Template
}

func NewCommandCenter(c *Config) *CommandCenter {
	cc := &CommandCenter{tld: c.Tld, autoStart: c.AutoStart}
	cc.name = "bam"
	cc.apps = make(map[string]App)
	cc.servers = []Server{cc}
	cc.parseTemplates()
	cc.loadApps(c)
	return cc
}

func (cc *CommandCenter) parseTemplates() {
	tf := template.FuncMap{
		"rootURL":   cc.rootURL,
		"assetPath": cc.assetPath,
		"appURL":    cc.appURL,
		"actionURL": cc.actionURL,
	}
	cc.templates = make(map[string]*template.Template)
	for name, html := range pagesHTML {
		t := template.New(name).Funcs(tf)
		template.Must(t.Parse(html))
		template.Must(t.Parse(baseHTML))
		cc.templates[name] = t
	}
}

func (cc *CommandCenter) render(w http.ResponseWriter, name string, d data) {
	w.Header().Add("Content-Type", "text/html")

	t, ok := cc.templates[name]
	if !ok {
		cc.renderError(w, http.StatusNotFound, fmt.Errorf("Page not found: %s", name))
		return
	}

	err := t.ExecuteTemplate(w, "root", d)
	if err != nil {
		cc.renderError(w, http.StatusInternalServerError, err)
	}
}

func (cc *CommandCenter) renderError(w http.ResponseWriter, status int, e error) {
	w.WriteHeader(status)
	err := cc.templates["error"].ExecuteTemplate(w, "root", data{
		"Title": fmt.Sprintf("Error %d", status),
		"Error": e,
	})

	if err != nil {
		log.Printf("ERROR: %s\n", err)
	}
}

func (cc *CommandCenter) rootURL() string {
	return cc.appURL(cc.name)
}

func (cc *CommandCenter) assetPath(path string) string {
	return fmt.Sprintf("%s/assets/%s", cc.rootURL(), path)
}

func (cc *CommandCenter) appURL(app string) string {
	return fmt.Sprintf("http://%s.%s", app, cc.tld)
}

func (cc *CommandCenter) actionURL(action, app string) string {
	return fmt.Sprintf("%s/%s?app=%s", cc.rootURL(), action, app)
}

func (cc *CommandCenter) List() []Server {
	return cc.servers
}

func (cc *CommandCenter) Start() error {
	port, err := FreePort()
	if err != nil {
		return err
	}

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
		go func(a App) {
			log.Printf("Starting app %s\n", a.Name())
			err := a.Start()
			if err != nil {
				log.Printf("Failed to start %s: %s\n", a.Name(), err)
			}
		}(app)
	}
}

func (cc *CommandCenter) createHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", cc.index)
	mux.HandleFunc("/start", cc.start)
	mux.HandleFunc("/stop", cc.stop)
	mux.HandleFunc("/not-found", cc.notFound)
	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(FS(false))))
	return mux
}

func (cc *CommandCenter) index(w http.ResponseWriter, r *http.Request) {
	cc.render(w, "index", data{
		"Title": "BAM!",
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
	app, found := cc.apps[name]
	if !found {
		cc.renderError(w, http.StatusNotFound, fmt.Errorf("Application not found: %s", name))
		return
	}

	err := action(app)
	if err != nil {
		cc.renderError(w, http.StatusInternalServerError, err)
	} else {
		http.Redirect(w, r, "/", http.StatusFound)
	}
}

func (cc *CommandCenter) notFound(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("app")
	cc.renderError(w, http.StatusNotFound, fmt.Errorf("Application doesn't exist: %s", name))
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
	procfiles, err := filepath.Glob(fmt.Sprintf("%s/*/Procfile", dir))
	if err != nil {
		log.Printf("An error occurred while searching for Procfiles at directory %s: %s\n", dir, err)
		return
	}

	for _, p := range procfiles {
		app, err := NewProcessApp(p)
		if err != nil {
			log.Printf("Unable to load application %s. Error: %s\n", p, err)
		} else {
			cc.register(app)
		}
	}
}

func (cc *CommandCenter) loadWebServerApps(dir string) {
	pages, err := filepath.Glob(fmt.Sprintf("%s/*/index.html", dir))
	if err != nil {
		log.Printf("An error occurred while searching for index.html at directory %s: %s\n", dir, err)
		return
	}

	for _, p := range pages {
		cc.register(NewWebServerApp(path.Dir(p)))
	}
}

const baseHTML = `
{{ define "root" }}
<html>
  <head>
    <title>{{.Title}}</title>
		<link rel="stylesheet" type="text/css" href="{{ assetPath "bam.css"}}">
  </head>
  <body>
    <div id="container">
			{{ template "body" . }}
    </div>
  </body>
</html>
{{ end }}
`

var pagesHTML = map[string]string{
	"index": `
	{{ define "body" }}
		<h1> <a href="{{ rootURL }}">BAM!</a> </h1>
		<ul class="list">
			{{range .Apps}}
				{{ if .Running}}
					<li class="green">
				{{ else }}
					<li class="red">
				{{ end }}
					<a class="title" href="{{ appURL .Name }}">{{.Name}}</a>
					<span></span>
					<ul class="actions">
						<li>
							{{ if .Running}}
								<a href="{{ actionURL "stop" .Name }}">
									<img src="{{ assetPath "images/stop.png" }}">
								</a>
							{{ else }}
								<a href="{{ actionURL "start" .Name }}">
									<img src="{{ assetPath "images/start.png" }}">
								</a>
							{{ end }}
						</li>
					</ul>
				</li>
			{{end}}
		</ul>
	{{ end }}`,
	"error": `
	{{ define "body" }}
		<h1> <a href="{{ rootURL }}">BAM!</a> </h1>
		<div class="error-box">
			<h3>{{.Title}}</h3>
			{{.Error}}
		</div>
	{{ end }}`,
}
