package config

import (
	"testing"
)

func TestCompile(t *testing.T) {
	cfg := &Config{
		Version: 1,
		Flags: map[string]Flag{
			"test_flag": {
				Enabled: true,
				Type:    "bool",
				Variants: map[string]int{
					"true":  50,
					"false": 50,
				},
				Rules: []Rule{
					{
						When: WhenCondition{
							All: []AttributeCondition{
								{Attr: "plan", Op: "eq", Value: "pro"},
							},
						},
						Then: ThenAction{
							Variants: map[string]int{
								"true":  90,
								"false": 10,
							},
						},
					},
				},
				Default: false,
			},
		},
	}

	compiled, err := Compile(cfg)
	if compiled == nil {
		t.Fatal("Compile() returned nil")
	}
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}

	flag, ok := compiled.Flags["test_flag"]
	if !ok {
		t.Fatal("compiled flag not found")
	}

	if !flag.Enabled {
		t.Error("flag should be enabled")
	}

	if len(flag.Rules) != 1 {
		t.Errorf("Rules length = %d, want 1", len(flag.Rules))
	}

	if flag.Rules[0].Conditions[0].Regex != nil {
		t.Error("non-regex condition should not have compiled regex")
	}
}

func TestCompile_WithRegex(t *testing.T) {
	cfg := &Config{
		Version: 1,
		Flags: map[string]Flag{
			"test": {
				Type: "bool",
				Rules: []Rule{
					{
						When: WhenCondition{
							All: []AttributeCondition{
								{Attr: "email", Op: "matches", Value: "^[a-z]+@example\\.com$"},
							},
						},
						Then: ThenAction{
							Variants: map[string]int{"true": 100},
						},
					},
				},
			},
		},
	}

	compiled, err := Compile(cfg)
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}

	rule := compiled.Flags["test"].Rules[0]
	if rule.Conditions[0].Regex == nil {
		t.Fatal("regex condition should have compiled regex")
	}

	// Test that regex works
	if !rule.Conditions[0].Regex.MatchString("test@example.com") {
		t.Error("regex should match valid email")
	}

	if rule.Conditions[0].Regex.MatchString("invalid") {
		t.Error("regex should not match invalid email")
	}
}

func TestCompile_NilConfig(t *testing.T) {
	_, err := Compile(nil)
	if err == nil {
		t.Error("Compile() expected error for nil config")
	}
}

