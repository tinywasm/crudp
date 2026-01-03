//go:build wasm

package main

import (
	"github.com/tinywasm/crudp/example/pkg/router"
)

func main() {
	// Initialize CRUDP router
	cp := router.NewRouter()

	// Connect fetch responses to CRUDP handlers
	cp.InitClient()

	// Keep WASM running
	select {}
}
