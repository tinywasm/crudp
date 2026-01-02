# Initialization and Serialization

## Overview

CRUDP is designed to be as simple as possible. Since most transport-related concerns have been moved to separate packages (like `tinywasm/broker`), the initialization focuses primarily on mandatory serialization functions.

## Serialization

CRUDP requires serialization functions to convert between Go types and the raw byte slices stored in `Packet.Data`. This is handled by passing two mandatory function parameters to the constructor.

This design allows for maximum flexibility and easy integration with libraries like `tinywasm/binary` or standard `encoding/json`.

### Using `tinywasm/binary` (Recommended)

```go
import (
    "github.com/tinywasm/crudp"
    "github.com/tinywasm/binary"
)

func NewRouter() *crudp.CrudP {
    // Pass Encode and Decode functions directly
    return crudp.New(binary.Encode, binary.Decode)
}
```

## Constructors

### `New(encode EncodeFunc, decode DecodeFunc)`

The primary constructor for `CrudP`. Both parameters are mandatory to ensure that the execution engine can process packet data.

```go
type EncodeFunc func(input any, output any) error
type DecodeFunc func(input any, output any) error
```

## Other Settings (Methods)

Secondary settings are configured via methods on the `CrudP` instance to keep the constructor simple.

### Logging

Logging is disabled by default (no-op).

- `SetLog(log func(...any))`: Sets a custom logging function. Passing `nil` restores the default no-op behavior (disables logging).
