package eval

import (
	"github.com/0mjs/goff/internal/config"
)

// Context represents the evaluation context.
type Context struct {
	Key   string
	Attrs map[string]any
}

// EvalRule evaluates a compiled rule against a context.
// Returns true if the rule matches, false otherwise.
// Errors are logged but don't prevent evaluation (rule is skipped).
func EvalRule(rule *config.CompiledRule, ctx Context) bool {
	if len(rule.Conditions) == 0 {
		return false
	}

	// Check if all conditions match (for "all") or any condition matches (for "any")
	allMatch := true
	anyMatch := false

	for _, cond := range rule.Conditions {
		attrValue, exists := ctx.Attrs[cond.Attr]
		if !exists {
			// Attribute missing - condition fails
			if cond.IsAll {
				allMatch = false
				break
			}
			continue
		}

		match, err := EvalOperator(attrValue, cond.Op, cond.Value, cond.Regex)
		if err != nil {
			// Evaluation error - skip this condition
			// For "all", this means the rule doesn't match
			// For "any", we continue checking other conditions
			if cond.IsAll {
				allMatch = false
				break
			}
			continue
		}

		if match {
			anyMatch = true
			if !cond.IsAll {
				// For "any", first match wins
				break
			}
		} else {
			if cond.IsAll {
				allMatch = false
				break
			}
		}
	}

	// If IsAll is true, all conditions must match
	// If IsAll is false, any condition must match
	if rule.Conditions[0].IsAll {
		return allMatch
	}
	return anyMatch
}
