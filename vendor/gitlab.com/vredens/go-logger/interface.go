package logger

// Writer ...
type Writer interface {
	Spawn(...Option) Logger
	SpawnSimple(...Option) Simple
	SpawnCompatible(...Option) Compatible
	Reconfigure(...WriterOption)
	Write(msg string)
	WriteData(msg string, data map[string]interface{})
}

// Simple interface is the a basic yet powerful interface for a logger. It encourages the use of tags instead of log levels for more control in terms of log filtering and search. It discourages fancy printf, requiring the user to run fmt.Printf or logger.F for a slightly shorter version.
type Simple interface {
	// Spawn creates a new logger with the provided configuration. Use this to change your default logger's fields and tags.
	SpawnSimple(...Option) Simple
	// Log writes a log line with the provided message and a list of tags. Use nil value for tags if you don't want to add any. Tags are appended to the logger's tags. There is no check for repeated tags.
	// @deprecated.
	Log(message string, tags []string)
	// Logd is similar to Log except it allows passing additional structured data in the form of map[string]interface{}.
	// @deprecated.
	Logd(message string, tags []string, data map[string]interface{})
	// Write writes a new log entry with the provided message and tags. The tags parameter can be nil and no extra tags will be included in the log entry.
	Write(msg string, tags []string)
	// Dump is similar to Write but it also logs extra data. The data parameter can be nil which has the same effect as calling the Write method.
	Dump(msg string, tags []string, data map[string]interface{})
	// Dbg is similar to Dump except it will not print anything if the minimum log level has been set to something higher than DEBUG. Note that the Simple logger itself will ignore any other minimum log level set.
	Dbg(message string, tags []string, data map[string]interface{})
}

// Entry is a special logger aimed at function/line scoped logging.
type Entry interface {
	// With method spawns a new Entry from the current one adding the provided tags and fields.
	With(tags []string, fields map[string]interface{}) Entry
	// WithTags updates  a new Entry logger with the provided tags.
	WithTags(tags ...string) Entry
	// WithData returns a new Entry logger with the provided data.
	WithData(data map[string]interface{}) Entry
	// WithFields returns a new Entry logger with the provided fields.
	WithFields(fields map[string]interface{}) Entry

	Debug(msg string)
	Info(msg string)
	Warn(msg string)
	Error(msg string)

	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}

// Logger is a typical logger interface.
type Logger interface {
	Spawn(...Option) Logger
	With(tags []string, fields map[string]interface{}) Entry
	WithTags(tags ...string) Entry
	WithFields(fields map[string]interface{}) Entry
	WithData(data map[string]interface{}) Entry

	Debug(msg string)
	Info(msg string)
	Warn(msg string)
	Error(msg string)

	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})

	// DebugData ...
	// @deprecated use WithData().Debug() instead
	DebugData(msg string, data map[string]interface{})
	// InfoData ...
	// @deprecated use WithData().Info() instead
	InfoData(msg string, data map[string]interface{})
	// WarnData ...
	// @deprecated use WithData().Warn() instead
	WarnData(msg string, data map[string]interface{})
	// ErrorData ...
	// @deprecated use WithData().Error() instead
	ErrorData(msg string, data map[string]interface{})
}

// Compatible is a compatibility interface for the standard logger for dropin replacement of the standard logger.
type Compatible interface {
	SpawnCompatible(...Option) Compatible
	Print(...interface{})
	Println(...interface{})
	Printf(string, ...interface{})
}
