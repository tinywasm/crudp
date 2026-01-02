//go:build wasm

package crudp_test

import (
	"testing"
)

func TestCrudP_WASM(t *testing.T) {
	t.Run("Basic", func(t *testing.T) {
		CrudPBasicFunctionalityShared(t)
	})

	t.Run("Logger", func(t *testing.T) {
		LoggerConfigShared(t)
	})
}
