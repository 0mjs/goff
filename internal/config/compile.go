package config

import (
	"fmt"
	"regexp"
)

// Compiled represents a compiled, immutable configuration ready for evaluation.
type Compiled struct {
	Flags map[string]*CompiledFlag
}

// CompiledFlag represents a compiled flag ready for evaluation.
type CompiledFlag struct {
	Enabled  bool
	Type     string // "bool" | "string"
	Variants map[string]int // variant -> percentage (0-100)
	Rules    []*CompiledRule
	Default  any // bool for bool flags, string for string flags
}

// CompiledRule represents a compiled rule ready for evaluation.
type CompiledRule struct {
	Conditions []*CompiledCondition
	Variants   map[string]int // variant -> percentage (0-100)
}

// CompiledCondition represents a compiled condition.
type CompiledCondition struct {
	Attr    string
	Op      string
	Value   interface{}
	Regex   *regexp.Regexp // compiled regex for "matches" operator
	IsAll   bool           // true if part of "all", false if part of "any"
}

// Compile compiles a Config into a Compiled configuration.
func Compile(cfg *Config) (*Compiled, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is nil")
	}

	compiled := &Compiled{
		Flags: make(map[string]*CompiledFlag, len(cfg.Flags)),
	}

	for flagKey, flag := range cfg.Flags {
		compiledFlag, err := compileFlag(flagKey, &flag)
		if err != nil {
			return nil, fmt.Errorf("compile flag %q: %w", flagKey, err)
		}
		compiled.Flags[flagKey] = compiledFlag
	}

	return compiled, nil
}

func compileFlag(flagKey string, flag *Flag) (*CompiledFlag, error) {
	compiledFlag := &CompiledFlag{
		Enabled:  flag.Enabled,
		Type:     flag.Type,
		Variants: make(map[string]int, len(flag.Variants)),
		Rules:   make([]*CompiledRule, 0, len(flag.Rules)),
		Default:  flag.Default,
	}

	// Copy variants
	for k, v := range flag.Variants {
		compiledFlag.Variants[k] = v
	}

	// Compile rules
	for i, rule := range flag.Rules {
		compiledRule, err := compileRule(&rule)
		if err != nil {
			return nil, fmt.Errorf("compile rule %d: %w", i, err)
		}
		compiledFlag.Rules = append(compiledFlag.Rules, compiledRule)
	}

	return compiledFlag, nil
}

func compileRule(rule *Rule) (*CompiledRule, error) {
	compiledRule := &CompiledRule{
		Variants: make(map[string]int, len(rule.Then.Variants)),
	}

	// Copy variants
	for k, v := range rule.Then.Variants {
		compiledRule.Variants[k] = v
	}

	// Compile conditions
	var conditions []*CompiledCondition
	isAll := len(rule.When.All) > 0

	if isAll {
		conditions = make([]*CompiledCondition, 0, len(rule.When.All))
		for _, cond := range rule.When.All {
			compiledCond, err := compileCondition(&cond, true)
			if err != nil {
				return nil, err
			}
			conditions = append(conditions, compiledCond)
		}
	} else {
		conditions = make([]*CompiledCondition, 0, len(rule.When.Any))
		for _, cond := range rule.When.Any {
			compiledCond, err := compileCondition(&cond, false)
			if err != nil {
				return nil, err
			}
			conditions = append(conditions, compiledCond)
		}
	}

	compiledRule.Conditions = conditions
	return compiledRule, nil
}

func compileCondition(cond *AttributeCondition, isAll bool) (*CompiledCondition, error) {
	compiled := &CompiledCondition{
		Attr:  cond.Attr,
		Op:    cond.Op,
		Value: cond.Value,
		IsAll: isAll,
	}

	// Compile regex for "matches" operator
	if cond.Op == "matches" {
		pattern, ok := cond.Value.(string)
		if !ok {
			return nil, fmt.Errorf("matches operator requires string value")
		}
		regex, err := regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("compile regex: %w", err)
		}
		compiled.Regex = regex
	}

	return compiled, nil
}

