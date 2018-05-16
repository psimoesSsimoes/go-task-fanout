package app

// This logger is a wrapper around another logger.

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	log "gitlab.com/vredens/go-logger"
)

// KV is an alias for map[string]interface{}
type KV map[string]interface{}

// LoggerConfig structure.
type LoggerConfig struct {
	// Level defines the minimum log level, anything lower is not logged. Levels are: debug, info, warn and error.
	Level string `json:"level" mapstructure:"level"`
	// Mode is the named template to use for each log line. Valid templates: noob, dev, simple, json. Defaults to json.
	Mode string `json:"mode" mapstructure:"mode"`
}

// ILogger defines the application logging interface
type ILogger interface {
	log.Logger
	LogError(err error, msg string)
	LogErrorf(err error, format string, args ...interface{})
	LogFatalError(err error, msg string)
	LogCallers(msg string, endDepth int)
}

type logger struct {
	log.Logger
}

// Logger is the global logger
var Logger = NewLogger("app")

// NewLogger ...
func NewLogger(component string, tags ...string) ILogger {
	return newLogger(component, tags...)
}

func newLogger(component string, tags ...string) *logger {
	return &logger{
		log.Spawn(
			log.WithFields(log.Fields{"component": component}),
			log.WithTags(tags...),
		),
	}
}

// LogError logs an error and evaluates application errors in order to also log the error stack.
func (l *logger) LogError(err error, msg string) {
	if msg != "" {
		l.Error(msg + "; " + StringifyError(err))
	} else {
		l.Error(StringifyError(err))
	}
}

// LogErrorf logs an error and evaluates application errors in order to also log the error stack.
func (l *logger) LogErrorf(err error, format string, args ...interface{}) {
	l.Error(fmt.Sprintf(format, args...) + "; " + StringifyError(err))
}

// LogFatalError ...
// @deprecated
func (l *logger) LogFatalError(err error, msg string) {
	if msg != "" {
		l.Error(msg + "; " + StringifyError(err))
		os.Exit(1)
	} else {
		l.Error(StringifyError(err))
		os.Exit(1)
	}
}

// LogCallers logs the provided msg and the list of callers ending after endDepth call stack.
func (l *logger) LogCallers(msg string, endDepth int) {
	tracing := make([]string, endDepth)
	tracing[0] = msg
	for i := 1; i < endDepth; i++ {
		_, file, line, ok := runtime.Caller(i)
		if ok {
			tracing[i] = fmt.Sprintf("[%s:%d]", file, line)
		}
	}

	l.Debugf("%s %s", msg, strings.Join(tracing, ";"))
}

// SetLogLevel updates the minimum log level, returns false if the provided level was not valid.
func SetLogLevel(level string) bool {
	switch level {
	case "debug":
		log.Reconfigure(log.WithLevel(log.DEBUG))
	case "info":
		log.Reconfigure(log.WithLevel(log.INFO))
	case "warn":
		log.Reconfigure(log.WithLevel(log.WARN))
	case "error":
		log.Reconfigure(log.WithLevel(log.ERROR))
	case "off":
		log.Reconfigure(log.WithLevel(log.OFF))
	default:
		return false
	}
	return true
}

// SetLogFormat sets logger output format based on preset templates.
func SetLogFormat(format string) {
	switch format {
	case "noob", "dev", "text", "simple":
		log.Reconfigure(log.WithFormat(log.FormatText))
	default:
		log.Reconfigure(log.WithFormat(log.FormatJSON))
	}
}

// ConfigureLogger sets the logger configuration for the default logger instance.
func ConfigureLogger(format, level string) {
	SetLogFormat(format)
	SetLogLevel(level)
}
