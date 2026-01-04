//go:build !wasm

package main

import (
	"log"
	"net/http"
	"os"

	"github.com/tinywasm/crudp"
	"github.com/tinywasm/crudp/example/modules"
)

func main() {
	publicDir := "public"

	// Debug: Print working directory and check if public exists
	if _, err := os.Stat(publicDir); os.IsNotExist(err) {
		log.Printf("WARNING: Public directory '%s' does not exist!", publicDir)
	}

	// Serve static files with no-cache headers
	fs := http.FileServer(http.Dir(publicDir))
	noCache := func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
			w.Header().Set("Pragma", "no-cache")
			w.Header().Set("Expires", "0")
			h.ServeHTTP(w, r)
		})
	}

	mux := http.NewServeMux()
	mux.Handle("/", noCache(fs))

	// Initialize CRUDP directly
	cp := crudp.New()
	cp.RegisterHandlers(modules.Init()...)
	cp.RegisterRoutes(mux)

	log.Printf("Server starting on http://localhost:6060")
	if err := http.ListenAndServe(":6060", mux); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
