package crudp

import (
	"github.com/tinywasm/binary"
)

type actionHandler struct {
	name    string
	index   uint8
	handler any
	Create  func(data ...any) any
	Read    func(data ...any) any
	Update  func(data ...any) any
	Delete  func(data ...any) any
}

// CrudP handles automatic handler processing
type CrudP struct {
	encode   func(input any, output any) error
	decode   func(input any, output any) error
	handlers []actionHandler
	log      func(...any) // Never nil - uses no-op by default
}

// New creates a new CrudP instance with binary codec by default
func New() *CrudP {
	cp := &CrudP{
		encode: binary.Encode,
		decode: binary.Decode,
		log:    func(...any) {},
	}

	return cp
}

// SetCodecs configures custom serialization functions
func (cp *CrudP) SetCodecs(encode, decode func(input any, output any) error) {
	cp.encode = encode
	cp.decode = decode
}

// SetLog configures a custom logging function
func (cp *CrudP) SetLog(log func(...any)) {
	if log == nil {
		cp.log = func(...any) {}
		return
	}
	cp.log = log
}
