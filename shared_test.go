package crudp_test

import (
	"github.com/tinywasm/binary"
	"github.com/tinywasm/crudp"
)

func NewTestCrudP() *crudp.CrudP {
	return crudp.New(binary.Encode, binary.Decode)
}

func testEncode(cp *crudp.CrudP, data any) ([]byte, error) {
	var out []byte
	err := binary.Encode(data, &out)
	return out, err
}

func testDecode(cp *crudp.CrudP, data []byte, target any) error {
	return binary.Decode(data, target)
}
