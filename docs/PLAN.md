# tinywasm/crudp — Enhancement Plan (Typed Interfaces + Explicit Parameters)

> **Goal:** Replace variadic `any` interface signatures with explicit, single-typed parameters.
> Remove the semantic ambiguity in `Read(data ...any) any` by splitting into `Read(id string)`
> and `List()`. Add proper `error` return to all methods. The `db *orm.DB` is captured
> in the handler's constructor — never passed as a method parameter.
>
> **Status:** Executed

---

## Development Rules

- **Testing Runner:** `go install github.com/tinywasm/devflow/cmd/gotest@latest`
- **SRP:** `interfaces.go` defines the contract. `handlers.go` wires it. Separate concerns.
- **Breaking Changes Allowed:** API fluidity takes priority.
- **Standard Library Only:** No external assertion libraries in tests.

---

## Context: What Is Wrong Today

Current interfaces:

```go
type Creator interface { Create(data ...any) any }
type Reader  interface { Read(data ...any) any }
type Updater interface { Update(data ...any) any }
type Deleter interface { Delete(data ...any) any }
```

Problems:
1. **Variadic `any`** — the caller must know the undocumented argument order at runtime.
2. **`Read` conflates two operations** — list all vs. read by ID both go into `Read(data...any)`.
3. **`any` return value hides errors** — current code detects `if err, ok := result.(error)` instead of using a proper `error` return.
4. **No `db` constraint** — handlers must use global state to access the database.

The fix: explicit parameters, proper error returns, and handler constructors that capture
`*orm.DB` via closure — so the interface methods have zero hidden dependencies.

---

## Step 1 — Redefine Interfaces in `interfaces.go`

**Target File:** `interfaces.go`

Replace all current interface definitions:

```go
package crudp

// Creator handles entity creation.
// payload is the entity to create (concrete type asserted internally by the handler).
// Returns the created entity or an error.
type Creator interface {
    Create(payload any) (any, error)
}

// Reader handles entity retrieval.
// Read returns a single entity by its string ID.
// List returns all entities (no filter).
type Reader interface {
    Read(id string) (any, error)
    List() (any, error)
}

// Updater handles entity mutation.
// payload is the entity with updated fields.
// Returns the updated entity or an error.
type Updater interface {
    Update(payload any) (any, error)
}

// Deleter handles entity removal by ID.
type Deleter interface {
    Delete(id string) error
}

// NamedHandler provides the resource name used for routing and RBAC.
type NamedHandler interface {
    HandlerName() string
}

// DataValidator validates payload before execution.
// action: 'c' create, 'r' read, 'u' update, 'd' delete.
type DataValidator interface {
    ValidateData(action byte, payload any) error
}

// AccessLevel declares which role codes are allowed per action.
// Used by standalone mode (without tinywasm/rbac).
type AccessLevel interface {
    AllowedRoles(action byte) []byte
}
```

**Key changes vs before:**
- `Create(payload any) (any, error)` — single param, explicit error return
- `Read(id string) (any, error)` — explicit id, no more ambiguous `data ...any`
- `List() (any, error)` — new separate method for listing all entities
- `Update(payload any) (any, error)` — single param, explicit error return
- `Delete(id string) error` — explicit id, error-only return
- `ValidateData(action byte, payload any) error` — single payload (not variadic)

---

## Step 2 — Update `actionHandler` Internal Struct in `handlers.go`

**Target File:** `handlers.go`

Update the internal `actionHandler` struct to match the new method signatures:

```go
type actionHandler struct {
    name         string
    index        uint8
    handler      any
    dataType     reflect.Type
    Create       func(payload any) (any, error)
    Read         func(id string) (any, error)
    List         func() (any, error)
    Update       func(payload any) (any, error)
    Delete       func(id string) error
    ValidateData func(action byte, payload any) error
    AllowedRoles func(action byte) []byte
}
```

Update `RegisterHandlers` to bind the new methods:

```go
if creator, ok := h.(Creator); ok {
    ah.Create = creator.Create
    hasCRUD = true
}
if reader, ok := h.(Reader); ok {
    ah.Read = reader.Read
    ah.List = reader.List
    hasCRUD = true
}
// ... same pattern for Updater, Deleter
```

---

## Step 3 — Update `CallHandler` in `handlers.go`

**Target File:** `handlers.go`

Update `CallHandler` to use the new typed methods. The incoming `data ...any` is now
only used for access checks (passing `*http.Request`). The actual CRUD payload
is decoded separately and passed as a single `any`:

```go
func (cp *CrudP) CallHandler(handlerID uint8, action byte, data ...any) (any, error) {
    handler := cp.handlers[handlerID]

    // 1. Access Control
    if err := cp.accessCheck(handler, action, data...); err != nil {
        return nil, err
    }

    // 2. Extract payload (first element that is not *http.Request)
    var payload any
    var id string
    for _, d := range data {
        switch v := d.(type) {
        case string:
            id = v
        default:
            payload = v
        }
    }

    // 3. Validate
    if handler.ValidateData != nil {
        if err := handler.ValidateData(action, payload); err != nil {
            return nil, err
        }
    }

    // 4. Execute
    switch action {
    case 'c':
        if handler.Create != nil {
            return handler.Create(payload)
        }
    case 'r':
        if id == "" && handler.List != nil {
            return handler.List()
        }
        if id != "" && handler.Read != nil {
            return handler.Read(id)
        }
    case 'u':
        if handler.Update != nil {
            return handler.Update(payload)
        }
    case 'd':
        if handler.Delete != nil {
            return handler.Delete(id)
        }
    }
    return nil, Errf("action '%c' not implemented for handler: %s", action, handler.name)
}
```

---

## Step 4 — Handler Wrapper Pattern (Reference for Consumers)

Document in `docs/HANDLER_REGISTER.md` the recommended pattern for implementing handlers
that capture `*orm.DB` without global state. This pattern must be used by all tinywasm
ecosystem packages (e.g., `tinywasm/user`):

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

---

## Step 5 — Update Example and Tests

- Update `example/modules/user/back.go` to use the new interface signatures.
- Update `integration_stlib_test.go` to use `Read(id)`, `List()`, `Create(payload)`, etc.
- Run `gotest` — 100% pass required.

---

## Step 6 — Verify & Submit

1. Run `gotest` from project root. All tests must pass.
2. Update `docs/HANDLER_REGISTER.md` with the new interface contract and the wrapper pattern.
3. Run `gopush 'feat: typed interfaces, explicit id/list/payload params, proper error returns'`
