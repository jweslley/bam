package main

import (
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"
)

func TestApps(t *testing.T) {
	apps := []App{}
	procfiles := []string{"./examples/fileserver/Procfile", "./examples/ping/Procfile"}
	for _, pf := range procfiles {
		fs, err := NewProcessApp(pf)
		if err != nil {
			t.Errorf("Failed to load procfile %s: %s", pf, err)
		}
		apps = append(apps, fs)
	}

	apps = append(apps, NewWebServerApp("./examples/static"))

	var wg sync.WaitGroup
	for _, app := range apps {
		wg.Add(1)

		if app.Running() {
			t.Errorf("app MUST not be started yet: %s", app.Name())
		}

		go func(a App) {
			defer func() { wg.Done() }()

			err := a.Start()
			if err != nil {
				t.Fatalf("Unable to start %s: %v", a.Name(), err)
			}

			if !a.Running() {
				t.Fatalf("app should be started now: %s", a.Name())
			}

			if a.Port() == 0 {
				t.Fatalf("Port should not be zeroed: %s", a.Name())
			}

			<-time.After(1 * time.Second) // wait for start

			req, _ := http.NewRequest("GET", fmt.Sprintf("http://localhost:%d", a.Port()), nil)
			res, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("Server unavailable %s: %v", a.Name(), err)
			}

			if res.StatusCode != 200 {
				t.Errorf("Unexpected status code for %s: %d", a.Name(), res.StatusCode)
			}

			a.Stop()
			if a.Running() {
				t.Fatalf("app should NOT be started now: %s", a.Name())
			}

		}(app)
	}

	wg.Wait()
}
