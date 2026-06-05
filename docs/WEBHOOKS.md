# Webhook Handling

This guide demonstrates how to receive and process webhooks (from Stripe, GitHub, etc.) using CRUDP's automatic endpoints.

## Handler Implementation

A webhook is just a `POST` request. Your handler receives `*http.Request` to access headers (for signature verification) and the raw body.

```go
package webhooks

import (
    "encoding/json"
    "io"
    "net/http"
)

// WebhookEvent is the entity for webhook handling
type WebhookEvent struct {
    Type string          `json:"type"`
    Data json.RawMessage `json:"data"`
}

func (w *WebhookEvent) HandlerName() string { return "webhooks" }

func (w *WebhookEvent) ValidateData(action byte, payload any) error { return nil }

// Access control (see [ACCESS_CONTROL.md](./ACCESS_CONTROL.md))
func (w *WebhookEvent) AllowedRoles(action byte) []byte { return []byte{'*'} } // Webhooks from any authenticated source

func (w *WebhookEvent) Create(payload any) (any, error) {
    // In this example we assume the payload is a struct holding the required context
    // or we fetch the *http.Request directly if it's passed as the payload
    r, ok := payload.(*http.Request)
    if !ok {
        return nil, errors.New("expected http.Request payload")
    }

    provider := r.URL.Path // Simplify getting provider from path or query params

    // 1. Read raw body for signature verification
    body, _ := io.ReadAll(r.Body)

    // 2. Verify signature based on provider
    switch provider {
    case "stripe":
        sig := r.Header.Get("Stripe-Signature")
        if !verifyStripeSignature(body, sig) {
            return errors.New("invalid stripe signature")
        }
    case "github":
        sig := r.Header.Get("X-Hub-Signature-256")
        if !verifyGitHubSignature(body, sig) {
            return errors.New("invalid github signature")
        }
    }

    // 3. Parse and process event
    var event WebhookEvent
    json.Unmarshal(body, &event)

    return handleEvent(provider, event)
}
```

## Resulting Routes

| Provider | URL | Handler Method |
|----------|-----|----------------|
| Stripe | `POST /webhooks/stripe` | Create |
| GitHub | `POST /webhooks/github` | Create |
| Generic | `POST /webhooks` | Create |

## Server Registration

```go
cp := crudp.New()
cp.RegisterHandlers(&webhooks.WebhookEvent{})

mux := http.NewServeMux()
cp.RegisterRoutes(mux) // Registers POST /webhooks/{path...}

http.ListenAndServe(":8080", mux)
```

## Key Points

- **Path Routing**: The `{path...}` captures the provider name (e.g., `stripe`, `github`).
- **Signature Verification**: Use `*http.Request` to access headers and raw body.
- **No Custom Routes Needed**: Webhooks are just Create operations.
