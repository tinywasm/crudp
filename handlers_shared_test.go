package crudp_test

import (
	"testing"

	"github.com/tinywasm/crudp"
)

// Test handler with explicit name
type explicitNameHandler struct{}

type ExplicitCreateResponse struct {
	Message string `json:"message"`
}

func (h *explicitNameHandler) HandlerName() string { return "my_custom_name" }
func (h *explicitNameHandler) Create(data ...any) any {
	return ExplicitCreateResponse{Message: "created"}
}
func (h *explicitNameHandler) ValidateData(action byte, data ...any) error { return nil }

// Test handler without explicit name (uses NamedHandler now)
type UserController struct{}

func (h *UserController) HandlerName() string { return "user_controller" }

type CreateResponse struct {
	ID     int    `json:"id"`
	Status string `json:"status"`
}

type ReadResponse struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func (h *UserController) Create(data ...any) any {
	return CreateResponse{ID: 1, Status: "created"}
}

func (h *UserController) Read(data ...any) any {
	return ReadResponse{ID: 1, Name: "test"}
}
func (h *UserController) ValidateData(action byte, data ...any) error { return nil }

// Handler with validation
type ValidatedHandler struct{}

func (h *ValidatedHandler) HandlerName() string { return "validated_handler" }

type ValidatedCreateResponse struct {
	Message string `json:"message"`
}

func (h *ValidatedHandler) Create(data ...any) any {
	return ValidatedCreateResponse{Message: "validated_created"}
}

func (h *ValidatedHandler) ValidateData(action byte, data ...any) error {
	if len(data) == 0 {
		return errorString("no data provided")
	}
	return nil
}

type errorString string

func (e errorString) Error() string { return string(e) }

