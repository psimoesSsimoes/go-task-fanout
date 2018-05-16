package logger

// With returns a new instance of an Entry logger with the provided tags and fields.
func (l *StdLogger) With(tags []string, fields map[string]interface{}) Entry {
	return l.spawn(WithFields(fields), WithTags(tags...))
}

// WithData creates a new instance of an Entry logger with the provided data.
func (l *StdLogger) WithData(data map[string]interface{}) Entry {
	return l.spawn(WithData(data))
}

// WithFields creates a new instance of an Entry logger with the provided fields.
func (l *StdLogger) WithFields(fields map[string]interface{}) Entry {
	return l.spawn(WithFields(fields))
}

// WithTags creates a new instance of an Entry logger with the provided tags.
func (l *StdLogger) WithTags(tags ...string) Entry {
	return l.spawn(WithTags(tags...))
}
