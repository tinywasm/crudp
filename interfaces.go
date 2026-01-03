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

// NamedHandler allows override of automatic name (optional)
// If not implemented, reflection is used: TypeName -> snake_case
type NamedHandler interface {
	HandlerName() string
}

// Validator validates complete data before action (optional)
type Validator interface {
	Validate(action byte, data ...any) error
}

// FieldValidator validates individual fields for UI (optional)
type FieldValidator interface {
	ValidateField(fieldName string, value string) error
}
