package logger

var gWriter = newWriter()
var gLogger = gWriter.spawn(WithFields(map[string]interface{}{
	"component": "app",
}))

// Spawn creates a new Logger directly from the log writer.
func Spawn(opts ...Option) Logger {
	return gWriter.spawn(opts...)
}

// SpawnCompatible creates a new logger which implements the Compatible interface.
func SpawnCompatible(opts ...Option) Compatible {
	return gWriter.spawn(opts...)
}

// SpawnSimple returns a new logger which implements the Simple interface.
func SpawnSimple(opts ...Option) Simple {
	return gWriter.spawn(opts...)
}

// Reconfigure ...
func Reconfigure(opts ...WriterOption) {
	gWriter.Reconfigure(opts...)
}

// Write ...
func Write(msg string) {
	gWriter.Write(msg)
}

// WriteData ...
func WriteData(msg string, data map[string]interface{}) {
	gWriter.WriteData(msg, data)
}

// Debug ...
func Debug(msg string) {
	gLogger.Debug(msg)
}

// Debugf ...
func Debugf(format string, args ...interface{}) {
	gLogger.Debugf(format, args...)
}

// DebugData ...
func DebugData(msg string, data map[string]interface{}) {
	gLogger.DebugData(msg, data)
}

// Info ...
func Info(msg string) {
	gLogger.Info(msg)
}

// Infof ...
func Infof(format string, args ...interface{}) {
	gLogger.Infof(format, args...)
}

// InfoData ...
func InfoData(msg string, data map[string]interface{}) {
	gLogger.InfoData(msg, data)
}

// Warn ...
func Warn(msg string) {
	gLogger.Warn(msg)
}

// Warnf ...
func Warnf(format string, args ...interface{}) {
	gLogger.Warnf(format, args...)
}

// WarnData ...
func WarnData(msg string, data map[string]interface{}) {
	gLogger.WarnData(msg, data)
}

// Error ...
func Error(msg string) {
	gLogger.Error(msg)
}

// Errorf ...
func Errorf(format string, args ...interface{}) {
	gLogger.Errorf(format, args...)
}

// ErrorData ...
func ErrorData(msg string, data map[string]interface{}) {
	gLogger.ErrorData(msg, data)
}

// Print for compatibility with Go's official log package.
func Print(args ...interface{}) {
	gLogger.Print(args...)
}

// Println for compatibility with Go's official log package.
func Println(args ...interface{}) {
	gLogger.Println(args...)
}

// Printf for compatibility with Go's official log package.
func Printf(format string, args ...interface{}) {
	gLogger.Printf(format, args...)
}

// Log ...
func Log(msg string, tags []string) {
	gLogger.Log(msg, tags)
}

// Logd ...
func Logd(msg string, tags []string, fields map[string]interface{}) {
	gLogger.Logd(msg, tags, fields)
}
