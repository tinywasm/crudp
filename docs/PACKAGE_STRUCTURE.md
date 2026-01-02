# Package Structure

CRUDP is designed for maximum efficiency in both standard Go and TinyGo (WASM).

## Main Components

| File | Responsibility |
|------|----------------|
| `crudp.go` | Core `CrudP` struct, initialization (mandatory serialization), and logger management. |
| `handlers.go` | Handler registration, name resolution (reflection), and method binding. |
| `interfaces.go` | Definition of CRUD, Validation, and Naming interfaces. |
| `packet.go` | Protocol data structures (`Packet`, `BatchRequest`, `BatchResponse`). |
| `actions.go` | Helper functions to map HTTP methods to CRUD actions (`c`, `r`, `u`, `d`). |
| `execute.go` | Main execution engine for `BatchRequest`. |
| `http_stlib.go` | Standard library HTTP integration and custom routes/middleware hooks. |

## Related Documentation

- [ARCHITECTURE.md](ARCHITECTURE.md) - High-level design.
- [INITIALIZATION.md](INITIALIZATION.md) - How to initialize and configure serialization.
- [HANDLER_REGISTER.md](HANDLER_REGISTER.md) - How to implement modules.