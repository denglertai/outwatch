// Package targets tests renderer registry behavior.
package targets

import (
	"testing"

	"github.com/denglertai/outwatch/internal/config"
)

// fakeRenderer is a minimal renderer used for registry tests.
type fakeRenderer struct{ name string }

// Name returns the fake renderer name.
func (f fakeRenderer) Name() string { return f.name }

// Render returns a static payload for testing.
func (f fakeRenderer) Render(config.OutputConfig) ([]byte, error) {
	return []byte("ok"), nil
}

// TestRegistry_RegisterAndGet verifies successful registration and lookup.
func TestRegistry_RegisterAndGet(t *testing.T) {
	r := NewRegistry()
	if err := r.Register(fakeRenderer{name: "x"}); err != nil {
		t.Fatalf("register: %v", err)
	}
	_, err := r.Get("x")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
}

// TestRegistry_Duplicate verifies duplicate registration is rejected.
func TestRegistry_Duplicate(t *testing.T) {
	r := NewRegistry()
	_ = r.Register(fakeRenderer{name: "x"})
	if err := r.Register(fakeRenderer{name: "x"}); err == nil {
		t.Fatalf("expected duplicate registration error")
	}
}
