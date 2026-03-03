# Entity Registration

## Overview

In CRUDP, your data models (**Entities**) implement the business logic. They are registered with a `CrudP` instance, which then dispatches incoming requests based on the `HandlerID` provided in the protocol packets.

For a complete step-by-step example, see the [Integration Guide](./INTEGRATION_GUIDE.md).

## CRUD Interfaces

Entities implement one or more of the CRUD interfaces defined in [`interfaces.go`](../interfaces.go):

- `Creator`: `Create(payload any) (any, error)`
- `Reader`: `Read(id string) (any, error)` and `List() (any, error)`
- `Updater`: `Update(payload any) (any, error)`
- `Deleter`: `Delete(id string) error`

**Key Points:**
- **Return types**: Returning an `error` allows CRUDP to automatically populate error messages in the response.
- **Dynamic Results**: Results can be structs, slices, or primitives.

## Mandatory Interfaces

**Security and Identification**: Every Entity that implements at least one CRUD operation **must** also implement:

1.  **`NamedHandler`**: Provides the unique name.
2.  **`DataValidator`**: Validates data before the action.
3.  **[`AccessLevel`](./ACCESS_CONTROL.md)**: Defines hierarchical permissions.

`RegisterHandlers` will return an error if a CRUD Entity is missing these implementations.

## Registration

Use `RegisterHandlers` to register Entity instances. The order in the slice determines the `HandlerID`.

```go
err := cp.RegisterHandlers(&User{}, &Product{})
```

## Handler Wrapper Pattern (Best Practice)

For handlers that need external dependencies (like a database connection) without using global state, use a wrapper struct that captures dependencies in its constructor. The entity model struct itself remains a pure data type.

```go
// The entity model struct (User) stays pure — no CRUDP methods on it.
// A separate handler wrapper captures db in its constructor.

type userCRUD struct{ db *orm.DB }

func (h *userCRUD) HandlerName() string                  { return "users" }
func (h *userCRUD) AllowedRoles(action byte) []byte      { return []byte{'a'} }
func (h *userCRUD) ValidateData(action byte, _ any) error { return nil }

func (h *userCRUD) Create(payload any) (any, error) {
    u := payload.(User)
    return createUser(h.db, u.Email, u.Name, u.Phone)
}
func (h *userCRUD) Read(id string) (any, error)   { return getUser(h.db, nil, id) }
func (h *userCRUD) List() (any, error)            { return listUsers(h.db) }
func (h *userCRUD) Update(payload any) (any, error) {
    u := payload.(User)
    return u, updateUser(h.db, u.ID, u.Name, u.Phone)
}
func (h *userCRUD) Delete(id string) error        { return deleteUser(h.db, id) }

// Registration in the consuming app:
// cp.RegisterHandlers(&userCRUD{db: db})
```

Key properties of this pattern:
- No global `store` — `db` is explicit in the constructor.
- Model struct (`User`) remains a pure data type — no behavior attached.
- Each entity gets a dedicated `*CRUD` type — follows SRP.
- Type assertion (`payload.(User)`) happens once inside the handler — all external code is clean.
