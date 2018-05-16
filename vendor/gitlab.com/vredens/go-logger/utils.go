package logger

import (
	"fmt"
	"path"
	"runtime"
)

// GetCallerInfo is a helper for returning a string with "filename:line" of the call depth provided. Use this to add to a log field.
func GetCallerInfo(depth int, fullPath bool) string {
	_, file, line, _ := runtime.Caller(depth + 1)

	if fullPath {
		return fmt.Sprintf("%s:%d", file, line)
	}

	return fmt.Sprintf("%s:%d", path.Base(file), line)
}

// Stringify attempts to convert any interface to a string. If the interface implements the fmt.Stringer interface then it returns the result of String().
func Stringify(v interface{}, quoted bool) string {
	s, ok := v.(string)
	if !ok {
		s = fmt.Sprint(v)
	}

	if len(s) == 0 {
		return ""
	}
	str := fmt.Sprintf("%q", s)

	if quoted {
		return str
	}
	return str[1 : len(str)-1]
}

// F is a wrapper for fmt.Sprintf, its use is not recommended. If you do use it, you are labeling yourself lazy.
func F(format string, args ...interface{}) string {
	return fmt.Sprintf(format, args...)
}
