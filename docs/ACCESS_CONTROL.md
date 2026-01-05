# Access Control (RBAC)

CRUDP implements a hierarchical, level-based access control system. Every Entity that implements CRUD operations must define its required permissions for each action.

## AccessLevel Interface

Entities must implement the `AccessLevel` interface defined in [`interfaces.go`](../interfaces.go):

```go
type AccessLevel interface {
    MinAccess(action byte) int
}
```

### Default Levels (Standard)

While levels can be any `int`, we recommend following this convention:

| Level | Meaning | Description |
|-------|---------|-------------|
| 0 | Public | No authentication required (if Resolver allows it) |
| 1 | Reader | Grant access to `Read` operations |
| 2 | Editor | Grant access to `Create`, `Update`, `Delete` |
| 255 | Admin | Full control over the resource |

## Server Configuration

To enable access control, you must configure how CRUDP determines the current user's level.

### 1. Set User Level Resolver

This function is called before every action. It usually extracts the level from the request context or headers.

```go
cp.SetUserLevel(func(data ...any) int {
    for _, item := range data {
        if ctx, ok := item.(*context.Context); ok {
            if level, ok := ctx.Value("user_level").(int); ok {
                return level
            }
        }
    }
    return 0 // Default to no access
})
```

### 2. Development Mode

During development, you can bypass all security checks:

```go
cp.SetDevMode(true)
```

### 3. Access Denied Notification

You can configure a callback to receive detailed information about failed access attempts (for logging, alerts, or audit trails).

```go
cp.SetAccessDeniedHandler(func(handler string, action byte, userLevel int, minRequired int) {
    log.Printf("SECURITY: User (level %d) tried to perform '%c' on %s (needs %d)", 
        userLevel, action, handler, minRequired)
})
```

## Entity Implementation

Each entity decides its own rules based on the `action` byte ('c', 'r', 'u', 'd'):

```go
func (u *User) MinAccess(action byte) int {
    switch action {
    case 'r': 
        return 1 // Readers and above can see users
    case 'c', 'u', 'd': 
        return 255 // Only Admins can modify users
    }
    return 255 // Default to most restrictive
}
```

## Security Flow

1. **Access Check**: `getUserLevel()` >= `MinAccess(action)`?
   - If fail: call `AccessDeniedHandler`, log generic message, return error.
2. **Data Validation**: `ValidateData(action, data)`
   - If fail: return validation error.
3. **Execution**: Execute the actual CRUD method.

## Requirements

- `SetUserLevel` is **mandatory** if any CRUD handlers are registered (unless `DevMode` is on).
- `RegisterHandlers` will return an error if an Entity implements CRUD but lacks `MinAccess`.
