package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEntryLogger(t *testing.T) {
	w, b := newTestWriter(WithImmutableTags("immutable"))
	l := w.Spawn(WithTags("hello"), WithFields(KV{"field": "field-value"}))

	l.WithData(KV{"data": "vdata"}).Info("sample")
	assert.Contains(t, b.String(), `immutable`)
	assert.Contains(t, b.String(), `hello`)
	assert.Contains(t, b.String(), `field:field-value`)
	assert.Contains(t, b.String(), `data="vdata"`)

	b.Reset()

	l.WithData(KV{"data": "vdata"}).WithData(KV{"new": "stuff"}).Info("sample")
	assert.Contains(t, b.String(), `data="vdata"`)
	assert.Contains(t, b.String(), `new="stuff"`)

	b.Reset()

	l.WithTags("tag1").WithTags("tag2").Info("sample")
	assert.Contains(t, b.String(), `immutable,hello,tag1,tag2,INFO`)

	b.Reset()

	l1 := l.With(Tags{"tag1", "tag2"}, KV{"key1": "val1", "key2": "val2"})
	l1.Info("info")
	assert.Contains(t, b.String(), `[immutable,hello,tag1,tag2,INFO]`)
	assert.Contains(t, b.String(), `info`)
	assert.Contains(t, b.String(), `[key1:val1]`)
	assert.Contains(t, b.String(), `[key2:val2]`)

	b.Reset()

	l2 := l.WithFields(KV{"key1": "val1", "key2": "val2"})
	l2.Info("info")
	assert.Contains(t, b.String(), `[immutable,hello,INFO]`)
	assert.Contains(t, b.String(), `info`)
	assert.Contains(t, b.String(), `[key1:val1]`)
	assert.Contains(t, b.String(), `[key2:val2]`)
}
