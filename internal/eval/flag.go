package eval

import (
	"sort"

	"github.com/0mjs/goff/internal/config"
)

// EvalBool evaluates a boolean flag.
func EvalBool(flag *config.CompiledFlag, flagKey string, ctx Context, def bool) (bool, Reason) {
	if flag == nil {
		return def, Missing
	}

	if !flag.Enabled {
		if flag.Default != nil {
			if d, ok := flag.Default.(bool); ok {
				return d, Disabled
			}
		}
		return def, Disabled
	}

	// Evaluate rules in order; first match wins
	for _, rule := range flag.Rules {
		if EvalRule(rule, ctx) {
			// Rule matched - use rule variants
			return selectVariantBool(flagKey, ctx.Key, rule.Variants, def)
		}
	}

	// No rule matched - fall back to percentage rollout
	if len(flag.Variants) > 0 {
		return selectVariantBool(flagKey, ctx.Key, flag.Variants, def)
	}

	// No variants defined - use default
	if flag.Default != nil {
		if d, ok := flag.Default.(bool); ok {
			return d, Default
		}
	}

	return def, Default
}

// EvalString evaluates a string flag.
func EvalString(flag *config.CompiledFlag, flagKey string, ctx Context, def string) (string, Reason) {
	if flag == nil {
		return def, Missing
	}

	if !flag.Enabled {
		if flag.Default != nil {
			if d, ok := flag.Default.(string); ok {
				return d, Disabled
			}
		}
		return def, Disabled
	}

	// Evaluate rules in order; first match wins
	for _, rule := range flag.Rules {
		if EvalRule(rule, ctx) {
			// Rule matched - use rule variants
			return selectVariantString(flagKey, ctx.Key, rule.Variants, def)
		}
	}

	// No rule matched - fall back to percentage rollout
	if len(flag.Variants) > 0 {
		return selectVariantString(flagKey, ctx.Key, flag.Variants, def)
	}

	// No variants defined - use default
	if flag.Default != nil {
		if d, ok := flag.Default.(string); ok {
			return d, Default
		}
	}

	return def, Default
}

// selectVariantBool selects a boolean variant based on percentage rollout.
func selectVariantBool(flagKey, contextKey string, variants map[string]int, def bool) (bool, Reason) {
	if len(variants) == 0 {
		return def, Default
	}

	bucket := HashFlagContext(flagKey, contextKey, 0)
	
	// Sort variants for deterministic iteration
	type variantPct struct {
		variant string
		pct     int
	}
	var sorted []variantPct
	for variant, pct := range variants {
		sorted = append(sorted, variantPct{variant, pct})
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].variant < sorted[j].variant
	})
	
	// Calculate cumulative percentages
	var cumulative int
	for _, v := range sorted {
		cumulative += v.pct
		if bucket < cumulative {
			if v.variant == "true" {
				return true, Match
			}
			return false, Match
		}
	}

	// Fallback (shouldn't happen if percentages sum to 100)
	return def, Percent
}

// selectVariantString selects a string variant based on percentage rollout.
func selectVariantString(flagKey, contextKey string, variants map[string]int, def string) (string, Reason) {
	if len(variants) == 0 {
		return def, Default
	}

	bucket := HashFlagContext(flagKey, contextKey, 0)
	
	// Sort variants for deterministic iteration
	type variantPct struct {
		variant string
		pct     int
	}
	var sorted []variantPct
	for variant, pct := range variants {
		sorted = append(sorted, variantPct{variant, pct})
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].variant < sorted[j].variant
	})
	
	// Calculate cumulative percentages
	var cumulative int
	for _, v := range sorted {
		cumulative += v.pct
		if bucket < cumulative {
			return v.variant, Match
		}
	}

	// Fallback (shouldn't happen if percentages sum to 100)
	return def, Percent
}

