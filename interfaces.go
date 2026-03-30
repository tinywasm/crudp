package crudp

// Creator handles entity creation.
// payload is the entity to create (concrete type asserted internally by the handler).
// Returns the created entity or an error.
type Creator interface {
	Create(payload any) (any, error)
}

// Reader handles entity retrieval.
// Read returns a single entity by its string ID.
// List returns all entities (no filter).
type Reader interface {
	Read(id string) (any, error)
	List() (any, error)
}

// Updater handles entity mutation.
// payload is the entity with updated fields.
// Returns the updated entity or an error.
type Updater interface {
	Update(payload any) (any, error)
}

// Deleter handles entity removal by ID.
type Deleter interface {
	Delete(id string) error
}

// NamedHandler provides the resource name used for routing and RBAC.
type NamedHandler interface {
	HandlerName() string
}

// DataValidator validates payload before execution.
// action: 'c' create, 'r' read, 'u' update, 'd' delete.
type DataValidator interface {
	ValidateData(action byte, payload any) error
}

// AccessLevel declares which role codes are allowed per action.
// Used by standalone mode (without tinywasm/rbac).
type AccessLevel interface {
	AllowedRoles(action byte) []byte
}
