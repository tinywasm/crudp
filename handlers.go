package crudp

import (
	"reflect"

	. "github.com/tinywasm/fmt"
)

// getHandlerName gets the handler name
// Priority: 1) HandlerName() if implemented, 2) reflection + snake_case
func getHandlerName(handler any) string {
	// First try NamedHandler interface
	if named, ok := handler.(NamedHandler); ok {
		return named.HandlerName()
	}

	// Fallback: use reflection and convert to snake_case
	t := reflect.TypeOf(handler)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// Use tinystring.SnakeLow for conversion
	return Convert(t.Name()).SnakeLow().String()
}

// RegisterHandler prepares the shared handler table
func (cp *CrudP) RegisterHandler(handlers ...any) error {
	cp.handlers = make([]actionHandler, len(handlers))

	for i, h := range handlers {
		if h == nil {
			return Errf("handler %d is nil", i)
		}

		// Get name (via interface or reflection)
		name := getHandlerName(h)

		cp.handlers[i] = actionHandler{
			name:    name,
			index:   uint8(i),
			handler: h,
		}

		cp.bind(uint8(i), h)

		cp.log("registered handler:", name, "at index", i)
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

// bind copies the CRUD functions without dynamic allocations
func (cp *CrudP) bind(index uint8, handler any) {
	if creator, ok := handler.(Creator); ok {
		cp.handlers[index].Create = creator.Create
	}
	if reader, ok := handler.(Reader); ok {
		cp.handlers[index].Read = reader.Read
	}
	if updater, ok := handler.(Updater); ok {
		cp.handlers[index].Update = updater.Update
	}
	if deleter, ok := handler.(Deleter); ok {
		cp.handlers[index].Delete = deleter.Delete
	}
}

// CallHandler searches and calls the handler directly by shared index
func (cp *CrudP) CallHandler(handlerID uint8, action byte, data ...any) (any, error) {
	if int(handlerID) >= len(cp.handlers) {
		return nil, Errf("no handler found for id: %d", handlerID)
	}

	handler := cp.handlers[handlerID]

	// Optional validation before executing
	if validator, ok := handler.handler.(Validator); ok {
		if err := validator.Validate(action, data...); err != nil {
			return nil, err
		}
	}

	switch action {
	case 'c':
		if handler.Create != nil {
			return handler.Create(data...)
		}
	case 'r':
		if handler.Read != nil {
			return handler.Read(data...)
		}
	case 'u':
		if handler.Update != nil {
			return handler.Update(data...)
		}
	case 'd':
		if handler.Delete != nil {
			return handler.Delete(data...)
		}
	}

	return nil, Errf("action '%c' not implemented for handler: %s", action, handler.name)
}

// decodeWithKnownType decodes packet data using cached type information
func (cp *CrudP) decodeWithKnownType(p *Packet, handlerID uint8) ([]any, error) {
	if int(handlerID) >= len(cp.handlers) {
		return nil, Errf("no handler found for id: %d", handlerID)
	}

	handler := cp.handlers[handlerID].handler
	if handler == nil {
		return cp.decodeWithRawBytes(p)
	}

	handlerValue := reflect.ValueOf(handler)
	handlerType := handlerValue.Type()

	var concreteType reflect.Type
	if handlerType.Kind() == reflect.Ptr {
		concreteType = handlerType.Elem()
	} else {
		concreteType = handlerType
	}

	if concreteType == nil {
		return cp.decodeWithRawBytes(p)
	}

	decodedData := make([]any, 0, len(p.Data))
	for _, itemBytes := range p.Data {
		// New instance for each item
		targetPtr := reflect.New(concreteType).Interface()

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
