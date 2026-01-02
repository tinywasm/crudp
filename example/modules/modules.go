package modules

import (
	"github.com/tinywasm/crudp/example/modules/patient"
	"github.com/tinywasm/crudp/example/modules/user"
)

// Init returns all business modules
func Init() []any {
	return []any{
		&user.Handler{},
		&patient.Handler{},
	}
}
