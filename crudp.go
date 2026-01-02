package crudp

import "context"

type actionHandler struct {
	name    string
	index   uint8
	handler any
	Create  func(ctx context.Context, data ...any) (any, error)
	Read    func(ctx context.Context, data ...any) (any, error)
	Update  func(ctx context.Context, data ...any) (any, error)
	Delete  func(ctx context.Context, data ...any) (any, error)
}

type EncodeFunc func(input any, output any) error
type DecodeFunc func(input any, output any) error

// CrudP handles automatic handler processing
type CrudP struct {
	encode   EncodeFunc
	decode   DecodeFunc
	handlers []actionHandler
	log      func(...any) // Never nil - uses no-op by default
}

// noopLogger is the default logger that does nothing
func noopLogger(...any) {}

// New creates a new CrudP instance with mandatory serialization functions
func New(encode EncodeFunc, decode DecodeFunc) *CrudP {
	cp := &CrudP{
		encode: encode,
		decode: decode,
		log:    noopLogger,
	}

	return cp
}

// SetLog configures a custom logging function
func (cp *CrudP) SetLog(log func(...any)) {
	if log == nil {
		cp.log = noopLogger
		return
	}
	cp.log = log
}
