package logger

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	errExample = errors.New("fail")

	_messages   = fakeMessages(1000)
	_tenInts    = []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 0}
	_tenStrings = []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}
	_tenTimes   = []time.Time{
		time.Unix(0, 0),
		time.Unix(1, 0),
		time.Unix(2, 0),
		time.Unix(3, 0),
		time.Unix(4, 0),
		time.Unix(5, 0),
		time.Unix(6, 0),
		time.Unix(7, 0),
		time.Unix(8, 0),
		time.Unix(9, 0),
	}
	_oneUser = &user{
		Name:      "Jane Doe",
		Email:     "jane@test.com",
		CreatedAt: time.Date(1980, 1, 1, 12, 0, 0, 0, time.UTC),
	}
	_tenUsers = users{
		_oneUser,
		_oneUser,
		_oneUser,
		_oneUser,
		_oneUser,
		_oneUser,
		_oneUser,
		_oneUser,
		_oneUser,
		_oneUser,
	}
)

func fakeMessages(n int) []string {
	messages := make([]string, n)
	for i := range messages {
		messages[i] = fmt.Sprintf("Test logging, but use a somewhat realistic message length. (#%v)", i)
	}
	return messages
}

func getMessage(iter int) string {
	return _messages[iter%1000]
}

type users []*user

