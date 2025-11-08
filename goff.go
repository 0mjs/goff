package goff

import (
	"time"

	pkggoff "github.com/0mjs/goff/pkg/goff"
)

// Re-export types
type (
	Client  = pkggoff.Client
	Context = pkggoff.Context
	Hooks   = pkggoff.Hooks
	Reason  = pkggoff.Reason
	Option  = pkggoff.Option
)

// Re-export constants
const (
	Match    = pkggoff.Match
	Percent  = pkggoff.Percent
	Default  = pkggoff.Default
	Disabled = pkggoff.Disabled
	Missing  = pkggoff.Missing
	Error    = pkggoff.Error
)

// New creates a new Client with the given options.
func New(opts ...Option) (Client, error) {
	return pkggoff.New(opts...)
}

// WithFile loads a configuration from a file.
func WithFile(path string) Option {
	return pkggoff.WithFile(path)
}

// WithAutoReload enables automatic reloading of the configuration file.
func WithAutoReload(interval time.Duration) Option {
	return pkggoff.WithAutoReload(interval)
}

// WithHooks sets observability hooks.
func WithHooks(hooks Hooks) Option {
	return pkggoff.WithHooks(hooks)
}
