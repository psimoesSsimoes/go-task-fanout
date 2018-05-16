package logger

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUtilGetCallerInfo(t *testing.T) {
	var str string

	str = GetCallerInfo(0, false)
	assert.Equal(t, "utils_test.go:13", str)

	cwd, _ := os.Getwd()

	str = GetCallerInfo(0, true)
	assert.Equal(t, cwd+"/utils_test.go:18", str)
}

func TestStringify(t *testing.T) {
	var str string

	str = Stringify("test", false)
	assert.Equal(t, `test`, str)

	str = Stringify([]string{"a", "b"}, false)
	assert.Equal(t, `[a b]`, str)

	str = Stringify(123, false)
	assert.Equal(t, `123`, str)

	str = Stringify("{an amazing test of power}", false)
	assert.Equal(t, `{an amazing test of power}`, str)

	str = Stringify("an amazing test of ultra power", true)
	assert.Equal(t, `"an amazing test of ultra power"`, str)

	str = Stringify("", false)
	assert.Equal(t, ``, str)
}

func TestF(t *testing.T) {
	var str string

	str = F("test %s", "a")
	assert.Equal(t, str, "test a")
}
