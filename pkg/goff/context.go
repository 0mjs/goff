package goff

// Context identifies the subject of evaluation and its attributes.
type Context struct {
	Key   string         // stable identifier; required for sticky bucketing
	Attrs map[string]any // optional attributes for targeting
}
