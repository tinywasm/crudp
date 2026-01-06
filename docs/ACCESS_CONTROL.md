# Access Control (RBAC)

CRUDP implements a non-hierarchical, role-based access control (RBAC) system. Every Entity that implements CRUD operations must define its required roles for each action.

## AccessLevel Interface

Entities must implement the `AccessLevel` interface defined in [`interfaces.go`](../interfaces.go):

```go
type AccessLevel interface {
    AllowedRoles(action byte) []byte
}
```

### Security by Default

- If `AllowedRoles` returns `nil` or an empty slice `[]byte{}`, `RegisterHandlers` will return an **error**.
- Every action ('c', 'r', 'u', 'd') implemented by the handler MUST have roles defined.

### Role Conventions

While roles can be any `byte`, we recommend using intuitive ASCII characters:

| Role | Meaning | Description |
|------|---------|-------------|
| `'*'` | Authenticated | Any user with at least one role assigned |
| `'a'` | Admin | Full control over the resource |
| `'e'` | Editor | Can create, update and read |
| `'v'` | Visitor | Read-only access |

## Server Configuration

To enable access control, you must configure how CRUDP determines the current user's roles.

### 1. Set User Roles Resolver

This function is called before every action. It usually extracts roles from a JWT, session, or request context.

```go
cp.SetUserRoles(func(data ...any) []byte {
    for _, item := range data {
        if ctx, ok := item.(*context.Context); ok {
            if roles, ok := ctx.Value("user_roles").([]byte); ok {
                return roles
            }
        }
    }
    return nil // No roles (unauthenticated)
})
```

### 2. Development Mode

During development, you can bypass all security checks:

```go
cp.SetDevMode(true)
```

### 3. Access Denied Notification

You can configure a callback to receive detailed information about failed access attempts.

```go
cp.SetAccessDeniedHandler(func(handler string, action byte, userRoles []byte, allowedRoles []byte, errMsg string) {
    log.Printf("SECURITY: %s (User roles %q, needs %q)", errMsg, userRoles, allowedRoles)
})
```

## Entity Implementation

Each entity decides its own rules based on the `action` byte ('c', 'r', 'u', 'd'):

```go
func (u *User) AllowedRoles(action byte) []byte {
    switch action {
    case 'r': 
        return []byte{'v', 'e', 'a'} // Visitors, Editors and Admins can see users
    case 'c', 'u', 'd': 
        return []byte{'a'}           // Only Admins can modify users
    }
    return []byte{'a'} // Safe default
}
```

### Multiple Roles Logic (OR)

The access check uses **OR** logic. If a user has **any** of the roles returned by `AllowedRoles`, access is granted.

Example:
- Resource allows: `['d', 'm']` (dentist or medic)
- User has: `['m', 'r']` (medic and reception)
- **Result**: Access GRANTED (matches 'm').

## Security Flow

1. **Access Check**: Does `getUserRoles()` contain ANY of `AllowedRoles(action)`?
   - Special case: If `AllowedRoles` contains `'*'`, any authenticated user (non-empty roles) can access.
   - If fail: call `AccessDeniedHandler`, log generic message, return error.
2. **Data Validation**: `ValidateData(action, data)`
   - If fail: return validation error.
3. **Execution**: Execute the actual CRUD method.

## Requirements

- `SetUserRoles` is **mandatory** if any CRUD handlers are registered (unless `DevMode` is on).
- `RegisterHandlers` will return an error if an Entity implements CRUD but lacks `AllowedRoles` or returns `nil`/empty for implemented actions.
