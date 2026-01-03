package router

import (
	"github.com/tinywasm/crudp"
	"github.com/tinywasm/crudp/example/modules"
)

func NewRouter() *crudp.CrudP {
	cp := crudp.NewDefault()

	// Get handlers from modules
	handlers := modules.Init()

	// Register handlers in CRUDP
	cp.RegisterHandler(handlers...)

	return cp
}
