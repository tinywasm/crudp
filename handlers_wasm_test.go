//go:build wasm

package crudp_test

import (
	"testing"
)

func TestHandlers_WASM(t *testing.T) {
	cp := NewTestCrudP()

	t.Run("Registration", func(t *testing.T) {
		HandlerRegistrationShared(t, cp)
	})

	t.Run("Validation", func(t *testing.T) {
		HandlerValidationShared(t, cp)
	})

	t.Run("CRUD", func(t *testing.T) {
		CRUDOperationsShared(t, cp)
	})
}
