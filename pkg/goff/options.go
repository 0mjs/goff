package goff

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/0mjs/goff/internal/config"
	"github.com/fsnotify/fsnotify"
)

type optionConfig struct {
	filePath      string
	autoReload    time.Duration
	hooks         *Hooks
	compiled      *atomic.Pointer[*config.Compiled]
	watcher       *fsnotify.Watcher
	stopWatcher   chan struct{}
	watcherDone   chan struct{}
}

// Option configures a Client.
type Option func(*optionConfig) error

// WithFile loads a configuration from a file.
func WithFile(path string) Option {
	return func(cfg *optionConfig) error {
		cfg.filePath = path
		return nil
	}
}

// WithAutoReload enables automatic reloading of the configuration file.
func WithAutoReload(interval time.Duration) Option {
	return func(cfg *optionConfig) error {
		cfg.autoReload = interval
		return nil
	}
}

// WithHooks sets observability hooks.
func WithHooks(hooks Hooks) Option {
	return func(cfg *optionConfig) error {
		cfg.hooks = &hooks
		return nil
	}
}

// New creates a new Client with the given options.
func New(opts ...Option) (Client, error) {
	cfg := &optionConfig{
		compiled: &atomic.Pointer[*config.Compiled]{},
	}

	// Apply options
	for _, opt := range opts {
		if err := opt(cfg); err != nil {
			return nil, fmt.Errorf("apply option: %w", err)
		}
	}

	// Load initial config
	if cfg.filePath == "" {
		return nil, fmt.Errorf("file path required (use WithFile)")
	}

	initialConfig, err := loadConfig(cfg.filePath)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	cfg.compiled.Store(&initialConfig)

	// Set up auto-reload if requested
	var closer func() error
	if cfg.autoReload > 0 {
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			return nil, fmt.Errorf("create watcher: %w", err)
		}

		if err := watcher.Add(cfg.filePath); err != nil {
			watcher.Close()
			return nil, fmt.Errorf("watch file: %w", err)
		}

		cfg.watcher = watcher
		cfg.stopWatcher = make(chan struct{})
		cfg.watcherDone = make(chan struct{})

		go watchFile(cfg)

		closer = func() error {
			close(cfg.stopWatcher)
			watcher.Close()
			<-cfg.watcherDone
			return nil
		}
	}

	return &client{
		config: cfg.compiled,
		hooks:  cfg.hooks,
		closer: closer,
	}, nil
}

func loadConfig(path string) (*config.Compiled, error) {
	cfg, err := config.LoadFromFile(path)
	if err != nil {
		return nil, err
	}

	compiled, err := config.Compile(cfg)
	if err != nil {
		return nil, err
	}

	return compiled, nil
}

func watchFile(cfg *optionConfig) {
	defer close(cfg.watcherDone)

	ticker := time.NewTicker(cfg.autoReload)
	defer ticker.Stop()

	lastError := time.Time{}
	errorCount := 0
	const maxErrors = 10
	const maxBackoff = 5 * time.Minute

	for {
		select {
		case <-cfg.stopWatcher:
			return
		case event := <-cfg.watcher.Events:
			if event.Op&fsnotify.Write == fsnotify.Write {
				// File was modified - reload
				reloadConfig(cfg, &lastError, &errorCount, maxErrors, maxBackoff)
			}
		case err := <-cfg.watcher.Errors:
			if err != nil {
				reloadConfig(cfg, &lastError, &errorCount, maxErrors, maxBackoff)
			}
		case <-ticker.C:
			// Periodic check (fallback if fsnotify misses changes)
			reloadConfig(cfg, &lastError, &errorCount, maxErrors, maxBackoff)
		}
	}
}

func reloadConfig(cfg *optionConfig, lastError *time.Time, errorCount *int, maxErrors int, maxBackoff time.Duration) {
	now := time.Now()
	
	// Exponential backoff: only try if enough time has passed
	backoff := time.Duration(*errorCount) * time.Second
	if backoff > maxBackoff {
		backoff = maxBackoff
	}
	
	if now.Sub(*lastError) < backoff {
		return
	}

	newConfig, err := loadConfig(cfg.filePath)
	if err != nil {
		*lastError = now
		*errorCount++
		if *errorCount >= maxErrors {
			// Stop trying after max errors
			return
		}
		// Log via hooks if available
		if cfg.hooks != nil && cfg.hooks.AfterEval != nil {
			cfg.hooks.AfterEval("", "", Error)
		}
		return
	}

	// Success - update config atomically
	cfg.compiled.Store(&newConfig)
	*errorCount = 0
}

