//go:build wasm

package crudp

import (
	"github.com/tinywasm/fetch"
)

// InitClient configures the global fetch handler to route responses
// back into the CrudP instance.
func (cp *CrudP) InitClient() {
	fetch.SetHandler(func(resp *fetch.Response) {
		var batchResp BatchResponse
		if cp.decode == nil {
			cp.log("decode function not configured")
			return
		}

		if err := cp.decode(resp.Body(), &batchResp); err != nil {
			cp.log("error decoding batch response:", err)
			return
		}

		cp.HandleResponse(&batchResp)
	})
}
