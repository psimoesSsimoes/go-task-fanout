package logger

// Level for defining log level integer values for granular filtering.
type Level uint8

const (
	unknown Level = iota
	// RAW log level should be used only when dumping large log entries or entire structures, use in development only.
	RAW
	// DEBUG log level should be used for development only and potentially as on/off in particular situations for gathering more information during tricky debugging situations.
	DEBUG
	// INFO is anything relevant for the log analysis/consumer. For example, informing that a database connection was established, a message was received, etc. This should be used for log analysis, metrics and basic profiling of bottlenecks.
	INFO
	// WARN log level should be used in situations where many entries of this type should trigger an allert in your centralized logging system. It only makes sense if you plan to use such a system architecture.
	WARN
	// ERROR log level should be used in all error situations. Careful with what you consider an error. Typically if you are handling the situation or returning the error to the caller, you should not log it as an ERROR.
	ERROR
	// OFF log level should be used only by SetLevel method to disable logging.
	OFF
)

// Format is a constant defining the log format.
type Format uint8

const (
	// FormatText is used for setting the log formatter as a human readable log line in the style of syslog and similar.
	FormatText Format = iota
	// FormatJSON is used for setting the log formatter to JSON where each line is a single JSON object representing the log entry.
	FormatJSON
)

var levels = map[Level]string{
	unknown: "",
	RAW:     "RAW",
	DEBUG:   "DEBUG",
	INFO:    "INFO",
	WARN:    "WARN",
	ERROR:   "ERROR",
}
