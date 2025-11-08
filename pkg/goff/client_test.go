package goff

import (
	"testing"
	"time"
)

func TestClient_Bool(t *testing.T) {
	client, err := New(
		WithFile("../../testdata/flags.yaml"),
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer client.Close()

	ctx := Context{
		Key: "user:123",
		Attrs: map[string]interface{}{
			"plan": "pro",
		},
	}

	result := client.Bool("new_checkout", ctx, false)
	if result != true && result != false {
		t.Errorf("Bool() = %v, want bool", result)
	}
}

func TestClient_String(t *testing.T) {
	client, err := New(
		WithFile("../../testdata/flags.yaml"),
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer client.Close()

	ctx := Context{
		Key: "user:123",
		Attrs: map[string]interface{}{
			"theme": "dark",
		},
	}

	result := client.String("checkout_theme", ctx, "default")
	if result == "" {
		t.Error("String() returned empty string")
	}
}

func TestClient_WithAutoReload(t *testing.T) {
	client, err := New(
		WithFile("../../testdata/flags.yaml"),
		WithAutoReload(1*time.Second),
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer client.Close()

	// Give it a moment to start watching
	time.Sleep(100 * time.Millisecond)

	ctx := Context{Key: "user:1", Attrs: map[string]interface{}{}}
	_ = client.Bool("new_checkout", ctx, false)
}

func TestClient_WithHooks(t *testing.T) {
	var called bool
	var calledFlag, calledVariant string

	client, err := New(
		WithFile("../../testdata/flags.yaml"),
		WithHooks(Hooks{
			AfterEval: func(flag, variant string, reason Reason) {
				called = true
				calledFlag = flag
				calledVariant = variant
				_ = reason // reason is available but not checked in this test
			},
		}),
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer client.Close()

	ctx := Context{Key: "user:1", Attrs: map[string]interface{}{}}
	_ = client.Bool("new_checkout", ctx, false)

	if !called {
		t.Error("AfterEval hook was not called")
	}
	if calledFlag != "new_checkout" {
		t.Errorf("hook flag = %v, want 'new_checkout'", calledFlag)
	}
	if calledVariant != "true" && calledVariant != "false" {
		t.Errorf("hook variant = %v, want 'true' or 'false'", calledVariant)
	}
}

