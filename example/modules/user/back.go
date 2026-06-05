//go:build !wasm

package user

import (
	. "github.com/tinywasm/fmt"
)

// Mock database
var users = []*User{
	{ID: 1, Name: "Alice", Email: "alice@example.com"},
	{ID: 2, Name: "Bob", Email: "bob@example.com"},
}

var nextID = 3

// Create handles user creation (server-side)
func (u *User) Create(payload any) (any, error) {
	if v, ok := payload.(*User); ok {
		v.ID = nextID
		nextID++
		users = append(users, v)
		return v, nil
	}
	return nil, nil
}

// Read handles user retrieval (server-side)
func (u *User) Read(id string) (any, error) {
	// Find user by ID
	for _, u := range users {
		if Sprintf("%d", u.ID) == id {
			return u, nil
		}
	}
	return nil, nil
}

// List handles all user retrieval (server-side)
func (u *User) List() (any, error) {
	return users, nil
}

// Update handles user modification (server-side)
func (u *User) Update(payload any) (any, error) {
	if v, ok := payload.(*User); ok {
		for _, u := range users {
			if u.ID == v.ID { // Assuming payload has the ID to update
				u.Name = v.Name
				u.Email = v.Email
				return u, nil
			}
		}
	}
	return nil, nil
}

// Delete handles user removal (server-side)
func (u *User) Delete(id string) error {
	for i, u := range users {
		if Sprintf("%d", u.ID) == id {
			users = append(users[:i], users[i+1:]...)
			return nil
		}
	}
	return nil
}
