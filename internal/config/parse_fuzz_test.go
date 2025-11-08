package config

import (
	"testing"
)

func FuzzLoadFromBytes(f *testing.F) {
	// Seed with valid YAML
	f.Add([]byte(`version: 1
flags:
  test:
    type: bool
    variants:
      true: 50
      false: 50
    default: false`))

	f.Fuzz(func(t *testing.T, data []byte) {
		// Should not panic
		cfg, err := LoadFromBytes(data)
		if err != nil {
			// Invalid input is expected
			return
		}

		// If we got a config, it should be valid
		if err := cfg.Validate(); err != nil {
			t.Errorf("Validate() error = %v", err)
		}

		// Should be able to compile
		_, err = Compile(cfg)
		if err != nil {
			t.Errorf("Compile() error = %v", err)
		}
	})
}

