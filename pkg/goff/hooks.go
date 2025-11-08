package goff

// Hooks are optional and fast to call; they must not allocate on the hot path.
type Hooks struct {
	AfterEval func(flag, variant string, reason Reason)
}

