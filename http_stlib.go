//go:build !wasm

package crudp

import (
	"io"
	"net/http"

	"github.com/tinywasm/context"
	. "github.com/tinywasm/fmt"
)

// MiddlewareProvider allows handlers to provide global HTTP middleware
type MiddlewareProvider interface {
	Middleware(next http.Handler) http.Handler
}

func (cp *CrudP) RegisterRoutes(mux *http.ServeMux) {
	// Enable access check when routes are registered
	cp.accessCheck = func(h actionHandler, a byte, d ...any) error {
		return cp.doAccessCheck(h, a, d...)
	}

	// 1. Register global batch endpoint
	mux.HandleFunc("POST /batch", cp.handleBatch)

	// 2. Generate automatic routes for each handler
	for _, h := range cp.handlers {

		if h.Create != nil {
			mux.HandleFunc("POST /"+h.name+"/{path...}", cp.makeHandler(h, 'c'))
		}
		if h.Read != nil {
			mux.HandleFunc("GET /"+h.name+"/{path...}", cp.makeHandler(h, 'r'))
		}
		if h.Update != nil {
			mux.HandleFunc("PUT /"+h.name+"/{path...}", cp.makeHandler(h, 'u'))
		}
		if h.Delete != nil {
			mux.HandleFunc("DELETE /"+h.name+"/{path...}", cp.makeHandler(h, 'd'))
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
	ctx := context.Background()
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

func (cp *CrudP) makeHandler(h actionHandler, action byte) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := r.PathValue("path")
		cp.handleSingle(w, r, h, action, path)
	}
}

func (cp *CrudP) handleSingle(w http.ResponseWriter, r *http.Request, h actionHandler, action byte, path string) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading body", http.StatusBadRequest)
		return
	}

	var req Request
	if len(body) > 0 {
		if cp.decode == nil {
			http.Error(w, "decode function not configured", http.StatusInternalServerError)
			return
		}
		if err := cp.decode(body, &req); err != nil {
			http.Error(w, "Error decoding request", http.StatusBadRequest)
			return
		}
	}

	// Prepare data for handler
	decodedData, err := cp.decodeWithKnownType(&Packet{Data: req.Data}, h.index)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Prepend path (as string) and other injectables (context, request)
	ctx := context.Background()
	inject := []any{ctx, r}
	if path != "" {
		inject = append(inject, path)
	}
	allData := append(inject, decodedData...)

	// Call handler directly via CallHandler (which handles the error detection logic we added)
	result, err := cp.CallHandler(h.index, action, allData...)

	resp := Response{
		ReqID: req.ReqID,
	}

	if err != nil {
		resp.MessageType = uint8(Msg.Error)
		resp.Message = err.Error()
	} else {
		resp.MessageType = uint8(Msg.Success)
		resp.Message = "OK"

		// Encode results
		if result != nil {
			pr := PacketResult{}
			if err := cp.encodeResult(&pr, result); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			resp.Data = pr.Data
		}
	}

	if cp.encode == nil {
		http.Error(w, "encode function not configured", http.StatusInternalServerError)
		return
	}

	encoded, err := cp.encodeBody(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(encoded)
}

func (cp *CrudP) encodeBody(data any) ([]byte, error) {
	var encoded []byte
	if err := cp.encode(data, &encoded); err != nil {
		return nil, err
	}
	return encoded, nil
}

// doAccessCheck performs the actual access validation (server-side only)
func (cp *CrudP) doAccessCheck(handler actionHandler, action byte, data ...any) error {
	if cp.devMode || handler.AllowedRoles == nil {
		return nil
	}

	var userRoles []byte
	if cp.getUserRoles != nil {
		userRoles = cp.getUserRoles(data...)
	}

	allowedRoles := handler.AllowedRoles(action)
	if !hasAnyRole(userRoles, allowedRoles) {
		// Access denied
		errMsg := Fmt("required roles %q, user has %q", allowedRoles, userRoles)
		if cp.accessDeniedHandler != nil {
			cp.accessDeniedHandler(handler.name, action, userRoles, allowedRoles, errMsg)
		}
		cp.log("access denied for handler:", handler.name)
		return Errf("access denied")
	}

	return nil
}

// hasAnyRole checks if user has at least one of the allowed roles (OR logic)
// Special case: '*' means any authenticated user (any non-empty userRoles)
func hasAnyRole(userRoles, allowedRoles []byte) bool {
	if len(allowedRoles) == 0 {
		return false // Security-by-default (already checked in registration, but just in case)
	}

	for _, allowed := range allowedRoles {
		if allowed == '*' && len(userRoles) > 0 {
			return true
		}
		for _, user := range userRoles {
			if allowed == user {
				return true
			}
		}
	}
	return false
}
