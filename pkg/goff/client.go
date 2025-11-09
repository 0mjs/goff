package goff

import (
	"sync/atomic"

	"github.com/0mjs/goff/internal/config"
	"github.com/0mjs/goff/internal/eval"
)

// Client evaluates flags against an immutable configuration snapshot.
type Client interface {
	Boolean(key string, ctx Context, def bool) bool
	String(key string, ctx Context, def string) string
	Close() error
}

type client struct {
	config *atomic.Pointer[*config.Compiled]
	hooks  *Hooks
	closer func() error
}

// Boolean evaluates a boolean flag.
func (c *client) Boolean(key string, ctx Context, def bool) bool {
	compiled := c.config.Load()
	if compiled == nil {
		return def
	}

	flag := (*compiled).Flags[key]

	evalCtx := eval.Context{
		Key:   ctx.Key,
		Attrs: ctx.Attrs,
	}

	result, reason := eval.EvalBool(flag, key, evalCtx, def)

	if c.hooks != nil && c.hooks.AfterEval != nil {
		variant := "false"
		if result {
			variant = "true"
		}
		c.hooks.AfterEval(key, variant, Reason(reason))
	}

	return result
}

// String evaluates a string flag.
func (c *client) String(key string, ctx Context, def string) string {
	compiled := c.config.Load()
	if compiled == nil {
		return def
	}

	flag := (*compiled).Flags[key]

	evalCtx := eval.Context{
		Key:   ctx.Key,
		Attrs: ctx.Attrs,
	}

	result, reason := eval.EvalString(flag, key, evalCtx, def)

	if c.hooks != nil && c.hooks.AfterEval != nil {
		c.hooks.AfterEval(key, result, Reason(reason))
	}

	return result
}

// Close closes the client and stops any background operations.
func (c *client) Close() error {
	if c.closer != nil {
		return c.closer()
	}
	return nil
}
