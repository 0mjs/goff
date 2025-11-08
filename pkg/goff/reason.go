package goff

// Reason is a small enum for audit/metrics.
type Reason uint8

const (
	Match Reason = iota
	Percent
	Default
	Disabled
	Missing
	Error
)

