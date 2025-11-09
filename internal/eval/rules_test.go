package eval

import (
	"testing"

	"github.com/0mjs/goff/internal/config"
)

func TestEvalRule_AllConditions(t *testing.T) {
	rule := &config.CompiledRule{
		Conditions: []*config.CompiledCondition{
			{Attr: "plan", Op: "eq", Value: "pro", IsAll: true},
			{Attr: "region", Op: "eq", Value: "us", IsAll: true},
		},
		Variants: map[string]int{"true": 100},
	}

	ctx := Context{
		Key: "user:1",
		Attrs: map[string]any{
			"plan":   "pro",
			"region": "us",
		},
	}

	if !EvalRule(rule, ctx) {
		t.Error("EvalRule() should return true when all conditions match")
	}

	ctx.Attrs["region"] = "eu"
	if EvalRule(rule, ctx) {
		t.Error("EvalRule() should return false when one condition doesn't match")
	}
}

func TestEvalRule_AnyConditions(t *testing.T) {
	rule := &config.CompiledRule{
		Conditions: []*config.CompiledCondition{
			{Attr: "plan", Op: "eq", Value: "pro", IsAll: false},
			{Attr: "plan", Op: "eq", Value: "enterprise", IsAll: false},
		},
		Variants: map[string]int{"true": 100},
	}

	ctx := Context{
		Key: "user:1",
		Attrs: map[string]any{
			"plan": "pro",
		},
	}

	if !EvalRule(rule, ctx) {
		t.Error("EvalRule() should return true when any condition matches")
	}

	ctx.Attrs["plan"] = "basic"
	if EvalRule(rule, ctx) {
		t.Error("EvalRule() should return false when no conditions match")
	}
}

func TestEvalRule_MissingAttribute(t *testing.T) {
	rule := &config.CompiledRule{
		Conditions: []*config.CompiledCondition{
			{Attr: "plan", Op: "eq", Value: "pro", IsAll: true},
		},
		Variants: map[string]int{"true": 100},
	}

	ctx := Context{
		Key:   "user:1",
		Attrs: map[string]any{},
	}

	if EvalRule(rule, ctx) {
		t.Error("EvalRule() should return false when required attribute is missing")
	}
}

func TestEvalRule_ErrorHandling(t *testing.T) {
	rule := &config.CompiledRule{
		Conditions: []*config.CompiledCondition{
			{Attr: "value", Op: "gt", Value: 10, IsAll: true},
		},
		Variants: map[string]int{"true": 100},
	}

	ctx := Context{
		Key: "user:1",
		Attrs: map[string]any{
			"value": "not a number",
		},
	}

	// Should not panic, should return false
	if EvalRule(rule, ctx) {
		t.Error("EvalRule() should return false when operator evaluation fails")
	}
}
