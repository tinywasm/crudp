//go:build !wasm

package crudp_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tinywasm/binary"
	"github.com/tinywasm/crudp"
)

// Integration test using the same pattern as example code

type IntegrationUser struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func (u *IntegrationUser) HandlerName() string { return "users" }

func (u *IntegrationUser) Create(data ...any) any {
	for _, item := range data {
		switch v := item.(type) {
		case *IntegrationUser:
			v.ID = 999
			return v
		}
	}
	return nil
}

func (u *IntegrationUser) Read(data ...any) any {
	for _, item := range data {
		switch v := item.(type) {
		case string:
			// Path parameter (e.g., "123" from /users/123)
			return &IntegrationUser{ID: 123, Name: "User from path: " + v}
		case *IntegrationUser:
			return &IntegrationUser{ID: v.ID, Name: "Found: " + v.Name}
		}
	}
	return nil
}

func (u *IntegrationUser) ValidateData(action byte, data ...any) error { return nil }
func (u *IntegrationUser) AllowedRoles(action byte) []byte             { return []byte{'*'} }

func TestIntegration_New(t *testing.T) {
	// Test New uses binary codec by default
	cp := NewTestCrudP()
	if cp == nil {
		t.Fatal("New returned nil")
	}

	err := cp.RegisterHandlers(&IntegrationUser{})
	if err != nil {
		t.Fatalf("RegisterHandlers failed: %v", err)
	}

	// Verify handler name
	if name := cp.GetHandlerName(0); name != "users" {
		t.Errorf("expected handler name 'users', got '%s'", name)
	}
}

func TestIntegration_AutomaticEndpoints(t *testing.T) {
	cp := NewTestCrudP()
	cp.RegisterHandlers(&IntegrationUser{})

	mux := http.NewServeMux()
	cp.RegisterRoutes(mux)

	t.Run("GET /users/{id}", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/users/42", nil)
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}

		// Decode response
		var resp crudp.Response
		if err := binary.Decode(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp.MessageType != 4 { // Msg.Success
			t.Errorf("expected success, got message: %s", resp.Message)
		}
	})

	t.Run("POST /users (empty body)", func(t *testing.T) {
		// Empty body - handler receives http.Request for multipart etc.
		req := httptest.NewRequest("POST", "/users/", httpBodyFromBytes([]byte{}))
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("POST /batch", func(t *testing.T) {
		var userData []byte
		binary.Encode(&IntegrationUser{Name: "Batch"}, &userData)

		batchReq := crudp.BatchRequest{
			Packets: []crudp.Packet{
				{Action: 'c', HandlerID: 0, ReqID: "batch-1", Data: [][]byte{userData}},
			},
		}

		var body []byte
		binary.Encode(batchReq, &body)

		req := httptest.NewRequest("POST", "/batch", httpBodyFromBytes(body))
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}

		var resp crudp.BatchResponse
		if err := binary.Decode(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode batch response: %v", err)
		}

		if len(resp.Results) != 1 {
			t.Fatalf("expected 1 result, got %d", len(resp.Results))
		}

		if resp.Results[0].MessageType != 4 { // Msg.Success
			t.Errorf("expected success, got: %s", resp.Results[0].Message)
		}
	})
}

type bytesBody struct {
	data []byte
	pos  int
}

func (b *bytesBody) Read(p []byte) (n int, err error) {
	if b.pos >= len(b.data) {
		return 0, io.EOF
	}
	n = copy(p, b.data[b.pos:])
	b.pos += n
	return n, nil
}

func (b *bytesBody) Close() error { return nil }

func httpBodyFromBytes(data []byte) *bytesBody {
	return &bytesBody{data: data}
}

// RestrictedResource requires level 100 for read access
type RestrictedResource struct{}

func (r *RestrictedResource) HandlerName() string                         { return "restricted" }
func (r *RestrictedResource) Read(data ...any) any                        { return "secret data" }
func (r *RestrictedResource) ValidateData(action byte, data ...any) error { return nil }
func (r *RestrictedResource) AllowedRoles(action byte) []byte             { return []byte{'a'} }

type PartialRolesHandler struct{ RestrictedResource }

func (p *PartialRolesHandler) Create(data ...any) any { return nil }
func (p *PartialRolesHandler) AllowedRoles(action byte) []byte {
	if action == 'r' {
		return []byte{'*'}
	}
	return nil // Missing 'c'
}

type InvalidHandler struct{ RestrictedResource }

func (i *InvalidHandler) AllowedRoles(action byte) []byte { return nil }

type EmptyRolesHandler struct{ RestrictedResource }

func (e *EmptyRolesHandler) AllowedRoles(action byte) []byte { return []byte{} }

type MultiRoleResource struct{ RestrictedResource }

func (m *MultiRoleResource) AllowedRoles(action byte) []byte {
	return []byte{'d', 'm'} // Dentist or Medic
}

