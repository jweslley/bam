package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"

	"github.com/jweslley/procker"
)

type App struct {
	server
	dir       string
	env       []string
	processes map[string]string
	process   procker.Process
}

func (a *App) Start() error {
	if a.Started() {
		return fmt.Errorf("bam: %s already started", a.Name())
	}

	a.process = a.buildProcess()
	return a.process.Start()
}

func (a *App) Stop() error {
	if !a.Started() {
		return fmt.Errorf("bam: %s not started", a.Name())
	}

	err := a.process.Stop(1000)
	a.process = nil
	return err
}

func (a *App) Started() bool {
	return a.process != nil && a.process.Running()
}

func (a *App) buildProcess() procker.Process {
	port, _ := FreePort()
	a.port = port
	p := []procker.Process{}
	for name, command := range a.processes {
		prefix := fmt.Sprintf("[%s:%s] ", a.Name(), name)
		process := procker.NewProcess(
			command,
			a.dir,
			append(a.env, fmt.Sprintf("PORT=%d", port)),
			procker.NewPrefixedWriter(os.Stdout, prefix),
			procker.NewPrefixedWriter(os.Stderr, prefix))
		p = append(p, process)
	}
	return procker.NewProcessGroup(p...)
}

func LoadApps(dir string) []*App {
	apps := []*App{}
	procfiles, _ := filepath.Glob(fmt.Sprintf("%s/*/Procfile", dir))
	for _, p := range procfiles {
		app, err := newApp(p)
		if err != nil {
			log.Printf("Unable to load application %s. Error: %s\n", p, err.Error())
		} else {
			apps = append(apps, app)
		}
	}
	return apps
}

func newApp(procfile string) (*App, error) {
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

	a := &App{dir: dir}
	a.name = name
	a.env = env
	a.processes = processes
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
