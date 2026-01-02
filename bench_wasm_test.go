//go:build wasm

package crudp

import (
	"testing"
)

func BenchmarkCrudP_Setup(b *testing.B) {
	BenchmarkCrudPSetupShared(b)
}

func BenchmarkCrudP_Execute(b *testing.B) {
	BenchmarkCrudPExecuteShared(b)
}
