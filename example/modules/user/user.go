package user

// User is the shared model between backend and frontend
type User struct {
	ID    int
	Name  string
	Email string
}

func (u *User) HandlerName() string { return "users" }

func (u *User) ValidateData(action byte, data ...any) error { return nil }
func (u *User) AllowedRoles(action byte) []byte             { return []byte{'*'} }

// Add returns all entities from this module
func Add() []any {
	return []any{&User{}}
}
