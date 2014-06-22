package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
)

func main() {
	port := flag.String("p", "9000", "Port to listen")
	dir := flag.String("d", ".", "Directory to serve")
	flag.Parse()

	http.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir(*dir))))
	address := fmt.Sprintf(":%s", *port)
	log.Printf("Serving directory %s at %s\n", *dir, address)
	log.Fatal(http.ListenAndServe(address, nil))
}
