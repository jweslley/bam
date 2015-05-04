package main

//go:generate esc -o command_center_assets.go -prefix=public public/bam.js public/bam.css public/images/

import (
	"bytes"
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
	templates map[string]*template.Template
}

func NewCommandCenter(name string, c *Config) *CommandCenter {
	cc := &CommandCenter{tld: c.Tld, autoStart: c.AutoStart}
	cc.name = name
	cc.apps = make(map[string]App)
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

	b := &bytes.Buffer{}
	err := t.ExecuteTemplate(b, "root", d)
	if err != nil {
		cc.renderError(w, http.StatusInternalServerError, err)
	} else {
		w.Write(b.Bytes())
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

func (cc *CommandCenter) Get(name string) (Server, bool) {
	if cc.name == name {
		return cc, true
	}
	app, ok := cc.apps[name]
	return app, ok
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
	cc.action(w, r, "starting", func(a App) error {
		return a.Start()
	})
}

func (cc *CommandCenter) stop(w http.ResponseWriter, r *http.Request) {
	cc.action(w, r, "stopping", func(a App) error {
		return a.Stop()
	})
}

func (cc *CommandCenter) action(w http.ResponseWriter, r *http.Request, desc string, action func(a App) error) {
	name := r.URL.Query().Get("app")
	app, found := cc.apps[name]
	if !found {
		cc.renderError(w, http.StatusNotFound, fmt.Errorf("Application not found: %s", name))
		return
	}

	log.Printf("%s %s\n", desc, name)
	err := action(app)
	if err != nil {
		log.Printf("ERROR: %s %s: %v\n", desc, name, err)
		cc.renderError(w, http.StatusInternalServerError,
			fmt.Errorf("An error occurred while %s %s: %v", desc, name, err))
	} else {
		http.Redirect(w, r, "/", http.StatusFound)
	}
}

func (cc *CommandCenter) notFound(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("app")
	log.Printf("WARN Application not found: %s\n", name)
	cc.renderError(w, http.StatusNotFound, fmt.Errorf("Application doesn't exist: %s", name))
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
    <meta charset="utf-8">
    <title>{{.Title}}</title>
		<link rel="stylesheet" type="text/css" href="{{ assetPath "bam.css" }}">
  </head>
  <body>
    <div id="container">
			{{ template "body" . }}
    </div>
    <script type="text/javascript" src="{{ assetPath "bam.js" }}"></script>
  </body>
</html>
{{ end }}
`

var pagesHTML = map[string]string{
	"index": `
	{{ define "body" }}
		<h1> <a href="{{ rootURL }}">BAM!</a> </h1>
		<input type="text" id="search-box" placeholder="Search" onkeyup="search();"></input>
		<ul class="list">
			{{range .Apps}}
				{{ if .Running}}
					<li data-app="{{.Name}}" class="green">
				{{ else }}
					<li data-app="{{.Name}}" class="red">
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
