package config

import (
	"testing"
)

func TestLoadFromBytes(t *testing.T) {
	validYAML := `
version: 1
flags:
  new_checkout:
    enabled: true
    type: "bool"
    variants:
      true: 50
      false: 50
    default: false
`

	cfg, err := LoadFromBytes([]byte(validYAML))
	if err != nil {
		t.Fatalf("LoadFromBytes() error = %v", err)
	}

	if cfg.Version != 1 {
		t.Errorf("Version = %d, want 1", cfg.Version)
	}

	flag, ok := cfg.Flags["new_checkout"]
	if !ok {
		t.Fatal("flag 'new_checkout' not found")
	}

	if !flag.Enabled {
		t.Error("flag should be enabled")
	}

	if flag.Type != "bool" {
		t.Errorf("Type = %s, want 'bool'", flag.Type)
	}

	if flag.Variants["true"] != 50 {
		t.Errorf("Variants['true'] = %d, want 50", flag.Variants["true"])
	}
}

func TestLoadFromBytes_InvalidYAML(t *testing.T) {
	invalidYAML := `invalid: yaml: content`

	_, err := LoadFromBytes([]byte(invalidYAML))
	if err == nil {
		t.Error("LoadFromBytes() expected error for invalid YAML")
	}
}

func TestLoadFromBytes_InvalidConfig(t *testing.T) {
	invalidConfig := `
version: 1
flags:
  test:
    type: "bool"
    variants:
      true: 60
      false: 30
`

	_, err := LoadFromBytes([]byte(invalidConfig))
	if err == nil {
		t.Error("LoadFromBytes() expected error for invalid config")
	}
}

