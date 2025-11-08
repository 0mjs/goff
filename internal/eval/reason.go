package eval

// Reason represents the reason for an evaluation result.
type Reason uint8

const (
	Match Reason = iota
	Percent
	Default
	Disabled
	Missing
	Error
)

