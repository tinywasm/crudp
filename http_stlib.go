//go:build !wasm

package crudp

import (
	"io"
	"net/http"

	"github.com/tinywasm/context"
)

// HttpRouteProvider allows handlers to register custom HTTP routes
type HttpRouteProvider interface {
	RegisterRoutes(mux *http.ServeMux)
}

// MiddlewareProvider allows handlers to provide global HTTP middleware
type MiddlewareProvider interface {
	Middleware(next http.Handler) http.Handler
}

// RegisterRoutes registers the default CRUD API endpoint and any custom routes from handlers
func (cp *CrudP) RegisterRoutes(mux *http.ServeMux) {
	// 1. Register default API endpoint
	mux.HandleFunc("/api", cp.handleBatch)

	// 2. Let handlers register their custom HTTP routes
	for _, h := range cp.handlers {
		if routeProvider, ok := h.handler.(HttpRouteProvider); ok {
			routeProvider.RegisterRoutes(mux)
		}
	}
}

// ApplyMiddleware collects all middleware from handlers and wraps the provided handler
func (cp *CrudP) ApplyMiddleware(handler http.Handler) http.Handler {
	for _, h := range cp.handlers {
		if mwProvider, ok := h.handler.(MiddlewareProvider); ok {
			handler = mwProvider.Middleware(handler)
		}
	}
	return handler
}

func (cp *CrudP) handleBatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading body", http.StatusBadRequest)
		return
	}

	var req BatchRequest
	if cp.decode == nil {
		http.Error(w, "decode function not configured", http.StatusInternalServerError)
		return
	}

	if err := cp.decode(body, &req); err != nil {
		// Create a minimal error response if decoding fails
		http.Error(w, "Error decoding request", http.StatusBadRequest)
		return
	}

	// Inject context and http.Request for handlers
	ctx := context.TODO()
	resp, err := cp.Execute(&req, ctx, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if cp.encode == nil {
		http.Error(w, "encode function not configured", http.StatusInternalServerError)
		return
	}

	var encoded []byte
	if err := cp.encode(resp, &encoded); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}

	// NOTE: Content-Type depends on codec, but defaulting to JSON is common
	w.Header().Set("Content-Type", "application/json")
	w.Write(encoded)
}
