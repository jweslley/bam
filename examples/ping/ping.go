package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
)

func root(w http.ResponseWriter, req *http.Request) {
	user := os.Getenv("OWNER")
	if user == "" {
		user = "world"
	}

	log.Printf("[ping] root: %s\n", user)
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, "Hello %s<br><a href='/ping'>Ping</a>", user)
}

func pong(w http.ResponseWriter, req *http.Request) {
	log.Println("[ping] pong")
	fmt.Fprint(w, "pong")
}

func main() {
	port := flag.String("p", "9000", "Port to listen")
	flag.Parse()

	http.HandleFunc("/", root)
	http.HandleFunc("/ping", pong)

	address := fmt.Sprintf(":%s", *port)
	log.Printf("Starting at %s\n", address)
	log.Fatal(http.ListenAndServe(address, nil))
}
