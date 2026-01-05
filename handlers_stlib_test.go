//go:build !wasm

package crudp_test

import (
	"testing"
)

func TestHandlers_Stdlib(t *testing.T) {
	cp := NewTestCrudP()

	t.Run("Registration", func(t *testing.T) {
		HandlerRegistrationShared(t, cp)
		HandlerRegistrationErrorsShared(t, cp)
		ModuleAddPatternShared(t, cp)
	})

	t.Run("Validation", func(t *testing.T) {
		HandlerValidationShared(t, cp)
		AccessControlShared(t, cp)
	})

	t.Run("CRUD", func(t *testing.T) {
		CRUDOperationsShared(t, cp)
	})
}
