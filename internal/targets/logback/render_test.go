// Package logback tests logback target rendering behavior.
package logback

import (
	"strings"
	"testing"

	"github.com/denglertai/outwatch/internal/config"
)

// TestRender_DeterministicOrder verifies stable sorted logger output in XML.
func TestRender_DeterministicOrder(t *testing.T) {
	r := Renderer{}
	payload, err := r.Render(config.OutputConfig{
		Target: "logback",
		File:   "x.xml",
		Loggers: map[string]string{
			"b.pkg": "DEBUG",
			"a.pkg": "INFO",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	xml := string(payload)
	a := strings.Index(xml, `name="a.pkg"`)
	b := strings.Index(xml, `name="b.pkg"`)
	if a == -1 || b == -1 || a > b {
		t.Fatalf("expected sorted logger output, got %s", xml)
	}
}
