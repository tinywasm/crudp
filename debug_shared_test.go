package crudp_test

import (
	"testing"

	"github.com/tinywasm/crudp"
)

func HandlerInstanceReuseShared(t *testing.T) {
	cp := NewTestCrudP()
	if err := cp.RegisterHandler(&User{}); err != nil {
		t.Fatalf("Failed to load handlers: %v", err)
	}

	for i := 0; i < 2; i++ {
		name := "User " + string(rune('A'+i))
		userData, _ := testEncode(&User{Name: name})

		req := &crudp.BatchRequest{
			Packets: []crudp.Packet{
				{Action: 'c', HandlerID: 0, Data: [][]byte{userData}},
			},
		}

		resp, err := cp.Execute(req)
		if err != nil {
			t.Fatalf("Execute failed: %v", err)
		}

		var result User
		testDecode(resp.Results[0].Data[0], &result)

		if result.Name != name {
			t.Errorf("Iteration %d: expected name %s, got %s", i, name, result.Name)
		}
	}
}

func ConcurrentHandlerAccessShared(t *testing.T) {
	cp := NewTestCrudP()
	if err := cp.RegisterHandler(&User{}); err != nil {
		t.Fatalf("Failed to load handlers: %v", err)
	}

	names := []string{"Alice", "Bob", "Charlie", "Dave"}

	for _, name := range names {
		userData, _ := testEncode(&User{Name: name})
		req := &crudp.BatchRequest{
			Packets: []crudp.Packet{
				{Action: 'c', HandlerID: 0, Data: [][]byte{userData}},
			},
		}

		resp, err := cp.Execute(req)
		if err != nil {
			t.Fatalf("Execute failed for %s: %v", name, err)
		}

		var result User
		testDecode(resp.Results[0].Data[0], &result)

		if result.Name != name {
			t.Errorf("Expected %s, got %s", name, result.Name)
		}
	}
}
