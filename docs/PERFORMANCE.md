# CRUDP Performance Analysis

> **Last Updated:** January 2026

## Overview

CRUDP has been refactored to prioritize execution efficiency and modularity. By moving transport and protocol definition to separate packages, the core execution engine is now leaner and faster.

## Benchmark Results

Testing Environment:
- **CPU**: 11th Gen Intel(R) Core(TM) i7-11800H @ 2.30GHz
- **OS**: Linux (amd64)

### Core Operation Performance

| Benchmark | Time/Op | Memory/Op | Allocs/Op |
|-----------|---------|-----------|-----------|
| `BenchmarkCrudP_Setup` | ~380 ns/op | 416 B/op | 12 allocs/op |
| `BenchmarkCrudP_Execute` | ~1500 ns/op | 608 B/op | 16 allocs/op |

*Note: Results were obtained using the standard library `json` codec via `tinywasm/json`.*

## Key Efficiency Improvements

### 1. Zero-Allocation Dispatching
CRUD methods are bound at registration time. This means that calling a handler by its `HandlerID` (integer index) doesn't involve any map lookups or reflection during the execution phase.

### 2. Reduced Complexity
By removing SSE, Broker, and Transport configuration from the main `CrudP` struct, the internal complexity of `Execute` has been significantly reduced, allowing for better compiler inlining and fewer execution branches.

### 3. Explicit Errors
Returning `(any, error)` instead of just `any` avoids the overhead of internal type assertions used previously to detect error states, leading to cleaner and more predictable performance.

## Best Practices for Performance

1. **Instance Reuse**: Always reuse your `CrudP` instance. Registration is fast (~380ns), but typically done once at startup.
2. **Batching**: Large batches reduce the overhead per operation. While `Execute` itself is fast (~1.5µs per batch), grouping operations reduces total system overhead.
3. **Codec Selection**: Using `tinywasm/json` is efficient for WASM. If you need even higher performance, consider a custom binary `Codec` implementation.