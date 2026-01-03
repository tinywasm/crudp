package crudp_test

import (
	"testing"

	"github.com/tinywasm/crudp"
	. "github.com/tinywasm/fmt"
)

func CrudPErrorHandlingShared(t *testing.T) {
	cp := NewTestCrudP()

	t.Run("Invalid Handler ID", func(t *testing.T) {
		req := &crudp.BatchRequest{
			Packets: []crudp.Packet{
				{Action: 'c', HandlerID: 99, ReqID: "err-1"},
			},
		}

		resp, err := cp.Execute(req)
		if err != nil {
			t.Fatalf("Execute failed: %v", err)
		}

		if resp.Results[0].MessageType != uint8(Msg.Error) {
			t.Errorf("Expected error message type, got %v", resp.Results[0].MessageType)
		}
	})
}
