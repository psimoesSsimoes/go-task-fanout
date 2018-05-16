package logger

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

type writer struct {
	format     Format
	output     io.Writer
	bufferPool sync.Pool
	mutex      sync.RWMutex
	minLevel   Level
	tags       []string
	fields     map[string]string
	fieldOrder []string
}

// NewWriter creates a new log Writer.
func NewWriter(opts ...WriterOption) Writer {
	return newWriter(opts...)
}

func newWriter(opts ...WriterOption) *writer {
	w := &writer{
		bufferPool: sync.Pool{
			New: func() interface{} {
				return bytes.NewBuffer(make([]byte, 512))
			},
		},
		fields:     map[string]string{},
		fieldOrder: make([]string, 0),
		mutex:      sync.RWMutex{},
		minLevel:   unknown,
	}
	w.output = os.Stdout

	w.Reconfigure(opts...)

	return w
}

// WriterOption are options for all loggers implementing the Basic logger interface.
type WriterOption func(*writer)

// Reconfigure ...
func (w *writer) Reconfigure(opts ...WriterOption) {
	for _, o := range opts {
		o(w)
	}
}

func (w *writer) write(msg string, tags []string, fields map[string]string, data map[string]interface{}) {
	switch w.format {
	case FormatJSON:
		w.writeJSON(msg, tags, fields, data)
	default:
		w.writeText(msg, tags, fields, data)
	}
}

// Write ...
func (w *writer) Write(msg string) {
	w.WriteData(msg, nil)
}

// WriteData ...
func (w *writer) WriteData(msg string, data map[string]interface{}) {
	switch w.format {
	case FormatJSON:
		w.writeJSON(msg, w.tags, w.fields, data)
	default:
		w.writeText(msg, w.tags, w.fields, data)
	}
}

func (w *writer) writeJSON(msg string, tags []string, fields map[string]string, data map[string]interface{}) {
	entry := make(map[string]interface{})

	for k, v := range fields {
		entry[k] = v
	}

	entry["msg"] = msg
	entry["tags"] = tags
	entry["data"] = data
	entry["timestamp"] = time.Now().Format(time.RFC3339Nano)

	j, err := json.Marshal(entry)
	if err != nil {
		w.output.Write([]byte(`{"msg":"failed to parse log entry into valid json"}`))
		return
	}

	w.mutex.Lock()
	w.output.Write(j)
	w.output.Write([]byte("\n"))
	w.mutex.Unlock()
}

func (w *writer) writeText(msg string, tags []string, fields map[string]string, data map[string]interface{}) {
	b := w.bufferPool.Get().(*bytes.Buffer)
	b.Reset()
	defer w.bufferPool.Put(b)

	b.WriteByte('[')
	b.WriteString(time.Now().Format(time.RFC3339Nano))
	b.WriteByte(']')

	if tags != nil && len(tags) > 0 {
		b.WriteByte('[')
		b.WriteString(strings.Join(tags, ","))
		b.WriteByte(']')
	}

	b.WriteByte(' ')
	b.WriteString(msg)

	if fields != nil && len(fields) > 0 {
		b.WriteByte(' ')
		fieldNames := make([]string, 0, len(fields))
		for k := range fields {
			fieldNames = append(fieldNames, k)
		}
		sort.Strings(fieldNames)
		for _, k := range fieldNames {
			b.WriteByte('[')
			b.WriteString(k)
			b.WriteByte(':')
			b.WriteString(fields[k])
			b.WriteByte(']')
		}
	}

	if data != nil && len(data) > 0 {
		for k, v := range data {
			b.WriteString(" ")
			b.WriteString(k)
			b.WriteByte('=')
			b.WriteString(Stringify(v, true))
		}
	}

	b.WriteByte('\n')

	w.mutex.Lock()
	w.output.Write(b.Bytes())
	w.mutex.Unlock()
}

// WithOutput allows setting the output writer for the logger.
func WithOutput(w io.Writer) WriterOption {
	return func(lw *writer) {
		lw.output = w
	}
}

// WithFormat ...
func WithFormat(format Format) WriterOption {
	return func(w *writer) {
		w.format = format
	}
}

// WithLevel ...
func WithLevel(level Level) WriterOption {
	return func(w *writer) {
		w.minLevel = level
	}
}

// WithImmutableTags option, a writer has the tags provided replace its list of tags.
// A writer's tags can not be changed or overriden by a logger and at Spawn time all tags are copied to the spawn.
func WithImmutableTags(tags ...string) WriterOption {
	return func(w *writer) {
		w.mutex.Lock()
		defer w.mutex.Unlock()

		if len(tags) == 0 {
			return
		}
		w.tags = make([]string, 0, len(tags))
		for _, tag := range tags {
			// TODO: validate tags, must contain only alphanumeric fields
			w.tags = append(w.tags, tag)
		}
	}
}

// WithImmutableFields option, a writer has the provided fields replace its list of fields.
// A writer's fields can not be changed or overriden by a logger and at Spawn time are copied to the spawn.
func WithImmutableFields(fields map[string]interface{}) WriterOption {
	return func(w *writer) {
		w.mutex.Lock()
		defer w.mutex.Unlock()
		w.fields = map[string]string{}
		for k, v := range fields {
			w.fields[k] = Stringify(v, false)
		}
	}
}
