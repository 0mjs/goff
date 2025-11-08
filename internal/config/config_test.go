package config

import (
	"testing"
)

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				Version: 1,
				Flags: map[string]Flag{
					"test_flag": {
						Enabled: true,
						Type:    "bool",
						Variants: map[string]int{
							"true":  50,
							"false": 50,
						},
						Default: false,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid version",
			config: &Config{
				Version: 2,
				Flags:   map[string]Flag{},
			},
			wantErr: true,
		},
		{
			name: "no flags",
			config: &Config{
				Version: 1,
				Flags:   map[string]Flag{},
			},
			wantErr: true,
		},
		{
			name: "invalid flag type",
			config: &Config{
				Version: 1,
				Flags: map[string]Flag{
					"test": {
						Type: "invalid",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "bool flag invalid variants sum",
			config: &Config{
				Version: 1,
				Flags: map[string]Flag{
					"test": {
						Type: "bool",
						Variants: map[string]int{
							"true":  60,
							"false": 30,
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "bool flag missing true variant",
			config: &Config{
				Version: 1,
				Flags: map[string]Flag{
					"test": {
						Type: "bool",
						Variants: map[string]int{
							"false": 100,
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "string flag invalid variants sum",
			config: &Config{
				Version: 1,
				Flags: map[string]Flag{
					"test": {
						Type: "string",
						Variants: map[string]int{
							"a": 60,
							"b": 30,
						},
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFlagValidate(t *testing.T) {
	tests := []struct {
		name    string
		flag    Flag
		wantErr bool
	}{
		{
			name: "valid bool flag",
			flag: Flag{
				Type: "bool",
				Variants: map[string]int{
					"true":  50,
					"false": 50,
				},
				Default: false,
			},
			wantErr: false,
		},
		{
			name: "valid string flag",
			flag: Flag{
				Type: "string",
				Variants: map[string]int{
					"red":   50,
					"blue":  30,
					"green": 20,
				},
				Default: "red",
			},
			wantErr: false,
		},
		{
			name: "invalid default type for bool",
			flag: Flag{
				Type:    "bool",
				Default: "not a bool",
			},
			wantErr: true,
		},
		{
			name: "invalid default type for string",
			flag: Flag{
				Type:    "string",
				Default: 123,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.flag.Validate("test_flag")
			if (err != nil) != tt.wantErr {
				t.Errorf("Flag.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRuleValidate(t *testing.T) {
	tests := []struct {
		name    string
		rule    Rule
		wantErr bool
	}{
		{
			name: "valid rule with all",
			rule: Rule{
				When: WhenCondition{
					All: []AttributeCondition{
						{Attr: "plan", Op: "eq", Value: "pro"},
					},
				},
				Then: ThenAction{
					Variants: map[string]int{
						"true":  90,
						"false": 10,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid rule with any",
			rule: Rule{
				When: WhenCondition{
					Any: []AttributeCondition{
						{Attr: "plan", Op: "eq", Value: "pro"},
					},
				},
				Then: ThenAction{
					Variants: map[string]int{
						"true": 100,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "rule with both all and any",
			rule: Rule{
				When: WhenCondition{
					All: []AttributeCondition{{Attr: "a", Op: "eq", Value: "b"}},
					Any: []AttributeCondition{{Attr: "c", Op: "eq", Value: "d"}},
				},
				Then: ThenAction{
					Variants: map[string]int{"true": 100},
				},
			},
			wantErr: true,
		},
		{
			name: "rule with no conditions",
			rule: Rule{
				When: WhenCondition{},
				Then: ThenAction{
					Variants: map[string]int{"true": 100},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid operator",
			rule: Rule{
				When: WhenCondition{
					All: []AttributeCondition{
						{Attr: "plan", Op: "invalid", Value: "pro"},
					},
				},
				Then: ThenAction{
					Variants: map[string]int{"true": 100},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid regex in matches",
			rule: Rule{
				When: WhenCondition{
					All: []AttributeCondition{
						{Attr: "plan", Op: "matches", Value: "[invalid"},
					},
				},
				Then: ThenAction{
					Variants: map[string]int{"true": 100},
				},
			},
			wantErr: true,
		},
		{
			name: "rule with no variants",
			rule: Rule{
				When: WhenCondition{
					All: []AttributeCondition{
						{Attr: "plan", Op: "eq", Value: "pro"},
					},
				},
				Then: ThenAction{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.rule.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Rule.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

