package modules

import (
	"github.com/tinywasm/crudp/example/modules/patient"
	"github.com/tinywasm/crudp/example/modules/user"
)

// Init collects all entities from all modules
func Init() []any {
	return append(
		user.Add(),
		patient.Add()...,
	)
}
