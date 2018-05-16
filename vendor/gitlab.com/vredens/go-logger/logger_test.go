package logger

import (
	"bytes"
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitializeLoggers(t *testing.T) {
	w, b := newTestWriter()

	l1 := w.Spawn()
	l1.Info("here")
	assert.Contains(t, b.String(), "here")

	b.Reset()

	l2 := w.SpawnCompatible()
	l2.Print("here2")
	assert.Contains(t, b.String(), "here2")

	b.Reset()

	l3 := l1.Spawn()
	l3.Info("here3")
	assert.Contains(t, b.String(), "here3")

	b.Reset()

	lt := w.spawn()
	l4 := lt.SpawnCompatible()
	l4.Print("here4")
	assert.Contains(t, b.String(), "here4")
}

func TestLoggingSimpleMessage(t *testing.T) {
	w, b := newTestWriter()
	l := w.Spawn()

	l.Debug("debug")
	assert.Contains(t, b.String(), `debug`)
	b.Reset()

	l.Info("info")
	assert.Contains(t, b.String(), `info`)
	b.Reset()

	l.Warn("warn")
	assert.Contains(t, b.String(), `warn`)
	b.Reset()

	l.Error("error")
	assert.Contains(t, b.String(), `error`)
	b.Reset()
}

func TestWriterReconfiguration(t *testing.T) {
	w, b := newTestWriter()
	l := w.Spawn()
	l.Debug("sample")
	checkBuffer(t, b)
	assert.Contains(t, b.String(), "sample")

	b.Reset()

	w.Reconfigure(WithLevel(INFO))
	l.Debug("sample")
	assert.Equal(t, b.String(), "")

	b2 := &bytes.Buffer{}
	w.Reconfigure(WithOutput(b2))

	l.Info("sample")
	checkBuffer(t, b2)
	assert.Equal(t, b.String(), "")
	assert.Contains(t, b2.String(), "sample")

	b2.Reset()

	// fields are immutable once a logger has been created
	w.Reconfigure(WithImmutableFields(map[string]interface{}{
		"test": "test",
	}))

	l.Info("sample")
	checkBuffer(t, b2)
	assert.Contains(t, b2.String(), "sample")
	assert.NotContains(t, b2.String(), `test="test"`)
}

func TestFieldsAndSpawnsOfLogger(t *testing.T) {
	w, b := newTestWriter()

	// fields are immutable once a logger has been created
	l := w.Spawn(WithFields(map[string]interface{}{
		"test": "test",
	}))

	l.Info("sample")
	checkBuffer(t, b)
	assert.Contains(t, b.String(), "sample")
	assert.Contains(t, b.String(), `[test:test]`)

	b.Reset()

	// fields are immutable once a logger has been created
	l2 := l.Spawn(WithFields(map[string]interface{}{
		"test-spawn": "test",
	}))

	l2.Info("sample")
	checkBuffer(t, b)
	assert.Contains(t, b.String(), "sample")
	assert.Contains(t, b.String(), `[test:test]`)
	assert.Contains(t, b.String(), `[test-spawn:test]`)
}

func TestTagsAndSpawnsOfLogger(t *testing.T) {
	w, b := newTestWriter(WithImmutableTags("immutable"))
	l := w.Spawn(WithTags("hello"))

	l.Info("sample")
	assert.Contains(t, b.String(), `immutable`)
	assert.Contains(t, b.String(), `hello`)

	b.Reset()

	w.Reconfigure(WithImmutableTags("extra"))
	l.Info("sample")
	checkBuffer(t, b)
	assert.Contains(t, b.String(), `immutable`)
	assert.Contains(t, b.String(), `hello`)
	assert.NotContains(t, b.String(), `extra`)

	b.Reset()

	l2 := w.Spawn()
	l2.Info("sample")
	checkBuffer(t, b)
	assert.NotContains(t, b.String(), `immutable`)
	assert.NotContains(t, b.String(), "hello")
	assert.Contains(t, b.String(), `extra`)

	b.Reset()

	l2 = l.Spawn(WithTags("goodbye"))
	l2.Info("sample #2")
	checkBuffer(t, b)
	assert.Contains(t, b.String(), `immutable`)
	assert.Contains(t, b.String(), `hello`)
	assert.Contains(t, b.String(), `goodbye`)
}

func TestLoggerInterfaceMethods(t *testing.T) {
	w, b := newTestWriter()
	l := w.Spawn(WithFields(KV{"field": "val"}))

	l2 := w.Spawn(
		WithFields(Fields{"component": "test"}),
		WithTags("one", "two"),
	)
	l3 := l2.Spawn(WithFields(Fields{"service": "test", "component": "test-2"}))
	l3.InfoData("stuff", KV{"xpto": "amen"})
	assert.Contains(t, b.String(), `xpto="amen"`)
	b.Reset()

	l.Debug("test")
	assert.Contains(t, b.String(), `test`)
	b.Reset()
	l.Debugf("test %s", "a")
	assert.Contains(t, b.String(), `test a`)
	b.Reset()
	l.DebugData("test", Fields{
		"test": "field",
	})
	assert.Contains(t, b.String(), `test [field:val] test="field"`)
	b.Reset()

	l.Info("test")
	assert.Contains(t, b.String(), `test`)
	b.Reset()
	l.Infof("test %s", "a")
	assert.Contains(t, b.String(), `test a`)
	b.Reset()
	l.InfoData("test", Fields{
		"test": "field",
	})
	assert.Contains(t, b.String(), `test [field:val] test="field"`)
	b.Reset()

	l.Warn("test")
	assert.Contains(t, b.String(), `test`)
	b.Reset()
	l.Warnf("test %s", "a")
	assert.Contains(t, b.String(), `test a`)
	b.Reset()
	l.WarnData("test", Fields{
		"test": "field",
	})
	assert.Contains(t, b.String(), `test [field:val] test="field"`)
	b.Reset()

	l.Error("test")
	assert.Contains(t, b.String(), `test`)
	b.Reset()
	l.Errorf("test %s", "a")
	assert.Contains(t, b.String(), `test a`)
	b.Reset()
	l.ErrorData("test", Fields{
		"test": "field",
	})
	assert.Contains(t, b.String(), `test [field:val] test="field"`)
	b.Reset()
}

func TestLoggerLogLevelChecking(t *testing.T) {
	w, b := newTestWriter()
	l := w.spawn(WithTags("ttag"))
	c := w.SpawnCompatible(WithTags("ttag"))

	w.Reconfigure(WithLevel(DEBUG))
	c.Print("test")
	c.Printf("test %s", "a")
	c.Println("test")
	assert.Empty(t, b.String())
	l.Debug("ok")
	assert.Contains(t, b.String(), `ok`)
	assert.Contains(t, b.String(), `[ttag,DEBUG]`)
	b.Reset()
	l.Info("ok")
	assert.Contains(t, b.String(), `ok`)
	assert.Contains(t, b.String(), `[ttag,INFO]`)
	b.Reset()
	l.Warn("ok")
	assert.Contains(t, b.String(), `ok`)
	assert.Contains(t, b.String(), `[ttag,WARN]`)
	b.Reset()
	l.Error("ok")
	assert.Contains(t, b.String(), `ok`)
	assert.Contains(t, b.String(), `[ttag,ERROR]`)
	b.Reset()

	w.Reconfigure(WithLevel(INFO))
	l.Debug("test")
	l.Debugf("test %s", "a")
	l.DebugData("test", Fields{"test": "field"})
	assert.Empty(t, b.String())
	l.Info("ok")
	assert.Contains(t, b.String(), `ok`)
	b.Reset()
	l.Warn("ok")
	assert.Contains(t, b.String(), `ok`)
	b.Reset()
	l.Error("ok")
	assert.Contains(t, b.String(), `ok`)
	b.Reset()

	w.Reconfigure(WithLevel(WARN))
	l.Info("test")
	l.Infof("test %s", "a")
	l.InfoData("test", Fields{"test": "field"})
	assert.Empty(t, b.String())
	l.Warn("ok")
	assert.Contains(t, b.String(), `ok`)
	b.Reset()
	l.Error("ok")
	assert.Contains(t, b.String(), `ok`)
	b.Reset()

	w.Reconfigure(WithLevel(ERROR))
	l.Warn("test")
	l.Warnf("test %s", "a")
	l.WarnData("test", Fields{"test": "field"})
	assert.Empty(t, b.String())
	l.Error("ok")
	assert.Contains(t, b.String(), `ok`)
	b.Reset()

	w.Reconfigure(WithLevel(OFF))
	l.Error("test")
	l.Errorf("test %s", "a")
	l.ErrorData("test", Fields{"test": "field"})
	assert.Empty(t, b.String())
}

func TestLogConcurrent(t *testing.T) {
	var wg sync.WaitGroup

	w, _ := newTestWriter()
	l := w.Spawn()

	for i := 0; i < 2; i++ {
		wg.Add(1)
		data := map[string]interface{}{
			"worker": i,
		}
		go func() {
			l.DebugData("worker", data)
			wg.Done()
		}()
	}
	wg.Wait()
}

func TestSpawnConcurrency(t *testing.T) {
	var wg sync.WaitGroup

	w, _ := newTestWriter(WithFormat(FormatJSON))
	l := w.spawn()

	for i := 0; i < 5; i++ {
		wg.Add(1)
		data := map[string]interface{}{
			"worker": i,
		}
		go func() {
			l.spawn(WithFields(data))
			l.WithFields(data).Info("test")
			wg.Done()
		}()
	}
	wg.Wait()
}

func BenchmarkLoggerSpawning(b *testing.B) {
	b.Run("Spawn/WithTags", func(b *testing.B) {
		w := NewWriter(optCoreFields, optDiscard)
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				w.Spawn(WithTags("some", "cool", "tags", "three", "are", "enough", "oh", "wait"))
			}
		})
	})

	b.Run("Spawn/WithFields", func(b *testing.B) {
		w := NewWriter(optDiscard)
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				w.Spawn(WithFields(fakeGoLoggerFields()))
			}
		})
	})

	b.Run("Spawn/WithStringerFields", func(b *testing.B) {
		w := NewWriter(optDiscard)
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				w.Spawn(WithFields(stringerFields))
			}
		})
	})
}

