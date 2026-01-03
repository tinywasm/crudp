# CRUDP Architecture

> **Status:** Specification Document  
> **Last Updated:** January 2026

## Overview

CRUDP (CRUD Protocol) is a logic-only execution engine for isomorphic Go applications. It handles the registration and execution of CRUD operations, delegating transport/batching to tools like `tinywasm/broker`.

## Core Principles

1. **Isomorphic Handlers:** Same handler code runs on frontend (WASM) and backend (Server).
2. **Zero Transport Coupling:** CRUDP doesn't know about HTTP, WebSockets, or SSE.
3. **Simplified API:** CRUD methods return `any` (which can be an error) for minimal binary size.
4. **Automatic Endpoints:** Native HTTP routing based on handler name and action.
5. **TinyGo Compatible:** Optimized for small WASM binaries.

## Modular Architecture

1. **`tinywasm/broker`**: Handles client-side batching and consolidation of requests.
2. **`tinywasm/crudp`**: (This repo) Defines protocol types (Packet, BatchRequest, BatchResponse) and maps them to registered handlers.

### Workflow Diagram

```mermaid
flowchart TD
    subgraph Client["FRONTEND (WASM)"]
        UI["UI Component"] --> B["tinywasm/broker\n(Queue / Consolidation)"]
    end

    B -- "Transport (HTTP/WS)" --> S["Server HTTP Handler\n(crudp.RegisterRoutes)"]

    subgraph Backend["BACKEND (Server)"]
        S --> E["crudp.Execute()"]
        E --> H["Registered Handlers\n(Create, Read, ...)"]
        H --> DB[(Database)]
    end

    H -- "(data, error)" --> E
    E -- "packet.BatchResponse" --> S
    S -- "JSON/Binary" --> Client
```

## Key Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Serialization | `Encode`/`Decode` funcs | Support JSON (tinywasm/json) or Binary (tinywasm/binary) formats |
| Handler signature | `func(data ...any) any` | Simplifies API and reduces binary size |
| Packet structure | Internal to `crudp` | Unified protocol definition and execution |
| Batching | Delegated to `tinywasm/broker` | Keep CRUDP core simple and focused on execution |
| Message types | `0-4` (uint8) | Normal, Info, Error, Warning, Success from `tinywasm/fmt` |
| HTTP methods | POST/GET/PUT/DELETE → c/r/u/d | Standard CRUD mapping |
| HandlerName | Automatic or Explicit | Use `NamedHandler` interface or reflection fallback |

## Dependencies

```
crudp
├── github.com/tinywasm/binary # Recommended codec
├── github.com/tinywasm/fmt    # String conversion and message types
└── github.com/tinywasm/fetch  # Client-side transport (WASM)
```

## Key Interfaces

```go
// CRUD interfaces - return any (result or error)
type Creator interface { Create(data ...any) any }
type Reader interface  { Read(data ...any) any }
type Updater interface { Update(data ...any) any }
type Deleter interface { Delete(data ...any) any }

// Optional interfaces
type NamedHandler interface { HandlerName() string }
type Validator interface { Validate(action byte, data ...any) error }
type FieldValidator interface { ValidateField(fieldName, value string) error }
```

## Implementation Status

| Feature | Status | Package |
|------|--------|----------|
| Batching | ✅ Done | `tinywasm/broker` |
| Execution | ✅ Done | `tinywasm/crudp` |
| Protocol Details | ✅ Done | `tinywasm/crudp` |
| HTTP Integration | ✅ Done | `tinywasm/crudp` (RegisterRoutes) |

## Related Documentation

- [INTEGRATION_GUIDE.md](INTEGRATION_GUIDE.md) - How to integrate CRUDP
- [INITIALIZATION.md](INITIALIZATION.md) - Initialization and serialization
- [HANDLER_REGISTER.md](HANDLER_REGISTER.md) - How to create and register handlers
- [LIMITATIONS.md](LIMITATIONS.md) - Supported data types
