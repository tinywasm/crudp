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
    Create(data ...any) (any, error)
}

// Reader handles read operations.
type Reader interface {
    Read(data ...any) (any, error)
}

// Updater handles update operations.
type Updater interface {
    Update(data ...any) (any, error)
}

// Deleter handles delete operations.
type Deleter interface {
    Delete(data ...any) (any, error)
}
```

**Key Points:**

-   **Return types**: Returning an explicit `error` allows CRUDP to automatically populate the `MessageType` (Error) and `Message` fields in the `PacketResult` sent back to the client.
-   **Dynamic Results**: The first return value (`any`) can be a simple struct, a slice of structs (for multi-item responses), or primitive values.

## Processing Data with For Loop

Handlers receive injected values (context, http.Request) plus user data in the `data` slice. Always iterate with a type switch:

```go
func (h *UserHandler) Create(data ...any) (any, error) {
    var ctx *context.Context
    var req *http.Request  // Only present on server
    var users []*User

    for _, item := range data {
        switch v := item.(type) {
        case *context.Context:
            ctx = v
        case *http.Request:
            req = v  // Only on server (!wasm)
        case *User:
            users = append(users, v)
        }
    }

    // Now use ctx, req (if present), and users
    return processedUsers, nil
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

CRUDP provides two optional interfaces for data validation:

-   `Validator`: Called **automatically** by `CallHandler` or `Execute` before the action method. If it returns an error, the action is aborted.
-   `FieldValidator`: For manual validation of individual fields (typically used by the UI).

```go
type Validator interface {
    Validate(action byte, data ...any) error
}

type FieldValidator interface {
    ValidateField(fieldName string, value string) error
}
```

## `RegisterHandler`

Use this method to register your handler instances. Order matters: the index in the slice becomes the `HandlerID` used in the protocol.

```go
cp := crudp.New(binary.Encode, binary.Decode)
err := cp.RegisterHandler(&UserHandler{}, &ProductHandler{})
```

During registration, CRUDP:
1.  Resolves the handler's name.
2.  Binds implemented CRUD methods into an internal table for zero-allocation dispatching.
3.  Logs the registration details if a logger is configured.
