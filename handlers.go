package crudp

import (
	"reflect"

	. "github.com/tinywasm/fmt"
)

// RegisterHandlers prepares the shared handler table
func (cp *CrudP) RegisterHandlers(handlers ...any) error {
	cp.handlers = make([]actionHandler, len(handlers))

	for i, h := range handlers {
		if h == nil {
			return Errf("handler %d is nil", i)
		}

		ah := actionHandler{
			index:   uint8(i),
			handler: h,
		}

		// Bind CRUD methods and track if any are implemented
		hasCRUD := false
		if creator, ok := h.(Creator); ok {
			ah.Create = creator.Create
			hasCRUD = true
		}
		if reader, ok := h.(Reader); ok {
			ah.Read = reader.Read
			hasCRUD = true
		}
		if updater, ok := h.(Updater); ok {
			ah.Update = updater.Update
			hasCRUD = true
		}
		if deleter, ok := h.(Deleter); ok {
			ah.Delete = deleter.Delete
			hasCRUD = true
		}

		if hasCRUD {
			// Enforce NamedHandler
			named, ok := h.(NamedHandler)
			if !ok {
				return Errf("missing interface: 'HandlerName() string' for handler at index %d", i)
			}
			ah.name = named.HandlerName()

			// Enforce DataValidator
			if validator, ok := h.(DataValidator); ok {
				ah.ValidateData = validator.ValidateData
			} else {
				return Errf("missing interface: 'ValidateData(action byte, data ...any) error' for handler: %s", ah.name)
			}

			// Enforce AccessLevel (optional when SetAccessCheck is configured)
			if access, ok := h.(AccessLevel); ok {
				ah.AllowedRoles = access.AllowedRoles

				// Security-by-default: validate all implemented actions have roles defined.
				// Skipped when accessCheckFn is set — the external hook owns access decisions.
				if cp.accessCheckFn == nil {
					for _, action := range []byte{'c', 'r', 'u', 'd'} {
						// Only validate actions that are implemented
						implemented := false
						switch action {
						case 'c':
							implemented = ah.Create != nil
						case 'r':
							implemented = ah.Read != nil
						case 'u':
							implemented = ah.Update != nil
						case 'd':
							implemented = ah.Delete != nil
						}

						if implemented {
							roles := ah.AllowedRoles(action)
							if len(roles) == 0 {
								return Errf("security error: AllowedRoles('%c') returned nil/empty for handler: %s (each action must define at least one role)", action, ah.name)
							}
						}
					}
				}
			} else if cp.accessCheckFn == nil {
				return Errf("missing interface: 'AllowedRoles(action byte) []byte' for handler: %s", ah.name)
			}

			// Validate AllowedAccess doesn't return -1 or invalid for implemented actions
			// Actually the plan says it must return non-nil if it was slice, but now it is int.
			// For int, level 0 might be "no access".
			// Let's check some actions to ensure it doesn't return something that means "not configured"
			// but since it's an interface call, we'll just cache it.

			// Cache the type for decode (Option A: Optimized Reflect)
			t := reflect.TypeOf(h)
			if t.Kind() == reflect.Ptr {
				t = t.Elem()
			}
			ah.dataType = t
		}

		cp.handlers[i] = ah
		if ah.name != "" {
			cp.log("registered handler:", ah.name, "at index", i)
		}
	}

	// Security: If CRUD handlers are registered but no access control is configured,
	// and we are NOT in dev mode, it's a security risk.
	if !cp.devMode && cp.getUserRoles == nil && cp.accessCheckFn == nil {
		hasCRUD := false
		for _, ah := range cp.handlers {
			if ah.AllowedRoles != nil {
				hasCRUD = true
				break
			}
		}
		if hasCRUD {
			return Errf("security error: CRUD handlers registered but SetUserRoles() not configured")
		}
	}

	return nil
}

// GetHandlerName returns the handler name by its ID
func (cp *CrudP) GetHandlerName(handlerID uint8) string {
	if int(handlerID) >= len(cp.handlers) {
		return ""
	}
	return cp.handlers[handlerID].name
}

// CallHandler searches and calls the handler directly by shared index
func (cp *CrudP) CallHandler(handlerID uint8, action byte, data ...any) (any, error) {
	if int(handlerID) >= len(cp.handlers) {
		return nil, Errf("no handler found for id: %d", handlerID)
	}

	handler := cp.handlers[handlerID]

	// 1. Access Control (first step)
	if err := cp.accessCheck(handler, action, data...); err != nil {
		return nil, err
	}

	// 2. Mandatory validation before executing
	if handler.ValidateData != nil {
		if err := handler.ValidateData(action, data...); err != nil {
			return nil, err
		}
	}

	var result any
	implemented := false
	switch action {
	case 'c':
		if handler.Create != nil {
			result = handler.Create(data...)
			implemented = true
		}
	case 'r':
		if handler.Read != nil {
			result = handler.Read(data...)
			implemented = true
		}
	case 'u':
		if handler.Update != nil {
			result = handler.Update(data...)
			implemented = true
		}
	case 'd':
		if handler.Delete != nil {
			result = handler.Delete(data...)
			implemented = true
		}
	default:
		return nil, Errf("unknown action '%c' for handler: %s", action, handler.name)
	}

	if !implemented {
		return nil, Errf("action '%c' not implemented for handler: %s", action, handler.name)
	}

	if result == nil {
		return nil, nil
	}

	// Detect error in result for backward compatibility with server expectations
	if err, ok := result.(error); ok {
		return nil, err
	}

	return result, nil
}

// decodeWithKnownType decodes packet data using cached type information
func (cp *CrudP) decodeWithKnownType(p *Packet, handlerID uint8) ([]any, error) {
	if int(handlerID) >= len(cp.handlers) {
		return nil, Errf("no handler found for id: %d", handlerID)
	}

	handler := cp.handlers[handlerID]
	if handler.dataType == nil {
		return cp.decodeWithRawBytes(p)
	}

	decodedData := make([]any, 0, len(p.Data))
	for _, itemBytes := range p.Data {
		// New instance for each item using CACHED type
		targetPtr := reflect.New(handler.dataType).Interface()

		if cp.decode == nil {
			return nil, Errf("decode function not configured")
		}

		if err := cp.decode(itemBytes, targetPtr); err != nil {
			return nil, err
		}

		decodedData = append(decodedData, targetPtr)
	}

	return decodedData, nil
}

// decodeWithRawBytes decodes packet data as raw bytes
func (cp *CrudP) decodeWithRawBytes(p *Packet) ([]any, error) {
	decodedData := make([]any, 0, len(p.Data))
	for _, itemBytes := range p.Data {
		decodedData = append(decodedData, itemBytes)
	}
	return decodedData, nil
}
