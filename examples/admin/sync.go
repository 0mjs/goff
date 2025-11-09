package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/0mjs/goff/internal/config"
	"gopkg.in/yaml.v3"
)

type SyncService struct {
	store *FlagStore
}

func NewSyncService(store *FlagStore) *SyncService {
	return &SyncService{store: store}
}

func (s *SyncService) SyncToYAML(yamlPath string) error {
	flags, err := s.store.GetAllFlags()
	if err != nil {
		return fmt.Errorf("get flags: %w", err)
	}

	cfg := config.Config{
		Version: 1,
		Flags:   make(map[string]config.Flag),
	}

	for _, dbFlag := range flags {
		if !dbFlag.Enabled {
			// For disabled flags, create minimal config
			var defaultValue interface{}
			if dbFlag.Type == "bool" {
				var b bool
				if err := json.Unmarshal([]byte(dbFlag.Default), &b); err != nil {
					return fmt.Errorf("parse default for %s: %w", dbFlag.Key, err)
				}
				defaultValue = b
			} else {
				var str string
				if err := json.Unmarshal([]byte(dbFlag.Default), &str); err != nil {
					return fmt.Errorf("parse default for %s: %w", dbFlag.Key, err)
				}
				defaultValue = str
			}

			cfg.Flags[dbFlag.Key] = config.Flag{
				Enabled: false,
				Type:    dbFlag.Type,
				Default: defaultValue,
			}
			continue
		}

		// Parse variants
		var variants map[string]int
		if err := json.Unmarshal([]byte(dbFlag.Variants), &variants); err != nil {
			return fmt.Errorf("parse variants for %s: %w", dbFlag.Key, err)
		}

		// Parse rules
		var rules []config.Rule
		if err := json.Unmarshal([]byte(dbFlag.Rules), &rules); err != nil {
			return fmt.Errorf("parse rules for %s: %w", dbFlag.Key, err)
		}

		// Parse default
		var defaultValue interface{}
		if dbFlag.Type == "bool" {
			var b bool
			if err := json.Unmarshal([]byte(dbFlag.Default), &b); err != nil {
				return fmt.Errorf("parse default for %s: %w", dbFlag.Key, err)
			}
			defaultValue = b
		} else {
			var str string
			if err := json.Unmarshal([]byte(dbFlag.Default), &str); err != nil {
				return fmt.Errorf("parse default for %s: %w", dbFlag.Key, err)
			}
			defaultValue = str
		}

		cfg.Flags[dbFlag.Key] = config.Flag{
			Enabled:  true,
			Type:     dbFlag.Type,
			Variants: variants,
			Rules:    rules,
			Default:  defaultValue,
		}
	}

	// Validate before writing
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("validate config: %w", err)
	}

	// Write to YAML
	data, err := yaml.Marshal(&cfg)
	if err != nil {
		return fmt.Errorf("marshal YAML: %w", err)
	}

	if err := os.WriteFile(yamlPath, data, 0644); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	return nil
}
