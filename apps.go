package main

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/NoahShen/gotunnelme/src/gotunnelme"
	"github.com/jweslley/procker"
)

type App interface {
	Name() string

	Port() int

	Start() error

	Stop() error

	Running() bool
}

var (
	errAlreadyStarted = errors.New("Already started")
	errNotStarted     = errors.New("Not started")
	errAlreadyShared  = errors.New("Already shared")
)

type app struct {
	name string
	port int
}

func (a *app) Name() string {
	return a.name
}

func (a *app) Port() int {
	return a.port
}

func (a *app) String() string {
	return fmt.Sprintf("%s:%d", a.name, a.port)
}

type processApp struct {
	app
	dir       string
	env       []string
	processes map[string]string
	process   procker.Process
}

func (a *processApp) Start() error {
	if a.Running() {
		return errAlreadyStarted
	}

	p, err := a.buildProcess()
	if err != nil {
		return err
	}

	a.process = p
	return a.process.Start()
}

func (a *processApp) Stop() error {
	if !a.Running() {
		return errNotStarted
	}

	err := a.process.Stop(3 * time.Second) // FIXME magic number
	a.process = nil
	return err
}

func (a *processApp) Running() bool {
	return a.process != nil && a.process.Running()
}

func (a *processApp) buildProcess() (procker.Process, error) {
	port, err := FreePort()
	if err != nil {
		return nil, err
	}

	a.port = port
	p := []procker.Process{}
	for name, command := range a.processes {
		prefix := fmt.Sprintf("[%s:%s] ", a.Name(), name)
		process := &procker.SysProcess{
			Command: command,
			Dir:     a.dir,
			Env:     append(a.env, fmt.Sprintf("PORT=%d", port)),
			Stdout:  procker.NewPrefixedWriter(os.Stdout, prefix),
			Stderr:  procker.NewPrefixedWriter(os.Stderr, prefix),
		}
		p = append(p, process)
	}
	return procker.NewProcessGroup(p...), nil
}

func NewProcessApp(procfile string) (App, error) {
	processes, err := parseProfile(procfile)
	if err != nil {
		return nil, err
	}

	dir := path.Dir(procfile)
	name := path.Base(dir)
	envFile := path.Join(dir, ".env")
	env, err := parseEnv(envFile)
	if err != nil {
		env = []string{}
	}

	a := &processApp{dir: dir, env: env, processes: processes}
	a.name = name
	return a, nil
}

func parseProfile(filepath string) (map[string]string, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	processes, err := procker.ParseProcfile(file)
	if err != nil {
		return nil, err
	}
	return processes, nil
}

func parseEnv(filepath string) ([]string, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	env, err := procker.ParseEnv(file)
	if err != nil {
		return nil, err
	}
	return env, nil
}

type aliasApp struct {
	app
	running bool
}

func (a *aliasApp) Start() error {
	a.running = true
	return nil
}

func (a *aliasApp) Stop() error {
	a.running = false
	return nil
}

func (a *aliasApp) Running() bool {
	return a.running
}

func NewAliasApp(name string, port int) App {
	a := &aliasApp{}
	a.name = name
	a.port = port
	a.running = true
	return a
}

type webApp struct {
	app
	handler  http.Handler
	listener net.Listener
}

func (a *webApp) Start() error {
	if a.Running() {
		return errAlreadyStarted
	}

	l, err := NewLocalListener()
	if err != nil {
		return err
	}

	port, err := AddrPort(l.Addr().String())
	if err != nil {
		return err
	}

	a.listener = l
	a.port = port

	s := &http.Server{Handler: a.handler}
	go func() {
		s.Serve(a.listener)
		a.listener = nil
	}()

	return nil
}

func (a *webApp) Stop() error {
	if !a.Running() {
		return errNotStarted
	}

	err := a.listener.Close()
	a.listener = nil
	return err
}

func (a *webApp) Running() bool {
	return a.listener != nil
}

func NewWebServerApp(dir string) App {
	a := &webApp{}
	a.name = path.Base(dir)
	a.handler = http.StripPrefix("/", http.FileServer(http.Dir(dir)))
	return a
}

type ShareableApp struct {
	App
	url    string
	tunnel *gotunnelme.Tunnel
}

func (a *ShareableApp) Stop() error {
	if a.Shared() {
		go func() { a.Unshare() }()
	}

	return a.App.Stop()
}

func (a *ShareableApp) Share() error {
	if !a.Running() {
		return errNotStarted
	}

	if a.Shared() {
		return errAlreadyShared
	}

	tunnel := gotunnelme.NewTunnel()
	url, err := tunnel.GetUrl("")
	if err != nil {
		return err
	}

	a.url = url
	a.tunnel = tunnel

	go func() {
		tunnel.CreateTunnel(a.Port())
		a.tunnel = nil
		a.url = ""
	}()

	return nil
}

func (a *ShareableApp) Unshare() error {
	if !a.Running() {
		return errNotStarted
	}

	if a.Shared() {
		a.tunnel.StopTunnel()
	}

	return nil
}

func (a *ShareableApp) Shared() bool {
	return a.tunnel != nil
}

func (a *ShareableApp) URL() string {
	return a.url
}
