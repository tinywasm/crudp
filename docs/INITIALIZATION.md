# Initialization and Serialization

## Overview

CRUDP is designed to be as simple as possible. By default, it uses `tinywasm/binary` for serialization, but you can configure custom codecs if needed.

## Serialization

CRUDP requires serialization functions to convert between Go types and the raw byte slices stored in `Packet.Data`. 

### Using `tinywasm/binary` (Default)

The standard constructor initializes CRUDP with the high-performance `tinywasm/binary` codec automatically.

```go
import (
    "github.com/tinywasm/crudp"
)

func NewRouter() *crudp.CrudP {
    // Uses tinywasm/binary by default
    return crudp.New()
}
```

### Custom Codecs

If you need to use a different format (like JSON), you can use the `SetCodecs` method.

```go
import (
    "encoding/json"
    "github.com/tinywasm/crudp"
)

func main() {
    cp := crudp.New()
    
    // Override default binary codec with JSON
    cp.SetCodecs(json.Marshal, json.Unmarshal)
}
```

## Public API

### `New()`

The primary constructor. Initializes a `CrudP` instance with:
- Default binary codec (`binary.Encode`/`binary.Decode`)
- Logging disabled (no-op)

### `SetCodecs(encode, decode func(any, any) error)`

Configures custom serialization functions.

- **encode**: Typically receives a Go struct as `input` and a `*[]byte` as `output`.
- **decode**: Typically receives a `[]byte` as `input` and a pointer to a Go struct as `output`.

### `SetLog(log func(...any))`

Configures a custom logging function. Passing `nil` restores the default no-op behavior (disables logging).