// Shared tests
func HandlerRegistrationShared(t *testing.T, cp *crudp.CrudP) {
	t.Run("Explicit HandlerName", func(t *testing.T) {
		cp := NewTestCrudP()
		err := cp.RegisterHandlers(&explicitNameHandler{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		name := cp.GetHandlerName(0)
		if name != "my_custom_name" {
			t.Errorf("expected 'my_custom_name', got '%s'", name)
		}
	})

	t.Run("Reflection fallback removed (NamedHandler required)", func(t *testing.T) {
		cp := NewTestCrudP()
		err := cp.RegisterHandlers(&UserController{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		name := cp.GetHandlerName(0)
		if name != "user_controller" {
			t.Errorf("expected 'user_controller', got '%s'", name)
		}
	})

	t.Run("Nil Handler Error", func(t *testing.T) {
		cp := NewTestCrudP()
		err := cp.RegisterHandlers(nil)
		if err == nil {
			t.Error("expected error for nil handler")
		}
	})

	t.Run("Multiple Handlers", func(t *testing.T) {
		cp := NewTestCrudP()
		err := cp.RegisterHandlers(
			&explicitNameHandler{},
			&UserController{},
			&ValidatedHandler{},
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if cp.GetHandlerName(0) != "my_custom_name" {
			t.Error("handler 0 name mismatch")
		}
		if cp.GetHandlerName(1) != "user_controller" {
			t.Error("handler 1 name mismatch")
		}
		if cp.GetHandlerName(2) != "validated_handler" {
			t.Error("handler 2 name mismatch")
		}
	})

	t.Run("Missing DataValidator Error", func(t *testing.T) {
		// Handled in HandlerRegistrationErrorsShared
	})
}

// Helper for Missing DataValidator test
type missingValidatorHandler struct{}

func (h *missingValidatorHandler) HandlerName() string  { return "missing_validator" }
func (h *missingValidatorHandler) Read(data ...any) any { return nil }

func HandlerRegistrationErrorsShared(t *testing.T, cp *crudp.CrudP) {
	t.Run("Missing NamedHandler", func(t *testing.T) {
		cp := NewTestCrudP()

		type ReaderOnly struct{ crudp.Reader }
		// Registering a struct that implements Reader but NOT NamedHandler should fail
		err := cp.RegisterHandlers(&ReaderOnly{})
		if err == nil {
			t.Error("expected error for missing NamedHandler")
		}
	})

	t.Run("Missing DataValidator", func(t *testing.T) {
		cp := NewTestCrudP()
		err := cp.RegisterHandlers(&missingValidatorHandler{})
		if err == nil {
			t.Error("expected error for missing DataValidator")
		} else {
			expectedMsg := "missing interface: 'ValidateData"
			if !contains(err.Error(), expectedMsg) {
				t.Errorf("expected error containing %q, got %q", expectedMsg, err.Error())
			}
		}
	})
}

func contains(s, substr string) bool {
	for i := 0; i < len(s)-len(substr)+1; i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Test the module Add() pattern works with RegisterHandlers
func ModuleAddPatternShared(t *testing.T, cp *crudp.CrudP) {
	t.Run("Module Add Pattern", func(t *testing.T) {
		cp := NewTestCrudP()

		// Simulates the Add() pattern from modules
		module1 := func() []any { return []any{&explicitNameHandler{}} }
		module2 := func() []any { return []any{&UserController{}} }

		all := append(module1(), module2()...)

		err := cp.RegisterHandlers(all...)
		if err != nil {
			t.Fatalf("RegisterHandlers failed with Add pattern: %v", err)
		}

		if cp.GetHandlerName(0) != "my_custom_name" {
			t.Errorf("expected handler 0 to be 'my_custom_name', got '%s'", cp.GetHandlerName(0))
		}
		if cp.GetHandlerName(1) != "user_controller" {
			t.Errorf("expected handler 1 to be 'user_controller', got '%s'", cp.GetHandlerName(1))
		}
	})
}

func HandlerValidationShared(t *testing.T, cp *crudp.CrudP) {
	t.Run("Validation Passes", func(t *testing.T) {
		cp := NewTestCrudP()
		cp.RegisterHandlers(&ValidatedHandler{})

		result, err := cp.CallHandler(0, 'c', "some data")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if resp, ok := result.(ValidatedCreateResponse); !ok || resp.Message != "validated_created" {
			t.Errorf("expected ValidatedCreateResponse with message 'validated_created', got %v", result)
		}
	})

	t.Run("Validation Fails", func(t *testing.T) {
		cp := NewTestCrudP()
		cp.RegisterHandlers(&ValidatedHandler{})

		_, err := cp.CallHandler(0, 'c') // No data
		if err == nil {
			t.Error("expected validation error")
		}
	})

}

func CRUDOperationsShared(t *testing.T, cp *crudp.CrudP) {
	t.Run("Create Operation", func(t *testing.T) {
		cp := NewTestCrudP()
		cp.RegisterHandlers(&UserController{})

		result, err := cp.CallHandler(0, 'c', map[string]any{"name": "test"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result == nil {
			t.Error("expected result, got nil")
		}
		if _, ok := result.(CreateResponse); !ok {
			t.Errorf("expected CreateResponse, got %T", result)
		}
	})

	t.Run("Read Operation", func(t *testing.T) {
		cp := NewTestCrudP()
		cp.RegisterHandlers(&UserController{})

		result, err := cp.CallHandler(0, 'r', 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result == nil {
			t.Error("expected result, got nil")
		}
		if _, ok := result.(ReadResponse); !ok {
			t.Errorf("expected ReadResponse, got %T", result)
		}
	})

	t.Run("Unimplemented Action", func(t *testing.T) {
		cp := NewTestCrudP()
		cp.RegisterHandlers(&UserController{}) // Only has Create and Read

		_, err := cp.CallHandler(0, 'd', 1) // Delete not implemented
		if err == nil {
			t.Error("expected error for unimplemented action")
		}
	})

	t.Run("Invalid Handler ID", func(t *testing.T) {
		cp := NewTestCrudP()
		cp.RegisterHandlers(&UserController{})

		_, err := cp.CallHandler(99, 'r', 1)
		if err == nil {
			t.Error("expected error for invalid handler ID")
		}
	})
}
