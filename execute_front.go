//go:build wasm

package crudp

// HandleResponse processes a BatchResponse by converting it back to a BatchRequest
// and executing it locally on the WASM side.
func (cp *CrudP) HandleResponse(resp *BatchResponse) {
	if resp == nil {
		return
	}

	req := &BatchRequest{
		Packets: make([]Packet, 0, len(resp.Results)),
	}

	for _, res := range resp.Results {
		req.Packets = append(req.Packets, res.Packet)
	}

	// In WASM, we don't usually care about the return value of Execute
	// as handlers update the DOM directly via tinywasm/dom
	_, _ = cp.Execute(req)
}
