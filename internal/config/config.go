package config

import (
	"fmt"
	"regexp"
)

// Config represents the root configuration structure.
type Config struct {
	Version int             `yaml:"version"`
	Flags   map[string]Flag `yaml:"flags"`
}

// Flag represents a single feature flag.
type Flag struct {
	Enabled  bool           `yaml:"enabled"`
	Type     string         `yaml:"type"`     // "bool" | "string"
	Variants map[string]int `yaml:"variants"` // For bool: "true"/"false" with 0-100 percentages; for string: variant names with percentages
	Rules    []Rule         `yaml:"rules,omitempty"`
	Default  any            `yaml:"default"` // bool for bool flags, string for string flags
}

// Rule represents a targeting rule for a flag.
type Rule struct {
	When WhenCondition `yaml:"when"`
	Then ThenAction    `yaml:"then"`
}

// WhenCondition represents the condition to match.
type WhenCondition struct {
	All []AttributeCondition `yaml:"all,omitempty"`
	Any []AttributeCondition `yaml:"any,omitempty"`
}

// AttributeCondition represents a single attribute condition.
type AttributeCondition struct {
	Attr  string `yaml:"attr"`
	Op    string `yaml:"op"` // eq | neq | gt | gte | lt | lte | in | contains | matches
	Value any    `yaml:"value"`
}

// ThenAction represents the action to take when a rule matches.
type ThenAction struct {
	Variants map[string]int `yaml:"variants"` // Same format as Flag.Variants
}

// Validate checks the configuration for errors.
func (c *Config) Validate() error {
	if c.Version != 1 {
		return fmt.Errorf("unsupported config version: %d (expected 1)", c.Version)
	}

	if len(c.Flags) == 0 {
		return fmt.Errorf("no flags defined")
	}

	for flagKey, flag := range c.Flags {
		if err := flag.Validate(flagKey); err != nil {
			return fmt.Errorf("flag %q: %w", flagKey, err)
		}
	}

	return nil
}

// Validate checks a flag for errors.
func (f *Flag) Validate(flagKey string) error {
	if f.Type != "bool" && f.Type != "string" {
		return fmt.Errorf("invalid type %q (must be 'bool' or 'string')", f.Type)
	}

	if f.Type == "bool" {
		if f.Default != nil {
			if _, ok := f.Default.(bool); !ok {
				return fmt.Errorf("default must be bool for bool flags, got %T", f.Default)
			}
		}
		// Validate bool variants
		if len(f.Variants) > 0 {
			total := 0
			hasTrue := false
			hasFalse := false
			for k, v := range f.Variants {
				if k != "true" && k != "false" {
					return fmt.Errorf("bool flag variants must be 'true' or 'false', got %q", k)
				}
				if k == "true" {
					hasTrue = true
				}
				if k == "false" {
					hasFalse = true
				}
				if v < 0 || v > 100 {
					return fmt.Errorf("variant %q percentage must be 0-100, got %d", k, v)
				}
				total += v
			}
			if !hasTrue || !hasFalse {
				return fmt.Errorf("bool flag must have both 'true' and 'false' variants")
			}
			if total != 100 {
				return fmt.Errorf("bool flag variant percentages must sum to 100, got %d", total)
			}
		}
	} else {
		// string type
		if f.Default != nil {
			if _, ok := f.Default.(string); !ok {
				return fmt.Errorf("default must be string for string flags, got %T", f.Default)
			}
		}
		if len(f.Variants) > 0 {
			total := 0
			for k, v := range f.Variants {
				if v < 0 || v > 100 {
					return fmt.Errorf("variant %q percentage must be 0-100, got %d", k, v)
				}
				total += v
			}
			if total != 100 {
				return fmt.Errorf("string flag variant percentages must sum to 100, got %d", total)
			}
		}
	}

	// Validate rules
	for i, rule := range f.Rules {
		if err := rule.Validate(); err != nil {
			return fmt.Errorf("rule %d: %w", i, err)
		}
		// Validate rule variants
		if len(rule.Then.Variants) > 0 {
			if f.Type == "bool" {
				total := 0
				for k, v := range rule.Then.Variants {
					if k != "true" && k != "false" {
						return fmt.Errorf("rule %d: bool flag variants must be 'true' or 'false', got %q", i, k)
					}
					if v < 0 || v > 100 {
						return fmt.Errorf("rule %d: variant %q percentage must be 0-100, got %d", i, k, v)
					}
					total += v
				}
				if total != 100 {
					return fmt.Errorf("rule %d: bool flag variant percentages must sum to 100, got %d", i, total)
				}
			} else {
				total := 0
				for k, v := range rule.Then.Variants {
					if v < 0 || v > 100 {
						return fmt.Errorf("rule %d: variant %q percentage must be 0-100, got %d", i, k, v)
					}
					total += v
				}
				if total != 100 {
					return fmt.Errorf("rule %d: string flag variant percentages must sum to 100, got %d", i, total)
				}
			}
		}
	}

	return nil
}

// Validate checks a rule for errors.
func (r *Rule) Validate() error {
	hasAll := len(r.When.All) > 0
	hasAny := len(r.When.Any) > 0

	if !hasAll && !hasAny {
		return fmt.Errorf("rule must have 'all' or 'any' condition")
	}

	if hasAll && hasAny {
		return fmt.Errorf("rule cannot have both 'all' and 'any' conditions")
	}

	conditions := r.When.All
	if hasAny {
		conditions = r.When.Any
	}

	validOps := map[string]bool{
		"eq":       true,
		"neq":      true,
		"gt":       true,
		"gte":      true,
		"lt":       true,
		"lte":      true,
		"in":       true,
		"contains": true,
		"matches":  true,
	}

	for i, cond := range conditions {
		if cond.Attr == "" {
			return fmt.Errorf("condition %d: attr is required", i)
		}
		if !validOps[cond.Op] {
			return fmt.Errorf("condition %d: invalid operator %q", i, cond.Op)
		}
		if cond.Op == "matches" {
			// Validate regex at parse time
			pattern, ok := cond.Value.(string)
			if !ok {
				return fmt.Errorf("condition %d: 'matches' operator requires string value", i)
			}
			if _, err := regexp.Compile(pattern); err != nil {
				return fmt.Errorf("condition %d: invalid regex pattern %q: %w", i, pattern, err)
			}
		}
	}

	if len(r.Then.Variants) == 0 {
		return fmt.Errorf("rule must have 'then.variants'")
	}

	return nil
}
