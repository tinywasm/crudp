# Plan — Refactor de `crudp` al arnés tipado (router + fmt.Encodable/Decodable)

> `crudp` expone un endpoint CRUD/batch sobre HTTP. Tiene dos deudas técnicas
> independientes que este plan aborda en dos fases:
> - **Fase 1 (transporte):** handlers atados a `net/http` → `router.Context` / `router.Router`. Consecuencia directa: `http_stlib.go` (actualmente `//go:build !wasm` solo por `net/http`) pierde su build tag y se renombra `routes.go` — es código isomórfico.
> - **Fase 2 (datos):** `any` en toda la capa de datos → `fmt.Encodable` / `fmt.Decodable` (patrón `tinywasm/json`)

---

## Reglas del arnés (extracto obligatorio)

Del AGENTS.md de esta librería y de `tinywasm/app/docs/CONSTRUCTION_HARNESS.md`:

1. **Tipado sobre `any`** — ninguna firma pública acepta ni devuelve `any`; el tipo incorrecto no compila.
2. **Explícito sobre implícito** — el nombre del método declara la intención; sin magia oculta.
3. **Estados ilegales irrepresentables** — existe exactamente un camino tipado para cada operación.
4. **Fallar en compilación, no en ejecución** — sin aserciones de tipo en producción.
5. **Superficie mínima** — solo lo que el consumidor necesita tipear; internos son privados.

---

## Contratos que consume (reexpresados para ser autocontenidos)

```go
// github.com/tinywasm/fmt
type Encodable interface { EncodeFields(w FieldWriter); IsNil() bool }
type Decodable interface { DecodeFields(r FieldReader); IsNil() bool }
type FielderSlice interface { Len() int; At(i int) Fielder; Append() Fielder }

// github.com/tinywasm/router
type Context interface {
    Method() string; Path() string; Body() []byte
    GetHeader(k string) string; SetHeader(k, v string)
    WriteStatus(code int); Write([]byte) (int, error)
    SetValue(key string, v any); Value(key string) any
}
type HandlerFunc func(Context)
type Middleware func(HandlerFunc) HandlerFunc
type Router interface {
    Get(path string, h HandlerFunc); Post(path string, h HandlerFunc)
    Handle(method, path string, h HandlerFunc); Use(m ...Middleware)
}
```

---

## FASE 1 — Transporte: `net/http` → `router`

### Estado de partida (Fase 1)

- `RegisterRoutes(mux *http.ServeMux)`
- `handleBatch(w http.ResponseWriter, r *http.Request)` / `handleSingle(w, r, …)`
- `makeHandler(…) http.HandlerFunc`
- `ApplyMiddleware(handler http.Handler) http.Handler`

### Cambios Fase 1

| Antes (`net/http`) | Después (`router`) |
|---|---|
| `RegisterRoutes(mux *http.ServeMux)` | `RegisterRoutes(r router.Router)` |
| `handleBatch(w, r)` | handler `func(router.Context)` — lee `ctx.Body()`, escribe con `ctx.Write`/`ctx.WriteStatus` |
| `makeHandler(…) http.HandlerFunc` | `makeHandler(…) router.HandlerFunc` |
| `ApplyMiddleware(http.Handler) http.Handler` | `router.Middleware` aplicado con `r.Use(...)` |
| `inject ...any` en Execute (ctx, r http.Request) | `ctx router.Context` — el contexto de petición ya carga los datos; roles se leen via `ctx.Value` |

---

## FASE 2 — Datos: `any` → `fmt.Encodable` / `fmt.Decodable`

Esta fase elimina `reflect` y el uso de `any` en la capa de datos/contratos, alineándose con el patrón `tinywasm/json`.

### Mapa completo de `any` a eliminar

