package crudp_test

import (
	"encoding/json"

	"github.com/tinywasm/binary"
	"github.com/tinywasm/crudp"
)

func NewTestCrudP() *crudp.CrudP {
	cp := crudp.New()
	cp.SetDevMode(true)
	return cp
}

func NewTestCrudPJSON() *crudp.CrudP {
	cp := crudp.New()
	cp.SetDevMode(true)
	cp.SetCodecs(jsonEncode, jsonDecode)
	return cp
}

func testEncodeBinary(data any) ([]byte, error) {
	var out []byte
	err := binary.Encode(data, &out)
	return out, err
}

func testDecodeBinary(data []byte, target any) error {
	return binary.Decode(data, target)
}

func testEncodeJSON(data any) ([]byte, error) {
	var out []byte
	err := jsonEncode(data, &out)
	return out, err
}

func testDecodeJSON(data []byte, target any) error {
	return jsonDecode(data, target)
}

func jsonEncode(input any, output any) error {
	b, err := json.Marshal(input)
	if err != nil {
		return err
	}
	*(output.(*[]byte)) = b
	return nil
}

func jsonDecode(input any, output any) error {
	return json.Unmarshal(input.([]byte), output)
}
