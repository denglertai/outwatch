package config

import (
	"strings"
	"testing"
)

func TestValidate_Valid_Logback_Config(t *testing.T) {
	cfg := &FileConfig{
		Target:   "logback",
		FileName: "app-loggers.xml",
		Logback: &LogbackConfig{
			Loggers: map[string]string{
				"com.example":       "INFO",
				"org.hibernate.SQL": "DEBUG",
			},
		},
	}

	err := cfg.Validate()
	if err != nil {
		t.Fatalf("validation failed: %v", err)
	}

	if cfg.Target != "logback" {
		t.Errorf("target should be normalized, got %q", cfg.Target)
	}
}

func TestValidate_Target_Required(t *testing.T) {
	cfg := &FileConfig{
		Target:   "",
		FileName: "config.xml",
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for missing target")
	}
	if err.Error() != "target is required" {
		t.Errorf("wrong error: %v", err)
	}
}

func TestValidate_Target_Whitespace_Only(t *testing.T) {
	cfg := &FileConfig{
		Target:   "   ",
		FileName: "config.xml",
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for whitespace-only target")
	}
}

func TestValidate_Filename_Required(t *testing.T) {
	cfg := &FileConfig{
		Target:   "logback",
		FileName: "",
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for missing filename")
	}
	if err.Error() != "filename is required" {
		t.Errorf("wrong error: %v", err)
	}
}

