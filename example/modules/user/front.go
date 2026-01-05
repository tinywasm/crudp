//go:build wasm

package user

// Create updates local state when server confirms creation
func (u *User) Create(data ...any) any {
	for _, item := range data {
		if u, ok := item.(*User); ok {
			// Update local state, DOM, etc.
			// Example: dom.El("#user-list").Append(renderUser(u))
			return u
		}
	}
	return nil
}

// Read updates UI with received users
func (u *User) Read(data ...any) any {
	for _, item := range data {
		switch v := item.(type) {
		case *User:
			// Single user received
			// Example: dom.El("#user-detail").SetHTML(renderUser(v))
			return v
		case []*User:
			// List of users received
			// Example: dom.El("#user-list").SetHTML(renderUsers(v))
			return v
		}
	}
	return nil
}

// Update updates local state after server confirms update
func (u *User) Update(data ...any) any {
	for _, item := range data {
		if u, ok := item.(*User); ok {
			// Update DOM element for this user
			// Example: dom.El("#user-" + strconv.Itoa(u.ID)).SetHTML(renderUser(u))
			return u
		}
	}
	return nil
}

// Delete removes element from DOM after server confirms
func (u *User) Delete(data ...any) any {
	/* 	for _, item := range data {
		if path, ok := item.(string); ok {
			// Remove from DOM
			// Example: dom.El("#user-" + path).Remove()
			return "deleted"
		}
	} */
	return nil
}
