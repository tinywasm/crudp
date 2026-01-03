package crudp

type actionHandler struct {
	name    string
	index   uint8
	handler any
	Create  func(data ...any) (any, error)
	Read    func(data ...any) (any, error)
	Update  func(data ...any) (any, error)
	Delete  func(data ...any) (any, error)
}

// CrudP handles automatic handler processing
type CrudP struct {
	encode   func(input any, output any) error
	decode   func(input any, output any) error
	handlers []actionHandler
	log      func(...any) // Never nil - uses no-op by default
}

// New creates a new CrudP instance with mandatory serialization functions
func New(encode, decode func(any, any) error) *CrudP {
	cp := &CrudP{
		encode: encode,
		decode: decode,
		log:    func(...any) {},
	}

	return cp
}

// SetLog configures a custom logging function
func (cp *CrudP) SetLog(log func(...any)) {
	if log == nil {
		cp.log = func(...any) {}
		return
	}
	cp.log = log
}