func TestValidate_Filename_With_Forward_Slash(t *testing.T) {
	cfg := &FileConfig{
		Target:   "logback",
		FileName: "subdir/config.xml",
		Logback: &LogbackConfig{
			Loggers: map[string]string{"root": "INFO"},
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for filename with path separator")
	}
	if err.Error() != "filename must be a file name only, without path segments" {
		t.Errorf("wrong error: %v", err)
	}
}

func TestValidate_Filename_With_Backslash(t *testing.T) {
	cfg := &FileConfig{
		Target:   "logback",
		FileName: "subdir\\config.xml",
		Logback: &LogbackConfig{
			Loggers: map[string]string{"root": "INFO"},
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for filename with path separator")
	}
	if err.Error() != "filename must be a file name only, without path segments" {
		t.Errorf("wrong error: %v", err)
	}
}

func TestValidate_Filename_Is_Dot(t *testing.T) {
	cfg := &FileConfig{
		Target:   "logback",
		FileName: ".",
		Logback: &LogbackConfig{
			Loggers: map[string]string{"root": "INFO"},
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for filename '.'")
	}
	if err.Error() != "filename must be a valid file name" {
		t.Errorf("wrong error: %v", err)
	}
}

func TestValidate_Filename_Is_DotDot(t *testing.T) {
	cfg := &FileConfig{
		Target:   "logback",
		FileName: "..",
		Logback: &LogbackConfig{
			Loggers: map[string]string{"root": "INFO"},
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for filename '..'")
	}
	if err.Error() != "filename must be a valid file name" {
		t.Errorf("wrong error: %v", err)
	}
}

func TestValidate_Logback_Section_Required(t *testing.T) {
	cfg := &FileConfig{
		Target:   "logback",
		FileName: "config.xml",
		Logback:  nil,
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for missing logback section")
	}
	if err.Error() != `logback section is required when target is "logback"` {
		t.Errorf("wrong error: %v", err)
	}
}

func TestValidate_Logback_Loggers_Empty(t *testing.T) {
	cfg := &FileConfig{
		Target:   "logback",
		FileName: "config.xml",
		Logback: &LogbackConfig{
			Loggers: map[string]string{},
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for empty loggers")
	}
	if err.Error() != "logback.loggers must contain at least one entry" {
		t.Errorf("wrong error: %v", err)
	}
}

func TestValidate_Logger_Name_Empty(t *testing.T) {
	cfg := &FileConfig{
		Target:   "logback",
		FileName: "config.xml",
		Logback: &LogbackConfig{
			Loggers: map[string]string{
				"": "INFO",
			},
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for empty logger name")
	}
	if err.Error() != "logger name must not be empty" {
		t.Errorf("wrong error: %v", err)
	}
}

func TestValidate_Logger_Name_Whitespace_Only(t *testing.T) {
	cfg := &FileConfig{
		Target:   "logback",
		FileName: "config.xml",
		Logback: &LogbackConfig{
			Loggers: map[string]string{
				"   ": "INFO",
			},
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for whitespace-only logger name")
	}
	if err.Error() != "logger name must not be empty" {
		t.Errorf("wrong error: %v", err)
	}
}

func TestValidate_Invalid_Log_Level(t *testing.T) {
	cfg := &FileConfig{
		Target:   "logback",
		FileName: "config.xml",
		Logback: &LogbackConfig{
			Loggers: map[string]string{
				"com.example": "INVALID",
			},
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for invalid log level")
	}
	if err.Error() != `invalid level "INVALID" for logger "com.example"` {
		t.Errorf("wrong error: %v", err)
	}
}

func TestValidate_All_Valid_Levels(t *testing.T) {
	levels := []string{"TRACE", "DEBUG", "INFO", "WARN", "ERROR", "OFF"}

	for _, level := range levels {
		t.Run(level, func(t *testing.T) {
			cfg := &FileConfig{
				Target:   "logback",
				FileName: "config.xml",
				Logback: &LogbackConfig{
					Loggers: map[string]string{
						"test": level,
					},
				},
			}

			err := cfg.Validate()
			if err != nil {
				t.Fatalf("validation failed for level %s: %v", level, err)
			}
		})
	}
}

func TestValidate_Case_Insensitive_Levels(t *testing.T) {
	cases := []string{"trace", "Debug", "info", "WARN", "ErRoR", "off"}

	for i, level := range cases {
		cfg := &FileConfig{
			Target:   "logback",
			FileName: "config.xml",
			Logback: &LogbackConfig{
				Loggers: map[string]string{
					"test": level,
				},
			},
		}

		err := cfg.Validate()
		if err != nil {
			t.Fatalf("validation failed for case variant %q: %v", level, err)
		}

		expected := strings.ToUpper(level)
		if cfg.Logback.Loggers["test"] != expected {
			t.Errorf("case %d: expected %s to be normalized to %s, got %s", i, level, expected, cfg.Logback.Loggers["test"])
		}
	}
}

func TestValidate_Target_Case_Normalization(t *testing.T) {
	cfg := &FileConfig{
		Target:   "LOGBACK",
		FileName: "config.xml",
		Logback: &LogbackConfig{
			Loggers: map[string]string{"root": "INFO"},
		},
	}

	err := cfg.Validate()
	if err != nil {
		t.Fatalf("validation failed: %v", err)
	}

	if cfg.Target != "logback" {
		t.Errorf("target should be normalized to lowercase, got %q", cfg.Target)
	}
}

func TestValidate_Unknown_Target_No_Extra_Section(t *testing.T) {
	cfg := &FileConfig{
		Target:   "custom",
		FileName: "config.xml",
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for unknown target without section")
	}
	if err.Error() != `section "custom" is required when target is "custom"` {
		t.Errorf("wrong error: %v", err)
	}
}

func TestValidate_Unknown_Target_With_Extra_Section(t *testing.T) {
	cfg := &FileConfig{
		Target:   "custom",
		FileName: "config.xml",
		Extra: map[string]any{
			"custom": map[string]string{
				"key": "value",
			},
		},
	}

	err := cfg.Validate()
	if err != nil {
		t.Fatalf("validation failed: %v", err)
	}
}

func TestSortedLoggerNames(t *testing.T) {
	loggers := map[string]string{
		"zebra":  "INFO",
		"apple":  "DEBUG",
		"middle": "WARN",
		"banana": "ERROR",
	}

	names := SortedLoggerNames(loggers)

	expected := []string{"apple", "banana", "middle", "zebra"}
	if len(names) != len(expected) {
		t.Fatalf("expected %d names, got %d", len(expected), len(names))
	}

	for i, exp := range expected {
		if names[i] != exp {
			t.Errorf("position %d: expected %s, got %s", i, exp, names[i])
		}
	}
}

func TestSortedLoggerNames_Empty(t *testing.T) {
	loggers := map[string]string{}

	names := SortedLoggerNames(loggers)

	if len(names) != 0 {
		t.Errorf("expected 0 names, got %d", len(names))
	}
}
