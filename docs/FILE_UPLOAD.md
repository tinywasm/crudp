# File Upload Handling: "Direct CRUD" Pattern

This guide demonstrates how to implement file uploads within CRUDP using the automatic endpoints. By injecting `*http.Request` into the handler, you can handle multipart forms directly within your CRUD methods.

## Core Pattern: Direct Request Handling

With automatic endpoints, a `POST /files` request is routed to the `Create` method of the `files` handler. If the request is a multipart form (file upload), the handler can detect this by checking the injected `*http.Request`.

## 1. File Handler Implementation

The handler uses a type switch to detect if it's receiving a JSON object (WASM or Batch) or the raw `*http.Request` (Direct HTTP Upload).

**File: `modules/files/files.go`**
```go
package files

import (
    "io"
    "net/http"
    "os"
)

type FileReference struct {
    ID    string `json:"id"`
    Path  string `json:"path"`
    Name  string `json:"name"`
}

type Handler struct{}

func (h *Handler) HandlerName() string { return "files" }

func (h *Handler) Create(data ...any) any {
    for _, item := range data {
        switch v := item.(type) {
        case *http.Request:
            // 1. Handle Multipart Upload (Server-side only logic)
            file, header, err := v.FormFile("file")
            if err != nil {
                return err
            }
            defer file.Close()

            // 2. Save file
            path := "/uploads/" + header.Filename
            dst, _ := os.Create(path)
            io.Copy(dst, file)

            // 3. Return the reference
            return &FileReference{
                ID:   "unique-id",
                Path: path,
                Name: header.Filename,
            }

        case *FileReference:
            // 4. Handle JSON Metadata (Batch or Client-side)
            // Save metadata to database...
            return v
        }
    }
    return nil
}
```

## 2. Server Integration

Registration is automatic. `RegisterRoutes` will create the `POST /files/{path...}` endpoint.

**File: `web/server.go`**
```go
func main() {
    cp := crudp.New(binary.Encode, binary.Decode)
    cp.RegisterHandler(&files.Handler{})
    
    mux := http.NewServeMux()
    cp.RegisterRoutes(mux) // Registers POST /files/{path...}
    
    http.ListenAndServe(":8080", mux)
}
```

## 3. Client Usage (HTML Form)

You can now upload files using a standard HTML form or `fetch` pointed directly at the handler endpoint.

```html
<form action="/files" method="POST" enctype="multipart/form-data">
    <input type="file" name="file">
    <button type="submit">Upload</button>
</form>
```

## Key Benefits

- **Simplified Routing**: No need for `HttpRouteProvider` or custom route registration.
- **Unified Logic**: All "Creation" logic (whether metadata or physical file) lives in the `Create` method.
- **Isomorphic Ready**: The same handler can process both raw uploads on the server and JSON metadata objects from WASM.