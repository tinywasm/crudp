# CRUDP Integration Guide

## Overview

This guide explains how to integrate CRUDP into a Go project for protocol execution between server and WASM client. CRUDP is used to map network packets to handler logic; business modules remain decoupled.

## Project Structure

```
myProject/
├── modules/
│   ├── modules.go          # Handler instances collection
│   └── users/
│       └── users.go        # Business logic (implements CRUD interfaces)
├── pkg/
│   └── router/
│       └── router.go       # CRUDP initialization
├── web/
│   ├── client.go           # WASM entry point
│   └── server.go           # Server entry point
└── go.mod                  # Dependencies: tinywasm/crudp, tinywasm/broker
```

## Implementation Steps

### 1. Define Handler with CRUD Interfaces

Business modules implement standard interfaces. They should return the result of the operation and an error.

**File: `modules/users/users.go`**
```go
package users

import (
    "github.com/tinywasm/context"
    "net/http"
)

type Handler struct{}

// Implement CRUDP interfaces - iterate data with type switch
func (h *Handler) Create(data ...any) (any, error) {
    var ctx *context.Context
    var users []*User

    for _, item := range data {
        switch v := item.(type) {
        case *context.Context:
            ctx = v
        case *http.Request:
            // Only on server - use for auth, headers, etc.
        case *User:
            users = append(users, v)
        }
    }

    // Process users with context available
    _ = ctx
    return "user created", nil
}
```

### 2. Register Handlers

Collect all business modules into a single list for registration.

**File: `modules/modules.go`**
```go
package modules

import "myProject/modules/users"

func Init() []any {
    return []any{
        &users.Handler{}, // Registered at index 0 (HandlerID: 0)
    }
}
```

### 3. Initialize CRUDP Router

Configure the CRUDP instance with your modules.

**File: `pkg/router/router.go`**
```go
package router

import (
    "github.com/tinywasm/crudp"
    "github.com/tinywasm/binary"
    "myProject/modules"
)

func NewRouter() *crudp.CrudP {
    cp := crudp.New(binary.Encode, binary.Decode)
    cp.RegisterHandler(modules.Init()...)
    return cp
}
```

### 4. Server Integration (Standard HTTP)

Use `RegisterRoutes` to expose the endpoint automatically.

**File: `web/server.go`**
```go
//go:build !wasm

package main

import (
    "net/http"
    "myProject/pkg/router"
)

func main() {
    cp := router.NewRouter()
    
    mux := http.NewServeMux()
    // Registers POST /api endpoint
    cp.RegisterRoutes(mux)
    
    http.ListenAndServe(":8080", mux)
}
```

### 5. Client Integration (WASM Batching)

On the client side, use `tinywasm/broker` to batch operations.

**File: `web/client.go`**
```go
//go:build wasm

package main

import (
    "github.com/tinywasm/broker"
    "myProject/pkg/router"
)

func main() {
    cp := router.NewRouter()
    
    // Create a broker to batch outgoing requests
    b := broker.New(50) // 50ms batch window
    
    b.SetOnFlush(func(items []broker.Item) {
        // Encode items using BatchRequest and send via fetch
    })
    
    // Add logic to enqueue operations...
}
```

## Key Principles

- **📦 Decoupling:** Business modules don't import CRUDP; they just implement basic Go interfaces.
- **⚡ Zero Overhead:** Internal method binding avoids reflection during request execution.
- **🎯 Execution Only:** CRUDP focuses on mapping packets to handlers; it doesn't care how they are transported.
