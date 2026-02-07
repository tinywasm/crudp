package crudp

import "github.com/tinywasm/dom"

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

// NamedHandler provides the unique name for the entity (required if CRUD implemented)
type NamedHandler interface {
	dom.NamedHandler
}

// DataValidator validates complete data before action (required if CRUD implemented)
type DataValidator interface {
	dom.DataValidator
}

// AccessLevel defines role-based access control (required if CRUD implemented)
type AccessLevel interface {
	dom.AccessLevel
}
