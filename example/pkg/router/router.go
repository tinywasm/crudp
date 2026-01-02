package router

import (
	"github.com/tinywasm/binary"
	"github.com/tinywasm/crudp"
	"github.com/tinywasm/crudp/example/modules"
)

func NewRouter() *crudp.CrudP {
	cp := crudp.New(binary.Encode, binary.Decode)

	// Get handlers from modules
	handlers := modules.Init()

	// Register handlers in CRUDP
	cp.RegisterHandler(handlers...)

	return cp
}
