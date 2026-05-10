package discovery

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsConfigFile_YAML_Extension(t *testing.T) {
	cases := []struct {
		filename string
		expected bool
	}{
		{"config.yaml", true},
		{"config.yml", true},
		{"CONFIG.YAML", true},
		{"CONFIG.YML", true},
		{"settings.Yaml", true},
		{"settings.Yml", true},
		{"config.json", false},
		{"config.txt", false},
		{"config.xml", false},
		{"yaml", false},
		{"yml", false},
		{"", false},
	}

	for _, tc := range cases {
		t.Run(tc.filename, func(t *testing.T) {
			result := IsConfigFile(tc.filename)
			if result != tc.expected {
				t.Errorf("IsConfigFile(%q) = %v, expected %v", tc.filename, result, tc.expected)
			}
		})
	}
}

func TestDiscover_Empty_Directory(t *testing.T) {
	tmpDir := t.TempDir()

	files, err := Discover(tmpDir)
	if err != nil {
		t.Fatalf("Discover failed: %v", err)
	}

	if len(files) != 0 {
		t.Errorf("expected 0 files, got %d", len(files))
	}
}

func TestDiscover_Only_YAML_Files(t *testing.T) {
	tmpDir := t.TempDir()

	// Create some YAML files
	yamlFiles := []string{"config1.yaml", "config2.yml", "settings.yaml"}
	for _, name := range yamlFiles {
		path := filepath.Join(tmpDir, name)
		if err := os.WriteFile(path, []byte("test"), 0o644); err != nil {
			t.Fatalf("WriteFile failed: %v", err)
		}
	}

	files, err := Discover(tmpDir)
	if err != nil {
		t.Fatalf("Discover failed: %v", err)
	}

	if len(files) != 3 {
		t.Errorf("expected 3 files, got %d", len(files))
	}

	// Verify files are sorted
	for i, expected := range []string{"config1.yaml", "config2.yml", "settings.yaml"} {
		if filepath.Base(files[i]) != expected {
			t.Errorf("expected %s at position %d, got %s", expected, i, filepath.Base(files[i]))
		}
	}
}

func TestDiscover_Mixed_Files(t *testing.T) {
	tmpDir := t.TempDir()

	// Create YAML and non-YAML files
	files := []string{"config.yaml", "data.json", "settings.yml", "readme.txt", "app.yaml"}
	for _, name := range files {
		path := filepath.Join(tmpDir, name)
		if err := os.WriteFile(path, []byte("test"), 0o644); err != nil {
			t.Fatalf("WriteFile failed: %v", err)
		}
	}

	discovered, err := Discover(tmpDir)
	if err != nil {
		t.Fatalf("Discover failed: %v", err)
	}

	// Should find only YAML files
	if len(discovered) != 3 {
		t.Errorf("expected 3 YAML files, got %d", len(discovered))
	}

	// Verify only YAML files are returned
	for _, f := range discovered {
		if !IsConfigFile(filepath.Base(f)) {
			t.Errorf("non-YAML file returned: %s", f)
		}
	}
}

func TestDiscover_Sorted_Output(t *testing.T) {
	tmpDir := t.TempDir()

	// Create files in non-alphabetical order
	filenames := []string{"zebra.yaml", "apple.yml", "middle.yaml", "banana.yml"}
	for _, name := range filenames {
		path := filepath.Join(tmpDir, name)
		if err := os.WriteFile(path, []byte("test"), 0o644); err != nil {
			t.Fatalf("WriteFile failed: %v", err)
		}
	}

	discovered, err := Discover(tmpDir)
	if err != nil {
		t.Fatalf("Discover failed: %v", err)
	}

	// Expected sorted order
	expected := []string{"apple.yml", "banana.yml", "middle.yaml", "zebra.yaml"}
	if len(discovered) != len(expected) {
		t.Fatalf("expected %d files, got %d", len(expected), len(discovered))
	}

	for i, exp := range expected {
		if filepath.Base(discovered[i]) != exp {
			t.Errorf("position %d: expected %s, got %s", i, exp, filepath.Base(discovered[i]))
		}
	}
}

func TestDiscover_Ignores_Directories(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a subdirectory and some files
	subdir := filepath.Join(tmpDir, "subdir")
	if err := os.Mkdir(subdir, 0o755); err != nil {
		t.Fatalf("Mkdir failed: %v", err)
	}

	// Create files in both directories
	if err := os.WriteFile(filepath.Join(tmpDir, "root.yaml"), []byte("test"), 0o644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(subdir, "nested.yaml"), []byte("test"), 0o644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	discovered, err := Discover(tmpDir)
	if err != nil {
		t.Fatalf("Discover failed: %v", err)
	}

	// Should only find root.yaml, not nested files
	if len(discovered) != 1 {
		t.Errorf("expected 1 file, got %d", len(discovered))
	}

	if filepath.Base(discovered[0]) != "root.yaml" {
		t.Errorf("expected root.yaml, got %s", filepath.Base(discovered[0]))
	}
}

func TestDiscover_Nonexistent_Directory(t *testing.T) {
	_, err := Discover("/nonexistent/directory/that/does/not/exist")
	if err == nil {
		t.Fatal("expected error for non-existent directory")
	}
}

func TestDiscover_Case_Insensitive_Extensions(t *testing.T) {
	tmpDir := t.TempDir()

	// Create files with various case combinations
	files := []string{"config.YAML", "settings.YML", "data.Yaml", "app.Yml"}
	for _, name := range files {
		path := filepath.Join(tmpDir, name)
		if err := os.WriteFile(path, []byte("test"), 0o644); err != nil {
			t.Fatalf("WriteFile failed: %v", err)
		}
	}

	discovered, err := Discover(tmpDir)
	if err != nil {
		t.Fatalf("Discover failed: %v", err)
	}

	// Should find all files regardless of case
	if len(discovered) != 4 {
		t.Errorf("expected 4 files, got %d", len(discovered))
	}
}
