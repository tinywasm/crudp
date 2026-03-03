package patient

type Patient struct {
	ID   int
	Name string
	Age  int
}

func (p *Patient) HandlerName() string { return "patients" }

func (p *Patient) Create(payload any) (any, error) {
	// Specific implementation for patients
	return nil, nil
}

func (p *Patient) Read(id string) (any, error) {
	// Specific implementation for patients
	return nil, nil
}

func (p *Patient) List() (any, error) {
	return nil, nil
}

func (p *Patient) ValidateData(action byte, payload any) error { return nil }
func (p *Patient) AllowedRoles(action byte) []byte             { return []byte{'*'} }

// Add returns all entities from this module
func Add() []any {
	return []any{&Patient{}}
}
