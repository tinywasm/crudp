package crudp

// Separate CRUD interfaces - handlers implement only what they need
// Return `any` which internally can be slice for multiple items
type Creator interface {
	Create(data ...any) any
}

type Reader interface {
	Read(data ...any) any
}

type Updater interface {
	Update(data ...any) any
}

type Deleter interface {
	Delete(data ...any) any
}

type NamedHandler interface {
	HandlerName() string
}

type DataValidator interface {
	ValidateData(action byte, data ...any) error
}

type AccessLevel interface {
	AllowedRoles(action byte) []byte
}
