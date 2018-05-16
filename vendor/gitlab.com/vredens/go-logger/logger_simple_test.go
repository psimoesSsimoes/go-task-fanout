package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSpawnSimple(t *testing.T) {
	w, b := newTestWriter()
	s := w.SpawnSimple()

	s.Dbg("test", nil, nil)
	assert.Contains(t, b.String(), "test")
	b.Reset()

	s1 := s.SpawnSimple()
	s1.Dbg("test", nil, nil)
	assert.Contains(t, b.String(), "test")
	b.Reset()

	w.Reconfigure(WithLevel(INFO))

	s.Dbg("test", nil, nil)
	assert.Empty(t, b.String())
}

func TestSimpleInterface(t *testing.T) {
	w, b := newTestWriter()
	s := w.SpawnSimple()

	s.Dbg("debug", Tags{"test-tag"}, KV{"key": "val"})
	assert.Contains(t, b.String(), `[test-tag] debug key="val"`)
	b.Reset()

	s.Write("message text", nil)
	assert.Contains(t, b.String(), `message text`)
	b.Reset()

	s.Write("message text", Tags{"test-tag"})
	assert.Contains(t, b.String(), `[test-tag] message text`)
	b.Reset()

	s.Dump("message text", nil, nil)
	assert.Contains(t, b.String(), `message text`)
	b.Reset()

	s.Dump("message text", nil, KV{"key1": "val1", "key2": "val2"})
	assert.Contains(t, b.String(), `message text`)
	assert.Contains(t, b.String(), `key1="val1"`)
	assert.Contains(t, b.String(), `key2="val2"`)
	b.Reset()

	s.Dump("message text", Tags{"test-tag1", "test-tag2"}, KV{"key1": "val1", "key2": "val2"})
	assert.Contains(t, b.String(), `[test-tag1,test-tag2] message text`)
	assert.Contains(t, b.String(), `key1="val1"`)
	assert.Contains(t, b.String(), `key2="val2"`)
	b.Reset()
}
