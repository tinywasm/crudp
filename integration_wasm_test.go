//go:build wasm

package crudp_test

import (
	"os"
	"testing"
)

func TestWASM_Integration(t *testing.T) {
	// Read server URL from file (written by test server on startup)
	urlBytes, err := os.ReadFile(".crudp_test_server_url")
	if err != nil {
		t.Skipf("Skipping WASM integration test: no server URL file. Run stdlib tests first.")
		return
	}
	serverURL := string(urlBytes)

	t.Run("AllCRUDOperations", func(t *testing.T) {
		IntegrationAllCRUDOperationsShared(t, serverURL)
	})

	t.Run("BatchAllOperations", func(t *testing.T) {
		IntegrationBatchAllOperationsShared(t, serverURL)
	})
}