func TestIntegration_AccessControl(t *testing.T) {
	t.Run("Access Granted", func(t *testing.T) {
		cp := crudp.New()
		cp.SetUserRoles(func(data ...any) []byte { return []byte{'a'} })
		cp.RegisterHandlers(&RestrictedResource{})

		mux := http.NewServeMux()
		cp.RegisterRoutes(mux)

		req := httptest.NewRequest("GET", "/restricted/", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("Access Denied", func(t *testing.T) {
		cp := crudp.New()
		cp.SetUserRoles(func(data ...any) []byte { return []byte{'v'} }) // Role 'v' != 'a'
		cp.RegisterHandlers(&RestrictedResource{})

		mux := http.NewServeMux()
		cp.RegisterRoutes(mux)

		req := httptest.NewRequest("GET", "/restricted/", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		// Should get 200 with error message in response body (CRUDP protocol)
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}

		var resp crudp.Response
		if err := binary.Decode(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp.MessageType != 2 { // Msg.Error
			t.Errorf("expected error message type, got %d: %s", resp.MessageType, resp.Message)
		}
	})

	t.Run("Access Denied Callback", func(t *testing.T) {
		cp := crudp.New()
		cp.SetUserRoles(func(data ...any) []byte { return []byte{'v'} })

		notified := false
		cp.SetAccessDeniedHandler(func(handler string, action byte, userRoles []byte, allowedRoles []byte, errMsg string) {
			notified = true
			if handler != "restricted" {
				t.Errorf("unexpected handler: %s", handler)
			}
			if string(userRoles) != "v" {
				t.Errorf("unexpected user roles: %s", userRoles)
			}
			if string(allowedRoles) != "a" {
				t.Errorf("unexpected allowed roles: %s", allowedRoles)
			}
		})

		cp.RegisterHandlers(&RestrictedResource{})

		mux := http.NewServeMux()
		cp.RegisterRoutes(mux)

		req := httptest.NewRequest("GET", "/restricted/", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if !notified {
			t.Error("AccessDeniedHandler was not called")
		}
	})

	t.Run("Security-by-Default (Empty Slice)", func(t *testing.T) {
		cp := crudp.New()
		cp.SetUserRoles(func(data ...any) []byte { return []byte{'*'} })
		err := cp.RegisterHandlers(&EmptyRolesHandler{})
		if err == nil {
			t.Error("expected error when registering handler with empty AllowedRoles, got nil")
		}
	})

	t.Run("OR Logic Match", func(t *testing.T) {
		cp := crudp.New()
		cp.SetUserRoles(func(data ...any) []byte { return []byte{'m', 'r'} }) // Medic and Reception
		cp.RegisterHandlers(&MultiRoleResource{})                             // Resource allows Dentist or Medic

		mux := http.NewServeMux()
		cp.RegisterRoutes(mux)

		req := httptest.NewRequest("GET", "/restricted/", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("Special '*' Role Access", func(t *testing.T) {
		cp := crudp.New()
		cp.SetUserRoles(func(data ...any) []byte { return []byte("any") })
		cp.RegisterHandlers(&IntegrationUser{}) // Uses '*'

		mux := http.NewServeMux()
		cp.RegisterRoutes(mux)

		req := httptest.NewRequest("GET", "/users/1", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("Unauthenticated Denied on '*'", func(t *testing.T) {
		cp := crudp.New()
		cp.SetUserRoles(func(data ...any) []byte { return nil }) // Unauthenticated
		cp.RegisterHandlers(&IntegrationUser{})                  // Uses '*'

		mux := http.NewServeMux()
		cp.RegisterRoutes(mux)

		req := httptest.NewRequest("GET", "/users/1", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		var resp crudp.Response
		binary.Decode(rec.Body.Bytes(), &resp)
		if resp.MessageType != 2 { // Msg.Error
			t.Errorf("expected access denied for unauthenticated user on '*' resource")
		}
	})

	t.Run("DevMode Bypass", func(t *testing.T) {
		cp := crudp.New()
		cp.SetDevMode(true)
		cp.RegisterHandlers(&RestrictedResource{}) // Requires 'a'

		mux := http.NewServeMux()
		cp.RegisterRoutes(mux)

		// No SetUserRoles called, should work in DevMode
		req := httptest.NewRequest("GET", "/restricted/", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("DevMode should bypass all access checks")
		}
	})

	t.Run("Missing SetUserRoles Error", func(t *testing.T) {
		cp := crudp.New()
		// No cp.SetUserRoles()
		err := cp.RegisterHandlers(&RestrictedResource{})
		if err == nil {
			t.Error("expected error when registering CRUD handlers without SetUserRoles()")
		}
	})
	t.Run("Security-by-Default (Partial Config)", func(t *testing.T) {
		cp := crudp.New()
		cp.SetUserRoles(func(data ...any) []byte { return []byte{'*'} })
		err := cp.RegisterHandlers(&PartialRolesHandler{})
		if err == nil {
			t.Error("expected error when registering handler with partial AllowedRoles configuration (missing 'c')")
		}
	})
}
