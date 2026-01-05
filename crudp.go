package crudp

import (
	"reflect"

	"github.com/tinywasm/binary"
)

type actionHandler struct {
	name         string
	index        uint8
	handler      any
	dataType     reflect.Type
	Create       func(data ...any) any
	Read         func(data ...any) any
	Update       func(data ...any) any
	Delete       func(data ...any) any
	ValidateData func(action byte, data ...any) error
	MinAccess    func(action byte) int
}

// AccessDeniedHandler defines the callback for failed access attempts
type AccessDeniedHandler func(handler string, action byte, userLevel int, minRequired int)

// CrudP handles automatic handler processing
type CrudP struct {
	encode              func(input any, output any) error
	decode              func(input any, output any) error
	handlers            []actionHandler
	log                 func(...any) // Never nil - uses no-op by default
	devMode             bool
	getUserLevel        func(data ...any) int
	accessDeniedHandler AccessDeniedHandler
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

// SetDevMode enables or disables development mode (bypasses access checks)
func (cp *CrudP) SetDevMode(enabled bool) {
	cp.devMode = enabled
}

// SetUserLevel configures the function to extract the current user's access level
func (cp *CrudP) SetUserLevel(fn func(data ...any) int) {
	cp.getUserLevel = fn
}

// SetAccessDeniedHandler configures a callback for failed access attempts
func (cp *CrudP) SetAccessDeniedHandler(fn AccessDeniedHandler) {
	cp.accessDeniedHandler = fn
}