func BenchmarkSimpleForZapComparison(b *testing.B) {
	w := NewWriter(optDiscard)

	b.Run("Text/Log", func(b *testing.B) {
		log := w.SpawnSimple(WithFields(fakeGoLoggerFields()), WithTags("tag-a"))
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				log.Log("some values", nil)
			}
		})
	})

	b.Run("Text/Log/Tags", func(b *testing.B) {
		log := w.SpawnSimple(WithFields(fakeGoLoggerFields()), WithTags("tag-a"))
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				log.Log("some values", Tags{"a"})
			}
		})
	})

	b.Run("Text/Logd/TagsAndFields", func(b *testing.B) {
		log := w.SpawnSimple(WithTags("tag-a"))
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				log.Logd("some values", Tags{"a"}, fakeGoLoggerFields())
			}
		})
	})
}

func BenchmarkLoggerForZapComparison(b *testing.B) {
	w := NewWriter(WithFormat(FormatText), optDiscard)

	b.Run("Text/ContextFields/Info", func(b *testing.B) {
		log := w.Spawn(WithFields(fakeGoLoggerFields()))
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				log.Info("some values")
			}
		})
	})

	b.Run("Text/InfoData", func(b *testing.B) {
		w := NewWriter(WithFormat(FormatText), optDiscard)
		log := w.Spawn()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				log.InfoData("some values", fakeGoLoggerFields())
			}
		})
	})

	b.Run("Text/WithData/Info", func(b *testing.B) {
		w := NewWriter(WithFormat(FormatText), optDiscard)
		log := w.Spawn()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				log.WithData(fakeGoLoggerFields()).Info("some values")
			}
		})
	})
}

var optDebugDiscard = WithOutput(os.Stdout)

func checkBuffer(t *testing.T, b *bytes.Buffer) {
	// t.Helper()
	if b.Len() == 0 {
		t.Log("nothing was logged")
		t.FailNow()
	}
	b.Truncate(b.Len() - 1) // remove trailing \n
}
