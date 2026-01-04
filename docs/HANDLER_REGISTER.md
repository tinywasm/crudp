# Handler Registration

## Core Concepts

Handlers are the core of a CRUDP application. They implement the business logic for your data models. Handlers are registered with a `CrudP` instance, which then dispatches incoming requests based on the `HandlerID` provided in the protocol packets.

## CRUD Interfaces

Handlers implement one or more of the following interfaces. Each method returns two values: the result (which can be `nil`) and an `error`.

**File: `interfaces.go`**
```go
package crudp

// Creator handles create operations.
type Creator interface {
    Create(data ...any) any
}

// Reader handles read operations.
type Reader interface {
    Read(data ...any) any
}

// Updater handles update operations.
type Updater interface {
    Update(data ...any) any
}

// Deleter handles delete operations.
type Deleter interface {
    Delete(data ...any) any
}
```

**Key Points:**

-   **Return types**: Returning an `error` object (e.g. `return errorString("failed")`) allows CRUDP to automatically populate the `MessageType` (Error) and `Message` fields. WASM handlers typically return `nil` if they only update the DOM.
-   **Dynamic Results**: The return value (`any`) can be a simple struct, a slice of structs, primitive values, or an `error`.

## Processing Data with For Loop

Handlers receive injected values (context, http.Request) plus user data in the `data` slice. Always iterate with a type switch:

```go
func (h *UserHandler) Create(data ...any) any {
    var ctx *context.Context
    var req *http.Request  // Only present on server
    var id  string         // From URL /users/{id}
    var users []*User

    for _, item := range data {
        switch v := item.(type) {
        case string:
            id = v    // Captured from URL path
        case *context.Context:
            ctx = v
        case *http.Request:
            req = v  // Only on server (!wasm)
        case *User:
            users = append(users, v)
        }
    }

    // Now use ctx, req (if present), id, and users
    return processedUsers
}
```

## Handler Naming

CRUDP automatically determines a handler's name, which is used for internal registration and logging.

1.  **By Convention (Reflection):** Default behavior. Converts the struct type name to `snake_case`. (e.g., `UserHandler` -> `"user_handler"`).
2.  **Explicitly (NamedHandler):** Implement this interface to override the automatic name.

```go
type NamedHandler interface {
    HandlerName() string
}
```

## Validation

CRUDP provides an optional interface for data validation:

-   `Validator`: Called **automatically** by `CallHandler` or `Execute` before the action method. If it returns an error, the action is aborted.

```go
type Validator interface {
    Validate(action byte, data ...any) error
}
```

> [!NOTE]
> For field-level validation (e.g., UI feedback), see `FieldValidator` in [`tinywasm/form`](https://github.com/tinywasm/form).

## `RegisterHandlers`

Use this method to register your handler instances. Order matters: the index in the slice becomes the `HandlerID` used in the protocol.

```go
cp := crudp.New()
err := cp.RegisterHandlers(&UserHandler{}, &ProductHandler{})
```

During registration, CRUDP:
1.  Resolves the handler's name.
2.  Binds implemented CRUD methods into an internal table for zero-allocation dispatching.
3.  Logs the registration details if a logger is configured.
