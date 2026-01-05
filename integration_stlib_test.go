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
