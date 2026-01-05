package user

// User is the shared model between backend and frontend
type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func (u *User) HandlerName() string { return "users" }

func (u *User) ValidateData(action byte, data ...any) error { return nil }

// Add returns all entities from this module
func Add() []any {
	return []any{&User{}}
}
