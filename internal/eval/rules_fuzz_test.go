package eval

import (
	"testing"

	"github.com/0mjs/goff/internal/config"
)

func FuzzEvalRule(f *testing.F) {
	f.Add("plan", "eq", "pro")
	f.Fuzz(func(t *testing.T, attr, op, value string) {
		// Create a rule with fuzzed values
		rule := &config.CompiledRule{
			Conditions: []*config.CompiledCondition{
				{Attr: attr, Op: op, Value: value, IsAll: true},
			},
			Variants: map[string]int{"true": 100},
		}

		ctx := Context{
			Key: "user:1",
			Attrs: map[string]any{
				attr: value,
			},
		}

		// Should not panic
		_ = EvalRule(rule, ctx)
	})
}
