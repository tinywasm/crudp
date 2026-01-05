//go:build !wasm

package crudp_test

import (
	"net"
	"net/http"
	"os"
	"sync"

	"github.com/tinywasm/binary"
	"github.com/tinywasm/crudp"
)

var (
	integrationServerURL  string
	integrationServerOnce sync.Once
)

// SetupIntegrationServer creates a CRUDP server for integration testing.
// Returns the server URL.
func SetupIntegrationServer() string {
	integrationServerOnce.Do(func() {
		cp := NewTestCrudP()
		cp.RegisterHandlers(&SharedUser{})

		mux := http.NewServeMux()
		cp.RegisterRoutes(mux)

		// Add shutdown handler
		shutdown := make(chan struct{})
		mux.HandleFunc("/shutdown", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			close(shutdown)
		})

		// Start server on random port
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			panic(err)
		}

		integrationServerURL = "http://" + listener.Addr().String()

		// Write URL to file for WASM test to discover
		os.WriteFile(".crudp_test_server_url", []byte(integrationServerURL), 0644)

		go func() {
			srv := &http.Server{Handler: mux}
			go srv.Serve(listener)
			<-shutdown
			srv.Close()
		}()
	})

	return integrationServerURL
}

// CreateIntegrationTestPayload creates a binary-encoded batch request
func CreateIntegrationTestPayload() []byte {
	var userData []byte
	binary.Encode(&SharedUser{Name: "WASM Test"}, &userData)

	req := crudp.BatchRequest{
		Packets: []crudp.Packet{
			{Action: 'c', HandlerID: 0, ReqID: "wasm-1", Data: [][]byte{userData}},
		},
	}

	var body []byte
	binary.Encode(req, &body)
	return body
}
