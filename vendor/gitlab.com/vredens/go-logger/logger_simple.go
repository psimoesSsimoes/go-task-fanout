package logger

// SpawnCompatible creates a new instance of a Logger using the parent Logger as origin.
func (w *writer) SpawnSimple(opts ...Option) Simple {
	return w.spawn(opts...)
}

// SpawnSimple creates a new instance of a Logger using the parent Logger as origin.
func (l *StdLogger) SpawnSimple(opts ...Option) Simple {
	return l.spawn(opts...)
}

// Log ...
func (l *StdLogger) Log(msg string, tags []string) {
	l.write(tags, msg, nil)
}

// Logd ...
func (l *StdLogger) Logd(msg string, tags []string, data map[string]interface{}) {
	l.write(tags, msg, data)
}

// Write writes a new log entry with the provided message and tags. The tags parameter can be nil and no extra tags will be included in the log entry.
func (l *StdLogger) Write(msg string, tags []string) {
	l.write(tags, msg, nil)
}

// Dump is similar to Write but it also logs extra data. The data parameter can be nil which has the same effect as calling the Write method.
func (l *StdLogger) Dump(msg string, tags []string, data map[string]interface{}) {
	l.write(tags, msg, data)
}

// Dbg is similar to Dump except it will not print anything if the minimum log level has been set to something higher than DEBUG. Note that the Simple logger itself will ignore any other minimum log level set.
func (l *StdLogger) Dbg(msg string, tags []string, data map[string]interface{}) {
	if l.writer.minLevel > DEBUG {
		return
	}
	l.write(tags, msg, data)
}
