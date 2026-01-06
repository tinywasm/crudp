# CRUDP Integration Guide

## Overview

This guide explains how to integrate CRUDP into a Go project for isomorphic execution between server and WASM client. CRUDP automatically routes packets to handlers; business modules remain decoupled.

## Project Structure

The example project follows this structure:

```
example/
├── modules/
│   ├── modules.go    # Handler instances collection & DI
│   └── users/
│       ├── user.go   # Shared: model + handler struct
│       ├── back.go   # Backend: server-side CRUD logic
│       └── front.go  # Frontend: WASM DOM updates
└── web/
    ├── client.go     # WASM entry point (Init CRUDP)
    └── server.go     # Server entry point (Init CRUDP)
```

## Implementation Steps

### 1. Define Shared Entity

In CRUDP, your data model (Entity) is also your handler. This simplifies the design and ensures consistency.

- [**File: `modules/user/user.go`**](./modules/user/user.go)

All CRUD entities must implement:
- `HandlerName() string`: Unique name for registration.
- `ValidateData(action byte, data ...any) error`: Data validation logic.
- `AllowedRoles(action byte) []byte`: Access control (see [ACCESS_CONTROL.md](../docs/ACCESS_CONTROL.md)).

```go
func (u *User) AllowedRoles(action byte) []byte {
    if action == 'r' { return []byte{'*'} } // Read: any authenticated user
    return []byte{'a'} // Write: only admins
}
```

### 2. Implement Backend Logic

Server-side CRUD operations receive `*context.Context`, `*http.Request`, and typed data.

- [**File: `modules/user/back.go`**](./modules/user/back.go)

Implement `Create`, `Read`, `Update`, or `Delete` methods as needed using build tags `//go:build !wasm`.

### 3. Implement Frontend Logic

WASM handlers update local state and DOM when receiving server responses.

- [**File: `modules/user/front.go`**](./modules/user/front.go)

Implement the same CRUD methods using build tags `//go:build wasm`.

### 4. Collect Modules

Each module exposes an `Add()` function returning its entities. Collect them in a central place.

- [**File: `modules/modules.go`**](./modules/modules.go)

### 5. Server Entry Point

Initialize CRUDP, configure access control, and register routes.

- [**File: `web/server.go`**](./web/server.go)

```go
func main() {
    cp := crudp.New()
    cp.SetUserRoles(func(data ...any) []byte {
        return []byte{'*'} // Implement your role extraction logic
    })
    cp.RegisterHandlers(modules.Init()...)
    cp.RegisterRoutes(mux)
}
```

### 6. Client Entry Point

Initialize CRUDP and connect it to the global fetch handler.

- [**File: `web/client.go`**](./web/client.go)

```go
func main() {
    cp := crudp.New()
    cp.RegisterHandlers(modules.Init()...)
    cp.InitClient() // Intercepts fetch responses
    select {} // Keep WASM alive
}
```

## Key Principles

- **📦 Decoupling**: Business modules don't import CRUDP.
- **🔄 Isomorphic**: Same handler struct, different logic per platform using build tags.
- **⚡ Automatic Routing**: RESTful endpoints generated from handler names.
- **🛠️ Direct Initialization**: No complex boilerplates; just `New()` + `RegisterHandlers()`.
