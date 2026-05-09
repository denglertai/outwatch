// Package targets defines render target contracts and registration helpers.
package targets

import "fmt"

// Registry stores renderer instances by target name.
type Registry struct {
	targets map[string]Renderer
}

// NewRegistry creates an empty target registry.
func NewRegistry() *Registry {
	return &Registry{targets: map[string]Renderer{}}
}

// Register inserts a renderer and fails on duplicate names.
func (r *Registry) Register(renderer Renderer) error {
	name := renderer.Name()
	if _, exists := r.targets[name]; exists {
		return fmt.Errorf("target %q already registered", name)
	}
	r.targets[name] = renderer
	return nil
}

// Get returns a registered renderer by name.
func (r *Registry) Get(name string) (Renderer, error) {
	target, ok := r.targets[name]
	if !ok {
		return nil, fmt.Errorf("target %q is not registered", name)
	}
	return target, nil
}
