package crudp_test

import (
	"testing"

	"github.com/tinywasm/crudp"
	. "github.com/tinywasm/fmt"
)

type User struct {
	ID    int
	Name  string
	Email string
}

func (u *User) Create(data ...any) any {
	created := make([]*User, 0, len(data))
	for _, item := range data {
		user, ok := item.(*User)
		if ok {
			user.ID = 123
			created = append(created, user)
		}
	}
	return created
}

func (u *User) Read(data ...any) any {
	results := make([]*User, 0, len(data))
	for _, item := range data {
		user, ok := item.(*User)
		if ok {
			results = append(results, &User{ID: user.ID, Name: "Found " + user.Name, Email: user.Email})
		}
	}
	return results
}

func CrudPBasicFunctionalityShared(t *testing.T) {
	// Initialize CRUDP with handlers
	cp := NewTestCrudP()
	if err := cp.RegisterHandlers(&User{}); err != nil {
		t.Fatalf("Failed to load handlers: %v", err)
	}

	// Test Create operation
	userData, err := testEncode(&User{Name: "John", Email: "john@example.com"})
	if err != nil {
		t.Fatalf("Failed to encode user data: %v", err)
	}

	createPacket := crudp.Packet{
		Action:    'c',
		HandlerID: 0,
		ReqID:     "test-create",
		Data:      [][]byte{userData},
	}

	batchReq := &crudp.BatchRequest{Packets: []crudp.Packet{createPacket}}

	batchResp, err := cp.Execute(batchReq)
	if err != nil {
		t.Fatalf("Failed to execute batch: %v", err)
	}

	if len(batchResp.Results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(batchResp.Results))
	}

	result := batchResp.Results[0]
	if result.ReqID != "test-create" {
		t.Errorf("Expected ReqID 'test-create', got '%s'", result.ReqID)
	}

	if result.MessageType != uint8(Msg.Success) {
		t.Errorf("Expected success, got failure: %s", result.Message)
	}

	if len(result.Data) != 1 {
		t.Fatalf("Expected 1 data element, got %d", len(result.Data))
	}

	var createdUser User
	if err := testDecode(result.Data[0], &createdUser); err != nil {
		t.Fatalf("Failed to decode created user: %v", err)
	}
	if createdUser.ID != 123 {
		t.Errorf("Expected created user ID 123, got %d", createdUser.ID)
	}

	// Test Read operation
	readUserData, err := testEncode(&User{ID: 123, Name: "John"})
	if err != nil {
		t.Fatalf("Failed to encode read user data: %v", err)
	}

	readPacket := crudp.Packet{
		Action:    'r',
		HandlerID: 0,
		ReqID:     "test-read",
		Data:      [][]byte{readUserData},
	}

	batchReq2 := &crudp.BatchRequest{Packets: []crudp.Packet{readPacket}}

	batchResp2, err := cp.Execute(batchReq2)
	if err != nil {
		t.Fatalf("Failed to execute read batch: %v", err)
	}

	if len(batchResp2.Results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(batchResp2.Results))
	}

	result2 := batchResp2.Results[0]
	if result2.ReqID != "test-read" {
		t.Errorf("Expected ReqID 'test-read', got '%s'", result2.ReqID)
	}

	if result2.MessageType != uint8(Msg.Success) {
		t.Errorf("Expected success, got failure: %s", result2.Message)
	}

	if len(result2.Data) != 1 {
		t.Fatalf("Expected 1 data element, got %d", len(result2.Data))
	}

	var readUser User
	if err := testDecode(result2.Data[0], &readUser); err != nil {
		t.Fatalf("Failed to decode read user: %v", err)
	}
	if readUser.Name != "Found John" {
		t.Errorf("Expected read user name 'Found John', got '%s'", readUser.Name)
	}
}

func LoggerConfigShared(t *testing.T) {
	t.Run("Logger Disabled By Default", func(t *testing.T) {
		cp := NewTestCrudP()
		cp.SetLog(nil)
	})

	t.Run("SetLog Custom", func(t *testing.T) {
		cp := NewTestCrudP()

		var logged []any
		cp.SetLog(func(args ...any) {
			logged = append(logged, args...)
		})

		// Register handler to trigger log
		err := cp.RegisterHandlers(&testLogHandler{})
		if err != nil {
			t.Fatal(err)
		}

		if len(logged) == 0 {
			t.Error("expected log output")
		}
	})

	t.Run("SetLog Nil Restores NoOp", func(t *testing.T) {
		cp := NewTestCrudP()
		cp.SetLog(func(args ...any) {})
		cp.SetLog(nil)
	})
}

type testLogHandler struct{}

func (h *testLogHandler) Create(data ...any) any {
	return "ok"
}
