package crudp

import (
	"context"
	"reflect"

	. "github.com/tinywasm/fmt"
)

// Execute processes a BatchRequest and returns a BatchResponse
func (cp *CrudP) Execute(ctx context.Context, req *BatchRequest) (*BatchResponse, error) {
	if req == nil {
		return nil, Errf("request is nil")
	}

	results := make([]PacketResult, 0, len(req.Packets))

	for _, p := range req.Packets {
		result := cp.executeSingle(ctx, &p)
		results = append(results, result)
	}

	return &BatchResponse{
		Results: results,
	}, nil
}

func (cp *CrudP) executeSingle(ctx context.Context, p *Packet) PacketResult {
	pr := PacketResult{
		Packet: *p,
	}

	// Decode data
	decodedData, err := cp.decodeWithKnownType(p, p.HandlerID)
	if err != nil {
		pr.MessageType = uint8(Msg.Error)
		pr.Message = err.Error()
		return pr
	}

	// Call handler
	result, err := cp.CallHandler(ctx, p.HandlerID, p.Action, decodedData...)
	if err != nil {
		pr.MessageType = uint8(Msg.Error)
		pr.Message = err.Error()
		return pr
	}

	// Encode result to Data
	if err := cp.encodeResult(&pr, result); err != nil {
		pr.MessageType = uint8(Msg.Error)
		pr.Message = err.Error()
		return pr
	}

	pr.MessageType = uint8(Msg.Success)
	pr.Message = "OK"
	return pr
}

func (cp *CrudP) encodeResult(pr *PacketResult, result any) error {
	if result == nil {
		return nil
	}

	if cp.encode == nil {
		return Errf("encode function not configured")
	}

	// Handle slices for multiple items
	v := reflect.ValueOf(result)
	if v.Kind() == reflect.Slice {
		pr.Data = make([][]byte, 0, v.Len())
		for i := 0; i < v.Len(); i++ {
			var encoded []byte
			if err := cp.encode(v.Index(i).Interface(), &encoded); err != nil {
				return err
			}
			pr.Data = append(pr.Data, encoded)
		}
		return nil
	}

	// Single item
	var encoded []byte
	if err := cp.encode(result, &encoded); err != nil {
		return err
	}
	pr.Data = [][]byte{encoded}
	return nil
}
