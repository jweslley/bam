package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/jweslley/procker"
)

type App interface {
	Server

	Start() error

	Stop() error

	Running() bool
}

type processApp struct {
	server
	dir       string
	env       []string
	processes map[string]string
	process   procker.Process
}

func (a *processApp) Start() error {
	if a.Running() {
		return fmt.Errorf("bam: %s already started", a.Name())
	}

	a.process = a.buildProcess()
	return a.process.Start()
}

func (a *processApp) Stop() error {
	if !a.Running() {
		return fmt.Errorf("bam: %s not started", a.Name())
	}

	err := a.process.Stop(3 * time.Second) // FIXME magic number
	a.process = nil
	return err
}

func (a *processApp) Running() bool {
	return a.process != nil && a.process.Running()
}

func (a *processApp) buildProcess() procker.Process {
	port, _ := FreePort() // FIXME swallow error
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
	return procker.NewProcessGroup(p...)
}

func LoadApps(dir string) []App {
	apps := []App{}
	procfiles, _ := filepath.Glob(fmt.Sprintf("%s/*/Procfile", dir))
	for _, p := range procfiles {
		app, err := NewProcessApp(p)
		if err != nil {
			log.Printf("Unable to load application %s. Error: %s\n", p, err.Error())
		} else {
			apps = append(apps, app)
		}
	}
	return apps
}

func NewProcessApp(procfile string) (App, error) {
	processes, err := parseProfile(procfile)
	if err != nil {
		log.Printf("Unable to load Procfile %s: %s\n", procfile, err)
		return nil, err
	}

	dir := path.Dir(procfile)
	name := path.Base(dir)
	envFile := path.Join(dir, ".env")
	env, err := parseEnv(envFile)
	if err != nil {
		log.Printf("Unable to load env file %s: %s\n", envFile, err)
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
