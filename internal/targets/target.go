// Package targets defines render target contracts and registration helpers.
package targets

import "github.com/denglertai/outwatch/internal/config"

// Renderer is implemented by output target renderers.
type Renderer interface {
	// Name returns the unique target identifier used in config (for example "logback").
	Name() string
	// Render transforms normalized output config into target bytes.
	Render(cfg config.OutputConfig) ([]byte, error)
}
