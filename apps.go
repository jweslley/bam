package main

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/jweslley/procker"
)

type App interface {
	Server

	Start() error

	Stop() error

	Running() bool
}

var (
	errAlreadyStarted = errors.New("Already started")
	errNotStarted     = errors.New("Not started")
)

type processApp struct {
	server
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
	server
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

type webServerApp struct {
	server
	dir      string
	listener net.Listener
}

func (a *webServerApp) Start() error {
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

	s := &http.Server{Handler: http.StripPrefix("/", http.FileServer(http.Dir(a.dir)))}
	go func() {
		s.Serve(a.listener)
		a.listener = nil
	}()

	return nil
}

func (a *webServerApp) Stop() error {
	if !a.Running() {
		return errNotStarted
	}

	err := a.listener.Close()
	a.listener = nil
	return err
}

func (a *webServerApp) Running() bool {
	return a.listener != nil
}

func NewWebServerApp(dir string) App {
	a := &webServerApp{dir: dir}
	a.name = path.Base(dir)
	return a
}
