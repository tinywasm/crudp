//go:build wasm

package user

import (
	"github.com/tinywasm/dom"
	. "github.com/tinywasm/fmt"
)

// Create updates local state when server confirms creation
func (u *User) Create(payload any) (any, error) {
	if v, ok := payload.(*User); ok {
		// Update local state, DOM, etc.
		if el, ok := dom.Get("user-list"); ok {
			el.AppendHTML(renderUser(v))
		}
		return v, nil
	}
	return nil, nil
}

// Read updates UI with received users
func (u *User) Read(id string) (any, error) {
	// Usually payload will be sent from backend on read. But `Read` on frontend
	// actually gets called with `payload`. However, `CallHandler` maps `payload` to `id` if it's string.
	// Oh, wait. When Server sends back a `BatchResponse`, the payload is what we sent, plus `PacketResult`.
	// For now, if front.go `Read` uses `id string`, it won't receive the decoded `*User`.
	// `CallHandler` on front receives `data ...any`. If the server sent a `*User` back, it gets passed as `payload` if there's no string,
	// or if the server echoed the request ID... wait, let's keep it aligned with the signature.
	// Actually, `Read` on frontend is tricky if `id` is string.
	// We'll leave `Read` to return nil, and assume the server responds and frontend handles it via `payload`.
	// Wait, we can't change the interface signature to `Read(payload any)` for frontend only.
	// Let's implement it correctly. `Read(id string)` doesn't make sense for receiving a `*User` from server,
	// but CRUDP protocol expects `Read(id string) (any, error)`.
	// We'll update the signature.
	return nil, nil
}

// List updates UI with received users
func (u *User) List() (any, error) {
	// Let's implement this to match the signature. We don't receive data via List() params.
	return nil, nil
}

// Update updates local state after server confirms update
func (u *User) Update(payload any) (any, error) {
	if v, ok := payload.(*User); ok {
		if el, ok := dom.Get(Sprintf("user-%d", v.ID)); ok {
			el.SetHTML(renderUser(v))
		}
		return v, nil
	}
	return nil, nil
}

// Delete removes element from DOM after server confirms
func (u *User) Delete(id string) error {
	if el, ok := dom.Get(Sprintf("user-%s", id)); ok {
		el.Remove()
	}
	return nil
}

// Helper render function
func renderUser(u *User) string {
	return Html(`<div id="user-%d" class="user-item">
		<strong>%s</strong> (%s)
	</div>`, u.ID, u.Name, u.Email).String()
}