type user struct {
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

func fakeGoLoggerFields() map[string]interface{} {
	return map[string]interface{}{
		"int":     _tenInts[0],
		"ints":    _tenInts,
		"string":  _tenStrings[0],
		"strings": _tenStrings,
		"time":    _tenTimes[0],
		"times":   _tenTimes,
		"user1":   _oneUser,
		"user2":   _oneUser,
		"users":   _tenUsers,
		"error":   errExample,
	}
}

func fakeFields() map[string]interface{} {
	mydata := struct {
		Name string `json:"name"`
	}{
		Name: "my name",
	}
	j, _ := json.Marshal(mydata)

	return map[string]interface{}{
		"data":  "complicated",
		"int":   1,
		"float": 0.15123,
		"struct": struct {
			Name string
		}{
			Name: "something",
		},
		"json": string(j),
	}
}

func fakeData() map[string]interface{} {
	mydata := struct {
		Name string `json:"name"`
	}{
		Name: "my name",
	}
	j, _ := json.Marshal(mydata)

	return map[string]interface{}{
		"data":  "complicated",
		"int":   1,
		"float": 0.15123,
		"at":    GetCallerInfo(1, false),
		"struct": struct {
			Name string
		}{
			Name: "something",
		},
		"json": string(j),
	}
}

var stringerFields = map[string]interface{}{
	"name":      "fields",
	"type":      "tringer",
	"id":        "yes",
	"are cool?": false,
}

var optCoreFields = WithImmutableFields(fakeGoLoggerFields())
var optDiscard = WithOutput(ioutil.Discard)

func newTestWriter(opts ...WriterOption) (*writer, *bytes.Buffer) {
	b := &bytes.Buffer{}
	w := newWriter(WithOutput(b))
	w.Reconfigure(opts...)

	return w, b
}

func TestWrite(t *testing.T) {
	w, b := newTestWriter(WithImmutableFields(fakeFields()))
	w.Write("simple message")
	assert.Contains(t, b.String(), `simple message [data:complicated][float:0.15123][int:1][json:{\"name\":\"my name\"}][struct:{something}]`)

	w, b = newTestWriter(WithImmutableFields(nil))
	w.Write("simple message")
	assert.NotContains(t, b.String(), `[data:complicated]`)
	assert.Contains(t, b.String(), `simple message`)
}

func TestWriteDataWithFields(t *testing.T) {
	w, b := newTestWriter(WithImmutableFields(fakeFields()))
	w.WriteData("some values", fakeData())
	assert.Contains(t, b.String(), `some values [data:complicated][float:0.15123][int:1][json:{\"name\":\"my name\"}][struct:{something}]`)
	assert.Contains(t, b.String(), `at="writer_test.go:162"`)
	assert.Contains(t, b.String(), `data="complicated"`)
	assert.Contains(t, b.String(), `float="0.15123"`)
	assert.Contains(t, b.String(), `int="1"`)
	assert.Contains(t, b.String(), `json="{\"name\":\"my name\"}"`)
	assert.Contains(t, b.String(), `struct="{something}"`)
}

func TestWriteDataWithFieldsAndTags(t *testing.T) {
	w, b := newTestWriter(WithImmutableFields(fakeFields()), WithImmutableTags("test", "tags", "fields"))
	w.WriteData("some values", fakeData())
	assert.Contains(t, b.String(), `[test,tags,fields] some values [data:complicated][float:0.15123][int:1][json:{\"name\":\"my name\"}][struct:{something}]`)
	assert.Contains(t, b.String(), `at="writer_test.go:174"`)
	assert.Contains(t, b.String(), `data="complicated"`)
	assert.Contains(t, b.String(), `float="0.15123"`)
	assert.Contains(t, b.String(), `int="1"`)
	assert.Contains(t, b.String(), `json="{\"name\":\"my name\"}"`)
	assert.Contains(t, b.String(), `struct="{something}"`)
}

func TestWriterAsJSON(t *testing.T) {
	w, b := newTestWriter(optCoreFields, WithFormat(FormatJSON))
	w.Write("simple message")
	assert.Contains(t, b.String(), `"msg":"simple message"`)
	// TODO: JSON unmarshal and compare fields

	b.Reset()

	w.Write("simple message")
	assert.Contains(t, b.String(), `"msg":"simple message"`)
	// TODO: JSON unmarshal and compare fields

	b.Reset()

	w2 := NewWriter(optCoreFields, WithFormat(FormatJSON), WithImmutableTags("test", "tags"), WithOutput(b))
	w2.WriteData("some values", fakeFields())
	assert.Contains(t, b.String(), `"msg":"some values"`)
	assert.Contains(t, b.String(), `"data":"complicated"`)
	assert.Contains(t, b.String(), `"float":0.15123`)
	assert.Contains(t, b.String(), `"int":1`)
	assert.Contains(t, b.String(), `"json":"{\"name\":\"my name\"}"`)
	assert.Contains(t, b.String(), `"struct":{"Name":"something"}`)
	assert.Contains(t, b.String(), `"tags":["test","tags"]`)
	// TODO: JSON unmarshal and compare fields

	b.Reset()

	w3 := newWriter(optCoreFields, WithFormat(FormatJSON), WithImmutableTags("test", "tags"), WithOutput(b))
	w3.write("some values", nil, nil, nil)
	assert.Contains(t, b.String(), `"msg":"some values"`)
	assert.NotContains(t, b.String(), `"data":"complicated"`)
	assert.NotContains(t, b.String(), `"float":0.15123`)
	assert.NotContains(t, b.String(), `"int":1`)
	assert.NotContains(t, b.String(), `"json":"{\"name\":\"my name\"}"`)
	assert.NotContains(t, b.String(), `"struct":{"Name":"something"}`)
	assert.NotContains(t, b.String(), `"tags":["test","tags"]`)

	b.Reset()

	w4 := newWriter(optCoreFields, WithFormat(FormatJSON), WithImmutableTags("test", "tags"), WithOutput(b))
	w4.write("some values", nil, nil, map[string]interface{}{
		"error": make(chan int),
	})
	assert.Contains(t, b.String(), `"msg":"failed to parse`)
}

func TestWriterConcurrency(t *testing.T) {
	var wg sync.WaitGroup

	w, _ := newTestWriter(WithFormat(FormatJSON))
	l := w.Spawn()

	for i := 0; i < 5; i++ {
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

func BenchmarkBasicWriteDataForZapComparison(b *testing.B) {
	b.Run("Text/Write", func(b *testing.B) {
		log := NewWriter(optCoreFields, optDiscard)
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				log.Write("some values")
			}
		})
	})

	b.Run("Text/WriteData", func(b *testing.B) {
		log := NewWriter(optCoreFields, optDiscard, WithFormat(FormatText))
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				log.WriteData("some values", fakeGoLoggerFields())
			}
		})
	})

	b.Run("Text/Fields/Write", func(b *testing.B) {
		log := NewWriter(optCoreFields, optDiscard)
		log.Reconfigure(
			WithImmutableFields(fakeGoLoggerFields()),
			WithFormat(FormatText),
			WithOutput(ioutil.Discard),
		)
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				log.WriteData("some values", nil)
			}
		})
	})
}

func BenchmarkJSONFormat(b *testing.B) {
	b.Run("JSON/WriteData", func(b *testing.B) {
		log := NewWriter(optCoreFields, optDiscard)
		log.Reconfigure(
			WithFormat(FormatJSON),
			WithOutput(ioutil.Discard),
		)
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				log.WriteData("some values", fakeGoLoggerFields())
			}
		})
	})

	b.Run("JSON/Fields/Write", func(b *testing.B) {
		log := NewWriter(optCoreFields, optDiscard)
		log.Reconfigure(
			WithImmutableFields(fakeGoLoggerFields()),
			WithFormat(FormatJSON),
			WithOutput(ioutil.Discard),
		)
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				log.WriteData("some values", nil)
			}
		})
	})

	b.Run("JSON/Fields/WriteData", func(b *testing.B) {
		log := NewWriter(optCoreFields, optDiscard)
		log.Reconfigure(
			WithImmutableFields(fakeGoLoggerFields()),
			WithFormat(FormatJSON),
			WithOutput(ioutil.Discard),
		)
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				log.WriteData("some values", fakeGoLoggerFields())
			}
		})
	})
}
