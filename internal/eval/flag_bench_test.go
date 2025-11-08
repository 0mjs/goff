package eval

import (
	"testing"

	"github.com/0mjs/goff/internal/config"
)

func BenchmarkEvalBool(b *testing.B) {
	flag := &config.CompiledFlag{
		Enabled: true,
		Type:    "bool",
		Variants: map[string]int{
			"true":  50,
			"false": 50,
		},
		Default: false,
	}

	ctx := Context{
		Key: "user:123",
		Attrs: map[string]interface{}{
			"plan": "pro",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = EvalBool(flag, "test_flag", ctx, false)
	}
}

func BenchmarkEvalBool_WithRule(b *testing.B) {
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
					"true":  90,
					"false": 10,
				},
			},
		},
		Default: false,
	}

	ctx := Context{
		Key: "user:123",
		Attrs: map[string]interface{}{
			"plan": "pro",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = EvalBool(flag, "test_flag", ctx, false)
	}
}

func BenchmarkEvalString(b *testing.B) {
	flag := &config.CompiledFlag{
		Enabled: true,
		Type:    "string",
		Variants: map[string]int{
			"red":   40,
			"blue":  30,
			"green": 30,
		},
		Default: "red",
	}

	ctx := Context{
		Key: "user:123",
		Attrs: map[string]interface{}{},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = EvalString(flag, "test_flag", ctx, "default")
	}
}

func BenchmarkEvalBool_Concurrent(b *testing.B) {
	flag := &config.CompiledFlag{
		Enabled: true,
		Type:    "bool",
		Variants: map[string]int{
			"true":  50,
			"false": 50,
		},
		Default: false,
	}

	b.RunParallel(func(pb *testing.PB) {
		ctx := Context{
			Key: "user:123",
			Attrs: map[string]interface{}{},
		}
		for pb.Next() {
			_, _ = EvalBool(flag, "test_flag", ctx, false)
		}
	})
}

func BenchmarkEvalString_Concurrent(b *testing.B) {
	flag := &config.CompiledFlag{
		Enabled: true,
		Type:    "string",
		Variants: map[string]int{
			"red":   50,
			"blue":  50,
		},
		Default: "red",
	}

	b.RunParallel(func(pb *testing.PB) {
		ctx := Context{
			Key: "user:123",
			Attrs: map[string]interface{}{},
		}
		for pb.Next() {
			_, _ = EvalString(flag, "test_flag", ctx, "default")
		}
	})
}

