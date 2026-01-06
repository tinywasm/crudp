//go:build !wasm

package user

import (
	"net/http"

	"github.com/tinywasm/context"
	. "github.com/tinywasm/fmt"
)

// Mock database
var users = []*User{
	{ID: 1, Name: "Alice", Email: "alice@example.com"},
	{ID: 2, Name: "Bob", Email: "bob@example.com"},
}

var nextID = 3

// Create handles user creation (server-side)
func (u *User) Create(data ...any) any {
	for _, item := range data {
		switch v := item.(type) {
		case *context.Context:
			// Use context for auth, tracing, etc.
		case *http.Request:
			// Access headers, parse multipart, etc.
		case *User:
			v.ID = nextID
			nextID++
			users = append(users, v)
			return v
		}
	}
	return nil
}

// Read handles user retrieval (server-side)
func (u *User) Read(data ...any) any {
	for _, item := range data {
		if path, ok := item.(string); ok {
			if path == "" {
				return users // All users
			}
			// Find user by ID
			for _, u := range users {
				if Fmt("%d", u.ID) == path {
					return u
				}
			}
			return nil
		}
	}
	return users
}

// Update handles user modification (server-side)
func (u *User) Update(data ...any) any {
	var targetID string
	var updateData *User

	for _, item := range data {
		switch v := item.(type) {
		case string:
			targetID = v
		case *User:
			updateData = v
		}
	}

	if targetID != "" && updateData != nil {
		for _, u := range users {
			if Fmt("%d", u.ID) == targetID {
				u.Name = updateData.Name
				u.Email = updateData.Email
				return u
			}
		}
	}
	return nil
}

// Delete handles user removal (server-side)
func (u *User) Delete(data ...any) any {
	for _, item := range data {
		if path, ok := item.(string); ok {
			for i, u := range users {
				if Fmt("%d", u.ID) == path {
					users = append(users[:i], users[i+1:]...)
					return "deleted"
				}
			}
		}
	}
	return nil
}
