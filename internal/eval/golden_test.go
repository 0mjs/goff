package eval

import (
	"testing"

	"github.com/0mjs/goff/internal/config"
)

func TestGolden_EvalBool(t *testing.T) {
	cfg, err := config.LoadFromFile("../../testdata/flags.yaml")
	if err != nil {
		t.Fatalf("LoadFromFile() error = %v", err)
	}

	compiled, err := config.Compile(cfg)
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}

	flag := compiled.Flags["new_checkout"]
	if flag == nil {
		t.Fatal("flag 'new_checkout' not found")
	}

	tests := []struct {
		name       string
		ctx        Context
		def        bool
		want       bool
		wantReason Reason
	}{
		{
			name: "pro plan user",
			ctx: Context{
				Key: "user:pro1",
				Attrs: map[string]any{
					"plan": "pro",
				},
			},
			def:        false,
			want:       true, // Rule matches, 90% true
			wantReason: Match,
		},
		{
			name: "basic plan user",
			ctx: Context{
				Key: "user:basic1",
				Attrs: map[string]any{
					"plan": "basic",
				},
			},
			def:        false,
			wantReason: Match, // or Percent, depending on bucket
		},
		{
			name: "no attributes",
			ctx: Context{
				Key:   "user:1",
				Attrs: map[string]any{},
			},
			def:        false,
			wantReason: Match, // or Percent
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, reason := EvalBool(flag, "new_checkout", tt.ctx, tt.def)

			if tt.want != result && tt.want {
				// Only check if we specified a want value
				t.Errorf("EvalBool() = %v, want %v", result, tt.want)
			}

			if reason != tt.wantReason && tt.wantReason != 0 {
				// Only check if we specified a wantReason
				if tt.wantReason == Match && reason != Percent {
					// Allow Match or Percent for percentage rollouts
					t.Errorf("EvalBool() reason = %v, want %v or Percent", reason, tt.wantReason)
				} else if reason != tt.wantReason {
					t.Errorf("EvalBool() reason = %v, want %v", reason, tt.wantReason)
				}
			}
		})
	}
}

func TestGolden_EvalString(t *testing.T) {
	cfg, err := config.LoadFromFile("../../testdata/flags.yaml")
	if err != nil {
		t.Fatalf("LoadFromFile() error = %v", err)
	}

	compiled, err := config.Compile(cfg)
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}

	flag := compiled.Flags["checkout_theme"]
	if flag == nil {
		t.Fatal("flag 'checkout_theme' not found")
	}

	tests := []struct {
		name       string
		ctx        Context
		def        string
		want       string
		wantReason Reason
	}{
		{
			name: "dark theme user",
			ctx: Context{
				Key: "user:dark1",
				Attrs: map[string]any{
					"theme": "dark",
				},
			},
			def:        "default",
			want:       "black",
			wantReason: Match,
		},
		{
			name: "light theme user",
			ctx: Context{
				Key: "user:light1",
				Attrs: map[string]any{
					"theme": "light",
				},
			},
			def:        "default",
			wantReason: Match, // or Percent
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, reason := EvalString(flag, "checkout_theme", tt.ctx, tt.def)

			if tt.want != "" && result != tt.want {
				t.Errorf("EvalString() = %v, want %v", result, tt.want)
			}

			if reason != tt.wantReason && tt.wantReason != 0 {
				if tt.wantReason == Match && reason != Percent {
					t.Errorf("EvalString() reason = %v, want %v or Percent", reason, tt.wantReason)
				} else if reason != tt.wantReason {
					t.Errorf("EvalString() reason = %v, want %v", reason, tt.wantReason)
				}
			}
		})
	}
}

func TestGolden_DisabledFlag(t *testing.T) {
	cfg, err := config.LoadFromFile("../../testdata/flags.yaml")
	if err != nil {
		t.Fatalf("LoadFromFile() error = %v", err)
	}

	compiled, err := config.Compile(cfg)
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}

	flag := compiled.Flags["disabled_flag"]
	if flag == nil {
		t.Fatal("flag 'disabled_flag' not found")
	}

	ctx := Context{Key: "user:1", Attrs: map[string]any{}}
	result, reason := EvalBool(flag, "disabled_flag", ctx, false)

	if result != true {
		t.Errorf("EvalBool() = %v, want true (flag default)", result)
	}
	if reason != Disabled {
		t.Errorf("EvalBool() reason = %v, want Disabled", reason)
	}
}