| Archivo | Símbolo | Problema | Solución |
|---|---|---|---|
| `interfaces.go` | `Creator.Create(payload any) (any, error)` | `any` en entrada y salida | `Create(payload fmt.Decodable) (fmt.Encodable, error)` |
| `interfaces.go` | `Reader.Read(id string) (any, error)` | `any` en retorno | `Read(id string) (fmt.Encodable, error)` |
| `interfaces.go` | `Reader.List() (any, error)` | retorno genérico (slice implícita) | `List() (fmt.Encodable, error)` — `fmt.Encodable` con `FielderSlice` maneja slice automáticamente |
| `interfaces.go` | `Updater.Update(payload any) (any, error)` | ídem Creator | `Update(payload fmt.Decodable) (fmt.Encodable, error)` |
| `interfaces.go` | `DataValidator.ValidateData(action byte, payload any) error` | `any` en payload | `ValidateData(action byte, payload fmt.Decodable) error` |
| `crudp.go` | `actionHandler.handler any` | `any` interno | eliminar: no se necesita con factory tipada (ver abajo) |
| `crudp.go` | `actionHandler.Create func(payload any) (any, error)` | funciones internas `any` | `Create func(fmt.Decodable) (fmt.Encodable, error)` — mismo cambio en todos los campos CRUD |
| `crudp.go` | `actionHandler.dataType reflect.Type` | reflejo para instanciar el payload | sustituir por `newPayload func() fmt.Decodable` — factory tipada; elimina import `reflect` |
| `crudp.go` | `encode func(input any, output any) error` | codec genérico | `encode func(fmt.Encodable, *[]byte) error` |
| `crudp.go` | `decode func(input any, output any) error` | codec genérico | `decode func([]byte, fmt.Decodable) error` |
| `crudp.go` | `SetCodecs(encode, decode func(any, any) error)` | firma pública con `any` | `SetCodecs(encode func(fmt.Encodable, *[]byte) error, decode func([]byte, fmt.Decodable) error)` |
| `crudp.go` | `getUserRoles func(data ...any) []byte` | inyección por `any` | `getUserRoles func(ctx router.Context) []byte` — roles viven en `ctx.Value("roles")` |
| `crudp.go` | `accessCheckFn func(resource string, action byte, data ...any) bool` | variadic `any` | `accessCheckFn func(resource string, action byte, ctx router.Context) bool` |
| `crudp.go` | `accessCheck func(handler actionHandler, action byte, data ...any) error` | variadic `any` | `accessCheck func(handler actionHandler, action byte, ctx router.Context) error` |
| `handlers.go` | `RegisterHandlers(handlers ...any)` | aceptar cualquier tipo | `RegisterHandlers(handlers ...NamedHandler)` — `NamedHandler` es el contrato mínimo; aserciones internas para `Creator`, `Reader`, etc. siguen igual pero ya no requieren reflect para el tipo de dato |
| `handlers.go` | `CallHandler(handlerID uint8, action byte, data ...any) (any, error)` | despacho genérico | `callHandler(handlerID uint8, action byte, payload fmt.Decodable, ctx router.Context) (fmt.Encodable, error)` — privado, tipado |
| `handlers.go` | `decodeWithKnownType(p, id) ([]any, error)` | slice de `any` | `decodePayload(p *Packet, id uint8) (fmt.Decodable, error)` usando `newPayload` factory |
| `handlers.go` | `decodeWithRawBytes(p) ([]any, error)` | slice de `any` | eliminar: sin tipo conocido no hay decode tipado; el handler debe siempre registrar su tipo |
| `execute.go` | `Execute(req, inject ...any)` | inject como `any` | `Execute(req *BatchRequest, ctx router.Context) (*BatchResponse, error)` |
| `execute.go` | `executeSingle(p, inject ...any)` | ídem | `executeSingle(p *Packet, ctx router.Context) PacketResult` |
| `execute.go` | `encodeResult(pr, result any) error` | `any` en result | `encodeResult(pr *PacketResult, result fmt.Encodable) error` — `reflect` eliminado: `fmt.Encodable` con `FielderSlice` maneja slices sin reflexión |

### Eliminaciones consecuentes

- **`import "reflect"`** en `execute.go` y `crudp.go` — ya no necesario.
- **`decodeWithRawBytes`** — eliminar: sin factory tipada no hay decodificación segura. Los handlers sin tipo registrado devuelven error explícito.
- **`actionHandler.handler any`** — eliminar campo, ya no referenciado.

### Registro de la factory tipada

Cada handler registrado debe proveer `NewPayload() fmt.Decodable` (o se añade al contrato interno):

```go
// Contrato interno (no exportado como interfaz):
// el campo newPayload en actionHandler se poblará desde el handler si implementa:
type payloadFactory interface {
    NewPayload() fmt.Decodable
}
// RegisterHandlers detecta esto via aserción en el momento del registro.
```

Esto es la única aserción de tipo permitida, y ocurre en `RegisterHandlers` (tiempo de inicialización), no en el path caliente de ejecución.

### Ejemplo del contrato tipado resultante

```go
// Implementador (módulo externo)
type UserHandler struct{ db *orm.DB }

func (h *UserHandler) HandlerName() string               { return "user" }
func (h *UserHandler) NewPayload() fmt.Decodable         { return &User{} }
func (h *UserHandler) Create(p fmt.Decodable) (fmt.Encodable, error) {
    u := p.(*User)
    return h.db.Insert(u)
}
func (h *UserHandler) Read(id string) (fmt.Encodable, error) {
    return h.db.FindByID(id)
}
func (h *UserHandler) List() (fmt.Encodable, error) {
    return h.db.FindAll() // devuelve fmt.Encodable que implementa FielderSlice
}

// Registro
cp := crudp.New()
cp.SetCodecs(json.EncodeBytes, json.DecodeBytes)
cp.RegisterHandlers(&UserHandler{db: db})
cp.RegisterRoutes(r) // r es router.Router
```

---

## Pasos de implementación

