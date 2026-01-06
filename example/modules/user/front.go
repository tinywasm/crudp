//go:build wasm

package user

import (
	"github.com/tinywasm/dom"
	. "github.com/tinywasm/fmt"
)

// Create updates local state when server confirms creation
func (u *User) Create(data ...any) any {
	for _, item := range data {
		if u, ok := item.(*User); ok {
			// Update local state, DOM, etc.
			if el, ok := dom.Get("user-list"); ok {
				el.AppendHTML(renderUser(u))
			}
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
			if el, ok := dom.Get("user-detail"); ok {
				el.SetHTML(renderUser(v))
			}
			return v
		case []*User:
			if el, ok := dom.Get("user-list"); ok {
				var content string
				for _, usr := range v {
					content += renderUser(usr)
				}
				el.SetHTML(content)
			}
			return v
		}
	}
	return nil
}

// Update updates local state after server confirms update
func (u *User) Update(data ...any) any {
	for _, item := range data {
		if u, ok := item.(*User); ok {
			if el, ok := dom.Get(Fmt("user-%d", u.ID)); ok {
				el.SetHTML(renderUser(u))
			}
			return u
		}
	}
	return nil
}

// Delete removes element from DOM after server confirms
func (u *User) Delete(data ...any) any {
	for _, item := range data {
		if path, ok := item.(string); ok {
			if el, ok := dom.Get(Fmt("user-%s", path)); ok {
				el.Remove()
			}
			return "deleted"
		}
	}
	return nil
}

// Helper render function
func renderUser(u *User) string {
	return Html(`<div id="user-%d" class="user-item">
		<strong>%s</strong> (%s)
	</div>`, u.ID, u.Name, u.Email).String()
}
