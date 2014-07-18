package main

import (
	"fmt"
	"net/http"
	"testing"
)

func TestApps(t *testing.T) {
	apps := LoadApps("./test")
	if len(apps) != 2 {
		t.Fatalf("App loading failed. Expected: %d, got: %d", 2, len(apps))
	}

	test := make(chan bool)
	for _, app := range apps {
		if app.Name() == "ping" {
			if len(app.env) != 1 {
				t.Fatalf("Unexpected env var count. got: %d", len(app.env))
			}
			if app.env[0] != "OWNER=master" {
				t.Fatalf("Unexpected env vars. got: %s", app.env[0])
			}
			if len(app.processes) != 1 {
				t.Fatalf("Unexpected process count. got: %d", len(app.processes))
			}
			if app.processes["web"] != "./ping -p $PORT" {
				t.Fatalf("Unexpected process. got: %v", app.processes)
			}
			if app.Port() != 0 {
				t.Fatalf("Unexpected port. got: %d", app.Port())
			}
		}

		if app.Name() == "fileserver" {
			if len(app.env) != 1 {
				t.Fatalf("Unexpected env var count. got: %d", len(app.env))
			}
			if app.env[0] != "FILE_SERVER_DIR=../.." {
				t.Fatalf("Unexpected env vars. got: %s", app.env[0])
			}
			if len(app.processes) != 1 {
				t.Fatalf("Unexpected process count. got: %d", len(app.processes))
			}
			if app.processes["web"] != "./fileserver -p $PORT -d $FILE_SERVER_DIR" {
				t.Fatalf("Unexpected process. got: %v", app.processes)
			}
			if app.Port() != 0 {
				t.Fatalf("Unexpected port. got: %d", app.Port())
			}
		}

		if app.Started() {
			t.Fatalf("app MUST not be started yet: %s", app.Name())
		}

		go func(a *App) {
			err := a.Start()
			if err != nil {
				t.Fatalf("Unable to start %s: %s", a.Name(), err)
			}

			if !a.Started() {
				t.Fatalf("app should be started now: %s", a.Name())
			}

			if a.Port() == 0 {
				t.Fatalf("Port should not be zeroed: %s", a.Name())
			}

			req, _ := http.NewRequest("GET", fmt.Sprintf("http://localhost:%d", a.Port()), nil)
			res, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("Server unavailable %s: %s", a.Name(), err)
			}

			if res.StatusCode != 200 {
				t.Errorf("Unexpected status code for %s: %d", a.Name(), res.StatusCode)
			}

			err = a.Stop()
			if err != nil {
				t.Fatalf("Unable to stop %s: %s", a.Name(), err)
			}

			if a.Started() {
				t.Fatalf("app should NOT be started now: %s", a.Name())
			}

			test <- true
		}(app)

	}

	<-test
	<-test
}