### Fase 1 (transporte)
1. Añadir `github.com/tinywasm/router` a `go.mod`.
2. Reescribir `RegisterRoutes` para registrar sobre `router.Router`.
3. Migrar `handleBatch`/`handleSingle`/`makeHandler` a `router.Context`/`router.HandlerFunc`.
4. Convertir `ApplyMiddleware` en `router.Middleware` aplicado con `r.Use(...)`.
5. Cambiar `Execute(req, inject ...any)` a `Execute(req, ctx router.Context)`.
6. Actualizar `SetUserRoles` y `SetAccessCheck` a firmas con `router.Context`.
7. **Eliminar build tag de `http_stlib.go` y renombrar a `routes.go`** — al no haber más `net/http` el archivo es isomórfico.
8. Extraer `dispatchLocal` en `execute_front.go`: `HandleResponse` deja de llamar `Execute`; usa path interno sin contexto HTTP.

### Fase 2 (datos)
7. Añadir `github.com/tinywasm/fmt` a `go.mod`.
8. Actualizar `interfaces.go`: reemplazar `any` por `fmt.Decodable`/`fmt.Encodable`.
9. Añadir interfaz interna `payloadFactory` y campo `newPayload func() fmt.Decodable` en `actionHandler`.
10. Actualizar `actionHandler` campos `Create`, `Read`, `List`, `Update`, `ValidateData` a firmas tipadas.
11. Actualizar `SetCodecs` a firma tipada; actualizar `encode`/`decode` fields.
12. Reescribir `decodePayload` (antes `decodeWithKnownType`) usando factory tipada; eliminar `decodeWithRawBytes`.
13. Reescribir `encodeResult` usando `fmt.Encodable` directamente; eliminar `reflect`.
14. Reescribir `callHandler` (privado, tipado) como reemplazo de `CallHandler(...any)`.
15. Cambiar `RegisterHandlers(...any)` a `RegisterHandlers(handlers ...NamedHandler)`.
16. Eliminar `import "reflect"` de `execute.go` y `crudp.go`.

---

## Isomorfismo: build tags tras el refactor

El objetivo del arnés es compilar el mismo código en ambos targets. Tras Fase 1 + Fase 2:

| Archivo actual | Build tag actual | Razón del tag | Tras refactor |
|---|---|---|---|
| `http_stlib.go` | `//go:build !wasm` | Importa `net/http` | **Eliminar tag** → renombrar `routes.go`; usa solo `router.Router`/`router.Context` (interfaces, compilan en ambos) |
| `client_wasm.go` | `//go:build wasm` | Usa `tinywasm/fetch` (API de browser) | **Permanece `wasm`** — razón legítima: fetch no existe en servidor |
| `execute_front.go` | `//go:build wasm` | `HandleResponse` procesa respuesta del servidor para actualizar DOM | **Permanece `wasm`** — razón legítima: flujo cliente sin contexto HTTP |

### Consecuencia sobre `execute_front.go`

`HandleResponse` llama actualmente `cp.Execute(req)` sin contexto. Después de Fase 1, `Execute(req, ctx router.Context)` requiere contexto. En el cliente wasm no existe contexto HTTP.

**Solución:** `HandleResponse` no llama `Execute` — llama un método interno privado `dispatchLocal(p *Packet)` que despacha directo al handler sin pasar por el path HTTP (sin decode de body, sin access check, sin encode de response). El dato ya viene decodificado del servidor en `PacketResult.Data`; el handler cliente lo usa para actualizar el DOM.

```go
// execute_front.go (//go:build wasm)
func (cp *CrudP) HandleResponse(resp *BatchResponse) {
    for _, res := range resp.Results {
        cp.dispatchLocal(&res) // privado, sin router.Context
    }
}
```

Esto es correcto por diseño: el cliente no re-ejecuta la lógica de servidor (no hay access check, no hay encode/decode de red), solo notifica al handler con los datos ya listos.

**El mismo struct `CrudP` funciona en ambos targets** — lo que difiere es el método de entrada:
- Servidor (`!wasm`): `RegisterRoutes(r router.Router)` → recibe peticiones HTTP via router
- Cliente (`wasm`): `InitClient()` + `HandleResponse(resp)` → envía fetch, procesa respuesta

---

## Estrategia de pruebas y criterios de aceptación

- **Sin `net/http` en superficie pública:** ninguna firma exportada nombra tipos de `net/http`. Verificable por `grep`.
- **Sin `any` en firmas exportadas:** `grep -n "any" interfaces.go crudp.go handlers.go execute.go` no muestra `any` en posición de tipo de parámetro/retorno exportado. (El `any` en `ctx.Value` del contrato de `router` es una excepción justificada por limitación del sistema de tipos de Go.)
- **Sin `import "reflect"`:** verificable tras Fase 2.
- **Batch y single** siguen funcionando: tests que envían un `Packet` con body JSON y verifican la respuesta escrita en el `Context`.
- **Middleware por contrato:** acceso denegado corta la cadena; test verifica con `router.Middleware`.
- **Factory tipada:** test de registro con handler que implementa `payloadFactory`; decode produce el tipo correcto sin aserción en path caliente.
- **Sin build tag en `routes.go`:** `go build ./...` en target wasm no falla — el archivo compila en ambos targets.
- **`execute_front.go` no llama `Execute`:** `HandleResponse` usa `dispatchLocal` interno; no existe dependencia del contexto HTTP en el path cliente.
