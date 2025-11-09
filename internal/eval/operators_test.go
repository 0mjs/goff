package eval

import (
	"regexp"
	"testing"
)

func TestEvalOperator_eq(t *testing.T) {
	tests := []struct {
		name    string
		attr    any
		opValue any
		want    bool
		wantErr bool
	}{
		{"string match", "pro", "pro", true, false},
		{"string mismatch", "basic", "pro", false, false},
		{"int match", 42, 42, true, false},
		{"int mismatch", 42, 43, false, false},
		{"float match", 3.14, 3.14, true, false},
		{"string to int", "42", 42, true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := EvalOperator(tt.attr, "eq", tt.opValue, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("EvalOperator() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("EvalOperator() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEvalOperator_neq(t *testing.T) {
	got, err := EvalOperator("pro", "neq", "basic", nil)
	if err != nil {
		t.Fatalf("EvalOperator() error = %v", err)
	}
	if !got {
		t.Error("EvalOperator() neq should return true for different values")
	}

	got, err = EvalOperator("pro", "neq", "pro", nil)
	if err != nil {
		t.Fatalf("EvalOperator() error = %v", err)
	}
	if got {
		t.Error("EvalOperator() neq should return false for same values")
	}
}

func TestEvalOperator_gt(t *testing.T) {
	tests := []struct {
		name    string
		attr    any
		opValue any
		want    bool
	}{
		{"5 > 3", 5, 3, true},
		{"3 > 5", 3, 5, false},
		{"5 > 5", 5, 5, false},
		{"3.5 > 3.0", 3.5, 3.0, true},
		{"string numbers", "10", "5", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := EvalOperator(tt.attr, "gt", tt.opValue, nil)
			if err != nil {
				t.Fatalf("EvalOperator() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("EvalOperator() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEvalOperator_gte(t *testing.T) {
	got, err := EvalOperator(5, "gte", 5, nil)
	if err != nil {
		t.Fatalf("EvalOperator() error = %v", err)
	}
	if !got {
		t.Error("EvalOperator() gte should return true for equal values")
	}

	got, err = EvalOperator(5, "gte", 3, nil)
	if err != nil {
		t.Fatalf("EvalOperator() error = %v", err)
	}
	if !got {
		t.Error("EvalOperator() gte should return true for greater values")
	}

	got, err = EvalOperator(3, "gte", 5, nil)
	if err != nil {
		t.Fatalf("EvalOperator() error = %v", err)
	}
	if got {
		t.Error("EvalOperator() gte should return false for lesser values")
	}
}

func TestEvalOperator_lt(t *testing.T) {
	got, err := EvalOperator(3, "lt", 5, nil)
	if err != nil {
		t.Fatalf("EvalOperator() error = %v", err)
	}
	if !got {
		t.Error("EvalOperator() lt should return true for lesser values")
	}

	got, err = EvalOperator(5, "lt", 3, nil)
	if err != nil {
		t.Fatalf("EvalOperator() error = %v", err)
	}
	if got {
		t.Error("EvalOperator() lt should return false for greater values")
	}

	got, err = EvalOperator(5, "lt", 5, nil)
	if err != nil {
		t.Fatalf("EvalOperator() error = %v", err)
	}
	if got {
		t.Error("EvalOperator() lt should return false for equal values")
	}
}

func TestEvalOperator_lte(t *testing.T) {
	got, err := EvalOperator(5, "lte", 5, nil)
	if err != nil {
		t.Fatalf("EvalOperator() error = %v", err)
	}
	if !got {
		t.Error("EvalOperator() lte should return true for equal values")
	}

	got, err = EvalOperator(3, "lte", 5, nil)
	if err != nil {
		t.Fatalf("EvalOperator() error = %v", err)
	}
	if !got {
		t.Error("EvalOperator() lte should return true for lesser values")
	}
}

func TestEvalOperator_in(t *testing.T) {
	got, err := EvalOperator("pro", "in", []any{"basic", "pro", "enterprise"}, nil)
	if err != nil {
		t.Fatalf("EvalOperator() error = %v", err)
	}
	if !got {
		t.Error("EvalOperator() in should return true when value is in array")
	}

	got, err = EvalOperator("free", "in", []any{"basic", "pro", "enterprise"}, nil)
	if err != nil {
		t.Fatalf("EvalOperator() error = %v", err)
	}
	if got {
		t.Error("EvalOperator() in should return false when value is not in array")
	}
}

func TestEvalOperator_contains(t *testing.T) {
	got, err := EvalOperator("hello world", "contains", "world", nil)
	if err != nil {
		t.Fatalf("EvalOperator() error = %v", err)
	}
	if !got {
		t.Error("EvalOperator() contains should return true when string contains substring")
	}

	got, err = EvalOperator("hello", "contains", "world", nil)
	if err != nil {
		t.Fatalf("EvalOperator() error = %v", err)
	}
	if got {
		t.Error("EvalOperator() contains should return false when string does not contain substring")
	}
}

func TestEvalOperator_matches(t *testing.T) {
	regex := regexp.MustCompile("^[a-z]+@example\\.com$")

	got, err := EvalOperator("test@example.com", "matches", nil, regex)
	if err != nil {
		t.Fatalf("EvalOperator() error = %v", err)
	}
	if !got {
		t.Error("EvalOperator() matches should return true for matching string")
	}

	got, err = EvalOperator("invalid", "matches", nil, regex)
	if err != nil {
		t.Fatalf("EvalOperator() error = %v", err)
	}
	if got {
		t.Error("EvalOperator() matches should return false for non-matching string")
	}
}

func TestEvalOperator_UnknownOp(t *testing.T) {
	_, err := EvalOperator("value", "unknown", "value", nil)
	if err == nil {
		t.Error("EvalOperator() expected error for unknown operator")
	}
}
