# HTTP Routes & Middleware System

**Prerequisites:** Read [INTEGRATION_GUIDE.md](INTEGRATION_GUIDE.md) first.

## Overview

While CRUDP focuses on the communication protocol, it also provides hooks to easily integrate custom HTTP endpoints (like file uploads) and global middleware (like authentication) when using the standard library server.

These features are only available in environments that support the standard library `net/http` (typically using the `!wasm` build tag).

## 1. Optional HTTP Interfaces

CRUDP looks for these interfaces on your registered handlers during route registration.

**File: `http_stlib.go`**
```go
// Allows handlers to register custom HTTP routes (e.g., /upload)
type HttpRouteProvider interface {
    RegisterRoutes(mux *http.ServeMux)
}

// Allows handlers to provide global HTTP middleware
type MiddlewareProvider interface {
    Middleware(next http.Handler) http.Handler
}
```

## 2. Server Implementation

### Automatic Route Registration

When you call `RegisterRoutes(mux)`, CRUDP automatically iterates through all registered handlers and calls their `RegisterRoutes` method if they implement `HttpRouteProvider`.

```go
cp := router.NewRouter()
mux := http.NewServeMux()
cp.RegisterRoutes(mux) // Registers /api AND handler-specific routes
```

### Applying Middleware

Middleware is applied by wrapping your main handler using `ApplyMiddleware`.

```go
handler := cp.ApplyMiddleware(mux)
http.ListenAndServe(":8080", handler)
```

## 3. Example: Handler with Custom Routes

You can keep the CRUD logic in a cross-platform file and put the HTTP-specific logic in a file with the `!wasm` build tag.

**File: `modules/users/users_stlib.go`**
```go
//go:build !wasm
package users

import "net/http"

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
    mux.HandleFunc("/api/users/export", h.handleExport)
}

func (h *Handler) handleExport(w http.ResponseWriter, r *http.Request) {
    // Custom logic to export users as CSV
}

func (h *Handler) Middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        log.Println("Request to:", r.URL.Path)
        next.ServeHTTP(w, r)
    })
}
```

## 4. Key Considerations

- **Separation of Concerns**: Keep your protocol logic (`Create`, `Read`, etc.) in files accessible to WASM. Keep HTTP-specific logic (`RegisterRoutes`, `Middleware`) in backend-only files.
- **Middleware Order**: Middleware is applied in the order handlers were registered with `RegisterHandler`.
- **Flexibility**: Only implement these interfaces if you actually need them. Most handlers only need the standard CRUD interfaces.
