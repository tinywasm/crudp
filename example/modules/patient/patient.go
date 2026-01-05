package patient

type Patient struct {
	ID   int
	Name string
	Age  int
}

func (p *Patient) HandlerName() string { return "patients" }

func (p *Patient) Create(data ...any) any {
	// Specific implementation for patients
	return nil
}

func (p *Patient) Read(data ...any) any {
	// Specific implementation for patients
	return nil
}

func (p *Patient) ValidateData(action byte, data ...any) error { return nil }

// Add returns all entities from this module
func Add() []any {
	return []any{&Patient{}}
}
