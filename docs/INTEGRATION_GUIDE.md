# CRUDP Integration Guide

## Overview

This guide explains how to integrate CRUDP into a Go project for isomorphic execution between server and WASM client. CRUDP automatically routes packets to handlers; business modules remain decoupled.

## Project Structure

```
myProject/
├── modules/
│   ├── modules.go              # Handler instances collection & DI
│   └── users/
│       ├── user.go             # Shared: model + handler struct
│       ├── back.go             # Backend: server-side CRUD logic
│       └── front.go            # Frontend: WASM DOM updates
└── web/
    ├── client.go               # WASM entry point (Init CRUDP)
    └── server.go               # Server entry point (Init CRUDP)
```

## Implementation Steps

### 1. Define Shared Model and Handler

Keep the model and handler struct in a file without build tags.

**File: `modules/users/user.go`**
```go
package user

// User is the shared model between backend and frontend
type User struct {
    ID    int    
    Name  string 
    Email string
}

// Handler implements CRUD operations for users
type Handler struct{}

func (h *Handler) HandlerName() string { return "users" }
```

### 2. Implement Backend Logic

Server-side CRUD operations receive `*context.Context`, `*http.Request`, and typed data.

**File: `modules/users/back.go`**
```go
//go:build !wasm

package user

import (
    "github.com/tinywasm/fmt"
    "github.com/tinywasm/context"
    "net/http"
)

// Mock database
var users = []*User{
    {ID: 1, Name: "Alice", Email: "alice@example.com"},
    {ID: 2, Name: "Bob", Email: "bob@example.com"},
}

func (h *Handler) Create(data ...any) any {
    for _, item := range data {
        switch v := item.(type) {
        case *context.Context:
            // Use for auth, tracing, cancellation
        case *http.Request:
            // Access headers, parse multipart uploads
        case *User:
            // Save to DB...
            return v
        }
    }
    return nil
}

func (h *Handler) Read(data ...any) any {
    for _, item := range data {
        if path, ok := item.(string); ok {
            if path == "" { return users } // All users
            for _, u := range users {
                if fmt.Fmt("%d", u.ID) == path { return u }
            }
        }
    }
    return users
}
```

### 3. Implement Frontend Logic

WASM handlers update local state and DOM when receiving server responses.

**File: `modules/users/front.go`**
```go
//go:build wasm

package user

import (
    "github.com/tinywasm/dom"
    . "github.com/tinywasm/fmt"
)

func (h *Handler) Read(data ...any) any {
    for _, item := range data {
        switch v := item.(type) {
        case *User:
            el, _ := dom.Get("user-detail")
            el.SetHTML(renderUserItem(v))
        case []*User:
            listEl, _ := dom.Get("user-list")
            listEl.SetHTML(renderUserList(v))
        }
    }
    return nil
}

// Helper render functions
func renderUserItem(u *User) string {
    return Html(`<div id="user-%d" class="user-item">
        <span>%s</span>
    </div>`, u.ID, u.Name).String()
}
```

### 4. Collect Modules (Dependency Injection)

Use this file to instantiate handlers and inject any required dependencies (DB, services).

**File: `modules/modules.go`**
```go
package modules

import "myProject/modules/user"

func Init() []any {
    return []any{
        &user.Handler{},
        // Add other handlers here...
    }
}
```

### 5. Server Entry Point

Initialize CRUDP and register routes directly.

**File: `web/server.go`**
```go
//go:build !wasm

package main

import (
    "net/http"
    "github.com/tinywasm/crudp"
    "myProject/modules"
)

func main() {
    mux := http.NewServeMux()

    // 1. Initialize CRUDP
    cp := crudp.New()
    
    // 2. Register Handlers
    cp.RegisterHandlers(modules.Init()...)
    
    // 3. Expose automatic endpoints
    cp.RegisterRoutes(mux) 
    
    http.ListenAndServe(":8080", mux)
}
```

### 6. Client Entry Point

Initialize CRUDP and connect it to the global fetch handler.

**File: `web/client.go`**
```go
//go:build wasm

package main

import (
    "github.com/tinywasm/crudp"
    "myProject/modules"
)

func main() {
    // 1. Initialize CRUDP
    cp := crudp.New()
    
    // 2. Register Handlers
    cp.RegisterHandlers(modules.Init()...)
    
    // 3. Connect responses to handlers
    cp.InitClient() 
    
    select {} 
}
```

## Key Principles

- **📦 Decoupling**: Business modules don't import CRUDP.
- **🔄 Isomorphic**: Same handler struct, different logic per platform using build tags.
- **⚡ Automatic Routing**: RESTful endpoints generated from handler names.
- **🛠️ Direct Initialization**: No complex boilerplates; just `New()` + `RegisterHandlers()`.
