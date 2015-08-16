package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestCommandCenter(t *testing.T) {
	apps := []string{"fileserver", "ping", "static", "PyServer"}
	c := &Config{AppsDir: "./examples/", Tld: "app"}

	cc := NewCommandCenter("bam", c)

	if len(cc.apps) != len(apps) {
		t.Errorf("Application count: got: %d; expected: %d", len(cc.apps), len(apps))
	}

	for _, name := range apps {
		app, ok := cc.apps[strings.ToLower(name)]
		if !ok {
			t.Errorf("Application not loaded: %s", name)
		}

		if app.Running() {
			t.Errorf("Application should be stoped: %s", name)
		}
	}

	go cc.Start()

	<-time.After(1 * time.Second) // wait for command center to start

	res := request(t, "GET", "http://localhost:%d", cc.Port())
	verifyResponse(t, res, http.StatusOK, apps...)

	for _, app := range cc.apps {
		if app.Running() {
			t.Errorf("Application should be stopped: %s", app.Name())
		}
	}

	static := cc.apps["static"]
	res = request(t, "GET", "http://localhost:%d/apps/%s/start", cc.Port(), static.Name())
	verifyResponse(t, res, http.StatusOK)

	<-time.After(1 * time.Second) // wait for static server to start

	if !static.Running() {
		t.Error("Static server should be running")
	}

	res = request(t, "GET", "http://localhost:%d", static.Port())
	verifyResponse(t, res, http.StatusOK, "It works!")

	res = request(t, "GET", "http://localhost:%d/apps/%s/stop", cc.Port(), static.Name())
	verifyResponse(t, res, http.StatusOK)

	<-time.After(1 * time.Second) // wait for static server to stop

	if static.Running() {
		t.Error("Static server should not be running")
	}
}

func request(t *testing.T, method string, url string, args ...interface{}) *http.Response {
	req, err := http.NewRequest(method, fmt.Sprintf(url, args...), nil)
	if err != nil {
		t.Errorf("Invalid request %s %s", method, fmt.Sprintf(url, args...))
	}

	req.Close = true
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	return res
}

func verifyResponse(t *testing.T, r *http.Response, status int, contentParts ...string) {
	if r.StatusCode != status {
		t.Errorf("Status code: got %d; expected %d", r.StatusCode, status)
	}

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		t.Errorf("Unexpected response: %s", err)
	}

	content := string(bodyBytes)
	for _, part := range contentParts {
		if !strings.Contains(content, part) {
			t.Errorf("Response doesnt contains %s", part)
		}
	}
}
