# File Upload Handling: "Upload & Reference" Pattern

This guide demonstrates how to implement file uploads within CRUDP using `HttpRouteProvider` for HTTP routes while keeping CRUD handlers decoupled from transport details.

## Core Pattern: "Upload & Reference"

**Why not pass `http.ResponseWriter` to handlers?**

Passing `w` to handlers breaks the asynchronous architecture and clean separation:
1. **Transport Coupling:** Handlers become HTTP-only, untestable without complex mocks.
2. **Single Responsibility:** HTTP layer handles network I/O; CRUD handlers manage business logic.

**Solution:** Use custom HTTP routes to handle physical storage and CRUD handlers to manage file metadata references.

## 1. File Reference & Handler

**File: `modules/files/files.go`**
```go
package files


type FileReference struct {
    ID    string `json:"id"`
    Path  string `json:"path"`
    Name  string `json:"name"`
}

type Handler struct {
    // Database or storage service
}

func (h *Handler) Create(data ...any) (any, error) {
    for _, item := range data {
        ref := item.(*FileReference)
        // Logic: Save file metadata to database
    }
    return "metadata saved", nil
}
```

## 2. HTTP Route Implementation

Implement the `HttpRouteProvider` interface in a backend-only file.

**File: `modules/files/files_stlib.go`**
```go
//go:build !wasm
package files

import (
    "io"
    "net/http"
    "os"
)

// Implement HttpRouteProvider interface
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
    mux.HandleFunc("POST /files/upload", h.handleFileUpload)
}

func (h *Handler) handleFileUpload(w http.ResponseWriter, r *http.Request) {
    // 1. Parse Multipart Form
    file, header, _ := r.FormFile("file")
    defer file.Close()

    // 2. Save file to physical storage (disk/S3/etc)
    dst, _ := os.Create("/uploads/" + header.Filename)
    io.Copy(dst, file)

    // 3. Create metadata reference
    ref := &FileReference{
        ID:   "unique-id",
        Path: "/uploads/" + header.Filename,
        Name: header.Filename,
    }

    // 4. Call CRUD logic directly
    _, err := h.Create(ref)
    if err != nil {
        http.Error(w, "Failed to save metadata", 500)
        return
    }

    w.Write([]byte("Upload successful"))
}
```

## 3. Integration

**File: `web/server.go`**
```go
func main() {
    cp := crudp.NewDefault()
    cp.RegisterHandler(&files.Handler{})
    
    mux := http.NewServeMux()
    cp.RegisterRoutes(mux) // This calls files.Handler.RegisterRoutes
    
    http.ListenAndServe(":8080", mux)
}
```

## Key Benefits

- **🧪 Testability**: CRUD handlers can be tested with mock references without needing a real HTTP server.
- **🔌 Reusability**: The same metadata logic works whether the file came from HTTP, a CLI import, or a background worker.
- **🛡️ Security**: You can wrap the entire `mux` with authentication middleware using `cp.ApplyMiddleware(mux)`.