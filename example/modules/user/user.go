package user

// User is the shared model between backend and frontend
type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// Handler implements CRUD operations for users
type Handler struct{}

func (h *Handler) HandlerName() string { return "users" }
