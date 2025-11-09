# goff

Go-native feature flags. Local evaluation, zero network on the hot path, deterministic and fast.

## Quickstart

```bash
go get github.com/0mjs/goff
```

Create a flags configuration file:

```yaml
# flags.yaml
version: 1
flags:
  new_checkout:
    enabled: true
    type: "bool"
    variants:
      true: 50
      false: 50
    rules:
      - when:
          all:
            - attr: "plan"
              op: "eq"
              value: "pro"
        then:
          variants:
            true: 90
            false: 10
    default: false
```

Use in your code:

```go
package main

import (
    "log"
    "time"
    
    "github.com/0mjs/goff"
)

func main() {
    client, err := goff.New(
        goff.WithFile("flags.yaml"),
        goff.WithAutoReload(5 * time.Second),
    )
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    ctx := goff.Context{
        Key: "user:123",
        Attrs: map[string]any{
            "plan": "pro",
        },
    }

    enabled := client.Boolean("new_checkout", ctx, false)
    if enabled {
        // Use new checkout flow
    }
}
```

## Performance

Benchmarks on Apple M1 Pro (arm64):

```
BenchmarkEvalBool-8                	 5205440	       213.5 ns/op
BenchmarkEvalBool_WithRule-8       	 5324262	       259.4 ns/op
BenchmarkEvalString-8              	 3940990	       284.1 ns/op
BenchmarkEvalBool_Concurrent-8     	13462598	        93.11 ns/op
BenchmarkEvalString_Concurrent-8   	14095626	       103.2 ns/op
```

- P50: <0.3µs per evaluation
- P99: <1µs per evaluation
- Throughput: >4M eval/s/core

## Configuration Schema

### Flag Types

**Boolean flags:**
```yaml
new_feature:
  enabled: true
  type: "bool"
  variants:
    true: 50    # 50% get true
    false: 50   # 50% get false
  default: false
```

**String flags:**
```yaml
theme:
  enabled: true
  type: "string"
  variants:
    red: 40
    blue: 30
    green: 30
  default: "red"
```

### Targeting Rules

Rules allow you to target specific users based on attributes:

```yaml
flags:
  feature:
    enabled: true
    type: "bool"
    variants:
      true: 50
      false: 50
    rules:
      - when:
          all:                    # all conditions must match
            - attr: "plan"
              op: "eq"
              value: "pro"
            - attr: "region"
              op: "eq"
              value: "us"
        then:
          variants:
            true: 100
            false: 0
      - when:
          any:                    # any condition can match
            - attr: "beta"
              op: "eq"
              value: "true"
        then:
          variants:
            true: 90
            false: 10
    default: false
```

### Operators

- `eq` - equals
- `neq` - not equals
- `gt` - greater than
- `gte` - greater than or equal
- `lt` - less than
- `lte` - less than or equal
- `in` - value is in array
- `contains` - string contains substring
- `matches` - string matches regex pattern

## CLI

### Validate configuration

```bash
ffctl validate -f flags.yaml
```

### Evaluate a flag

```bash
ffctl eval -f flags.yaml -flag new_checkout -key user:123 -attr plan=pro -def=false
```

Output:
```
variant: true
reason: match
```

## API Reference

### Client

```go
type Client interface {
    Boolean(key string, ctx Context, def bool) bool
    String(key string, ctx Context, def string) string
    Close() error
}
```

### Context

```go
type Context struct {
    Key   string                 // stable identifier for sticky bucketing
    Attrs map[string]any // attributes for targeting
}
```

### Options

- `WithFile(path string)` - load configuration from file
- `WithAutoReload(interval time.Duration)` - automatically reload on file changes
- `WithHooks(hooks Hooks)` - set observability hooks

### Hooks

```go
type Hooks struct {
    AfterEval func(flag, variant string, reason Reason)
}
```

## Status

🚧 In development - v0.1.0 coming soon

## License

MIT
