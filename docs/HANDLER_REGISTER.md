# Entity Registration

## Overview

In CRUDP, your data models (**Entities**) implement the business logic. They are registered with a `CrudP` instance, which then dispatches incoming requests based on the `HandlerID` provided in the protocol packets.

For a complete step-by-step example, see the [Integration Guide](./INTEGRATION_GUIDE.md).

## CRUD Interfaces

Entities implement one or more of the CRUD interfaces defined in [`interfaces.go`](../interfaces.go):

- `Creator`: `Create(data ...any) any`
- `Reader`: `Read(data ...any) any`
- `Updater`: `Update(data ...any) any`
- `Deleter`: `Delete(data ...any) any`

**Key Points:**
- **Return types**: Returning an `error` allows CRUDP to automatically populate error messages in the response.
- **Dynamic Results**: Results can be structs, slices, or primitives.

## Mandatory Interfaces

**Security and Identification**: Every Entity that implements at least one CRUD operation **must** also implement:

1.  **`NamedHandler`**: Provides the unique name.
    ```go
    func (u *User) HandlerName() string { return "users" }
    ```
2.  **`DataValidator`**: Validates data before the action.
    ```go
    func (u *User) ValidateData(action byte, data ...any) error { return nil }
    ```

`RegisterHandlers` will return an error if a CRUD Entity is missing these implementations.

## Registration

Use `RegisterHandlers` to register Entity instances. The order in the slice determines the `HandlerID`.

```go
err := cp.RegisterHandlers(&User{}, &Product{})
```
