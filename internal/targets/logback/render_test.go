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

// TestRender_Name returns the correct renderer name.
func TestRender_Name(t *testing.T) {
	r := Renderer{}
	if r.Name() != "logback" {
		t.Fatalf("expected name 'logback', got %q", r.Name())
	}
}

// TestRender_Empty_Loggers generates valid XML with no loggers.
func TestRender_Empty_Loggers(t *testing.T) {
	r := Renderer{}
	payload, err := r.Render(config.OutputConfig{
		Target:  "logback",
		File:    "x.xml",
		Loggers: map[string]string{},
	})
	if err != nil {
		t.Fatalf("render empty loggers failed: %v", err)
	}

	xml := string(payload)
	if !strings.Contains(xml, "<?xml") {
		t.Fatalf("expected XML declaration, got %s", xml)
	}
	if !strings.Contains(xml, "<configuration>") {
		t.Fatalf("expected configuration element, got %s", xml)
	}
}

// TestRender_Single_Logger verifies single logger renders correctly.
func TestRender_Single_Logger(t *testing.T) {
	r := Renderer{}
	payload, err := r.Render(config.OutputConfig{
		Target:  "logback",
		File:    "x.xml",
		Loggers: map[string]string{"root": "DEBUG"},
	})
	if err != nil {
		t.Fatalf("render single logger failed: %v", err)
	}

	xml := string(payload)
	if !strings.Contains(xml, `name="root"`) {
		t.Fatalf("expected logger name in output")
	}
	if !strings.Contains(xml, `level="DEBUG"`) {
		t.Fatalf("expected logger level in output")
	}
}
