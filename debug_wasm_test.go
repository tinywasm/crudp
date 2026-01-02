//go:build wasm

package crudp_test

import (
	"testing"
)

func TestHandlerInstanceReuse(t *testing.T) {
	HandlerInstanceReuseShared(t)
}

func TestConcurrentHandlerAccess(t *testing.T) {
	ConcurrentHandlerAccessShared(t)
}
