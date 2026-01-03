# Vision of the CRUDP Protocol

## Introduction

CRUDP is a protocol for isomorphic Go applications where **everything is CRUD**. It provides centralized handler registration and automatic HTTP endpoint generation for both client (WASM) and server.

## Core Principles

### 1. Everything is CRUD

All operations map to Create, Read, Update, or Delete. File uploads, webhooks, exports—they're all CRUD operations with access to `*http.Request` when needed.

### 2. Automatic Endpoints

When you register a handler, CRUDP generates HTTP routes automatically:

| Action | HTTP Method | Route Pattern |
|--------|-------------|---------------|
| Create | `POST` | `/{handler_name}/{path...}` |
| Read   | `GET` | `/{handler_name}/{path...}` |
| Update | `PUT` | `/{handler_name}/{path...}` |
| Delete | `DELETE` | `/{handler_name}/{path...}` |

No manual route configuration needed.

### 3. Simplified Return Type

Handlers return `any` which can be:
- A result (struct, slice, primitive)
- An `error` (detected automatically by the server)

```go
func (h *Handler) Create(data ...any) any {
    // Return result or error
}
```

### 4. Request Injection

Server-side handlers receive `*http.Request` in the `data` slice, enabling:
- Multipart form parsing for file uploads
- Header access for webhook signature verification
- Any HTTP-specific logic without custom routes

```go
func (h *Handler) Create(data ...any) any {
    for _, item := range data {
        if r, ok := item.(*http.Request); ok {
            // Access headers, parse multipart, etc.
        }
    }
}
```

## Protocol Requirements

- TinyGo-friendly (minimal binary size)
- Batch operations via `/batch` endpoint
- Correlation IDs for async response tracking
- Transport decoupling (handlers don't import `net/http`)

## Related Documentation

- [ARCHITECTURE.md](ARCHITECTURE.md) - Technical architecture
- [HANDLER_REGISTER.md](HANDLER_REGISTER.md) - Handler interfaces
- [FILE_UPLOAD.md](FILE_UPLOAD.md) - File upload pattern
- [WEBHOOKS.md](WEBHOOKS.md) - Webhook handling
