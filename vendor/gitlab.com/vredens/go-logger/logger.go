package logger

import "fmt"
import "sync"

// ConfigurableLogger interface defines the interface for a logger which can be reconfigured.
type ConfigurableLogger interface {
	Reconfigure(...Option)
	AddFields(map[string]interface{})
	AddTags(...string)
	AddData(KV)
}

// Option ...
type Option func(ConfigurableLogger)

// WithFields ...
func WithFields(fields map[string]interface{}) Option {
	return func(l ConfigurableLogger) {
		l.AddFields(fields)
	}
}

// WithTags ...
func WithTags(tags ...string) Option {
	return func(l ConfigurableLogger) {
		l.AddTags(tags...)
	}
}

// WithData ...
func WithData(data map[string]interface{}) Option {
	return func(l ConfigurableLogger) {
		l.AddData(data)
	}
}

// StdLogger ...
type StdLogger struct {
	writer  *writer
	mutex   sync.RWMutex
	tags    []string
	tagsIdx map[string]byte
	fields  map[string]string
	data    map[string]interface{}
}

// Spawn creates a new instance of a Logger from the log writer.
func (w *writer) Spawn(opts ...Option) Logger {
	return w.spawn(opts...)
}

// SpawnCompatible creates a new instance of a Logger using the parent Logger as origin.
func (w *writer) SpawnCompatible(opts ...Option) Compatible {
	return w.spawn(opts...)
}

func (w *writer) spawn(opts ...Option) *StdLogger {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	l := &StdLogger{
		writer:  w,
		fields:  map[string]string{},
		tags:    []string{},
		tagsIdx: map[string]byte{},
	}
	l.addFields(w.fields)
	l.AddTags(w.tags...)
	for _, opt := range opts {
		opt(l)
	}

	return l
}

// Spawn creates a new instance of a Logger using the parent Logger as origin.
func (l *StdLogger) Spawn(opts ...Option) Logger {
	return l.spawn(opts...)
}

// SpawnCompatible creates a new instance of a Logger using the parent Logger as origin.
func (l *StdLogger) SpawnCompatible(opts ...Option) Compatible {
	return l.spawn(opts...)
}

func (l *StdLogger) spawn(opts ...Option) *StdLogger {
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	spawn := &StdLogger{
		writer:  l.writer,
		fields:  map[string]string{},
		tags:    []string{},
		tagsIdx: map[string]byte{},
		data:    map[string]interface{}{},
	}

	spawn.addFields(l.fields)
	spawn.AddData(l.data)
	spawn.AddTags(l.tags...)

	for _, opt := range opts {
		opt(spawn)
	}

	return spawn
}

// Reconfigure ...
func (l *StdLogger) Reconfigure(opts ...Option) {
	for _, opt := range opts {
		opt(l)
	}
}

// AddFields ...
func (l *StdLogger) AddFields(fields map[string]interface{}) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if fields == nil || len(fields) == 0 {
		return
	}

	for k, v := range fields {
		l.fields[k] = Stringify(v, false)
	}
}

func (l *StdLogger) addFields(fields map[string]string) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if fields == nil || len(fields) == 0 {
		return
	}

	for k, v := range fields {
		l.fields[k] = v
	}
}

// AddTags ...
func (l *StdLogger) AddTags(tags ...string) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if len(tags) == 0 {
		return
	}
	for _, t := range tags {
		// TODO: validate tags, must contain only alphanumeric fields
		if _, found := l.tagsIdx[t]; !found {
			l.tagsIdx[t] = 0
			l.tags = append(l.tags, t)
		}
	}
}

// AddData ...
func (l *StdLogger) AddData(data KV) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if data == nil || len(data) == 0 {
		return
	}

	for k, v := range data {
		l.data[k] = v
	}
}

func (l *StdLogger) write(tags []string, msg string, data map[string]interface{}) {
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	if tags != nil && len(tags) > 0 {
		if l.data != nil && len(l.data) > 0 {
			l.writer.write(msg, append(l.tags, tags...), l.fields, l.data)
		} else {
			l.writer.write(msg, append(l.tags, tags...), l.fields, data)
		}
	} else {
		if l.data != nil && len(l.data) > 0 {
			l.writer.write(msg, l.tags, l.fields, l.data)
		} else {
			l.writer.write(msg, l.tags, l.fields, data)
		}
	}
}

// Debug ...
func (l *StdLogger) Debug(msg string) {
	if l.writer.minLevel > DEBUG {
		return
	}
	l.write([]string{"DEBUG"}, msg, nil)
}

// Info ...
func (l *StdLogger) Info(msg string) {
	if l.writer.minLevel > INFO {
		return
	}
	l.write([]string{"INFO"}, msg, nil)
}

// Warn ...
func (l *StdLogger) Warn(msg string) {
	if l.writer.minLevel > WARN {
		return
	}
	l.write([]string{"WARN"}, msg, nil)
}

// Error ...
func (l *StdLogger) Error(msg string) {
	if l.writer.minLevel > ERROR {
		return
	}
	l.write([]string{"ERROR"}, msg, nil)
}

// DebugData ...
func (l *StdLogger) DebugData(msg string, data map[string]interface{}) {
	if l.writer.minLevel > DEBUG {
		return
	}
	l.write([]string{"DEBUG"}, msg, data)
}

// InfoData ...
func (l *StdLogger) InfoData(msg string, data map[string]interface{}) {
	if l.writer.minLevel > INFO {
		return
	}
	l.write([]string{"INFO"}, msg, data)
}

// WarnData ...
func (l *StdLogger) WarnData(msg string, data map[string]interface{}) {
	if l.writer.minLevel > WARN {
		return
	}
	l.write([]string{"WARN"}, msg, data)
}

// ErrorData ...
func (l *StdLogger) ErrorData(msg string, data map[string]interface{}) {
	if l.writer.minLevel > ERROR {
		return
	}
	l.write([]string{"ERROR"}, msg, data)
}

// Debugf ...
func (l *StdLogger) Debugf(format string, args ...interface{}) {
	if l.writer.minLevel > DEBUG {
		return
	}
	l.write([]string{"DEBUG"}, fmt.Sprintf(format, args...), nil)
}

// Infof ...
func (l *StdLogger) Infof(format string, args ...interface{}) {
	if l.writer.minLevel > INFO {
		return
	}
	l.write([]string{"INFO"}, fmt.Sprintf(format, args...), nil)
}

// Warnf ...
func (l *StdLogger) Warnf(format string, args ...interface{}) {
	if l.writer.minLevel > WARN {
		return
	}
	l.write([]string{"WARN"}, fmt.Sprintf(format, args...), nil)
}

// Errorf ...
func (l *StdLogger) Errorf(format string, args ...interface{}) {
	if l.writer.minLevel > ERROR {
		return
	}
	l.write([]string{"ERROR"}, fmt.Sprintf(format, args...), nil)
}

// Print ...
func (l *StdLogger) Print(args ...interface{}) {
	if l.writer.minLevel > RAW {
		return
	}
	l.write(nil, fmt.Sprint(args...), nil)
}

// Println ...
func (l *StdLogger) Println(args ...interface{}) {
	if l.writer.minLevel > RAW {
		return
	}
	l.write(nil, fmt.Sprint(args...), nil)
}

// Printf ...
func (l *StdLogger) Printf(format string, args ...interface{}) {
	if l.writer.minLevel > RAW {
		return
	}
	l.write(nil, fmt.Sprintf(format, args...), nil)
}
