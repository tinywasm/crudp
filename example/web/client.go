//go:build wasm

package main

import (
	"github.com/tinywasm/crudp"
	"github.com/tinywasm/crudp/example/modules"
)

func main() {
	// Initialize CRUDP directly
	cp := crudp.NewDefault()
	cp.RegisterHandlers(modules.Init()...)

	// Connect fetch responses to CRUDP handlers
	cp.InitClient()

	// Keep WASM running
	select {}
}
