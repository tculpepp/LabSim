// Command serve is a tiny static file server for local testing of the WASM build.
package main

import (
	"flag"
	"log"
	"mime"
	"net/http"
)

func main() {
	dir := flag.String("dir", "dist", "directory to serve")
	addr := flag.String("addr", ":8080", "listen address")
	flag.Parse()
	_ = mime.AddExtensionType(".wasm", "application/wasm")
	_ = mime.AddExtensionType(".webmanifest", "application/manifest+json")
	log.Printf("LabSim: http://localhost%s", *addr)
	log.Fatal(http.ListenAndServe(*addr, http.FileServer(http.Dir(*dir))))
}
