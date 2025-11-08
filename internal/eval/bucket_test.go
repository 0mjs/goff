package eval

import (
	"testing"
)

func TestHashFlagContext_Deterministic(t *testing.T) {
	flagKey := "test_flag"
	contextKey := "user:123"

	hash1 := HashFlagContext(flagKey, contextKey, 0)
	hash2 := HashFlagContext(flagKey, contextKey, 0)

	if hash1 != hash2 {
		t.Errorf("HashFlagContext() not deterministic: %d != %d", hash1, hash2)
	}

	// Should be in [0..99] range
	if hash1 < 0 || hash1 >= 100 {
		t.Errorf("HashFlagContext() out of range: %d", hash1)
	}
}

func TestHashFlagContext_DifferentInputs(t *testing.T) {
	hash1 := HashFlagContext("flag1", "user:1", 0)
	hash2 := HashFlagContext("flag2", "user:1", 0)
	hash3 := HashFlagContext("flag1", "user:2", 0)

	if hash1 == hash2 {
		t.Error("different flags should produce different hashes")
	}

	if hash1 == hash3 {
		t.Error("different context keys should produce different hashes")
	}
}

func TestHashFlagContext_WithSalt(t *testing.T) {
	hash1 := HashFlagContext("flag", "user:1", 0)
	hash2 := HashFlagContext("flag", "user:1", 1)

	if hash1 == hash2 {
		t.Error("different salts should produce different hashes")
	}
}

func TestHashFlagContext_Distribution(t *testing.T) {
	// Test that hashes are reasonably distributed
	buckets := make(map[int]int)
	for i := 0; i < 1000; i++ {
		hash := HashFlagContext("flag", "user:"+string(rune(i)), 0)
		buckets[hash]++
	}

	// Should have at least 50 unique buckets out of 100
	if len(buckets) < 50 {
		t.Errorf("poor distribution: only %d unique buckets out of 1000 samples", len(buckets))
	}
}

