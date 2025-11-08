package eval

import (
	"github.com/cespare/xxhash/v2"
)

// HashFlagContext computes a deterministic hash for flag evaluation.
// Input: flagKey + '\x1f' + contextKey + salt
// Returns a value in [0..99] for percentage rollouts.
func HashFlagContext(flagKey, contextKey string, salt uint64) int {
	h := xxhash.New()
	h.WriteString(flagKey)
	h.Write([]byte{'\x1f'})
	h.WriteString(contextKey)
	
	// Add salt if provided
	if salt != 0 {
		var saltBytes [8]byte
		for i := 0; i < 8; i++ {
			saltBytes[i] = byte(salt >> (i * 8))
		}
		h.Write(saltBytes[:])
	}
	
	hash := h.Sum64()
	// Map to [0..99] for percentage rollouts
	return int(hash % 100)
}

