# HTTP Routes & Middleware System

## Overview

CRUDP generates automatic HTTP endpoints for your handlers. It also provides a hook for global middleware when using the standard library server (`net/http`).

## 1. Automatic Routing

When you call `RegisterRoutes(mux)`, CRUDP generates endpoints for every handler and its implemented CRUD actions.

| Action | HTTP Method | URL Pattern |
|--------|-------------|-------------|
| Create | `POST` | `/{handler_name}/{path...}` |
| Read   | `GET` | `/{handler_name}/{path...}` |
| Update | `PUT` | `/{handler_name}/{path...}` |
| Delete | `DELETE` | `/{handler_name}/{path...}` |

### Accessing Request Details

Handlers receive the following injected values in the `data ...any` slice:
1. `*context.Context` (always)
2. `*http.Request` (server-side only)
3. `string` (the `{path...}` wildcard value)

## 2. Middleware

Handlers can provide global HTTP middleware by implementing the `MiddlewareProvider` interface.

**File: `http_stlib.go`**
```go
type MiddlewareProvider interface {
    Middleware(next http.Handler) http.Handler
}
```

### Applying Middleware

Middleware is applied by wrapping your main mux/handler using `ApplyMiddleware`.

```go
cp := router.NewRouter()
mux := http.NewServeMux()
cp.RegisterRoutes(mux)

// Apply all middleware from handlers
handler := cp.ApplyMiddleware(mux)

http.ListenAndServe(":8080", handler)
```

## 3. Example: Handling Custom Logic

Since handlers receive the `*http.Request`, you can handle any HTTP-specific logic (like file uploads or webhooks) directly inside your CRUD methods.

```go
func (h *UserHandler) Create(data ...any) any {
    for _, item := range data {
        if r, ok := item.(*http.Request); ok {
            // Check headers, handle multipart, etc.
            if r.Header.Get("X-Custom-Webhook") != "" {
                return h.handleWebhook(r)
            }
        }
    }
    // Default JSON processing...
    return nil
}
```

## 4. Key Considerations

- **Middleware Order**: Middleware is applied in the order handlers were registered.
- **WASM Compatibility**: Middleware logic should be in files with the `//go:build !wasm` tag to avoid including `net/http` in the WASM binary.
