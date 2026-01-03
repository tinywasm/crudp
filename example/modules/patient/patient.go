package patient

type Handler struct{}

type Patient struct {
	ID   int
	Name string
	Age  int
}

func (h *Handler) Create(data ...any) (any, error) {
	// Specific implementation for patients
	return nil, nil
}

func (h *Handler) Read(data ...any) (any, error) {
	// Specific implementation for patients
	return nil, nil
}
