// Command webserve serves the web build locally for testing in a browser.
//
// Build the bundle first with build_web.ps1 at the repo root, then from src/:
//
//	go run ./cmd/webserve
package main

import (
	"flag"
	"log"
	"net/http"
)

func main() {
	dir := flag.String("dir", "../dist/web", "directory holding index.html, wasm_exec.js and dungeoneer.wasm")
	addr := flag.String("addr", "localhost:8080", "address to listen on")
	flag.Parse()

	log.Printf("serving %s at http://%s", *dir, *addr)
	log.Fatal(http.ListenAndServe(*addr, http.FileServer(http.Dir(*dir))))
}
