package crudp_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/tinywasm/binary"
	"github.com/tinywasm/crudp"
	"github.com/tinywasm/fetch"
)

// SharedUser is used in shared integration tests
type SharedUser struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func (u *SharedUser) HandlerName() string { return "users" }

func (u *SharedUser) Create(data ...any) any {
	for _, item := range data {
		if user, ok := item.(*SharedUser); ok {
			user.ID = 999
			return user
		}
	}
	return nil
}

func (u *SharedUser) Read(data ...any) any {
	// Mock database
	users := []*SharedUser{
		{ID: 1, Name: "Alice"},
		{ID: 2, Name: "Bob"},
		{ID: 3, Name: "Charlie"},
	}

	for _, item := range data {
		if path, ok := item.(string); ok {
			if path == "" {
				// No ID - return all users
				return users
			}
			// ID provided - find user by ID
			for _, u := range users {
				if fmt.Sprintf("%d", u.ID) == path {
					return u
				}
			}
			return nil // Not found
		}
	}
	return nil
}

func (u *SharedUser) Update(data ...any) any {
	for _, item := range data {
		if user, ok := item.(*SharedUser); ok {
			user.Name = "Updated: " + user.Name
			return user
		}
	}
	return nil
}

func (u *SharedUser) Delete(data ...any) any {
	for _, item := range data {
		if path, ok := item.(string); ok {
			return "Deleted: " + path
		}
	}
	return "deleted"
}

func (u *SharedUser) ValidateData(action byte, data ...any) error { return nil }
func (u *SharedUser) MinAccess(action byte) int                   { return 0 }

// Test all 4 CRUD operations via automatic endpoints
func IntegrationAllCRUDOperationsShared(t *testing.T, serverURL string) {
	t.Run("Create", func(t *testing.T) {
		testCRUDOperation(t, serverURL, "POST", "/users/", &SharedUser{Name: "New"})
	})

	t.Run("ReadOne", func(t *testing.T) {
		testCRUDOperation(t, serverURL, "GET", "/users/2", nil) // Returns Bob
	})

	t.Run("ReadAll", func(t *testing.T) {
		testCRUDOperation(t, serverURL, "GET", "/users/", nil)
	})

	t.Run("Update", func(t *testing.T) {
		testCRUDOperation(t, serverURL, "PUT", "/users/42", &SharedUser{ID: 42, Name: "Updated"})
	})

	t.Run("Delete", func(t *testing.T) {
		testCRUDOperation(t, serverURL, "DELETE", "/users/42", nil)
	})
}

func testCRUDOperation(t *testing.T, serverURL, method, path string, payload *SharedUser) {
	received := make(chan *crudp.Response, 1)

	fetch.SetHandler(func(resp *fetch.Response) {
		var r crudp.Response
		if err := binary.Decode(resp.Body(), &r); err != nil {
			t.Errorf("Failed to decode: %v", err)
			received <- nil
			return
		}
		received <- &r
	})

	// Build request based on method
	var req *fetch.Request
	switch method {
	case "POST":
		req = fetch.Post(serverURL + path)
	case "GET":
		req = fetch.Get(serverURL + path)
	case "PUT":
		req = fetch.Put(serverURL + path)
	case "DELETE":
		req = fetch.Delete(serverURL + path)
	}

	// Add body if payload provided
	if payload != nil {
		var userData []byte
		binary.Encode(payload, &userData)
		crudpReq := crudp.Request{
			ReqID: "test-" + method,
			Data:  [][]byte{userData},
		}
		var body []byte
		binary.Encode(crudpReq, &body)
		req.ContentTypeBinary().Body(body)
	}

	req.Dispatch()

	select {
	case resp := <-received:
		if resp == nil {
			t.Fatalf("%s %s: nil response", method, path)
		}
		if resp.MessageType != 4 { // Msg.Success
			t.Errorf("%s %s: expected success, got: %s", method, path, resp.Message)
		}
	case <-time.After(5 * time.Second):
		t.Fatalf("%s %s: timeout", method, path)
	}
}

// Test batch endpoint with all 4 operations
func IntegrationBatchAllOperationsShared(t *testing.T, serverURL string) {
	received := make(chan *crudp.BatchResponse, 1)

	fetch.SetHandler(func(resp *fetch.Response) {
		var batchResp crudp.BatchResponse
		if err := binary.Decode(resp.Body(), &batchResp); err != nil {
			t.Errorf("Failed to decode: %v", err)
			received <- nil
			return
		}
		received <- &batchResp
	})

	// Create batch with all 4 operations
	var createData, updateData []byte
	binary.Encode(&SharedUser{Name: "Create"}, &createData)
	binary.Encode(&SharedUser{ID: 1, Name: "Update"}, &updateData)

	batchReq := crudp.BatchRequest{
		Packets: []crudp.Packet{
			{Action: 'c', HandlerID: 0, ReqID: "batch-create", Data: [][]byte{createData}},
			{Action: 'r', HandlerID: 0, ReqID: "batch-read", Data: nil},
			{Action: 'u', HandlerID: 0, ReqID: "batch-update", Data: [][]byte{updateData}},
			{Action: 'd', HandlerID: 0, ReqID: "batch-delete", Data: nil},
		},
	}

	var body []byte
	binary.Encode(batchReq, &body)

	fetch.Post(serverURL + "/batch").
		ContentTypeBinary().
		Body(body).
		Dispatch()

	select {
	case resp := <-received:
		if resp == nil {
			t.Fatal("Nil batch response")
		}
		if len(resp.Results) != 4 {
			t.Fatalf("Expected 4 results, got %d", len(resp.Results))
		}
		for i, r := range resp.Results {
			if r.MessageType != 4 {
				t.Errorf("Result %d: expected success, got: %s", i, r.Message)
			}
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Timeout waiting for batch response")
	}
}
