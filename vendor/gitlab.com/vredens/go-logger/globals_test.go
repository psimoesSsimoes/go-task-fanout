package logger

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGlobalLoggerInterface(t *testing.T) {
	b := &bytes.Buffer{}
	Reconfigure(WithOutput(b))

	Debug("test")
	assert.Contains(t, b.String(), "test")
	b.Reset()
	Debugf("test %s", "a")
	assert.Contains(t, b.String(), "test a")
	b.Reset()
	DebugData("test", Fields{"field": "sample"})
	assert.Contains(t, b.String(), ` test `)
	assert.Contains(t, b.String(), `field="sample"`)
	b.Reset()

	Info("test")
	assert.Contains(t, b.String(), "test")
	b.Reset()
	Infof("test %s", "a")
	assert.Contains(t, b.String(), "test a")
	b.Reset()
	InfoData("test", Fields{"field": "sample"})
	assert.Contains(t, b.String(), ` test `)
	assert.Contains(t, b.String(), `field="sample"`)
	b.Reset()

	Warn("test")
	assert.Contains(t, b.String(), "test")
	b.Reset()
	Warnf("test %s", "a")
	assert.Contains(t, b.String(), "test a")
	b.Reset()
	WarnData("test", Fields{"field": "sample"})
	assert.Contains(t, b.String(), ` test `)
	assert.Contains(t, b.String(), `field="sample"`)
	b.Reset()

	Error("test")
	assert.Contains(t, b.String(), "test")
	b.Reset()
	Errorf("test %s", "a")
	assert.Contains(t, b.String(), "test a")
	b.Reset()
	ErrorData("test", Fields{"field": "sample"})
	assert.Contains(t, b.String(), ` test `)
	assert.Contains(t, b.String(), `field="sample"`)
	b.Reset()

	Log("test", nil)
	assert.Contains(t, b.String(), "test")
	b.Reset()
	Log("test", Tags{"tag"})
	assert.Contains(t, b.String(), "test")
	assert.Contains(t, b.String(), "[tag]")
	b.Reset()
	Logd("test", nil, Fields{"key": "val"})
	assert.Contains(t, b.String(), ` test `)
	assert.Contains(t, b.String(), `key="val"`)
	b.Reset()
	Logd("test", Tags{"tag"}, Fields{"key": "val"})
	assert.Contains(t, b.String(), ` test `)
	assert.Contains(t, b.String(), `key="val"`)
	assert.Contains(t, b.String(), "[tag]")
	b.Reset()
}

func TestGlobalCompatibleInterface(t *testing.T) {
	b := &bytes.Buffer{}
	Reconfigure(WithOutput(b))

	Print("test")
	assert.Contains(t, b.String(), "test")
	b.Reset()
	Println("test")
	assert.Contains(t, b.String(), "test")
	b.Reset()
	Printf("test %s", "a")
	assert.Contains(t, b.String(), "test a")
	b.Reset()
}

func TestGlobalWriterInterface(t *testing.T) {
	b := &bytes.Buffer{}
	Reconfigure(WithOutput(b))

	l := Spawn()
	l.Debug("test")
	assert.Contains(t, b.String(), "test")
	b.Reset()
	c := SpawnCompatible()
	c.Print("test")
	assert.Contains(t, b.String(), "test")
	b.Reset()
	Write("test")
	assert.Contains(t, b.String(), "test")
	b.Reset()
	WriteData("test", Fields{"field": "sample"})
	assert.Contains(t, b.String(), `test field="sample"`)
	b.Reset()
}

func TestGlobalSpawns(t *testing.T) {
	l := SpawnSimple()

	if _, ok := l.(*StdLogger); !ok {
		t.Error("spawn simple failed")
	}
}
