package eval

import (
	"testing"

	"github.com/0mjs/goff/internal/config"
)

func TestEvalBool_Enabled(t *testing.T) {
	flag := &config.CompiledFlag{
		Enabled: true,
		Type:    "bool",
		Variants: map[string]int{
			"true":  50,
			"false": 50,
		},
		Default: false,
	}

	ctx := Context{Key: "user:1", Attrs: map[string]any{}}

	result, reason := EvalBool(flag, "test", ctx, false)

	// Should be either Match or Percent depending on bucket
	if reason != Match && reason != Percent {
		t.Errorf("EvalBool() reason = %v, want Match or Percent", reason)
	}

	// Result should be deterministic for same context
	result2, _ := EvalBool(flag, "test", ctx, false)
	if result != result2 {
		t.Error("EvalBool() should be deterministic")
	}
}

func TestEvalBool_Disabled(t *testing.T) {
	flag := &config.CompiledFlag{
		Enabled: false,
		Type:    "bool",
		Default: false,
	}

	ctx := Context{Key: "user:1", Attrs: map[string]any{}}

	result, reason := EvalBool(flag, "test", ctx, true)

	if result != false {
		t.Errorf("EvalBool() = %v, want false (default)", result)
	}
	if reason != Disabled {
		t.Errorf("EvalBool() reason = %v, want Disabled", reason)
	}
}

func TestEvalBool_Missing(t *testing.T) {
	ctx := Context{Key: "user:1", Attrs: map[string]any{}}

	result, reason := EvalBool(nil, "test", ctx, true)

	if result != true {
		t.Errorf("EvalBool() = %v, want true (default)", result)
	}
	if reason != Missing {
		t.Errorf("EvalBool() reason = %v, want Missing", reason)
	}
}

func TestEvalBool_WithRule(t *testing.T) {
	flag := &config.CompiledFlag{
		Enabled: true,
		Type:    "bool",
		Variants: map[string]int{
			"true":  50,
			"false": 50,
		},
		Rules: []*config.CompiledRule{
			{
				Conditions: []*config.CompiledCondition{
					{Attr: "plan", Op: "eq", Value: "pro", IsAll: true},
				},
				Variants: map[string]int{
					"true":  100,
					"false": 0,
				},
			},
		},
		Default: false,
	}

	ctx := Context{
		Key: "user:1",
		Attrs: map[string]any{
			"plan": "pro",
		},
	}

	result, reason := EvalBool(flag, "test", ctx, false)

	if result != true {
		t.Errorf("EvalBool() = %v, want true (rule match)", result)
	}
	if reason != Match {
		t.Errorf("EvalBool() reason = %v, want Match", reason)
	}
}

func TestEvalString_Enabled(t *testing.T) {
	flag := &config.CompiledFlag{
		Enabled: true,
		Type:    "string",
		Variants: map[string]int{
			"red":   50,
			"blue":  30,
			"green": 20,
		},
		Default: "red",
	}

	ctx := Context{Key: "user:1", Attrs: map[string]any{}}

	result, reason := EvalString(flag, "test", ctx, "default")

	if result == "" {
		t.Error("EvalString() should return a variant")
	}
	if reason != Match && reason != Percent {
		t.Errorf("EvalString() reason = %v, want Match or Percent", reason)
	}
}

func TestEvalString_Disabled(t *testing.T) {
	flag := &config.CompiledFlag{
		Enabled: false,
		Type:    "string",
		Default: "red",
	}

	ctx := Context{Key: "user:1", Attrs: map[string]any{}}

	result, reason := EvalString(flag, "test", ctx, "default")

	if result != "red" {
		t.Errorf("EvalString() = %v, want 'red' (default)", result)
	}
	if reason != Disabled {
		t.Errorf("EvalString() reason = %v, want Disabled", reason)
	}
}

func TestEvalString_WithRule(t *testing.T) {
	flag := &config.CompiledFlag{
		Enabled: true,
		Type:    "string",
		Variants: map[string]int{
			"red":  50,
			"blue": 50,
		},
		Rules: []*config.CompiledRule{
			{
				Conditions: []*config.CompiledCondition{
					{Attr: "theme", Op: "eq", Value: "dark", IsAll: true},
				},
				Variants: map[string]int{
					"black": 100,
				},
			},
		},
		Default: "red",
	}

	ctx := Context{
		Key: "user:1",
		Attrs: map[string]any{
			"theme": "dark",
		},
	}

	result, reason := EvalString(flag, "test", ctx, "default")

	if result != "black" {
		t.Errorf("EvalString() = %v, want 'black' (rule match)", result)
	}
	if reason != Match {
		t.Errorf("EvalString() reason = %v, want Match", reason)
	}
}
